'use strict';

const MODE = Object.freeze({ STEP: 'step', FULL: 'full' });
const PLAY_INTERVAL_MS = 800;
const WASM_READY_TIMEOUT_MS = 5000;
const WASM_READY_POLL_MS = 25;
const END_MARKER = '$';
const EPSILON = 'ε';

const TEXT = Object.freeze({
  accepted: 'ACEITA',
  rejected: 'REJEITADA',
  play: '▶ Reproduzir',
  pause: '⏸ Pausar',
  noTokens: 'Nenhum token reconhecido.',
  noConflicts: 'Gramática LL(1): nenhuma célula da tabela tem conflito. Células vazias indicam erro sintático.',
  conflicts: 'Conflitos encontrados: ',
  loadFailed: 'Não foi possível carregar o analisador. Recarregue a página.',
  stepOf: (current, total) => `Passo ${current} de ${total}`,
});

const $ = (id) => document.getElementById(id);

const app = {
  grammar: null,
  nonTerminals: new Set(),
  result: null,
  mode: MODE.FULL,
  current: 0,
  timer: null,
};

// ---------- DOM helpers (text is always set with textContent) ----------

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

// The Go side writes productions with "->"; the screen uses the arrow symbol.
function arrowText(text) {
  return text.replaceAll('->', '→');
}

function symbolClass(symbol) {
  if (symbol === EPSILON) return 'eps';
  if (app.nonTerminals.has(symbol)) return 'nt';
  return 'term';
}

function symbolNode(symbol) {
  return el('span', symbolClass(symbol), symbol);
}

function appendSymbols(parent, symbols) {
  symbols.forEach((symbol, index) => {
    if (index > 0) parent.append(el('span', 'sym-gap', ' '));
    parent.append(symbolNode(symbol));
  });
}

function productionNode(head, body) {
  const node = el('span');
  node.append(symbolNode(head), el('span', 'arrow', ' → '));
  appendSymbols(node, body);
  return node;
}

// ---------- Static sections ----------

function renderGrammar() {
  const list = $('productions');
  for (const production of app.grammar.productions) {
    const item = el('li');
    item.append(el('span', 'prod-number', String(production.number)), productionNode(production.head, production.body));
    list.append(item);
  }
  $('grouped').textContent = app.grammar.grouped.join('\n').replaceAll('->', '→');
}

function renderSets(rows) {
  const body = $('sets-table').tBodies[0];
  for (const row of rows) {
    const tr = el('tr');
    const head = el('th');
    head.scope = 'row';
    head.append(symbolNode(row.nonTerminal));
    tr.append(head);
    for (const set of [row.first, row.follow]) {
      const cell = el('td');
      cell.append(el('span', 'arrow', '{ '));
      set.forEach((symbol) => {
        const item = el('span', 'set-item');
        item.append(symbolNode(symbol));
        cell.append(item);
      });
      cell.append(el('span', 'arrow', ' }'));
      tr.append(cell);
    }
    body.append(tr);
  }
}

function buildParseTable(table, compact) {
  const tableNode = el('table', compact ? 'ptable compact' : 'ptable');
  const headRow = el('tr');
  headRow.append(el('th', null, 'M[A, a]'));
  for (const column of table.columns) {
    const th = el('th');
    th.scope = 'col';
    th.dataset.column = column;
    th.append(symbolNode(column));
    headRow.append(th);
  }
  tableNode.createTHead().append(headRow);
  const body = tableNode.createTBody();
  for (const row of table.rows) {
    const tr = el('tr');
    const th = el('th');
    th.scope = 'row';
    th.dataset.nonTerminal = row.nonTerminal;
    th.append(symbolNode(row.nonTerminal));
    tr.append(th);
    for (const cell of row.cells) {
      const td = el('td', cell.number ? 'filled' : 'empty');
      td.dataset.nonTerminal = row.nonTerminal;
      td.dataset.column = cell.column;
      if (cell.number) {
        const production = app.grammar.productions.find((p) => p.number === cell.number);
        td.append(el('span', 'cell-number', String(cell.number)), productionNode(production.head, production.body));
      } else {
        td.textContent = '·';
      }
      tr.append(td);
    }
    body.append(tr);
  }
  return tableNode;
}

function renderTables(table) {
  $('parse-table').append(buildParseTable(table, false));
  $('parse-table-mini').append(buildParseTable(table, true));
  $('table-note').textContent = table.conflicts.length ? TEXT.conflicts + table.conflicts.join('; ') : TEXT.noConflicts;
  $('table-note').classList.toggle('conflicts', table.conflicts.length > 0);
}

