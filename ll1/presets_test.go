package ll1

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// generatedPerPreset é quantas sentenças geradas cada gramática pronta tem de tratar.
const generatedPerPreset = 200

// handSets guarda FIRST e FOLLOW calculados à mão: o arquivo da etapa para os
// exemplos próprios e steps/06-presets-research.md para as 15 gramáticas com fonte.
type handSets struct {
	first, follow map[string][]string
}

var presetHandSets = map[string]handSets{
	"expressoes": {
		first: map[string][]string{
			"E": {"(", "id", "num"}, "E'": {"+", "-", Epsilon}, "T": {"(", "id", "num"},
			"T'": {"*", "/", Epsilon}, "F": {"(", "id", "num"},
		},
		follow: map[string][]string{
			"E": {")", "$"}, "E'": {")", "$"}, "T": {"+", "-", ")", "$"},
			"T'": {"+", "-", ")", "$"}, "F": {"*", "/", "+", "-", ")", "$"},
		},
	},
	"booleanas": {
		first: map[string][]string{
			"B": {"not", "(", "true", "false", "id"}, "B'": {"or", Epsilon},
			"T": {"not", "(", "true", "false", "id"}, "T'": {"and", Epsilon},
			"F": {"not", "(", "true", "false", "id"},
		},
		follow: map[string][]string{
			"B": {")", "$"}, "B'": {")", "$"}, "T": {"or", ")", "$"},
			"T'": {"or", ")", "$"}, "F": {"and", "or", ")", "$"},
		},
	},
	"declaracoes": {
		first: map[string][]string{
			"D": {"int", "float", "char"}, "T": {"int", "float", "char"},
			"L": {"id"}, "L'": {",", Epsilon},
		},
		follow: map[string][]string{
			"D": {"$"}, "T": {"id"}, "L": {";"}, "L'": {";"},
		},
	},
	"listas": {
		first: map[string][]string{
			"L": {"["}, "E": {"num", "id", "[", Epsilon}, "E'": {",", Epsilon}, "X": {"num", "id", "["},
		},
		follow: map[string][]string{
			"L": {"$", ",", "]"}, "E": {"]"}, "E'": {"]"}, "X": {",", "]"},
		},
	},
	"nao-ll1": {
		first:  map[string][]string{"E": {"id"}, "T": {"id"}},
		follow: map[string][]string{"E": {"+", "$"}, "T": {"+", "$"}},
	},
	"json": {
		first: map[string][]string{
			"Value": {"{", "[", "id", "num", "true", "false", "null"}, "Object": {"{"},
			"Members": {"id", Epsilon}, "Members'": {",", Epsilon}, "Member": {"id"}, "Array": {"["},
			"Elements":  {"{", "[", "id", "num", "true", "false", "null", Epsilon},
			"Elements'": {",", Epsilon},
		},
		follow: map[string][]string{
			"Value": {"$", ",", "}", "]"}, "Object": {"$", ",", "}", "]"}, "Members": {"}"},
			"Members'": {"}"}, "Member": {",", "}"}, "Array": {"$", ",", "}", "]"},
			"Elements": {"]"}, "Elements'": {"]"},
		},
	},
	"s-expressoes": {
		first:  map[string][]string{"S": {"id", "("}, "Rest": {")", ".", ","}},
		follow: map[string][]string{"S": {"$", ")", ".", ","}, "Rest": {"$", ")", ".", ","}},
	},
	"pl0": {
		first: map[string][]string{
			"Program": {"var", "id", "call", "begin", "if", "while", "."},
			"Var":     {"var", Epsilon}, "Ids": {",", Epsilon},
			"Stmt":  {"id", "call", "begin", "if", "while", Epsilon},
			"Stmts": {";", Epsilon}, "Cond": {"odd", "id", "num", "("}, "Rel": {"=", "#", "<", ">"},
			"Expr": {"id", "num", "("}, "Expr'": {"+", "-", Epsilon}, "Term": {"id", "num", "("},
		},
		follow: map[string][]string{
			"Program": {"$"}, "Var": {"id", "call", "begin", "if", "while", "."}, "Ids": {";"},
			"Stmt": {".", ";", "end"}, "Stmts": {"end"}, "Cond": {"then", "do"}, "Rel": {"id", "num", "("},
			"Expr":  {".", ";", "end", "then", "do", "=", "#", "<", ">", ")"},
			"Expr'": {".", ";", "end", "then", "do", "=", "#", "<", ">", ")"},
			"Term":  {"+", "-", ".", ";", "end", "then", "do", "=", "#", "<", ">", ")"},
		},
	},
	"linha-reta": {
		first: map[string][]string{
			"Stm": {"id", "print"}, "Stm'": {";", Epsilon}, "Simple": {"id", "print"},
			"Exp": {"id", "num", "("}, "Exp'": {"+", "-", "*", "/", Epsilon}, "Atom": {"id", "num", "("},
			"ExpList": {"id", "num", "("}, "ExpList'": {",", Epsilon}, "Binop": {"+", "-", "*", "/"},
		},
		follow: map[string][]string{
			"Stm": {"$", ","}, "Stm'": {"$", ","}, "Simple": {";", "$", ","},
			"Exp": {";", "$", ",", ")"}, "Exp'": {";", "$", ",", ")"},
			"Atom":    {"+", "-", "*", "/", ";", "$", ",", ")"},
			"ExpList": {")"}, "ExpList'": {")"}, "Binop": {"id", "num", "("},
		},
	},
	"tiny": {
		first: map[string][]string{
			"StmtSeq": {"if", "repeat", "id", "read", "write"}, "StmtSeq'": {";", Epsilon},
			"Stmt": {"if", "repeat", "id", "read", "write"}, "Else": {"else", Epsilon},
			"Exp": {"num", "id", "("}, "Exp'": {"<", "=", Epsilon}, "Simple": {"num", "id", "("},
			"Simple'": {"+", "-", Epsilon}, "Factor": {"num", "id", "("},
		},
		follow: map[string][]string{
			"StmtSeq": {"$", "else", "end", "until"}, "StmtSeq'": {"$", "else", "end", "until"},
			"Stmt": {";", "$", "else", "end", "until"}, "Else": {"end"},
			"Exp":     {"then", ")", ";", "$", "else", "end", "until"},
			"Exp'":    {"then", ")", ";", "$", "else", "end", "until"},
			"Simple":  {"<", "=", "then", ")", ";", "$", "else", "end", "until"},
			"Simple'": {"<", "=", "then", ")", ";", "$", "else", "end", "until"},
			"Factor":  {"+", "-", "<", "=", "then", ")", ";", "$", "else", "end", "until"},
		},
	},
	"regex": {
		first: map[string][]string{
			"Regex": {"id", "(", "+", Epsilon}, "Regex'": {"+", Epsilon}, "Term": {"id", "(", Epsilon},
			"Factor": {"id", "("}, "Star": {"*", Epsilon}, "Base": {"id", "("},
		},
		follow: map[string][]string{
			"Regex": {"$", ")"}, "Regex'": {"$", ")"}, "Term": {"+", "$", ")"},
			"Factor": {"id", "(", "+", "$", ")"}, "Star": {"id", "(", "+", "$", ")"},
			"Base": {"*", "id", "(", "+", "$", ")"},
		},
	},
	"dyck": {
		first:  map[string][]string{"S": {"(", "[", Epsilon}},
		follow: map[string][]string{"S": {"$", ")", "]"}},
	},
	"xml": {
		first: map[string][]string{
			"Element": {"<"}, "Attrs": {"id", Epsilon}, "Tail": {"/", ">"},
			"Content": {"id", "<"}, "After": {"id", "/"},
		},
		follow: map[string][]string{
			"Element": {"$"}, "Attrs": {"/", ">"}, "Tail": {"$", "id", "<"},
			"Content": {"$", "id", "<"}, "After": {"$", "id", "<"},
		},
	},
	"csv": {
		first: map[string][]string{
			"File": {"id", "num", ",", ";", Epsilon}, "File'": {";", Epsilon},
			"Record": {"id", "num", ",", Epsilon}, "Record'": {",", Epsilon}, "Field": {"id", "num", Epsilon},
		},
		follow: map[string][]string{
			"File": {"$"}, "File'": {"$"}, "Record": {";", "$"}, "Record'": {";", "$"}, "Field": {",", ";", "$"},
		},
	},
	"logica-proposicional": {
		first:  map[string][]string{"F": {"id", "~", "("}, "Op": {"&", "+", ">", "="}},
		follow: map[string][]string{"F": {"$", "&", "+", ">", "=", ")"}, "Op": {"id", "~", "("}},
	},
	"lambda": {
		first: map[string][]string{
			"Term": {"id", "(", "\\"}, "Term'": {"id", "(", "\\", Epsilon}, "Atom": {"id", "(", "\\"},
		},
		follow: map[string][]string{
			"Term": {"$", ")", "]"}, "Term'": {"$", ")", "]"}, "Atom": {"id", "(", "\\", "$", ")", "]"},
		},
	},
	"lua-tabelas": {
		first: map[string][]string{
			"Table": {"{"}, "Fields": {"[", "id", "num", "{", Epsilon}, "Fields'": {",", ";", Epsilon},
			"Sep": {",", ";"}, "Field": {"[", "id", "num", "{"}, "Field'": {"=", Epsilon},
			"Exp": {"id", "num", "{"},
		},
		follow: map[string][]string{
			"Table": {"$", ",", ";", "}", "]"}, "Fields": {"}"}, "Fields'": {"}"},
			"Sep": {"[", "id", "num", "{", "}"}, "Field": {",", ";", "}"}, "Field'": {",", ";", "}"},
			"Exp": {"]", ",", ";", "}"},
		},
	},
	"wirth-ebnf": {
		first: map[string][]string{
			"Syntax": {"id", Epsilon}, "Production": {"id"}, "Expression": {"id", "[", "(", "{"},
			"Expression'": {"/", Epsilon}, "Term": {"id", "[", "(", "{"},
			"Term'": {"id", "[", "(", "{", Epsilon}, "Factor": {"id", "[", "(", "{"},
		},
		follow: map[string][]string{
			"Syntax": {"$"}, "Production": {"id", "$"}, "Expression": {".", "]", ")", "}"},
			"Expression'": {".", "]", ")", "}"}, "Term": {"/", ".", "]", ")", "}"},
			"Term'": {"/", ".", "]", ")", "}"}, "Factor": {"id", "[", "(", "{", "/", ".", "]", ")", "}"},
		},
	},
	"brainfuck": {
		first: map[string][]string{
			"Program": {"+", "-", "<", ">", ".", ",", "[", Epsilon},
			"Cmd":     {"+", "-", "<", ">", ".", ",", "["},
		},
		follow: map[string][]string{
			"Program": {"$", "]"},
			"Cmd":     {"+", "-", "<", ">", ".", ",", "[", "$", "]"},
		},
	},
	"pascal-tipos": {
		first: map[string][]string{
			"VarPart": {"var"}, "Decls": {"id", Epsilon}, "Decl": {"id"}, "Ids": {",", Epsilon},
			"Type": {"id", "num", "^", "array", "record"}, "Range": {"num"},
			"Rest": {";", Epsilon}, "Rest'": {"id", Epsilon},
		},
		follow: map[string][]string{
			"VarPart": {"$"}, "Decls": {"$"}, "Decl": {";", "end"}, "Ids": {":"},
			"Type": {";", "end"}, "Range": {";", "end", "]"}, "Rest": {"end"}, "Rest'": {"end"},
		},
	},
}

