#!/usr/bin/env node
// gui-smoke.mjs — render the shipped bundle in a real browser engine.
//
// Why this exists: the Wails window needs a desktop session and WebKitGTK, so on
// a headless build agent the GUI cannot be launched at all, and nothing shows
// what a user would see. This loads the REAL production bundle from dist/ in
// headless Chromium with a stubbed Wails bridge and drives it: analyze a report,
// switch to the Definitions Studio, screenshot both. It exercises the built
// React code, the generated Wails client, the CSS and the layout — everything
// except the Go <-> WebKit boundary, which only a human on a desktop can reach
// and which the Go tests cover from the other side.
//
// The report is NOT a hand-written fixture: pass the JSON of a real run
// (`sssd-inspector -anonymize -json <supportconfig> <outdir>`), so the smoke
// also proves the report the engine actually produces renders in the UI.
//
// Usage:
//   node scripts/gui-smoke.mjs --report sc_report.json [--out /tmp/gui-smoke]
import { createServer } from 'node:http';
import { readFile, writeFile, mkdir, rm, cp } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { extname, join, resolve } from 'node:path';
import { execFile } from 'node:child_process';
import { promisify } from 'node:util';

const run = promisify(execFile);
const args = process.argv.slice(2);
const argOf = (name, fallback) => {
  const i = args.indexOf(`--${name}`);
  return i >= 0 && args[i + 1] ? args[i + 1] : fallback;
};

const reportPath = argOf('report');
const outDir = resolve(argOf('out', '/tmp/sssd-gui-smoke'));
const browser = argOf('browser', process.env.CHROME_BIN || 'chromium');
const distDir = resolve('dist');

if (!reportPath) {
  console.error('usage: node scripts/gui-smoke.mjs --report <report.json>');
  process.exit(2);
}
if (!existsSync(join(distDir, 'index.html'))) {
  console.error('frontend/dist is missing — run "npm run build" first.');
  process.exit(2);
}
const report = JSON.parse(await readFile(reportPath, 'utf8'));

const MIME = {
  '.html': 'text/html', '.js': 'text/javascript', '.css': 'text/css',
  '.woff2': 'font/woff2', '.png': 'image/png', '.svg': 'image/svg+xml',
  '.json': 'application/json',
};

function serve(root, port) {
  const server = createServer(async (req, res) => {
    const path = decodeURIComponent((req.url || '/').split('?')[0]);
    const file = join(root, path === '/' ? 'index.html' : path);
    if (!file.startsWith(root) || !existsSync(file)) {
      res.writeHead(404).end('not found');
      return;
    }
    res.writeHead(200, { 'content-type': MIME[extname(file)] || 'application/octet-stream' });
    res.end(await readFile(file));
  });
  return new Promise((ok) => server.listen(port, '127.0.0.1', () => ok(server)));
}

