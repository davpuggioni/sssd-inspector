// catalog_gen.go
//
// Generates sssd_catalog/catalog.json and sssd_catalog/catalog.md from upstream
// SSSD API definitions (src/config/etc/sssd.api.conf and sssd.api.d/*.conf)
// and/or DocBook XML man pages (src/man/*.xml, src/man/include/*.xml) or compiled
// roff man pages (sssd*.5, sssd*.5.gz).
package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// catalogOutputDir is where RunGenerateCatalog writes catalog.json/catalog.md.
// It is a variable (not a constant) so tests can redirect the output into a
// temporary directory instead of overwriting the committed catalog.
var catalogOutputDir = "sssd_catalog"

// OptionMeta holds metadata for a single SSSD configuration option.
type OptionMeta struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"` // "bool", "int", "string", "list"
	Sections []string `json:"sections"`
	Default  string   `json:"default,omitempty"`
	Values   []string `json:"values,omitempty"`
	Doc      string   `json:"doc,omitempty"`
	// Source is the documentation file this option was extracted from
	// (for example "sssd.conf.5.xml" or "sssd.api.conf"). It is surfaced in
	// the report so every catalog finding can be traced back to the exact
	// man page it came from.
	Source string `json:"source,omitempty"`
}

// SssdCatalog represents the consolidated schema catalog of SSSD options.
type SssdCatalog struct {
	Source      string                `json:"source"`
	Version     string                `json:"version"`
	Generated   string                `json:"generated"`
	Sources     []string              `json:"sources"`
	Sections    map[string][]string   `json:"sections"`
	Options     map[string]OptionMeta `json:"options"`
	OptionNames []string              `json:"option_names"`
}

// RunGenerateCatalog parses the contents of sourceDir and writes
// sssd_catalog/catalog.json and sssd_catalog/catalog.md.
//
// The destination is catalogOutputDir (the committed sssd_catalog/ directory in
// the repository root); tests override it so a generation run never clobbers
// the committed catalog.
func RunGenerateCatalog(sourceDir string) error {
	if sourceDir == "" {
		sourceDir = "upstream"
	}
	info, err := os.Stat(sourceDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("source directory %q does not exist or is not a directory.\n"+
			"Please copy upstream SSSD definitions or man pages into %q.\n"+
			"See %s/README.md for instructions", sourceDir, sourceDir, sourceDir)
	}

	catalog := &SssdCatalog{
		Source:    "SSSD upstream",
		Version:   "unknown",
		Generated: time.Now().UTC().Format("2006-01-02"),
		Sections:  make(map[string][]string),
		Options:   make(map[string]OptionMeta),
	}

	discoveredSources := make(map[string]bool)

	// Step 1: Detect SSSD version from version.m4 if present
	if vData, err := os.ReadFile(filepath.Join(sourceDir, "version.m4")); err == nil {
		re := regexp.MustCompile(`m4_define\(\[VERSION_NUMBER\],\s*\[([^\]]+)\]\)`)
		if m := re.FindStringSubmatch(string(vData)); len(m) > 1 {
			catalog.Version = strings.TrimSpace(m[1])
			discoveredSources["version.m4"] = true
		}
	}

	// Step 2: Parse sssd.api.conf and sssd.api.d/*.conf
	parseAPIConfigs(sourceDir, catalog, discoveredSources)

	// Step 3: Parse DocBook XML man pages (*.xml)
	parseDocBookXMLMan(sourceDir, catalog, discoveredSources)

	// Step 4: Parse roff man pages (*.5, *.5.gz)
	parseRoffManPages(sourceDir, catalog, discoveredSources)

	if len(catalog.Options) == 0 {
		return fmt.Errorf("no SSSD configuration options found in %q.\n"+
			"Ensure sssd.api.conf, sssd.api.d/*.conf or *.xml man pages are placed there", sourceDir)
	}

	// Apply known enum constraints verified against SSSD C code and man pages
	enrichKnownEnums(catalog)

	// Add options that are real but undocumented in this release's man pages,
	// so a correct sssd.conf is never reported as invalid.
	applyCuratedOptions(catalog)

	// Drop any entry that never acquired a section: it cannot be validated
	// meaningfully and would otherwise mask a genuine typo.
	pruneSectionlessOptions(catalog)

	// Finalize sorted option names and sorted section option lists
	for name := range catalog.Options {
		catalog.OptionNames = append(catalog.OptionNames, name)
	}
	sort.Strings(catalog.OptionNames)
	for sec, opts := range catalog.Sections {
		if len(opts) == 0 {
			// Never publish an empty section: it makes the reader believe a
			// section exists that documents no options.
			delete(catalog.Sections, sec)
			continue
		}
		sort.Strings(opts)
		catalog.Sections[sec] = opts
	}
	for src := range discoveredSources {
		catalog.Sources = append(catalog.Sources, src)
	}
	sort.Strings(catalog.Sources)

	// Write output files
	outDir := catalogOutputDir
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outDir, err)
	}

	jsonPath := filepath.Join(outDir, "catalog.json")
	jsonData, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize catalog JSON: %w", err)
	}
	if err := os.WriteFile(jsonPath, append(jsonData, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", jsonPath, err)
	}

	mdPath := filepath.Join(outDir, "catalog.md")
	mdContent := generateCatalogMarkdown(catalog)
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", mdPath, err)
	}

	fmt.Printf("[+] SSSD Catalog Generated Successfully\n")
	fmt.Printf("    Version:     %s\n", catalog.Version)
	fmt.Printf("    Sources:     %s\n", strings.Join(catalog.Sources, ", "))
	fmt.Printf("    Options:     %d unique options\n", len(catalog.Options))
	fmt.Printf("    Sections:    %d section definitions\n", len(catalog.Sections))
	fmt.Printf("    Output JSON: %s (%d bytes)\n", jsonPath, len(jsonData))
	fmt.Printf("    Output MD:   %s (%d bytes)\n", mdPath, len(mdContent))

	return nil
}

