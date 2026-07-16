// Command gostructor-gen generates a reflection-free Fill method for a config
// struct (Theme 7). Point it at one or more struct types and it emits a
// <type>.gs.go next to the source with a
//
//	func (c *T) Fill(opts ...gostructor.Option) error
//
// that resolves every field through the same sources, in the same slice order,
// with the same conversion and error taxonomy as gostructor.Configure — but with
// no per-fill reflection: fields are known at build time and each converts into
// its concrete Go type with a direct call into gostructor/gen. gostructor.Configure
// detects the generated Filler and dispatches to it automatically (EngineAdaptive).
//
// Two ways to select what to generate:
//
//	//go:generate go run github.com/goreflect/gostructor/cmd/gostructor-gen -type Config
//
// or mark the struct with a doc comment and let the tool discover it — then you
// only ever hand gostructor-gen a path:
//
//	//gostructor:gen
//	type Config struct { ... }
//
//	go run github.com/goreflect/gostructor/cmd/gostructor-gen -dir . -recursive
//
// Flags:
//
//	-type       comma-separated struct type names to generate a Fill for
//	-dir        directory holding the package (project root with -recursive; default ".")
//	-recursive  discover //gostructor:gen-marked structs in -dir and every subdirectory
//
// The output file is named after the struct (lower-cased) with a short .gs.go
// postfix — Config -> config.gs.go, plus config.gs_test.go for the golden
// equivalence test.
//
// The generator is deliberately stdlib-only (go/ast) and classifies field types
// syntactically. The string/bool/sized-numeric primitives, time.Duration, and
// []string get a dedicated reflection-free conversion; other slices, arrays,
// maps, and named types are filled through gen.Reflective (the same reflective
// core the engine uses, so parity holds); and an untagged nested struct is
// flattened into its leaves exactly as the reflective engine flattens it. A
// field of any other shape (a pointer, a channel, an out-of-package named type)
// fails generation rather than emitting a partial Fill — generation is total, so
// a build either gets a fully-correct fast path or a clear error telling it to
// stay on the reflective engine for that type.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/goreflect/gostructor"
	"github.com/goreflect/gostructor/internal/structplan"
)

// genMarker is the doc-comment directive that flags a struct for generation in
// discovery mode (no -type). It is matched anywhere in the type's doc comment.
const genMarker = "gostructor:gen"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "gostructor-gen: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		typeList  = flag.String("type", "", "comma-separated struct type names to generate a Fill for")
		dir       = flag.String("dir", ".", "directory holding the package (project root with -recursive)")
		recursive = flag.Bool("recursive", false, "discover //gostructor:gen-marked structs in dir and every subdirectory")
	)
	flag.Parse()

	types := splitList(*typeList)

	if *recursive {
		if len(types) > 0 {
			return fmt.Errorf("-type and -recursive are mutually exclusive: -recursive discovers types from //%s markers", genMarker)
		}
		return generateTree(*dir)
	}
	return generateDir(*dir, types)
}

func splitList(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// generateTree walks root and every subdirectory, generating a Fill for each
// struct marked //gostructor:gen. Directories with no marked struct are skipped
// silently, so a single invocation can cover a whole project.
func generateTree(root string) error {
	generated := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root {
			if base := d.Name(); strings.HasPrefix(base, ".") || base == "vendor" || base == "testdata" {
				return filepath.SkipDir
			}
		}
		n, err := generateMarked(path)
		if err != nil {
			return err
		}
		generated += n
		return nil
	})
	if err != nil {
		return err
	}
	if generated == 0 {
		return fmt.Errorf("no structs marked //%s found under %s", genMarker, root)
	}
	return nil
}

// generateMarked generates a Fill for every marked struct in dir, returning how
// many it wrote. A directory that holds no Go package, or no marked struct, is a
// no-op (returns 0) so the tree walk can pass over it.
func generateMarked(dir string) (int, error) {
	pkg, err := loadPackage(dir)
	if err != nil {
		if _, ok := err.(noPackageError); ok {
			return 0, nil
		}
		return 0, err
	}
	types := markedStructs(pkg)
	for _, t := range types {
		if err := generateType(dir, pkg, t); err != nil {
			return 0, err
		}
	}
	return len(types), nil
}

// generateDir generates a Fill for the named types in dir, or — when no types
// are named — for every struct marked //gostructor:gen there.
func generateDir(dir string, types []string) error {
	pkg, err := loadPackage(dir)
	if err != nil {
		return err
	}
	if len(types) == 0 {
		types = markedStructs(pkg)
		if len(types) == 0 {
			return fmt.Errorf("no -type given and no structs marked //%s found in %s", genMarker, dir)
		}
	}
	for _, t := range types {
		if err := generateType(dir, pkg, t); err != nil {
			return err
		}
	}
	return nil
}

