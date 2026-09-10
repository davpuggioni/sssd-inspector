export namespace main {
	
	export class ConfigFinding {
	    severity: number;
	    category: string;
	    message: string;
	    source_path: string;
	    source_key: string;
	    source_line: number;
	    evidence: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigFinding(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.severity = source["severity"];
	        this.category = source["category"];
	        this.message = source["message"];
	        this.source_path = source["source_path"];
	        this.source_key = source["source_key"];
	        this.source_line = source["source_line"];
	        this.evidence = source["evidence"];
	    }
	}
	export class GraphEdge {
	    from: string;
	    to: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new GraphEdge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.kind = source["kind"];
	    }
	}
	export class GraphSourceNode {
	    id: string;
	    source_path: string;
	    source_line: number;
	    line_text: string;
	
	    static createFrom(source: any = {}) {
	        return new GraphSourceNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.source_path = source["source_path"];
	        this.source_line = source["source_line"];
	        this.line_text = source["line_text"];
	    }
	}
	export class GraphFindingNode {
	    id: string;
	    category: string;
	    message: string;
	    severity: number;
	    source_path?: string;
	    source_key?: string;
	    source_line?: number;
	    evidence?: string;
	    event_count?: number;
	
	    static createFrom(source: any = {}) {
	        return new GraphFindingNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.category = source["category"];
	        this.message = source["message"];
	        this.severity = source["severity"];
	        this.source_path = source["source_path"];
	        this.source_key = source["source_key"];
	        this.source_line = source["source_line"];
	        this.evidence = source["evidence"];
	        this.event_count = source["event_count"];
	    }
	}
	export class GraphEntity {
	    id: string;
	    kind: string;
	    label: string;
	    value?: string;
	    severity: number;
	    source_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new GraphEntity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.kind = source["kind"];
	        this.label = source["label"];
	        this.value = source["value"];
	        this.severity = source["severity"];
	        this.source_path = source["source_path"];
	    }
	}
	export class CorrelationGraph {
	    entities: GraphEntity[];
	    findings: GraphFindingNode[];
	    sources: GraphSourceNode[];
	    edges: GraphEdge[];
	
	    static createFrom(source: any = {}) {
	        return new CorrelationGraph(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entities = this.convertValues(source["entities"], GraphEntity);
	        this.findings = this.convertValues(source["findings"], GraphFindingNode);
	        this.sources = this.convertValues(source["sources"], GraphSourceNode);
	        this.edges = this.convertValues(source["edges"], GraphEdge);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExecutiveSummary {
	    health_score: number;
	    critical_count: number;
	    error_count: number;
	    warning_count: number;
	    problem_count: number;
	    log_error_count: number;
	    top_category: string;
	    top_category_hits: number;
	    headline: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecutiveSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.health_score = source["health_score"];
	        this.critical_count = source["critical_count"];
	        this.error_count = source["error_count"];
	        this.warning_count = source["warning_count"];
	        this.problem_count = source["problem_count"];
	        this.log_error_count = source["log_error_count"];
	        this.top_category = source["top_category"];
	        this.top_category_hits = source["top_category_hits"];
	        this.headline = source["headline"];
	    }
	}
	
	
	
	
	export class KBSuggestion {
	    tid_id: string;
	    title: string;
	    url: string;
	    score: number;
	    sample_line: string;
	
	    static createFrom(source: any = {}) {
	        return new KBSuggestion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tid_id = source["tid_id"];
	        this.title = source["title"];
	        this.url = source["url"];
	        this.score = source["score"];
	        this.sample_line = source["sample_line"];
	    }
	}
	export class TemporalCluster {
	    description: string;
	    event_count: number;
	    window_start: string;
	    window_end: string;
	    sample_raw_log: string;
	
	    static createFrom(source: any = {}) {
	        return new TemporalCluster(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.description = source["description"];
	        this.event_count = source["event_count"];
	        this.window_start = source["window_start"];
	        this.window_end = source["window_end"];
	        this.sample_raw_log = source["sample_raw_log"];
	    }
	}
	export class TimelineEvent {
	    timestamp: string;
	    message: string;
	    raw_log: string;
	    occurrences?: number;
	    samples?: string[];
	
	    static createFrom(source: any = {}) {
	        return new TimelineEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timestamp = source["timestamp"];
	        this.message = source["message"];
	        this.raw_log = source["raw_log"];
	        this.occurrences = source["occurrences"];
	        this.samples = source["samples"];
	    }
	}
	export class TIDConditions {
	    log_patterns: string[];
	    config_patterns: string[];
	    min_sssd_version: string;
	
	    static createFrom(source: any = {}) {
	        return new TIDConditions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.log_patterns = source["log_patterns"];
	        this.config_patterns = source["config_patterns"];
	        this.min_sssd_version = source["min_sssd_version"];
	    }
	}
	export class TIDArticle {
	    tid_id: string;
	    title: string;
	    url: string;
	    description: string;
	    log_patterns: string[];
	    config_patterns: string[];
	    kb_id: string;
	    plain_text: string;
	    situation: string;
	    resolution: string;
	    cause: string;
	    environment: string;
	    created: string;
	    changed: string;
	    additional_information: string;
	    conditions?: TIDConditions;
	
	    static createFrom(source: any = {}) {
	        return new TIDArticle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tid_id = source["tid_id"];
	        this.title = source["title"];
	        this.url = source["url"];
	        this.description = source["description"];
	        this.log_patterns = source["log_patterns"];
	        this.config_patterns = source["config_patterns"];
	        this.kb_id = source["kb_id"];
	        this.plain_text = source["plain_text"];
	        this.situation = source["situation"];
	        this.resolution = source["resolution"];
	        this.cause = source["cause"];
	        this.environment = source["environment"];
	        this.created = source["created"];
	        this.changed = source["changed"];
	        this.additional_information = source["additional_information"];
	        this.conditions = this.convertValues(source["conditions"], TIDConditions);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SSSDLogError {
	    description: string;
	    examples: string[];
	
	    static createFrom(source: any = {}) {
	        return new SSSDLogError(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.description = source["description"];
	        this.examples = source["examples"];
	    }
	}
	export class ReportData {
	    app_version: string;
	    timestamp: string;
	    support_case_id: string;
	    kernel_version: string;
	    sles_release: string;
	    scc_status: string;
	    summary: ExecutiveSummary;
	    hardware_manufacturer: string;
	    hardware_model: string;
	    hypervisor: string;
	    virtual_identity: string;
	    mac_type: string;
	    sssd_installed: boolean;
	    sssd_config_found: boolean;
	    sssd_service: string;
	    winbind_service: string;
	    nscd_status: string;
	    nscd_caching: string[];
	    sssd_packages: string[];
	    nsswitch_valid: boolean;
	    pam_sss_installed: boolean;
	    pam_gdpr_restricted: boolean;
	    hosts_issues: string[];
	    nameservers: string[];
	    search_domain: string;
	    time_service: string;
	    kerberos_realm: string;
	    keytab_found: boolean;
	    ad_provider_mode: boolean;
	    enumerate_issue: boolean;
	    use_fqdn_set: boolean;
	    ad_domain: string;
	    hostname: string;
	    config_findings: ConfigFinding[];
	    sssd_log_errors: SSSDLogError[];
	    sssd_config_snippet: string;
	    mac_denial_examples: string[];
	    problems: string[];
	    warnings: string[];
	    matched_tids: TIDArticle[];
	    timeline: TimelineEvent[];
	    temporal_clusters?: TemporalCluster[];
	    kb_suggestions?: KBSuggestion[];
	    graph: CorrelationGraph;
	
	    static createFrom(source: any = {}) {
	        return new ReportData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_version = source["app_version"];
	        this.timestamp = source["timestamp"];
	        this.support_case_id = source["support_case_id"];
	        this.kernel_version = source["kernel_version"];
	        this.sles_release = source["sles_release"];
	        this.scc_status = source["scc_status"];
	        this.summary = this.convertValues(source["summary"], ExecutiveSummary);
	        this.hardware_manufacturer = source["hardware_manufacturer"];
	        this.hardware_model = source["hardware_model"];
	        this.hypervisor = source["hypervisor"];
	        this.virtual_identity = source["virtual_identity"];
	        this.mac_type = source["mac_type"];
	        this.sssd_installed = source["sssd_installed"];
	        this.sssd_config_found = source["sssd_config_found"];
	        this.sssd_service = source["sssd_service"];
	        this.winbind_service = source["winbind_service"];
	        this.nscd_status = source["nscd_status"];
	        this.nscd_caching = source["nscd_caching"];
	        this.sssd_packages = source["sssd_packages"];
	        this.nsswitch_valid = source["nsswitch_valid"];
	        this.pam_sss_installed = source["pam_sss_installed"];
	        this.pam_gdpr_restricted = source["pam_gdpr_restricted"];
	        this.hosts_issues = source["hosts_issues"];
	        this.nameservers = source["nameservers"];
	        this.search_domain = source["search_domain"];
	        this.time_service = source["time_service"];
	        this.kerberos_realm = source["kerberos_realm"];
	        this.keytab_found = source["keytab_found"];
	        this.ad_provider_mode = source["ad_provider_mode"];
	        this.enumerate_issue = source["enumerate_issue"];
	        this.use_fqdn_set = source["use_fqdn_set"];
	        this.ad_domain = source["ad_domain"];
	        this.hostname = source["hostname"];
	        this.config_findings = this.convertValues(source["config_findings"], ConfigFinding);
	        this.sssd_log_errors = this.convertValues(source["sssd_log_errors"], SSSDLogError);
	        this.sssd_config_snippet = source["sssd_config_snippet"];
	        this.mac_denial_examples = source["mac_denial_examples"];
	        this.problems = source["problems"];
	        this.warnings = source["warnings"];
	        this.matched_tids = this.convertValues(source["matched_tids"], TIDArticle);
	        this.timeline = this.convertValues(source["timeline"], TimelineEvent);
	        this.temporal_clusters = this.convertValues(source["temporal_clusters"], TemporalCluster);
	        this.kb_suggestions = this.convertValues(source["kb_suggestions"], KBSuggestion);
	        this.graph = this.convertValues(source["graph"], CorrelationGraph);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	

}

