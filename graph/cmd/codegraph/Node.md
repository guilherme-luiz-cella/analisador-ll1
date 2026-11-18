---
class: codegraph.Node
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# Node

One package, file, type, function or method of the code graph: id, label, kind, package, file and line, first doc paragraph, signature and a source snippet of up to 40 lines.

## Notes
Ids are `pkg:<dir>`, `file:<path>`, `type:<dir>.<Name>` and `func:<dir>.<Recv>.<Name>`; a repeated id gets `:<line>` appended.
