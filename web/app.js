'use strict';

const MODE = Object.freeze({ STEP: 'step', FULL: 'full' });
const PLAY_INTERVAL_MS = 800;
const WASM_READY_TIMEOUT_MS = 5000;
const WASM_READY_POLL_MS = 25;
const END_MARKER = '$';
const EPSILON = 'ε';
// Analisada para montar a derivação mais à esquerda de "id + num * id" da seção da gramática.
const DERIVATION_SAMPLE = 'x + 1 * y';
// A gramática escolhida fica na query, "?g=<id>": o hash é dos links de seção,
// que o trocam a cada clique.
const GRAMMAR_PARAM = 'g';
// Links antigos guardavam a gramática no hash.
const LEGACY_HASH_PREFIX = '#g=';
const MAX_SENTENCE_LENGTH = 200;

const TEXT = Object.freeze({
  accepted: 'ACEITA',
  rejected: 'REJEITADA',
  refused: 'NÃO ANALISADA',
  play: '▶ Reproduzir',
  pause: '⏸ Pausar',
  noTokens: 'Nenhum token reconhecido.',
  noConflicts: 'Gramática LL(1): nenhuma célula da tabela tem conflito. Células vazias indicam erro sintático.',
  conflicts: 'A gramática não é LL(1): as células destacadas recebem mais de uma produção, e um único token não basta para escolher. Por isso o analisador se recusa a analisar em vez de adivinhar. ',
  source: 'Fonte: ',
  adapted: 'adaptado para LL(1)',
  examplePrefix: 'ex.: ',
  terminalsNote: (terminals) => `Até ${MAX_SENTENCE_LENGTH} caracteres. Terminais desta gramática: ${terminals.join(' ')}`,
  loadFailed: 'Não foi possível carregar o analisador. Recarregue a página.',
  stepOf: (current, total) => `Passo ${current} de ${total}`,
  ruleOk: 'satisfeita',
  ruleConflict: 'conflito',
  singleProduction: (number) => `Uma única produção (${number}): não há escolha a fazer.`,
  emptySet: '∅',
});

const $ = (id) => document.getElementById(id);

const app = {
  presets: [],
  preset: null,
  // Textos da página escritos para a gramática do trabalho, devolvidos quando ela volta a ser escolhida.
  defaults: null,
  grammar: null,
  nonTerminals: new Set(),
  table: null,
  result: null,
  rows: [],
  mode: MODE.FULL,
  current: 0,
  timer: null,
};

// ---------- Ajudantes de DOM (texto sempre por textContent, nunca HTML) ----------

function el(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}

// O Go escreve as produções com "->"; a tela usa a seta.
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

// ---------- Seções fixas ----------

function setNode(symbols) {
  const node = el('span', 'set');
  if (symbols.length === 0) {
    node.textContent = TEXT.emptySet;
    return node;
  }
  node.append(el('span', 'arrow', '{ '));
  for (const symbol of symbols) {
    const item = el('span', 'set-item');
    item.append(symbolNode(symbol));
    node.append(item);
  }
  node.append(el('span', 'arrow', ' }'));
  return node;
}

// Forma agrupada ("E' → + T E' | - T E' | ε"), com espaços depois das cabeças para as setas alinharem.
function renderGrouped() {
  const box = $('grouped');
  const width = Math.max(...app.grammar.nonTerminals.map((head) => head.length));
  app.grammar.nonTerminals.forEach((head, line) => {
    if (line > 0) box.append('\n');
    box.append(symbolNode(head), ' '.repeat(width - head.length), el('span', 'arrow', ' → '));
    app.grammar.productions.filter((production) => production.head === head).forEach((production, index) => {
      if (index > 0) box.append(el('span', 'arrow', ' | '));
      appendSymbols(box, production.body);
    });
  });
}

function renderGrammar() {
  const list = $('productions');
  for (const id of ['productions', 'grouped', 'def-nonterminals', 'def-terminals', 'def-start']) $(id).replaceChildren();
  for (const production of app.grammar.productions) {
    const item = el('li');
    item.append(el('span', 'prod-number', String(production.number)), productionNode(production.head, production.body));
    list.append(item);
  }
  renderGrouped();
  $('def-nonterminals').append(setNode(app.grammar.nonTerminals));
  $('def-terminals').append(setNode(app.grammar.terminals));
  $('def-start').append(symbolNode(app.grammar.start));
  $('def-count').textContent = String(app.grammar.productions.length);
}

