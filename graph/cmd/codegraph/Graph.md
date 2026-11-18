---
class: codegraph.Graph
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# Graph

The code-graph JSON document: module name, nodes and edges, both sorted so the output is stable between runs.

## Connections
- Holds [[Node]]: every declared package, file, type and function.
- Holds [[Edge]]: only edges whose two ends exist.
