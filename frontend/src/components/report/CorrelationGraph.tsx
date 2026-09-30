// CorrelationGraph — the interactive graph view.
//
// TypeScript/React port of the former CorrelationGraph.js renderer. Same
// features (force layout, severity filter, hover dimming, click detail, node
// drag, wheel zoom and background pan) and the same CSS classes, so the
// existing report stylesheet still applies. What changed: the SVG and the
// detail panel are React elements instead of HTML strings, so node labels and
// evidence coming from supportconfig content are escaped by the framework, and
// the layout maths lives in correlationGraph.ts where it can be unit-tested.
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { MouseEvent as ReactMouseEvent } from 'react';
// The payload type is aliased: the component and the Go model would otherwise
// share the name CorrelationGraph.
import type { CorrelationGraph as GraphPayload } from '../../api/backend';
import { severityLabel } from '../../api/backend';
import { listOf } from '../../utils/payload';
import { SEV_FILL, adjacency, canvasSize, layoutGraph } from './correlationGraph';
import type { LayoutEdge, LayoutNode } from './correlationGraph';

const MAX_LABEL = 34;
const MIN_ZOOM = 0.5;
const MAX_ZOOM = 8;

interface ViewBox {
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface CorrelationGraphProps {
  graph: GraphPayload;
}

function shortLabel(s: string): string {
  return s.length > MAX_LABEL ? `${s.slice(0, MAX_LABEL - 1)}…` : s;
}

function shapeOf(node: LayoutNode) {
  const r = node.r;
  if (node.kind === 'finding') {
    return <polygon className="corr-shape" points={`0,${-r} ${r},0 0,${r} ${-r},0`} />;
  }
  if (node.kind === 'source') {
    return <rect className="corr-shape" x={-r} y={-r} width={r * 2} height={r * 2} />;
  }
  return <circle className="corr-shape" r={r} />;
}

function SeverityFilters(props: { enabled: Set<number>; toggle: (sev: number) => void }) {
  const { enabled, toggle } = props;
  return (
    <div className="corr-filters">
      <span className="corr-filter-label">Show:</span>
      {[2, 1, 0].map((sev) => (
        <label className="corr-filter" key={sev}>
          <input type="checkbox" checked={enabled.has(sev)} onChange={() => toggle(sev)} />
          <span className="corr-sev-dot" style={{ background: SEV_FILL[sev] }} />
          {sev === 0 ? 'Warning+' : sev === 1 ? 'Error' : 'Critical'}
        </label>
      ))}
    </div>
  );
}

function DetailPanel({ node }: { node: LayoutNode }) {
  return (
    <div className="corr-detail" data-testid="corr-detail">
      <h4>{node.label}</h4>
      <div className="corr-detail-kind">
        {node.kind}
        {node.category ? ` · ${node.category}` : ''}
      </div>
      {node.kind === 'finding' && (
        <>
          <div className={`corr-detail-sev sev-${severityLabel(node.severity)}`}>
            {severityLabel(node.severity).toUpperCase()}
          </div>
          {node.source_path ? (
            <div>
              <strong>Source:</strong> {node.source_path}
              {node.source_line ? `:${node.source_line}` : ''}
            </div>
          ) : null}
          {node.source_key ? <div><strong>Key:</strong> {node.source_key}</div> : null}
          {node.evidence ? <div className="corr-detail-evidence">{node.evidence}</div> : null}
          {node.event_count ? <div><strong>Events:</strong> {node.event_count}</div> : null}
        </>
      )}
      {node.kind === 'source' && (
        <>
          <div><strong>File:</strong> {node.source_path}:{node.source_line}</div>
          {node.line_text ? <div className="corr-detail-evidence">{node.line_text}</div> : null}
        </>
      )}
      {node.kind === 'entity' && (
        <>
          {node.value ? <div><strong>Value:</strong> {node.value}</div> : null}
          {node.source_path ? <div><strong>From:</strong> {node.source_path}</div> : null}
        </>
      )}
    </div>
  );
}

export function CorrelationGraph({ graph }: CorrelationGraphProps) {
  const nodeCount = listOf(graph.entities).length + listOf(graph.findings).length + listOf(graph.sources).length;
  const { width, height } = useMemo(() => canvasSize(nodeCount), [nodeCount]);
  const { nodes, edges } = useMemo(
    () => layoutGraph(
      {
        entities: listOf(graph.entities),
        findings: listOf(graph.findings),
        sources: listOf(graph.sources),
        edges: listOf(graph.edges),
      },
      width,
      height,
    ),
    [graph, width, height],
  );
  const adj = useMemo(() => adjacency(edges), [edges]);
  const home = useMemo<ViewBox>(() => ({ x: 0, y: 0, w: width, h: height }), [width, height]);

  const [view, setView] = useState<ViewBox>(home);
  const [hovered, setHovered] = useState<string | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [visible, setVisible] = useState<Set<number>>(() => new Set([0, 1, 2]));
  // A dragged node is kept in state so it re-renders; the layout itself comes
  // from useMemo and is never mutated in place.
  const [dragged, setDragged] = useState<{ id: string; x: number; y: number } | null>(null);

  const svgRef = useRef<SVGSVGElement>(null);
  const dragState = useRef<{ id: string; offsetX: number; offsetY: number } | null>(null);
  const panState = useRef<{ px: number; py: number; vx: number; vy: number } | null>(null);

  const positioned = useMemo<LayoutNode[]>(
    () => (dragged ? nodes.map((n) => (n.id === dragged.id ? { ...n, x: dragged.x, y: dragged.y } : n)) : nodes),
    [nodes, dragged],
  );
  const byId = useMemo(() => new Map(positioned.map((n) => [n.id, n])), [positioned]);
  const selectedNode = selected ? byId.get(selected) : undefined;

  const toSvgPoint = useCallback((clientX: number, clientY: number) => {
    const svg = svgRef.current;
    // Guard the methods, not just the element. getScreenCTM and
    // createSVGPoint are part of the SVG 2 DOM, which a real browser
    // implements and a headless DOM (jsdom) does not: there the optional
    // chain on `svg` passes but the call throws, and an exception thrown
    // inside a mousemove handler is reported as an unhandled error that
    // outlives the test. Returning null degrades the interaction to "no
    // drag in progress" instead.
    if (!svg || typeof svg.getScreenCTM !== 'function' || typeof svg.createSVGPoint !== 'function') {
      return null;
    }
    const ctm = svg.getScreenCTM();
    if (!ctm) {
      return null;
    }
    const point = svg.createSVGPoint();
    point.x = clientX;
    point.y = clientY;
    return point.matrixTransform(ctm.inverse());
  }, []);

  // Pure viewBox maths, shared by the wheel (zoom at the pointer) and the
  // toolbar / keyboard (zoom at the centre).
  const zoomView = useCallback((current: ViewBox, px: number, py: number, factor: number): ViewBox => {
    const nextW = current.w * factor;
    if (nextW < home.w / MAX_ZOOM || nextW > home.w / MIN_ZOOM) {
      return current;
    }
    return {
      x: px - (px - current.x) * factor,
      y: py - (py - current.y) * factor,
      w: nextW,
      h: current.h * factor,
    };
  }, [home.w]);

  const zoomAt = useCallback((px: number, py: number, factor: number) => {
    setView((current) => zoomView(current, px, py, factor));
  }, [zoomView]);

  const zoomBy = useCallback((factor: number) => {
    setView((current) => zoomView(current, current.x + current.w / 2, current.y + current.h / 2, factor));
  }, [zoomView]);

  const reset = useCallback(() => setView(home), [home]);

  // Node drag and background pan track the pointer at the window level: while
  // dragging, the pointer regularly leaves the SVG.
  useEffect(() => {
    const onMove = (event: MouseEvent) => {
      const drag = dragState.current;
      if (drag) {
        const point = toSvgPoint(event.clientX, event.clientY);
        if (point) {
          setDragged({ id: drag.id, x: point.x - drag.offsetX, y: point.y - drag.offsetY });
        }
        return;
      }
      const pan = panState.current;
      const svg = svgRef.current;
      if (!pan || !svg) {
        return;
      }
      const rect = svg.getBoundingClientRect();
      setView((current) => ({
        ...current,
        x: pan.vx - (event.clientX - pan.px) * (current.w / rect.width),
        y: pan.vy - (event.clientY - pan.py) * (current.h / rect.height),
      }));
    };
    const onUp = () => {
      dragState.current = null;
      panState.current = null;
    };
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
    return () => {
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };
  }, [toSvgPoint]);

  const startNodeDrag = (event: ReactMouseEvent, node: LayoutNode) => {
    const point = toSvgPoint(event.clientX, event.clientY);
    if (!point) {
      return;
    }
    dragState.current = { id: node.id, offsetX: point.x - node.x, offsetY: point.y - node.y };
    event.preventDefault();
    event.stopPropagation();
  };

  const startPan = (event: ReactMouseEvent) => {
    panState.current = { px: event.clientX, py: event.clientY, vx: view.x, vy: view.y };
    event.preventDefault();
  };

  const toggleSeverity = (sev: number) => setVisible((current) => {
    const next = new Set(current);
    if (next.has(sev)) {
      next.delete(sev);
    } else {
      next.add(sev);
    }
    return next;
  });

  const neighbours = hovered ? adj.get(hovered) : undefined;
  const dimmed = (id: string) => Boolean(hovered) && id !== hovered && !neighbours?.has(id);
  /** A severity the user switched off hides the node entirely (exact match). */
  const filteredOut = (node: LayoutNode) => !visible.has(node.severity);


  return (
    <div className="corr-graph-wrap">
      <SeverityFilters enabled={visible} toggle={toggleSeverity} />

      <div className="corr-zoombar">
        <button type="button" className="corr-zoom-btn" title="Zoom in (+)" onClick={() => zoomBy(1 / 1.3)}>＋</button>
        <button type="button" className="corr-zoom-btn" title="Zoom out (−)" onClick={() => zoomBy(1.3)}>－</button>
        <button type="button" className="corr-zoom-btn" title="Reset view (0)" onClick={reset}>⤢</button>
        <span className="corr-zoom-hint">wheel = zoom · drag background = pan · drag node = move</span>
      </div>

      <div
        className="corr-canvas"
        tabIndex={0}
        role="application"
        aria-label="Correlation graph"
        onKeyDown={(event) => {
          if (event.key === '+' || event.key === '=') {
            zoomBy(1 / 1.25);
          } else if (event.key === '-' || event.key === '_') {
            zoomBy(1.25);
          } else if (event.key === '0') {
            reset();
          }
        }}
        onMouseDown={startPan}
        onMouseMove={(event) => {
          const group = (event.target as Element).closest('.corr-node');
          setHovered(group ? group.getAttribute('data-id') : null);
        }}
        onMouseLeave={() => setHovered(null)}
        onClick={(event) => {
          const group = (event.target as Element).closest('.corr-node');
          setSelected(group ? group.getAttribute('data-id') : null);
        }}
      >
        <svg
          ref={svgRef}
          className="corr-svg"
          viewBox={`${view.x} ${view.y} ${view.w} ${view.h}`}
          preserveAspectRatio="xMidYMid meet"
          xmlns="http://www.w3.org/2000/svg"
          onWheel={(event) => {
            const point = toSvgPoint(event.clientX, event.clientY);
            if (point) {
              zoomAt(point.x, point.y, event.deltaY > 0 ? 1.15 : 1 / 1.15);
            }
          }}
        >
          <defs>
            <marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
              <path d="M 0 0 L 10 5 L 0 10 z" fill="#888" />
            </marker>
          </defs>
          {edges.map((edge: LayoutEdge) => {
            const touching = hovered === edge.from || hovered === edge.to;
            const from = byId.get(edge.from);
            const to = byId.get(edge.to);
            return (
              <line
                key={`${edge.from}->${edge.to}`}
                className="corr-edge"
                x1={from?.x ?? edge.x1}
                y1={from?.y ?? edge.y1}
                x2={to?.x ?? edge.x2}
                y2={to?.y ?? edge.y2}
                stroke={hovered && touching ? '#aaa' : '#555'}
                strokeWidth="1.2"
                opacity={hovered ? (touching ? 1 : 0.08) : 1}
                markerEnd="url(#arrow)"
              />
            );
          })}
          {positioned.map((node) => (
            <g
              key={node.id}
              className={`corr-node corr-${node.kind}`}
              data-id={node.id}
              data-severity={node.severity}
              transform={`translate(${node.x.toFixed(1)},${node.y.toFixed(1)})`}
              opacity={dimmed(node.id) ? 0.15 : 1}
              style={filteredOut(node) ? { display: 'none' } : undefined}
              onMouseDown={(event) => startNodeDrag(event, node)}
            >
              <title>{node.label}</title>
              {shapeOf(node)}
              <text className="corr-node-label" x={node.r + 6} y="4" fill="#ddd" fontSize="11">
                {shortLabel(node.label)}
              </text>
            </g>
          ))}
        </svg>
      </div>

      {selectedNode ? <DetailPanel node={selectedNode} /> : null}
    </div>
  );
}

export default CorrelationGraph;

