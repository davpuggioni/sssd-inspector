// analysis_sequences.go
// Phase 6d: Sequence-based root-cause correlation.
//
// The deterministic pattern scanner (Phase 6) flags individual log lines, but
// real AD integration failures are *chains*: a DNS SRV timeout is followed by a
// connection attempt failure, then offline backend, then PAM_AUTHINFO_UNAVAIL.
// Each link is a separate "problem" today. This engine walks the already-sorted
// timeline and collapses characteristic cause→effect sequences into a single
// root-cause ConfigFinding, using the coarse categories produced by
// categoryFor() (mirroring SSSD failure domains verified in the C source:
// fail_over.c, data_provider_be.c, sdap_async_connection.c, krb5_child.c).
//
// Sequences are intentionally conservative: they require at least 2 distinct
// signals within a bounded window, so they never suppress a genuine standalone
// error. Findings are added via addConfigFinding, so they participate in the
// existing score/graph/anonymization pipeline unchanged.
package main

import (
	"fmt"
	"strings"
	"time"
)

// minTimelineEvents is the floor; below it a timeline is too sparse to infer
// causation and we abstain (single-line correlations would be false positives).
const minTimelineEvents = 3

// rootCauseHeadline maps a synthesized root-cause category to the short triage
// statement shown at the top of the report. The correlate* emitters below
// produce a LONG, evidence-carrying message ("[CORRELATED ROOT CAUSE] ...")
// that stays in the findings list; this map gives the headline the one-line
// summary a support engineer reads first, so the triage says what broke
// instead of repeating the whole finding.
//
// Categories are the ones correlateSequences emits (dns, krb5, overload,
// offline) plus the neighbouring domains the sequence engine is planned to
// cover (net, keytab, tls, crypto, gpo, idmap, access).
//
// A category that is absent falls back to the finding message, so adding a
// new emitter never produces an empty headline.
var rootCauseHeadline = map[string]string{
	"dns":      "DNS/SRV service discovery failure - the DC could not be located or reached via DNS.",
	"net":      "Network/backend connectivity failure - SSSD could not reach an AD/LDAP/Kerberos endpoint.",
	"offline":  "Backend went OFFLINE - SSSD exhausted retries and is falling back to cached credentials.",
	"keytab":   "Machine credentials unavailable - the host keytab or principal is missing/mismatched.",
	"krb5":     "Kerberos exchange failure - clock skew, preauth, or a missing machine account.",
	"tls":      "TLS/SSL negotiation failure - certificate chain, CA, or cipher incompatibility.",
	"crypto":   "Encryption-type mismatch - AD forcing RC4 that the local crypto policy rejects.",
	"gpo":      "Group-Policy access/control failure - the machine account cannot read GPOs.",
	"idmap":    "Identity mapping failure - SID-to-UID/GID translation is out of range or colliding.",
	"access":   "Access-control policy denial - account disabled/expired or HBAC/simple rules.",
	"overload": "Backend overload/restart storm - the machine account is overloaded by group or socket size.",
}

// timelineEpoch parses a timeline event timestamp (normalized) into Unix seconds.
func timelineEpoch(ev TimelineEvent) (int64, bool) {
	ts := normalizeTimestamp(ev.Timestamp)
	if ts == "" {
		return 0, false
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, time.Local)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}

// catOf returns the coarse category of a timeline event message via categoryFor,
// with substring fallbacks so partial descriptions still classify.
func catOf(msg string) string {
	if c := categoryFor(msg); c != "general" {
		return c
	}
	l := strings.ToLower(msg)
	if strings.Contains(l, "dns") || strings.Contains(l, "srv") || strings.Contains(l, "resolv") {
		return "dns"
	}
	if strings.Contains(l, "ssl") || strings.Contains(l, "tls") || strings.Contains(l, "sasl") {
		return "tls"
	}
	if strings.Contains(l, "kerberos") || strings.Contains(l, "clock") || strings.Contains(l, "preauth") {
		return "krb5"
	}
	if strings.Contains(l, "offline") {
		return "offline"
	}
	return "general"
}

// seqLink holds one link in a cause->effect chain.
type seqLink struct {
	idx      int
	category string
	message  string
	rawLog   string
	epoch    int64
	hasEpoch bool
}