// Sentenças conferidas além dos exemplos de cada gramática, quase todas sem
// espaços, para exercitar o léxico que conhece a gramática.
var extraAccepted = map[string][]string{
	"expressoes":           {"a-b/c+d*(e-1)"},
	"booleanas":            {"not not x", "(a or b) and c"},
	"declaracoes":          {"int a,b,c;", "char x;"},
	"listas":               {"[[[]]]", "[1,[a,2],[]]"},
	"json":                 {"[]", "{a:{b:[1,2,{}]}}"},
	"s-expressoes":         {"(a.(b.c))"},
	"pl0":                  {".", "var a;a=1."},
	"linha-reta":           {"print(1)", "a=(b=1,b)"},
	"tiny":                 {"repeat read x until x<1"},
	"regex":                {"", "+", "(a)**b"},
	"dyck":                 {"", "([])[()]"},
	"xml":                  {"<a><b/></a>"},
	"csv":                  {"", ",,", "1;;2"},
	"logica-proposicional": {"~~p", "((p&q)+r)"},
	"lambda":               {"\\f[\\x[f(f x)]]"},
	"lua-tabelas":          {"{{}}", "{x;y;}"},
	"wirth-ebnf":           {"", "a=b/c.d={e}."},
	"brainfuck":            {"", "++[->+<].", "[[]]"},
	"pascal-tipos":         {"var v:array[1..10]of real;"},
}

