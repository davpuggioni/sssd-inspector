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

	// Finalize sorted option names and sorted section option lists
	for name := range catalog.Options {
		catalog.OptionNames = append(catalog.OptionNames, name)
	}
	sort.Strings(catalog.OptionNames)
	for sec, opts := range catalog.Sections {
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

		// Infer section from filename if obvious
		inferredSec := ""
		base := filepath.Base(file)
		switch {
		case strings.HasPrefix(base, "sssd-ad"):
			inferredSec = "provider/ad"
		case strings.HasPrefix(base, "sssd-ldap"):
			inferredSec = "provider/ldap"
		case strings.HasPrefix(base, "sssd-ipa"):
			inferredSec = "provider/ipa"
		case strings.HasPrefix(base, "sssd-krb5"):
			inferredSec = "provider/krb5"
		case strings.HasPrefix(base, "sssd-simple"):
			inferredSec = "provider/simple"
		case strings.HasPrefix(base, "sssd-kcm"):
			inferredSec = "kcm"
		case strings.HasPrefix(base, "sssd-sudo"):
			inferredSec = "sudo"
		case strings.HasPrefix(base, "sssd-session-recording"):
			inferredSec = "session_recording"
		}

		idx := reTerm.FindAllStringSubmatchIndex(text, -1)
		for _, loc := range idx {
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
			for _, rName := range rawNames {
				optName := strings.ToLower(strings.TrimSpace(rName))
				if !isPlausibleOptionName(optName) {
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
				if inferredSec != "" && !sliceContains(opt.Sections, inferredSec) {
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
			catalog.Options[name] = opt
		}

		for _, m := range reRoffB.FindAllStringSubmatch(text, -1) {
			name := strings.ToLower(strings.TrimSpace(m[1]))
			if len(name) > 2 && !strings.Contains(name, "-") {
				opt := catalog.Options[name]
				opt.Name = name
				catalog.Options[name] = opt
			}
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
