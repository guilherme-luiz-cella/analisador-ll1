---
class: ll1.Preset
path: ll1/presets.go
kind: support
step: 6
---

# Preset

A fixed example grammar the user can pick (20 in all): id, pt-BR name and description, [[PresetGroup]], author, source title and URL for the published ones, whether it was adapted to LL(1), the [[Grammar]] itself, sentences of the language (`Examples`) and sentences that must be rejected (`InvalidExamples`).

## Public functions
- `Presets() []Preset`: every preset, the assignment grammar (`DefaultPresetID = "expressoes"`, built by `Default()`) first, then 4 own examples, then 15 from the literature. Each call builds fresh grammars.
- `PresetByID(id) (Preset, bool)`: lookup by id; false for an unknown id.

## Connections
- Holds [[Grammar]]: built from rule text such as `"E -> E + T | T"` by the unexported `grammarFrom`.
- Grouped by [[PresetGroup]]: Trabalho, Exemplos or Da literatura e da web.

## Notes
`grammarFrom` takes the heads as non-terminals (the first is the start symbol) and every other symbol as a terminal, in order of first appearance; `|` only separates alternatives, so it can never be a terminal. The FIRST and FOLLOW of every preset are checked in `presets_test.go` against the hand-computed tables of `steps/06-presets-research.md`. Only `nao-ll1` has a conflict (`M[E, id]`).