var extraRejected = map[string][]string{
	"expressoes":           {"a**b", "()"},
	"booleanas":            {"a not b", "or"},
	"declaracoes":          {"int a,b,c", "a;", "int int;"},
	"listas":               {"[", "1", "[,]"},
	"json":                 {"{a}", "a = 1"},
	"s-expressoes":         {"(a.b.c)", ""},
	"pl0":                  {"", "x = 1"},
	"linha-reta":           {"", "a = 1 ;"},
	"tiny":                 {"", "x = 1 else"},
	"regex":                {")", "a |"},
	"dyck":                 {"(", "]"},
	"xml":                  {"", "< a > < / >"},
	"csv":                  {"a b", "\"a\""},
	"logica-proposicional": {"", "( p )"},
	"lambda":               {"", "\\x x"},
	"lua-tabelas":          {"{ = }", "{ [ k ] }"},
	"wirth-ebnf":           {"a", "a = b . ."},
	"brainfuck":            {"]", "a"},
	"pascal-tipos":         {"var x : integer", "var x : 1 . 9 ;"},
}

const naoLL1ID = "nao-ll1"

func presetByIDForTest(t *testing.T, id string) Preset {
	t.Helper()
	p, ok := PresetByID(id)
	if !ok {
		t.Fatalf("preset %q not found", id)
	}
	return p
}

