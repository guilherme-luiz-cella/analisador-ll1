// Command codegraph lê o código Go deste módulo e grava um grafo em JSON
// (pacotes, arquivos, tipos, funções, métodos e as ligações entre eles) que a
// página /graph desenha. Com -notes, lê em vez disso as notas Markdown de
// graph/ e grava as notas e os seus [[wikilinks]].
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Tipos de nó, como aparecem no JSON.
const (
	kindPackage = "package"
	kindFile    = "file"
	kindType    = "type"
	kindFunc    = "func"
	kindMethod  = "method"
)

// Tipos de ligação, como aparecem no JSON.
const (
	edgeImports  = "imports"
	edgeContains = "contains"
	edgeMethod   = "method"
	edgeCalls    = "calls"
)

const (
	maxSnippetLines = 40
	maxSignatureLen = 200
	modulePrefix    = "module "
	rootDir         = "."
	outputFileMode  = 0o644
)

// Pastas que nunca têm código do módulo.
var skippedDirs = map[string]bool{
	"node_modules": true, "vendor": true, "site": true, "dist": true, "web": true,
}

// Node é um pacote, arquivo, tipo, função ou método.
type Node struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	Package   string `json:"package"`
	File      string `json:"file,omitempty"`
	Line      int    `json:"line,omitempty"`
	Doc       string `json:"doc,omitempty"`
	Signature string `json:"signature,omitempty"`
	Code      string `json:"code,omitempty"`
}

// Edge é uma relação com sentido entre dois nós.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// Graph é o documento JSON inteiro.
type Graph struct {
	Module string `json:"module"`
	Nodes  []Node `json:"nodes"`
	Edges  []Edge `json:"edges"`
}

// parsedFile guarda o que a segunda passada precisa para resolver as chamadas.
type parsedFile struct {
	dir     string
	rel     string
	file    *ast.File
	imports map[string]string // nome local do import -> pasta no módulo ("" se for de fora)
	decls   []*ast.FuncDecl
}

type builder struct {
	module string
	fset   *token.FileSet
	nodes  map[string]*Node
	edges  map[Edge]bool

	files         []*parsedFile
	funcsByDir    map[string]map[string]string // pasta -> nome da função -> id do nó
	methodsByName map[string][]string          // nome do método -> ids dos nós
	declIDs       map[*ast.FuncDecl]string     // declaração -> id do nó
	sources       map[string][]string          // caminho do arquivo -> linhas
}

func main() {
	root := flag.String("root", ".", "module root directory")
	notes := flag.String("notes", "", "vault directory: write the note graph instead of the code graph")
	out := flag.String("o", "", "output file (default: stdout)")
	flag.Parse()

	var document any
	var err error
	if *notes != "" {
		document, err = buildNotes(*notes, *root)
	} else {
		document, err = buildGraph(*root)
	}
	if err == nil {
		err = writeJSON(document, *out)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "codegraph:", err)
		os.Exit(1)
	}
}

func writeJSON(document any, out string) error {
	data, err := json.MarshalIndent(document, "", " ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if out == "" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(out, data, outputFileMode)
}

func buildGraph(root string) (*Graph, error) {
	module, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	b := &builder{
		module:        module,
		fset:          token.NewFileSet(),
		nodes:         map[string]*Node{},
		edges:         map[Edge]bool{},
		funcsByDir:    map[string]map[string]string{},
		methodsByName: map[string][]string{},
		declIDs:       map[*ast.FuncDecl]string{},
		sources:       map[string][]string{},
	}
	if err := b.parseTree(root); err != nil {
		return nil, err
	}
	b.declare()
	b.link()
	return b.result(), nil
}

func readModulePath(goMod string) (string, error) {
	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, modulePrefix) {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, modulePrefix)), `"`), nil
		}
	}
	return "", fmt.Errorf("no module line in %s", goMod)
}

