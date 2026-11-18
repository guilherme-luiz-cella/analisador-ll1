---
class: wasm.tableJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# tableJSON

The whole parsing table for the page: `columns`, `rows` and the list of `conflicts` (empty for an LL(1) grammar).

## Connections
- Holds [[tableRowJSON]]: one per non-terminal.
- Mirrors [[Table]]: columns and conflicts are copied from it.
