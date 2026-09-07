// CorrelationGraph — interactive force-directed SVG renderer (P5/P9).
// Zero external dependencies: a tiny homemade spring-electrical layout
// (Coulomb repulsion + Hooke links + central gravity) runs for a fixed
// number of ticks, then the SVG is frozen and made interactive
// (hover, click, severity filter, drag).
//
// Nodes are colour-coded by kind:
//   entity  — blue circles   (AD domain, realm, hostname, DNS, ad_server, keytab, gpo)
//   finding — severity-coloured diamonds (critical/error/warning)
//   source  — grey squares   (physical evidence: file:line)
//
// The graph is built by the Go backend (analysis_graph.go) and shipped
// inside report.graph as plain JSON.

const NODE_R = { entity: 22, finding: 18, source: 12 };
const SEV_FILL = { 0: '#17a2b8', 1: '#ffc107', 2: '#dc3545' };
const KIND_FILL = { entity: '#4a90d9', source: '#6c757d' };

function severityLabel(sev) {
  return ['warning', 'error', 'critical'][sev] || 'info';
}

function escapeHtml(s) {
  return String(s ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}
function escapeAttr(s) { return escapeHtml(s); }

/**
 * Run the force simulation and produce a layout (x/y per node).
 * Pure function — no DOM. Returns { nodes, edges } with coordinates.
 */
function layoutGraph(graph, width, height) {
  const cx = width / 2, cy = height / 2;
  const nodes = graph.entities.map((e, i) => ({
    id: e.id, kind: 'entity', label: e.label, value: e.value,
    severity: 0, r: NODE_R.entity, fill: KIND_FILL.entity,
    x: cx + Math.cos(i * 0.5) * 100, y: cy + Math.sin(i * 0.5) * 100, vx: 0, vy: 0,
  }));
  const off = nodes.length;
  graph.findings.forEach((f, i) => {
    nodes.push({
      id: f.id, kind: 'finding', label: f.message, category: f.category,
      severity: f.severity, source_path: f.source_path, source_key: f.source_key,
      source_line: f.source_line, evidence: f.evidence, event_count: f.event_count,
      r: NODE_R.finding, fill: SEV_FILL[f.severity] || SEV_FILL[0],
      x: cx + Math.cos((off + i) * 0.5) * 150, y: cy + Math.sin((off + i) * 0.5) * 150, vx: 0, vy: 0,
    });
  });
  const off2 = nodes.length;
  graph.sources.forEach((s, i) => {
    nodes.push({
      id: s.id, kind: 'source', label: s.source_path + ':' + s.source_line,
      source_path: s.source_path, source_line: s.source_line, line_text: s.line_text,
      r: NODE_R.source, fill: KIND_FILL.source,
      x: cx + Math.cos((off2 + i) * 0.5) * 200, y: cy + Math.sin((off2 + i) * 0.5) * 200, vx: 0, vy: 0,
    });
  });

  const byId = new Map(nodes.map(n => [n.id, n]));
  const edges = graph.edges
    .filter(e => byId.has(e.from) && byId.has(e.to))
    .map(e => ({ from: e.from, to: e.to, kind: e.kind, f: byId.get(e.from), t: byId.get(e.to) }));

  const REPULSION = 8000, SPRING = 0.02, SPRING_LEN = 90, GRAVITY = 0.015, DAMP = 0.85, TICKS = 250;
  for (let t = 0; t < TICKS; t++) {
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const a = nodes[i], b = nodes[j];
        let dx = a.x - b.x, dy = a.y - b.y;
        let d2 = dx * dx + dy * dy;
        if (d2 < 1) { d2 = 1; dx = 1; dy = 0; }
        const f = REPULSION / d2, d = Math.sqrt(d2);
        const fx = (dx / d) * f, fy = (dy / d) * f;
        a.vx += fx; a.vy += fy; b.vx -= fx; b.vy -= fy;
      }
    }
    for (const e of edges) {
      const a = e.f, b = e.t;
      const dx = b.x - a.x, dy = b.y - a.y;
      const d = Math.sqrt(dx * dx + dy * dy) || 1;
      const f = (d - SPRING_LEN) * SPRING;
      const fx = (dx / d) * f, fy = (dy / d) * f;
      a.vx += fx; a.vy += fy; b.vx -= fx; b.vy -= fy;
    }
    for (const n of nodes) {
      const dx = cx - n.x, dy = cy - n.y;
      n.vx += dx * GRAVITY; n.vy += dy * GRAVITY;
      n.x += n.vx; n.y += n.vy;
    }
  }

  const edgeLines = edges.map(e => ({ from: e.from, to: e.to, x1: e.f.x, y1: e.f.y, x2: e.t.x, y2: e.t.y }));
  return { nodes, edges: edgeLines };
}

