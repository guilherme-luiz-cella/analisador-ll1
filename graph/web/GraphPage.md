---
class: web.graph
path: web/graph/graph.js
kind: web
step: 3
---

# GraphPage

The `/graph` page: a force-directed graph in the style of Obsidian's graph view, drawn with the pinned `vis-network` 9.1.9 file kept in `web/graph/vendor/`. The "Notas" mode shows this vault; the "Código" mode shows the Go code graph (packages, files, types, functions and calls).

## Public functions
- Mode switch "Notas | Código": loads `notes.json` or `graph.json` on first use.
- Node click: opens the side panel; a note is rendered as Markdown with a link to its source file on GitHub, a code node shows signature, doc and snippet.
- Search, kind filters, "Ajustar à tela".

## Connections
- Reads [[NotesGraph]]: `notes.json`, written by `codegraph -notes`.
- Reads [[Graph]]: `graph.json`, written by `codegraph`.
- Renders bodies with [[MiniMarkdown]]: headings, lists, code and clickable wikilinks.

## Notes
Node size follows the link count; labels fade in with the zoom (hubs first); hovering a node dims everything but its neighbors. Dark by default, light with `prefers-color-scheme: light`; the CSS colors are read once and cached until the color scheme changes. The URL hash names the open node (`#notas/<nota>` or `#codigo/<id>`); a malformed hash just opens the notes view.
