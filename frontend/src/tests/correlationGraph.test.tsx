// correlationGraph.test.tsx — the graph layout and the component that draws it.
//
// The layout was previously unreachable from tests: it lived inside an
// imperative renderer that needed a real DOM and a host element. Split out as a
// pure function it can be checked directly — determinism, nothing escaping the
// canvas, no collapsed pairs — and the component is checked through the DOM.
import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import type { CorrelationGraph } from '../api/backend';
import { CorrelationGraph as CorrelationGraphView } from '../components/report/CorrelationGraph';
import { KIND_FILL, NODE_R, SEV_FILL, adjacency, canvasSize, layoutGraph } from '../components/report/correlationGraph';

function makeGraph(): CorrelationGraph {
  return {
    entities: [
      { id: 'e1', kind: 'ad_domain', label: 'ad.corp.example', value: 'ad.corp.example', severity: 0 },
      { id: 'e2', kind: 'realm', label: 'CORP.EXAMPLE', severity: 0 },
    ],
    findings: [
      { id: 'f1', category: 'crypto', message: 'RC4 enctype found', severity: 2, source_path: 'sssd.conf', source_line: 12, evidence: 'ldap_default_authtok = rc4-hmac' },
      { id: 'f2', category: 'dns', message: 'No nameservers', severity: 0, source_path: 'resolv.conf', source_line: 1 },
    ],
    sources: [
      { id: 's1', source_path: 'sssd.conf', source_line: 12, line_text: 'ldap_default_authtok = rc4-hmac' },
    ],
    edges: [
      { from: 'f1', to: 's1', kind: 'evidence' },
      { from: 'f1', to: 'e1', kind: 'about' },
      { from: 'e1', to: 'e2', kind: 'related' },
      { from: 'f2', to: 'e1', kind: 'about' },
    ],
  } as unknown as CorrelationGraph;
}

describe('canvasSize', () => {
  it('grows with the node count and stays inside its bounds', () => {
    expect(canvasSize(0).width).toBe(720);
    expect(canvasSize(0).height).toBe(480);
    const dense = canvasSize(500);
    expect(dense.width).toBeLessThanOrEqual(2000);
    expect(dense.height).toBeLessThanOrEqual(1400);
  });
});

describe('layoutGraph', () => {
  const graph = makeGraph();

  it('is deterministic: the same input yields the same picture', () => {
    const a = layoutGraph(graph, 800, 600);
    const b = layoutGraph(graph, 800, 600);
    expect(a.nodes.map((n) => [n.x, n.y])).toEqual(b.nodes.map((n) => [n.x, n.y]));
  });

  it('lays out every node with the radius of its kind', () => {
    const { nodes } = layoutGraph(graph, 800, 600);
    expect(nodes).toHaveLength(5);
    for (const node of nodes) {
      expect(node.r).toBe(NODE_R[node.kind]);
      expect(Number.isFinite(node.x)).toBe(true);
      expect(Number.isFinite(node.y)).toBe(true);
    }
  });

  it('labels a source node with its file and line', () => {
    const { nodes } = layoutGraph(graph, 800, 600);
    expect(nodes.find((n) => n.id === 's1')?.label).toBe('sssd.conf:12');
  });

  it('keeps every node inside the canvas', () => {
    const { width, height } = canvasSize(5);
    const { nodes } = layoutGraph(graph, width, height);
    for (const node of nodes) {
      expect(node.x).toBeGreaterThanOrEqual(node.r - 0.001);
      expect(node.x).toBeLessThanOrEqual(width - node.r + 0.001);
      expect(node.y).toBeGreaterThanOrEqual(node.r - 0.001);
      expect(node.y).toBeLessThanOrEqual(height - node.r + 0.001);
    }
  });

  it('does not let nodes collapse on top of each other', () => {
    const { nodes } = layoutGraph(graph, 800, 600);
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const d = Math.hypot(nodes[i].x - nodes[j].x, nodes[i].y - nodes[j].y);
        expect(d).toBeGreaterThan((nodes[i].r + nodes[j].r) * 0.5);
      }
    }
  });

  it('drops edges that point at nodes which are not in the graph', () => {
    const { edges } = layoutGraph(
      { ...graph, edges: [...graph.edges, { from: 'f1', to: 'ghost', kind: 'evidence' }] },
      800,
      600,
    );
    expect(edges).toHaveLength(graph.edges.length);
  });

  it('builds a symmetric adjacency map for the hover dimming', () => {
    const { edges } = layoutGraph(graph, 800, 600);
    const adj = adjacency(edges);
    expect(adj.get('f1')?.has('s1')).toBe(true);
    expect(adj.get('s1')?.has('f1')).toBe(true);
  });
});

