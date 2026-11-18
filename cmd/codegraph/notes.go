package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// noteKinds are the values the "kind" field of a note may take.
var noteKinds = map[string]bool{"support": true, "cli": true, "wasm": true, "tool": true, "web": true}

const (
	noteExt        = ".md"
	vaultReadme    = "README.md"
	frontmatterSep = "---"
)

var (
	wikilinkPattern   = regexp.MustCompile(`\[\[([^\[\]\n]+)\]\]`)
	fencedCodePattern = regexp.MustCompile("(?s)```.*?```")
	inlineCodePattern = regexp.MustCompile("`[^`\n]*`")
)

// Note is one Markdown note of the vault: front matter, title and body.
type Note struct {
	ID    string `json:"id"` // file name without .md, the name [[wikilinks]] use
	Title string `json:"title"`
	Class string `json:"class,omitempty"`
	Path  string `json:"path,omitempty"` // source file under the module root
	Kind  string `json:"kind,omitempty"`
	Step  string `json:"step,omitempty"`
	Dir   string `json:"dir,omitempty"` // folder of the note under the vault root
	Body  string `json:"body"`          // Markdown without the front matter and the title line
}

// NoteEdge is a [[wikilink]] from one note to another.
type NoteEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NotesGraph is the JSON document written by -notes.
type NotesGraph struct {
	Nodes []Note     `json:"nodes"`
	Edges []NoteEdge `json:"edges"`
}

// buildNotes reads every note under vaultDir. Problems that make the vault
// inconsistent with the code (duplicate names, broken links, a missing source
// file, an unknown kind) are all reported together.
func buildNotes(vaultDir, root string) (*NotesGraph, error) {
	var notes []Note
	var problems []error
	err := filepath.WalkDir(vaultDir, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			if p != vaultDir && strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, noteExt) || (name == vaultReadme && filepath.Dir(p) == filepath.Clean(vaultDir)) {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(vaultDir, p)
		if err != nil {
			return err
		}
		note, err := parseNote(filepath.ToSlash(rel), string(data))
		if err != nil {
			problems = append(problems, err)
			return nil
		}
		notes = append(notes, note)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].ID < notes[j].ID })

	byName := map[string]string{} // lower-case name -> id
	for _, n := range notes {
		key := strings.ToLower(n.ID)
		if other, dup := byName[key]; dup {
			problems = append(problems, fmt.Errorf("%s: name already used by another note (%s)", n.ID, other))
		}
		byName[key] = n.ID
	}

	graph := &NotesGraph{Nodes: notes, Edges: []NoteEdge{}}
	seen := map[NoteEdge]bool{}
	for _, n := range notes {
		if n.Path != "" && root != "" {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(n.Path))); err != nil {
				problems = append(problems, fmt.Errorf("%s: path %q does not exist", n.ID, n.Path))
			}
		}
		for _, target := range extractLinks(n.Body) {
			id, ok := byName[strings.ToLower(target)]
			if !ok {
				problems = append(problems, fmt.Errorf("%s: link [[%s]] has no note", n.ID, target))
				continue
			}
			edge := NoteEdge{From: n.ID, To: id}
			if id != n.ID && !seen[edge] {
				seen[edge] = true
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}
	if len(problems) > 0 {
		return nil, errors.Join(problems...)
	}
	sort.Slice(graph.Edges, func(i, j int) bool {
		a, b := graph.Edges[i], graph.Edges[j]
		if a.From != b.From {
			return a.From < b.From
		}
		return a.To < b.To
	})
	return graph, nil
}

// parseNote reads one note; rel is its path under the vault, e.g. "ll1/Parser.md".
func parseNote(rel, text string) (Note, error) {
	id := strings.TrimSuffix(path.Base(rel), noteExt)
	fields, body, err := splitFrontmatter(text)
	if err != nil {
		return Note{}, fmt.Errorf("%s: %w", rel, err)
	}
	note := Note{ID: id, Title: id, Class: fields["class"], Path: fields["path"],
		Kind: fields["kind"], Step: fields["step"]}
	if dir := path.Dir(rel); dir != "." {
		note.Dir = dir
	}
	if note.Kind != "" && !noteKinds[note.Kind] {
		return Note{}, fmt.Errorf("%s: unknown kind %q", rel, note.Kind)
	}
	note.Title, note.Body = splitTitle(id, body)
	return note, nil
}

// splitFrontmatter separates the YAML block (flat "key: value" lines) from the
// Markdown. A note without front matter is allowed and has no fields.
func splitFrontmatter(text string) (map[string]string, string, error) {
	text = strings.ReplaceAll(strings.TrimPrefix(text, "\ufeff"), "\r\n", "\n")
	fields := map[string]string{}
	if !strings.HasPrefix(text, frontmatterSep+"\n") {
		return fields, text, nil
	}
	lines := strings.Split(text, "\n")
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") == frontmatterSep {
			parseFields(strings.Join(lines[1:i], "\n"), fields)
			return fields, strings.TrimLeft(strings.Join(lines[i+1:], "\n"), "\n"), nil
		}
	}
	return nil, "", errors.New("front matter is not closed with ---")
}

func parseFields(block string, fields map[string]string) map[string]string {
	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	return fields
}

// splitTitle takes the first "# Title" line as the title and removes it from the body.
func splitTitle(fallback, body string) (string, string) {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if title, ok := strings.CutPrefix(line, "# "); ok {
			rest := strings.Join(lines[i+1:], "\n")
			return strings.TrimSpace(title), strings.TrimSpace(rest)
		}
		break
	}
	return fallback, strings.TrimSpace(body)
}

// extractLinks returns the target of every [[wikilink]] outside code, without
// the |alias or #heading part.
func extractLinks(body string) []string {
	body = fencedCodePattern.ReplaceAllString(body, "")
	body = inlineCodePattern.ReplaceAllString(body, "")
	var targets []string
	for _, match := range wikilinkPattern.FindAllStringSubmatch(body, -1) {
		target := match[1]
		if i := strings.IndexAny(target, "|#"); i >= 0 {
			target = target[:i]
		}
		if target = strings.TrimSpace(target); target != "" {
			targets = append(targets, target)
		}
	}
	return targets
}
