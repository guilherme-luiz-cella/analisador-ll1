---
class: codegraph.parsedFile
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# parsedFile

A parsed `.go` file with what the second pass needs to resolve calls: its directory, relative path, AST, the import names mapped to module directories and the function declarations.

## Connections
- Collected by [[builder]]: one per non-test file of the module.
