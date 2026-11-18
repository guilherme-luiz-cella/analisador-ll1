---
class: wasm.presetJSON
path: cmd/wasm/main.go
kind: wasm
step: 6
---

# presetJSON

One entry of `ll1.presets()`: id, name, description, group label, `isDefault`, author, source, URL, `adapted`, `isLL1` (the table has no conflicts), the terminals and the valid and invalid example sentences. The page builds the "Gramática" dropdown and the example buttons from it.

## Connections
- Mirrors [[Preset]]: every field except the grammar, which comes from `ll1.grammar(id)`.
- Labels from [[PresetGroup]]: `group` is its pt-BR label.
