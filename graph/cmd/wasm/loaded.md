---
class: wasm.loaded
path: cmd/wasm/main.go
kind: wasm
step: 6
---

# loaded

A [[Preset]] together with its parsing [[Table]], built once when the WebAssembly program starts, so switching grammars on the page never rebuilds a table.

## Connections
- Holds [[Preset]]: the grammar and its texts.
- Holds [[Table]]: used by `parse` and `table`.