// ---------- Condição LL(1) por regra ----------

// Ordem de exibição dos conjuntos, a mesma do Go: terminais na ordem da gramática, depois $ e ε.
function sortSymbols(symbols) {
  const order = [...app.grammar.terminals, END_MARKER, EPSILON];
  return [...symbols].sort((a, b) => order.indexOf(a) - order.indexOf(b));
}

function isEmptyProduction(production) {
  return production.body.length === 1 && production.body[0] === EPSILON;
}

// "FIRST(α)", ou "(FIRST(α) ∪ FIRST(β))" quando há vários corpos.
function firstCall(bodies) {
  const node = el('span');
  if (bodies.length > 1) node.append('(');
  bodies.forEach((body, index) => {
    if (index > 0) node.append(' ∪ ');
    node.append('FIRST(');
    appendSymbols(node, body);
    node.append(')');
  });
  if (bodies.length > 1) node.append(')');
  return node;
}

// Desenha "<esquerda> = A ∩ B = resultado" e diz se A e B não têm símbolo em comum.
function checkLine(left, setA, setB) {
  const shared = setA.filter((symbol) => setB.includes(symbol));
  const result = setNode(shared);
  result.classList.add('result');
  const line = el('p', 'check-line');
  line.append(left, ' = ', setNode(setA), ' ∩ ', setNode(setB), ' = ', result);
  return { line, disjoint: shared.length === 0 };
}

// O FIRST de cada corpo (production.first) já vem calculado pelo Go.
function ruleCheck(head, follow) {
  const alternatives = app.grammar.productions.filter((production) => production.head === head);
  const lines = [];
  let ok = true;
  if (alternatives.length === 1) lines.push(el('p', 'check-note', TEXT.singleProduction(alternatives[0].number)));

  // Condição 1: alternativas que começam por algum símbolo têm de começar diferente,
  // senão um único token não decide qual delas usar.
  const choices = alternatives.filter((production) => !isEmptyProduction(production));
  choices.forEach((a, i) => {
    for (const b of choices.slice(i + 1)) {
      const left = el('span');
      left.append(firstCall([a.body]), ' ∩ ', firstCall([b.body]));
      const check = checkLine(left, a.first, b.first);
      lines.push(check.line);
      ok = ok && check.disjoint;
    }
  });

  // Condição 2: se uma alternativa deriva ε, as outras não podem começar pelo que vem
  // depois de A (FOLLOW(A)), senão o analisador não sabe se aplica a que deriva ε.
  for (const nullable of alternatives.filter((production) => production.first.includes(EPSILON))) {
    const others = alternatives.filter((production) => production !== nullable);
    if (others.length === 0) continue;
    const othersFirst = sortSymbols(new Set(others.flatMap((production) => production.first)));
    const left = el('span');
    left.append(firstCall(others.map((production) => production.body)), ' ∩ FOLLOW(', symbolNode(head), ')');
    const check = checkLine(left, othersFirst, follow.get(head));
    lines.push(check.line);
    ok = ok && check.disjoint;
  }

  const box = el('div', 'check ' + (ok ? 'ok' : 'bad'));
  const header = el('div', 'check-head');
  header.append(symbolNode(head), el('span', 'tag ' + (ok ? 'ok' : 'bad'), ok ? TEXT.ruleOk : TEXT.ruleConflict));
  box.append(header, ...lines);
  return box;
}

function renderChecks(rows) {
  $('ll1-checks').replaceChildren();
  const follow = new Map(rows.map((row) => [row.nonTerminal, row.follow]));
  for (const head of app.grammar.nonTerminals) $('ll1-checks').append(ruleCheck(head, follow));
}

// ---------- Derivação mais à esquerda ----------

function productionAt(cellRef) {
  const row = app.table.rows.find((candidate) => candidate.nonTerminal === cellRef.nonTerminal);
  const cell = row.cells.find((candidate) => candidate.column === cellRef.column);
  return app.grammar.productions.find((production) => production.number === cell.number);
}