describe('CorrelationGraph (component)', () => {
  it('draws one shape per node and one line per edge', () => {
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    expect(container.querySelectorAll('.corr-node')).toHaveLength(5);
    expect(container.querySelectorAll('.corr-edge')).toHaveLength(4);
    // Entities are circles, findings diamonds, sources squares.
    expect(container.querySelectorAll('.corr-entity circle')).toHaveLength(2);
    expect(container.querySelectorAll('.corr-finding polygon')).toHaveLength(2);
    expect(container.querySelectorAll('.corr-source rect')).toHaveLength(1);
  });

  it('fills and outlines every shape, so no node renders black on a dark canvas', () => {
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);

    // Regression: the port to React dropped the fill and stroke attributes the
    // imperative renderer had, leaving every shape on the SVG default of black
    // over a near-black .corr-svg. The layout already computes a fill per node;
    // it simply stopped being rendered, so nothing in the tests failed.
    const shapes = Array.from(container.querySelectorAll('.corr-shape'));
    expect(shapes).toHaveLength(5);
    for (const shape of shapes) {
      const fill = shape.getAttribute('fill');
      expect(fill, `${shape.tagName} has no fill attribute`).toBeTruthy();
      // Black on the #1e1e24 canvas is the exact regression being pinned.
      expect(fill?.toLowerCase(), `${shape.tagName} is black on a dark canvas`).not.toBe('#000');
      expect(fill?.toLowerCase()).not.toBe('black');
      expect(shape.getAttribute('stroke')).toBe('#fff');
    }
  });

  it('colours each node by severity or by kind', () => {
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    const fillOf = (id: string) =>
      container.querySelector(`[data-id="${id}"] .corr-shape`)?.getAttribute('fill');

    // Findings carry the severity colour: f1 is severity 2 (critical, red),
    // f2 is severity 0 (warning, cyan). Entities and sources are fixed.
    expect(fillOf('f1')).toBe(SEV_FILL[2]);
    expect(fillOf('f2')).toBe(SEV_FILL[0]);
    expect(fillOf('e1')).toBe(KIND_FILL.entity);
    expect(fillOf('e2')).toBe(KIND_FILL.entity);
    expect(fillOf('s1')).toBe(KIND_FILL.source);
  });

  it('escapes node labels coming from the supportconfig', () => {
    const graph = makeGraph();
    graph.entities[0].label = '<script>alert(1)</script>';
    const { container } = render(<CorrelationGraphView graph={graph} />);
    expect(container.querySelector('script')).toBeNull();
    expect(container.textContent).toContain('<script>alert(1)</script>');
  });

  it('shows the detail panel of the clicked finding, with its evidence', async () => {
    const user = userEvent.setup();
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    await user.click(container.querySelector('[data-id="f1"]') as Element);

    const detail = screen.getByTestId('corr-detail');
    expect(detail).toHaveTextContent('RC4 enctype found');
    expect(detail).toHaveTextContent('crypto');
    expect(detail).toHaveTextContent('CRITICAL');
    expect(detail).toHaveTextContent('ldap_default_authtok = rc4-hmac');
  });

  // Regression: pressing a node calls svg.getScreenCTM, which jsdom does not
  // implement. The optional chain on the element passed, the call threw, and
  // the exception escaped as an unhandled error even though every test passed.
  // The component must degrade to "no drag" in a DOM without SVG geometry.
  it('survives pressing a node in a DOM without SVG geometry', async () => {
    const user = userEvent.setup();
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    const node = container.querySelector('[data-id="f1"]') as Element;

    // jsdom has no SVG geometry: the prototype genuinely lacks the method.
    // The cast is deliberate — TypeScript types it from the SVG 2 DOM, which
    // is the very mismatch that made the unguarded call throw here.
    expect(typeof (SVGElement.prototype as unknown as Record<string, unknown>).getScreenCTM).toBe('undefined');
    await user.pointer([{ keys: '[MouseLeft>]', target: node }]);
    await user.pointer([{ keys: '[/MouseLeft]', target: node }]);

    // The graph is still rendered and still interactive after the gesture.
    expect(screen.getByTestId('corr-detail')).toHaveTextContent('RC4 enctype found');
  });

  it('hides findings of a severity the user switched off', async () => {
    const user = userEvent.setup();
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    const boxes = container.querySelectorAll('.corr-filter input');
    expect(boxes).toHaveLength(3);

    // The filter order is Critical, Error, Warning+, so the first box is the
    // critical severity that finding f1 carries.
    await user.click(boxes[0]);
    const byId = (id: string) => container.querySelector(`[data-id="${id}"]`) as SVGElement;
    expect(byId('f1').style.display).toBe('none');
    expect(byId('f2').style.display).toBe('');

    await user.click(boxes[0]);
    expect(byId('f1').style.display).toBe('');
  });

  it('zooms and resets through the toolbar', async () => {
    const user = userEvent.setup();
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    const svg = container.querySelector('svg') as SVGSVGElement;
    const home = svg.getAttribute('viewBox');

    await user.click(screen.getByTitle('Zoom in (+)'));
    expect(svg.getAttribute('viewBox')).not.toBe(home);

    await user.click(screen.getByTitle('Reset view (0)'));
    expect(svg.getAttribute('viewBox')).toBe(home);
  });

  it('zooms with the keyboard and resets with 0', async () => {
    const user = userEvent.setup();
    const { container } = render(<CorrelationGraphView graph={makeGraph()} />);
    const svg = container.querySelector('svg') as SVGSVGElement;
    const canvas = container.querySelector('.corr-canvas') as HTMLElement;
    const home = svg.getAttribute('viewBox');

    canvas.focus();
    await user.keyboard('+');
    expect(svg.getAttribute('viewBox')).not.toBe(home);
    await user.keyboard('0');
    expect(svg.getAttribute('viewBox')).toBe(home);
  });
});

