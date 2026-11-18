package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range files {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func hasEdge(graph *Graph, from, to, kind string) bool {
	for _, e := range graph.Edges {
		if e.From == from && e.To == to && e.Kind == kind {
			return true
		}
	}
	return false
}

func TestGraphHasNodesAndResolvedEdges(t *testing.T) {
	root := writeTree(t, map[string]string{
		"go.mod": "module demo\n\ngo 1.22\n",
		"lib/lib.go": `package lib

// Box holds a value.
type Box struct{ v int }

// Get returns the value.
func (b Box) Get() int { return b.v }

// Make builds a Box.
func Make() Box { return Box{} }
`,
		"lib/lib_test.go": "package lib\n\nfunc TestIgnored() {}\n",
		"main.go": `package main

import "demo/lib"

func main() {
	b := lib.Make()
	_ = b.Get()
	helper()
}

func helper() {}
`,
	})

	graph, err := buildGraph(root)
	if err != nil {
		t.Fatal(err)
	}

	ids := map[string]Node{}
	for _, n := range graph.Nodes {
		ids[n.ID] = n
	}
	for _, id := range []string{"pkg:lib", "pkg:.", "file:lib/lib.go", "type:lib.Box", "func:lib.Box.Get", "func:lib.Make", "func:..main", "func:..helper"} {
		if _, ok := ids[id]; !ok {
			t.Errorf("missing node %s", id)
		}
	}
	if _, ok := ids["func:lib.TestIgnored"]; ok {
		t.Error("test files must be skipped")
	}
	if got := ids["type:lib.Box"].Doc; got != "Box holds a value." {
		t.Errorf("type doc = %q", got)
	}

	wantEdges := [][3]string{
		{"pkg:.", "pkg:lib", edgeImports},
		{"file:lib/lib.go", "type:lib.Box", edgeContains},
		{"type:lib.Box", "func:lib.Box.Get", edgeMethod},
		{"func:..main", "func:lib.Make", edgeCalls},
		{"func:..main", "func:lib.Box.Get", edgeCalls},
		{"func:..main", "func:..helper", edgeCalls},
	}
	for _, w := range wantEdges {
		if !hasEdge(graph, w[0], w[1], w[2]) {
			t.Errorf("missing %s edge %s -> %s", w[2], w[0], w[1])
		}
	}
}

func TestSnippetIsCutAtTheLimit(t *testing.T) {
	body := "package p\n\nfunc long() {\n"
	for i := 0; i < 60; i++ {
		body += "\t_ = 1\n"
	}
	body += "}\n"
	root := writeTree(t, map[string]string{"go.mod": "module demo\n", "p.go": body})

	graph, err := buildGraph(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range graph.Nodes {
		if n.ID == "func:..long" {
			lines := 0
			for _, c := range n.Code {
				if c == '\n' {
					lines++
				}
			}
			if lines != maxSnippetLines {
				t.Errorf("snippet has %d line breaks, want %d", lines, maxSnippetLines)
			}
			return
		}
	}
	t.Fatal("func:..long not found")
}
