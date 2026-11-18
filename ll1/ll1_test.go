package ll1

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

const generatedSentences = 500

func TestFirstSetsMatchTheTextbook(t *testing.T) {
	want := map[string][]string{
		"E":  {"num", "id", "("},
		"E'": {"+", "-", Epsilon},
		"T":  {"num", "id", "("},
		"T'": {"*", "/", Epsilon},
		"F":  {"num", "id", "("},
	}
	assertSets(t, "FIRST", Default().First(), want)
}

func TestFollowSetsMatchTheTextbook(t *testing.T) {
	want := map[string][]string{
		"E":  {")", EndMarker},
		"E'": {")", EndMarker},
		"T":  {"+", "-", ")", EndMarker},
		"T'": {"+", "-", ")", EndMarker},
		"F":  {"*", "/", "+", "-", ")", EndMarker},
	}
	assertSets(t, "FOLLOW", Default().Follow(), want)
}

func TestSetsAreOrderedLikeTheTerminalsWithTheExtraSymbolLast(t *testing.T) {
	g := Default()
	if got, want := g.First()["F"], []string{"id", "num", "("}; !reflect.DeepEqual(got, want) {
		t.Errorf("FIRST(F) = %v, want %v", got, want)
	}
	if got, want := g.Follow()["F"], []string{"+", "-", "*", "/", ")", EndMarker}; !reflect.DeepEqual(got, want) {
		t.Errorf("FOLLOW(F) = %v, want %v", got, want)
	}
}

func TestFirstOfSequence(t *testing.T) {
	g := Default()
	cases := []struct {
		seq  []string
		want []string
	}{
		{[]string{"T", "E'"}, []string{"id", "num", "("}},
		{[]string{"E'", "T'"}, []string{"+", "-", "*", "/", Epsilon}},
		{[]string{"T'", ")"}, []string{"*", "/", ")"}},
		{[]string{Epsilon}, []string{Epsilon}},
		{nil, []string{Epsilon}},
	}
	for _, c := range cases {
		if got := g.FirstOfSequence(c.seq); !reflect.DeepEqual(got, c.want) {
			t.Errorf("FIRST(%v) = %v, want %v", c.seq, got, c.want)
		}
	}
}

func TestGrammarPrintsGroupedAlternatives(t *testing.T) {
	want := "E -> T E'\n" +
		"E' -> + T E' | - T E' | ε\n" +
		"T -> F T'\n" +
		"T' -> * F T' | / F T' | ε\n" +
		"F -> ( E ) | id | num\n"
	if got := Default().String(); got != want {
		t.Errorf("grammar:\n%s\nwant:\n%s", got, want)
	}
	if got := Default().Productions[1].String(); got != "E' -> + T E'" {
		t.Errorf("production 2 = %q", got)
	}
}

func TestTableHasNoConflictsAndTheTextbookCells(t *testing.T) {
	table := BuildTable(Default())
	if len(table.Conflicts) != 0 {
		t.Fatalf("conflicts: %v", table.Conflicts)
	}
	want := map[string]map[string]int{
		"E":  {"id": 1, "num": 1, "(": 1},
		"E'": {"+": 2, "-": 3, ")": 4, EndMarker: 4},
		"T":  {"id": 5, "num": 5, "(": 5},
		"T'": {"*": 6, "/": 7, "+": 8, "-": 8, ")": 8, EndMarker: 8},
		"F":  {"(": 9, "id": 10, "num": 11},
	}
	if !reflect.DeepEqual(table.Cells, want) {
		t.Errorf("cells = %v, want %v", table.Cells, want)
	}
	if got, want := table.Columns, []string{"id", "num", "+", "-", "*", "/", "(", ")", EndMarker}; !reflect.DeepEqual(got, want) {
		t.Errorf("columns = %v, want %v", got, want)
	}
	if p, ok := table.Production("T'", "+"); !ok || p.Number != 8 || !p.IsEmpty() {
		t.Errorf("M[T', +] = %v %v, want production 8", p, ok)
	}
	if _, ok := table.Production("F", "+"); ok {
		t.Error("M[F, +] should be empty")
	}
}

func TestTableRecordsConflictsOfANonLL1Grammar(t *testing.T) {
	g := &Grammar{
		Start:        "S",
		NonTerminals: []string{"S"},
		Terminals:    []string{"a"},
		Productions: []Production{
			{Number: 1, Head: "S", Body: []string{"a"}},
			{Number: 2, Head: "S", Body: []string{"a", "S"}},
		},
	}
	if table := BuildTable(g); len(table.Conflicts) != 1 {
		t.Errorf("conflicts = %v, want one", table.Conflicts)
	}
}

func TestTokenizeRecognizesEveryTokenKind(t *testing.T) {
	tokens, err := Tokenize(" soma_1+42 * (x)-y/_z")
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{"id", "soma_1", 2}, {"+", "+", 8}, {"num", "42", 9}, {"*", "*", 12},
		{"(", "(", 14}, {"id", "x", 15}, {")", ")", 16}, {"-", "-", 17},
		{"id", "y", 18}, {"/", "/", 19}, {"id", "_z", 20}, {EndMarker, EndMarker, 22},
	}
	if !reflect.DeepEqual(tokens, want) {
		t.Errorf("tokens = %v\nwant %v", tokens, want)
	}
}

func TestTokenizeReportsTheIllegalCharacterAndItsPosition(t *testing.T) {
	_, err := Tokenize("a + #b")
	lexErr, ok := err.(*LexicalError)
	if !ok || lexErr.Char != '#' || lexErr.Pos != 5 {
		t.Errorf("err = %v, want '#' at 5", err)
	}
}