// generateType analyzes one struct and writes its Fill file and golden test.
func generateType(dir string, pkg *loadedPackage, typeName string) error {
	spec, err := analyzeStruct(pkg, typeName)
	if err != nil {
		return err
	}

	base := strings.ToLower(typeName)
	fillName := base + ".gs.go"
	testName := base + ".gs_test.go"

	fillSrc, err := renderFill(pkg.name, spec)
	if err != nil {
		return fmt.Errorf("formatting %s: %w", fillName, err)
	}
	if err := os.WriteFile(filepath.Join(dir, fillName), fillSrc, 0o644); err != nil {
		return err
	}

	testSrc, err := renderGoldenTest(pkg.name, spec)
	if err != nil {
		return fmt.Errorf("formatting %s: %w", testName, err)
	}
	if err := os.WriteFile(filepath.Join(dir, testName), testSrc, 0o644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "gostructor-gen: wrote %s and %s\n", fillName, testName)
	return nil
}

// noPackageError marks a directory that holds no Go package, so a tree walk can
// distinguish "nothing here" from a real parse failure.
type noPackageError struct{ dir string }

func (e noPackageError) Error() string { return "no Go package found in " + e.dir }

// loadedPackage is the parsed, non-test package in a directory.
type loadedPackage struct {
	name  string
	files []*ast.File
	fset  *token.FileSet
	// typeNames is every named type declared in the package, so a field type
	// can be recognized as an in-package name (fillable via gen.Reflective)
	// rather than an unresolvable out-of-package reference.
	typeNames map[string]bool
}

func loadPackage(dir string) (*loadedPackage, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing package in %s: %w", dir, err)
	}
	// Prefer the package go generate named; otherwise the sole package present.
	want := os.Getenv("GOPACKAGE")
	for name, pkg := range pkgs {
		if want != "" && name != want {
			continue
		}
		files := make([]*ast.File, 0, len(pkg.Files))
		for _, f := range pkg.Files {
			files = append(files, f)
		}
		lp := &loadedPackage{name: name, files: files, fset: fset, typeNames: map[string]bool{}}
		for _, f := range files {
			collectTypeNames(f, lp.typeNames)
		}
		return lp, nil
	}
	if want != "" {
		return nil, fmt.Errorf("package %q not found in %s", want, dir)
	}
	return nil, noPackageError{dir: dir}
}

func collectTypeNames(file *ast.File, into map[string]bool) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				into[ts.Name.Name] = true
			}
		}
	}
}

// markedStructs returns, in declaration order, every struct type in the package
// whose doc comment carries the //gostructor:gen marker.
func markedStructs(pkg *loadedPackage) []string {
	var out []string
	for _, file := range pkg.files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
					continue
				}
				// The marker may sit on the whole `type (...)` block (gd.Doc) or
				// on the individual spec (ts.Doc), or trail it (ts.Comment).
				if hasMarker(gd.Doc) || hasMarker(ts.Doc) || hasMarker(ts.Comment) {
					out = append(out, ts.Name.Name)
				}
			}
		}
	}
	return out
}

func hasMarker(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, genMarker) {
			return true
		}
	}
	return false
}

func findStruct(pkg *loadedPackage, typeName string) (*ast.StructType, error) {
	if st, ok := lookupStruct(pkg, typeName); ok {
		return st, nil
	}
	if pkg.typeNames[typeName] {
		return nil, fmt.Errorf("type %s is not a struct", typeName)
	}
	return nil, fmt.Errorf("struct type %s not found", typeName)
}

// lookupStruct returns the struct type declared under typeName in the package,
// if any (false for a non-struct type or an unknown name).
func lookupStruct(pkg *loadedPackage, typeName string) (*ast.StructType, bool) {
	for _, file := range pkg.files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != typeName {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				return st, ok
			}
		}
	}
	return nil, false
}

// fieldSpec is one resolvable leaf field the generator will emit code for.
type fieldSpec struct {
	Name       string // leaf struct field name (matches the runtime FieldContext.Name)
	Path       string // selector path from the root struct, e.g. "Service.Name"
	Index      int    // position in the flattened Fields slice (== reflective order)
	GoType     string // Go type spelling for the value assertion, e.g. "int", "map[string]int"
	Conv       string // reflection-free conversion func (e.g. "gen.Int"); empty when Reflective
	Reflective bool   // true when the field is filled via gen.Reflective[GoType]
	EnvKey     string // env variable name for the golden test ("" if not env-addressable)
	EnvValue   string // a type-appropriate value for the golden test
}

// method is the unique helper-method name for this field's parsing block, so
// Fill delegates to one small method per field instead of one monolith.
func (f fieldSpec) method() string {
	return "gostructorFill" + strings.ReplaceAll(f.Path, ".", "")
}