func TestPresetsListTwentyGrammarsInThreeGroups(t *testing.T) {
	presets := Presets()
	if len(presets) != 20 {
		t.Fatalf("presets = %d, want 20", len(presets))
	}
	if presets[0].ID != DefaultPresetID || DefaultPresetID != "expressoes" {
		t.Errorf("first preset = %q, want expressoes", presets[0].ID)
	}
	counts := map[PresetGroup]int{}
	seen := map[string]bool{}
	for _, p := range presets {
		if seen[p.ID] {
			t.Errorf("duplicate id %q", p.ID)
		}
		seen[p.ID] = true
		counts[p.Group]++
		if p.Name == "" || p.Description == "" || p.Grammar == nil || len(p.Examples) == 0 {
			t.Errorf("%s: missing name, description, grammar or examples", p.ID)
		}
		if p.Group == GroupLiterature && (p.Author == "" || p.Source == "" || !strings.HasPrefix(p.URL, "https://") && !strings.HasPrefix(p.URL, "http://")) {
			t.Errorf("%s: sourced preset lacks author, source or URL", p.ID)
		}
		if _, ok := presetHandSets[p.ID]; !ok {
			t.Errorf("%s: no hand-computed sets in the test", p.ID)
		}
	}
	want := map[PresetGroup]int{GroupAssignment: 1, GroupExamples: 4, GroupLiterature: 15}
	if !reflect.DeepEqual(counts, want) {
		t.Errorf("group sizes = %v, want %v", counts, want)
	}
	for group, label := range map[PresetGroup]string{GroupAssignment: "Trabalho", GroupExamples: "Exemplos", GroupLiterature: "Da literatura e da web"} {
		if group.Label() != label {
			t.Errorf("label = %q, want %q", group.Label(), label)
		}
	}
}

func TestBrainfuckIsDisplayedWithAnAsterisk(t *testing.T) {
	if got := presetByIDForTest(t, "brainfuck").Name; got != "Brainf*ck" {
		t.Errorf("name = %q", got)
	}
}

func TestDefaultPresetIsTheAssignmentGrammar(t *testing.T) {
	p := presetByIDForTest(t, DefaultPresetID)
	if !reflect.DeepEqual(p.Grammar, Default()) {
		t.Errorf("default preset grammar differs from Default()")
	}
	if _, ok := PresetByID("nao-existe"); ok {
		t.Error("unknown id found")
	}
}

func TestPresetSetsMatchTheHandComputedTables(t *testing.T) {
	for _, p := range Presets() {
		want, ok := presetHandSets[p.ID]
		if !ok {
			continue
		}
		t.Run(p.ID, func(t *testing.T) {
			assertSets(t, "FIRST", p.Grammar.First(), want.first)
			assertSets(t, "FOLLOW", p.Grammar.Follow(), want.follow)
			if len(p.Grammar.NonTerminals) != len(want.first) {
				t.Errorf("non-terminals = %v", p.Grammar.NonTerminals)
			}
		})
	}
}

