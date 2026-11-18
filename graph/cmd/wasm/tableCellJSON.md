---
class: wasm.tableCellJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# tableCellJSON

One cell of the parsing table: the column, the production `number` (0 means an empty, error cell) and its printed `text`. A conflicting cell also lists every production claiming it in `conflicts`.

## Connections
- Mirrors [[Production]]: number and text of the production the cell holds.
- Copies [[ConflictCell]]: the production numbers of a conflicting cell.