// parseAPIConfigs parses sssd.api.conf and sssd.api.d/*.conf
func parseAPIConfigs(sourceDir string, catalog *SssdCatalog, sources map[string]bool) {
	var confFiles []string
	_ = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "sssd.api.conf" || (strings.HasPrefix(name, "sssd-") && strings.HasSuffix(name, ".conf")) {
			confFiles = append(confFiles, path)
		}
		return nil
	})

	for _, file := range confFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		sources[filepath.Base(file)] = true
		curSection := ""

		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
				continue
			}
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				curSection = strings.Trim(line, "[]")
				if _, ok := catalog.Sections[curSection]; !ok {
					catalog.Sections[curSection] = []string{}
				}
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			optName := strings.ToLower(strings.TrimSpace(parts[0]))
			metaParts := strings.Split(parts[1], ",")
			rawType := "string"
			if len(metaParts) > 0 {
				rawType = mapAPIType(strings.TrimSpace(metaParts[0]))
			}
			defaultVal := ""
			if len(metaParts) >= 4 {
				defaultVal = strings.TrimSpace(metaParts[3])
			}

			opt := catalog.Options[optName]
			opt.Name = optName
			if opt.Type == "" {
				opt.Type = rawType
			}
			if opt.Default == "" && defaultVal != "" && defaultVal != "None" {
				opt.Default = defaultVal
			}
			if opt.Source == "" {
				opt.Source = filepath.Base(file)
			}
			if curSection != "" && !sliceContains(opt.Sections, curSection) {
				opt.Sections = append(opt.Sections, curSection)
				catalog.Sections[curSection] = append(catalog.Sections[curSection], optName)
			}
			catalog.Options[optName] = opt
		}
	}
}

func mapAPIType(t string) string {
	switch strings.ToLower(t) {
	case "bool", "boolean":
		return "bool"
	case "int", "integer":
		return "int"
	case "list":
		return "list"
	default:
		return "string"
	}
}

