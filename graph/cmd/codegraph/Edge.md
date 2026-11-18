---
class: codegraph.Edge
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# Edge

A directed relation between two [[Node]] ids of the code graph. Kinds: `imports` (package to package), `contains` (package to file, file to type or function), `method` (type to its methods) and `calls` (function to function).

## Connections
- Joins [[Node]]: `from` and `to` are node ids.
