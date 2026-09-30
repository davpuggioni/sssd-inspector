// SystemInfoTable.test.tsx — the hosts-file badge must not assert more than
// the supportconfig actually showed.
//
// Four states, and they are not interchangeable: a file that was never
// collected is not evidence of anything, and a file that parses but lacks a
// loopback line is not corrupt. Rendering any of them as "Present" would be a
// claim the analysis cannot support.
import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SystemInfoTable } from '../components/report/SystemInfoTable';
import { makeReport } from './helpers';

function badgeFor(status: string) {
  render(<SystemInfoTable report={{ ...makeReport(), hosts_file_status: status } as never} />);
  const row = screen.getByText('Hosts File Status').closest('tr');
  const badge = row?.querySelector('.status-badge') as HTMLElement;
  return { text: badge.textContent, modifier: badge.className };
}

describe('SystemInfoTable — hosts file status', () => {
  it('reports a collected, parsable file as present', () => {
    expect(badgeFor('present').text).toBe('Present');
  });

  it('reports a file absent from the host as missing', () => {
    const badge = badgeFor('missing_on_host');
    expect(badge.text).toBe('Missing from Host');
    expect(badge.modifier).toContain('error');
  });

  it('does not claim a file it never saw is present', () => {
    const badge = badgeFor('not_collected');
    expect(badge.text).toBe('Not Collected');
    // A warning, not an error: the file may well be fine, the bundle is
    // simply silent about it.
    expect(badge.modifier).toContain('warn');
    expect(badge.text).not.toBe('Present');
  });

  it('separates an unreadable file from a parsable one', () => {
    const badge = badgeFor('malformed');
    expect(badge.text).toBe('Present but Malformed');
    expect(badge.modifier).toContain('error');
  });
});