func sliceContains(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}

var reTerm = regexp.MustCompile(`<term>\s*([^<(]+?)(?:\s*\(([^)]+)\))?\s*</term>`)
var reDefault = regexp.MustCompile(`(?i)Default:\s*([^\n<]+)`)
var reXMLTag = regexp.MustCompile(`<[^>]+>`)
var reOptionName = regexp.MustCompile(`^[a-z][a-z0-9_]{2,}$`)

// isPlausibleOptionName rejects DocBook <term> values that are not SSSD option
// names (man page cross references, section headers, [domain/NAME] markers...).
func isPlausibleOptionName(name string) bool {
	return reOptionName.MatchString(name)
}

// extractXMLText flattens a DocBook fragment to single-spaced plain text,
// decoding the XML entities that appear in SSSD documentation.
func extractXMLText(chunk string, maxLen int) string {
	flat := reXMLTag.ReplaceAllString(chunk, " ")
	flat = strings.Join(strings.Fields(decodeXMLEntities(flat)), " ")
	if maxLen > 0 && len(flat) > maxLen {
		flat = strings.TrimSpace(flat[:maxLen]) + "..."
	}
	return flat
}

// decodeXMLEntities replaces the XML entity references used by SSSD man pages.
func decodeXMLEntities(s string) string {
	return strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&apos;", "'",
	).Replace(s)
}

// docFileSection maps a man page file name to the sssd.conf section it
// documents. Only man pages that document sssd.conf options are listed: other
// pages (pam_sss(8), sssd(8), sss_rpcidmapd(5), the InfoPipe attribute tables,
// the ad_modified_defaults/ldap_id_mapping include fragments) contain
// <term> elements that are NOT configuration options -- PAM return codes,
// signal names, LDAP attribute names -- and harvesting them produced ~88
// phantom entries that made real typos look valid.
//
// The provider pages document the options of the [domain/<provider>] section,
// which is where they are actually written in sssd.conf; an sssd.conf never
// contains a [provider/ad] section. The provider is only a qualifier here, so
// the validator compares the "domain" family, not the full name.
var docFileSection = map[string]string{
	"sssd-ad.5.xml":                "domain/ad",
	"sssd-ldap.5.xml":              "domain/ldap",
	"sssd-ipa.5.xml":               "domain/ipa",
	"sssd-krb5.5.xml":              "domain/krb5",
	"sssd-simple.5.xml":            "domain/simple",
	"sssd-kcm.8.xml":               "kcm",
	"sssd-sudo.5.xml":              "sudo",
	"sssd-session-recording.5.xml": "session_recording",
	"sssd-idp.5.xml":               "domain", // IDP options live in [domain/NAME]
}

// sectionRefID maps a DocBook refsect id inside sssd.conf.5.xml to the sssd.conf
// section (or the "*" wildcard for options usable in every section) that the
// refsection documents. The ids are stable, machine readable anchors in the
// upstream man page, so they are a far more reliable anchor than the prose
// title, which is localised.
var sectionRefID = map[string]string{
	"all-section-options":                  "*",
	"services-and-domains-section-options": "*",
	"services":                             "sssd",
	"general":                              "*",
	"NSS":                                  "nss",
	"PAM":                                  "pam",
	"SUDO":                                 "sudo",
	"AUTOFS":                               "autofs",
	"SSH":                                  "ssh",
	"PAC_RESPONDER":                        "pac",
	"SESSION_RECORDING":                    "session_recording",
	"domain-sections":                      "domain",
	"trusted-domains":                      "subdomain",
	"certmap":                              "certmap",
	"prompting_configuration":              "pam",
	"app_domains":                          "domain",
}

// nonOptionRefID lists refsects of sssd.conf.5.xml that carry examples and
// prose rather than option definitions. Their <term> elements are values and
// fragments ("env", "always", "false"), never option names.
var nonOptionRefID = map[string]bool{
	"file-format":           true,
	"config-snippets":       true,
	"vendor-dir":            true,
	"example":               true,
	"seealso":               true,
	"config-file-hierarchy": true,
}

