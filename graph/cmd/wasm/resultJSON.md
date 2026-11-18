---
class: wasm.resultJSON
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# resultJSON

What `ll1.parse` returns: `accepted`, the Portuguese `message`, the tokens and the steps.

## Connections
- Mirrors [[Result]]: the verdict and message are copied.
- Holds [[tokenJSON]]: the recognized tokens.
- Holds [[stepJSON]]: the full trace.
