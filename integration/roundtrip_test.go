package integration_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func runTests(t *testing.T, when spec.G, it spec.S) {
	log.SetOutput(io.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	var (
		baseDir         string
		relativeDir     string
		testDir         string
		copyDirFunc     func()
		copyFileFunc    func(name string)
		initModuleFunc  func()
		writeToTestData bool
	)

	name := "working with a module"

	it.Before(func() {
		RegisterTestingT(t)
		var err error
		testDir, err = os.MkdirTemp("", "counterfeiter-integration")
		Expect(err).NotTo(HaveOccurred())
		os.Unsetenv("GOPATH")
		baseDir = testDir
		err = os.MkdirAll(baseDir, 0777)
		Expect(err).ToNot(HaveOccurred())
		relativeDir = filepath.Join("..", "fixtures")
		copyDirFunc = func() {
			err = os.MkdirAll(baseDir, 0777)
			Expect(err).ToNot(HaveOccurred())
			err = Copy(relativeDir, baseDir)
			Expect(err).ToNot(HaveOccurred())
		}
		copyFileFunc = func(name string) {
			dir := baseDir
			d := filepath.Dir(name)
			if d != "." {
				dir = filepath.Join(dir, d)
			}

			err = os.MkdirAll(dir, 0777)
			Expect(err).ToNot(HaveOccurred())
			b, err := os.ReadFile(filepath.Join(relativeDir, name))
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(filepath.Join(baseDir, name), b, 0755)
			Expect(err).ToNot(HaveOccurred())
		}
		initModuleFunc = func() {
			copyFileFunc("blank.go")
			err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures\n\ngo 1.18\n"), 0755)
			Expect(err).ToNot(HaveOccurred())
		}
		// Set this to true to write the output of tests to the testdata/output
		// directory 🙃 happy debugging!
		// writeToTestData = true
	})

	it.After(func() {
		if baseDir == "" {
			return
		}
		err := os.RemoveAll(testDir)
		Expect(err).ToNot(HaveOccurred())
	})

	when("generating a fake for stdlib interfaces", func() {
		const (
			noHeader   = "noheader"
			withHeader = "header"
		)
		t := func(header, variant string) {
			it("succeeds", func() {
				initModuleFunc()
				cache := &generator.FakeCache{}
				f, err := generator.NewFake(generator.InterfaceOrFunction, "WriteCloser", "io", "FakeWriteCloser", "custom", header, baseDir, cache)
				Expect(err).NotTo(HaveOccurred())
				b, err := f.Generate(true) // Flip to false to see output if goimports fails
				Expect(err).NotTo(HaveOccurred())
				if writeToTestData {
					WriteOutput(b, filepath.Join("testdata", "output", "write_closer", "actual."+variant+".go"))
				}
				WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_write_closer."+variant+".go"))
				RunBuild(baseDir)
				b2, err := os.ReadFile(filepath.Join("testdata", "expected_fake_writecloser."+variant+".txt"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b2)).To(Equal(string(b)))
			})
		}
		t("", noHeader)
		t("// some header\n//\n\n", withHeader)
	})

	when("generating an interface for a package", func() {
		it("succeeds", func() {
			initModuleFunc()
			cache := &generator.FakeCache{}
			f, err := generator.NewFake(generator.Package, "", "os", "Os", "custom", "", baseDir, cache)
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true) // Flip to false to see output if goimports fails
			Expect(err).NotTo(HaveOccurred())
			if writeToTestData {
				WriteOutput(b, filepath.Join("testdata", "output", "package_mode", "actual.go"))
			}
			WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_os.go"))
			RunBuild(baseDir)
		})
	})

	when("generating interfaces using type aliases", func() {
		it.Before(func() {
			relativeDir = filepath.Join(relativeDir, "type_aliases")
			copyDirFunc()
		})
		it("imports the aliased type, not the underlying type", func() {
			cache := &generator.FakeCache{}
			pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/type_aliases"
			interfaceName := "WithAliasedType"
			fakePackageName := "type_aliasesfakes"
			f, err := generator.NewFake(generator.InterfaceOrFunction, interfaceName, pkgPath, "Fake"+interfaceName, fakePackageName, "", baseDir, cache)
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(false)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).NotTo(ContainSubstring("primitive"))
			Expect(string(b)).To(ContainSubstring(`"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/type_aliases/extra"`))
		})
	})

	when("generating a fake into the same package as the interface", func() {
		it.Before(func() {
			baseDir = filepath.Join(baseDir, "samepackage")
			relativeDir = filepath.Join(relativeDir, "samepackage")
			copyFileFunc("samepackage.go")
			WriteOutput([]byte("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage\n\ngo 1.18\n"), filepath.Join(baseDir, "go.mod"))
		})

		it("builds without an import cycle, for exported and unexported interfaces", func() {
			cache := &generator.FakeCache{}
			pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage"
			for _, target := range []struct{ name, fake, file string }{
				{"Widget", "FakeWidget", "fake_widget.go"},
				{"gadget", "FakeGadget", "fake_gadget.go"},
			} {
				f, err := generator.NewFake(generator.InterfaceOrFunction, target.name, pkgPath, target.fake, "samepackage", "", baseDir, cache, generator.WithDestinationDir(baseDir))
				Expect(err).NotTo(HaveOccurred())
				b, err := f.Generate(true)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("var _ " + target.name + " = new(" + target.fake + ")"))
				WriteOutput(b, filepath.Join(baseDir, target.file))
			}
			RunBuild(baseDir)
		})

		it("regenerates after the interface changed, even though the stale fake no longer compiles", func() {
			cache := &generator.FakeCache{}
			pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage"
			generate := func() {
				f, err := generator.NewFake(generator.InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "samepackage", "", baseDir, cache, generator.WithDestinationDir(baseDir))
				Expect(err).NotTo(HaveOccurred())
				b, err := f.Generate(true)
				Expect(err).NotTo(HaveOccurred())
				WriteOutput(b, filepath.Join(baseDir, "fake_widget.go"))
			}
			generate()
			RunBuild(baseDir)

			src, err := os.ReadFile(filepath.Join(baseDir, "samepackage.go"))
			Expect(err).NotTo(HaveOccurred())
			changed := strings.Replace(string(src), "Do(Thing) (Thing, error)\n", "Do(Thing) (Thing, error)\n\tUndo() error\n", 1)
			Expect(changed).NotTo(Equal(string(src)))
			WriteOutput([]byte(changed), filepath.Join(baseDir, "samepackage.go"))

			generate()
			RunBuild(baseDir)
		})
	})

	when("generating a fake into the external test package of the interface's package", func() {
		it.Before(func() {
			baseDir = filepath.Join(baseDir, "samepackage")
			relativeDir = filepath.Join(relativeDir, "samepackage")
			copyFileFunc("samepackage.go")
			WriteOutput([]byte("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage\n\ngo 1.18\n"), filepath.Join(baseDir, "go.mod"))
			WriteOutput([]byte("package samepackage_test\n\nimport \"testing\"\n\nfunc TestFake(t *testing.T) {\n\tw := &FakeWidget{}\n\tif w.DoCallCount() != 0 {\n\t\tt.Fatal(\"unexpected call\")\n\t}\n}\n"), filepath.Join(baseDir, "samepackage_test.go"))
		})

		it("imports the package under test and vets together with the black-box tests", func() {
			cache := &generator.FakeCache{}
			pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage"
			f, err := generator.NewFake(generator.InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "samepackage_test", "", baseDir, cache, generator.WithDestinationDir(baseDir), generator.WithTestPackage())
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(ContainSubstring("package samepackage_test\n"))
			Expect(string(b)).To(ContainSubstring("var _ samepackage.Widget = new(FakeWidget)"))
			WriteOutput(b, filepath.Join(baseDir, "fake_widget_test.go"))
			RunBuild(baseDir)
			RunVet(baseDir)
		})
	})

	when("a test file imports the fake package that does not exist yet", func() {
		it.Before(func() {
			baseDir = filepath.Join(baseDir, "widgets")
			WriteOutput([]byte("module example.com/widgets\n\ngo 1.18\n"), filepath.Join(baseDir, "go.mod"))
			WriteOutput([]byte("package widgets\n\ntype Widget interface {\n\tDo(string) error\n}\n"), filepath.Join(baseDir, "widgets.go"))
			WriteOutput([]byte("package widgets_test\n\nimport (\n\t\"testing\"\n\n\t\"example.com/widgets/widgetsfakes\"\n)\n\nfunc TestWidget(t *testing.T) {\n\tvar _ = &widgetsfakes.FakeWidget{}\n}\n"), filepath.Join(baseDir, "widgets_test.go"))
		})

		it("generates the fake the test file is waiting for", func() {
			cache := &generator.FakeCache{}
			f, err := generator.NewFake(generator.InterfaceOrFunction, "Widget", "example.com/widgets", "FakeWidget", "widgetsfakes", "", baseDir, cache)
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true)
			Expect(err).NotTo(HaveOccurred())
			WriteOutput(b, filepath.Join(baseDir, "widgetsfakes", "fake_widget.go"))
			RunBuild(baseDir)
			RunVet(baseDir)
		})
	})

	when("the target package has errors that do not involve the target", func() {
		const (
			pkgPath       = "example.com/fooer"
			missingImport = "fake.com/this/package/doesnt/exist"
		)
		var (
			write    func(name, content string)
			generate func(target string) (string, error)
		)

		it.Before(func() {
			baseDir = filepath.Join(baseDir, "fooer")
			write = func(name, content string) {
				WriteOutput([]byte(content), filepath.Join(baseDir, name))
			}
			generate = func(target string) (string, error) {
				cache := &generator.FakeCache{}
				f, err := generator.NewFake(generator.InterfaceOrFunction, target, pkgPath, "Fake"+target, "fooerfakes", "", baseDir, cache)
				if err != nil {
					return "", err
				}
				b, err := f.Generate(true)
				return string(b), err
			}
			write("go.mod", "module "+pkgPath+"\n\ngo 1.18\n")
			write("fooer.go", "package fooer\n\ntype Fooer interface {\n\tSayHello(audience string) string\n}\n")
		})

		it("ignores an import that cannot be resolved when the target does not use it", func() {
			write("report.go", "package fooer\n\nimport nonexistent \""+missingImport+"\"\n\nfunc ProcessFooer(f Fooer) nonexistent.Report {\n\treturn nonexistent.NewReport().WithFooer(f)\n}\n")
			out, err := generate("Fooer")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("SayHello(arg1 string) string"))
			Expect(out).NotTo(ContainSubstring("invalid type"))
		})

		it("ignores a type error in another file", func() {
			write("broken.go", "package fooer\n\nvar broken int = \"not an int\"\n")
			out, err := generate("Fooer")
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring("SayHello(arg1 string) string"))
			Expect(out).NotTo(ContainSubstring("invalid type"))
		})

		it("fails when a method uses a type that could not be loaded", func() {
			write("report.go", "package fooer\n\nimport nonexistent \""+missingImport+"\"\n\ntype Reporter interface {\n\tReport() nonexistent.Report\n}\n")
			_, err := generate("Reporter")
			Expect(err).To(MatchError(ContainSubstring("method Report")))
			Expect(err).To(MatchError(ContainSubstring(missingImport)))
		})

		it("fails when the interface embeds an interface that could not be loaded", func() {
			write("report.go", "package fooer\n\nimport nonexistent \""+missingImport+"\"\n\ntype Reporter interface {\n\tnonexistent.Reporter\n\tFooer\n}\n")
			_, err := generate("Reporter")
			Expect(err).To(MatchError(ContainSubstring("embedded interface")))
			Expect(err).To(MatchError(ContainSubstring(missingImport)))
		})

		it("still fails on a syntax error in another file", func() {
			write("broken.go", "package fooer\n\nfunc (\n")
			_, err := generate("Fooer")
			Expect(err).To(HaveOccurred())
		})
	})

	when(name, func() {
		t := func(interfaceName string, filename string, subDir string, files ...string) {
			when("working with "+filename, func() {
				it.Before(func() {
					if subDir != "" {
						baseDir = filepath.Join(baseDir, subDir)
						relativeDir = filepath.Join(relativeDir, subDir)
					}
					log.Println(testDir)
					copyFileFunc(filename)
					for i := range files {
						copyFileFunc(files[i])
					}
				})

				it("succeeds", func() {
					suffix := strings.Replace(subDir, "\\", "/", -1)
					if suffix != "" {
						suffix = "/" + suffix
					}
					WriteOutput([]byte(fmt.Sprintf("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures%s\n\ngo 1.18\n", suffix)), filepath.Join(baseDir, "go.mod"))
					cache := &generator.FakeCache{}
					f, err := generator.NewFake(generator.InterfaceOrFunction, interfaceName, fmt.Sprintf("github.com/maxbrunsfeld/counterfeiter/v6/fixtures%s", suffix), "Fake"+interfaceName, "fixturesfakes", "", baseDir, cache)
					Expect(err).NotTo(HaveOccurred())
					b, err := f.Generate(true) // Flip to false to see output if goimports fails
					Expect(err).NotTo(HaveOccurred())
					if writeToTestData {
						WriteOutput(b, filepath.Join("testdata", "output", strings.Replace(filename, ".go", "", -1), "actual.go"))
					}
					WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_"+filename))
					RunBuild(baseDir)
				})
			})
		}
		t("SomethingElse", "compound_return.go", "")
		t("DotImports", "dot_imports.go", "")
		t("EmbedsInterfaces", "embeds_interfaces.go", "", filepath.Join("another_package", "types.go"))
		t("AliasedInterface", "aliased_interfaces.go", "", filepath.Join("another_package", "types.go"))
		t("HasImports", "has_imports.go", "")
		t("HasOtherTypes", "has_other_types.go", "", "other_types.go")
		t("HasVarArgs", "has_var_args.go", "")
		t("HasVarArgsWithLocalTypes", "has_var_args.go", "")
		t("ImportsGoHyphenPackage", "imports_go_hyphen_package.go", "", filepath.Join("go-hyphenpackage", "fixture.go"))
		t("GenericImportedConstraint", "generic_imported_constraint.go", "", filepath.Join("go-hyphenpackage", "fixture.go"))
		t("FirstInterface", "multiple_interfaces.go", "")
		t("SecondInterface", "multiple_interfaces.go", "")
		t("InlineStructParams", "inline_struct_params.go", "")
		t("RequestFactory", "request_factory.go", "")
		t("ReusesArgTypes", "reuses_arg_types.go", "")
		t("SomethingWithForeignInterface", "something_remote.go", "", filepath.Join("aliased_package", "in_aliased_package.go"))
		t("Something", "something.go", "")
		t("SomethingFactory", "typed_function.go", "")
		t("SyncSomething", "interface.go", "sync")
		t("GenericInterfaceComparable", "genericinterface.go", "genericinterface")

		when("working with duplicate packages", func() {
			t := func(interfaceName string, offset string, fakePackageName string) {
				when("working with "+interfaceName, func() {
					it.Before(func() {
						relativeDir = filepath.Join(relativeDir, "dup_packages")
						copyDirFunc()
					})

					it("succeeds", func() {
						pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"
						if offset != "" {
							pkgPath = pkgPath + "/" + offset
						}
						cache := &generator.FakeCache{}
						f, err := generator.NewFake(generator.InterfaceOrFunction, interfaceName, pkgPath, "Fake"+interfaceName, fakePackageName, "", baseDir, cache)
						Expect(err).NotTo(HaveOccurred())
						b, err := f.Generate(false) // Flip to false to see output if goimports fails
						Expect(err).NotTo(HaveOccurred())
						if writeToTestData {
							WriteOutput(b, filepath.Join("testdata", "output", "dup_"+strings.ToLower(interfaceName), "actual.go"))
						}
						WriteOutput(b, filepath.Join(baseDir, offset, fakePackageName, "fake_"+strings.ToLower(interfaceName)+".go"))
						RunBuild(filepath.Join(baseDir, offset, fakePackageName))
					})
				})
			}

			t("MultiAB", "foo", "foofakes")
			t("AliasV1", "", "dup_packagesfakes")
		})
	})
}
