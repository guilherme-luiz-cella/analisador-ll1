---
class: wasm.main
path: cmd/wasm/main.go
kind: wasm
step: 2
---

# Wasm

Entry point of the browser build (`GOOS=js GOARCH=wasm`). It exposes the `ll1` package as `globalThis.ll1` and keeps the program alive with `select {}`.

## Public functions
- `ll1.presets()`: JSON list of the preset grammars ([[presetJSON]]).
- `ll1.grammar(id?)`, `ll1.sets(id?)`, `ll1.table(id?)`: JSON strings for the grammar, the FIRST and FOLLOW rows and the parsing table.
- `ll1.parse(sentence, id?)`: JSON string with the verdict, tokens and steps.
- `ll1.generate(id?)`, `ll1.generateInvalid(id?)`: the sentence itself, as plain text.

The optional `id` picks a preset; a missing or unknown id means the assignment grammar.

## Connections
- Uses [[Preset]]: every preset is loaded at start as a [[loaded]] value with its table.
- Uses [[Table]]: built once per preset; cells and conflicts are copied to [[tableJSON]].
- Calls [[Parser]]: `parse` runs it and converts the [[Result]].
- Calls [[Generator]]: for the two generate buttons.
- Builds [[grammarJSON]], [[setsRowJSON]], [[tableJSON]] and [[resultJSON]]: the shapes the page receives.

## Notes
The JSON types below mirror the Go types so that the page never sees Go structs. Errors in encoding come back as `{"error": ...}`.
