package main

import (
	"math/rand"
	"strings"
	"testing"
)

func runWith(input string) string {
	var out strings.Builder
	run(strings.NewReader(input), &out, false, rand.New(rand.NewSource(1)))
	return out.String()
}

func TestTypedValidSentenceShowsTheTraceAndIsAccepted(t *testing.T) {
	out := runWith("4\na+b*c\n\n0\n")
	for _, want := range []string{"Passo", "Pilha", "Entrada", "Ação", "Casa 'id' (a)", "Aceita", "ACEITA", "Pressione Enter", "Até logo!"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
}

func TestTypedInvalidSentenceIsRejectedWithTheReason(t *testing.T) {
	out := runWith("4\na+#\n\n0\n")
	if !strings.Contains(out, "REJEITADA") || !strings.Contains(out, "posição 3") {
		t.Errorf("output lacks the rejection:\n%s", out)
	}
}

func TestEveryMenuOptionRunsAndReturnsToTheMenu(t *testing.T) {
	out := runWith("1\n\n2\n\n3\n\n5\n\n6\n\n9\n0\n")
	for _, want := range []string{"E' -> + T E' | - T E' | ε", "FOLLOW", "{ +, -, *, /, ), $ }", "Nenhum conflito", "ACEITA", "REJEITADA", "Opção inválida."} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q", want)
		}
	}
	if got := strings.Count(out, "Escolha uma opção"); got != 7 {
		t.Errorf("menu shown %d times, want 7", got)
	}
}

func TestEndOfInputExitsCleanly(t *testing.T) {
	if out := runWith("4\n"); !strings.HasSuffix(out, "Até logo!\n") {
		t.Errorf("output ends with %q", out[len(out)-20:])
	}
}

func TestTableKeepsColumnsAligned(t *testing.T) {
	out := table(palette{enabled: true}, []string{"A", "Bε"}, [][]string{{"xyz", "1"}}, func(_, _ int, _ string) string { return styleGood })
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 5 {
		t.Fatalf("lines = %d", len(lines))
	}
	if !strings.Contains(lines[3], styleGood+"xyz"+styleReset) {
		t.Errorf("row = %q", lines[3])
	}
}