// structSpec is a struct type and its flattened, resolvable fields.
type structSpec struct {
	Type     string
	Fields   []fieldSpec
	needTime bool
}

func analyzeStruct(pkg *loadedPackage, typeName string) (*structSpec, error) {
	st, err := findStruct(pkg, typeName)
	if err != nil {
		return nil, err
	}
	spec := &structSpec{Type: typeName}
	if err := walkStruct(pkg, st, nil, typeName, spec, map[string]bool{typeName: true}); err != nil {
		return nil, err
	}
	if len(spec.Fields) == 0 {
		return nil, fmt.Errorf("%s: no exported, resolvable fields to generate a Fill for", typeName)
	}
	for i := range spec.Fields {
		spec.Fields[i].Index = i
		if strings.Contains(spec.Fields[i].GoType, "time.") {
			spec.needTime = true
		}
	}
	return spec, nil
}

// walkStruct flattens st into spec.Fields in the same order the reflective
// engine (internal/structplan) does: fields in declaration order, recursing into
// an untagged nested struct and appending every other exported field as a leaf.
// pathPrefix is the selector path to st from the root; owner names the root type
// for error messages; seen guards against a recursive struct type.
func walkStruct(pkg *loadedPackage, st *ast.StructType, pathPrefix []string, owner string, spec *structSpec, seen map[string]bool) error {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			return fmt.Errorf("%s: embedded fields are not supported by gostructor-gen (use the reflective engine for this type)", owner)
		}
		nestedName, isNested := nestedStruct(pkg, field.Type)
		tagged := tagHasConfig(field.Tag)
		for _, name := range field.Names {
			if !name.IsExported() {
				continue // unexported: the engine skips it, so do we
			}
			path := append(append([]string{}, pathPrefix...), name.Name)

			// An untagged nested struct is flattened into its leaves, exactly as
			// the reflective engine treats a non-atomic struct field.
			if isNested && !tagged {
				if seen[nestedName] {
					return fmt.Errorf("%s.%s: recursive struct type %s is not supported by gostructor-gen", owner, name.Name, nestedName)
				}
				nst, _ := lookupStruct(pkg, nestedName)
				seen[nestedName] = true
				if err := walkStruct(pkg, nst, path, owner, spec, seen); err != nil {
					return err
				}
				delete(seen, nestedName)
				continue
			}

			goType, conv, reflective, err := classifyLeaf(pkg, field.Type)
			if err != nil {
				return fmt.Errorf("%s.%s: %w (gostructor-gen supports the string/bool/numeric primitives, time.Duration, []string, and — via reflection — other slices, arrays, maps, and in-package named types; nested untagged structs are flattened)", owner, name.Name, err)
			}
			fs := fieldSpec{
				Name:       name.Name,
				Path:       strings.Join(path, "."),
				GoType:     goType,
				Conv:       conv,
				Reflective: reflective,
			}
			fillTagInfo(&fs, field.Tag, goType)
			spec.Fields = append(spec.Fields, fs)
		}
	}
	return nil
}

// nestedStruct reports whether expr names an in-package struct type with at
// least one exported field — the shape the engine flattens. A struct with no
// exported fields is atomic to the engine, so it is not reported here (it is
// then handled as a Reflective leaf).
func nestedStruct(pkg *loadedPackage, expr ast.Expr) (name string, ok bool) {
	ident, isIdent := expr.(*ast.Ident)
	if !isIdent {
		return "", false
	}
	st, isStruct := lookupStruct(pkg, ident.Name)
	if !isStruct || !hasExportedField(st) {
		return "", false
	}
	return ident.Name, true
}

func hasExportedField(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			return true // embedded exported type
		}
		for _, name := range field.Names {
			if name.IsExported() {
				return true
			}
		}
	}
	return false
}

// tagHasConfig reports whether a struct tag carries a cfg or gos tag, which
// makes a struct-typed field atomic to the engine (resolved as one leaf value
// rather than flattened).
func tagHasConfig(tag *ast.BasicLit) bool {
	if tag == nil {
		return false
	}
	raw, err := stringLit(tag.Value)
	if err != nil {
		return false
	}
	stag := reflect.StructTag(raw)
	if _, ok := stag.Lookup(structplan.CfgTag); ok {
		return true
	}
	_, ok := stag.Lookup(structplan.GosTag)
	return ok
}

