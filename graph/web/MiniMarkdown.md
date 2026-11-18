---
class: web.MiniMarkdown
path: web/graph/markdown.js
kind: web
step: 3
---

# MiniMarkdown

A small Markdown renderer for the notes panel. It builds DOM nodes, so note text can never inject HTML.

## Public functions
- `MiniMarkdown.render(markdown, { wikilink })`: returns a document fragment with headings, paragraphs, bullet and numbered lists, fenced code, quotes, `code`, bold, italic, `http(s)` links and `[[wikilinks]]`. The `wikilink(name, label)` callback returns the link element, or null so the text stays plain.

## Connections
- Used by [[GraphPage]]: the callback turns a wikilink into a jump to that node.

## Notes
Bold and italic call the inline pass again for their inner text, so each call uses its own copy of the inline regular expression; a shared global one lost its position and looped forever. `test/markdown.test.js` (`node --test`) guards this.
