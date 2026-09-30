# SSSD Inspector 🔍

**SSSD Inspector** is a native diagnostic utility designed to parse supportconfig archives generated on SLES or OpenSUSE and identify complex identity management failures. Unlike generic log viewers, this tool uses a specialized pattern-matching engine derived directly from the SSSD C source code to provide human-readable explanations for cryptic error messages.

---

## ✨ Features

### License

This program is released under the terms of the **GNU General Public License
v3.0 or later** (GPL-3.0-or-later). See [`LICENSE`](LICENSE) for the full text.
The license note is also printed at the bottom of every generated report.

### Correlation Graph & Diff Mode (Phase 4)
- **Interactive SVG force-directed graph** — entities (domain/realm/hostname/DNS) → findings → evidence, rendered client-side with zero external dependencies
- **Hover/click interactivity** — highlight connected nodes, drill-down panel with source path + line + evidence, severity filters, draggable nodes
- **Diff mode (`-compare A:B`)** — differential analysis between two supportconfigs (paths separated by `:`): common findings, only-in-A, only-in-B, health-score delta; prints to stdout and optionally writes `compare_report.json` (with `-json`)
- **PII-safe** — graph labels, values, and evidence are anonymized alongside the rest of the report

### Executive Summary & Scoring (Phase 3+)
- **Health Score (0-100)** — weighted penalty model over all findings (critical ×25, error ×12, problem ×4, warning ×2, log-error ×3)
- **Dominant root-cause inference** — the most weighted finding category drives the triage headline
- **Temporal Clusters** — sliding-window (300 s) bursts of the same diagnostic event = retry loops, timeouts, offline flapping
- **KB Suggestions (TF-IDF)** — log lines that match no known pattern are fuzzy-correlated (cosine similarity) with the Knowledge Base corpus
- **JSON export** — full machine-readable report (summary + provenance-aware findings + clusters + suggestions)

### Advanced AD/Kerberos/DNS Checks (Phase 2)
- krb5.conf `allow_weak_crypto` and RC4-only enctypes analysis
- `ldap_id_use_start_tls` incompatibility with the AD provider
- `ad_gpo_access_control` value validation
- `ad_site` + `ad_enable_dns_sites = false` contradiction
- `ad_machine_account_password_renewal_opts` format validation
- `ad_hostname` vs system hostname consistency
- Loopback-only DNS resolver detection
- Missing `[domain/]` sections, missing `nss`/`pam` responders, overlapping `ldap_idmap` ranges

### Data-Driven Rules (Phase 5)
- Optional `rules.yaml` / `rules/*.yaml`: add site-specific detectors **without recompiling** (see `rules/example.yaml`)

### Core Analysis Engine
- **Single-Pass Log Scanning** — Scans the relevant log files (`sssd.txt`, `messages`, `messages.txt` — see `analyzer_singlepass.go`) in one pass, extracting errors, warnings, timeline events, and KB article evidence simultaneously
- **220+ Error Patterns** — Pattern database of SSSD error signatures mapped to human-readable descriptions (plus test-only auxiliary entries), covering:
  - **Kerberos** — Clock skew, encryption type mismatches, KDC unreachable, FAST tunnel failures
  - **LDAP/AD** — TLS handshake failures, SASL bind errors, USN rollbacks, LDAP size limits
  - **PAM/NSS** — Offline authentication blocks, shell vetoes, negative cache rejections
  - **SSSD Service** — Watchdog terminations, configuration errors, file permission issues, cache corruption
  - **IPA/FreeIPA** — HBAC rule evaluation, SELinux user mapping, cross-forest AD trusts
  - **OAuth2/OIDC** — Identity Provider configuration validation
  - **And many more** — Sudo, SSH, InfoPipe, PAC, proxy providers
- **Knowledge Base Integration** — 16 bundled JSON articles (TIDs) correlated with log evidence for precise troubleshooting
- **PII Anonymization** — Built-in redaction of IPv4/IPv6, MAC, email, domain/realm, hostname (FQDN, short and syslog forms) for safe report sharing; opt-in via `-anonymize` (CLI) or the GUI checkbox; provenance fields (`source_key`/`source_path`) are redacted while keeping the actionable `domain/` section structure

