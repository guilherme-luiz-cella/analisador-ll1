// Teste do web/graph/markdown.js no Node, sem navegador: rode com
//   node --test
// Fica fora de web/ porque tudo em web/ é copiado para o site.
'use strict';

const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

// DOM mínimo: só o que markdown.js usa (createElement, fragmento, append, textContent).
function fakeNode(tag) {
  return {
    tag,
    children: [],
    textContent: '',
    append(...items) { this.children.push(...items); },
  };
}

function load() {
  const context = vm.createContext({
    window: {},
    document: { createElement: fakeNode, createDocumentFragment: () => fakeNode('#fragment') },
  });
  const source = fs.readFileSync(path.join(__dirname, '..', 'web', 'graph', 'markdown.js'), 'utf8');
  vm.runInContext(source, context);
  // O tempo limite transforma um laço infinito em falha em vez de travar o teste.
  return (markdown) => vm.runInContext(`window.MiniMarkdown.render(${JSON.stringify(markdown)})`, context, { timeout: 1000 });
}

function text(node) {
  if (typeof node === 'string') return node;
  return node.textContent + node.children.map(text).join('');
}

function find(node, tag) {
  if (typeof node === 'string') return null;
  if (node.tag === tag) return node;
  for (const child of node.children) {
    const found = find(child, tag);
    if (found) return found;
  }
  return null;
}

test('italic and bold render without hanging', () => {
  const render = load();
  for (const [markdown, tag, inner] of [['a *it* b', 'em', 'it'], ['a **b** c', 'strong', 'b'], ['x *y* e **z** fim', 'strong', 'z']]) {
    const fragment = render(markdown);
    const paragraph = fragment.children[0];
    assert.strictEqual(paragraph.tag, 'p');
    assert.strictEqual(text(paragraph), markdown.replaceAll('*', ''));
    assert.strictEqual(text(find(paragraph, tag)), inner);
  }
});

test('bold inside a list item keeps the text after it', () => {
  const render = load();
  const list = render('- **Regra 2** do FOLLOW e `código`').children[0];
  assert.strictEqual(list.tag, 'ul');
  assert.strictEqual(text(list), 'Regra 2 do FOLLOW e código');
});