func (b *builder) parseTree(root string) error {
	return filepath.WalkDir(root, func(p string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if entry.IsDir() {
			if p != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || skippedDirs[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		file, err := parser.ParseFile(b.fset, rel, data, parser.ParseComments)
		if err != nil {
			return err
		}
		dir := path.Dir(rel)
		b.sources[rel] = strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		b.files = append(b.files, &parsedFile{dir: dir, rel: rel, file: file})
		return nil
	})
}

func packageID(dir string) string { return "pkg:" + dir }

func (b *builder) packageLabel(dir string) string {
	if dir == rootDir {
		return path.Base(b.module)
	}
	return dir
}

// uniqueID evita ids repetidos, como várias funções init num mesmo pacote.
func (b *builder) uniqueID(id string, line int) string {
	if _, taken := b.nodes[id]; taken {
		return fmt.Sprintf("%s:%d", id, line)
	}
	return id
}

func (b *builder) addEdge(from, to, kind string) {
	if from != "" && to != "" && from != to {
		b.edges[Edge{From: from, To: to, Kind: kind}] = true
	}
}

// declare cria todos os nós e as ligações de estrutura (primeira passada).
func (b *builder) declare() {
	for _, pf := range b.files {
		pkgID := packageID(pf.dir)
		pkg := b.nodes[pkgID]
		if pkg == nil {
			pkg = &Node{ID: pkgID, Label: b.packageLabel(pf.dir), Kind: kindPackage, Package: pf.dir,
				Signature: "package " + pf.file.Name.Name}
			b.nodes[pkgID] = pkg
		}
		if pkg.Doc == "" {
			pkg.Doc = firstParagraph(pf.file.Doc)
		}

		fileID := "file:" + pf.rel
		b.nodes[fileID] = &Node{ID: fileID, Label: path.Base(pf.rel), Kind: kindFile, Package: pf.dir,
			File: pf.rel, Line: 1, Signature: importSummary(pf.file)}
		b.addEdge(pkgID, fileID, edgeContains)

		pf.imports = map[string]string{}
		for _, spec := range pf.file.Imports {
			b.declareImport(pf, pkgID, spec)
		}

		for _, decl := range pf.file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				b.declareTypes(pf, fileID, d)
			case *ast.FuncDecl:
				b.declareFunc(pf, fileID, d)
			}
		}
	}
}

func importSummary(file *ast.File) string {
	var paths []string
	for _, spec := range file.Imports {
		paths = append(paths, strings.Trim(spec.Path.Value, `"`))
	}
	if len(paths) == 0 {
		return ""
	}
	return "import " + strings.Join(paths, ", ")
}

func (b *builder) declareImport(pf *parsedFile, pkgID string, spec *ast.ImportSpec) {
	importPath := strings.Trim(spec.Path.Value, `"`)
	local := path.Base(importPath)
	if spec.Name != nil {
		local = spec.Name.Name
	}
	if local == "_" || local == "." {
		return
	}
	dir := ""
	switch {
	case importPath == b.module:
		dir = rootDir
	case strings.HasPrefix(importPath, b.module+"/"):
		dir = strings.TrimPrefix(importPath, b.module+"/")
	}
	pf.imports[local] = dir
	if dir != "" {
		b.addEdge(pkgID, packageID(dir), edgeImports)
	}
}

func (b *builder) declareTypes(pf *parsedFile, fileID string, d *ast.GenDecl) {
	if d.Tok != token.TYPE {
		return
	}
	for _, spec := range d.Specs {
		ts := spec.(*ast.TypeSpec)
		line := b.fset.Position(ts.Pos()).Line
		id := b.uniqueID("type:"+pf.dir+"."+ts.Name.Name, line)
		doc := ts.Doc
		if doc == nil && len(d.Specs) == 1 {
			doc = d.Doc
		}
		end := b.fset.Position(ts.End()).Line
		b.nodes[id] = &Node{ID: id, Label: ts.Name.Name, Kind: kindType, Package: pf.dir, File: pf.rel,
			Line: line, Doc: firstParagraph(doc), Signature: typeSignature(ts), Code: b.snippet(pf.rel, line, end)}
		b.addEdge(fileID, id, edgeContains)
	}
}

func typeSignature(ts *ast.TypeSpec) string {
	switch t := ts.Type.(type) {
	case *ast.StructType:
		return "type " + ts.Name.Name + " struct"
	case *ast.InterfaceType:
		return "type " + ts.Name.Name + " interface"
	default:
		return "type " + ts.Name.Name + " " + printNode(t)
	}
}

func (b *builder) declareFunc(pf *parsedFile, fileID string, d *ast.FuncDecl) {
	line := b.fset.Position(d.Pos()).Line
	end := b.fset.Position(d.End()).Line
	kind, label, prefix := kindFunc, d.Name.Name, "func:"+pf.dir+"."
	recv := receiverName(d)
	if recv != "" {
		kind, label, prefix = kindMethod, recv+"."+d.Name.Name, "func:"+pf.dir+"."+recv+"."
	}
	id := b.uniqueID(prefix+d.Name.Name, line)
	b.nodes[id] = &Node{ID: id, Label: label, Kind: kind, Package: pf.dir, File: pf.rel, Line: line,
		Doc: firstParagraph(d.Doc), Signature: signature(d), Code: b.snippet(pf.rel, line, end)}
	b.declIDs[d] = id
	b.addEdge(fileID, id, edgeContains)
	pf.decls = append(pf.decls, d)

	if recv == "" {
		if b.funcsByDir[pf.dir] == nil {
			b.funcsByDir[pf.dir] = map[string]string{}
		}
		b.funcsByDir[pf.dir][d.Name.Name] = id
		return
	}
	b.methodsByName[d.Name.Name] = append(b.methodsByName[d.Name.Name], id)
	b.addEdge("type:"+pf.dir+"."+recv, id, edgeMethod)
}