function highlightCell(cellRef) {
  for (const root of [$('parse-table'), $('parse-table-mini')]) {
    for (const node of root.querySelectorAll('.hl')) node.classList.remove('hl');
    if (!cellRef) continue;
    for (const node of root.querySelectorAll('td, th')) {
      const sameRow = node.dataset.nonTerminal === cellRef.nonTerminal;
      const sameColumn = node.dataset.column === cellRef.column;
      if ((node.tagName === 'TD' && sameRow && sameColumn) ||
          (node.tagName === 'TH' && (sameRow || sameColumn))) {
        node.classList.add('hl');
        if (node.tagName === 'TD') centerInScroll(node);
      }
    }
  }
}

// Scrolls only the table's own box, never the page, so the used cell stays visible.
function centerInScroll(cell) {
  const box = cell.closest('.scroll');
  const boxRect = box.getBoundingClientRect();
  if (boxRect.width === 0) return;
  const cellRect = cell.getBoundingClientRect();
  box.scrollLeft += cellRect.left - boxRect.left - (boxRect.width - cellRect.width) / 2;
}

// ---------- Analysis ----------

function tokenIndexForStep(step) {
  return app.result.tokens.length - step.input.length;
}

function renderTokens() {
  const box = $('tokens');
  box.replaceChildren();
  const tokens = app.result.tokens;
  if (tokens.length === 0) {
    box.append(el('span', 'empty-note', TEXT.noTokens));
    return;
  }
  const step = app.mode === MODE.STEP ? app.result.steps[app.current] : null;
  const activeIndex = step ? tokenIndexForStep(step) : -1;
  tokens.forEach((token, index) => {
    const chip = el('div', 'token');
    chip.append(el('span', 'kind', token.kind), el('span', 'lexeme', token.kind === END_MARKER ? 'fim' : token.lexeme));
    if (index === activeIndex) chip.classList.add('current');
    else if (index < activeIndex) chip.classList.add('consumed');
    box.append(chip);
  });
}

function stackNode(symbols, highlightTop) {
  const node = el('span');
  symbols.forEach((symbol, index) => {
    if (index > 0) node.append(el('span', 'sym-gap', ' '));
    const symbolElement = symbolNode(symbol);
    if (highlightTop && index === symbols.length - 1) {
      const wrapper = el('span', 'top-sym');
      wrapper.append(symbolElement);
      node.append(wrapper);
    } else {
      node.append(symbolElement);
    }
  });
  return node;
}

function renderStepRows() {
  const body = $('steps-table').tBodies[0];
  body.replaceChildren();
  const steps = app.result.steps;
  const visible = app.mode === MODE.STEP ? app.current + 1 : steps.length;
  steps.slice(0, visible).forEach((step, index) => {
    const tr = el('tr');
    const isLast = index === steps.length - 1;
    if (isLast) tr.className = app.result.accepted ? 'final-ok' : 'final-bad';
    if (app.mode === MODE.STEP && index === app.current) tr.classList.add('current');
    const stackCell = el('td', 'stack-col');
    stackCell.append(stackNode(step.stack, true));
    const inputCell = el('td', 'input-col');
    appendSymbols(inputCell, step.input);
    tr.append(el('td', 'number', String(step.number)), stackCell, inputCell, el('td', null, arrowText(step.action)));
    body.append(tr);
  });
  const scroll = $('steps-scroll');
  if (app.mode === MODE.STEP) scroll.scrollTop = scroll.scrollHeight;
  else scroll.scrollTop = 0;
}

function renderStepView() {
  const steps = app.result.steps;
  const step = steps[app.current];
  const stack = $('state-stack');
  stack.replaceChildren();
  step.stack.forEach((symbol, index) => {
    const cell = el('span', index === step.stack.length - 1 ? 'cell top' : 'cell');
    cell.append(symbolNode(symbol));
    stack.append(cell);
  });
  const input = $('state-input');
  input.replaceChildren();
  step.input.forEach((symbol, index) => {
    const cell = el('span', index === 0 ? 'cell next' : 'cell');
    cell.append(symbolNode(symbol));
    input.append(cell);
  });
  const action = $('state-action');
  action.textContent = arrowText(step.action);
  const isLast = app.current === steps.length - 1;
  action.className = 'action mono' + (isLast ? (app.result.accepted ? ' accept' : ' error') : '');
  $('step-counter').textContent = TEXT.stepOf(app.current + 1, steps.length);
  $('btn-prev').disabled = app.current === 0;
  $('btn-next').disabled = isLast;
  highlightCell(step.cell);
}

function renderAnalysis() {
  if (!app.result) return;
  const hasSteps = app.result.steps.length > 0;
  const stepMode = app.mode === MODE.STEP && hasSteps;
  $('mode-step').setAttribute('aria-pressed', String(app.mode === MODE.STEP));
  $('mode-full').setAttribute('aria-pressed', String(app.mode === MODE.FULL));
  $('mode-step').disabled = !hasSteps;
  $('step-controls').hidden = !stepMode;
  $('step-view').hidden = !stepMode;
  if (stepMode) renderStepView();
  else highlightCell(null);
  renderTokens();
  renderStepRows();
}