var reRefSect = regexp.MustCompile(`<refsect([12])\b[^>]*\bid=['"]([^'"]+)['"]`)
var reVarlistStart = regexp.MustCompile(`<varlistentry`)

// sectionMarker is a resolved position in a man page at which the documented
// sssd.conf section changes ("" meaning "not an option section").
type sectionMarker struct {
	pos     int
	section string
}

// docSectionMarkers pre-scans a DocBook man page and returns the positions at
// which the documented section changes, so each <term> can be attributed to
// the refsection that encloses it. A refsect2 inherits the section of its
// enclosing refsect1 unless it has an explicit mapping.
func docSectionMarkers(text string) []sectionMarker {
	type raw struct {
		pos   int
		id    string
		level int
		end   int
	}
	var raws []raw
	for _, m := range reRefSect.FindAllStringSubmatchIndex(text, -1) {
		level := 1
		if text[m[2]:m[3]] == "2" {
			level = 2
		}
		id := text[m[4]:m[5]]
		// Find where this refsect ends so we can close the scope.
		end := len(text)
		closer := "</refsect" + strconv.Itoa(level) + ">"
		if i := strings.Index(text[m[1]:], closer); i >= 0 {
			end = m[1] + i
		}
		raws = append(raws, raw{pos: m[1], id: id, level: level, end: end})
	}

	// Walk the document maintaining the currently open refsect1 section.
	var markers []sectionMarker
	openLevel1 := ""
	openLevel1End := 0
	for _, r := range raws {
		// Close any refsect1 that ended before this one starts.
		for openLevel1 != "" && r.pos >= openLevel1End {
			markers = append(markers, sectionMarker{pos: r.pos, section: ""})
			openLevel1 = ""
		}
		sec, has := sectionRefID[r.id]
		if !has {
			sec = ""
			if r.level == 2 {
				sec = openLevel1 // inherit the enclosing refsect1
			}
		}
		if r.level == 1 {
			openLevel1 = sec
			openLevel1End = r.end
		}
		markers = append(markers, sectionMarker{pos: r.pos, section: sec})
	}
	return markers
}

// sectionAt returns the sssd.conf section documented at byte offset pos, or ""
// when the position is not inside an option-bearing refsection.
func sectionAt(markers []sectionMarker, pos int) string {
	cur := ""
	for _, mk := range markers {
		if mk.pos > pos {
			break
		}
		cur = mk.section
	}
	if nonOptionRefID == nil { // defensive: never nil in practice
		return cur
	}
	return cur
}

// termDepthEntry pairs a <term> match with the varlistentry nesting depth at
// which it occurs. SSSD documents enumerated options with a nested
// <variablelist> inside the option's own <varlistentry>:
//
//	<varlistentry>                     <- depth 1
//	  <term>pam_initgroups_scheme (string)</term>   <- an OPTION
//	  <listitem>...<variablelist>
//	    <varlistentry>                   <- depth 2
//	      <term>always</term>             <- an enumerated VALUE, not an option
//
// Harvesting every <term> unconditionally injected the enumerated values
// ("always", "env", "true", "false", "hybrid") into the option list, which
// made those nonsense names validate as real options.
type termDepthEntry struct {
	loc    []int
	depth  int
	name   string
	parent string
}

var reVarListOpen = regexp.MustCompile(`<varlistentry\b`)
var reVarListClose = regexp.MustCompile(`</varlistentry>`)

