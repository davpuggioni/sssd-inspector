// SystemInfoTable — the "System & Virtualization" / "Authentication" / "sssd.conf"
// part of the report. Values come straight from the Go report; React escapes
// them on render, so supportconfig content can never inject markup.
import type { ReportData } from '../../api/backend';
import { IniSnippet } from './IniSnippet';

function YesNo({ value }: { value: boolean }) {
  return value ? <span className="success">Yes</span> : <span className="fail">No</span>;
}

function ListItems({ items }: { items?: string[] }) {
  if (!items || items.length === 0) {
    return null;
  }
  return (
    <>
      {items.map((item, index) => (
        <li key={`${index}-${item}`}>{item}</li>
      ))}
    </>
  );
}

function HostsStatusBadge({ status }: { status: string }) {
  if (status === 'missing_on_host') {
    return <span className="status-badge error">Missing from Host</span>;
  }
  if (status === 'not_collected') {
    return <span className="status-badge warn">Not Collected</span>;
  }
  return <span className="status-badge success">Present</span>;
}

export interface SystemInfoTableProps {
  report: ReportData;
}

export function SystemInfoTable({ report }: SystemInfoTableProps) {
  return (
    <table className="info-table">
      <tbody>
        <tr>
          <td colSpan={2} className="section-title">System &amp; Virtualization Information</td>
        </tr>
        <tr><th>OS Release</th><td>{report.sles_release || 'Unknown'}</td></tr>
        <tr><th>Kernel Version</th><td>{report.kernel_version || 'Unknown'}</td></tr>
        <tr><th>SCC Status</th><td>{report.scc_status || 'Unknown'}</td></tr>
        <tr><th>Hardware</th><td>{report.hardware_manufacturer} {report.hardware_model}</td></tr>
        <tr>
          <th>Virtualization</th>
          <td>{report.hypervisor || 'Unknown'} (Identity: {report.virtual_identity || 'Unknown'})</td>
        </tr>
        <tr><th>MAC Security</th><td>{report.mac_type || 'Unknown/None'}</td></tr>

        <tr>
          <td colSpan={2} className="section-title">Base Authentication Services</td>
        </tr>
        <tr><th>SSSD Installed</th><td><YesNo value={report.sssd_installed} /></td></tr>
        <tr>
          <th>SSSD Packages</th>
          <td>
            {report.sssd_packages && report.sssd_packages.length > 0
              ? <pre className="pkg-list">{report.sssd_packages.join('\n')}</pre>
              : 'None Detected'}
          </td>
        </tr>
        <tr><th>SSSD Config Found</th><td><YesNo value={report.sssd_config_found} /></td></tr>
        <tr><th>SSSD Service Status</th><td>{report.sssd_service || 'Unknown'}</td></tr>
        <tr><th>Winbind Service</th><td>{report.winbind_service || 'Unknown'}</td></tr>
        <tr><th>nscd Status</th><td>{report.nscd_status || 'Unknown'}</td></tr>
        <tr><th>nscd Caching</th><td><ListItems items={report.nscd_caching} /></td></tr>

        <tr>
          <td colSpan={2} className="section-title">Authentication Configuration</td>
        </tr>
        <tr><th>nsswitch.conf Valid</th><td><YesNo value={report.nsswitch_valid} /></td></tr>
        <tr><th>pam_sss Installed</th><td><YesNo value={report.pam_sss_installed} /></td></tr>
        <tr><th>GDPR Restricted Mode</th><td><YesNo value={report.pam_gdpr_restricted} /></td></tr>
        <tr><th>Hosts File Status</th><td><HostsStatusBadge status={report.hosts_file_status} /></td></tr>
        <tr><th>Hosts File Issues</th><td><ListItems items={report.hosts_issues} /></td></tr>
        <tr><th>Nameservers</th><td><ListItems items={report.nameservers} /></td></tr>
        <tr><th>Search Domain</th><td>{report.search_domain || 'None'}</td></tr>
        <tr><th>Time Service</th><td>{report.time_service || 'Unknown'}</td></tr>
        <tr><th>Kerberos Realm</th><td>{report.kerberos_realm || 'Not configured'}</td></tr>
        <tr><th>Keytab Found</th><td><YesNo value={report.keytab_found} /></td></tr>

        <tr>
          <td colSpan={2} className="section-title">Advanced SSSD Configuration</td>
        </tr>
        <tr><th>AD Provider Mode</th><td><YesNo value={report.ad_provider_mode} /></td></tr>
        <tr><th>Enumerate Issue</th><td><YesNo value={report.enumerate_issue} /></td></tr>
        <tr><th>Use FQDN Set</th><td><YesNo value={report.use_fqdn_set} /></td></tr>
        {report.catalog_provenance ? (
          <tr><th>Option Catalog</th><td>{report.catalog_provenance}</td></tr>
        ) : null}

        <tr>
          <td colSpan={2} className="section-title">Configuration Files</td>
        </tr>
        <tr>
          <td colSpan={2}>
            {report.sssd_config_snippet ? (
              <details open>
                <summary style={{ cursor: 'pointer', fontWeight: 'bold', padding: '5px 0' }}>
                  View sssd.conf
                </summary>
                <IniSnippet content={report.sssd_config_snippet} />
              </details>
            ) : 'sssd.conf not found / not readable'}
          </td>
        </tr>
      </tbody>
    </table>
  );
}

export default SystemInfoTable;
