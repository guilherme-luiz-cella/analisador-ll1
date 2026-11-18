// Tiny Markdown renderer for the notes panel. It builds DOM nodes (never HTML
// strings), so note text can't inject markup. Supported: headings, paragraphs,
// bullet and numbered lists, fenced code, block quotes, rules, `code`, **bold**,
// *italic*, [text](https://...) and [[wikilinks]].
(() => {
  'use strict';

  const INLINE = /`([^`]+)`|\[\[([^\]\n]+)\]\]|\[([^\]\n]+)\]\((https?:\/\/[^)\s]+)\)|\*\*([^*\n]+)\*\*|\*([^*\s][^*\n]*?)\*/g;
  const HEADING_SHIFT = 2; // "# Title" is the panel title, so "##" renders as an h3
  const MAX_HEADING = 6;

  function make(tag, props = {}, ...children) {
    const node = document.createElement(tag);
    Object.assign(node, props);
    node.append(...children);
    return node;
  }

  // options.wikilink(target, label) returns a Node, or null when the note does not exist.
  function inline(text, options) {
    const out = [];
    let last = 0;
    INLINE.lastIndex = 0;
    for (let m = INLINE.exec(text); m; m = INLINE.exec(text)) {
      if (m.index > last) out.push(text.slice(last, m.index));
      if (m[1] !== undefined) {
        out.push(make('code', { textContent: m[1] }));
      } else if (m[2] !== undefined) {
        const [target, alias] = m[2].split('|');
        const name = target.split('#')[0].trim();
        const node = options.wikilink ? options.wikilink(name, (alias || name).trim()) : null;
        out.push(node || (alias || name).trim());
      } else if (m[3] !== undefined) {
        out.push(make('a', { href: m[4], target: '_blank', rel: 'noopener', textContent: m[3] }));
      } else if (m[5] !== undefined) {
        out.push(make('strong', {}, ...inline(m[5], options)));
      } else {
        out.push(make('em', {}, ...inline(m[6], options)));
      }
      last = INLINE.lastIndex;
    }
    if (last < text.length) out.push(text.slice(last));
    return out;
  }

  const isBlank = (line) => line.trim() === '';
  const listMatch = (line) => /^(\s*)([-*]|\d+[.)])\s+(.*)$/.exec(line);

  function render(markdown, options = {}) {
    const lines = markdown.replace(/\r\n/g, '\n').split('\n');
    const fragment = document.createDocumentFragment();
    let i = 0;
    while (i < lines.length) {
      const line = lines[i];
      if (isBlank(line)) { i += 1; continue; }

      const fence = /^```/.exec(line);
      if (fence) {
        const code = [];
        i += 1;
        while (i < lines.length && !/^```/.test(lines[i])) { code.push(lines[i]); i += 1; }
        i += 1;
        fragment.append(make('pre', { className: 'code' }, make('code', { textContent: code.join('\n') })));
        continue;
      }

      const heading = /^(#{1,6})\s+(.*)$/.exec(line);
      if (heading) {
        const level = Math.min(MAX_HEADING, heading[1].length + HEADING_SHIFT - 1);
        fragment.append(make(`h${level}`, {}, ...inline(heading[2], options)));
        i += 1;
        continue;
      }

      if (/^\s*([-*_])\1{2,}\s*$/.test(line)) { fragment.append(make('hr')); i += 1; continue; }

      if (/^>\s?/.test(line)) {
        const quote = [];
        while (i < lines.length && /^>\s?/.test(lines[i])) { quote.push(lines[i].replace(/^>\s?/, '')); i += 1; }
        fragment.append(make('blockquote', {}, ...inline(quote.join(' '), options)));
        continue;
      }

      const item = listMatch(line);
      if (item) {
        const ordered = /\d/.test(item[2]);
        const list = make(ordered ? 'ol' : 'ul');
        while (i < lines.length && listMatch(lines[i])) {
          const parts = [listMatch(lines[i])[3]];
          i += 1;
          // Indented lines that are not new items continue the current item.
          while (i < lines.length && !isBlank(lines[i]) && !listMatch(lines[i]) && /^\s+\S/.test(lines[i])) {
            parts.push(lines[i].trim());
            i += 1;
          }
          list.append(make('li', {}, ...inline(parts.join(' '), options)));
        }
        fragment.append(list);
        continue;
      }

      const paragraph = [];
      while (i < lines.length && !isBlank(lines[i]) && !/^(```|#{1,6}\s|>)/.test(lines[i]) && !listMatch(lines[i])) {
        paragraph.push(lines[i].trim());
        i += 1;
      }
      fragment.append(make('p', {}, ...inline(paragraph.join(' '), options)));
    }
    return fragment;
  }

  window.MiniMarkdown = { render };
})();