// indexedTerms returns every <term> match together with the variablelist
// nesting depth that encloses it. Depth 1 is the option itself; depth >= 2 is
// an enumerated value of that option, which is what lets the catalog learn
// valid values straight from the man page (for example the values of
// pac_check: no_check, pac_present, check_upn, ...).
func indexedTerms(text string) []termDepthEntry {
	type evt struct {
		pos   int
		delta int
	}
	var evts []evt
	for _, m := range reVarListOpen.FindAllStringIndex(text, -1) {
		evts = append(evts, evt{pos: m[0], delta: 1})
	}
	for _, m := range reVarListClose.FindAllStringIndex(text, -1) {
		evts = append(evts, evt{pos: m[0], delta: -1})
	}
	// Stable ordering by position so the depth can be replayed in document order.
	sort.SliceStable(evts, func(i, j int) bool { return evts[i].pos < evts[j].pos })

	var out []termDepthEntry
	depth := 0
	ei := 0
	for _, loc := range reTerm.FindAllStringSubmatchIndex(text, -1) {
		for ei < len(evts) && evts[ei].pos <= loc[0] {
			depth += evts[ei].delta
			ei++
		}
		// parent is the name of the enclosing depth-1 option, so a nested term
		// can be recorded as one of that option's valid values.
		parent := ""
		for i := len(out) - 1; i >= 0; i-- {
			if out[i].depth == 1 {
				parent = out[i].name
				break
			}
		}
		out = append(out, termDepthEntry{loc: loc, depth: depth, name: termName(text, loc), parent: parent})
	}
	return out
}

// termName extracts the bare option or value name from a <term> match,
// discarding the trailing "(type)" annotation.
func termName(text string, loc []int) string {
	raw := text[loc[2]:loc[3]]
	if i := strings.IndexByte(raw, '('); i >= 0 {
		raw = raw[:i]
	}
	return strings.ToLower(strings.TrimSpace(raw))
}

// parseDocBookXMLMan parses *.xml DocBook files from upstream/src/man
func parseDocBookXMLMan(sourceDir string, catalog *SssdCatalog, sources map[string]bool) {
	var xmlFiles []string
	_ = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".xml") && !strings.Contains(path, "po/") {
			xmlFiles = append(xmlFiles, path)
		}
		return nil
	})

	for _, file := range xmlFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		sources[filepath.Base(file)] = true
		text := string(content)
		base := filepath.Base(file)
		markers := docSectionMarkers(text)

		for _, te := range indexedTerms(text) {
			// Depth 1 is the option itself; anything deeper is an enumerated
			// value of that option, not an option name. Harvesting them as
			// options is what previously injected "always", "env", "true",
			// "false" and "hybrid" into the catalog.
			if te.depth >= 2 {
				val := te.name
				if te.parent == "" || !isPlausibleOptionName(val) {
					continue
				}
				if opt, ok := catalog.Options[te.parent]; ok && !containsFold(opt.Values, val) {
					opt.Values = append(opt.Values, val)
					catalog.Options[te.parent] = opt
				}
				continue
			}
			if te.depth != 1 {
				continue
			}
			loc := te.loc
			m := reTerm.FindStringSubmatch(text[loc[0]:loc[1]])
			if len(m) < 2 {
				continue
			}
			// Slice the enclosing varlistentry so "Default:" and the prose
			// description that follows the <term> can be captured.
			entry := text[loc[1]:]
			if end := strings.Index(entry, "</varlistentry>"); end >= 0 {
				entry = entry[:end]
			}
			defaultVal := ""
			if dm := reDefault.FindStringSubmatch(entry); len(dm) > 1 {
				defaultVal = strings.TrimSpace(dm[1])
			}
			doc := extractXMLText(entry, 240)

			rawNames := strings.Split(m[1], ",")
			rawType := "string"
			if len(m) > 2 && m[2] != "" {
				rawType = mapAPIType(m[2])
			}
			// Resolve the documented sssd.conf section: from the enclosing
			// refsection for sssd.conf(5), otherwise from the file name.
			inferredSec := sectionAt(markers, loc[0])
			if inferredSec == "" {
				inferredSec = docFileSection[base]
			}
			for _, rName := range rawNames {
				optName := strings.ToLower(strings.TrimSpace(rName))
				if !isPlausibleOptionName(optName) {
					continue
				}
				// An option that cannot be attributed to any sssd.conf
				// section is not a configuration option: it is a PAM return
				// code, a signal name or an LDAP attribute documented in the
				// same man pages. Dropping it is what stops a real typo
				// (gecos, env, sighup...) from being accepted as valid.
				if inferredSec == "" {
					continue
				}
				opt := catalog.Options[optName]
				opt.Name = optName
				if opt.Type == "" {
					opt.Type = rawType
				}
				if opt.Default == "" && defaultVal != "" {
					opt.Default = defaultVal
				}
				if opt.Doc == "" && doc != "" {
					opt.Doc = doc
				}
				if opt.Source == "" {
					opt.Source = base
				}
				if !sliceContains(opt.Sections, inferredSec) {
					opt.Sections = append(opt.Sections, inferredSec)
					catalog.Sections[inferredSec] = append(catalog.Sections[inferredSec], optName)
				}
				catalog.Options[optName] = opt
			}
		}
	}
}

