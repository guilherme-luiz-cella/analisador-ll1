---
class: wasm.grammarJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# grammarJSON

The grammar for the page: start symbol, productions, the grouped lines (`E' -> + T E' | - T E' | ε`), terminals and non-terminals.

## Connections
- Holds [[productionJSON]]: the numbered rules.
- Mirrors [[Grammar]]: start symbol, symbol lists and `String()` lines.
