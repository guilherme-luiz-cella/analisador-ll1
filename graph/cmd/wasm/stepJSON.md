---
class: wasm.stepJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# stepJSON

One row of the trace for the page: number, stack, remaining input, action and the optional consulted cell.

## Connections
- Mirrors [[Step]]: same four fields.
- Holds [[cellRefJSON]]: the table cell this step consulted, or null.