// parseRoffManPages parses *.5 and *.5.gz manual pages
func parseRoffManPages(sourceDir string, catalog *SssdCatalog, sources map[string]bool) {
	var roffFiles []string
	_ = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.Contains(name, ".5") && (strings.HasSuffix(name, ".5") || strings.HasSuffix(name, ".5.gz")) {
			roffFiles = append(roffFiles, path)
		}
		return nil
	})

	reRoffTP := regexp.MustCompile(`(?m)^\.TP(?:\s+\d+)?\s*\n\\fB([a-zA-Z0-9_]+)\\fR(?:\s*\(([^)]+)\))?`)
	reRoffB := regexp.MustCompile(`(?m)^\.B\s+([a-zA-Z0-9_]+)`)

	for _, file := range roffFiles {
		var data []byte
		var err error
		if strings.HasSuffix(file, ".gz") {
			f, oErr := os.Open(file)
			if oErr != nil {
				continue
			}
			gz, gErr := gzip.NewReader(f)
			if gErr != nil {
				f.Close()
				continue
			}
			data, err = io.ReadAll(gz)
			gz.Close()
			f.Close()
		} else {
			data, err = os.ReadFile(file)
		}
		if err != nil {
			continue
		}
		sources[filepath.Base(file)] = true
		text := string(data)

		for _, m := range reRoffTP.FindAllStringSubmatch(text, -1) {
			name := strings.ToLower(strings.TrimSpace(m[1]))
			if name == "" {
				continue
			}
			opt := catalog.Options[name]
			opt.Name = name
			if opt.Type == "" && len(m) > 2 && m[2] != "" {
				opt.Type = mapAPIType(m[2])
			}
			if opt.Source == "" {
				opt.Source = filepath.Base(file)
			}
			// A type-bearing ".TP\fBname\fR (type)" entry is a strong signal
			// that name is a real option, so it is kept even when the page
			// name does not map to a known section: it is then treated as
			// valid everywhere rather than pruned.
			sec, known := roffFileSection[filepath.Base(file)]
			if !known {
				sec = "*"
			}
			if !sliceContains(opt.Sections, sec) {
				opt.Sections = append(opt.Sections, sec)
				if sec != "*" {
					catalog.Sections[sec] = append(catalog.Sections[sec], name)
				}
			}
			catalog.Options[name] = opt
		}

		// The bare ".B <word>" fallback may only enrich options that are
		// already known from a type-bearing ".TP\fBname\fR (type)" entry or
		// from the API definitions. Inventing options from bare .B words is
		// what previously injected phantom entries such as "always", "env"
		// and "true" into the catalog.
		for _, m := range reRoffB.FindAllStringSubmatch(text, -1) {
			name := strings.ToLower(strings.TrimSpace(m[1]))
			opt, known := catalog.Options[name]
			if !known || len(name) <= 2 || strings.Contains(name, "-") {
				continue
			}
			opt.Name = name
			catalog.Options[name] = opt
		}
	}
}