### Raw SSSD Log Mode (`-logdir`)
- **No supportconfig needed** — analyze `*.log` files directly (e.g. `/var/log/sssd/sssd_<domain>.log`, `ldap_child.log`, `krb5_child.log`), including rotated variants (`*.log.N`, `*.log-<date>`)
- **Same engines as supportconfig mode** — single-pass pattern scan, timeline, temporal clusters, KB suggestions, root-cause sequence correlation, executive summary, correlation graph
- **Honest about the unknown** — fields that can only come from a supportconfig are reported as `N/A (raw log mode)` instead of being guessed; supportconfig-only findings (e.g. *sssd.conf not found*) are never emitted
- **PII-safe by construction** — with `-anonymize`, the domain/realm/hostname are harvested from the log lines themselves (`Domain [...]`, principals, `ldap://` URIs, syslog prefixes) and redacted, since there is no `sssd.conf` to read them from

### Performance Optimizations
- **Regex Cache** — Thread-safe compilation cache, compiles each regex pattern once per application lifetime
- **File Cache** — Per-run cache of file contents: each file is read only once per analysis
- **Scanner Pool** — Reusable buffer allocations via `sync.Pool` to reduce GC pressure
- **Context-Aware Scanning** — Timeout-based cancellation prevents hangs on corrupted or NFS-mounted files
- **Parallel-Analysis Helpers** — Optional worker-pool based phase helper functions (`analyzePhase1Parallel`/`analyzePhase2Parallel`) exist for large archives; the default CLI path runs phases sequentially
- **Multi-Core File Processing** — Parallel file scanning and batch processing (`utils_parallel.go`) for large supportconfig archives

### Output Formats
- **Text Report** — Formatted terminal output with problem severity indicators
- **HTML Report** — Rich, styled HTML with collapsible sections and visual hierarchy
- **PDF Export** — Print-ready PDF generation via browser (GUI mode)

### User Interface
- **Native GUI** — Built with Wails (Go + WebKit), providing a fast, lightweight desktop experience
- **Dark Mode** — Automatic system theme detection with manual toggle
- **Drag & Drop** — Supportconfig files can be dragged directly into the application
- **Progress Tracking** — Real-time progress bar with detailed status messages during analysis
- **Export Controls** — One-click export to TXT and PDF
- **Zoom Controls** — Adjust report text size for comfortable reading
- **Keyboard Shortcuts** — Ctrl+O (open), Ctrl+Enter (analyze), Ctrl+P (export PDF), Ctrl+S (export TXT)

---

## 📋 Requirements

### Runtime
- **Linux** (SLES 12+, OpenSUSE Leap 15+, or any modern Linux distribution)
- Supportconfig archive (.txz or .tar.xz) from SLES or OpenSUSE

