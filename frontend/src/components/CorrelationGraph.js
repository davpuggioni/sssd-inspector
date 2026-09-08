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
  const total = graph.entities.length + graph.findings.length + graph.sources.length;
  // Golden-angle spiral: guarantees an even, non-overlapping seeding ring
  // regardless of node count (fixed 0.5 rad steps piled nodes on top of
  // each other once the count grew).
  const GOLDEN = 2.399963;
  const seedR = (kindScale) => {
    const base = Math.min(width, height) * 0.10;
    const spread = Math.min(width, height) * 0.42;
    return (i) => base + spread * Math.sqrt((i + 1) / Math.max(total, 1)) * kindScale;
  };
  const entityR = seedR(0.75), findingR = seedR(1.0), sourceR = seedR(1.2);
  const nodes = graph.entities.map((e, i) => ({
    id: e.id, kind: 'entity', label: e.label, value: e.value,
    severity: 0, r: NODE_R.entity, fill: KIND_FILL.entity,
    x: cx + Math.cos(i * GOLDEN) * entityR(i), y: cy + Math.sin(i * GOLDEN) * entityR(i), vx: 0, vy: 0,
  }));
  const off = nodes.length;
  graph.findings.forEach((f, i) => {
    nodes.push({
      id: f.id, kind: 'finding', label: f.message, category: f.category,
      severity: f.severity, source_path: f.source_path, source_key: f.source_key,
      source_line: f.source_line, evidence: f.evidence, event_count: f.event_count,
      r: NODE_R.finding, fill: SEV_FILL[f.severity] || SEV_FILL[0],
      x: cx + Math.cos((off + i) * GOLDEN) * findingR(off + i), y: cy + Math.sin((off + i) * GOLDEN) * findingR(off + i), vx: 0, vy: 0,
    });
  });
  const off2 = nodes.length;
  graph.sources.forEach((s, i) => {
    nodes.push({
      id: s.id, kind: 'source', label: s.source_path + ':' + s.source_line,
      source_path: s.source_path, source_line: s.source_line, line_text: s.line_text,
      r: NODE_R.source, fill: KIND_FILL.source,
      x: cx + Math.cos((off2 + i) * GOLDEN) * sourceR(off2 + i), y: cy + Math.sin((off2 + i) * GOLDEN) * sourceR(off2 + i), vx: 0, vy: 0,
    });
  });

  const byId = new Map(nodes.map(n => [n.id, n]));
  const edges = graph.edges
    .filter(e => byId.has(e.from) && byId.has(e.to))
    .map(e => ({ from: e.from, to: e.to, kind: e.kind, f: byId.get(e.from), t: byId.get(e.to) }));

  // Repulsion scales with node count: a fixed constant lets dense graphs
  // collapse into overlaps even after 250 ticks.
  const REPULSION = 8000 * (1 + nodes.length / 12), SPRING = 0.02, SPRING_LEN = 90, GRAVITY = 0.015, DAMP = 0.85, TICKS = 250;
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
      // Damped integration: velocity is decayed every tick, otherwise it
      // accumulates and nodes fly into the canvas corners and pile up.
      n.vx = (n.vx + dx * GRAVITY) * DAMP;
      n.vy = (n.vy + dy * GRAVITY) * DAMP;
      n.x += n.vx; n.y += n.vy;
    }
  }

  // Post-layout collision relaxation: push apart any node pairs that still
  // overlap after the force simulation (radius + label gutter).
  const GUTTER = 26;
  for (let pass = 0; pass < 80; pass++) {
    let moved = false;
    for (let i = 0; i < nodes.length; i++) {
      for (let j = i + 1; j < nodes.length; j++) {
        const a = nodes[i], b = nodes[j];
        const minD = a.r + b.r + GUTTER;
        let dx = a.x - b.x, dy = a.y - b.y;
        let d = Math.sqrt(dx * dx + dy * dy);
        if (d >= minD) continue;
        if (d < 0.01) { d = 0.01; dx = 0.01; dy = 0; }
        const push = (minD - d) / 2;
        const ux = dx / d, uy = dy / d;
        a.x += ux * push; a.y += uy * push;
        b.x -= ux * push; b.y -= uy * push;
        moved = true;
      }
    }
    // keep relaxed nodes inside the canvas
    for (const n of nodes) {
      n.x = Math.max(n.r, Math.min(width - n.r, n.x));
      n.y = Math.max(n.r, Math.min(height - n.r, n.y));
    }
    if (!moved) break;
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

  // Canvas scales gently with node count so dense graphs have room to spread.
  const nNodes = graph.entities.length + graph.findings.length + graph.sources.length;
  const W = Math.min(2000, Math.max(720, nNodes * 36));
  const H = Math.min(1400, Math.max(480, nNodes * 24));
  const laid = layoutGraph(graph, W, H);
  const byId = new Map(laid.nodes.map(n => [n.id, n]));
  const adj = new Map(laid.nodes.map(n => [n.id, []]));
  for (const e of laid.edges) { adj.get(e.from).push(e.to); adj.get(e.to).push(e.from); }

  // Short on-canvas label: the full text stays in <title> and the detail panel.
  const MAX_LABEL = 34;
  const shortLabel = s => s.length > MAX_LABEL ? s.slice(0, MAX_LABEL - 1) + '…' : s;

  let svg = `<svg class="corr-svg" viewBox="0 0 ${W} ${H}" preserveAspectRatio="xMidYMid meet" xmlns="http://www.w3.org/2000/svg">`;
  svg += `<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#888"/></marker></defs>`;
  for (const e of laid.edges) {
    svg += `<line class="corr-edge" data-from="${e.from}" data-to="${e.to}" x1="${e.x1.toFixed(1)}" y1="${e.y1.toFixed(1)}" x2="${e.x2.toFixed(1)}" y2="${e.y2.toFixed(1)}" stroke="#555" stroke-width="1.2" marker-end="url(#arrow)"/>`;
  }
  for (const n of laid.nodes) {
    const cls = `corr-node corr-${n.kind}`;
    const short = shortLabel(n.label);
    if (n.kind === 'finding') {
      const r = n.r;
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="${n.severity}" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(n.label)}</title><polygon class="corr-shape" points="0,${-r} ${r},0 0,${r} ${-r},0" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(short)}</text></g>`;
    } else if (n.kind === 'source') {
      const r = n.r;
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="0" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(n.label)}</title><rect class="corr-shape" x="${-r}" y="${-r}" width="${r * 2}" height="${r * 2}" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(short)}</text></g>`;
    } else {
      svg += `<g class="${cls}" data-id="${n.id}" data-severity="0" transform="translate(${n.x.toFixed(1)},${n.y.toFixed(1)})"><title>${escapeAttr(n.label)}</title><circle class="corr-shape" r="${n.r}" fill="${n.fill}" stroke="#fff" stroke-width="1.5"/><text class="corr-node-label" x="${n.r + 6}" y="4" fill="#ddd" font-size="11">${escapeHtml(short)}</text></g>`;
    }
  }
  svg += '</svg>';
  const filterHtml = `<div class="corr-filters"><span class="corr-filter-label">Show:</span><label class="corr-filter"><input type="checkbox" data-sev="2" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[2]}"></span>Critical</label><label class="corr-filter"><input type="checkbox" data-sev="1" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[1]}"></span>Error</label><label class="corr-filter"><input type="checkbox" data-sev="0" checked/> <span class="corr-sev-dot" style="background:${SEV_FILL[0]}"></span>Warning+</label></div>`;
  const zoomHtml = `<div class="corr-zoombar"><button type="button" class="corr-zoom-btn" data-zoom="in" title="Zoom in (+)">＋</button><button type="button" class="corr-zoom-btn" data-zoom="out" title="Zoom out (−)">－</button><button type="button" class="corr-zoom-btn" data-zoom="reset" title="Reset view (0)">⤢</button><span class="corr-zoom-hint">wheel = zoom · drag background = pan · drag node = move</span></div>`;
  container.innerHTML = `<div class="corr-graph-wrap">${filterHtml}${zoomHtml}<div class="corr-canvas">${svg}</div><div class="corr-detail" style="display:none;"></div></div>`;

  const svgEl = container.querySelector('.corr-svg');
  const detailEl = container.querySelector('.corr-detail');

  // ── Zoom & pan (viewBox-based, no external libraries) ────────────────────
  const HOME_VIEW = { x: 0, y: 0, w: W, h: H };
  let view = { ...HOME_VIEW };
  const MIN_ZOOM = 0.5, MAX_ZOOM = 8; // relative to home width
  function applyView() { svgEl.setAttribute('viewBox', `${view.x} ${view.y} ${view.w} ${view.h}`); }
  function toSvgPoint(e) {
    const pt = svgEl.createSVGPoint(); pt.x = e.clientX; pt.y = e.clientY;
    const ctm = svgEl.getScreenCTM(); if (!ctm) return null;
    return pt.matrixTransform(ctm.inverse());
  }
  function zoomAt(px, py, factor) {
    const newW = view.w * factor;
    if (newW < HOME_VIEW.w / MAX_ZOOM || newW > HOME_VIEW.w / MIN_ZOOM) return;
    view.x = px - (px - view.x) * factor;
    view.y = py - (py - view.y) * factor;
    view.w = newW; view.h = view.h * factor;
    applyView();
  }
  svgEl.addEventListener('wheel', e => {
    e.preventDefault();
    const p = toSvgPoint(e); if (!p) return;
    zoomAt(p.x, p.y, e.deltaY > 0 ? 1.15 : 1 / 1.15);
  }, { passive: false });
  container.querySelectorAll('.corr-zoom-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const mode = btn.getAttribute('data-zoom');
      if (mode === 'reset') { view = { ...HOME_VIEW }; applyView(); return; }
      const f = mode === 'in' ? 1 / 1.3 : 1.3;
      zoomAt(view.x + view.w / 2, view.y + view.h / 2, f);
    });
  });
  // Pan by dragging the background (node drag still takes precedence).
  let panning = false, panStart = null;
  svgEl.addEventListener('mousedown', e => {
    if (e.target.closest('.corr-node')) return;
    panning = true;
    panStart = { px: e.clientX, py: e.clientY, vx: view.x, vy: view.y };
    e.preventDefault();
  });
  window.addEventListener('mousemove', e => {
    if (!panning) return;
    const ctm = svgEl.getScreenCTM(); if (!ctm) return;
    const rect = svgEl.getBoundingClientRect();
    const scaleX = view.w / rect.width, scaleY = view.h / rect.height;
    view.x = panStart.vx - (e.clientX - panStart.px) * scaleX;
    view.y = panStart.vy - (e.clientY - panStart.py) * scaleY;
    applyView();
  });
  window.addEventListener('mouseup', () => { panning = false; });
  // Keyboard shortcuts: +/− zoom, 0 reset (only while the canvas has focus or is hovered).
  container.addEventListener('keydown', e => {
    if (e.key === '+' || e.key === '=') zoomAt(view.x + view.w / 2, view.y + view.h / 2, 1 / 1.25);
    else if (e.key === '-' || e.key === '_') zoomAt(view.x + view.w / 2, view.y + view.h / 2, 1.25);
    else if (e.key === '0') { view = { ...HOME_VIEW }; applyView(); }
  });
  container.querySelector('.corr-canvas').setAttribute('tabindex', '0');

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