// roffFileSection maps a compiled man page name to the sssd.conf section it
// documents, for the roff fallback path (the API and DocBook passes are
// authoritative when available).
var roffFileSection = map[string]string{
	"sssd.conf.5":                 "sssd",
	"sssd.conf.5.gz":              "sssd",
	"sssd-ad.5":                   "provider/ad",
	"sssd-ad.5.gz":                "provider/ad",
	"sssd-ipa.5":                  "provider/ipa",
	"sssd-ipa.5.gz":               "provider/ipa",
	"sssd-ldap.5":                 "provider/ldap",
	"sssd-ldap.5.gz":              "provider/ldap",
	"sssd-krb5.5":                 "provider/krb5",
	"sssd-krb5.5.gz":              "provider/krb5",
	"sssd-simple.5":               "provider/simple",
	"sssd-simple.5.gz":            "provider/simple",
	"sssd-session-recording.5":    "session_recording",
	"sssd-session-recording.5.gz": "session_recording",
}

// curatedOptions lists options that SSSD genuinely supports but that the
// upstream man pages of some releases fail to document. Without this table the
// validator reports a documented, working option as "unknown", which is the
// worst kind of false positive: it sends the reader to fix a correct config.
//
// Each entry is a deliberate, verifiable claim: the option must exist in the
// SSSD source tree (src/confdb/confdb.h / the provider sources) and be
// accepted by the running daemon. Entries are tagged Source "curated" so the
// report can distinguish curated knowledge from man-page knowledge.
var curatedOptions = map[string]OptionMeta{
	// config_file_version (integer) guards sssd.conf against being edited
	// with a newer release than the running daemon. It is read by
	// confdb_init and shipped in distro sssd.conf(5) pages, but is absent
	// from the sssd.conf.5.xml of several upstream releases.
	"config_file_version": {
		Name:     "config_file_version",
		Type:     "int",
		Sections: []string{"sssd"},
		Doc:      "Declares the sssd.conf file format version. The daemon refuses to start if the file was written for a newer release than the running one.",
		Source:   "curated",
	},
}

// applyCuratedOptions merges curatedOptions into the catalog. Documentation
// found in the man pages always wins; curated entries only fill genuine gaps.
func applyCuratedOptions(catalog *SssdCatalog) {
	for name, meta := range curatedOptions {
		existing, present := catalog.Options[name]
		if !present {
			opt := meta
			opt.Name = name
			catalog.Options[name] = opt
			for _, sec := range meta.Sections {
				if !sliceContains(catalog.Sections[sec], name) {
					catalog.Sections[sec] = append(catalog.Sections[sec], name)
				}
			}
			continue
		}
		// Documented but section-less: keep the documentation, add the section.
		if len(existing.Sections) == 0 && len(meta.Sections) > 0 {
			existing.Sections = meta.Sections
			catalog.Options[name] = existing
			for _, sec := range meta.Sections {
				if !sliceContains(catalog.Sections[sec], name) {
					catalog.Sections[sec] = append(catalog.Sections[sec], name)
				}
			}
		}
	}
}

// pruneSectionlessOptions removes catalog entries that belong to no sssd.conf
// section. Such entries are extraction artefacts (PAM return codes from
// pam_sss(8), signal names from sssd(8), LDAP attribute names from the
// InfoPipe and mapping tables). Keeping them would let a genuine typo such as
// "gecos" or "env" pass validation as a known option.
func pruneSectionlessOptions(catalog *SssdCatalog) {
	for name, opt := range catalog.Options {
		if len(opt.Sections) == 0 {
			delete(catalog.Options, name)
		}
	}
}