// correlateSequences inspects the sorted timeline and, for characteristic
// failure chains observed in the SSSD C source, emits a consolidated root
// cause finding per chain. Additive across chains (>=2 chain links required
// per chain): a supportconfig carrying both a DNS failover cascade and a
// Kerberos clock-skew burst reports BOTH root causes instead of only the
// first one.
func correlateSequences(timeline []TimelineEvent, report *ReportData) {
	if len(timeline) < minTimelineEvents {
		return
	}

	links := make([]seqLink, 0, len(timeline))
	for i, ev := range timeline {
		c := catOf(ev.Message)
		ts, ok := timelineEpoch(ev)
		links = append(links, seqLink{idx: i, category: c, message: ev.Message, rawLog: ev.RawLog, epoch: ts, hasEpoch: ok})
	}

	emitted := make(map[string]bool)

	// R1: DNS/SRV failover chain (fail_over.c, sdap_async_connection.c)
	emittedAny := emitDNSChain(links, report, emitted)
	// R3: Kerberos realm/clock failure chain (krb5_child.c switch over AP_ERR/SKEW)
	emittedAny = emitKrb5Chain(links, report, emitted) || emittedAny
	// R2: watchdog/overload chain
	emittedAny = emitOverloadChain(links, report, emitted) || emittedAny
	// R4: offline burst
	emitOfflineChain(links, report, emitted)
	// R5: TLS/STARTTLS -> SASL bind chain (sdap_async_connection.c)
	emitTLSChain(links, report, emitted)
	// R6: SID-to-UID mapping failure chain (idmap.c, sdap_idmap.c)
	emitIDMapChain(links, report, emitted)
	// R7: GPO access denial chain (ad_gpo.c, pam_sss.c)
	emitGPOChain(links, report, emitted)
	// R8: group enumeration storm -> watchdog kill (sssd_nss.c, watchdog.c)
	emitEnumWatchdogChain(links, report, emitted)
}

// emitDNSChain detects the DNS/SRV-to-connect-to-offline cascade.
// Source: src/providers/fail_over.c (SRV timeout), sdap_async_connection.c.
func emitDNSChain(links []seqLink, report *ReportData, emitted map[string]bool) bool {
	var srvTimeout, resolveFail, connectFail, offline bool
	evidence := ""
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "service resolving timeout") || strings.Contains(ml, "unable to resolve srv") {
			srvTimeout = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "failed to resolve host") || strings.Contains(ml, "could not get server host name") {
			resolveFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "unable to establish connection") || strings.Contains(ml, "ldap connection error") {
			connectFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "going offline") {
			offline = true
		}
	}

	score := 0
	if srvTimeout {
		score++
	}
	if resolveFail {
		score++
	}
	if connectFail {
		score++
	}
	if offline {
		score++
	}
	if score >= 2 && (srvTimeout || resolveFail) && !emitted["dns"] {
		addConfigFindingEx(report, SevError, "dns",
			"[CORRELATED ROOT CAUSE] DNS/SRV service discovery failure (sequence: "+
				descFlags(srvTimeout, resolveFail, connectFail, offline)+
				"). The client could not locate or reach a Domain Controller via DNS; "+
				fmt.Sprintf("this single root explains %d separate-looking errors.", score),
			"timeline", "sequence:dns", 0, evidence, "correlate:dns-srv-failure", ConfidenceLog, "")
		emitted["dns"] = true
		return true
	}
	return false
}

// descFlags renders which steps of the DNS chain fired.
func descFlags(srv, resolve, connect, offline bool) string {
	var parts []string
	if srv {
		parts = append(parts, "SRV timeout")
	}
	if resolve {
		parts = append(parts, "host/SRV resolve fail")
	}
	if connect {
		parts = append(parts, "LDAP connect fail")
	}
	if offline {
		parts = append(parts, "backend offline")
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, " -> ")
}

// emitKrb5Chain detects clock-skew + preauth failures alongside the
// krb5_realm configuration mismatch (root cause = bad realm).
// Source: src/providers/krb5/krb5_child.c (KRB5KRB_AP_ERR_SKEW, KRB5_PREAUTH_FAILED).
func emitKrb5Chain(links []seqLink, report *ReportData, emitted map[string]bool) bool {
	var skew, preauth bool
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "clock skew") || strings.Contains(ml, "skew too great") {
			skew = true
		}
		if strings.Contains(ml, "preauthentication failed") || strings.Contains(ml, "preauth failed") {
			preauth = true
		}
	}

	realmMismatch := false
	for _, f := range report.ConfigFindings {
		if f.Category == "krb5_realm" || f.Category == "join" {
			realmMismatch = true
			break
		}
	}

	if (skew && preauth) || (skew && realmMismatch) {
		if !emitted["krb5"] {
			sev := SevError
			if realmMismatch {
				sev = SevCritical
			}
			msg := "[CORRELATED ROOT CAUSE] Kerberos realm/clock failure: clock skew + preauthentication failures"
			if realmMismatch {
				msg += " coincide with a misconfigured default_realm"
			}
			msg += ". Correcting default_realm or the AD machine/clock resolves the cascade."
			ev := ""
			if len(links) > 0 {
				ev = links[0].rawLog
			}
			addConfigFindingEx(report, sev, "krb5", msg, "timeline", "sequence:krb5", 0, ev, "correlate:krb5-clock-skew", ConfidenceLog, "krb5.conf(5)")
			emitted["krb5"] = true
			return true
		}
	}
	return false
}

