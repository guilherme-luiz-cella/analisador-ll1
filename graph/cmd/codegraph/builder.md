---
class: codegraph.builder
path: cmd/codegraph/main.go
kind: tool
step: 3
---

# builder

Builds the code graph with `go/parser` and `go/ast`. Pass one (`declare`) creates the nodes and the structural edges; pass two (`link`) resolves calls between functions declared in the module.

## Public functions
- `buildGraph(root) (*Graph, error)`: the whole pipeline for a module directory.

## Connections
- Reads [[parsedFile]]: every non-test Go file, skipping `web`, `site`, `dist`, `vendor` and hidden folders.
- Creates [[Node]]: one per package, file, type, function and method.
- Creates [[Edge]]: imports, contains, method and calls.
- Returns [[Graph]]: sorted nodes and edges.

## Notes
A call `x.M()` resolves to a method only when `M` is unique in the module; a plain call or a function passed as a value resolves inside its own package.
