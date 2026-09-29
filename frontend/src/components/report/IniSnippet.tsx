// IniSnippet — sssd.conf viewer with a lightweight INI highlighter.
//
// React port of the legacy highlightIni(): sections, comments, keys and values
// become spans while indentation and blank lines stay verbatim. Rendering real
// elements (instead of an HTML string injected with dangerouslySetInnerHTML)
// means the report content is escaped by React by construction.
import type { ReactNode } from 'react';

const KEY_RE = /^([A-Za-z_][A-Za-z0-9_.-]*)(\s*=\s*)(.*)$/;

function highlightLine(line: string, index: number): ReactNode {
  const trimmed = line.trim();
  if (trimmed === '' || trimmed.startsWith('#') || trimmed.startsWith(';')) {
    return (
      <span className="ini-comment" key={index}>
        {line}
        {'\n'}
      </span>
    );
  }

  const section = trimmed.match(/^\[([^\]]*)\]$/);
  if (section) {
    const start = line.indexOf('[');
    const end = line.lastIndexOf(']');
    return (
      <span key={index}>
        {line.slice(0, start)}
        <span className="ini-section">[{section[1]}]</span>
        {line.slice(end + 1)}
        {'\n'}
      </span>
    );
  }

  const kv = line.match(KEY_RE);
  if (kv) {
    return (
      <span key={index}>
        <span className="ini-key">{kv[1]}</span>
        <span className="ini-op">{kv[2]}</span>
        <span className="ini-value">{kv[3]}</span>
        {'\n'}
      </span>
    );
  }

  return (
    <span key={index}>
      {line}
      {'\n'}
    </span>
  );
}

export interface IniSnippetProps {
  content: string;
}

export function IniSnippet({ content }: IniSnippetProps) {
  return (
    <pre className="log-block ini-snippet" style={{ whiteSpace: 'pre-wrap', marginTop: '10px' }}>
      {content.split('\n').map(highlightLine)}
    </pre>
  );
}

export default IniSnippet;
