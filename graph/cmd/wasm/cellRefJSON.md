---
class: wasm.cellRefJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# cellRefJSON

A pointer to a table cell (`nonTerminal`, `column`). A step carries one when the stack top is a non-terminal with an entry for the current token, so the page can highlight the cell that step read.

## Connections
- Points into [[tableJSON]]: the cell it names.
