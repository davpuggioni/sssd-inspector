// rulesTemplate — starter document and schema reference for the Studio editor.
//
// The starter is a real, valid rules.yaml: the Go loader requires name,
// severity, patterns and message, and skips anything else (with a diagnostic
// the Studio displays). The help text mirrors rules_engine.go, which is the
// authority — keep the two in step when the schema changes.

export const STARTER_RULES = `# sssd-inspector rules
# Every rule needs: name, severity, patterns, message.
rules:
  - name: "legacy-rc4-enctype"
    severity: warning          # critical | error | warning
    category: "crypto"
    files: ["sssd.conf", "krb5.conf"]
    message: "Legacy RC4 enctype found; modern AD rejects it."
    patterns:
      - "rc4-hmac"
    match: any                 # any (default) | all
    pattern_type: literal      # literal (default) | regex
`;

export const SCHEMA_HELP: Array<{ field: string; required: boolean; description: string }> = [
  { field: 'name', required: true, description: 'Unique rule id. A duplicate name is reported: both rules load and their findings cannot be told apart.' },
  { field: 'severity', required: true, description: 'critical | error | warning. Anything else is skipped.' },
  { field: 'patterns', required: true, description: 'One or more strings. Matched case-insensitively (literal) or through RE2 (regex).' },
  { field: 'message', required: true, description: 'Finding text, rendered as "[RULE: name] message" in the report.' },
  { field: 'category', required: false, description: 'Groups the finding (e.g. crypto, dns, kerberos).' },
  { field: 'files', required: false, description: 'File names scanned inside the supportconfig. Default: sssd.conf, sssd.txt, messages, messages.txt.' },
  { field: 'match', required: false, description: 'any (default) fires on the first matching line; all requires every pattern on one line.' },
  { field: 'pattern_type', required: false, description: 'literal (default) or regex (RE2, case-insensitive).' },
];
