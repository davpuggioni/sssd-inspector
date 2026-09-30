// helpers.ts — shared fixtures and the backend mock used by the component
// tests. The fixtures mirror the generated Wails models, so a test failure
// means the UI disagrees with the Go payloads.
import { vi } from 'vitest';
import type { CatalogInfo, DefinitionsInventory, RuleTestResult, RuleValidationResult, ReportData } from '../api/backend';

/** Minimal ReportData with the fields the report view actually reads. */
export function makeReport(overrides: Partial<ReportData> = {}): ReportData {
  const base = {
    app_version: '0.2.4',
    timestamp: '2026-09-29 10:00:00',
    support_case_id: 'SR#12345',
    kernel_version: '5.14.21',
    sles_release: 'SLES 15 SP6',
    scc_status: 'out of date',
    summary: {
      health_score: 72,
      critical_count: 1,
      error_count: 2,
      warning_count: 3,
      problem_count: 1,
      log_error_count: 4,
      top_category: 'crypto',
      top_category_hits: 3,
      headline: 'RC4 enctype in sssd.conf',
    },
    hardware_manufacturer: 'QEMU',
    hardware_model: 'Standard PC',
    hypervisor: 'kvm',
    virtual_identity: 'kvm',
    mac_type: 'Unknown/None',
    sssd_installed: true,
    sssd_config_found: true,
    sssd_service: 'active',
    winbind_service: 'inactive',
    nscd_status: 'inactive',
    nscd_caching: ['passwd', 'group'],
    sssd_packages: ['sssd-2.9.1'],
    nsswitch_valid: true,
    pam_sss_installed: true,
    pam_gdpr_restricted: false,
    hosts_issues: [],
    hosts_file_status: 'present',
    nameservers: ['10.0.0.1'],
    search_domain: 'example.com',
    time_service: 'ntp',
    kerberos_realm: 'EXAMPLE.COM',
    keytab_found: true,
    ad_provider_mode: true,
    enumerate_issue: false,
    use_fqdn_set: true,
    ad_domain: 'ad.example.com',
    hostname: 'client.example.com',
    config_findings: [
      {
        severity: 2,
        category: 'crypto',
        message: 'RC4 enctype present',
        source_path: 'sssd.conf',
        source_key: 'ad_config',
        source_line: 12,
        evidence: 'ldap_default_authtok = rc4-hmac',
        rule_id: 'rule:legacy-rc4',
        confidence: 1,
        remediation: '',
      },
    ],
    sssd_log_errors: [
      { description: 'Failed to authenticate', examples: ['[sssd] Authentication failure'] },
    ],
    sssd_config_snippet: '[sssd]\nservices = nss, pam\n',
    mac_denial_examples: [],
    problems: ['SSSD service is not responding'],
    warnings: ['Consider increasing the debug level'],
    matched_tids: [
      {
        tid_id: 'TID#1234',
        title: 'SSSD fails with RC4',
        url: 'https://example.com/tid/1234',
        description: 'AD rejects RC4 enctypes',
        log_patterns: ['rc4-hmac'],
        config_patterns: [],
        kb_id: '',
        plain_text: '',
        situation: '',
        resolution: '',
        cause: '',
        environment: '',
        created: '',
        changed: '',
        additional_information: '',
        evidence: ['Aug 18 10:00:00 host01 sssd: rc4-hmac rejected by dc01'],
      },
    ],
    timeline: [
      { timestamp: '10:00:01', message: 'service started', raw_log: 'sssd starting', occurrences: 1 },
    ],
    diagnostics: [
      { file: '/home/u/.sssd-inspector/rules.yaml', line: 4, message: 'rule "x": missing patterns or message, skipped', severity: 0 },
    ],
    temporal_clusters: [
      { description: 'repeated auth failures', event_count: 12, window_start: '10:00', window_end: '10:05', sample_raw_log: 'auth failure' },
    ],
    kb_suggestions: [
      { tid_id: 'TID#4321', title: 'Fuzzy match', url: 'https://example.com/tid/4321', score: 0.82, sample_line: 'some log line' },
    ],
    graph: { entities: [], findings: [], sources: [], edges: [] },
  };
  return { ...base, ...overrides } as ReportData;
}