// --- the stubbed Wails bridge --------------------------------------------
// Mirrors the generated client (frontend/wailsjs): every App method returns a
// promise, and runtime events are plain callbacks. ?scene= decides what the
// smoke drives, so each view is dumped in the state a user would see.
function shim() {
  const report = window.__SMOKE_REPORT__;
  const inventory = {
    user_root: '/home/tester/.sssd-inspector',
    system_root: '/etc/sssd-inspector',
    files: [
      { path: '/home/tester/.sssd-inspector/rules.yaml', kind: 'rules', scope: 'user', exists: true, is_dir: false, writable: true, rule_count: 1, article_count: 0, size_bytes: 512, mod_time: '2026-09-29 10:00:00' },
      { path: '/etc/sssd-inspector/rules.yaml', kind: 'rules', scope: 'system', exists: false, is_dir: false, writable: false, rule_count: 0, article_count: 0 },
      { path: '/home/tester/.sssd-inspector/catalog.json', kind: 'catalog', scope: 'user', exists: true, is_dir: false, writable: true, rule_count: 0, article_count: 0, option_count: 519, size_bytes: 250086, mod_time: '2026-09-29 10:00:00' },
      { path: '/etc/sssd-inspector/catalog.json', kind: 'catalog', scope: 'system', exists: false, is_dir: false, writable: false, rule_count: 0, article_count: 0 },
    ],
    rules: [{
      rule: { name: 'legacy-rc4', severity: 'warning', category: 'crypto', files: ['sssd.conf'], patterns: ['rc4-hmac'], match: 'any', pattern_type: 'literal', message: 'RC4 enctype found' },
      file: '/home/tester/.sssd-inspector/rules.yaml', scope: 'user', line: 3,
    }],
    rule_count: 1,
    article_count: 16,
  };
  const rulesDoc = 'rules:\n  - name: "legacy-rc4"\n    severity: warning\n    category: crypto\n    files: ["sssd.conf"]\n    message: "RC4 enctype found"\n    patterns: ["rc4-hmac"]\n';
  const valid = { label: 'valid', valid: true, rule_count: 1, rules: [], diagnostics: [] };
  const rule = { name: 'legacy-rc4', severity: 'warning', category: 'crypto', files: ['sssd.conf'], patterns: ['rc4-hmac'], match: 'any', pattern_type: 'literal', message: 'RC4 enctype found' };
  const ok = (value) => () => Promise.resolve(value);
  const listeners = {};

  window.go = { main: { App: {
    Analyze: ok(report),
    OpenFileBrowser: ok('/tmp/supportconfig.txz'),
    SaveTXT: ok('report.txt'),
    SaveJSON: ok('report.json'),
    SavePDF: ok('report.pdf'),
    ListDefinitions: ok(inventory),
    ValidateRuleYAML: ok(valid),
    ReadRuleYAML: ok({ path: '/home/tester/.sssd-inspector/rules.yaml', scope: 'user', exists: true, bytes: 140, content: rulesDoc }),
    SaveRuleYAML: ok({ path: '/home/tester/.sssd-inspector/rules.yaml', scope: 'user', saved: true, bytes: 140, validation: valid }),
    TestRulesAgainst: ok({
      target_path: '/tmp/supportconfig.txz', total: 1, matched: 1,
      outcomes: [{ rule, file: 'sssd.conf', scope: 'user', line: 3, matched: true, evidence: 'ldap_default_authtok = rc4-hmac' }],
      diagnostics: [],
    }),
    GetCatalogInfo: ok({
      source: 'SSSD upstream', version: '2.14.0', generated: '2026-09-29', option_count: 519,
      section_count: 12, available: true, effective: '/home/tester/.sssd-inspector/catalog.json',
      using_override: true, override_paths: ['/home/tester/.sssd-inspector/catalog.json', '/etc/sssd-inspector/catalog.json'],
      diagnostics: [],
    }),
    OpenDefinitionsRoot: ok(undefined),
  } } };
  window.runtime = {
    EventsOn: (name, cb) => { (listeners[name] ||= []).push(cb); return () => {}; },
    EventsOnMultiple: (name, cb) => { (listeners[name] ||= []).push(cb); return () => {}; },
    EventsOff: () => {}, EventsOnce: () => () => {},
    LogPrint: () => {}, LogInfo: () => {}, LogTrace: () => {}, LogError: () => {},
    OnFileDrop: (cb) => { window.runtime.__drop = cb; },
    OnFileDropOff: () => {},
  };

  const scene = new URLSearchParams(location.search).get('scene');
  // A fresh installation has no custom rules, so Go sends "rules": null: a nil
  // slice marshals as null, not []. That is the state which crashed the Studio
  // before listOf() existed, so the smoke drives it instead of only the happy
  // path. The render-crash detector below turns it into a hard failure.
  if (scene === 'studio-empty') {
    inventory.files = null;
    inventory.rules = null;
    inventory.diagnostics = null;
    inventory.rule_count = 0;
  }
  const byText = (selector, text) =>
    [...document.querySelectorAll(selector)].find((el) => el.textContent.trim().includes(text));
  const type = (input, value) => {
    const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
    setter.call(input, value);
    input.dispatchEvent(new Event('input', { bubbles: true }));
  };

  if (scene === 'analysis') {
    setTimeout(() => {
      type(document.querySelector('#filePath'), '/tmp/supportconfig.txz');
      document.querySelector('#anonymizeCheck').click();
      byText('button', 'Analyze').click();
      (listeners['analyze-progress'] || []).forEach((cb) => cb('Scanning SSSD logs', 40));
    }, 150);
  } else if (scene === 'studio') {
    setTimeout(() => byText('button', 'Definitions Studio').click(), 150);
    setTimeout(() => byText('button', 'Validate').click(), 400);
    setTimeout(() => byText('button', 'Run dry-run').click(), 600);
  } else if (scene === 'studio-empty') {
    setTimeout(() => byText('button', 'Definitions Studio').click(), 150);
  } else if (scene === 'graph') {
    // The analysis scene stops at the top of the report, so the graph is below
    // the fold and its screenshot never showed it. This scene runs the same
    // analysis, opens the graph section, scrolls the canvas into view and
    // clicks a node, so the screenshot contains the drawn shapes rather than
    // an empty area of the page. The node colour is asserted from the DOM
    // below: a black shape on the dark canvas passes every other marker here.
    setTimeout(() => {
      type(document.querySelector('#filePath'), '/tmp/supportconfig.txz');
      byText('button', 'Analyze').click();
    }, 150);
    setTimeout(() => {
      const section = document.querySelector('.corr-graph-section');
      if (section) {
        const details = section.closest('details');
        if (details) details.open = true;
        section.scrollIntoView({ block: 'center' });
      }
      const node = document.querySelector('.corr-node');
      if (node) node.dispatchEvent(new MouseEvent('click', { bubbles: true }));
    }, 900);
  }
}


