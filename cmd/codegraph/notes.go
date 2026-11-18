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

// noteKinds são os valores aceitos no campo "kind" de uma nota.
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

// Note é uma nota Markdown do cofre: cabeçalho, título e corpo.
type Note struct {
	ID    string `json:"id"` // nome do arquivo sem .md, o nome usado nos [[wikilinks]]
	Title string `json:"title"`
	Class string `json:"class,omitempty"`
	Path  string `json:"path,omitempty"` // arquivo de código, a partir da raiz do módulo
	Kind  string `json:"kind,omitempty"`
	Step  string `json:"step,omitempty"`
	Dir   string `json:"dir,omitempty"` // pasta da nota dentro do cofre
	Body  string `json:"body"`          // Markdown sem o cabeçalho e sem a linha do título
}

// NoteEdge é um [[wikilink]] de uma nota para outra.
type NoteEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// NotesGraph é o documento JSON gravado com -notes.
type NotesGraph struct {
	Nodes []Note     `json:"nodes"`
	Edges []NoteEdge `json:"edges"`
}

// buildNotes lê todas as notas de vaultDir. Os problemas que deixam o cofre
// diferente do código (nome repetido, link quebrado, arquivo de código que
// não existe, kind desconhecido) são todos informados juntos.
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

	byName := map[string]string{} // nome em minúsculas -> id
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

// parseNote lê uma nota; rel é o caminho dentro do cofre, ex.: "ll1/Parser.md".
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

// splitFrontmatter separa o bloco YAML (linhas simples "chave: valor") do
// Markdown. Uma nota sem cabeçalho é aceita e fica sem campos.
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

// parseFields lê as linhas "chave: valor" para dentro de fields; linhas que
// começam com # são comentários do YAML.
func parseFields(block string, fields map[string]string) {
	for _, line := range strings.Split(block, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		fields[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
}

// splitTitle usa a primeira linha "# Título" como título e a tira do corpo.
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

// extractLinks devolve o destino de cada [[wikilink]] fora de código, sem a
// parte |apelido ou #seção.
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
