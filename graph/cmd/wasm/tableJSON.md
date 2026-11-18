---
class: wasm.tableJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# tableJSON

The whole parsing table for the page: `columns`, `rows` and the list of `conflicts` (empty for this grammar).

## Connections
- Holds [[tableRowJSON]]: one per non-terminal.
- Mirrors [[Table]]: columns and conflicts are copied from it.
