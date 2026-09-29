// correlationGraph.ts — the force-directed layout for the correlation graph.
//
// TypeScript port of the former CorrelationGraph.js layout pass, unchanged in
// behaviour: the constants, the seeding spiral, the simulation and the collision
// relaxation are the same, so the picture a user sees is identical. What
// changed is that it is now a pure, typed, unit-testable module: no DOM, no
// string building, and the node shapes are plain data for the component to
// render.
import type { CorrelationGraph } from '../../api/backend';

export type GraphNodeKind = 'entity' | 'finding' | 'source';

export interface LayoutNode {
  id: string;
  kind: GraphNodeKind;
  label: string;
  /** radius in px, from NODE_R */
  r: number;
  fill: string;
  x: number;
  y: number;
  vx: number;
  vy: number;
  severity: number;
  // Detail-panel fields, carried through from the Go graph nodes.
  value?: string;
  category?: string;
  source_path?: string;
  source_key?: string;
  source_line?: number;
  evidence?: string;
  event_count?: number;
  line_text?: string;
}

export interface LayoutEdge {
  from: string;
  to: string;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
}

export interface GraphInput {
  entities: CorrelationGraph['entities'];
  findings: CorrelationGraph['findings'];
  sources: CorrelationGraph['sources'];
  edges: CorrelationGraph['edges'];
}

export const NODE_R: Record<GraphNodeKind, number> = { entity: 22, finding: 18, source: 12 };
/** Severity ordinal (types.go): 0 warning, 1 error, 2 critical. */
export const SEV_FILL: Record<number, string> = { 0: '#17a2b8', 1: '#ffc107', 2: '#dc3545' };
export const KIND_FILL: Record<'entity' | 'source', string> = { entity: '#4a90d9', source: '#6c757d' };

/** Canvas size for a graph of n nodes: it grows gently so dense graphs fit. */
export function canvasSize(nodeCount: number): { width: number; height: number } {
  return {
    width: Math.min(2000, Math.max(720, nodeCount * 36)),
    height: Math.min(1400, Math.max(480, nodeCount * 24)),
  };
}

/**
 * Run the force simulation and produce coordinates for every node.
 * Pure function — no DOM. Deterministic: the same input always yields the same
 * layout, which is what makes it testable.
 */
