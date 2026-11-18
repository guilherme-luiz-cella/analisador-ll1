---
class: codegraph.Note
path: cmd/codegraph/notes.go
kind: tool
step: 3
---

# Note

One vault note as the page receives it: `id` (file name without `.md`), `title` (first `# ` line), the front matter fields `class`, `path`, `kind` and `step`, the folder `dir` and the Markdown `body` without front matter and title.

## Public functions
- `parseNote(rel, text) (Note, error)`: reads one note; rejects a kind outside `support`, `cli`, `wasm`, `tool` and `web`.

## Notes
A note may omit the front matter; its title then falls back to the file name.