// receiverName devolve o nome do tipo receptor, ou "" numa função comum.
func receiverName(d *ast.FuncDecl) string {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return ""
	}
	expr := d.Recv.List[0].Type
	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.IndexExpr:
			expr = t.X
		case *ast.IndexListExpr:
			expr = t.X
		case *ast.ParenExpr:
			expr = t.X
		case *ast.Ident:
			return t.Name
		default:
			return ""
		}
	}
}

// signature imprime a declaração sem o corpo e sem o comentário.
func signature(d *ast.FuncDecl) string {
	header := *d
	header.Body = nil
	header.Doc = nil
	sig := strings.Join(strings.Fields(printNode(&header)), " ")
	// Corta por runa, não por byte, para não partir um "ε" ou "→" ao meio.
	if runes := []rune(sig); len(runes) > maxSignatureLen {
		sig = string(runes[:maxSignatureLen]) + "…"
	}
	return sig
}

func printNode(node any) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, token.NewFileSet(), node); err != nil {
		return ""
	}
	return buf.String()
}

func (b *builder) snippet(file string, start, end int) string {
	lines := b.sources[file]
	if start < 1 || start > len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}
	total := end - start + 1
	if total > maxSnippetLines {
		cut := lines[start-1 : start-1+maxSnippetLines]
		return strings.Join(cut, "\n") + fmt.Sprintf("\n// … (+%d linhas)", total-maxSnippetLines)
	}
	return strings.Join(lines[start-1:end], "\n")
}

func firstParagraph(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	// A primeira linha de um comentário costuma parar no meio da frase, então fica o primeiro parágrafo inteiro.
	text := strings.TrimSpace(doc.Text())
	if i := strings.Index(text, "\n\n"); i >= 0 {
		text = text[:i]
	}
	return strings.Join(strings.Fields(text), " ")
}

// link resolve as chamadas entre as funções declaradas (segunda passada).
func (b *builder) link() {
	for _, pf := range b.files {
		for _, d := range pf.decls {
			if d.Body != nil {
				b.collectCalls(pf, b.declIDs[d], d.Body)
			}
		}
	}
}

func (b *builder) collectCalls(pf *parsedFile, caller string, body *ast.BlockStmt) {
	// Nomes depois de ponto e chaves de struct parecem identificadores, mas não são referências.
	skipIdents := map[*ast.Ident]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			skipIdents[x.Sel] = true
		case *ast.KeyValueExpr:
			if key, ok := x.Key.(*ast.Ident); ok {
				skipIdents[key] = true
			}
		}
		return true
	})

	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
				b.addEdge(caller, b.resolveSelector(pf, sel), edgeCalls)
			}
		case *ast.Ident:
			// Pega chamadas diretas e funções passadas como valor; Obj existe para nomes locais.
			if !skipIdents[x] && (x.Obj == nil || x.Obj.Kind == ast.Fun) {
				b.addEdge(caller, b.funcsByDir[pf.dir][x.Name], edgeCalls)
			}
		}
		return true
	})
}

func (b *builder) resolveSelector(pf *parsedFile, sel *ast.SelectorExpr) string {
	if x, ok := sel.X.(*ast.Ident); ok && x.Obj == nil {
		if dir, imported := pf.imports[x.Name]; imported {
			if dir == "" {
				return ""
			}
			return b.funcsByDir[dir][sel.Sel.Name]
		}
	}
	if ids := b.methodsByName[sel.Sel.Name]; len(ids) == 1 {
		return ids[0]
	}
	return ""
}

func (b *builder) result() *Graph {
	graph := &Graph{Module: b.module, Nodes: []Node{}, Edges: []Edge{}}
	for _, n := range b.nodes {
		graph.Nodes = append(graph.Nodes, *n)
	}
	for e := range b.edges {
		if b.nodes[e.From] != nil && b.nodes[e.To] != nil {
			graph.Edges = append(graph.Edges, e)
		}
	}
	sort.Slice(graph.Nodes, func(i, j int) bool { return graph.Nodes[i].ID < graph.Nodes[j].ID })
	sort.Slice(graph.Edges, func(i, j int) bool {
		a, c := graph.Edges[i], graph.Edges[j]
		if a.From != c.From {
			return a.From < c.From
		}
		if a.To != c.To {
			return a.To < c.To
		}
		return a.Kind < c.Kind
	})
	return graph
}