// fillTagInfo derives the env key and a type-appropriate golden-test value from
// the field's cfg/gos tags, using the exact same parsing and naming as the
// engine.
func fillTagInfo(fs *fieldSpec, tag *ast.BasicLit, goType string) {
	var raw string
	if tag != nil {
		// tag.Value includes the surrounding back-quotes / quotes.
		if unq, err := stringLit(tag.Value); err == nil {
			raw = unq
		}
	}
	stag := reflect.StructTag(raw)
	base, overrides := structplan.ParseCfg(stag.Get(structplan.CfgTag))
	if key, ok := overrides[gostructor.SourceEnv]; ok {
		fs.EnvKey = key
	} else if base != "" {
		fs.EnvKey = gostructor.ScreamingSnake(base)
	}
	fs.EnvValue = sampleValue(goType)
}

// sampleValue returns a value that converts cleanly into goType, for the
// generated golden test's env fixtures. It is only consulted for the scalar
// leaf types (env is a flat string source, so a map/slice field is never given
// an env fixture — see renderGoldenTest).
func sampleValue(goType string) string {
	switch goType {
	case "string":
		return "example"
	case "bool":
		return "true"
	case "float32", "float64":
		return "3.5"
	case "time.Duration":
		return "1s"
	case "[]string":
		return "a,b,c"
	default: // the int/uint family
		return "42"
	}
}

// classifyLeaf maps a leaf field's AST type to its Go spelling, the conversion
// to use, and whether that conversion is the generic reflective one. ok is
// false for any type gostructor-gen cannot fill.
//
//   - the string/bool/sized-numeric primitives and time.Duration and []string
//     get a dedicated reflection-free gen.* helper (reflective == false);
//   - any other self-contained slice, array, map, or in-package named type is
//     filled with gen.Reflective[T] (reflective == true), which routes through
//     the same reflective core the engine uses so results match exactly.
func classifyLeaf(pkg *loadedPackage, expr ast.Expr) (goType, conv string, reflective bool, err error) {
	switch t := expr.(type) {
	case *ast.Ident:
		if c, found := primitiveConv[t.Name]; found {
			return t.Name, c, false, nil
		}
		if pkg.typeNames[t.Name] {
			return t.Name, "", true, nil
		}
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok && x.Name == "time" && t.Sel.Name == "Duration" {
			return "time.Duration", "gen.Duration", false, nil
		}
	case *ast.ArrayType:
		if t.Len == nil {
			if elem, ok := t.Elt.(*ast.Ident); ok && elem.Name == "string" {
				return "[]string", "gen.StringSlice", false, nil
			}
		}
		if selfContained(t) {
			return exprString(pkg.fset, expr), "", true, nil
		}
	case *ast.MapType:
		if selfContained(t) {
			return exprString(pkg.fset, expr), "", true, nil
		}
	}
	return "", "", false, fmt.Errorf("unsupported field type %s", exprString(pkg.fset, expr))
}

// selfContained reports whether a compound type spelling refers only to types
// the generated file can name without importing anything beyond time: builtins,
// in-package names (no package qualifier), time.Duration, and slices/arrays/maps
// thereof. A pointer, channel, func, out-of-package selector, or anonymous
// struct is rejected so generation stays total.
func selfContained(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		// A builtin or an in-package name; either way, no import is needed.
		return true
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)
		return ok && x.Name == "time" && t.Sel.Name == "Duration"
	case *ast.ArrayType:
		return selfContained(t.Elt)
	case *ast.MapType:
		return selfContained(t.Key) && selfContained(t.Value)
	default:
		return false
	}
}

var primitiveConv = map[string]string{
	"string":  "gen.String",
	"bool":    "gen.Bool",
	"int":     "gen.Int",
	"int8":    "gen.Int8",
	"int16":   "gen.Int16",
	"int32":   "gen.Int32",
	"int64":   "gen.Int64",
	"uint":    "gen.Uint",
	"uint8":   "gen.Uint8",
	"uint16":  "gen.Uint16",
	"uint32":  "gen.Uint32",
	"uint64":  "gen.Uint64",
	"float32": "gen.Float32",
	"float64": "gen.Float64",
}

// stringLit unquotes a Go string/back-quote literal from the AST.
func stringLit(lit string) (string, error) {
	if len(lit) < 2 {
		return "", fmt.Errorf("bad literal %q", lit)
	}
	// Strip the surrounding quote or back-quote; struct tag literals are simple.
	q := lit[0]
	if (q == '`' && lit[len(lit)-1] == '`') || (q == '"' && lit[len(lit)-1] == '"') {
		inner := lit[1 : len(lit)-1]
		if q == '"' {
			return strings.NewReplacer(`\"`, `"`, `\\`, `\`).Replace(inner), nil
		}
		return inner, nil
	}
	return "", fmt.Errorf("bad literal %q", lit)
}

// exprString renders a type expression back to source for error messages and
// type spellings, using the package's own file set so compound types print
// correctly.
func exprString(fset *token.FileSet, expr ast.Expr) string {
	var b bytes.Buffer
	if err := format.Node(&b, fset, expr); err != nil {
		return "?"
	}
	return b.String()
}