func TestOnlyTheNonLL1PresetHasConflicts(t *testing.T) {
	for _, p := range Presets() {
		table := BuildTable(p.Grammar)
		if p.ID == naoLL1ID {
			want := []ConflictCell{{NonTerminal: "E", Column: "id", Numbers: []int{1, 2}}}
			if !reflect.DeepEqual(table.ConflictCells, want) || len(table.Conflicts) != 1 {
				t.Errorf("nao-ll1 conflicts = %v %v, want M[E, id] with 1 and 2", table.ConflictCells, table.Conflicts)
			}
			continue
		}
		if len(table.Conflicts) != 0 || len(table.ConflictCells) != 0 {
			t.Errorf("%s: conflicts %v", p.ID, table.Conflicts)
		}
	}
}

func TestEveryTerminalCanBeTyped(t *testing.T) {
	for _, p := range Presets() {
		for _, terminal := range p.Grammar.Terminals {
			runes := []rune(terminal)
			word := isLetter(runes[0])
			for _, r := range runes {
				word = word && (isLetter(r) || isDigit(r))
			}
			single := len(runes) == 1 && !isLetter(runes[0]) && !isDigit(runes[0]) && !isSpace(runes[0])
			if terminal == "|" || terminal == EndMarker || terminal == Epsilon || (!word && !single) {
				t.Errorf("%s: terminal %q cannot be produced by the lexer", p.ID, terminal)
			}
		}
	}
}

func TestPresetExamplesGetTheExpectedVerdicts(t *testing.T) {
	for _, p := range Presets() {
		if p.ID == naoLL1ID {
			continue
		}
		table := BuildTable(p.Grammar)
		for _, input := range append(append([]string{}, p.Examples...), extraAccepted[p.ID]...) {
			if result := Parse(table, input); !result.Accepted {
				t.Errorf("%s: %q rejected: %s", p.ID, input, result.Message)
			}
		}
		for _, input := range append(append([]string{}, p.InvalidExamples...), extraRejected[p.ID]...) {
			result := Parse(table, input)
			if result.Accepted {
				t.Errorf("%s: %q accepted", p.ID, input)
			}
			if len(result.Steps) == 0 || !strings.HasPrefix(result.Steps[len(result.Steps)-1].Action, ErrorPrefix) {
				t.Errorf("%s: %q trace does not end in an error", p.ID, input)
			}
		}
	}
}

func TestParseRefusesAGrammarWithConflicts(t *testing.T) {
	p := presetByIDForTest(t, naoLL1ID)
	table := BuildTable(p.Grammar)
	for _, input := range append(append([]string{}, p.Examples...), p.InvalidExamples...) {
		result := Parse(table, input)
		if result.Accepted || !strings.Contains(result.Message, "A gramática não é LL(1): a tabela tem conflitos") {
			t.Errorf("%q: accepted=%v message=%q", input, result.Accepted, result.Message)
		}
		if len(result.Steps) != 1 || !strings.HasPrefix(result.Steps[0].Action, ErrorPrefix) {
			t.Errorf("%q: steps = %v", input, result.Steps)
		}
	}
}

func TestGeneratedSentencesAreAcceptedForEveryLL1Preset(t *testing.T) {
	for _, p := range Presets() {
		if p.ID == naoLL1ID {
			continue
		}
		table := BuildTable(p.Grammar)
		rng := rand.New(rand.NewSource(11))
		for i := 0; i < generatedPerPreset; i++ {
			sentence := Generate(p.Grammar, rng, 1+i%8)
			if result := Parse(table, sentence); !result.Accepted {
				t.Fatalf("%s: generated %q rejected: %s", p.ID, sentence, result.Message)
			}
		}
	}
}

func TestGeneratedInvalidSentencesAreRejectedForEveryLL1Preset(t *testing.T) {
	for _, p := range Presets() {
		if p.ID == naoLL1ID {
			continue
		}
		table := BuildTable(p.Grammar)
		rng := rand.New(rand.NewSource(13))
		for i := 0; i < generatedPerPreset; i++ {
			sentence := GenerateInvalid(p.Grammar, rng)
			if Parse(table, sentence).Accepted {
				t.Fatalf("%s: invalid %q accepted", p.ID, sentence)
			}
		}
	}
}

