# Implementation Summary

## 2026-09-29 — M3 follow-up: docs, an anonymization gap, and a GUI smoke that runs headless

### Documentation
`docs/api.md` had been left behind at M0: it documented `Analyze` and the
exports but none of the seven Definitions Studio bindings, and still described
`TIDArticle.Evidence` as `json:"-"`. Both fixed, plus the `CatalogInfo` fields
the M3 override added. `README.md` now documents the catalog override (precedence
table, the diagnostic-and-fallback rule, the content-keyed cache) and the Studio
in the GUI workflow, including the `Ctrl+1`/`Ctrl+2` view switch.

### Anonymization: a real gap, found by fixing a wrong test
The KB evidence test asserted that a subdomain of the search domain survives
anonymization. It does not — that assertion was wrong: the fixture used
`SearchDomain: "example.com"`, which is the *replacement token*, so the
substitution was a no-op and the "leak" was an artefact. With a realistic
domain, `dc01.corp.example` → `dc01.example.com` and the IP → `XXX.XXX.XXX.XXX`.

Re-testing all three identities properly did surface a genuine gap: the AD
domain is masked as a substring, but a log line naming one of the AD's own hosts
does not contain it (`dc01.corp.example` does not contain `ad.corp.example`), so
a report whose only known identity is `ad_domain` shipped the customer domain
through every DC host name. `parentDomain` now masks the shared parent as well,
and refuses to go below two labels so it cannot mask a bare word. Pinned by
`TestAnonymizationMasksSubdomainsOfEveryRedactedDomain`,
`TestParentDomainOnlyForMultiLabelDomains` and
`TestAnonymizationKeepsTwoLabelAdDomainReadable`.

### GUI smoke without a desktop session
The Wails window needs a desktop session and WebKitGTK, which a headless build
agent does not have, so the GUI was never actually looked at. `frontend/scripts/
gui-smoke.mjs` (`npm run smoke:gui`) loads the **built** `dist/` bundle in
headless Chromium behind a stubbed `window.go` / `window.runtime` bridge, drives
the UI (fill the path, tick *Anonymize PII*, **Analyze**, switch to the Studio,
**Validate**, **Run dry-run**) and asserts the markers a user would look for,
failing on a render crash or a missing bridge, and writing a screenshot per view.

The report it renders is a real one produced by the engine
(`-anonymize -json`), not a hand-written fixture, so it also proves the engine's
output renders in the UI. Both scenes pass 7/7 markers; the screenshots carry
10k+ distinct colours, i.e. real content. The Go ↔ WebKit boundary still needs a
manual pass on a workstation, and the guide says so.


## 2026-09-29 — M3 closing: install a catalog override from the Studio

`InstallCatalog(path, scope)` copies a generated `catalog.json` into a
definitions root, so replacing the catalog no longer means "know the path and
copy the file by hand". The OS chooser (`OpenCatalogFile`) hands it the path.

It reuses the machinery that already exists rather than adding a second way to
write definitions: the document is decoded with `decodeCatalog` — the loader's
own validation — so a file that would be skipped on load is **refused** (nothing
is written, the error names the reason), the write is atomic, and the previous
catalog is kept as `catalog.json.bak`. Because the override cache is keyed by
content, the next analysis picks the new catalog up without restarting.

The Studio's catalog panel gained "Install catalog (user)" and "Install for
system" (the latter needs root, and the service returns the same actionable
hint as the rules editor). After a successful install the panel reloads, so it
cannot keep claiming the release that was in effect before.

`TestInstallCatalogMakesTheOverrideEffective` is the test that matters: after
installing a 9.9.9 catalog, `loadOptionCatalog` must return it — an install
that did not change the effective catalog would be a copy, not a feature.
Refusal, backup and scope errors are covered alongside it, plus five frontend
cases (install per scope, cancelled chooser, refused install, service failure).
70 frontend tests in total.


## 2026-09-29 — M3: the option catalog can be overridden, KB evidence ships, the legacy JavaScript is gone

Three follow-ups from M0–M2, each closing a gap those milestones documented
instead of hiding.

### Catalog override (the M1 scope deviation)
`CatalogInfo` was read-only because the embedded catalog is build-time and no
override precedence existed, so `ImportCatalog` would have been a dead feature.
Now a `catalog.json` dropped in a definitions root overrides the embedded
catalog, with the rules loader's precedence (per-user before per-system) and the
embedded copy as the fallback — a build stays pinned to a release, but a support
engineer on a newer distro can validate against the release the customer runs.

`catalog_resolve.go` owns the resolution and one `decodeCatalog` validates both
layers, so an override can never be accepted for being "close enough". A broken
override (unreadable, malformed, or an empty options map) is a **diagnostic plus
a fallback**, never a half-applied state: validation continues against the
embedded catalog and the report states which release it actually used. Parsed
overrides are cached by content hash, so an edit takes effect on the next
analysis without restarting the GUI. `GetCatalogInfo` now reports `effective`,
`using_override`, `override_paths` and `diagnostics`; `-definitions-info` lists
the override locations and names the catalog in effect; the Studio's catalog
panel shows the same, and the inventory gained a `catalog` kind.

`TestCatalogOverrideChangesValidation` is the test that matters: with an
override that does not know `ldap_uri`, the analysis must start flagging it —
an override that only changes the display would be a lie.

### KB evidence
`TIDArticle.Evidence` was `json:"-"`, so the log lines that made a KB article
match never reached the report and the "Evidence Found" block could not render.
They are ordinary log excerpts and `anonymizeReport` already scrubbed them, so
they now ship (`evidence,omitempty`) and the block is back. The anonymization
test pins **parity** with the other log excerpts rather than inventing a
stronger guarantee: a subdomain of the search domain (`dc01.example.com`)
survives anonymization everywhere in the report today, and narrowing that mask
is a separate, report-wide decision.

