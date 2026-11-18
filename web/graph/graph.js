(() => {
  'use strict';

  const DATA_URL = 'graph.json';
  const STACKED_QUERY = '(max-width: 60rem)';
  const DIM_NODE_OPACITY = 0.15;
  const DIM_EDGE_OPACITY = 0.07;
  const FOCUS_SCALE = 1.1;
  const FOCUS_MS = 400;
  const LAYOUT_SEED = 7;
  const STABILIZATION_ITERATIONS = 300;

  const TEXT = {
    loading: 'Montando o grafo…',
    loadFailed: 'Não foi possível carregar o grafo (graph.json).',
    libFailed: 'Não foi possível carregar a biblioteca do grafo (vis-network). Verifique a conexão com a internet.',
    noResults: 'Nenhum resultado',
    oneResult: '1 resultado',
    manyResults: (n) => `${n} resultados`,
    noDoc: 'Sem comentário de documentação.',
    howToTitle: 'Como usar',
    howTo: [
      'Clique em um nó para ver a assinatura, o comentário e o código.',
      'Os vizinhos do nó ficam em destaque; o resto escurece.',
      'Use a busca e os filtros para achar uma função ou esconder ligações.',
      'Arraste os nós para reorganizar; role para dar zoom.',
    ],
    startTitle: 'Por onde começar',
    summaryTitle: 'Resumo',
    signature: 'Assinatura',
    doc: 'Descrição',
    location: 'Arquivo',
    code: 'Código',
    back: 'Voltar ao grafo',
  };

  // Node kinds: order is the display order in filters and in the summary.
  const KINDS = [
    { key: 'package', label: 'Pacotes', one: 'pacote', many: 'pacotes', shape: 'hexagon', swatch: 'hexagon', size: 30, color: '--k-package' },
    { key: 'file', label: 'Arquivos', one: 'arquivo', many: 'arquivos', shape: 'square', swatch: 'square', size: 16, color: '--k-file' },
    { key: 'type', label: 'Tipos', one: 'tipo', many: 'tipos', shape: 'diamond', swatch: 'diamond', size: 20, color: '--k-type' },
    { key: 'func', label: 'Funções', one: 'função', many: 'funções', shape: 'dot', swatch: 'circle', size: 13, color: '--k-func' },
    { key: 'method', label: 'Métodos', one: 'método', many: 'métodos', shape: 'dot', swatch: 'circle', size: 13, color: '--k-method' },
  ];

  // Edge kinds. out/in are the panel headings for the two directions.
  const EDGES = [
    { key: 'calls', label: 'Chamadas', out: 'Chama', in: 'Chamado por', color: '--e-calls', arrow: true, width: 1.6, length: 150 },
    { key: 'contains', label: 'Contém', out: 'Contém', in: 'Está em', color: '--e-contains', arrow: false, width: 1, length: 70 },
    { key: 'method', label: 'Métodos de tipo', out: 'Métodos', in: 'Método de', color: '--e-method', arrow: false, width: 1.6, length: 70 },
    { key: 'imports', label: 'Importações', out: 'Importa', in: 'Importado por', color: '--e-imports', arrow: true, width: 3, length: 230 },
  ];

  const kindByKey = Object.fromEntries(KINDS.map((k) => [k.key, k]));
  const edgeByKey = Object.fromEntries(EDGES.map((e) => [e.key, e]));

  const $ = (id) => document.getElementById(id);
  const root = document.documentElement;
  const cssVar = (name) => getComputedStyle(root).getPropertyValue(name).trim();

  function el(tag, props = {}, ...children) {
    const node = document.createElement(tag);
    Object.assign(node, props);
    node.append(...children);
    return node;
  }

  const state = {
    graph: null,
    network: null,
    nodes: null,
    edges: null,
    byId: new Map(),
    visibleKinds: new Set(KINDS.map((k) => k.key)),
    visibleEdges: new Set(EDGES.map((e) => e.key)),
    selectedId: null,
    matches: null,
  };

  const repoUrl = $('repo-link').href.replace(/\/$/, '');
  const isStacked = () => window.matchMedia(STACKED_QUERY).matches;

  function nodeVisible(node) {
    return state.visibleKinds.has(node.kind);
  }

  function edgeVisible(edge) {
    return state.visibleEdges.has(edge.kind)
      && nodeVisible(state.byId.get(edge.from))
      && nodeVisible(state.byId.get(edge.to));
  }

  // ---- styling -----------------------------------------------------------

  function nodeStyle(node, opacity) {
    const kind = kindByKey[node.kind];
    const color = cssVar(kind.color);
    return {
      id: node.id,
      hidden: !nodeVisible(node),
      opacity,
      color: {
        background: color,
        border: cssVar('--ink'),
        highlight: { background: color, border: cssVar('--accent') },
        hover: { background: color, border: cssVar('--accent') },
      },
      font: { color: cssVar('--ink'), size: 14, face: cssVar('--mono'), strokeWidth: 4, strokeColor: cssVar('--bg') },
    };
  }

  function edgeStyle(edge, index, opacity) {
    const spec = edgeByKey[edge.kind];
    const color = cssVar(spec.color);
    return {
      id: index,
      hidden: !state.visibleEdges.has(edge.kind),
      color: { color, highlight: color, hover: color, opacity, inherit: false },
    };
  }

  // Emphasis comes from the selected node (with its neighbors) or from the search.
  function emphasizedSet() {
    if (state.selectedId) {
      const set = new Set([state.selectedId]);
      state.graph.edges.filter(edgeVisible).forEach((edge) => {
        if (edge.from === state.selectedId) set.add(edge.to);
        if (edge.to === state.selectedId) set.add(edge.from);
      });
      return set;
    }
    return state.matches;
  }

  function applyStyles() {
    const emphasized = emphasizedSet();
    state.nodes.update(state.graph.nodes.map((node) =>
      nodeStyle(node, !emphasized || emphasized.has(node.id) ? 1 : DIM_NODE_OPACITY)));
    state.edges.update(state.graph.edges.map((edge, index) => {
      const lit = !emphasized || (state.selectedId
        ? edge.from === state.selectedId || edge.to === state.selectedId
        : emphasized.has(edge.from) && emphasized.has(edge.to));
      return edgeStyle(edge, index, lit ? 1 : DIM_EDGE_OPACITY);
    }));
  }

  // ---- side panel --------------------------------------------------------

  function linkButton(node) {
    const kind = kindByKey[node.kind];
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

  function relatedBlocks(node) {
    const blocks = [];
    EDGES.forEach((spec) => {
      const outgoing = state.graph.edges.filter((e) => e.kind === spec.key && e.from === node.id);
      const incoming = state.graph.edges.filter((e) => e.kind === spec.key && e.to === node.id);
      const byLabel = (a, b) => a.label.localeCompare(b.label);
      blocks.push(...linkList(spec.out, outgoing.map((e) => state.byId.get(e.to)).sort(byLabel)));
      blocks.push(...linkList(spec.in, incoming.map((e) => state.byId.get(e.from)).sort(byLabel)));
    });
    return blocks;
  }

  function locationNode(node) {
    if (!node.file) return [];
    const text = node.line ? `${node.file}:${node.line}` : node.file;
    const href = `${repoUrl}/blob/main/${node.file}${node.line ? `#L${node.line}` : ''}`;
    return [el('h3', {}, TEXT.location),
      el('a', { className: 'location', href, target: '_blank', rel: 'noopener', textContent: text })];
  }

  function renderNode(node) {
    const kind = kindByKey[node.kind];
    const badge = el('span', { className: 'badge', textContent: kind.one });
    badge.style.setProperty('--badge-color', cssVar(kind.color));
    const panel = $('panel');
    panel.replaceChildren(
      badge,
      el('h2', { textContent: node.label }),
      ...(node.signature ? [el('h3', {}, TEXT.signature), el('div', { className: 'signature', textContent: node.signature })] : []),
      el('h3', {}, TEXT.doc),
      node.doc
        ? el('p', { textContent: node.doc })
        : el('p', { className: 'empty-doc', textContent: TEXT.noDoc }),
      ...locationNode(node),
      ...(node.code ? [el('h3', {}, TEXT.code), el('pre', { className: 'code', textContent: node.code })] : []),
      ...relatedBlocks(node),
      el('button', { type: 'button', className: 'btn ghost back', textContent: TEXT.back,
        onclick: () => $('graph').scrollIntoView({ behavior: 'smooth', block: 'start' }) }),
    );
    panel.scrollTop = 0;
  }

  function renderIntro() {
    const counts = KINDS.map((kind) => {
      const total = state.graph.nodes.filter((n) => n.kind === kind.key).length;
      return `${total} ${total === 1 ? kind.one : kind.many}`;
    });
    const packages = state.graph.nodes.filter((n) => n.kind === 'package');
    $('panel').replaceChildren(
      el('h3', {}, TEXT.howToTitle),
      el('ul', {}, ...TEXT.howTo.map((line) => el('li', { textContent: line }))),
      el('h3', {}, TEXT.summaryTitle),
      el('p', { className: 'hint', textContent: `${state.graph.module}: ${counts.join(' · ')}.` }),
      ...linkList(TEXT.startTitle, packages),
    );
  }

  // ---- selection and search ----------------------------------------------

  function select(id, { focus = false, scroll = false } = {}) {
    const node = state.byId.get(id);
    if (!node) return;
    if (!nodeVisible(node)) {
      state.visibleKinds.add(node.kind);
      syncFilterBoxes();
    }
    state.selectedId = id;
    state.network.selectNodes([id]);
    renderNode(node);
    applyStyles();
    history.replaceState(null, '', `#${encodeURIComponent(id)}`);
    if (focus) {
      state.network.focus(id, { scale: FOCUS_SCALE, animation: { duration: FOCUS_MS, easingFunction: 'easeInOutQuad' } });
    }
    if (scroll && isStacked()) $('panel').scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  function clearSelection() {
    state.selectedId = null;
    state.network.unselectAll();
    renderIntro();
    applyStyles();
    history.replaceState(null, '', location.pathname + location.search);
  }

  function matchingNodes(query) {
    const needle = query.trim().toLowerCase();
    if (!needle) return null;
    // Best matches first: exact name, then name prefix, then name contains, then the rest.
    const rank = (node) => {
      const label = node.label.toLowerCase();
      const name = label.split('.').pop();
      if (label === needle || name === needle) return 0;
      if (name.startsWith(needle)) return 1;
      return label.includes(needle) ? 2 : 3;
    };
    return state.graph.nodes
      .filter((node) => nodeVisible(node)
        && [node.label, node.file, node.doc, node.signature].some((text) => (text || '').toLowerCase().includes(needle)))
      .sort((a, b) => rank(a) - rank(b));
  }

  function onSearch() {
    const found = matchingNodes($('search').value);
    state.matches = found ? new Set(found.map((n) => n.id)) : null;
    $('search-count').textContent = !found ? ''
      : found.length === 0 ? TEXT.noResults
        : found.length === 1 ? TEXT.oneResult : TEXT.manyResults(found.length);
    applyStyles();
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

  // ---- filters -----------------------------------------------------------

  function chip(spec, checked, swatchClass, onChange) {
    const input = el('input', { type: 'checkbox', checked });
    input.addEventListener('change', () => onChange(input.checked));
    const swatch = el('span', { className: `swatch ${swatchClass}` });
    const label = el('label', { className: 'chip' }, input, swatch, el('span', { className: 'chip-label', textContent: spec.label }));
    label.style.setProperty('--chip-color', `var(${spec.color})`);
    return label;
  }

  function buildFilters() {
    KINDS.forEach((kind) => $('node-filters').append(chip(kind, true, kind.swatch, (on) => {
      if (on) state.visibleKinds.add(kind.key); else state.visibleKinds.delete(kind.key);
      if (state.selectedId && !nodeVisible(state.byId.get(state.selectedId))) clearSelection();
      refreshVisibility();
    })));
    EDGES.forEach((edge) => $('edge-filters').append(chip(edge, true, 'line', (on) => {
      if (on) state.visibleEdges.add(edge.key); else state.visibleEdges.delete(edge.key);
      refreshVisibility();
    })));
  }

  function syncFilterBoxes() {
    const boxes = $('node-filters').querySelectorAll('input');
    KINDS.forEach((kind, i) => { boxes[i].checked = state.visibleKinds.has(kind.key); });
    refreshVisibility();
  }

  function refreshVisibility() {
    onSearch();
  }

  // ---- startup -----------------------------------------------------------

  function createNetwork() {
    state.nodes = new vis.DataSet(state.graph.nodes.map((node) => ({
      ...nodeStyle(node, 1),
      label: node.label,
      shape: kindByKey[node.kind].shape,
      size: kindByKey[node.kind].size,
      title: node.signature || node.label,
    })));
    state.edges = new vis.DataSet(state.graph.edges.map((edge, index) => {
      const spec = edgeByKey[edge.kind];
      return {
        ...edgeStyle(edge, index, 1),
        from: edge.from,
        to: edge.to,
        width: spec.width,
        length: spec.length,
        arrows: spec.arrow ? { to: { enabled: true, scaleFactor: 0.6 } } : undefined,
        smooth: false,
        selectionWidth: 0,
      };
    }));

    state.network = new vis.Network($('graph'), { nodes: state.nodes, edges: state.edges }, {
      layout: { randomSeed: LAYOUT_SEED },
      physics: {
        solver: 'forceAtlas2Based',
        forceAtlas2Based: { gravitationalConstant: -70, centralGravity: 0.03, springConstant: 0.05, avoidOverlap: 0.6 },
        stabilization: { iterations: STABILIZATION_ITERATIONS, updateInterval: 50 },
      },
      interaction: { hover: true, tooltipDelay: 250, zoomView: true, multiselect: false },
      nodes: { borderWidth: 1.5, borderWidthSelected: 3 },
    });

    state.network.once('stabilizationIterationsDone', () => {
      state.network.setOptions({ physics: false });
      state.network.fit();
      $('status').hidden = true;
      const initial = decodeURIComponent(location.hash.slice(1));
      if (initial && state.byId.has(initial)) select(initial, { focus: true });
    });
    state.network.on('click', (params) => {
      if (params.nodes.length) select(params.nodes[0]);
      else if (state.selectedId) clearSelection();
    });
  }

  async function start() {
    if (typeof vis === 'undefined') {
      $('status').textContent = TEXT.libFailed;
      return;
    }
    try {
      const response = await fetch(DATA_URL);
      if (!response.ok) throw new Error(response.statusText);
      state.graph = await response.json();
    } catch (error) {
      $('status').textContent = TEXT.loadFailed;
      return;
    }
    state.graph.nodes.forEach((node) => state.byId.set(node.id, node));
    buildFilters();
    if (isStacked()) $('filters-box').open = false;
    renderIntro();
    createNetwork();
    $('search').addEventListener('input', onSearch);
    $('search').addEventListener('keydown', onSearchKey);
    $('fit').addEventListener('click', () => state.network.fit({ animation: { duration: FOCUS_MS } }));
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      applyStyles();
      if (state.selectedId) renderNode(state.byId.get(state.selectedId)); else renderIntro();
    });
  }

  $('status').textContent = TEXT.loading;
  start();
})();