// Render the correlation graph into `container`.
// `report` is the full ReportData object; the graph lives in report.graph.
// Returns a controller with destroy() and filterBySeverity(minSev), or null.
export function renderCorrelationGraph(container, report) {
  const graph = report.graph;
  if (!graph || (!graph.entities.length && !graph.findings.length && !graph.sources.length)) {
    container.innerHTML = '';
    return null;
  }

  const W = 720, H = 480;
  const laid = layoutGraph(graph, W, H);
  const byId = new Map(laid.nodes.map(n => [n.id, n]));
  const adj = new Map(laid.nodes.map(n => [n.id, []]));
  for (const e of laid.edges) { adj.get(e.from).push(e.to); adj.get(e.to).push(e.from); }

  let svg = `<svg class="corr-svg" viewBox="0 0 ${W} ${H}" preserveAspectRatio="xMidYMid meet" xmlns="http://www.w3.org/2000/svg">`;
  svg += `<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#888"/></marker></defs>`;
  for (const e of laid.edges) {
    svg += `<line class="corr-edge" data-from="${e.from}" data-to="${e.to}" x1="${e.x1.toFixed(1)}" y1="${e.y1.toFixed(1)}" x2="${e.x2.toFixed(1)}" y2="${e.y2.toFixed(1)}" stroke="#555" stroke-width="1.2" marker-end="url(#arrow)"/>`;
  }
  for (const n of laid.nodes) {
    const cls = `corr-node corr-${n.kind}`;
    const title = n.label.length > 80 ? n.label.slice(0, 77) + '…' : n.label;
    if (n.kind === 'finding') {
      const r = n.r;
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="${n.severity}" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(title)}</title><polygon class="corr-shape" points="0,${-r} ${r},0 0,${r} ${-r},0" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(title)}</text></g>`;
    } else if (n.kind === 'source') {
      const r = n.r;
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="0" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(title)}</title><rect class="corr-shape" x="${-r}" y="${-r}" width="${r * 2}" height="${r * 2}" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(title)}</text></g>`;
    } else {
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="0" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(title)}</title><circle class="corr-shape" r="${n.r}" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${n.r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(title)}</text></g>`;
    }
  }
  svg += '</svg>';
  const filterHtml = `<div class="corr-filters"><span class="corr-filter-label">Show:</span><label class="corr-filter"><input type="checkbox" data-sev="2" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[2]}"></span>Critical</label><label class="corr-filter"><input type="checkbox" data-sev="1" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[1]}"></span>Error</label><label class="corr-filter"><input type="checkbox" data-sev="0" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[0]}"></span>Warning+</label></div>`;
  container.innerHTML = `<div class="corr-graph-wrap">${filterHtml}<div class="corr-canvas">${svg}</div><div class="corr-detail" style="display:none;"></div></div>`;

  const svgEl = container.querySelector('.corr-svg');
  const detailEl = container.querySelector('.corr-detail');

  function highlight(id) {
    const conn = id ? adj.get(id) : null;
    svgEl.querySelectorAll('.corr-node').forEach(g => {
      const nid = g.getAttribute('data-id');
      g.style.opacity = (!id || nid === id || (conn && conn.has(nid))) ? '1' : '0.15';
    });
    svgEl.querySelectorAll('.corr-edge').forEach(l => {
      const f = l.getAttribute('data-from'), t = l.getAttribute('data-to');
      if (!id || f === id || t === id) { l.style.opacity = '1'; l.style.stroke = '#aaa'; }
      else { l.style.opacity = '0.08'; l.style.stroke = '#555'; }
    });
  }
  function clearHighlight() {
    svgEl.querySelectorAll('.corr-node').forEach(g => g.style.opacity = '1');
    svgEl.querySelectorAll('.corr-edge').forEach(l => { l.style.opacity = '1'; l.style.stroke = '#555'; });
  }
  function showDetail(node) {
    let html = `<h4>${escapeHtml(node.label)}</h4><div class="corr-detail-kind">${node.kind}` + (node.category ? ' · ' + escapeHtml(node.category) : '') + '</div>';
    if (node.kind === 'finding') {
      html += `<div class="corr-detail-sev sev-${severityLabel(node.severity)}">${severityLabel(node.severity).toUpperCase()}</div>`;
      if (node.source_path) html += `<div><strong>Source:</strong> ${escapeHtml(node.source_path)}` + (node.source_line ? ':' + node.source_line : '') + '</div>';
      if (node.source_key) html += `<div><strong>Key:</strong> ${escapeHtml(node.source_key)}</div>`;
      if (node.evidence) html += `<div class="corr-detail-evidence">${escapeHtml(node.evidence)}</div>`;
      if (node.event_count) html += `<div><strong>Events:</strong> ${node.event_count}</div>`;
    } else if (node.kind === 'source') {
      html += `<div><strong>File:</strong> ${escapeHtml(node.source_path)}:${node.source_line}</div>`;
      if (node.line_text) html += `<div class="corr-detail-evidence">${escapeHtml(node.line_text)}</div>`;
    } else {
      if (node.value) html += `<div><strong>Value:</strong> ${escapeHtml(node.value)}</div>`;
      if (node.source_path) html += `<div><strong>From:</strong> ${escapeHtml(node.source_path)}</div>`;
    }
    detailEl.innerHTML = html;
    detailEl.style.display = 'block';
  }

  svgEl.addEventListener('mouseover', e => { const g = e.target.closest('.corr-node'); if (g) highlight(g.getAttribute('data-id')); });
  svgEl.addEventListener('mouseout', e => { if (!e.target.closest('.corr-node')) clearHighlight(); });
  svgEl.addEventListener('click', e => {
    const g = e.target.closest('.corr-node');
    if (!g) { detailEl.style.display = 'none'; clearHighlight(); return; }
    showDetail(byId.get(g.getAttribute('data-id')));
    highlight(g.getAttribute('data-id'));
  });

  container.querySelectorAll('.corr-filter input').forEach(cb => {
    cb.addEventListener('change', () => {
      const active = new Set();
      container.querySelectorAll('.corr-filter input').forEach(c => { if (c.checked) active.add(parseInt(c.getAttribute('data-sev'))); });
      svgEl.querySelectorAll('.corr-node').forEach(g => {
        g.style.display = active.has(parseInt(g.getAttribute('data-severity'))) ? '' : 'none';
      });
    });
  });

  let dragId = null, dragOffset = { x: 0, y: 0 };
  svgEl.addEventListener('mousedown', e => {
    const g = e.target.closest('.corr-node');
    if (!g) return;
    dragId = g.getAttribute('data-id');
    const node = byId.get(dragId);
    const pt = svgEl.createSVGPoint(); pt.x = e.clientX; pt.y = e.clientY;
    const ctm = svgEl.getScreenCTM(); if (!ctm) return;
    const loc = pt.matrixTransform(ctm.inverse());
    dragOffset.x = loc.x - node.x; dragOffset.y = loc.y - node.y;
    e.preventDefault();
  });
  window.addEventListener('mousemove', e => {
    if (!dragId) return;
    const node = byId.get(dragId);
    const pt = svgEl.createSVGPoint(); pt.x = e.clientX; pt.y = e.clientY;
    const ctm = svgEl.getScreenCTM(); if (!ctm) return;
    const loc = pt.matrixTransform(ctm.inverse());
    node.x = loc.x - dragOffset.x; node.y = loc.y - dragOffset.y;
    const g = svgEl.querySelector(`.corr-node[data-id="${dragId}"]`);
    if (g) g.setAttribute('transform', `translate(${node.x.toFixed(1)},${node.y.toFixed(1)})`);
    svgEl.querySelectorAll('.corr-edge').forEach(l => {
      const f = l.getAttribute('data-from'), t = l.getAttribute('data-to');
      if (f === dragId) { l.setAttribute('x1', node.x.toFixed(1)); l.setAttribute('y1', node.y.toFixed(1)); }
      if (t === dragId) { l.setAttribute('x2', node.x.toFixed(1)); l.setAttribute('y2', node.y.toFixed(1)); }
    });
  });
  window.addEventListener('mouseup', () => { dragId = null; });

  return {
    destroy() { container.innerHTML = ''; },
    filterBySeverity(minSev) {
      svgEl.querySelectorAll('.corr-node').forEach(g => {
        g.style.display = parseInt(g.getAttribute('data-severity')) >= minSev ? '' : 'none';
      });
    },
  };
}

export default { renderCorrelationGraph, layoutGraph };