### Legacy JavaScript
`config/constants.js`, `config/FrontendConfig.js`, `utils/validators.js`,
`utils/helpers.js`, `components/UIComponents.js` and the browser test-runner are
deleted. What was used moved to `config/ui.ts` (nine strings, the zoom bounds,
the accepted extensions) and `utils/fileValidation.ts`; the drag & drop and
Wails event coverage moved into `App.test.tsx`, which drives the real
components through a backend stub that records the registered callbacks.

Porting `validateFileExtension` surfaced a bug: it took the substring from the
**last** dot and compared it against `['.txz', '.tar.xz']`, so
`supportconfig.tar.xz` reduced to `.xz` and the GUI refused the most common
supportconfig filename its own file dialog offers. It now matches the longest
supported suffix, and an already-extracted supportconfig directory is accepted
because the analysis supports directories.

The porting also surfaced a real UI defect: a progress event arriving after the
analysis promise resolved (Wails can deliver the final 100% tick late) left the
progress panel stuck over a finished report. `useAnalysis` now ignores progress
outside an active run, with a regression test.

### Tests
`catalog_resolve_test.go` (precedence, fallback + diagnostic, empty-catalog
rejection, content-keyed cache, `GetCatalogInfo`, inventory),
`analyzer_config_catalog_e2e_test.go` (an override really changes validation; a
broken override reaches the report), `kb_evidence_test.go` (JSON round trip,
omission when empty, anonymization parity) and 16 new vitest cases — 51 in
total, all green, plus the full Go suite on both build tags.



## 2026-09-29 — Definitions Studio M2: the GUI is React + TypeScript, and the Studio exists

M1 built the service and the CLI surface but left the rules editor with nothing
to talk to. M2 migrates the whole frontend to React 18 + TypeScript and builds
the Definitions Studio on top of the M1 service.

### The migration
`src/main.js` (one 537-line file, imperative DOM, one `innerHTML` template per
report) is now `src/main.tsx` + typed components. The report is no longer a
string of HTML: `ReportView` and its section components (exec summary, system
info, diagnostics, config findings, timeline, KB articles, temporal clusters,
correlation graph) are React elements typed against the **generated Wails
models** (`wailsjs/go/models.ts`), so a field rename in Go breaks the build
instead of rendering `undefined`. Escaping is now a property of the framework
rather than a function that must be remembered. `CorrelationGraph.js` is
untouched and mounted through a ref — it is a self-contained SVG renderer and
rewriting it is not part of this migration. Dead legacy files (`main_old.js`,
`main_original.js`, `style.css`, `app.css`) are gone; the browser test-runner
and the JS utilities it exercises are kept.

Toolchain: React 18.3, Vite 5, TypeScript 5.9 in `strict` mode. `npm run build`
now runs `tsc --noEmit` first, so the Wails build fails on a type error. New
`npm test` (vitest + jsdom + Testing Library): 23 tests over the report view,
the Studio and the bridge guard.

### api/backend.ts
One typed façade over the generated bindings. Components never import
`wailsjs/*` directly, and every call is guarded by `backendAvailable()`: running
`npm run dev` in a plain browser now explains itself instead of throwing a
`TypeError` on `window.go`.

### The Definitions Studio
Second tab (`Ctrl+2`; analysis is `Ctrl+1`), four panels:

- **Discovery inventory** — every search location in loader order with its
  state, rule/article counts, size and a real write probe; the loaded rules with
  their file and line; the skipped inputs. "Open user/system folder" calls
  `OpenDefinitionsRoot`.
- **Rule editor** — load, edit, validate, save, per scope. Changing the scope
  reloads the document so the editor can never show one scope while saving into
  another. A refused save renders the diagnostics and says "refused"; it is
  never reported as a success.
- **Dry-run** — prefilled with the path the analysis view holds; firing rules
  first with the evidence line, silent rules listed too ("my rule did nothing"
  is answerable). A refresh back to the inventory happens on a successful save.
- **Option catalog** — the embedded catalog's provenance (read-only).

A "Schema reference" table in the editor mirrors `rules_engine.go`: the loader is
the authority, and a starter template that is valid by construction.

### New Go surface
`ReadRuleYAML(scope)` (`RuleDocument`): the editor needs the document that is
actually in effect, and a scope without one is the normal state of a fresh
install — `Exists=false` is not an error. Exposed as the seventh Wails binding
and covered by `TestReadRuleYAML*`.

### Known gap found on the way
`TIDArticle.Evidence` is tagged `json:"-"` in `types.go`, so the legacy "Evidence
Found" block in the KB articles section could never render. The React port does
not fake it; shipping that evidence needs a backend decision (it is scrubbed
only in anonymize mode) and is left to a later milestone.



## 2026-09-29 — Definitions Studio M0+M1: definition loading is visible and verifiable

Every data-driven definition (YAML rules, external KB articles) has always been
loaded fail-safe — a bad file is skipped — but silently, from a set of search
paths a user could not discover. That is now impossible: the loader reports
what it skipped, and a new service exposes the search paths, the validation and
a rule dry-run to both the CLI and the GUI.

### M0 — loading diagnostics (never silent)
`Diagnostic{File,Line,Message,Severity}` is attached to every report
(`ReportData.Diagnostics`, `omitempty` so existing consumers keep parsing) and
`ReportData.AddDiagnostics` deduplicates. The rules loader (`loadAnalysisRules`)
and the KB loaders (`loadKBArticlesDiag`, external directories merged by
precedence) return diagnostics with real line numbers (`yaml.Node`,
`jsonErrorLine`) and absolute paths. They reach the user through the
TXT/HTML/JSON reports, a stderr banner (`[!] definitions: file:line: message`)
and the `definitions-warning` GUI event; anonymization scrubs `File`/`Message`.

