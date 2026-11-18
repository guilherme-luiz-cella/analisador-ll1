---
class: wasm.grammarJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# grammarJSON

The grammar for the page: start symbol, productions, terminals and non-terminals. The page draws the grouped form (`E' → + T E' | - T E' | ε`) itself from the productions.

## Connections
- Holds [[productionJSON]]: the numbered rules.
- Mirrors [[Grammar]]: start symbol and symbol lists.
