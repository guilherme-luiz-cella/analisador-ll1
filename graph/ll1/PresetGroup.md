---
class: ll1.PresetGroup
path: ll1/presets.go
kind: support
step: 6
---

# PresetGroup

Where a [[Preset]] comes from: `GroupAssignment` (the graded grammar), `GroupExamples` (written for this project) or `GroupLiterature` (published grammars). The web groups the dropdown with one `<optgroup>` per group and the CLI prints one heading per group.

## Public functions
- `Label() string`: the pt-BR heading, "Trabalho", "Exemplos" or "Da literatura e da web".
