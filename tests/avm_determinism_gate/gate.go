package avmdeterminismgate

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ProductionDirs = []string{
	filepath.FromSlash("x/aetravm"),
	filepath.FromSlash("x/contracts"),
	filepath.FromSlash("x/vm"),
	filepath.FromSlash("x/messages"),
	filepath.FromSlash("x/messaging"),
}

type Violation struct {
	File	string
	Line	int
	Col	int
	Rule	string
	Text	string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s", filepath.ToSlash(v.File), v.Line, v.Col, v.Rule, v.Text)
}

type Gate struct {
	Root string
}

func (g Gate) ScanProduction() ([]Violation, error) {
	var all []Violation
	for _, dir := range ProductionDirs {
		violations, err := g.ScanDir(filepath.Join(g.Root, dir))
		if err != nil {
			return nil, err
		}
		all = append(all, violations...)
	}
	sortViolations(all)
	return all, nil
}

func (g Gate) ScanDir(dir string) ([]Violation, error) {
	var violations []Violation
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fileViolations, err := g.ScanFile(path)
		if err != nil {
			return err
		}
		violations = append(violations, fileViolations...)
		return nil
	})
	sortViolations(violations)
	return violations, err
}

func (g Gate) ScanFile(path string) ([]Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	scanner := fileScanner{
		root:		g.Root,
		path:		path,
		fset:		fset,
		importPaths:	make(map[string]string),
	}
	scanner.scanImports(file)
	ast.Inspect(file, scanner.inspectNode)
	sortViolations(scanner.violations)
	return scanner.violations, nil
}

func FormatViolations(violations []Violation) string {
	var b strings.Builder
	for _, violation := range violations {
		b.WriteString(violation.String())
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func AssertNoViolations(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	return errors.New(FormatViolations(violations))
}

type fileScanner struct {
	root		string
	path		string
	fset		*token.FileSet
	importPaths	map[string]string
	violations	[]Violation
	fn		*ast.FuncDecl
	mapVars		[]map[string]struct{}
}

func (s *fileScanner) scanImports(file *ast.File) {
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		name := filepath.Base(path)
		if spec.Name != nil {
			name = spec.Name.Name
		}
		s.importPaths[name] = path
		switch {
		case path == "time":
			s.add(spec.Pos(), "forbidden-import", "time package is forbidden in AVM production runtime paths")
		case path == "math/rand":
			s.add(spec.Pos(), "forbidden-import", "math/rand is forbidden in AVM production runtime paths")
		case path == "crypto/rand":
			s.add(spec.Pos(), "forbidden-import", "crypto/rand is forbidden inside AVM execution")
		case path == "os":
			s.add(spec.Pos(), "forbidden-import", "os package is forbidden: filesystem/env access must not enter AVM runtime")
		case path == "io/fs":
			s.add(spec.Pos(), "forbidden-import", "io/fs is forbidden: filesystem access must not enter AVM runtime")
		case path == "path/filepath":
			s.add(spec.Pos(), "forbidden-import", "path/filepath is forbidden: filesystem access must not enter AVM runtime")
		case path == "net" || strings.HasPrefix(path, "net/"):
			s.add(spec.Pos(), "forbidden-import", "network packages are forbidden in AVM runtime")
		}
	}
}

func (s *fileScanner) inspectNode(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.FuncDecl:
		previousFn := s.fn
		previousMapVars := s.mapVars
		s.fn = n
		s.mapVars = append(s.mapVars, collectFuncMapVars(n))
		ast.Inspect(n.Body, s.inspectNode)
		s.mapVars = previousMapVars
		s.fn = previousFn
		return false
	case *ast.GoStmt:
		s.add(n.Go, "forbidden-goroutine", "goroutines are forbidden inside AVM runtime execution paths")
	case *ast.SelectorExpr:
		s.scanSelector(n)
	case *ast.Ident:
		if n.Name == "float32" || n.Name == "float64" {
			s.add(n.Pos(), "forbidden-float", "floating point types are forbidden in AVM runtime paths")
		}
	case *ast.BasicLit:
		if n.Kind == token.FLOAT {
			s.add(n.Pos(), "forbidden-float", "floating point literals are forbidden in AVM runtime paths")
		}
	case *ast.AssignStmt:
		s.trackMapAssignments(n)
	case *ast.ValueSpec:
		s.trackMapValueSpec(n)
	case *ast.RangeStmt:
		s.scanRange(n)
	}
	return true
}

