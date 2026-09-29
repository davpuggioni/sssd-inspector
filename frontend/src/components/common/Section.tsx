// Section — one collapsible report section.
//
// The legacy renderer emitted native <details>/<summary> blocks; React keeps
// them (keyboard accessible, zero JS) and only sets the initial `open` state
// imperatively so the user stays in control of the toggle afterwards.
import { useEffect, useRef } from 'react';
import type { ReactNode } from 'react';

export interface SectionProps {
  title: string;
  headingClass?: string;
  /** Extra class on the <details> element (e.g. 'corr-graph-section'). */
  className?: string;
  count?: number;
  defaultOpen?: boolean;
  children: ReactNode;
}

export function Section({ title, headingClass, className, count, defaultOpen = true, children }: SectionProps) {
  const ref = useRef<HTMLDetailsElement>(null);

  useEffect(() => {
    if (ref.current) {
      ref.current.open = defaultOpen;
    }
  }, [defaultOpen]);

  return (
    <details className={`report-section${className ? ` ${className}` : ''}`} ref={ref}>
      <summary>
        <h2 className={headingClass ?? ''}>
          {title}
          {typeof count === 'number' && <span className="section-count"> {count}</span>}
        </h2>
      </summary>
      <div className="section-body">{children}</div>
    </details>
  );
}

export default Section;
