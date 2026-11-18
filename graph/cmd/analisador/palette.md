---
class: analisador.palette
path: cmd/analisador/render.go
kind: cli
step: 1
---

# palette

Terminal drawing helpers of `render.go`. `palette` wraps text in ANSI styles or leaves it plain when disabled.

## Public functions
- `paint(style, text) string`: wraps `text` in the style, or returns it as is when colors are off.
- `box(p, lines...) string`: double-line frame used for the title and section headers.
- `table(p, headers, rows, style) string`: aligned grid; styles are applied after padding so escape codes never break the columns.

## Connections
- Used by [[app]]: every screen goes through it.

## Notes
Widths are counted in runes, so accented text stays aligned.
