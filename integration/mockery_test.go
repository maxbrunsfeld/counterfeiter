package integration_test

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"golang.org/x/tools/go/packages"
)

// mockeryConfig is the configuration the mockery template's README
// recommends; it lays fakes out the way counterfeiter does by default.
const mockeryConfig = `template: "file://%s"
dir: "{{.InterfaceDir}}/{{.SrcPackageName}}fakes"
filename: "fake_{{.InterfaceName | snakecase}}.go"
pkgname: "{{.SrcPackageName}}fakes"
structname: "Fake{{.InterfaceName | firstUpper}}"
`

// mockerySamePackageConfig is the README's override for writing fakes into the
// interface's own package. Like counterfeiter, it keeps the fake of an
// unexported interface unexported.
const mockerySamePackageConfig = `{dir: "{{.InterfaceDir}}", pkgname: "{{.SrcPackageName}}", structname: "{{if firstIsLower .InterfaceName}}fake{{else}}Fake{{end}}{{.InterfaceName | firstUpper}}"}`

func testMockeryTemplate(t *testing.T, when spec.G, it spec.S) {
	g := NewWithT(t)
	var (
		mockery string
		baseDir string
	)

	it.Before(func() {
		var err error
		mockery, err = exec.LookPath("mockery")
		if err != nil {
			if os.Getenv("CI") != "" {
				t.Fatal("mockery must be on PATH in CI")
			}
			t.Skip("mockery is not on PATH")
		}
		baseDir = t.TempDir()
		// Copy skips the *fakes directories, so every fake compared below
		// comes from mockery.
		g.Expect(Copy(filepath.Join("..", "fixtures"), baseDir)).To(Succeed())
		g.Expect(os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures\n\ngo 1.18\n"), 0600)).To(Succeed())
	})

	// generate runs mockery with the counterfeiter template over the fixture
	// package in pkgDir. pkgConfig, if set, is the package's config and
	// interfaces maps each interface to its own; both are YAML flow mappings.
	generate := func(pkgDir string, pkgConfig string, interfaces map[string]string) {
		template, err := filepath.Abs(filepath.Join("..", "mockery", "counterfeiter.templ"))
		g.Expect(err).NotTo(HaveOccurred())
		config := &strings.Builder{}
		fmt.Fprintf(config, mockeryConfig, filepath.ToSlash(template))
		pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures"
		if pkgDir != "" {
			pkgPath += "/" + pkgDir
		}
		fmt.Fprintf(config, "packages:\n  %s:\n", pkgPath)
		if pkgConfig != "" {
			fmt.Fprintf(config, "    config: %s\n", pkgConfig)
		}
		config.WriteString("    interfaces:\n")
		names := make([]string, 0, len(interfaces))
		for name := range interfaces {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintf(config, "      %s: %s\n", name, interfaces[name])
		}
		g.Expect(os.WriteFile(filepath.Join(baseDir, ".mockery.yml"), []byte(config.String()), 0600)).To(Succeed())

		cmd := exec.Command(mockery)
		cmd.Dir = baseDir
		out, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred(), string(out))
	}

	// expectCounterfeiterAPI asserts that every type counterfeiter declares in
	// fixtures/fakesDir/files has the same API in the fakes mockery wrote, and
	// that the fixture package and its tests build against them.
	expectCounterfeiterAPI := func(pkgDir string, fakesDir string, files ...string) {
		if len(files) > 0 {
			expected, err := fakeAPI(filepath.Join("..", "fixtures", fakesDir), files)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(expected).NotTo(BeEmpty())
			actual, err := fakeAPI(filepath.Join(baseDir, fakesDir), files)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(actual).To(Equal(expected))
		}

		cmd := exec.Command("go", "vet", "./"+filepath.ToSlash(filepath.Join(pkgDir, "...")))
		cmd.Dir = baseDir
		out, err := cmd.CombinedOutput()
		g.Expect(err).NotTo(HaveOccurred(), string(out))
	}

	when("faking interfaces from another package", func() {
		it("produces fakes with the same API as counterfeiter", func() {
			interfaces := map[string]string{
				"AliasedInterface":              "fake_aliased_interface.go",
				"DotImports":                    "fake_dot_imports.go",
				"EmbedsInterfaces":              "fake_embeds_interfaces.go",
				"FirstInterface":                "fake_first_interface.go",
				"GenericImportedConstraint":     "fake_generic_imported_constraint.go",
				"HasImports":                    "fake_has_imports.go",
				"HasOtherTypes":                 "fake_has_other_types.go",
				"HasVarArgs":                    "fake_has_var_args.go",
				"HasVarArgsWithLocalTypes":      "fake_has_var_args_with_local_types.go",
				"ImportsGoHyphenPackage":        "fake_imports_go_hyphen_package.go",
				"InlineStructParams":            "fake_inline_struct_params.go",
				"ReusesArgTypes":                "fake_reuses_arg_types.go",
				"SecondInterface":               "fake_second_interface.go",
				"Something":                     "fake_something.go",
				"SomethingElse":                 "fake_something_else.go",
				"SomethingWithForeignInterface": "fake_something_with_foreign_interface.go",
				"unexportedInterface":           "fake_unexported_interface.go",
			}
			configs := map[string]string{}
			var files []string
			for name, file := range interfaces {
				configs[name] = "{}"
				files = append(files, file)
			}
			generate("", "", configs)
			expectCounterfeiterAPI("fixturesfakes", "fixturesfakes", files...)
		})
	})

	when("faking generic interfaces", func() {
		it("produces fakes with the same API as counterfeiter", func() {
			generate("genericinterface", "", map[string]string{
				"GenericInterface":                      "{}",
				"GenericInterfaceAny":                   "{}",
				"GenericInterfaceComparable":            "{}",
				"GenericInterfaceCustomTypeConstraintT": "{}",
				"GenericInterfaceCustomTypeConstraintU": "{}",
				"GenericInterfaceMultipleTypes":         "{}",
			})
			expectCounterfeiterAPI("genericinterface", filepath.Join("genericinterface", "genericinterfacefakes"),
				"fake_generic_interface.go",
				"fake_generic_interface_any.go",
				"fake_generic_interface_comparable.go",
				"fake_generic_interface_custom_type_constraint_t.go",
				"fake_generic_interface_custom_type_constraint_u.go",
				"fake_generic_interface_multiple_types.go",
			)
		})
	})

	when("writing fakes into the package that declares the interface", func() {
		it("produces fakes with the same API as counterfeiter, without importing that package", func() {
			for _, file := range []string{"fake_widget.go", "fake_gadget.go"} {
				g.Expect(os.Remove(filepath.Join(baseDir, "samepackage", file))).To(Succeed())
			}
			generate("samepackage", mockerySamePackageConfig, map[string]string{
				"Widget": "{}",
				"gadget": "{}",
			})
			expectCounterfeiterAPI("samepackage", "samepackage", "fake_widget.go", "fake_gadget.go")
		})
	})

	when("a boilerplate-file is given", func() {
		it("writes it at the top of the fake like counterfeiter's -header", func() {
			generate("headers/defaultheader", "{template-data: {boilerplate-file: headers/default.header.go.txt}}", map[string]string{
				"HeaderDefault":  "{}",
				"HeaderSpecific": "{config: {template-data: {boilerplate-file: headers/specific.header.go.txt}}}",
			})
			fakesDir := filepath.Join("headers", "defaultheader", "defaultheaderfakes")
			files := []string{"fake_header_default.go", "fake_header_specific.go"}
			expectCounterfeiterAPI("headers/defaultheader", fakesDir, files...)
			// Everything before the package clause: the header, then the
			// generated-code comment.
			preamble := func(b []byte) string {
				return string(b[:strings.Index(string(b), "package ")])
			}
			for _, file := range files {
				expected, err := os.ReadFile(filepath.Join("..", "fixtures", fakesDir, file))
				g.Expect(err).NotTo(HaveOccurred())
				actual, err := os.ReadFile(filepath.Join(baseDir, fakesDir, file))
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(preamble(actual)).To(Equal(preamble(expected)), file)
			}
		})
	})

	when("the interface names its parameters", func() {
		it("uses mockery's parameter names in the fake's methods", func() {
			generate("", "", map[string]string{"InlineStructParams": "{}"})
			b, err := os.ReadFile(filepath.Join(baseDir, "fixturesfakes", "fake_inline_struct_params.go"))
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(string(b)).To(ContainSubstring(") DoSomething(ctx context.Context, body struct {"))
			g.Expect(string(b)).NotTo(ContainSubstring("arg1"))
			expectCounterfeiterAPI("fixturesfakes", "fixturesfakes", "fake_inline_struct_params.go")
		})
	})

	when("the fakes are used in a test", func() {
		it("behaves like counterfeiter's fakes", func() {
			generate("", "", map[string]string{"Something": "{}", "HasVarArgs": "{}"})
			g.Expect(os.WriteFile(filepath.Join(baseDir, "fixturesfakes", "behaviour_test.go"), []byte(mockeryBehaviourTest), 0600)).To(Succeed())
			cmd := exec.Command("go", "test", "./fixturesfakes/")
			cmd.Dir = baseDir
			out, err := cmd.CombinedOutput()
			g.Expect(err).NotTo(HaveOccurred(), string(out))
		})
	})

	when("the interface's signatures use another package named sync", func() {
		it("still guards the fake with the standard library's sync", func() {
			usesSync := filepath.Join(baseDir, "usessync")
			g.Expect(os.MkdirAll(usesSync, 0700)).To(Succeed())
			g.Expect(os.WriteFile(filepath.Join(usesSync, "usessync.go"), []byte("package usessync\n\nimport \"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/sync\"\n\ntype UsesSync interface {\n\tDo(sync.SyncSomething) sync.SyncSomething\n}\n"), 0600)).To(Succeed())
			generate("usessync", "", map[string]string{"UsesSync": "{}"})
			expectCounterfeiterAPI("usessync", filepath.Join("usessync", "usessyncfakes"))
		})
	})

	when("the interface's signatures use packages named like the fake's local variables", func() {
		it("gives the locals names that do not shadow the packages", func() {
			for _, name := range []string{"fake", "stub", "ret"} {
				dir := filepath.Join(baseDir, "shadow", name)
				g.Expect(os.MkdirAll(dir, 0700)).To(Succeed())
				g.Expect(os.WriteFile(filepath.Join(dir, name+".go"), []byte("package "+name+"\n\ntype Thing struct{}\n"), 0600)).To(Succeed())
			}
			g.Expect(os.WriteFile(filepath.Join(baseDir, "shadow", "shadow.go"), []byte(`package shadow

import (
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/shadow/fake"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/shadow/ret"
	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/shadow/stub"
)

type Shadowed interface {
	Do([]fake.Thing, stub.Thing) ret.Thing
}
`), 0600)).To(Succeed())
			generate("shadow", "", map[string]string{"Shadowed": "{}"})
			expectCounterfeiterAPI("shadow", filepath.Join("shadow", "shadowfakes"))
		})
	})

	when("several interfaces share one file", func() {
		it("generates fakes that build together", func() {
			generate("", `{filename: "fakes.go"}`, map[string]string{
				"FirstInterface":      "{}",
				"HasVarArgs":          "{}",
				"Something":           "{}",
				"unexportedInterface": "{}",
			})
			b, err := os.ReadFile(filepath.Join(baseDir, "fixturesfakes", "fakes.go"))
			g.Expect(err).NotTo(HaveOccurred())
			for _, name := range []string{"FakeFirstInterface", "FakeHasVarArgs", "FakeSomething", "FakeUnexportedInterface"} {
				g.Expect(string(b)).To(ContainSubstring("type " + name + " struct"))
			}
			expectCounterfeiterAPI("fixturesfakes", "fixturesfakes")
		})
	})
}

