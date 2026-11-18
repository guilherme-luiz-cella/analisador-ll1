---
class: codegraph.NotesGraph
path: cmd/codegraph/notes.go
kind: tool
step: 3
---

# NotesGraph

The `notes.json` document and the code that builds it: reads every `.md` under the vault (except its `README.md`), parses front matter, title and body, and turns wikilinks into edges.

## Public functions
- `buildNotes(vaultDir, root) (*NotesGraph, error)`: returns all problems together: duplicate names, a link with no note, a `path` that does not exist under `root`, an unknown kind, an unclosed front matter.

## Connections
- Holds [[Note]]: sorted by id.
- Holds [[NoteEdge]]: sorted, without duplicates.

## Notes
The page draws the notes as an Obsidian-style graph and shows the Markdown body in the side panel.