func (s *fileScanner) scanSelector(n *ast.SelectorExpr) {
	x, ok := n.X.(*ast.Ident)
	if !ok {
		return
	}
	path := s.importPaths[x.Name]
	switch path {
	case "time":
		if n.Sel.Name == "Now" {
			s.add(n.Pos(), "forbidden-time", "time.Now is forbidden in AVM runtime")
		}
	case "os":
		if isOSForbiddenSelector(n.Sel.Name) {
			s.add(n.Pos(), "forbidden-os", "filesystem/env/process access is forbidden in AVM runtime")
		}
	case "math/rand":
		s.add(n.Pos(), "forbidden-randomness", "math/rand use is forbidden in AVM runtime")
	case "crypto/rand":
		s.add(n.Pos(), "forbidden-randomness", "crypto/rand use is forbidden inside execution")
	}
	if path == "net" || strings.HasPrefix(path, "net/") {
		s.add(n.Pos(), "forbidden-network", "network access is forbidden in AVM runtime")
	}
}

func (s *fileScanner) trackMapAssignments(n *ast.AssignStmt) {
	if len(s.mapVars) == 0 {
		return
	}
	scope := s.mapVars[len(s.mapVars)-1]
	for i, lhs := range n.Lhs {
		if i >= len(n.Rhs) {
			continue
		}
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		if isMapExpression(n.Rhs[i]) {
			scope[ident.Name] = struct{}{}
		}
	}
}

func (s *fileScanner) trackMapValueSpec(n *ast.ValueSpec) {
	if len(s.mapVars) == 0 {
		return
	}
	scope := s.mapVars[len(s.mapVars)-1]
	for _, name := range n.Names {
		if _, ok := n.Type.(*ast.MapType); ok {
			scope[name.Name] = struct{}{}
		}
	}
	for i, value := range n.Values {
		if i >= len(n.Names) {
			continue
		}
		if isMapExpression(value) {
			scope[n.Names[i].Name] = struct{}{}
		}
	}
}

func (s *fileScanner) scanRange(n *ast.RangeStmt) {
	if s.fn == nil || !isDeterministicStateBuilder(s.fn.Name.Name) {
		return
	}
	if !s.isKnownMapRange(n.X) {
		return
	}
	if rangeIsSortedKeyCollection(n, s.fn.Body) {
		return
	}
	s.add(n.For, "unsorted-map-range", "state/root/export builders must collect map keys and sort before consuming map entries")
}

func (s *fileScanner) isKnownMapRange(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	for i := len(s.mapVars) - 1; i >= 0; i-- {
		if _, ok := s.mapVars[i][ident.Name]; ok {
			return true
		}
	}
	return false
}

func (s *fileScanner) add(pos token.Pos, rule, text string) {
	position := s.fset.Position(pos)
	path := position.Filename
	if s.root != "" {
		if rel, err := filepath.Rel(s.root, path); err == nil {
			path = rel
		}
	}
	s.violations = append(s.violations, Violation{
		File:	path,
		Line:	position.Line,
		Col:	position.Column,
		Rule:	rule,
		Text:	text,
	})
}

func collectFuncMapVars(fn *ast.FuncDecl) map[string]struct{} {
	out := make(map[string]struct{})
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if _, ok := field.Type.(*ast.MapType); ok {
				for _, name := range field.Names {
					out[name.Name] = struct{}{}
				}
			}
		}
	}
	return out
}

