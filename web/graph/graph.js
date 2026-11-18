(() => {
  'use strict';

  const STACKED_QUERY = '(max-width: 60rem)';
  const DIM_ALPHA = 0.12;
  const DIM_EDGE_ALPHA = 0.05;
  const EDGE_ALPHA = 0.55;
  const FOCUS_SCALE = 1.4;
  const FOCUS_MS = 400;
  const LAYOUT_SEED = 7;
  const STABILIZATION_ITERATIONS = 250;
  const NODE_MIN = 5;
  const NODE_MAX = 17;
  const LABEL_SIZE = 12;
  const HUB_LABEL_BOOST = 1.5;
  const STARTER_COUNT = 8;

  const TEXT = {
    loading: 'Montando o grafo…',
    loadFailed: (file) => `Não foi possível carregar o grafo (${file}).`,
    libFailed: 'Não foi possível carregar a biblioteca do grafo (vis-network).',
    noResults: 'Nenhum resultado',
    oneResult: '1 resultado',
    manyResults: (n) => `${n} resultados`,
    noDoc: 'Sem comentário de documentação.',
    howToTitle: 'Como usar',
    startTitle: 'Por onde começar',
    summaryTitle: 'Resumo',
    signature: 'Assinatura',
    doc: 'Descrição',
    location: 'Arquivo',
    code: 'Código',
    back: 'Voltar ao grafo',
    backlinks: 'Citada por',
    source: (path) => `Ver ${path} no GitHub ↗`,
    noSource: 'Esta nota não descreve um arquivo único.',
  };

  const MODES = {
    notes: {
      key: 'notes',
      toggle: 'mode-notes',
      url: 'notes.json',
      hashPrefix: 'notas/',
      placeholder: 'Nota, tipo ou arquivo (ex.: Parser)',
      labelFull: 1,
      intro: 'Cada nota descreve um tipo ou ponto de entrada do projeto. Uma ligação é um [[link]] de uma nota para outra, e o tamanho do nó mostra quantas ligações ele tem.',
      howTo: [
        'Passe o mouse sobre um nó para destacar os vizinhos; o resto escurece.',
        'Clique em um nó para ler a nota; os links dentro dela levam a outras notas.',
        'Role para dar zoom: os nomes aparecem conforme você se aproxima.',
        'Arraste os nós para reorganizar. Use a busca para achar uma nota.',
      ],
      // one e many são as palavras do resumo e do selo; nas notas são o nome da categoria.
      kinds: [
        { key: 'support', label: 'Biblioteca ll1', one: 'biblioteca ll1', many: 'biblioteca ll1', color: '--k-support' },
        { key: 'cli', label: 'Terminal', one: 'terminal', many: 'terminal', color: '--k-cli' },
        { key: 'wasm', label: 'WebAssembly', one: 'webassembly', many: 'webassembly', color: '--k-wasm' },
        { key: 'tool', label: 'Ferramenta', one: 'ferramenta', many: 'ferramenta', color: '--k-tool' },
        { key: 'web', label: 'Página web', one: 'página web', many: 'página web', color: '--k-web' },
      ],
      edges: [{ key: 'link', label: 'Links', color: '--e-link', arrow: false, width: 1, length: 85 }],
    },
    code: {
      key: 'code',
      toggle: 'mode-code',
      url: 'graph.json',
      hashPrefix: 'codigo/',
      placeholder: 'Função, tipo ou arquivo (ex.: Parse)',
      labelFull: 1.8,
      intro: 'Pacotes, arquivos, tipos, funções e métodos lidos direto do código Go, com as chamadas entre eles.',
      howTo: [
        'Clique em um nó para ver a assinatura, o comentário e o código.',
        'Passe o mouse sobre um nó para destacar os vizinhos; o resto escurece.',
        'Aproxime o zoom para ver os nomes das funções.',
        'Use a busca e os filtros para achar uma função ou esconder ligações.',
      ],
      kinds: [
        { key: 'package', label: 'Pacotes', one: 'pacote', many: 'pacotes', color: '--k-package' },
        { key: 'file', label: 'Arquivos', one: 'arquivo', many: 'arquivos', color: '--k-file' },
        { key: 'type', label: 'Tipos', one: 'tipo', many: 'tipos', color: '--k-type' },
        { key: 'func', label: 'Funções', one: 'função', many: 'funções', color: '--k-func' },
        { key: 'method', label: 'Métodos', one: 'método', many: 'métodos', color: '--k-method' },
      ],
      edges: [
        { key: 'calls', label: 'Chamadas', out: 'Chama', in: 'Chamado por', color: '--e-calls', arrow: true, width: 1.2, length: 130 },
        { key: 'contains', label: 'Contém', out: 'Contém', in: 'Está em', color: '--e-contains', arrow: false, width: 1, length: 60 },
        { key: 'method', label: 'Métodos de tipo', out: 'Métodos', in: 'Método de', color: '--e-method', arrow: false, width: 1.2, length: 60 },
        { key: 'imports', label: 'Importações', out: 'Importa', in: 'Importado por', color: '--e-imports', arrow: true, width: 2, length: 200 },
      ],
    },
  };

  const $ = (id) => document.getElementById(id);
  const root = document.documentElement;
  // getComputedStyle é caro e seria chamado várias vezes por nó a cada quadro de zoom;
  // as cores só mudam quando o tema muda, então ficam guardadas até lá.
  const palette = new Map();
  const cssVar = (name) => {
    if (!palette.has(name)) palette.set(name, getComputedStyle(root).getPropertyValue(name).trim());
    return palette.get(name);
  };
  const clamp = (x, lo, hi) => Math.min(hi, Math.max(lo, x));

  function rgba(color, alpha) {
    let hex = color.replace('#', '');
    if (hex.length === 3) hex = [...hex].map((c) => c + c).join('');
    const n = parseInt(hex, 16);
    return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, ${alpha})`;
  }

  function el(tag, props = {}, ...children) {
    const node = document.createElement(tag);
    Object.assign(node, props);
    node.append(...children);
    return node;
  }

  const state = {
    mode: null,
    data: null,
    network: null,
    nodes: null,
    edges: null,
    selectedId: null,
    hoverId: null,
    matches: null,
    scale: 1,
    fitScale: 0,
    ticket: 0,
    labelFrame: 0,
  };
  const cache = {};

  const repoUrl = $('repo-link').href.replace(/\/$/, '');
  const isStacked = () => window.matchMedia(STACKED_QUERY).matches;
  const kindOf = (node) => state.mode.kinds.find((k) => k.key === node.kind) || state.mode.kinds[0];
  const edgeSpec = (key) => state.mode.edges.find((e) => e.key === key);

  // ---- dados -------------------------------------------------------------

  function prepare(mode, raw) {
    const graph = mode.key === 'notes'
      ? {
        nodes: raw.nodes.map((n) => ({ ...n, label: n.title })),
        edges: raw.edges.map((e) => ({ ...e, kind: 'link' })),
      }
      : raw;
    const byId = new Map(graph.nodes.map((node) => [node.id, node]));
    const degree = new Map(graph.nodes.map((node) => [node.id, 0]));
    graph.edges.forEach((edge) => {
      degree.set(edge.from, (degree.get(edge.from) || 0) + 1);
      degree.set(edge.to, (degree.get(edge.to) || 0) + 1);
    });
    const names = new Map();
    if (mode.key === 'notes') graph.nodes.forEach((node) => names.set(node.id.toLowerCase(), node.id));
    return {
      graph,
      byId,
      degree,
      names,
      module: raw.module,
      maxDegree: Math.max(1, ...degree.values()),
      visibleKinds: new Set(mode.kinds.map((k) => k.key)),
      visibleEdges: new Set(mode.edges.map((e) => e.key)),
    };
  }

  async function load(mode) {
    if (cache[mode.key]) return cache[mode.key];
    const response = await fetch(mode.url);
    if (!response.ok) throw new Error(response.statusText);
    cache[mode.key] = prepare(mode, await response.json());
    return cache[mode.key];
  }

  const nodeVisible = (node) => state.data.visibleKinds.has(node.kind);
  const edgeVisible = (edge) => state.data.visibleEdges.has(edge.kind)
    && nodeVisible(state.data.byId.get(edge.from))
    && nodeVisible(state.data.byId.get(edge.to));

  // ---- estilo ------------------------------------------------------------

  const focusId = () => state.hoverId || state.selectedId;

  // O destaque vem do nó sob o mouse ou selecionado (com os vizinhos) ou da busca.
  function emphasized() {
    const id = focusId();
    if (id) {
      const set = new Set([id]);
      state.data.graph.edges.filter(edgeVisible).forEach((edge) => {
        if (edge.from === id) set.add(edge.to);
        if (edge.to === id) set.add(edge.from);
      });
      return set;
    }
    return state.matches;
  }

  // Os nomes aparecem aos poucos com o zoom; nós com muitas ligações aparecem antes das folhas.
  function zoomAlpha(node) {
    const relative = state.scale / (state.fitScale || state.scale || 1);
    const boost = 1 + HUB_LABEL_BOOST * ((state.data.degree.get(node.id) || 0) / state.data.maxDegree);
    const full = state.mode.labelFull;
    return clamp((relative * boost - full * 0.45) / (full * 0.55), 0, 1);
  }

  function fontFor(node, emph) {
    const alpha = emph ? (emph.has(node.id) ? 1 : 0) : zoomAlpha(node);
    return {
      color: rgba(cssVar('--ink'), alpha),
      strokeColor: rgba(cssVar('--bg'), alpha),
      strokeWidth: 3,
      size: LABEL_SIZE,
      face: cssVar('--sans'),
    };
  }

  function nodeStyle(node, emph) {
    const base = cssVar(kindOf(node).color);
    const lit = !emph || emph.has(node.id);
    const alpha = lit ? 1 : DIM_ALPHA;
    const isFocus = node.id === focusId();
    const fill = rgba(base, alpha);
    const border = isFocus ? cssVar('--ink') : fill;
    return {
      id: node.id,
      hidden: !nodeVisible(node),
      borderWidth: isFocus ? 2 : 1,
      color: { background: fill, border, highlight: { background: fill, border }, hover: { background: fill, border } },
      font: fontFor(node, emph),
    };
  }

  function edgeStyle(edge, index, emph) {
    const spec = edgeSpec(edge.kind);
    const id = focusId();
    let alpha = EDGE_ALPHA;
    let color = cssVar(spec.color);
    let width = spec.width;
    if (id) {
      const touches = edge.from === id || edge.to === id;
      alpha = touches ? 0.95 : DIM_EDGE_ALPHA;
      if (touches) { color = cssVar('--accent'); width += 0.6; }
    } else if (emph) {
      alpha = emph.has(edge.from) && emph.has(edge.to) ? EDGE_ALPHA : DIM_EDGE_ALPHA;
    }
    const paint = rgba(color, alpha);
    return { id: index, hidden: !edgeVisible(edge), width, color: { color: paint, highlight: paint, hover: paint, inherit: false } };
  }

  function applyStyles() {
    const emph = emphasized();
    state.nodes.update(state.data.graph.nodes.map((node) => nodeStyle(node, emph)));
    state.edges.update(state.data.graph.edges.map((edge, index) => edgeStyle(edge, index, emph)));
  }

  // O zoom só muda os nomes, então atualiza apenas as fontes.
  function refreshLabels() {
    state.labelFrame = 0;
    const emph = emphasized();
    state.nodes.update(state.data.graph.nodes.map((node) => ({ id: node.id, font: fontFor(node, emph) })));
  }

  function onFrame() {
    const scale = state.network.getScale();
    if (Math.abs(scale - state.scale) < 1e-4 || state.labelFrame) return;
    state.scale = scale;
    state.labelFrame = requestAnimationFrame(refreshLabels);
  }

  // ---- painel lateral: comum ---------------------------------------------

  function linkButton(node) {
    const kind = kindOf(node);
    const button = el('button', { type: 'button', className: 'link-btn', title: node.signature || node.label },
      el('span', { className: 'dot' }), node.label);
    button.firstChild.style.setProperty('--dot-color', cssVar(kind.color));
    button.addEventListener('click', () => select(node.id, { focus: true, scroll: true }));
    return button;
  }

  function linkList(title, nodes) {
    if (!nodes.length) return [];
    const list = el('ul', { className: 'links' });
    nodes.forEach((node) => list.append(el('li', {}, linkButton(node))));
    return [el('h3', {}, title), list];
  }

  const byLabel = (a, b) => a.label.localeCompare(b.label);

  function backButton() {
    return el('button', {
      type: 'button',
      className: 'btn ghost back',
      textContent: TEXT.back,
      onclick: () => $('graph').scrollIntoView({ behavior: 'smooth', block: 'start' }),
    });
  }

  function badge(node) {
    const kind = kindOf(node);
    const span = el('span', { className: 'badge', textContent: kind.one });
    span.style.setProperty('--badge-color', cssVar(kind.color));
    return span;
  }

  function sourceLink(path, line, text) {
    const href = `${repoUrl}/blob/main/${path}${line ? `#L${line}` : ''}`;
    return el('a', { className: 'location', href, target: '_blank', rel: 'noopener', textContent: text });
  }

  // ---- painel lateral: notas ---------------------------------------------

  function wikilink(name, label) {
    const id = state.data.names.get(name.toLowerCase());
    if (!id) return null;
    const anchor = el('a', { className: 'wikilink', href: `#${state.mode.hashPrefix}${encodeURIComponent(id)}`, textContent: label });
    anchor.addEventListener('click', (event) => {
      event.preventDefault();
      select(id, { focus: true, scroll: true });
    });
    return anchor;
  }

  function renderNote(node) {
    const backlinks = state.data.graph.edges
      .filter((e) => e.to === node.id)
      .map((e) => state.data.byId.get(e.from))
      .sort(byLabel);
    $('panel').replaceChildren(
      badge(node),
      el('h2', { textContent: node.label }),
      el('div', { className: 'note-meta' },
        ...(node.class ? [el('code', { textContent: node.class })] : []),
        ...(node.step ? [el('span', { textContent: `etapa ${node.step}` })] : [])),
      el('div', { className: 'md' }, window.MiniMarkdown.render(node.body || '', { wikilink })),
      ...linkList(TEXT.backlinks, backlinks),
      el('h3', {}, TEXT.location),
      node.path
        ? sourceLink(node.path, 0, TEXT.source(node.path))
        : el('p', { className: 'empty-doc', textContent: TEXT.noSource }),
      backButton(),
    );
  }

  // ---- painel lateral: código --------------------------------------------

  function relatedBlocks(node) {
    const blocks = [];
    state.mode.edges.forEach((spec) => {
      const outgoing = state.data.graph.edges.filter((e) => e.kind === spec.key && e.from === node.id);
      const incoming = state.data.graph.edges.filter((e) => e.kind === spec.key && e.to === node.id);
      blocks.push(...linkList(spec.out, outgoing.map((e) => state.data.byId.get(e.to)).sort(byLabel)));
      blocks.push(...linkList(spec.in, incoming.map((e) => state.data.byId.get(e.from)).sort(byLabel)));
    });
    return blocks;
  }

  function renderCode(node) {
    const location = node.file
      ? [el('h3', {}, TEXT.location), sourceLink(node.file, node.line, node.line ? `${node.file}:${node.line}` : node.file)]
      : [];
    $('panel').replaceChildren(
      badge(node),
      el('h2', { textContent: node.label }),
      ...(node.signature ? [el('h3', {}, TEXT.signature), el('div', { className: 'signature', textContent: node.signature })] : []),
      el('h3', {}, TEXT.doc),
      node.doc
        ? el('p', { textContent: node.doc })
        : el('p', { className: 'empty-doc', textContent: TEXT.noDoc }),
      ...location,
      ...(node.code ? [el('h3', {}, TEXT.code), el('pre', { className: 'code', textContent: node.code })] : []),
      ...relatedBlocks(node),
      backButton(),
    );
  }

  function renderNode(node) {
    if (state.mode.key === 'notes') renderNote(node); else renderCode(node);
    $('panel').scrollTop = 0;
  }

  function renderIntro() {
    const { mode, data } = state;
    const counts = mode.kinds.map((kind) => {
      const total = data.graph.nodes.filter((n) => n.kind === kind.key).length;
      return `${total} ${total === 1 ? kind.one : kind.many}`;
    });
    const summary = mode.key === 'notes'
      ? `${data.graph.nodes.length} notas e ${data.graph.edges.length} ligações. ${counts.join(' · ')}.`
      : `${data.module}: ${counts.join(' · ')}.`;
    const starters = mode.key === 'notes'
      ? [...data.graph.nodes].sort((a, b) => data.degree.get(b.id) - data.degree.get(a.id)).slice(0, STARTER_COUNT)
      : data.graph.nodes.filter((n) => n.kind === 'package');
    $('panel').replaceChildren(
      el('p', { className: 'hint', textContent: mode.intro }),
      el('h3', {}, TEXT.howToTitle),
      el('ul', {}, ...mode.howTo.map((line) => el('li', { textContent: line }))),
      el('h3', {}, TEXT.summaryTitle),
      el('p', { className: 'hint', textContent: summary }),
      ...linkList(TEXT.startTitle, starters),
    );
  }

  // ---- seleção e busca ---------------------------------------------------

  function select(id, { focus = false, scroll = false } = {}) {
    const node = state.data.byId.get(id);
    if (!node) return;
    if (!nodeVisible(node)) {
      state.data.visibleKinds.add(node.kind);
      syncFilterBoxes();
    }
    state.selectedId = id;
    renderNode(node);
    applyStyles();
    history.replaceState(null, '', `#${state.mode.hashPrefix}${encodeURIComponent(id)}`);
    if (focus) {
      const scale = Math.max(state.network.getScale(), FOCUS_SCALE);
      state.network.focus(id, { scale, animation: { duration: FOCUS_MS, easingFunction: 'easeInOutQuad' } });
    }
    if (scroll && isStacked()) $('panel').scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  function clearSelection() {
    state.selectedId = null;
    renderIntro();
    applyStyles();
    history.replaceState(null, '', location.pathname + location.search);
  }

  function haystack(node) {
    if (!node.hay) {
      node.hay = [node.label, node.id, node.class, node.path, node.file, node.doc, node.signature, node.body]
        .filter(Boolean).join('\n').toLowerCase();
    }
    return node.hay;
  }

  function matchingNodes(query) {
    const needle = query.trim().toLowerCase();
    if (!needle) return null;
    // Melhores primeiro: nome exato, depois começo do nome, depois nome que contém, depois o resto.
    const rank = (node) => {
      const label = node.label.toLowerCase();
      const name = label.split('.').pop();
      if (label === needle || name === needle) return 0;
      if (name.startsWith(needle)) return 1;
      return label.includes(needle) ? 2 : 3;
    };
    return state.data.graph.nodes
      .filter((node) => nodeVisible(node) && haystack(node).includes(needle))
      .sort((a, b) => rank(a) - rank(b));
  }

  function onSearch() {
    const found = matchingNodes($('search').value);
    state.matches = found ? new Set(found.map((n) => n.id)) : null;
    $('search-count').textContent = !found ? ''
      : found.length === 0 ? TEXT.noResults
        : found.length === 1 ? TEXT.oneResult : TEXT.manyResults(found.length);
    if (state.nodes) applyStyles();
  }

  function onSearchKey(event) {
    if (event.key === 'Enter') {
      const found = matchingNodes($('search').value);
      if (found && found.length) select(found[0].id, { focus: true });
    } else if (event.key === 'Escape') {
      $('search').value = '';
      onSearch();
    }
  }

  // ---- filtros -----------------------------------------------------------

  function chip(spec, checked, swatchClass, onChange) {
    const input = el('input', { type: 'checkbox', checked });
    input.addEventListener('change', () => onChange(input.checked));
    const swatch = el('span', { className: `swatch ${swatchClass}` });
    const label = el('label', { className: 'chip' }, input, swatch, el('span', { className: 'chip-label', textContent: spec.label }));
    label.style.setProperty('--chip-color', `var(${spec.color})`);
    return label;
  }

  function buildFilters() {
    const { mode, data } = state;
    $('node-filters').replaceChildren(...mode.kinds.map((kind) => chip(kind, data.visibleKinds.has(kind.key), 'circle', (on) => {
      if (on) data.visibleKinds.add(kind.key); else data.visibleKinds.delete(kind.key);
      if (state.selectedId && !nodeVisible(data.byId.get(state.selectedId))) clearSelection();
      onSearch();
    })));
    $('edge-filters').replaceChildren(...mode.edges.map((edge) => chip(edge, data.visibleEdges.has(edge.key), 'line', (on) => {
      if (on) data.visibleEdges.add(edge.key); else data.visibleEdges.delete(edge.key);
      onSearch();
    })));
    $('edge-fieldset').hidden = mode.edges.length < 2;
  }

  function syncFilterBoxes() {
    const boxes = $('node-filters').querySelectorAll('input');
    state.mode.kinds.forEach((kind, i) => { boxes[i].checked = state.data.visibleKinds.has(kind.key); });
    onSearch();
  }

  // ---- rede --------------------------------------------------------------

  function createNetwork(initialId) {
    const { mode, data } = state;
    if (state.network) state.network.destroy();
    state.nodes = new vis.DataSet(data.graph.nodes.map((node) => ({
      ...nodeStyle(node, state.matches),
      label: node.label,
      value: Math.max(1, data.degree.get(node.id) || 0),
    })));
    state.edges = new vis.DataSet(data.graph.edges.map((edge, index) => {
      const spec = edgeSpec(edge.kind);
      return {
        ...edgeStyle(edge, index, state.matches),
        from: edge.from,
        to: edge.to,
        length: spec.length,
        arrows: spec.arrow ? { to: { enabled: true, scaleFactor: 0.4 } } : undefined,
      };
    }));

    state.fitScale = 0;
    state.scale = 1;
    state.network = new vis.Network($('graph'), { nodes: state.nodes, edges: state.edges }, {
      layout: { randomSeed: LAYOUT_SEED },
      physics: {
        solver: 'forceAtlas2Based',
        forceAtlas2Based: {
          gravitationalConstant: mode.key === 'notes' ? -60 : -70,
          centralGravity: 0.02,
          springConstant: 0.06,
          damping: 0.5,
          avoidOverlap: 0.4,
        },
        stabilization: { iterations: STABILIZATION_ITERATIONS, updateInterval: 50 },
      },
      interaction: { hover: true, zoomView: true, multiselect: false, tooltipDelay: 100000 },
      nodes: {
        shape: 'dot',
        chosen: false,
        scaling: { min: NODE_MIN, max: NODE_MAX, label: { enabled: false } },
      },
      edges: { chosen: false, smooth: false, selectionWidth: 0, hoverWidth: 0 },
    });

    state.network.once('stabilizationIterationsDone', () => {
      state.network.fit();
      state.fitScale = state.network.getScale();
      state.scale = state.fitScale;
      $('status').hidden = true;
      applyStyles();
      if (initialId && data.byId.has(initialId)) select(initialId, { focus: true });
    });
    state.network.on('afterDrawing', onFrame);
    state.network.on('hoverNode', (params) => { state.hoverId = params.node; applyStyles(); });
    state.network.on('blurNode', () => { state.hoverId = null; applyStyles(); });
    state.network.on('click', (params) => {
      if (params.nodes.length) select(params.nodes[0], { scroll: true });
      else if (state.selectedId) clearSelection();
    });
  }

  async function setMode(key, initialId = null) {
    const mode = MODES[key];
    const ticket = ++state.ticket;
    $('status').hidden = false;
    $('status').textContent = TEXT.loading;
    let data;
    try {
      data = await load(mode);
    } catch (error) {
      if (ticket === state.ticket) $('status').textContent = TEXT.loadFailed(mode.url);
      return;
    }
    if (ticket !== state.ticket) return;

    state.mode = mode;
    state.data = data;
    state.selectedId = null;
    state.hoverId = null;
    state.nodes = null;
    Object.values(MODES).forEach((m) => $(m.toggle).setAttribute('aria-pressed', String(m === mode)));
    $('search').placeholder = mode.placeholder;
    buildFilters();
    renderIntro();
    onSearch();
    createNetwork(initialId);
    if (!initialId) history.replaceState(null, '', location.pathname + location.search);
  }

  function initialTarget() {
    let hash;
    try {
      hash = decodeURIComponent(location.hash.slice(1));
    } catch {
      // Um hash malformado (ex.: "%E0") abre só o grafo de notas, sem travar a página.
      hash = '';
    }
    for (const mode of Object.values(MODES)) {
      if (hash.startsWith(mode.hashPrefix)) return { key: mode.key, id: hash.slice(mode.hashPrefix.length) };
    }
    return { key: 'notes', id: null };
  }

  function start() {
    if (typeof vis === 'undefined') {
      $('status').textContent = TEXT.libFailed;
      return;
    }
    if (isStacked()) $('filters-box').open = false;
    $('search').addEventListener('input', onSearch);
    $('search').addEventListener('keydown', onSearchKey);
    $('fit').addEventListener('click', () => state.network && state.network.fit({ animation: { duration: FOCUS_MS } }));
    Object.values(MODES).forEach((mode) => $(mode.toggle).addEventListener('click', () => {
      if (state.mode !== mode) setMode(mode.key);
    }));
    window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
      palette.clear();
      if (!state.nodes) return;
      buildFilters();
      applyStyles();
      if (state.selectedId) renderNode(state.data.byId.get(state.selectedId)); else renderIntro();
    });
    const target = initialTarget();
    setMode(target.key, target.id);
  }

  start();
})();