### M1 — definitions service, CLI surface, Wails bindings
`definitions_service.go` is the single implementation behind both front-ends:
`ListDefinitions` (every candidate path, what is there, whether it is
writable), `ValidateRuleYAML` (`parseRulesDocument` — the loader's own
validator, so the editor cannot disagree with the analysis), `TestRulesAgainst`
(`applyAnalysisRules` on a throwaway report: the real matcher, no side
effects), `SaveRuleYAML` (refuses documents the analysis would skip, atomic
write, keeps `rules.yaml.bak`, only the user/system scopes) and
`GetCatalogInfo`.

Rules parsing now has one validation path shared by the loader, the editor and
the CLI gate. It also catches the mistakes that silently produced **zero**
rules: a missing top-level `rules:` header, another top-level key, a rule
without a name, and duplicate names — within a file and across files.

New CLI flags, registered for BOTH binaries (`definitions_cli.go` dispatches
them through one implementation, so they cannot drift as `-logdir` once did):

| Flag | Behaviour |
|------|-----------|
| `-definitions-info` | prints every definition search path, its contents and writability |
| `-validate-rules` | validates the installed definitions; exits 1 if anything would be skipped |
| `-rules-test <path>` | dry-runs the loaded rules against a supportconfig; exits 2 without a path |

Wails bindings (`app_definitions.go`) expose the same five operations plus
`OpenDefinitionsRoot`, which opens the definitions folder in the file manager;
`frontend/wailsjs` was regenerated with `wails generate module`.