// --- build the served copy ------------------------------------------------
await rm(outDir, { recursive: true, force: true });
await mkdir(join(outDir, 'site'), { recursive: true });
await cp(distDir, join(outDir, 'site'), { recursive: true });

const indexPath = join(outDir, 'site', 'index.html');
const html = await readFile(indexPath, 'utf8');
await writeFile(indexPath, html.replace('</head>',
  `  <script>window.__SMOKE_REPORT__ = ${JSON.stringify(report).replace(/</g, '\\u003c')};</script>\n` +
  `  <script>(${shim.toString()})();</script>\n</head>`));

const port = 8173 + Math.floor(Math.random() * 400);
const server = await serve(join(outDir, 'site'), port);
const base = `http://127.0.0.1:${port}/index.html`;

// --- scenes and the markers a user would look for -------------------------
const scenes = [
  { name: 'analysis', markers: [
    ['report heading', 'Analysis Report'],
    ['health score', 'Health Score'],
    ['configuration findings', 'Configuration Findings'],
    ['actionable problems', 'Critical Problems'],
    ['sssd.conf snippet', 'View sssd.conf'],
    ['catalog provenance row', 'Option Catalog'],
    ['correlation graph', 'Correlation Graph'],
  ] },
  { name: 'graph', markers: [
    ['graph section', 'Correlation Graph'],
    ['graph canvas', 'corr-svg'],
    ['drawn node shape', 'corr-shape'],
    ['node detail panel', 'corr-detail'],
  ] },
  { name: 'studio-empty', markers: [
    ['inventory panel', 'Discovery inventory'],
    ['rule count badge', 'rules loaded'],
    ['rule editor', 'Rule editor'],
    ['catalog panel', 'SSSD option catalog'],
  ] },
  { name: 'studio', markers: [
    ['inventory panel', 'Discovery inventory'],
    ['rule editor', 'Rule editor'],
    ['dry-run panel', 'Dry-run against a supportconfig'],
    ['catalog panel', 'SSSD option catalog'],
    ['catalog override in effect', 'override'],
    ['validation verdict', 'Valid'],
    ['dry-run fired rule', 'legacy-rc4'],
  ] },
];

let failures = 0;
for (const scene of scenes) {
  const shot = join(outDir, `${scene.name}.png`);
  const { stdout } = await run(browser, [
    '--headless', '--disable-gpu', '--no-sandbox', '--hide-scrollbars',
    '--window-size=1400,1600',
    '--virtual-time-budget=4000',
    `--screenshot=${shot}`,
    '--dump-dom',
    `${base}?scene=${scene.name}`,
  ], { maxBuffer: 64 * 1024 * 1024 });

  const missing = scene.markers.filter(([, marker]) => !stdout.includes(marker)).map(([, m]) => m);
  console.log(`[${missing.length === 0 ? 'PASS' : 'FAIL'}] scene "${scene.name}" — ${scene.markers.length - missing.length}/${scene.markers.length} markers, screenshot: ${shot}`);
  for (const marker of missing) {
    console.log(`         missing marker: ${marker}`);
    failures += 1;
  }
  // A page that threw while rendering still dumps a DOM: catch that too.
  for (const [name, marker] of [['render crash', 'Something went wrong'], ['missing bridge', 'backend is not attached']]) {
    if (stdout.includes(marker)) {
      console.log(`         [FAIL] ${name} in the rendered page`);
      failures += 1;
    }
  }

  // The graph drawn in the page has to be visible, not merely present. A shape
  // with no fill attribute renders black, and black on the #1e1e24 canvas is
  // invisible: the DOM still contains the node and every text marker above
  // passes, which is why this needed a check of its own.
  if (scene.name === 'graph') {
    const shapes = [...stdout.matchAll(/class="corr-shape"[^>]*?fill="([^"]*)"/g)].map((m) => m[1].toLowerCase());
    if (shapes.length === 0) {
      console.log('         [FAIL] no corr-shape carries a fill attribute — nodes would render black on the dark canvas');
      failures += 1;
    }
    const black = shapes.filter((f) => f === '#000' || f === 'black' || f === '');
    if (black.length > 0) {
      console.log(`         [FAIL] ${black.length} node shape(s) filled black: ${black.join(', ')}`);
      failures += 1;
    } else {
      console.log(`         [ok] ${shapes.length} node shape(s) filled, none black: ${[...new Set(shapes)].join(', ')}`);
    }
  }
}

server.close();
console.log(failures === 0
  ? `\ngui-smoke OK — artifacts in ${outDir}`
  : `\ngui-smoke FAILED with ${failures} problem(s) — artifacts in ${outDir}`);
process.exit(failures === 0 ? 0 : 1);