### Build Prerequisites
| Dependency | Version | Purpose |
|-----------|---------|---------|
| Go | 1.23+ | Backend compilation |
| Node.js | 18+ | Frontend asset build |
| npm | 9+ | Frontend dependency management |
| Wails CLI | 2.12+ | GUI framework (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`) |

---

## 🚀 Installation

### Option 1: Download Pre-built Binary

Download the latest release from the [Releases](https://github.com/davpuggioni/sssd-inspector/releases) page:

```bash
# CLI version
chmod +x sssd-inspector-cli
./sssd-inspector-cli --help

# GUI version
chmod +x sssd-inspector-gui
./sssd-inspector-gui
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/davpuggioni/sssd-inspector.git
cd sssd-inspector

# Install dependencies
npm install --prefix frontend

# Build the CLI binary
wails build -tags cli -platform linux/amd64 -ldflags "-w -s" -clean

# Build the GUI binary
wails build -platform linux/amd64 -ldflags "-w -s" -clean
```

---

## 💻 CLI Usage

### Basic Analysis

```bash
# Analyze a supportconfig archive
sssd-inspector /path/to/supportconfig.txz

# Analyze an already-extracted directory
sssd-inspector /path/to/supportconfig-directory/

# Generate both TXT and HTML reports
sssd-inspector /path/to/supportconfig.txz -txt -html
```

### Options

| Flag | Description |
|------|-------------|
| `-v, --version` | Print program version |
| `-analyze <path>` | Path to supportconfig directory or archive |
| `-logdir <path>` | Analyze raw SSSD logs (no supportconfig): a directory of `*.log` files or a single log file, e.g. `/var/log/sssd`. Rotated logs (`*.log.N`, `*.log-<date>`) are included; `*.log.gz` are not. Takes precedence over `-analyze` |
| `-compare <A:B>` | Differential analysis between two supportconfigs (CLI-only flag; paths separated by a single `:`) |
| `-txt` | Generate a TXT report (default: both formats) |
| `-html` | Generate an HTML report (default: both formats) |
| `-json` | Generate a structured JSON report (full findings + graph + clusters) |
| `-anonymize` | Redact PII (IPs, domains, hostnames, emails) from the report |
| `-gen-catalog <dir>` | Generate sssd_catalog/catalog.{json,md} from upstream man pages and API definitions |
| `-definitions-info` | List every definition search path (rules and KB articles), what was found there, and whether it is writable |
| `-validate-rules` | Validate the installed definition files and exit non-zero if any rule or KB article would be skipped |
| `-rules-test <path>` | Dry-run the installed analysis rules against a supportconfig directory or archive and show which rules would fire, with the evidence line |

### Analysis definitions (rules and KB articles)

Detectors can be added without recompiling the inspector: the binary ships with
built-in checks plus an embedded knowledge base, and loads extra definitions
from the filesystem. Every location is searched in this order (for KB articles
a later directory overrides an earlier one for the same article):

| Kind | Locations (in order) |
|------|----------------------|
| Rules | `<exe dir>/rules.yaml`, `<exe dir>/rules/*.yaml`, `./rules.yaml`, `./rules/*.yaml`, `/etc/sssd-inspector/rules.yaml`, `/etc/sssd-inspector/rules/`, `~/.sssd-inspector/rules.yaml`, `~/.sssd-inspector/rules/` |
| KB articles | `./kb_articles/*.json`, `<exe dir>/kb_articles/*.json`, `/etc/sssd-inspector/kb_articles/`, `~/.sssd-inspector/kb_articles/`, the embedded corpus |

`~/.sssd-inspector` is the per-user root (no privileges needed);
`/etc/sssd-inspector` is the system-wide one (root only). The GUI's Definitions
Studio creates them on demand and can open them in the file manager.

Three commands make the search paths and the definitions verifiable — the GUI
Studio calls the same Go service, so the answers cannot differ:

```bash
# 1. Where does the inspector look, what did it find there, is it writable?
sssd-inspector -definitions-info

# 2. Are the installed definitions valid? Exit code 1 = something would be skipped.
sssd-inspector -validate-rules

# 3. Would my new rule fire on this supportconfig? Dry-run: no report is written.
sssd-inspector -rules-test /tmp/supportconfig-abc123.txz
```

A definition that cannot be loaded is never ignored silently: it is reported
with its file and line, both in the report (the `diagnostics` section of
TXT/HTML/JSON) and on stderr. The same validation covers the mistakes that
silently produce **zero rules**: a missing top-level `rules:` header, a
different top-level key, a rule without a name, or duplicate rule names.

### Examples

```bash
# Full analysis with PII redaction and all report formats
sssd-inspector /tmp/supportconfig-abc123.txz -txt -html -anonymize

# Export structured JSON (includes findings, graph, clusters, suggestions)
sssd-inspector /tmp/supportconfig-abc123.txz -json -anonymize

# Diff two supportconfig (before/after fix) — delta to stdout, plus compare_report.json with -json
sssd-inspector -compare /tmp/sc-before:/tmp/sc-after -anonymize

# Raw SSSD logs, no supportconfig (same format flags as -analyze)
sssd-inspector -logdir /var/log/sssd -txt -html -anonymize

# Quick analysis with default TXT output
sssd-inspector /var/log/supportconfig/

# Version check
sssd-inspector -v
```

### Development: tests and regression guards

```bash
# Run the full test suite (default/hybrid build)
go test ./...

# Run the suite for the static CLI binary build
go test -tags cli ./...

# Race detector (both build modes) and vet
go test -race ./... && go test -race -tags cli ./...
go vet ./... && go vet -tags cli ./...
```

The CLI surface is guarded against silent flag loss: every user-visible
flag is registered once in `cli_flags.go` (`registerCLIFlags`), pinned by a
hardcoded contract in `cli_flags_test.go`, and cross-checked against the
Options table above — removing a documented flag without updating the code,
the tests, this table and `docs/CHANGES.md` fails the suite.

Test data must be generic: fixtures, test comments, CI smoke data and docs
may only use reserved names (RFC 2606/6761 domains such as `example.test` or
`example.com`, documentation IPs such as `192.0.2.10`, invented hostnames such
as `testhost01`) — never a real company domain or a hostname taken from an
analysed machine. `test_fixture_hygiene_test.go` enforces the ban list.

### Offline sssd.conf option catalog (`-gen-catalog`)

The binary embeds a catalog of every SSSD configuration option
(`sssd_catalog/catalog.json`, generated ahead of time). The catalog backs the
configuration validator, so a typo like `ldap_url = ldap://dc01` is reported as
`Did you mean 'ldap_uri'?`, a bad value like `ldap_schema = rfc2309` is
reported together with the documented value list, and an option used in a
section that never accepts it (for example an `[nss]` option written under
`[domain/...]`) is reported as a section mistake rather than as a typo. The
whole check runs offline — nothing is fetched at runtime.

Every finding states where the knowledge came from, for example
`[checked against the SSSD 2.14.0 option catalog: 519 options from 54
documentation sources]`. The same line is published in the HTML and text
reports, so a configuration claim is never presented without the documentation
release it was checked against.

Regenerate the catalog whenever the upstream SSSD release changes:

```bash
# 1. Drop the upstream sources into the (git-ignored) drop-zone.
#    Expected layout:
#      upstream/sssd.api.conf            # SSSD option API definitions
#      upstream/sssd.api.d/*.conf        # per-provider option definitions
#      upstream/src/man/*.xml            # DocBook man pages
#      upstream/src/man/include/*.xml    # shared man-page includes
#      upstream/version.m4               # upstream version (fill the VERSION string)
#    See upstream/README.md for the exact files to copy.

# 2. Regenerate both catalog.json and catalog.md.
go run . -gen-catalog upstream
# or, for the hybrid binary:
go run -tags cli . -gen-catalog upstream

# 3. Commit the result. catalog.md is the reviewable diff; catalog.json is
#    what the binary embeds.
git add sssd_catalog/catalog.json sssd_catalog/catalog.md
```

The generator is tolerant by design: it parses the API `*.conf` files
(`option = type, subtype, mandatory[, default]`) first for the authoritative
option list and types, then the DocBook XML (`<term>` + `Default:` lines) for
documentation and defaults, and finally falls back to compiled roff man pages
(`sssd*.5`, `sssd*.5.gz`) when XML sources are unavailable. A curated table in
`catalog_gen.go` (`enrichKnownEnums`) fills in the value lists that upstream
documents only as prose.

Every option is tagged with the ground-truth layer that produced it
(`OptionMeta.Confidence`): `"api"` for `sssd.api.conf` / `sssd.api.d/*.conf`
(authoritative for existence, types and defaults — written from the C code),
`"man"` for DocBook XML and roff pages (authoritative for documentation,
defaults and enumerations only), and `"curated"` for the hand-verified table
below. `catalog.md` renders the layer next to each option
(`sssd-ad.conf (api)`), and findings carry it as `confidence`, so a claim
backed only by prose is visibly weaker than one backed by the API definitions.

Only options that can be attributed to an `sssd.conf` section are published.
A man page contains plenty of `<term>` elements that are **not** options — PAM
return codes in `pam_sss(8)`, signal names in `sssd(8)`, LDAP attribute names in
the InfoPipe tables — and a nested `<variablelist>` inside an option's own entry
lists its *enumerated values* (`always`, `true`, `no_session`, ...). Harvesting
those made nonsense names validate as real options, so they are now filtered:
nested terms become the parent option's allowed values, and entries that cannot
be placed in a section are dropped.

A second curated table (`curatedOptions`) covers options that SSSD genuinely
supports but that some releases fail to document. `config_file_version` is the
current example: it is read by the daemon and shipped in distro `sssd.conf(5)`
pages, but is absent from the upstream `sssd.conf.5.xml`. Such entries are
tagged `"source": "curated"` in `catalog.json` to keep curated knowledge
distinguishable from man-page knowledge. When adding one, verify the option in
the SSSD source tree first — a false "unknown parameter" sends the reader to
break a correct configuration.

### Severity policy and semantic rules

Severity is decided centrally in `config_severity.go` (`catalogSeverity`),
not at each call site: every finding passes through one allowlist
(`errorCapableRules`) and anything not on it is capped at `SevWarning` — SSSD
ignores unknown or misplaced options and keeps running, so only provably fatal
conditions (no domains, missing/invalid `id_provider`, overlapping idmap
ranges, unresolvable KRB5 realm, ...) may escalate. Pre-existing checks keep
their historical severity; only new rules start capped.

Every finding carries an evidence contract (`rule_id`, `confidence`,
`doc_ref`), rendered in the HTML report and the JSON output, so a false
positive is traceable to the rule and documentation that produced it.

`config_semantics.go` adds provider-aware cross-option rules the syntactic
catalog cannot express: `ldap_uri` / `ldap_search_base` on a domain whose
`id_provider` never speaks LDAP, bind credentials with an `auth_provider`
that never binds, and LDAP search tuning (`ldap_schema`, `ldap_tls_*`, ...) on
non-LDAP providers. Adding a rule means appending to `semanticRules` with a
dedicated test.

Both outputs are committed because they serve different audiences:

- `sssd_catalog/catalog.json` — embedded with `go:embed` and consumed by the
  validator (`config_catalog.go`).
- `sssd_catalog/catalog.md` — human-readable reference that makes catalog
  changes reviewable in a pull request.

If `sssd_catalog/catalog.json` is missing or malformed the validator degrades
silently (no findings) instead of failing the analysis run;
`TestEmbeddedCatalogLoads` fails the build instead.

#### Overriding the catalog per machine

The embedded catalog is frozen at build time, which is right for an air-gapped
host and wrong for an engineer validating a newer distro. A generated
`catalog.json` dropped in a definitions root overrides it:

| Scope | Location | Precedence |
|-------|----------|------------|
| user | `~/.sssd-inspector/catalog.json` | highest (per-user) |
| system | `/etc/sssd-inspector/catalog.json` | per-system (needs root) |
| embedded | in the binary | fallback |

The first candidate that exists **and parses** wins; the report's
`catalog_provenance` states the release the claims are based on and appends
`(override in effect: <path>)`. An override that is unreadable, malformed, or has
an empty `options` map is **reported as a diagnostic and skipped** — validation
continues against the embedded catalog rather than silently disappearing.
`-definitions-info` lists the candidate paths and names the catalog in effect;
the Definitions Studio shows the same and can **install** a generated
`catalog.json` for you (`InstallCatalog`: refused if unusable, previous copy kept
as `catalog.json.bak`), with "Open user folder" for the manual route. The cache
is keyed by content, so an edited catalog takes effect on the next analysis
without restarting the application.

### Output

The CLI prints a human-readable summary to stdout plus `[Progress N%]` lines,
and writes `<basename>_report.{txt,html,json}` next to the working directory
(only for the formats requested via `-txt`/`-html`/`-json`):

1. **Console output** — Human-readable summary of all findings directly in the terminal
2. **TXT report** — `<basename>_report.txt` (if `-txt` is specified)
3. **HTML report** — `<basename>_report.html` (if `-html` is specified)
4. **JSON report** — `<basename>_report.json` (if `-json` is specified)
5. **Compare JSON** — `compare_report.json` (only for `-compare`, and only with `-json`)

---

## 🖥️ GUI Usage

### Building and Running

```bash
# Build the GUI
wails build -platform linux/amd64 -tags webkit2_41 -ldflags "-w -s" -clean 2>&1

# Run (or launch the binary)
./sssd-inspector
```

### Workflow

1. **Open the application** — A native window appears with the SSSD Inspector interface
2. **Select a file** — Click "Browse..." or drag a supportconfig.txz (or an already-extracted supportconfig directory) into the input field
3. **Configure options** — Check "Anonymize PII" if you need to redact sensitive information
4. **Start analysis** — Click "Analyze" (or `Ctrl+Enter`) and watch the progress bar fill in real-time
5. **Review results** — Browse through system info, problems, warnings, error details, timeline, KB articles (each with the log lines that matched) and the correlation graph
6. **Export** — Click "Export PDF" (print-ready via browser dialog), "Export TXT" or "Export JSON" for the report file

### Definitions Studio (`Ctrl+2`)

The second tab answers the three questions a custom definition raises, through
the same Go service the `-definitions-info`, `-validate-rules` and
`-rules-test` flags use:

- **Discovery inventory** — every search location in discovery order, with what
  is there, whether you can write there, the rules currently loaded and any
  input that was skipped. "Open user folder" / "Open system folder" take you to
  where definition files go, and **Reload** re-reads the inventory, so a file
  you edited outside the app shows up without restarting. A reload that fails
  says so and keeps the previous view rather than showing an empty one.
- **Rule editor** — load the `rules.yaml` of a scope, edit, `Validate` and
  `Save`. A save is refused when the analysis would skip the document, and the
  verdict is shown inline with file and line. The system scope needs root.
- **Dry-run** — point it at a supportconfig and see which loaded rules *would*
  fire, and on which evidence line, without writing anything. The target is
  prefilled with the archive the analysis view is holding.
- **SSSD option catalog** — which release the configuration is validated
  against, and **Install catalog** to replace it with a generated
  `catalog.json` (user scope, or system with root). An unusable file is
  refused, the previous catalog is kept as `catalog.json.bak`, and the panel
  reloads so it cannot keep claiming the old release.

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+1` | Switch to the analysis view |
| `Ctrl+2` | Switch to the Definitions Studio |
| `Ctrl+O` | Open file browser |
| `Ctrl+Enter` | Start analysis |
| `Ctrl+P` | Export PDF report |
| `Ctrl+S` | Export TXT report |
| `Ctrl+J` | Export JSON report |

---

## 📊 Report Sections

### System Information
- OS release and kernel version
- Hardware manufacturer and model
- Virtualization platform (VMware, KVM, Xen, etc.)
- SCC registration status

### Authentication Services
- SSSD installation status and package versions
- SSSD, Winbind, and NSCD service states
- NSSwitch configuration validity
- PAM module presence

### AD / Kerberos Integration
- DNS nameserver configuration and connectivity
- Time synchronization service status (chronyd/ntpd)
- Kerberos realm detection and keytab verification
- krb5.conf analysis (encryption types, domain realm mapping)

### SSSD Configuration Deep Dive
- AD provider detection
- Enumerate flag warning (performance risk)
- FQDN usage verification
- File permission analysis (sssd.conf ownership and mode)
- /var/lib/sss/ ownership verification
- Configuration conflict detection (duplicate parameters)
- Cross-source AD correlation — realm vs DNS vs hostname reconciliation (Phase B engine)

### Log Analysis
- **Critical Errors** — Kerberos failures, LDAP connection issues, TLS handshake errors, cache corruption
- **Error Timeline** — Chronological view of all error events with timestamps
- **MAC Security Denials** — AppArmor/SELinux denials related to SSSD

### Knowledge Base Articles
- TID articles matched against actual log evidence
- Direct links to SUSE Technical Information Database
- Relevant configuration suggestions

---

## 🏗️ Architecture

```
sssd-inspector/
├── main_cli.go              # CLI entry point (build tag: cli)
├── main_gui.go              # GUI entry point (Wails framework)
├── app.go                   # Application structure & Wails bindings
├── shared.go                # Shared CLI/GUI logic & configuration
│
├── analyzer_core.go         # Main analysis orchestrator (analyzeData)
├── analyzer_singlepass.go   # Single-pass log scanner (Aho-Corasick engine)
├── analyzer_parallel.go     # Parallel analysis phases
├── analyzer_system.go       # OS, hardware, services analysis
├── analyzer_auth.go         # DNS, Kerberos, PAM, NSS analysis
├── analyzer_logs.go         # SSSD log pattern definitions (220+ patterns)
├── config_parser.go         # sssd.conf INI parser + AD typed option validator (Phase A)
├── analyzer_correlate.go    # Cross-source correlation/reconciliation engine (Phase B)
│
├── utils.go                 # File scanning, section extraction
├── utils_parallel.go        # Parallel file/batch processing
├── cache.go                 # RegexCache, FileCache, Scanner Pool
├── ahocorasick.go           # Aho-Corasick multi-pattern matcher (O(n) literal scanning)
├── interfaces.go            # Core interfaces for DI
│
├── loaders.go               # Secure archive extraction (.txz)
├── kb.go                    # Knowledge base article matching
├── report.go                # Text & HTML report generation
├── types.go                 # Data structures (ReportData, TIDArticle, etc.)
│
├── config/                  # YAML-based configuration
├── constants/               # Application constants
├── errors/                  # Custom error types
├── logger/                  # Structured logging
│
├── frontend/                # Web UI (Vite + React + TypeScript)
│   ├── src/
│   │   ├── main.tsx          # Application entry point
│   │   ├── App.tsx           # Shell: top bar, view switch, status, shortcuts
│   │   ├── api/              # Typed façade over the Wails bindings
│   │   ├── hooks/            # Analysis, theme, status, definitions state
│   │   ├── components/       # shell/ report/ definitions/ common/
│   │   ├── styles/           # Modular CSS (layout, components, report, studio)
│   │   ├── config/           # Frontend configuration & constants
│   │   └── utils/            # Validation & helper utilities
│   ├── wailsjs/              # Generated Wails bindings
│   └── tsconfig.json         # TypeScript (strict)
│
├── kb_articles/             # Knowledge base articles (JSON)
└── docs/                    # Documentation
```

### Single-Pass Scanning Engine

The core innovation of SSSD Inspector is its single-pass log scanning engine. Instead of reading the same log files ~29 times during analysis (as traditional tools do), the engine:

1. **Builds an Aho-Corasick automaton** — 230+ error patterns + quick checks + KB article patterns are almost entirely literal strings, so they run through an Aho-Corasick trie: **O(n) matching per line in a single pass regardless of how many patterns are registered**, avoiding the classic trap of throwing hundreds of patterns into one giant regex alternation and watching performance fall off a cliff
2. **Pre-filters lines** — Fast keyword check (`sssd`, `krb5`, `ldap`, `pam`, etc.) skips ~90% of unrelated syslog lines
3. **Single pass** — Each surviving line is walked once through the automaton, extracting errors, keytab info, watchdog alerts, crypto bugs, evidence, and timestamps simultaneously
4. **Ordered phases** — Analysis runs through discrete sequential phases (config, logs, Kerberos, cross-source correlation) so dependent checks always see fully-populated report fields; parallel helper functions exist but are not invoked on this path

### Caching System

- **RegexCache** — Thread-safe `sync.RWMutex`-protected map of compiled regexp, compile-once per pattern per application lifetime
- **ACCache** — Process-wide cache for compiled Aho-Corasick automatons, so the trie over all literal patterns is built exactly once per run
- **FileCache** — LRU-like cache (max 20 files) with double-checked locking, reads each file only once per analysis
- **ScannerPool** — `sync.Pool` of 64KB byte slices for scanner buffers, reducing GC pressure

---

## 🧪 Testing

```bash
# Run all tests
cd sssd-inspector
go test -count=1 ./...

# Run with verbose output
go test -count=1 -v ./...

# Run benchmarks
go test -bench=. -benchmem -count=1 ./...

# Run specific test
go test -count=1 -run TestAnonymizeReport_DeepPII -v

# Coverage report
go test -count=1 -coverprofile=cover.out ./... && go tool cover -func=cover.out | tail -1
```

The test suite includes:
- **Unit tests** — Individual analyzer functions with mock supportconfig files
- **Integration tests** — End-to-end analysis pipeline with asset embedding verification
- **Configuration tests** — YAML config loading, validation, and defaults
- **Parallel processing tests** — Worker pool correctness, batch processing, edge cases
- **Benchmark tests** — Performance comparison between sequential and parallel scanning
- **Error handling tests** — Custom error types, wrapping, and context propagation
- **Logger tests** — Structured logging levels, fields, and filtering
- **UI component tests** — Button, ProgressBar, StatusMessage, FileInput lifecycle
- **CLI contract tests** (`cli_flags_test.go`) — Flag registry, usage output and a
  bidirectional lock between the registered flags and the Options table above
- **Entry-point tests** (`main_cli_entry_test.go`, `main_gui_entry_test.go`) —
  Dispatch order, exit codes and the CLI-vs-GUI split, for both binaries
- **Process-level tests** (`main_coverage_test.go`) — The real compiled binaries
  are built and executed, so `main()` itself (which cannot be called from a unit
  test because it ends in `os.Exit`) is guarded too
- **Entry-point coverage measurement** (`main_cover_measure_test.go`) — The
  binaries are rebuilt with `go build -cover` and run with `GOCOVERDIR`, then the
  profile is read back with `go tool covdata func`: `main()` of the CLI must be
  100%, which pins it as a pure delegation to the tested dispatchers
- **GUI export tests** (`app_export_test.go`) — `Analyze`, PDF/JSON/TXT export
  and the file dialog paths, with the native dialogs substituted through seams
- **PII redaction tests** — `-anonymize` output is asserted to be free of raw
  IPs, domains, e-mails and MAC addresses, including graph entity IDs


---

## ⚙️ Configuration

The application supports YAML-based configuration in multiple locations (checked in order):

1. `config.yaml` in the current directory
2. `~/.sssd-inspector/config.yaml`
3. `/etc/sssd-inspector/config.yaml`
4. Built-in defaults (if no file found)

### Configuration File Example

```yaml
# SSSD Inspector Configuration
app:
  name: "SSSD Inspector"
  version: "0.2.3"

analysis:
  max_file_size: "100MB"
  max_line_length: "1MB"
  buffer_size: "64KB"
  timeout: "30m"

gui:
  window:
    width: 1024
    height: 768
    title: "SSSD Inspector"

anonymization:
  enabled: true
  patterns:
    ip_v4: '\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b'
    email: '(?i)\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b'
    # (full schema: patterns + replacements maps — see config/config.go;
    # note the GUI PII toggle defaults to off, and -anonymize is opt-in)

knowledge_base:
  tid_directory: "./kb_articles"
  auto_refresh: true
  refresh_interval: "24h"

performance:
  max_workers: 4
  gc_percent: 100

cli:
  default_generate_both_formats: true
```

---

## 🛡️ Security Features

- **Path Traversal Prevention** — Archive extraction validates all paths against directory boundaries
- **Zip/Tar Bomb Protection** — Maximum file size limit (2GB) during extraction
- **PII Redaction** — Automatic masking of IPv4/IPv6, MAC, email, and domain information
- **Safe File Handling** — Streaming file processing with near-zero RAM usage
- **Timeout Protection** — Context-aware operations prevent hangs on corrupted files

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Add new error patterns** — Extend `buildErrorPatterns()` in `analyzer_logs.go` with SSSD error messages from real-world troubleshooting
2. **Create KB articles** — Add JSON files to `kb_articles/` with log patterns and configuration checks
3. **Improve tests** — Add test cases for new scenarios and edge cases
4. **Report issues** — Open GitHub issues with supportconfig samples (anonymized) showing missing detections

### Adding a New Error Pattern

```go
// In analyzer_logs.go, add to the errorPatterns map:
"Your new SSSD error message": "Human-readable explanation of what this means and how to fix it",
```


---

## 📄 License

This program is free software; you can redistribute it and/or modify it under the terms of the **GNU General Public License version 3** as published by the Free Software Foundation.

---

## ⚠️ Disclaimer

This tool is intended for diagnostic purposes. It parses logs based on patterns found in the SSSD source code; however, always verify system configurations manually before applying changes in production environments.

---

## 👤 Author

**Davide M. Puggioni** — SUSE Technical Support

The code in this repository was written with the help of Artificial
Intelligence (AI) coding assistants, used for implementation, refactoring,
test generation and documentation. All AI-assisted output was reviewed,
verified and validated by the author (builds, unit/integration tests and
real-world supportconfig analyses) before being committed.

> **Note:** the tool does not embed author/AI attribution in its output.
> Every generated report ends with a version + license footer (see
> `toolSignature` in `report.go`) rather than a copyright line, so the
> report footer never carries a year that can drift from the truth or an
> attribution that is not mirrored in this file.

---

## 🙏 Acknowledgments

- Built with [Wails](https://wails.io/) — Native Go + WebKit desktop applications
- Pattern engine inspired by SSSD C source code analysis - https://github.com/sssd/sssd
- Knowledge Base articles sourced from SUSE Knowledge Base Articles