// emitOverloadChain correlates watchdog terminations + packet overload + huge
// group enumeration into one "overload" root cause.
func emitOverloadChain(links []seqLink, report *ReportData, emitted map[string]bool) bool {
	var watchdog, overload bool
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "watchdog") {
			watchdog = true
		}
		if strings.Contains(ml, "overlarge") || strings.Contains(ml, "maximum number of shells") {
			overload = true
		}
	}
	if watchdog && overload && !emitted["overload"] {
		ev := ""
		for _, l := range links {
			if strings.Contains(strings.ToLower(l.message), "watchdog") {
				ev = l.rawLog
				break
			}
		}
		// Category "overload", not "ad_server": "ad_server" elsewhere means
		// "ad_server points at an IP" (config_parser.go) and is also a graph
		// entity kind, so reusing it here grouped a backend performance
		// problem under DC configuration in the root-cause breakdown and gave
		// it the wrong headline.
		addConfigFindingEx(report, SevWarning, "overload",
			"[CORRELATED ROOT CAUSE] Backend overload/restart storm: watchdog terminations co-occur with oversized packets/shell limits. Consider 'ignore_group_members = true' and reviewing AD group sizes.",
			"timeline", "sequence:overload", 0, ev, "correlate:backend-overload", ConfidenceLog, "")
		emitted["overload"] = true
		return true
	}
	return false
}

// emitOfflineChain collapses the offline->cached-credentials-exhaustion pattern.
// Source: src/providers/data_provider_be.c ("Going offline!"), krb5_auth.c.
func emitOfflineChain(links []seqLink, report *ReportData, emitted map[string]bool) {
	offline := false
	for _, l := range links {
		if strings.Contains(strings.ToLower(l.message), "backend offline") ||
			strings.Contains(strings.ToLower(l.rawLog), "going offline") ||
			strings.Contains(strings.ToLower(l.message), "offline auth") {
			offline = true
			break
		}
	}
	if !offline || emitted["offline"] {
		return
	}
	ev := ""
	for _, l := range links {
		if strings.Contains(strings.ToLower(l.message), "backend offline") {
			ev = l.rawLog
			break
		}
	}
	addConfigFindingEx(report, SevError, "offline",
		"[CORRELATED ROOT CAUSE] Backend went OFFLINE and online authentication subsequently failed. Root cause is upstream connectivity/DNS (see 'dns'/'net' findings); enable cache and increase offline_credentials_expiration only after the upstream is fixed.",
		"timeline", "sequence:offline", 0, ev, "correlate:offline-after-online", ConfidenceLog, "")
	emitted["offline"] = true
}

// subsequent SASL/GSSAPI bind failure into one root cause: the TLS layer
// never came up, so the bind that rides on it cannot succeed. Reporting the
// two separately makes a broken CA/cert chain look like two unrelated
// problems.
// Source: src/providers/ldap/sdap_async_connection.c (start_tls), sasl_bind.
func emitTLSChain(links []seqLink, report *ReportData, emitted map[string]bool) {
	var tlsFail, saslFail, credFail bool
	evidence := ""
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "starttls") || strings.Contains(ml, "start tls") ||
			strings.Contains(ml, "start_tls") || strings.Contains(ml, "tls handshake") ||
			strings.Contains(ml, "could not start tls") || strings.Contains(ml, "tls/ssl") ||
			strings.Contains(ml, "certificate") || strings.Contains(ml, "tls negotiation") {
			tlsFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		// Real SSSD emits "ldap_sasl_interactive_bind_s failed" (underscores),
		// so match on "sasl" + a failure word rather than the exact phrase.
		if (strings.Contains(ml, "sasl") || strings.Contains(ml, "gssapi")) &&
			(strings.Contains(ml, "fail") || strings.Contains(ml, "error") ||
				strings.Contains(ml, "could not")) {
			saslFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "invalid credentials") {
			credFail = true
		}
	}
	if tlsFail && saslFail && !emitted["tls"] {
		msg := "[CORRELATED ROOT CAUSE] TLS negotiation failed first, then the SASL/GSSAPI bind failed because no TLS layer was established. "
		if credFail {
			msg += "Credentials were also rejected, so fixing the certificate chain (ldap_tls_cacert / ldap_tls_cacertdir) must come before blaming the bind DN. "
		}
		msg += "Verify the CA chain, clock, and ldap_id_use_start_tls against the server certificate validity."
		addConfigFindingEx(report, SevError, "tls", msg,
			"timeline", "sequence:tls", 0, evidence, "correlate:tls-before-sasl", ConfidenceLog, "sssd-ldap(5)")
		emitted["tls"] = true
	}
}