export function makeInventory(overrides: Partial<DefinitionsInventory> = {}): DefinitionsInventory {
  return {
    user_root: '/home/u/.sssd-inspector',
    system_root: '/etc/sssd-inspector',
    files: [
      { path: '/home/u/.sssd-inspector/rules.yaml', kind: 'rules', scope: 'user', exists: true, is_dir: false, writable: true, rule_count: 2, article_count: 0, size_bytes: 128, mod_time: '2026-09-29 09:00:00' },
      { path: '/etc/sssd-inspector/rules.yaml', kind: 'rules', scope: 'system', exists: false, is_dir: false, writable: false, rule_count: 0, article_count: 0 },
    ],
    rules: [
      { rule: { name: 'legacy-rc4', severity: 'warning', category: 'crypto', files: ['sssd.conf'], patterns: ['rc4-hmac'], match: 'any', pattern_type: 'literal', message: 'RC4 present' }, file: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', line: 3 },
    ],
    rule_count: 1,
    article_count: 0,
    diagnostics: [],
    ...overrides,
  } as DefinitionsInventory;
}

export function makeValidation(valid: boolean): RuleValidationResult {
  return {
    label: valid ? 'valid' : 'invalid',
    valid,
    rule_count: valid ? 1 : 0,
    rules: [],
    diagnostics: valid ? [] : [{ file: '/home/u/.sssd-inspector/rules.yaml', line: 4, message: 'rule "x": missing patterns or message, skipped', severity: 0 }],
  } as unknown as RuleValidationResult;
}

export function makeDryRun(): RuleTestResult {
  return {
    target_path: '/tmp/supportconfig',
    total: 2,
    matched: 1,
    outcomes: [
      { rule: { name: 'fires', severity: 'critical', category: 'crypto', files: ['sssd.conf'], patterns: ['rc4-hmac'], match: 'any', pattern_type: 'literal', message: 'RC4 present' }, file: '/tmp/supportconfig/sssd.conf', scope: 'user', line: 3, matched: true, evidence: 'ldap_default_authtok = rc4-hmac' },
      { rule: { name: 'silent', severity: 'warning', category: 'dns', files: [], patterns: ['never-there'], match: 'any', pattern_type: 'literal', message: 'never' }, file: '/tmp/supportconfig/sssd.conf', scope: 'user', line: 9, matched: false },
    ],
    diagnostics: [],
  } as unknown as RuleTestResult;
}

/** CatalogInfo fixture: the embedded catalog is the default, no override. */
export function makeCatalog(overrides: Partial<CatalogInfo> = {}): CatalogInfo {
  return {
    source: 'SSSD upstream',
    version: '2.9',
    generated: '2026-01-01',
    option_count: 120,
    section_count: 8,
    available: true,
    effective: 'embedded',
    using_override: false,
    override_paths: ['/home/u/.sssd-inspector/catalog.json'],
    ...overrides,
  } as unknown as CatalogInfo;
}

/**
 * Event handlers the shell registered with the backend. The Wails runtime
 * delivers progress, definition warnings and OS drops through callbacks, so a
 * test that wants to exercise them has to reach the function the component
 * handed over — this registry is that seam.
 */
export interface BackendEventRegistry {
  progress: Array<(message: string, percentage: number) => void>;
  definitionsWarning: Array<(diagnostics: Array<{ file: string; line?: number; message: string; severity: number }>) => void>;
  fileDrop: Array<(paths: string[]) => void>;
}

/** Backend stub: every call resolves to a benign default unless overridden. */
export function mockBackend(overrides: Record<string, unknown> = {}) {
  const events: BackendEventRegistry = { progress: [], definitionsWarning: [], fileDrop: [] };
  return {
    events,
    backendAvailable: vi.fn(() => true),
    analyze: vi.fn(async () => makeReport()),
    openFileBrowser: vi.fn(async () => ''),
    savePdf: vi.fn(async () => 'report.pdf'),
    saveTxt: vi.fn(async () => 'report.txt'),
    saveJson: vi.fn(async () => 'report.json'),
    listDefinitions: vi.fn(async () => makeInventory()),
    validateRuleYaml: vi.fn(async () => makeValidation(true)),
    readRuleYaml: vi.fn(async () => ({ path: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', exists: false, bytes: 0, content: '' })),
    saveRuleYaml: vi.fn(async () => ({ path: '/home/u/.sssd-inspector/rules.yaml', scope: 'user', saved: true, bytes: 10, validation: makeValidation(true) })),
    testRulesAgainst: vi.fn(async () => makeDryRun()),
    getCatalogInfo: vi.fn(async () => makeCatalog()),
    openDefinitionsRoot: vi.fn(async () => undefined),
    openCatalogFile: vi.fn(async () => '/tmp/catalog.json'),
    installCatalog: vi.fn(async (_path: string, scope: string) => ({
      path: '/home/u/.sssd-inspector/catalog.json', scope, saved: true, bytes: 250086,
      backup: '/home/u/.sssd-inspector/catalog.json.bak',
      validation: { label: 'valid', valid: true, rule_count: 0, rules: [], diagnostics: [] },
    })),
    onAnalyzeProgress: vi.fn((cb: (message: string, percentage: number) => void) => {
      events.progress.push(cb);
      return () => undefined;
    }),
    onDefinitionsWarning: vi.fn((cb: (diagnostics: never[]) => void) => {
      events.definitionsWarning.push(cb as never);
      return () => undefined;
    }),
    onFileDrop: vi.fn((cb: (paths: string[]) => void) => {
      events.fileDrop.push(cb);
      return true;
    }),
    severityLabel: (severity: number | undefined) => (['warning', 'error', 'critical'] as const)[severity ?? 0] ?? 'warning',
    CANCELLED: 'cancelled',
    SCOPE_USER: 'user',
    SCOPE_SYSTEM: 'system',
    ...overrides,
  };
}