function setMode(mode) {
  stopPlaying();
  app.mode = mode;
  if (mode === MODE.STEP) app.current = 0;
  renderAnalysis();
}

function showResult(result) {
  app.result = result;
  app.current = 0;
  const banner = $('banner');
  banner.className = 'banner ' + (result.accepted ? 'ok' : 'bad');
  $('banner-verdict').textContent = (result.accepted ? '✔ ' : '✘ ') + (result.accepted ? TEXT.accepted : TEXT.rejected);
  $('banner-message').textContent = result.message;
  $('result').hidden = false;
  renderAnalysis();
}

function analyze() {
  stopPlaying();
  const input = $('sentence').value;
  const result = JSON.parse(globalThis.ll1.parse(input));
  result.tokens = result.tokens || [];
  result.steps = result.steps || [];
  showResult(result);
  $('banner').scrollIntoView({ block: 'nearest', behavior: 'smooth' });
}

function reset() {
  stopPlaying();
  app.result = null;
  app.mode = MODE.FULL;
  app.current = 0;
  $('sentence').value = '';
  $('result').hidden = true;
  highlightCell(null);
  $('sentence').focus();
}

// ---------- Step navigation ----------

function goToStep(index) {
  const last = app.result.steps.length - 1;
  app.current = Math.max(0, Math.min(last, index));
  renderAnalysis();
}

function stopPlaying() {
  if (app.timer === null) return;
  clearInterval(app.timer);
  app.timer = null;
  $('btn-play').textContent = TEXT.play;
}

function togglePlay() {
  if (app.timer !== null) {
    stopPlaying();
    return;
  }
  const last = app.result.steps.length - 1;
  if (app.current >= last) goToStep(0);
  $('btn-play').textContent = TEXT.pause;
  app.timer = setInterval(() => {
    if (app.current >= last) {
      stopPlaying();
      return;
    }
    goToStep(app.current + 1);
    if (app.current >= last) stopPlaying();
  }, PLAY_INTERVAL_MS);
}

// ---------- Wiring ----------

function setSentence(text) {
  $('sentence').value = text;
  analyze();
}

function wireEvents() {
  $('analyze-form').addEventListener('submit', (event) => {
    event.preventDefault();
    analyze();
  });
  $('btn-valid').addEventListener('click', () => setSentence(globalThis.ll1.generate()));
  $('btn-invalid').addEventListener('click', () => setSentence(globalThis.ll1.generateInvalid()));
  $('btn-clear').addEventListener('click', reset);
  for (const button of document.querySelectorAll('[data-example]')) {
    button.addEventListener('click', () => setSentence(button.dataset.example));
  }
  $('mode-step').addEventListener('click', () => setMode(MODE.STEP));
  $('mode-full').addEventListener('click', () => setMode(MODE.FULL));
  $('btn-prev').addEventListener('click', () => { stopPlaying(); goToStep(app.current - 1); });
  $('btn-next').addEventListener('click', () => { stopPlaying(); goToStep(app.current + 1); });
  $('btn-play').addEventListener('click', togglePlay);
}

function enableControls() {
  for (const id of ['sentence', 'btn-analyze', 'btn-valid', 'btn-invalid', 'btn-clear']) $(id).disabled = false;
}

function failLoading() {
  $('loading').classList.add('failed');
  $('loading-text').textContent = TEXT.loadFailed;
}

async function waitForApi() {
  const deadline = Date.now() + WASM_READY_TIMEOUT_MS;
  while (!globalThis.ll1) {
    if (Date.now() > deadline) throw new Error('ll1 API not available');
    await new Promise((resolve) => setTimeout(resolve, WASM_READY_POLL_MS));
  }
}

async function loadWasm() {
  const go = new Go();
  const response = await fetch('main.wasm');
  if (!response.ok) throw new Error('main.wasm: HTTP ' + response.status);
  let instance;
  if (WebAssembly.instantiateStreaming && response.headers.get('Content-Type') === 'application/wasm') {
    ({ instance } = await WebAssembly.instantiateStreaming(response, go.importObject));
  } else {
    ({ instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject));
  }
  go.run(instance);
  await waitForApi();
}

async function start() {
  try {
    await loadWasm();
    const api = globalThis.ll1;
    app.grammar = JSON.parse(api.grammar());
    app.nonTerminals = new Set(app.grammar.nonTerminals);
    renderGrammar();
    renderSets(JSON.parse(api.sets()));
    renderTables(JSON.parse(api.table()));
    wireEvents();
    enableControls();
    $('loading').classList.add('done');
    $('sentence').focus();
  } catch (error) {
    console.error(error);
    failLoading();
  }
}

start();