export function layoutGraph(graph: GraphInput, width: number, height: number): { nodes: LayoutNode[]; edges: LayoutEdge[] } {
  const cx = width / 2;
  const cy = height / 2;
  const total = graph.entities.length + graph.findings.length + graph.sources.length;
  // Golden-angle spiral: guarantees an even, non-overlapping seeding ring
  // regardless of node count (fixed 0.5 rad steps piled nodes on top of
  // each other once the count grew).
  const GOLDEN = 2.399963;
  const seedR = (kindScale: number) => {
    const base = Math.min(width, height) * 0.10;
    const spread = Math.min(width, height) * 0.42;
    return (i: number) => base + spread * Math.sqrt((i + 1) / Math.max(total, 1)) * kindScale;
  };
  const entityR = seedR(0.75);
  const findingR = seedR(1.0);
  const sourceR = seedR(1.2);

  const nodes: LayoutNode[] = graph.entities.map((e, i) => ({
    id: e.id,
    kind: 'entity',
    label: e.label,
    value: e.value,
    source_path: e.source_path,
    severity: 0,
    r: NODE_R.entity,
    fill: KIND_FILL.entity,
    x: cx + Math.cos(i * GOLDEN) * entityR(i),
    y: cy + Math.sin(i * GOLDEN) * entityR(i),
    vx: 0,
    vy: 0,
  }));

  const off = nodes.length;
  graph.findings.forEach((f, i) => {
    nodes.push({
      id: f.id,
      kind: 'finding',
      label: f.message,
      category: f.category,
      severity: f.severity,
      source_path: f.source_path,
      source_key: f.source_key,
      source_line: f.source_line,
      evidence: f.evidence,
      event_count: f.event_count,
      r: NODE_R.finding,
      fill: SEV_FILL[f.severity] ?? SEV_FILL[0],
      x: cx + Math.cos((off + i) * GOLDEN) * findingR(off + i),
      y: cy + Math.sin((off + i) * GOLDEN) * findingR(off + i),
      vx: 0,
      vy: 0,
    });
  });

  const off2 = nodes.length;
  graph.sources.forEach((s, i) => {
    nodes.push({
      id: s.id,
      kind: 'source',
      label: `${s.source_path}:${s.source_line}`,
      source_path: s.source_path,
      source_line: s.source_line,
      line_text: s.line_text,
      severity: 0,
      r: NODE_R.source,
      fill: KIND_FILL.source,
      x: cx + Math.cos((off2 + i) * GOLDEN) * sourceR(off2 + i),
      y: cy + Math.sin((off2 + i) * GOLDEN) * sourceR(off2 + i),
      vx: 0,
      vy: 0,
    });
  });

  const byId = new Map(nodes.map((n) => [n.id, n]));
  const linked = graph.edges
    .filter((e) => byId.has(e.from) && byId.has(e.to))
    .map((e) => ({ from: e.from, to: e.to, f: byId.get(e.from) as LayoutNode, t: byId.get(e.to) as LayoutNode }));

  // Repulsion scales with node count: a fixed constant lets dense graphs
  // collapse into overlaps even after 250 ticks.
  const REPULSION = 8000 * (1 + nodes.length / 12);
  const SPRING = 0.02;
  const SPRING_LEN = 90;
  const GRAVITY = 0.015;
  const DAMP = 0.85;
  const TICKS = 250;
  for (let t = 0; t < TICKS; t++) {
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const a = nodes[i];
        const b = nodes[j];
        let dx = a.x - b.x;
        let dy = a.y - b.y;
        let d2 = dx * dx + dy * dy;
        if (d2 < 1) {
          d2 = 1;
          dx = 1;
          dy = 0;
        }
        const f = REPULSION / d2;
        const d = Math.sqrt(d2);
        const fx = (dx / d) * f;
        const fy = (dy / d) * f;
        a.vx += fx;
        a.vy += fy;
        b.vx -= fx;
        b.vy -= fy;
      }
    }
    for (const e of linked) {
      const a = e.f;
      const b = e.t;
      const dx = b.x - a.x;
      const dy = b.y - a.y;
      const d = Math.sqrt(dx * dx + dy * dy) || 1;
      const f = (d - SPRING_LEN) * SPRING;
      const fx = (dx / d) * f;
      const fy = (dy / d) * f;
      a.vx += fx;
      a.vy += fy;
      b.vx -= fx;
      b.vy -= fy;
    }
    for (const n of nodes) {
      const dx = cx - n.x;
      const dy = cy - n.y;
      // Damped integration: velocity is decayed every tick, otherwise it
      // accumulates and nodes fly into the canvas corners and pile up.
      n.vx = (n.vx + dx * GRAVITY) * DAMP;
      n.vy = (n.vy + dy * GRAVITY) * DAMP;
      n.x += n.vx;
      n.y += n.vy;
    }
  }

  // Post-layout collision relaxation: push apart any node pairs that still
  // overlap after the force simulation (radius + label gutter).
  const GUTTER = 26;
  for (let pass = 0; pass < 80; pass++) {
    let moved = false;
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const a = nodes[i];
        const b = nodes[j];
        const minD = a.r + b.r + GUTTER;
        let dx = a.x - b.x;
        let dy = a.y - b.y;
        let d = Math.sqrt(dx * dx + dy * dy);
        if (d >= minD) {
          continue;
        }
        if (d < 0.01) {
          d = 0.01;
          dx = 0.01;
          dy = 0;
        }
        const push = (minD - d) / 2;
        const ux = dx / d;
        const uy = dy / d;
        a.x += ux * push;
        a.y += uy * push;
        b.x -= ux * push;
        b.y -= uy * push;
        moved = true;
      }
    }
    // keep relaxed nodes inside the canvas
    for (const n of nodes) {
      n.x = Math.max(n.r, Math.min(width - n.r, n.x));
      n.y = Math.max(n.r, Math.min(height - n.r, n.y));
    }
    if (!moved) {
      break;
    }
  }

  const edges: LayoutEdge[] = linked.map((e) => ({
    from: e.from,
    to: e.to,
    x1: e.f.x,
    y1: e.f.y,
    x2: e.t.x,
    y2: e.t.y,
  }));
  return { nodes, edges };
}

/** Adjacency of the laid-out graph, used to dim everything but the neighbourhood. */
export function adjacency(edges: LayoutEdge[]): Map<string, Set<string>> {
  const adj = new Map<string, Set<string>>();
  const link = (from: string, to: string) => {
    if (!adj.has(from)) adj.set(from, new Set());
    adj.get(from)!.add(to);
  };
  for (const e of edges) {
    link(e.from, e.to);
    link(e.to, e.from);
  }
  return adj;
}