function derivationLine(form, sign, number) {
  const item = el('li');
  const formNode = el('span', 'form');
  const next = form.findIndex((symbol) => app.nonTerminals.has(symbol));
  form.forEach((symbol, index) => {
    if (index > 0) formNode.append(el('span', 'sym-gap', ' '));
    const node = symbolNode(symbol);
    if (index === next) node.classList.add('next-nt');
    formNode.append(node);
  });
  item.append(el('span', 'step-sign', sign), formNode);
  if (number) item.append(el('span', 'prod-number', String(number)));
  return item;
}

// O analisador preditivo sempre expande o não terminal mais à esquerda, então as
// produções que ele aplica, em ordem, são exatamente a derivação mais à esquerda.
function renderDerivation() {
  const result = JSON.parse(globalThis.ll1.parse(DERIVATION_SAMPLE, app.preset.id));
  const applied = result.steps.filter((step) => step.cell).map((step) => productionAt(step.cell));
  const list = $('derivation');
  list.replaceChildren();
  let form = [app.grammar.start];
  list.append(derivationLine(form, '', 0));
  for (const production of applied) {
    const index = form.findIndex((symbol) => app.nonTerminals.has(symbol));
    const body = production.body.filter((symbol) => symbol !== EPSILON);
    form = [...form.slice(0, index), ...body, ...form.slice(index + 1)];
    list.append(derivationLine(form, '⇒', production.number));
  }
}