// enrichKnownEnums attaches curated valid values for options with discrete enumerations
func enrichKnownEnums(catalog *SssdCatalog) {
	enums := map[string][]string{
		"ad_gpo_access_control":    {"disabled", "enforcing", "permissive"},
		"ad_gpo_default_right":     {"interactive", "remote_interactive", "network", "batch", "service", "permit", "deny"},
		"ldap_tls_reqcert":         {"never", "allow", "try", "demand", "hard"},
		"ldap_schema":              {"rfc2307", "rfc2307bis", "ipa", "ad"},
		"ldap_deref":               {"never", "searching", "finding", "always"},
		"ldap_pwmodify_mode":       {"ldap_modify", "exop"},
		"krb5_use_fast":            {"never", "try", "demand"},
		"certificate_verification": {"no_verification", "ocsp_dgst", "ocsp", "crl"},
		"access_provider":          {"permit", "deny", "ad", "ipa", "ldap", "simple", "proxy"},
		"id_provider":              {"ad", "ipa", "ldap", "proxy", "simple", "files"},
		"auth_provider":            {"none", "ad", "ipa", "ldap", "krb5", "proxy"},
		"chpass_provider":          {"none", "ad", "ipa", "ldap", "krb5", "proxy"},
		"sudo_provider":            {"none", "ad", "ipa", "ldap"},
		"autofs_provider":          {"none", "ad", "ipa", "ldap"},
		"case_sensitive":           {"true", "false", "preserving"},
		"auto_private_groups":      {"true", "false", "hybrid"},
		"local_auth_policy":        {"match", "only", "enable", "disable"},
	}

	for name, values := range enums {
		if opt, ok := catalog.Options[name]; ok {
			opt.Values = values
			catalog.Options[name] = opt
		}
	}
}

// generateCatalogMarkdown builds a readable reference guide of all SSSD options
func generateCatalogMarkdown(catalog *SssdCatalog) string {
	var sb strings.Builder
	sb.WriteString("# SSSD Configuration Options Reference Catalog\n\n")
	sb.WriteString(fmt.Sprintf("Auto-generated from SSSD %s on %s.\n", catalog.Version, catalog.Generated))
	sb.WriteString(fmt.Sprintf("Sources: `%s`\n\n", strings.Join(catalog.Sources, "`, `")))
	sb.WriteString(fmt.Sprintf("Total Options: **%d** across **%d** sections.\n\n", len(catalog.Options), len(catalog.Sections)))

	// Section overview
	sb.WriteString("## Sections\n\n")
	var secNames []string
	for sec := range catalog.Sections {
		secNames = append(secNames, sec)
	}
	sort.Strings(secNames)
	for _, sec := range secNames {
		opts := catalog.Sections[sec]
		sb.WriteString(fmt.Sprintf("- [`[%s]`](#section-%s) (%d options)\n", sec, strings.ReplaceAll(sec, "/", "-"), len(opts)))
	}
	sb.WriteString("\n---\n\n")

	// Options by section
	for _, sec := range secNames {
		opts := catalog.Sections[sec]
		sb.WriteString(fmt.Sprintf("## Section `[%s]`\n\n", sec))
		sb.WriteString("| Option | Type | Default | Allowed Values |\n")
		sb.WriteString("|--------|------|---------|----------------|\n")
		for _, optName := range opts {
			opt := catalog.Options[optName]
			defVal := opt.Default
			if defVal == "" {
				defVal = "-"
			}
			valStr := "-"
			if len(opt.Values) > 0 {
				valStr = strings.Join(opt.Values, ", ")
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s |\n", opt.Name, opt.Type, defVal, valStr))
		}
		sb.WriteString("\n")
	}

	// Alphabetical all options table
	sb.WriteString("## All Options (Alphabetical)\n\n")
	sb.WriteString("| Option | Type | Sections | Default | Allowed Values |\n")
	sb.WriteString("|--------|------|----------|---------|----------------|\n")
	for _, name := range catalog.OptionNames {
		opt := catalog.Options[name]
		defVal := opt.Default
		if defVal == "" {
			defVal = "-"
		}
		valStr := "-"
		if len(opt.Values) > 0 {
			valStr = strings.Join(opt.Values, ", ")
		}
		secStr := "-"
		if len(opt.Sections) > 0 {
			secStr = strings.Join(opt.Sections, ", ")
		}
		sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | %s |\n", opt.Name, opt.Type, secStr, defVal, valStr))
	}

	return sb.String()
}
