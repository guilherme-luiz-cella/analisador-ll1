package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const noteA = "---\nclass: demo.A\npath: a.go\nkind: support\nstep: 1\n---\n\n# A\n\nUses [[B]] and [[b|alias]]; `[[code]]` is not a link.\n\n```\n[[also-code]]\n```\n"

func TestNotesParseFrontmatterTitleAndLinks(t *testing.T) {
	vault := writeTree(t, map[string]string{
		"README.md": "# vault\n",
		"x/A.md":    noteA,
		"B.md":      "---\r\nkind: cli\r\n---\r\n# Bee\r\nSee [[A#Heading]].\r\n",
	})
	root := writeTree(t, map[string]string{"a.go": "package a\n"})

	graph, err := buildNotes(vault, root)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) != 2 {
		t.Fatalf("got %d notes, want 2 (README is not a note)", len(graph.Nodes))
	}
	a, b := graph.Nodes[0], graph.Nodes[1]
	if a.ID != "A" || a.Class != "demo.A" || a.Path != "a.go" || a.Kind != "support" || a.Step != "1" || a.Dir != "x" {
		t.Errorf("note A parsed wrong: %+v", a)
	}
	if a.Title != "A" || strings.HasPrefix(a.Body, "# ") || strings.HasPrefix(a.Body, "---") {
		t.Errorf("title or body wrong: %q / %q", a.Title, a.Body)
	}
	if b.Title != "Bee" || b.Kind != "cli" {
		t.Errorf("note B parsed wrong: %+v", b)
	}
	want := []NoteEdge{{From: "A", To: "B"}, {From: "B", To: "A"}}
	if len(graph.Edges) != len(want) || graph.Edges[0] != want[0] || graph.Edges[1] != want[1] {
		t.Errorf("edges = %+v, want %+v", graph.Edges, want)
	}
}

func TestNotesReportEveryProblem(t *testing.T) {
	vault := writeTree(t, map[string]string{
		"A.md":     "---\npath: missing.go\nkind: bogus\n---\n# A\n",
		"B.md":     "---\npath: missing.go\n---\n# B\n[[Nowhere]]\n",
		"sub/B.md": "# B again\n",
	})
	_, err := buildNotes(vault, t.TempDir())
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, part := range []string{"unknown kind", "has no note", "does not exist", "already used"} {
		if !strings.Contains(err.Error(), part) {
			t.Errorf("error does not mention %q:\n%v", part, err)
		}
	}
}

func TestFrontmatterMustBeClosed(t *testing.T) {
	if _, _, err := splitFrontmatter("---\nclass: x\n# title\n"); err == nil {
		t.Error("expected an error for an unclosed front matter")
	}
}

// O cofre de notas real tem de bater com o código. O teste é pulado quando a
// pasta não existe (graph/ nem sempre vai junto com o código-fonte).
func TestProjectVaultIsConsistent(t *testing.T) {
	vault := filepath.Join("..", "..", "graph")
	if _, err := os.Stat(vault); err != nil {
		t.Skip("no graph/ folder")
	}
	graph, err := buildNotes(vault, filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Nodes) == 0 || len(graph.Edges) == 0 {
		t.Errorf("empty vault: %d notes, %d links", len(graph.Nodes), len(graph.Edges))
	}
}