### Tests
`definitions_service_test.go`, `definitions_cli_test.go`,
`app_definitions_test.go` and `definitions_e2e_test.go` (real binaries: exit
codes 0/1/2 of the three flags, on both the CLI and the hybrid binary). M0
added `definitions_diagnostics_test.go` plus `main_test.go` (TestMain isolates
the definition roots from the machine's real ones). The entry-point coverage
guard now drives the three new flags as well.


## 2026-09-29 — evidence contract, severity policy, semantic rules

Follow-up pass from the plan-mode review: three structural weaknesses behind
the recent false positives are fixed, with no new dependencies and the binary
staying standalone and deterministic.

### Ground-truth confidence on every catalog option
`OptionMeta` gains `Confidence` (`api` > `man` > `curated`), derived once from
`Source` by `applyConfidenceTags` at the end of generation. Of the 519 SSSD
2.14.0 options, 481 are API-backed, 37 man-only and 1 curated. `catalog.md`
renders the layer per option (`sssd-ad.conf (api)`); findings carry it as
`confidence`, and unknown-option verdicts cite the full API-backed catalog.

### Central severity policy
`catalogSeverity` in `config_severity.go` caps every non-allowlisted rule at
`SevWarning`, so a weak knowledge layer can never sink the health score on
its own. Pre-existing checks are grandfathered with stable rule IDs
(`structure:*`, `domain:*`, `idmap:*`, `correlate:*`, `log:*`); user YAML
rules (`rule:*`) keep their declared severity. Severity is therefore a table,
not a per-call-site decision.

### Evidence contract
`ConfigFinding` (and `GraphFindingNode`) gain `rule_id`, `confidence` and
`doc_ref`, threaded through `addConfigFindingEx`, the correlation graph, the
HTML template and the Wails `models.ts`. Every finding is now traceable to
the check and documentation that produced it.

### Provider-aware semantic rules
`config_semantics.go` adds three cross-option checks the syntactic catalog
cannot express: `semantic:ldap-uri-without-ldap-provider`,
`semantic:dead-bind-credentials` (evidence redacts the secret as `***`) and
`semantic:dead-ldap-options`. All stay `SevWarning`. The user LDAP
supportconfig from the previous pass still yields exactly its 2 genuine
findings, plus zero semantic noise.

### Tests
9 new tests in `config_semantics_test.go`: per-rule positives and negatives
(including the real LDAP config staying clean), the severity cap, the
evidence contract on catalog findings, and the confidence tags surviving the
embed round-trip.

## 2026-09-28 — fix the provider/* section false positives on a real LDAP domain

A second review of a real supportconfig (SLES 15 SP7, SSSD 2.10.2, `id_provider = ldap`)
surfaced two more false positives, both regressions of the previous pass.

### `ldap_group_*` / `ldap_user_member_of` rejected inside `[domain/NAME]`
The report claimed four options were "only valid in [provider/ad/id],
[provider/ipa/id], [provider/ldap/id]". Those section names come from the API
definitions (`sssd.api.d/*.conf`), which name SSSD's internal configuration
hierarchy — an sssd.conf has no `[provider/...]` section at all. The options
affected are declared *only* by the API files, so the man-page pass never
reached them and their `provider/*` sections were the only ones recorded,
making the family comparison fail. `normalizeAPISection` now maps the API
headers to their real sssd.conf names (`provider/ad/id` -> `domain/ad/id`,
`provider` -> `domain`); no `provider/` section survives generation. The
previously fixed `ldap_uri` had escaped the bug only because the man pages also
declared it.

### `id_provider = files` reported as a fatal CONFIGURATION ERROR
`validateDomainStructure` compared against a hardcoded literal that had drifted
from the documentation and did not contain `files`, so a legitimate
`[domain/files]` section produced a `SevError` finding, a `problems` entry and
a depressed health score. The accepted values now come from the embedded
catalog via `idProviderList`, with `fallbackIdProviders` used only when the
catalog is unavailable; the list can no longer drift from the documentation.

On the reported configuration the findings drop from 7 to 2, and both remaining
ones are genuine: `reconnection_retries` in `[nss]` and `ldap_group_number`,
neither of which is an SSSD option.

### Tests
`TestNormalizeAPISection`, `TestCatalog_LdapGroupOptionsValidInDomainSection`
(the catalog shape plus the end-to-end LDAP configuration),
`TestCatalog_IdProviderFilesIsAccepted` and `TestIdProviderList` (both the
catalog-derived list and the fallback). An earlier version of the
`id_provider` test did not actually guard the fallback path and passed against
the reverted code; extracting `idProviderList` made it testable, and a mutation
check now confirms the test fails when `files` is removed from the fallback.

## 2026-09-28 — catalog section awareness, provenance, false-positive fixes

A user review of a real supportconfig report found that the catalog validator,
although working, produced two false positives out of four findings and never
stated where its knowledge came from. This pass fixes all of it.

### Findings were not attributed to their source
The report asserted *"It is not part of the SSSD option list for this release"*
without ever saying which release, or how many options had been compared —
`grep -rn 'catalog' report.go types.go` returned nothing. Now every catalog
finding ends with `[checked against the SSSD 2.14.0 option catalog: 519
options from 54 documentation sources]`, and the same line is published as
`ReportData.CatalogProvenance` (`catalog_provenance` in JSON) and rendered in
the HTML and text reports, so the audit trail is visible even when no finding
is raised. `OptionMeta.Source` records the exact man page each option came from.

### `config_file_version` reported as unknown (false positive)
It is a real `[sssd]` option, absent from the upstream `sssd.conf.5.xml` of
this release. New curated table `curatedOptions` in `catalog_gen.go` fills such
gaps; entries are tagged `"source": "curated"` so curated knowledge stays
distinguishable from man-page knowledge.

### 88 phantom options in the catalog
`<term>` elements were harvested unconditionally, but a man page is full of
things that are not options: PAM return codes (`pam_auth_err` in pam_sss(8)),
signal names (`sighup` in sssd(8)), LDAP attribute names (`gecos` in the
InfoPipe tables), and the *enumerated values* of an option, which upstream
writes as a nested `<variablelist>` inside the option's own entry
(`always`, `true`, `no_session`, `hybrid`, ...). The consequence was the worst
kind of false negative: a genuine typo like `gecos = x` validated as a known
option, and those names could be offered as "did you mean" suggestions.

Fixed by resolving the sssd.conf section of every `<term>` from the enclosing
DocBook `refsect` id (`docSectionMarkers` / `sectionRefID`, which also maps
`all-section-options` and the responder sections), tracking the
`variablelist` nesting depth (`indexedTerms`), turning nested terms into the
parent option's allowed values, and pruning anything that cannot be placed in a
section. The bare `.B <word>` roff fallback may now only enrich options that
are already known, never invent them. Net effect: 596 → 519 options, 54 → 49
sections, and the man-page enumerations for `pac_check` and
`pam_initgroups_scheme` are now learned automatically instead of being curated.

### Section-blind validation
`OptionMeta.Sections` was generated but never read: a valid option in the wrong
section could not be distinguished from a typo. New check `wrongSectionReason`
emits a `config_section` finding naming the section the option belongs to,
which has a different remediation (move the key, do not rename it).

Provider options are written in `[domain/<name>]`, while the provider man pages
document them per provider, so the comparison is done on the section *family*
(`sectionFamily`): an `sssd.conf` has no `[provider/ad]` section, and whether
`ldap_uri` applies to a given domain depends on its `id_provider`, which the
catalog does not model. Reporting `ldap_uri` or `ad_gpo_access_control` as
invalid inside `[domain/example.com]` would have been a new false positive.

### Regression tests
`config_catalog_test.go` gained the four cases from the reported supportconfig:
`config_file_version` known, `reconnection_retries` still unknown but with
provenance, enumerated values absent from the option list, wrong section
distinguished from typo, plus a full valid AD configuration that must produce
no finding, provenance publication, `sectionFamily` and `wrongSectionReason`.

## 2026-09-28 — /etc/hosts states, embedded KB, offline option catalog

### False "Malformed /etc/hosts" fixed (`conffiles.go`)
- Supportconfig never ships `/etc/hosts` as a standalone file: it is inlined
  into `network.txt`, so the old lookup for a file literally named `hosts`
  always missed and the report claimed the file was malformed.
- `analyzeHosts` now classifies three states (`HostsFileStatus`):
  `HostsOK`, `HostsMissing`, `HostsMalformed`. Only genuinely broken content
  raises a finding; an absent file is reported as "not captured" instead.
- Covered by `conffiles_test.go` / `analyzer_hosts_test.go`.

### Knowledge base embedded in the binary (`kb.go`)
- `kb_articles/*.json` is embedded with `go:embed`; external
  `kb_articles/` next to the binary or in the CWD is still merged on top
  (external entries win on ID collision). The inspector no longer needs the
  data directory to be shipped alongside it.

### Offline sssd.conf option catalog (`-gen-catalog`)
- New maintainer-oriented generator flag `-gen-catalog <dir>` (registered in
  `cli_flags.go`, dispatched from both `main_cli.go` and `main_gui.go`).
- `catalog_gen.go` builds `sssd_catalog/catalog.json` + `catalog.md` from
  `upstream/sssd.api.conf`, `upstream/sssd.api.d/*.conf`, DocBook man pages
  (`upstream/src/man/**/*.xml`) and, as a fallback, compiled roff pages
  (`sssd*.5`, `sssd*.5.gz`). `enrichKnownEnums` fills the value lists that
  upstream only documents as prose.
- `sssd_catalog/catalog.json` (519 options, 49 sections for SSSD 2.14.0) is
  embedded with `go:embed` and committed, together with the readable
  `catalog.md`.
- New `config_catalog.go` validates every `sssd.conf` key against the
  catalog: unknown parameter (with a "did you mean?" suggestion via
  Levenshtein distance), wrong type for bool/int options, and values outside
  the documented enumeration. All three are `SevWarning` — SSSD ignores such
  settings, it does not fail.
- `upstream/` is a git-ignored drop-zone (`upstream/README.md` documents the
  exact files to copy); regenerate with `go run . -gen-catalog upstream`.

### Bug fixes found by the catalog
- `ldap_idmap_min_id` / `ldap_idmap_max_id` are **not** SSSD options: the
  ID-mapping range overlap check never matched a real `sssd.conf`. It now
  reads `ldap_idmap_range_min` / `ldap_idmap_range_max` (with the legacy
  `idmap_range_*` spellings kept as aliases) and treats `range_max` as
  exclusive.
- `ad_gpo_access_control = disabled` is documented in `sssd-ad(5)` but was
  reported as a `CONFIGURATION ERROR`. All three documented modes
  (`disabled`, `permissive`, `enforcing`) are now accepted.

### Regression guards
- `config_catalog_test.go`: embedded catalog integrity, typo suggestion,
  type/enum checks, pure helper coverage.
- `analyzer_config_catalog_e2e_test.go`: end-to-end `analyzeData` run proving
  the findings reach the report warnings (and only the warnings).
- `analyzer_config_checks_test.go` / `analyzer_config_validate_test.go`:
  real option names, adjacent-range acceptance, GPO `disabled`.
- `main_cover_measure_test.go` now runs `-gen-catalog` both ways (success
  from a minimal fixture and a missing source directory).
- `cli_flags_test.go` README cross-check rewritten to capture whole
  backticked cells, so it validates both directions.

---

## 2026-09-25 — Version 0.2.3: single-source version + release packaging

### Version alignment
- `constants.AppVersion` (`0.2.3`) is now the only literal: `analyzer_core.go`
  hardcoded `"0.2.0"` and `config.DefaultAppVersion` duplicated the value.
- The CLI path masked the drift (`shared.go` re-copies the constant onto the
  report), but the GUI path (`app.go` → `analyzeData`, and therefore the
  PDF/TXT/JSON exports) and the config default still carried `0.2.0`.
- `analyzer_core.go` and `config.DefaultAppVersion` now reference
  `constants.AppVersion`; `config.yaml`, `README.md`,
  `docs/configuration.md`, the sample reports (`scA_report.json`,
  `mocksc_report.json`) and the frontend `APP_VERSION`
  (`FrontendConfig.app.version`) quote `0.2.3`.

### Regression guard
- New `version_drift_test.go`: no `AppVersion` string literal outside
  `constants/`, `config.DefaultAppVersion == constants.AppVersion`, and every
  `version:` in `config.yaml`, `README.md` and `docs/configuration.md` must
  match the constant. The drift that shipped in the GUI could not pass it.
- `test_fixture_hygiene_test.go` skips the generated `dist/` directory.

### Release packaging
- New `build_release.sh`: `wails build -platform linux/amd64 -tags webkit2_41
  -ldflags "-w -s" -clean` (output captured with `tee`, `pipefail` so a failed
  build aborts the script), a `-v` smoke test comparing the binary output with
  `constants.AppVersion` (safe headless: `runHybridCLI` returns before
  `launchGUI`), then `dist/sssd-inspector-0.2.3-linux-amd64.tar.gz` containing
  the binary, `kb_articles/` (loaded next to the executable), `LICENSE`,
  `LICENCE.md`, `README.md`, `config.yaml` and `build_instructions.txt`, plus a
  `.sha256` checksum.
- `.gitignore` ignores the generated `dist/`; `build_instructions.txt`
  documents the current command (the old one lacked `-tags webkit2_41`).
- `docs/development.md` Release Process now points at `build_release.sh` and
  uses the `v0.2.3` tag example.

---


## Overview

This document summarizes the comprehensive refactoring and enhancement of the SSSD Inspector project, implementing Senior Go Developer and Senior Frontend Developer best practices while maintaining all existing functionality.

## Completed Improvements

### 1. 📚 Complete Documentation System

**Created comprehensive documentation structure:**
- `docs/README.md` - Project overview and architecture
- `docs/configuration.md` - Detailed configuration guide
- `docs/api.md` - Complete API reference
- `docs/development.md` - Development guide and standards
- `docs/CHANGES.md` - This summary document

**Documentation features:**
- Architecture diagrams and explanations
- API method documentation with examples
- Configuration reference with all options
- Development workflow and coding standards
- Troubleshooting guides

### 2. ⚙️ Configuration Management System

**Implemented flexible YAML-based configuration:**
- `config.yaml` - Main configuration file with all settings
- `config/config.go` - Configuration structures and loading logic
- Support for multiple configuration locations
- Environment variable overrides
- Configuration validation

**Configuration includes:**
- Application metadata and versioning
- Analysis parameters (file sizes, timeouts, buffers)
- GUI settings (window dimensions, colors)
- Anonymization patterns and replacements
- Knowledge base settings
- Logging configuration
- Performance tuning options
- Report generation settings

### 3. 🔧 Constants Management

**Created comprehensive constants system:**
- `constants/constants.go` - All application-wide constants
- Eliminated magic numbers throughout codebase
- Centralized string constants for UI elements
- File pattern definitions
- Error message constants
- Progress message constants

**Benefits:**
- Improved maintainability
- Reduced duplication
- Easier configuration management
- Better testing capabilities

### 4. 📖 GoDoc Documentation

**Added comprehensive GoDoc comments:**
- All exported functions documented
- Struct and type documentation
- Parameter and return value descriptions
- Usage examples and context
- Error handling documentation

**Files enhanced:**
- `app.go` - Complete API method documentation
- `main.go` - Entry point and CLI documentation
- `utils.go` - Utility function documentation
- `config/config.go` - Configuration documentation

### 5. 🛡️ Enhanced Error Handling

**Implemented proper error wrapping:**
- Replaced `fmt.Errorf("%v", err)` with `fmt.Errorf("context: %w", err)`
- Added context to all error messages
- Proper error propagation in call chains
- Structured error types for different scenarios

**Error handling improvements:**
- CLI functions now return errors instead of calling `log.Fatalf`
- Better error context and debugging information
- Consistent error handling patterns
- Graceful degradation where possible

### 6. 🏗️ Clean Code Refactoring

**Refactored utils.go with clean architecture:**
- `FileProcessor` struct for streaming operations
- `SectionExtractor` struct for file section parsing
- `SafeFileReader` struct for small file operations
- `FileFilter` struct for file relevance checking
- Backward compatibility functions for existing code

**Clean code principles applied:**
- Single Responsibility Principle
- Dependency Injection
- Struct-based organization
- Clear separation of concerns
- Comprehensive error handling

### 7. ✅ Testing Compatibility

**Ensured all existing tests pass:**
- Maintained backward compatibility
- Updated function signatures where needed
- Preserved all existing functionality
- Added configuration-aware testing

**Test results:**
```
ok      sssd-inspector  0.013s
?       sssd-inspector/config   [no test files]
?       sssd-inspector/constants        [no test files]
```

## Technical Improvements

### Architecture Enhancements

1. **Modular Design**: Clear separation between configuration, constants, utilities, and application logic
2. **Dependency Management**: Proper package structure with minimal coupling
3. **Configuration-Driven**: Application behavior controlled through YAML configuration
4. **Error Resilience**: Robust error handling with proper context and wrapping

### Code Quality Improvements

1. **Documentation**: Comprehensive GoDoc and markdown documentation
2. **Constants**: Eliminated magic numbers and strings
3. **Type Safety**: Strong typing with proper struct definitions
4. **Error Handling**: Consistent error patterns throughout codebase

### Maintainability Improvements

1. **Configuration**: Externalized all configurable parameters
2. **Documentation**: Complete API and development documentation
3. **Testing**: Maintained test compatibility while improving structure
4. **Standards**: Applied Go best practices and clean code principles

## Configuration Example

```yaml
app:
  name: "SSSD Inspector"
  version: "0.2.3"

analysis:
  max_file_size: "100MB"
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
  replacements:
    ip_v4: "XXX.XXX.XXX.XXX"
```

## API Documentation Example

```go
// Analyze handles the secure extraction, routing, and cleanup of the target logs.
// This is the main analysis method that orchestrates the entire diagnostic process.
//
// Parameters:
//   - targetPath: Path to the supportconfig directory or archive file
//   - anonymize: Whether to redact PII (Personally Identifiable Information) from the report
//
// Returns:
//   - ReportData: Comprehensive analysis results
//   - error: Any error that occurred during the analysis process
func (a *App) Analyze(targetPath string, anonymize bool) (ReportData, error)
```

## Usage Examples

### CLI Usage
```bash
# Basic analysis
./sssd-inspector /path/to/supportconfig.txz

# With anonymization
./sssd-inspector -anonymize /path/to/supportconfig.txz

# Generate specific formats
./sssd-inspector -txt -html /path/to/supportconfig.txz
```

### Configuration Override
```bash
# Override configuration location
SSSD_INSPECTOR_CONFIG=/custom/path/config.yaml ./sssd-inspector

# Override specific settings
SSSD_INSPECTOR_LOG_LEVEL=debug ./sssd-inspector
```

## Benefits Achieved

### For Developers
- **Easier Maintenance**: Clear structure and comprehensive documentation
- **Better Testing**: Modular design enables focused testing
- **Configuration Flexibility**: Externalized configuration for different environments
- **Code Standards**: Consistent patterns and best practices

### For Users
- **Customizable Behavior**: Configuration file controls all aspects
- **Better Error Messages**: Clear, actionable error information
- **Documentation**: Complete guides for usage and troubleshooting
- **Stability**: Robust error handling and graceful degradation

### For Operations
- **Deployment**: Configuration-driven deployment
- **Monitoring**: Structured logging and error reporting
- **Maintenance**: Clear documentation and modular design
- **Scaling**: Performance tuning through configuration

## Future Enhancements Enabled

The refactored architecture enables several future improvements:

1. **Plugin System**: Modular design supports plugin architecture
2. **Configuration Templates**: Environment-specific configurations
3. **Advanced Logging**: Structured logging with configurable levels
4. **Performance Monitoring**: Built-in performance metrics
5. **API Extensions**: Clean API structure for future enhancements

## Validation

✅ **All Tests Pass**: Existing functionality preserved  
✅ **Build Success**: Project compiles without errors  
✅ **Configuration Loading**: YAML configuration works correctly  
✅ **CLI Functionality**: Command-line interface operates properly  
✅ **Documentation Complete**: Comprehensive docs created  
✅ **Code Standards**: Go best practices applied  

## Conclusion

This refactoring successfully transformed the SSSD Inspector project into a well-structured, documented, and maintainable codebase while preserving all existing functionality. The implementation follows Senior Go Developer and Senior Frontend Developer best practices, providing a solid foundation for future development and maintenance.

The project now features:
- Professional documentation system
- Flexible configuration management
- Clean, maintainable code architecture
- Robust error handling
- Comprehensive testing compatibility
- Developer-friendly standards and guidelines

All improvements were implemented without removing any existing functionality, ensuring backward compatibility while significantly enhancing code quality and maintainability.

---

## 2026-09-18 — Regression test net + PII-leak fix (unreleased work)

### Regression guard for the CLI flag surface (the `-logdir` lesson)
- The CLI flag set lived in two duplicated `main()` functions (`main_cli.go`,
  `main_gui.go`); a previous `-logdir` feature was silently lost in a merge.
- All flags now come from a single registry: `registerCLIFlags()` in
  `cli_flags.go` (tag-free, shared by both binaries; `-compare` stays
  CLI-only by construction).
- `cli_flags_test.go` locks: (1) a hardcoded flag contract (`v, analyze,
  txt, html, json, anonymize, compare`), (2) every flag shown in `-h`
  usage, (3) a two-way lock between the registry and the README.md Options
  table. Deleting or silently adding a user-visible flag now fails the suite.
- `integration_test.go` is GUI-only: added `//go:build !cli` so
  `go test -tags cli ./...` builds and runs (it previously failed), enabling
  the CLI regression matrix.
- New CI workflow `.github/workflows/ci.yml`: gofmt, `go vet` in both
  build modes, both binary builds, `go test` default + `-tags cli`, `-race`
  in both modes, and a CLI smoke test on the built binary.

### Coverage of previously untested areas (71.6% -> 82.5% statements)
- `runCLI` end-to-end (`shared_cli_test.go`, `shared_cli_modes_test.go`):
  dir + tar.xz input, txt/html/json generation, stdout contract, PII
  redaction, missing-path error, no-output mode.
- Archive extraction (`loaders_test.go`): real `.tar.xz` built in-test,
  relevant-files-only contract, invalid archive error, path-traversal flattening.
- Parallel phases (`analyzer_parallel_test.go`): equivalence and
  repeatability of `analyzePhase1/2Parallel` vs the sequential phases.
- Correlation engine (`analyzer_correlate_test.go`,
  `analyzer_correlate_checks_test.go`): helper tables + the four
  contradiction checks with no-false-positive anchors.
- `analyzeKerberosAndKeytab` (`analyzer_auth_keytab*.go`), FileFilter +
  context scanning (`utils_filter_test.go`), caches
  (`cache_test.go`), plus `findingKeys`/`categoryFor`/`chainable`
  (`analyzer_extras_test.go`).

### Entry-point coverage and remaining test gaps closed
- `main()` cannot be called from a unit test (it ends in `os.Exit`), so the
  decidable logic of both binaries moved into unit-testable functions and the
  entry points are now guarded at process level:
  - `runCLIEntry` / `dispatchCLI` (`main_cli_entry_test.go`, `//go:build cli`):
    dispatch order (`-v` → `-compare` → `-analyze` → positional), exit codes
    (0/1/2), usage output, trailing-format-flag rescan and the
    default-both-formats behaviour.
  - `runHybridCLI` / `launchGUI` split (`main_gui_entry_test.go`): the hybrid
    binary must handle a path without ever starting the GUI, plus the
    CLI-only `-compare` rejection.
  - `main_coverage_test.go`: builds the real `cli` and hybrid binaries and runs
    them (`-v`, `-h`, unknown flag, `-analyze`, positional path, `-compare`),
    asserting exit codes and output. This is what covers `main()` itself and
    protects against turning an `os.Exit(1)` into a `log.Fatalf`.
- GUI export paths are now testable through dialog seams
  (`openFileDialogFn` / `saveFileDialogFn` in `app.go`):
  `OpenFileBrowser`, `SavePDF` (base64 data-URL decode), `SaveJSON`, `SaveTXT`,
  cancellation semantics (`ErrCancelled`), dialog errors and write errors
  (`app_export_test.go`, `//go:build !cli`), plus the Wails `startup` hook.
- Remaining zero-coverage code is only the parts that require a display server:
  `launchGUI()` and the `main()` statements that call it.
- `README.md`: Author section now states that the code was developed with the
  help of AI assistants, and the Testing section documents the new
  entry-point/process-level tests and both regression guards.

### PII fix found by the new tests
- `TestRunCLI_AnonymizeRedacts` exposed a real leak: with `-anonymize`, the
  top-level `/ad_domain` and `/hostname` fields and the graph entity IDs
  still carried raw PII; the uppercased AD-domain variant leaked into the
  sssd.conf snippet and the correlation messages.
- `anonymizeReport` (analyzer_core.go) now: snapshots the originals before
  overwriting, masks `ad_domain`/hostname case variants inside every
  embedding field, replaces the top-level fields with placeholders, and
  re-keys graph entity IDs from the scrubbed values (rewriting edge
  endpoints and collapsing duplicates).

---

## 2026-09-23 — Raw SSSD log mode restored (`-logdir`)

### The feature
- `-logdir <path>` analyses raw SSSD logs without any supportconfig: a
  directory of `*.log` files (rotated `*.log.N` / `*.log-<date>` included) or
  a single log file, e.g. `/var/log/sssd`.
- The flag was originally added in `9f2d01e` and silently lost during a
  repository restructure (duplicated flag surface in two `main()` functions).
  It is now registered in the shared `cli_flags.go` registry, so it exists on
  BOTH binaries and is pinned by the flag contract test + the README lock.
- Dispatch order (pinned by tests): `Version -> Compare -> LogDir -> Analyze ->
  positional`. `-logdir` is handled by the hybrid dispatcher too, so it can
  never fall through to `launchGUI()` (headless guard).
- Same engines as supportconfig mode: single-pass scan, timeline, temporal
  clusters, KB suggestions (the historical version passed a nil KB list,
  disabling all KB features), root-cause sequence correlation, executive
  summary and correlation graph.
- Fields that only a supportconfig can provide are reported as
  `N/A (raw log mode)` (`constants.RawLogModeNA`); supportconfig-only findings
  ("sssd.conf not found", "sssd.service not running", "No Kerberos Keytab")
  are never emitted.
- Report output reuses `writeReports` (extracted from `runCLI`), so both modes
  write `<basename>_report.{txt,html,json}`; like `-analyze`, only explicitly
  requested formats are written.

### Anonymization hardening (found while building the mode)
- Without an `sssd.conf` there is no parsed domain/realm/hostname, so raw-log
  mode HARVESTS them from the log lines themselves (`Domain [...]`,
  `[domain/...]`, `realm=`, principals, `ldap://` URIs, syslog `prog[pid]`
  prefixes) and seeds both the report fields and a new `extraTokens` parameter
  of `anonymizeReport` — otherwise `-anonymize` would have been a silent no-op
  (the same failure class as the Hostname-less supportconfig leak fixed in
  `208a05d`).
- New `isRedactableToken` guard: the N/A marker and the unknown-value markers
  (`""`, `None`, `Not configured`, `Unknown`) can never be used as redaction
  tokens — replacing them would corrupt every field carrying them and would
  turn graph nodes into `redacted-host` without anything being redacted.
  Applied across `maskString`, the top-level field overwrites and the graph
  entity/replacement guards (`analysis_graph.go` included).
- Harvest anti-corruption rules: explicit domain markers allow single labels,
  anything else must contain a dot (keeps `[be[ldap_id]]`-style service names
  out of the domain list); `looksLikeRealm` rejects `realm=supports`-style
  captures; the syslog host must be followed by `prog[pid]:`; a shared
  blacklist drops `root`/`kernel`/generic words. Host tokens are applied BEFORE
  domain tokens so `dc01.example.test` collapses whole instead of leaving the
  `dc01` label behind.
- `finalizeReport` now holds the graph+anonymize closing steps, shared by
  `analyzeData` and `analyzeLogsOnly`, so the PII-critical tail has exactly one
  implementation. (Dedup/executive summary stay inline: the supportconfig
  `debug_level` hint is appended after the summary and moving it would change
  existing health scores.)
- `constants.SupportconfigLogFiles()` replaces the four duplicated
  `[]string{"sssd.txt", "messages", "messages.txt"}` literals
  (`analyzer_logs.go`, `analyzer_singlepass.go`, `analyzer_auth.go`, `kb.go`),
  and `performSinglePassScanOnFiles` parameterises the single-pass scanner so
  raw-log mode reuses it while the supportconfig path keeps its exact file set.

### Tests added
- `logdir_test.go`: file-selection table (`*.log` + rotations; not `.gz`, not
  supportconfig names), `collectLogFiles` paths (dir / single file / missing /
  empty / non-log), harvesting (identity values found + false positives
  rejected), `analyzeLogsOnly` (errors detected, NO supportconfig-only
  findings, N/A markers, seeded identity fields, summary/clusters/graph),
  anonymization (zero raw PII in JSON, N/A markers survive, raw values kept
  without `-anonymize`), `runLogDirAnalyze` end-to-end for all three formats,
  and the bidirectional isolation guard (`analyzeData` ignores `*.log`).
- `main_cli_entry_test.go`: dispatch precedence (`-logdir` beats `-analyze`),
  missing path -> exit 1, end-to-end JSON.
- `main_gui_entry_test.go`: `-logdir` is handled (never reaches `launchGUI()`)
  and keeps the same precedence as the CLI binary.
- `main_cover_measure_test.go`: `-logdir` success + failure runs in both
  binaries, keeping the `dispatchCLI`/`runHybridCLI` coverage thresholds honest.
- `cli_flags_test.go`: `FlagLogDir` added to the hardcoded flag contract (this
  test failed first, before the README/docs were updated — exactly as designed).
- CI: smoke step builds raw logs, runs `-logdir -json -anonymize`, asserts the
  error is found, the domain is redacted, and `-h` advertises the flag.

---

## 2026-09-23 — Test-data hygiene: no real domains or hostnames

### The problem
- Test fixtures, test comments, CI smoke data and the `-logdir` documentation
  contained real-world values taken from an analysed environment: a real
  third-party company domain, internal hostnames taken from web/prod machines,
  a private IP address and an employer domain used as test data. Tests must
  only ever use reserved/generic names — so none of those literals is repeated
  here either (the ban list in the guard test is where they are named).

### The fix
- All occurrences replaced with reserved values: domain `example.test`
  (RFC 6761 `.test`, and deliberately NOT `example.com` so the redaction tests
  can still observe the placeholder), hostnames `testhost01`/`testhost02`, IP
  `192.0.2.10` (RFC 5737 TEST-NET-1), realm `EXAMPLE.TEST`.
- `anonymize_sources_intra_test.go` renamed to
  `anonymize_sources_domain_test.go` (the name itself carried the domain).
- New guard `test_fixture_hygiene_test.go` walks `*.go`, `*.md` and the CI
  workflows, strips the legitimate knowledge-base documentation URLs, and
  fails when any banned real value appears. The ban list is assembled from
  string parts so the guard does not trip on itself, and grows whenever a new
  real value is discovered.
- Left untouched on purpose: the knowledge-base article URLs (product
  documentation) and the author identification in `wails.json`/`config.yaml`.