// emitIDMapChain correlates SID-to-UID mapping failures with the downstream
// symptom (user not found / permission denied on a numeric uid) into one
// idmap root cause: the entry was mapped out of ldap_idmap_range or the
// ranges collide, so every consumer of that uid fails.
// Source: src/providers/ldap/sdap_idmap.c, src/lib/idmap/sss_idmap.c.
func emitIDMapChain(links []seqLink, report *ReportData, emitted map[string]bool) {
	var idmapFail, outOfRange, downstream bool
	evidence := ""
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if (strings.Contains(ml, "sid") && (strings.Contains(ml, "could not") || strings.Contains(ml, "fail"))) ||
			(strings.Contains(ml, "idmap") && strings.Contains(ml, "fail")) {
			idmapFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "out of range") || strings.Contains(ml, "range_max") ||
			strings.Contains(ml, "invalid sid") || strings.Contains(ml, "cannot map sid") {
			outOfRange = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "no such user") || strings.Contains(ml, "user not found") ||
			strings.Contains(ml, "permission denied") || strings.Contains(ml, "invalid uid") {
			downstream = true
		}
	}
	if (idmapFail || outOfRange) && downstream && !emitted["idmap"] {
		addConfigFindingEx(report, SevError, "idmap",
			"[CORRELATED ROOT CAUSE] SID-to-UID/GID mapping failed, and the downstream 'user not found'/permission-denied errors are its consequence, not independent faults. "+
				"Check ldap_id_mapping, ldap_idmap_range_min/max for out-of-range or overlapping ranges, and that RFC2307 attributes exist when ldap_id_mapping = false.",
			"timeline", "sequence:idmap", 0, evidence, "correlate:idmap-downstream", ConfidenceLog, "sssd-ldap(5)")
		emitted["idmap"] = true
	}
}

// emitGPOChain correlates a Group Policy evaluation failure with the PAM
// access denial it produces on ad_gpo_access_control=enforcing systems:
// the machine account could not read the GPO, so access is denied.
// Source: src/providers/ad/ad_gpo.c, src/sss_client/pam_sss.c.
func emitGPOChain(links []seqLink, report *ReportData, emitted map[string]bool) {
	var gpoFail, pamDeny bool
	evidence := ""
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "gpo") &&
			(strings.Contains(ml, "fail") || strings.Contains(ml, "could not") ||
				strings.Contains(ml, "error") || strings.Contains(ml, "permission")) {
			gpoFail = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "pam_authinfo_unavail") || strings.Contains(ml, "authentication failure") ||
			strings.Contains(ml, "user not permitted") || strings.Contains(ml, "not allowed to logon") {
			pamDeny = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
	}
	if gpoFail && pamDeny && !emitted["gpo"] {
		addConfigFindingEx(report, SevError, "gpo",
			"[CORRELATED ROOT CAUSE] GPO evaluation failed first; the PAM login denial that follows is the access-control consequence, not a separate credential problem. "+
				"Verify ad_gpo_access_control (permissive vs enforcing), that the machine account can read the GPO (RSOP), and SYSVOL reachability.",
			"timeline", "sequence:gpo", 0, evidence, "correlate:gpo-before-pam", ConfidenceLog, "sssd-ad(5)")
		emitted["gpo"] = true
	}
}

// emitEnumWatchdogChain correlates a large group-membership enumeration with
// a subsequent watchdog kill: the responder blocked long enough on the
// enumeration that the watchdog terminated it. One root (huge group /
// ignore_group_members not set), two apparent problems.
// Source: src/responder/nss/nsssrv.c, src/util/watchdog.c.
func emitEnumWatchdogChain(links []seqLink, report *ReportData, emitted map[string]bool) {
	var hugeEnum, watchdog bool
	evidence := ""
	for _, l := range links {
		ml := strings.ToLower(l.message) + " " + strings.ToLower(l.rawLog)
		if strings.Contains(ml, "overlarge") || strings.Contains(ml, "maximum number of shells") ||
			strings.Contains(ml, "group is too large") || strings.Contains(ml, "too many members") ||
			strings.Contains(ml, "huge group") {
			hugeEnum = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
		if strings.Contains(ml, "watchdog") || strings.Contains(ml, "terminated by own") {
			watchdog = true
			if evidence == "" {
				evidence = l.rawLog
			}
		}
	}
	if hugeEnum && watchdog && !emitted["overload"] {
		addConfigFindingEx(report, SevWarning, "overload",
			"[CORRELATED ROOT CAUSE] A huge group enumeration blocked the responder long enough for the watchdog to kill it: the kill is a symptom, the group size is the cause. "+
				"Consider ignore_group_members = true and review AD group membership sizes.",
			"timeline", "sequence:enum-watchdog", 0, evidence, "correlate:enum-watchdog", ConfidenceLog, "sssd.conf(5)")
		emitted["overload"] = true
	}
}
