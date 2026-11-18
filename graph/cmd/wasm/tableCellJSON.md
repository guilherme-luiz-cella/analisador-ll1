---
class: wasm.tableCellJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# tableCellJSON

One cell of the parsing table: the column and the production `number` (0 means an empty, error cell). A conflicting cell also lists every production claiming it in `conflicts`.

## Connections
- Mirrors [[Production]]: the number of the production the cell holds; the page looks up its body in [[grammarJSON]].
- Copies [[ConflictCell]]: the production numbers of a conflicting cell.