function renderSets(rows) {
  const body = $('sets-table').tBodies[0];
  body.replaceChildren();
  for (const row of rows) {
    const tr = el('tr');
    const head = el('th');
    head.scope = 'row';
    head.append(symbolNode(row.nonTerminal));
    tr.append(head);
    for (const set of [row.first, row.follow]) {
      const cell = el('td');
      cell.append(setNode(set));
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
      if (cell.conflicts) {
        // Todas as produções que pedem a célula, para o conflito ficar visível.
        td.classList.add('conflict');
        for (const number of cell.conflicts) {
          const production = app.grammar.productions.find((p) => p.number === number);
          const line = el('span', 'conflict-line');
          line.append(el('span', 'cell-number', String(number)), productionNode(production.head, production.body));
          td.append(line);
        }
      } else if (cell.number) {
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
  $('parse-table').replaceChildren(buildParseTable(table, false));
  $('parse-table-mini').replaceChildren(buildParseTable(table, true));
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

// Rola só a caixa da tabela, nunca a página, para a célula usada ficar à vista.
function centerInScroll(cell) {
  const box = cell.closest('.scroll');
  const boxRect = box.getBoundingClientRect();
  if (boxRect.width === 0) return;
  const cellRect = cell.getBoundingClientRect();
  box.scrollLeft += cellRect.left - boxRect.left - (boxRect.width - cellRect.width) / 2;
}

// ---------- Análise ----------

// A entrada do passo é o fim da lista de tokens, então a diferença dá o token atual.
// No erro léxico a entrada tem os tokens lidos mais o caractere inválido e "$", que
// não estão nos tokens, e a conta fica negativa: aí todos os lidos contam como consumidos.
function tokenIndexForStep(step) {
  const index = app.result.tokens.length - step.input.length;
  return index < 0 ? app.result.tokens.length : index;
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

// Cada linha é montada uma vez por resultado e reaproveitada: andar pelos passos só
// recoloca linhas prontas em vez de refazer cada símbolo.
function stepRow(index) {
  if (app.rows[index]) return app.rows[index];
  const steps = app.result.steps;
  const step = steps[index];
  const tr = el('tr');
  if (index === steps.length - 1) tr.className = app.result.accepted ? 'final-ok' : 'final-bad';
  const stackCell = el('td', 'stack-col');
  stackCell.append(stackNode(step.stack, true));
  const inputCell = el('td', 'input-col');
  appendSymbols(inputCell, step.input);
  tr.append(el('td', 'number', String(step.number)), stackCell, inputCell, el('td', 'action-col', arrowText(step.action)));
  app.rows[index] = tr;
  return tr;
}

function renderStepRows() {
  const steps = app.result.steps;
  const stepMode = app.mode === MODE.STEP;
  const visible = stepMode ? app.current + 1 : steps.length;
  const rows = [];
  for (let index = 0; index < visible; index++) {
    const row = stepRow(index);
    row.classList.toggle('current', stepMode && index === app.current);
    rows.push(row);
  }
  $('steps-table').tBodies[0].replaceChildren(...rows);
  const scroll = $('steps-scroll');
  scroll.scrollTop = stepMode ? scroll.scrollHeight : 0;
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
  // Traz para a vista os controles e o estado atual, não só o botão.
  if (mode === MODE.STEP) $('view-toolbar').scrollIntoView({ block: 'start', behavior: 'smooth' });
}

function showResult(result) {
  app.result = result;
  app.rows = [];
  app.current = 0;
  const banner = $('banner');
  banner.className = 'banner ' + (result.accepted ? 'ok' : 'bad');
  // Ler offsetWidth reinicia a animação de entrada a cada novo resultado.
  void banner.offsetWidth;
  banner.classList.add('shown');
  // Uma gramática com conflitos nunca roda, então as sentenças dela não são "rejeitadas".
  const verdict = result.accepted ? TEXT.accepted : (app.preset.isLL1 ? TEXT.rejected : TEXT.refused);
  $('banner-verdict').textContent = (result.accepted ? '✔ ' : '✘ ') + verdict;
  $('banner-message').textContent = result.message;
  $('result').hidden = false;
  renderAnalysis();
}

function analyze() {
  stopPlaying();
  const input = $('sentence').value;
  showResult(JSON.parse(globalThis.ll1.parse(input, app.preset.id)));
  $('banner').scrollIntoView({ block: 'nearest', behavior: 'smooth' });
}

function reset() {
  stopPlaying();
  app.result = null;
  app.rows = [];
  app.mode = MODE.FULL;
  app.current = 0;
  $('sentence').value = '';
  $('result').hidden = true;
  highlightCell(null);
  $('sentence').focus({ preventScroll: true });
}

// ---------- Navegação pelos passos ----------

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
  // No último passo, reproduzir recomeça do primeiro.
  if (app.current >= last) goToStep(0);
  $('btn-play').textContent = TEXT.pause;
  app.timer = setInterval(() => {
    goToStep(app.current + 1);
    if (app.current >= last) stopPlaying();
  }, PLAY_INTERVAL_MS);
}

// ---------- Escolha da gramática ----------

function presetFromUrl() {
  const id = new URLSearchParams(location.search).get(GRAMMAR_PARAM);
  return app.presets.find((preset) => preset.id === id) || null;
}

// Troca um link antigo "#g=<id>" por "?g=<id>", uma vez, ao abrir a página.
function moveLegacyHash() {
  if (!location.hash.startsWith(LEGACY_HASH_PREFIX)) return;
  const url = new URL(location.href);
  url.searchParams.set(GRAMMAR_PARAM, location.hash.slice(LEGACY_HASH_PREFIX.length));
  url.hash = '';
  history.replaceState(null, '', url);
}

function defaultPreset() {
  return app.presets.find((preset) => preset.isDefault);
}

function renderPresetOptions() {
  const select = $('preset');
  const groups = new Map();
  for (const preset of app.presets) {
    if (!groups.has(preset.group)) {
      const group = el('optgroup');
      group.label = preset.group;
      groups.set(preset.group, group);
      select.append(group);
    }
    const option = el('option', null, preset.name);
    option.value = preset.id;
    groups.get(preset.group).append(option);
  }
}

// "Fonte: <autor>, <título>" com link para a fonte, mais o aviso de adaptação.
function fillSource(node, preset) {
  node.replaceChildren();
  node.hidden = !preset.author;
  if (!preset.author) return;
  const link = el('a', null, `${preset.author}, ${preset.source}`);
  link.href = preset.url;
  link.target = '_blank';
  link.rel = 'noopener';
  node.append(TEXT.source, link);
  if (preset.adapted) node.append(' · ', el('span', 'adapted', TEXT.adapted));
}

function exampleButton(sentence) {
  const button = el('button', 'example mono', sentence);
  button.type = 'button';
  button.addEventListener('click', () => setSentence(sentence));
  return button;
}

function renderPresetTexts() {
  const preset = app.preset;
  const isDefault = preset.isDefault;
  for (const node of document.querySelectorAll('[data-assignment]')) node.hidden = !isDefault;
  $('preset-about').hidden = isDefault;
  $('preset-name').textContent = preset.name;
  $('preset-description').textContent = preset.description;
  fillSource($('preset-source'), preset);
  fillSource($('preset-about-source'), preset);
  $('grouped-caption').textContent = isDefault ? app.defaults.caption : preset.name;
  $('sentence').placeholder = isDefault ? app.defaults.placeholder : TEXT.examplePrefix + preset.examples[0];
  $('sentence-note').textContent = isDefault ? app.defaults.note : TEXT.terminalsNote(preset.terminals);
  const examples = $('examples');
  examples.replaceChildren(app.defaults.examplesLabel);
  for (const sentence of [...preset.examples, ...preset.invalidExamples]) examples.append(' ', exampleButton(sentence));
}

// Carrega uma gramática pronta: gramática, conjuntos, condição LL(1), tabela e
// exemplos; a análise anterior é apagada.
function selectPreset(preset) {
  const api = globalThis.ll1;
  app.preset = preset;
  $('preset').value = preset.id;
  app.grammar = JSON.parse(api.grammar(preset.id));
  app.nonTerminals = new Set(app.grammar.nonTerminals);
  app.table = JSON.parse(api.table(preset.id));
  const sets = JSON.parse(api.sets(preset.id));
  renderGrammar();
  renderSets(sets);
  renderChecks(sets);
  renderTables(app.table);
  if (preset.isDefault) renderDerivation();
  renderPresetTexts();
  reset();
}

function onPresetChange() {
  const preset = app.presets.find((candidate) => candidate.id === $('preset').value) || defaultPreset();
  // A gramática do trabalho fica com o endereço limpo, sem ?g=.
  const url = new URL(location.href);
  if (preset.isDefault) url.searchParams.delete(GRAMMAR_PARAM);
  else url.searchParams.set(GRAMMAR_PARAM, preset.id);
  history.replaceState(null, '', url);
  selectPreset(preset);
}

// ---------- Ligação dos eventos ----------

function setSentence(text) {
  $('sentence').value = text;
  analyze();
}

function wireEvents() {
  $('analyze-form').addEventListener('submit', (event) => {
    event.preventDefault();
    analyze();
  });
  $('btn-valid').addEventListener('click', () => setSentence(globalThis.ll1.generate(app.preset.id)));
  $('btn-invalid').addEventListener('click', () => setSentence(globalThis.ll1.generateInvalid(app.preset.id)));
  $('btn-clear').addEventListener('click', reset);
  $('preset').addEventListener('change', onPresetChange);
  $('mode-step').addEventListener('click', () => setMode(MODE.STEP));
  $('mode-full').addEventListener('click', () => setMode(MODE.FULL));
  $('btn-prev').addEventListener('click', () => { stopPlaying(); goToStep(app.current - 1); });
  $('btn-next').addEventListener('click', () => { stopPlaying(); goToStep(app.current + 1); });
  $('btn-play').addEventListener('click', togglePlay);
}

function enableControls() {
  for (const id of ['preset', 'sentence', 'btn-analyze', 'btn-valid', 'btn-invalid', 'btn-clear']) $(id).disabled = false;
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
    app.presets = JSON.parse(globalThis.ll1.presets());
    app.defaults = {
      caption: $('grouped-caption').textContent,
      placeholder: $('sentence').placeholder,
      note: $('sentence-note').textContent,
      examplesLabel: $('examples').textContent.trim(),
    };
    $('sentence').maxLength = MAX_SENTENCE_LENGTH;
    renderPresetOptions();
    moveLegacyHash();
    selectPreset(presetFromUrl() || defaultPreset());
    wireEvents();
    enableControls();
    $('loading').classList.add('done');
    $('sentence').focus({ preventScroll: true });
  } catch (error) {
    console.error(error);
    failLoading();
  }
}

start();