func TestGeneratedInvalidSentencesAreNeverEmpty(t *testing.T) {
	for _, p := range Presets() {
		table := BuildTable(p.Grammar)
		rng := rand.New(rand.NewSource(1))
		for i := 0; i < 500; i++ {
			sentence := GenerateInvalid(p.Grammar, rng)
			if strings.TrimSpace(sentence) == "" {
				t.Fatalf("%s: GenerateInvalid gave an empty sentence (call %d)", p.ID, i+1)
			}
			if Parse(table, sentence).Accepted {
				t.Fatalf("%s: invalid %q accepted", p.ID, sentence)
			}
		}
	}
}

func TestGenerateTerminatesForTheLeftRecursiveGrammar(t *testing.T) {
	g := presetByIDForTest(t, naoLL1ID).Grammar
	rng := rand.New(rand.NewSource(17))
	for i := 0; i < generatedPerPreset; i++ {
		tokens, err := TokenizeFor(g, Generate(g, rng, i%8))
		if err != nil {
			t.Fatal(err)
		}
		for j, token := range tokens[:len(tokens)-1] {
			want := TerminalID
			if j%2 == 1 {
				want = "+"
			}
			if token.Kind != want {
				t.Fatalf("token %d = %q, want %q (sentence of E -> E + T | T)", j, token.Kind, want)
			}
		}
	}
	if GenerateInvalid(g, rng) == "" {
		t.Error("GenerateInvalid gave an empty sentence")
	}
}

func TestEpsilonInsideALongerBodyIsDropped(t *testing.T) {
	g := grammarFrom(`
		A -> ε b A | ε ε | c`)
	want := [][]string{{"b", "A"}, {Epsilon}, {"c"}}
	for i, p := range g.Productions {
		if !reflect.DeepEqual(p.Body, want[i]) {
			t.Errorf("production %d body = %v, want %v", p.Number, p.Body, want[i])
		}
	}
	if result := Parse(BuildTable(g), "b b c"); !result.Accepted {
		t.Errorf("b b c rejected: %s", result.Message)
	}
}

func TestTokenizeForTurnsKeywordsIntoTerminals(t *testing.T) {
	g := presetByIDForTest(t, "booleanas").Grammar
	tokens, err := TokenizeFor(g, "android and not(x)")
	if err != nil {
		t.Fatal(err)
	}
	want := []Token{
		{"id", "android", 1}, {"and", "and", 9}, {"not", "not", 13}, {"(", "(", 16},
		{"id", "x", 17}, {")", ")", 18}, {EndMarker, EndMarker, 19},
	}
	if !reflect.DeepEqual(tokens, want) {
		t.Errorf("tokens = %v\nwant %v", tokens, want)
	}
}

func TestTokenizeForRejectsCharactersOutsideTheGrammar(t *testing.T) {
	g := presetByIDForTest(t, "json").Grammar
	_, err := TokenizeFor(g, "{ a = 1 }")
	if lexErr, ok := err.(*LexicalError); !ok || lexErr.Char != '=' || lexErr.Pos != 5 {
		t.Errorf("err = %v, want '=' at 5", err)
	}
	result := Parse(BuildTable(g), "[ 1 , + ]")
	if got, want := result.Steps[0].Input, []string{"[", "num", ",", "+", "$"}; !reflect.DeepEqual(got, want) {
		t.Errorf("lexical trace input = %v, want %v", got, want)
	}
}

func TestTokenizeKeepsTheAssignmentLexer(t *testing.T) {
	for _, input := range []string{"num + id", "a+b*c", "2 $ 3", "x # y", "if then"} {
		gotTokens, gotErr := Tokenize(input)
		wantTokens, wantErr := TokenizeFor(Default(), input)
		if !reflect.DeepEqual(gotTokens, wantTokens) || !reflect.DeepEqual(gotErr, wantErr) {
			t.Errorf("%q: Tokenize = %v %v, TokenizeFor(Default) = %v %v", input, gotTokens, gotErr, wantTokens, wantErr)
		}
	}
	tokens, _ := Tokenize("num")
	if tokens[0].Kind != TerminalID {
		t.Errorf("the word num must stay an identifier, got %q", tokens[0].Kind)
	}
}
