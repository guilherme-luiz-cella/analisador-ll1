---
class: codegraph.main
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# Codegraph

Command that feeds the `/graph` page, using only the standard library. Without flags it reads the Go source of the module and writes the code graph; with `-notes` it reads the Markdown vault and writes the note graph.

## Public functions
- `codegraph -root . -o web/graph/graph.json`: code graph, built by [[builder]].
- `codegraph -notes graph -o web/graph/notes.json`: note graph, built by [[NotesGraph]]; exits with an error if a link, a kind or a source path in the vault is wrong.

## Connections
- Writes [[Graph]]: the code-level document (`graph.json`).
- Writes [[NotesGraph]]: the vault document (`notes.json`).
- Runs [[builder]]: one pass to declare nodes, one to link calls.

## Notes
`build-site.ps1` runs both modes while building `site/`.