func TestParserAcceptsValidSentences(t *testing.T) {
	table := BuildTable(Default())
	for _, input := range []string{"a", "3", "a1", "a+b*c", "(a+b)*c", "((x))", "a-b/c+d*(e-1)"} {
		result := Parse(table, input)
		if !result.Accepted {
			t.Errorf("%q rejected: %s", input, result.Message)
		}
		assertTraceShape(t, input, result)
	}
}

func TestParserRejectsInvalidSentences(t *testing.T) {
	table := BuildTable(Default())
	for _, input := range []string{"", "+", "1a", "a+", "(a", "a)", "a b", "a**b", "()", "a+#"} {
		result := Parse(table, input)
		if result.Accepted {
			t.Errorf("%q accepted", input)
		}
		if !strings.Contains(result.Message, "REJEITADA") {
			t.Errorf("%q message = %q", input, result.Message)
		}
		assertTraceShape(t, input, result)
	}
}

func TestParserTraceForASimpleSentence(t *testing.T) {
	result := Parse(BuildTable(Default()), "soma")
	want := []Step{
		{1, []string{"$", "E"}, []string{"id", "$"}, "E -> T E'"},
		{2, []string{"$", "E'", "T"}, []string{"id", "$"}, "T -> F T'"},
		{3, []string{"$", "E'", "T'", "F"}, []string{"id", "$"}, "F -> id"},
		{4, []string{"$", "E'", "T'", "id"}, []string{"id", "$"}, "Casa 'id' (soma)"},
		{5, []string{"$", "E'", "T'"}, []string{"$"}, "T' -> ε"},
		{6, []string{"$", "E'"}, []string{"$"}, "E' -> ε"},
		{7, []string{"$"}, []string{"$"}, "Aceita"},
	}
	if !reflect.DeepEqual(result.Steps, want) {
		t.Errorf("steps:\n%v\nwant:\n%v", result.Steps, want)
	}
}

func TestParserNamesTheOffendingTokenAndTheExpectation(t *testing.T) {
	table := BuildTable(Default())
	cases := map[string]string{
		"a)":  "token ')' inesperado na posição 2, esperado fim da entrada",
		"(a":  "fim da entrada inesperado, esperado ')'",
		"a+*": "token '*' inesperado na posição 3: M[T, *] vazia",
	}
	for input, want := range cases {
		if got := Parse(table, input).Message; !strings.Contains(got, want) {
			t.Errorf("%q message = %q, want it to contain %q", input, got, want)
		}
	}
}

func TestParserMatchesOperatorsWithoutRepeatingTheLexeme(t *testing.T) {
	steps := Parse(BuildTable(Default()), "a+b").Steps
	found := false
	for _, step := range steps {
		found = found || step.Action == "Casa '+'"
	}
	if !found {
		t.Errorf("no \"Casa '+'\" step in %v", steps)
	}
}

func TestParserReportsTheLexicalErrorPosition(t *testing.T) {
	result := Parse(BuildTable(Default()), "a+#")
	if len(result.Steps) != 1 || !strings.Contains(result.Steps[0].Action, "posição 3") {
		t.Errorf("steps = %v", result.Steps)
	}
	if got, want := result.Steps[0].Input, []string{"id", "+", "#", "$"}; !reflect.DeepEqual(got, want) {
		t.Errorf("input = %v, want %v", got, want)
	}
}

func TestGeneratedSentencesAreAccepted(t *testing.T) {
	g := Default()
	table := BuildTable(g)
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < generatedSentences; i++ {
		sentence := Generate(g, rng, 1+i%8)
		if result := Parse(table, sentence); !result.Accepted {
			t.Fatalf("generated %q rejected: %s", sentence, result.Message)
		}
	}
}

func TestGenerateWithZeroDepthGivesASingleOperand(t *testing.T) {
	sentence := Generate(Default(), rand.New(rand.NewSource(2)), 0)
	if strings.Contains(sentence, " ") {
		t.Errorf("sentence = %q, want one token", sentence)
	}
}

func TestGeneratedInvalidSentencesAreRejected(t *testing.T) {
	g := Default()
	table := BuildTable(g)
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < generatedSentences; i++ {
		sentence := GenerateInvalid(g, rng)
		if Parse(table, sentence).Accepted {
			t.Fatalf("invalid %q accepted", sentence)
		}
	}
}

func assertTraceShape(t *testing.T, input string, result Result) {
	t.Helper()
	if len(result.Steps) == 0 {
		t.Fatalf("%q: no steps", input)
	}
	if got := result.Steps[0].Stack; !reflect.DeepEqual(got, []string{"$", "E"}) {
		t.Errorf("%q: first stack = %v", input, got)
	}
	last := result.Steps[len(result.Steps)-1].Action
	if result.Accepted && last != "Aceita" {
		t.Errorf("%q: last action = %q, want Aceita", input, last)
	}
	if !result.Accepted && !strings.HasPrefix(last, "Erro") {
		t.Errorf("%q: last action = %q, want an Erro", input, last)
	}
	for i, step := range result.Steps {
		if step.Number != i+1 {
			t.Errorf("%q: step %d numbered %d", input, i+1, step.Number)
		}
		if step.Input[len(step.Input)-1] != EndMarker {
			t.Errorf("%q: step %d input does not end with $", input, i+1)
		}
	}
}

func assertSets(t *testing.T, name string, got, want map[string][]string) {
	t.Helper()
	for nt, symbols := range want {
		if !sameElements(got[nt], symbols) {
			t.Errorf("%s(%s) = %v, want %v", name, nt, got[nt], symbols)
		}
	}
	if len(got) != len(want) {
		t.Errorf("%s has %d entries, want %d", name, len(got), len(want))
	}
}

func sameElements(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, s := range a {
		seen[s]++
	}
	for _, s := range b {
		seen[s]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}