func isMapExpression(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		_, ok := e.Type.(*ast.MapType)
		return ok
	case *ast.CallExpr:
		if ident, ok := e.Fun.(*ast.Ident); ok && ident.Name == "make" && len(e.Args) > 0 {
			_, ok := e.Args[0].(*ast.MapType)
			return ok
		}
	}
	return false
}

func isDeterministicStateBuilder(name string) bool {
	if strings.HasPrefix(name, "clone") || strings.HasPrefix(name, "Clone") {
		return false
	}
	if name == "Export" || strings.HasPrefix(name, "Export") || strings.HasSuffix(name, "Export") {
		return true
	}
	if strings.HasPrefix(name, "Compute") && (strings.Contains(name, "Root") || strings.Contains(name, "Hash")) {
		return true
	}
	for _, prefix := range []string{"Normalize", "Canonical", "Encode", "Snapshot"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func rangeIsSortedKeyCollection(rangeStmt *ast.RangeStmt, body *ast.BlockStmt) bool {
	keyIdent, ok := rangeStmt.Key.(*ast.Ident)
	if !ok || keyIdent.Name == "_" {
		return false
	}
	valueIdent, hasValue := rangeStmt.Value.(*ast.Ident)
	if hasValue && valueIdent.Name != "_" {
		return false
	}
	appendTargets := map[string]struct{}{}
	ok = true
	ast.Inspect(rangeStmt.Body, func(node ast.Node) bool {
		if node == nil || !ok {
			return false
		}
		assign, isAssign := node.(*ast.AssignStmt)
		if !isAssign || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		lhs, lhsOK := assign.Lhs[0].(*ast.Ident)
		call, callOK := assign.Rhs[0].(*ast.CallExpr)
		if !lhsOK || !callOK || len(call.Args) != 2 {
			ok = false
			return false
		}
		fun, funOK := call.Fun.(*ast.Ident)
		arg0, arg0OK := call.Args[0].(*ast.Ident)
		arg1, arg1OK := call.Args[1].(*ast.Ident)
		if !funOK || fun.Name != "append" || !arg0OK || !arg1OK || lhs.Name != arg0.Name || arg1.Name != keyIdent.Name {
			ok = false
			return false
		}
		appendTargets[lhs.Name] = struct{}{}
		return true
	})
	if !ok || len(appendTargets) == 0 {
		return false
	}
	for target := range appendTargets {
		if !bodyHasSortCall(body, target, rangeStmt.End()) {
			return false
		}
	}
	return true
}

func bodyHasSortCall(body *ast.BlockStmt, target string, after token.Pos) bool {
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found || node == nil || node.Pos() <= after {
			return !found
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := selector.X.(*ast.Ident)
		if !ok || pkg.Name != "sort" || len(call.Args) == 0 {
			return true
		}
		arg, ok := call.Args[0].(*ast.Ident)
		if ok && arg.Name == target {
			found = true
		}
		return !found
	})
	return found
}

func isOSForbiddenSelector(name string) bool {
	switch name {
	case "Open", "OpenFile", "Create", "ReadFile", "WriteFile", "ReadDir", "Mkdir", "MkdirAll",
		"Remove", "RemoveAll", "Rename", "Stat", "Lstat", "Chdir", "Getwd",
		"Getenv", "LookupEnv", "Environ", "Setenv", "Unsetenv", "ExpandEnv",
		"Exec", "StartProcess", "FindProcess":
		return true
	default:
		return false
	}
}

func sortViolations(violations []Violation) {
	sort.SliceStable(violations, func(i, j int) bool {
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		if violations[i].Col != violations[j].Col {
			return violations[i].Col < violations[j].Col
		}
		if violations[i].Rule != violations[j].Rule {
			return violations[i].Rule < violations[j].Rule
		}
		return violations[i].Text < violations[j].Text
	})
}
