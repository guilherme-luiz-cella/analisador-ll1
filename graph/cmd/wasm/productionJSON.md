---
class: wasm.productionJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# productionJSON

A grammar rule for the page: `number`, `head`, `body` and `first`, the FIRST of the body computed by Go, which the LL(1) condition panel uses.

## Connections
- Mirrors [[Production]]: same fields, plus FIRST of the body.
- Uses [[Sets]]: `FirstOfSequence` fills `first`, so the page has no second FIRST implementation.