// fakeAPI type-checks the package in dir and describes the API of every type
// declared in the given files: its type parameters, its exported fields and
// the exported methods of its pointer type. Parameter names are left out, so
// fakes that name their parameters differently compare equal.
func fakeAPI(dir string, files []string) (map[string]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 || len(pkgs[0].Errors) > 0 {
		return nil, fmt.Errorf("loading %s: %v", dir, pkgs)
	}
	pkg := pkgs[0]
	wanted := map[string]bool{}
	for _, file := range files {
		wanted[file] = true
	}
	api := map[string]string{}
	for _, file := range pkg.Syntax {
		if !wanted[filepath.Base(pkg.Fset.File(file.Pos()).Name())] {
			continue
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, s := range gen.Specs {
				name := s.(*ast.TypeSpec).Name.Name
				api[name] = typeAPI(pkg.Types.Scope().Lookup(name).Type().(*types.Named))
			}
		}
	}
	return api, nil
}

func typeAPI(named *types.Named) string {
	var lines []string
	for i := 0; i < named.TypeParams().Len(); i++ {
		tp := named.TypeParams().At(i)
		lines = append(lines, "type parameter "+tp.Obj().Name()+" "+types.TypeString(tp.Constraint(), nil))
	}
	if st, ok := named.Underlying().(*types.Struct); ok {
		for i := 0; i < st.NumFields(); i++ {
			if f := st.Field(i); f.Exported() {
				lines = append(lines, "field "+f.Name()+" "+types.TypeString(f.Type(), nil))
			}
		}
	}
	methods := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < methods.Len(); i++ {
		if fn := methods.At(i).Obj(); fn.Exported() {
			lines = append(lines, "method "+fn.Name()+signatureTypes(fn.Type().(*types.Signature)))
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func signatureTypes(sig *types.Signature) string {
	list := func(vars *types.Tuple, variadic bool) string {
		parts := make([]string, vars.Len())
		for i := range parts {
			parts[i] = types.TypeString(vars.At(i).Type(), nil)
			if variadic && i == len(parts)-1 {
				parts[i] = "..." + strings.TrimPrefix(parts[i], "[]")
			}
		}
		return strings.Join(parts, ", ")
	}
	return "(" + list(sig.Params(), sig.Variadic()) + ") (" + list(sig.Results(), false) + ")"
}

// mockeryBehaviourTest exercises a mockery-generated fake the way
// generated_fakes_test.go exercises counterfeiter's, using only the standard
// library so it runs inside the temporary fixtures module.
const mockeryBehaviourTest = `package fixturesfakes_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/fixturesfakes"
)

func TestBehaviour(t *testing.T) {
	fake := new(fixturesfakes.FakeSomething)
	fake.DoThingsReturns(1, nil)
	fake.DoThingsReturnsOnCall(1, 2, errors.New("second"))
	if n, err := fake.DoThings("a", 1); n != 1 || err != nil {
		t.Fatalf("first call returned %d, %v", n, err)
	}
	if n, err := fake.DoThings("b", 2); n != 2 || err == nil {
		t.Fatalf("second call returned %d, %v", n, err)
	}
	if got := fake.DoThingsCallCount(); got != 2 {
		t.Fatalf("call count %d", got)
	}
	if s, v := fake.DoThingsArgsForCall(1); s != "b" || v != 2 {
		t.Fatalf("args for call 1: %q, %d", s, v)
	}
	want := []fixturesfakes.FakeSomethingDoThingsArgs{{Arg1: "a", Arg2: 1}, {Arg1: "b", Arg2: 2}}
	if got := fake.DoThingsArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("args %v", got)
	}
	fake.DoThingsCalls(func(s string, _ uint64) (int, error) { return len(s), nil })
	if n, _ := fake.DoThings("abc", 0); n != 3 {
		t.Fatalf("stub returned %d", n)
	}

	slice := []byte{1}
	fake.DoASlice(slice)
	slice[0] = 2
	if got := fake.DoASliceArgsForCall(0); got[0] != 1 {
		t.Fatal("slice argument was not copied")
	}

	if got := fake.CallOrder(); !reflect.DeepEqual(got, []string{"DoThings", "DoThings", "DoThings", "DoASlice"}) {
		t.Fatalf("call order %v", got)
	}
	if got := fake.Invocations()["DoThings"]; !reflect.DeepEqual(got[0], []interface{}{"a", uint64(1)}) {
		t.Fatalf("invocations %v", got)
	}

	varArgs := new(fixturesfakes.FakeHasVarArgs)
	varArgs.DoMoreThingsReturns(5)
	if n := varArgs.DoMoreThings(1, 2, "x", "y"); n != 5 {
		t.Fatalf("variadic call returned %d", n)
	}
	if _, _, rest := varArgs.DoMoreThingsArgsForCall(0); !reflect.DeepEqual(rest, []string{"x", "y"}) {
		t.Fatalf("variadic args %v", rest)
	}
}
`
