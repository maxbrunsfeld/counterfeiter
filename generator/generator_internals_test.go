package generator

import (
	"go/types"
	"io"
	"log"
	"path/filepath"
	"sync"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/tools/go/packages"
)

func TestGenerator(t *testing.T) {
	log.SetOutput(io.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	spec.Run(t, "Generator", testGenerator, spec.Report(report.Terminal{}))
}

func testGenerator(t *testing.T, when spec.G, it spec.S) {
	var (
		f   *Fake
		err error
	)

	it.Before(func() {
		RegisterTestingT(t)
	})

	when("generating the same fake from several goroutines", func() {
		it("produces the same output every time", func() {
			f, err = NewFake(InterfaceOrFunction, "FileInfo", "os", "FakeFileInfo", "osfakes", "", "", &Cache{})
			Expect(err).NotTo(HaveOccurred())
			want, err := f.Generate(false)
			Expect(err).NotTo(HaveOccurred())

			results := make([][]byte, 8)
			errs := make([]error, len(results))
			var wg sync.WaitGroup
			for i := range results {
				wg.Add(1)
				go func() {
					defer wg.Done()
					results[i], errs[i] = f.Generate(false)
				}()
			}
			wg.Wait()
			for i := range results {
				Expect(errs[i]).NotTo(HaveOccurred())
				Expect(string(results[i])).To(Equal(string(want)))
			}
		})
	})

	when("constructing a fake with NewFake()", func() {
		when("the target is a nonexistent package", func() {
			it("errors", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "nonexistentpackage", "FakeNonExistent", "nonexistentpackagefakes", "", "", c)
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is a package with a nonexistent interface", func() {
			it("errors", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "os", "FakeNonExistent", "osfakes", "", "", c)
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is an interface that exists", func() {
			it("succeeds", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "FileInfo", "os", "FakeFileInfo", "osfakes", "", "", c)
				Expect(err).NotTo(HaveOccurred())
				Expect(f).NotTo(BeNil())
				Expect(f.TargetAlias).To(Equal("os"))
				Expect(f.TargetName).To(Equal("FileInfo"))
				Expect(f.TargetPackage).To(Equal("os"))
				Expect(f.Name).To(Equal("FakeFileInfo"))
				Expect(f.Mode).To(Equal(InterfaceOrFunction))
				Expect(f.DestinationPackage).To(Equal("osfakes"))
				Expect(f.Imports).To(BeEquivalentTo(Imports{
					ByAlias: map[string]Import{
						"os":   {Alias: "os", PkgPath: "os"},
						"sync": {Alias: "sync", PkgPath: "sync"},
						"time": {Alias: "time", PkgPath: "time"},
						"fs":   {Alias: "fs", PkgPath: "io/fs"},
					},
					ByPkgPath: map[string]Import{
						"os":    {Alias: "os", PkgPath: "os"},
						"sync":  {Alias: "sync", PkgPath: "sync"},
						"time":  {Alias: "time", PkgPath: "time"},
						"io/fs": {Alias: "fs", PkgPath: "io/fs"},
					},
				}))
				Expect(f.Function).To(BeZero())
				Expect(f.Packages).NotTo(BeNil())
				Expect(f.Package).NotTo(BeNil())
				Expect(f.Methods).To(HaveLen(6))
			})
		})

		when("the target is an interface in a third-party module", func() {
			it("succeeds", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "GomegaMatcher", "github.com/onsi/gomega/types", "FakeGomegaMatcher", "typesfakes", "", "", c)
				Expect(err).NotTo(HaveOccurred())
				Expect(f.TargetPackage).To(Equal("github.com/onsi/gomega/types"))
				b, err := f.Generate(true)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring(`"github.com/onsi/gomega/types"`))
				Expect(string(b)).To(ContainSubstring("func (fake *FakeGomegaMatcher) Match(arg1 any) (bool, error)"))
				Expect(string(b)).To(ContainSubstring("var _ types.GomegaMatcher = new(FakeGomegaMatcher)"))
			})
		})

		when("the target is an unexported type that nothing else in its package refers to", func() {
			it("still finds it", func() {
				c := &Cache{}
				for _, target := range []string{"unexportedInterface", "unexportedFunc"} {
					f, err = NewFake(InterfaceOrFunction, target, "github.com/maxbrunsfeld/counterfeiter/v6/fixtures", "Fake"+target, "fixturesfakes", "", "", c)
					Expect(err).NotTo(HaveOccurred(), target)
					Expect(f.TargetName).To(Equal(target))
				}
			})
		})

		when("the target is a generic interface whose constraint comes from another package", func() {
			it("imports the constraint's package and qualifies it with that package's name", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "GenericImportedConstraint", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures", "FakeGenericImportedConstraint", "fixturesfakes", "", "", c)
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Imports.ByPkgPath).To(HaveKey("github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage"))
				Expect(f.GenericTypeParametersAndConstraints).To(Equal("[T hyphenpackage.Hyphenated]"))
				Expect(f.GenericTypeParameters).To(Equal("[T]"))
			})
		})

		when("the destination directory already holds a package", func() {
			it("joins that package rather than naming one after the directory", func() {
				dir, err := filepath.Abs(filepath.Join("..", "fixtures", "seeded", "fakes"))
				Expect(err).NotTo(HaveOccurred())
				f, err = NewFake(InterfaceOrFunction, "Sower", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/seeded", "FakeSower", "fakes", "", "", &Cache{}, WithDestinationDir(dir))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.DestinationPackage).To(Equal("seeded_fakes"))
				b, err := f.Generate(false)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("package seeded_fakes\n"))
				Expect(string(b)).To(ContainSubstring("var _ seeded.Sower = new(FakeSower)"))
			})

			it("keeps the given name when the directory holds no Go files", func() {
				f, err = NewFake(InterfaceOrFunction, "Sower", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/seeded", "FakeSower", "fakes", "", "", &Cache{}, WithDestinationDir(t.TempDir()))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.DestinationPackage).To(Equal("fakes"))
			})
		})

		when("the destination is the same package as the target", func() {
			var (
				pkgPath string
				dir     string
			)

			it.Before(func() {
				pkgPath = "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/samepackage"
				var err error
				dir, err = filepath.Abs(filepath.Join("..", "fixtures", "samepackage"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("does not import the target package and leaves its types unqualified", func() {
				f, err = NewFake(InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "samepackage", "", "", &Cache{}, WithDestinationDir(dir))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.TargetAlias).To(BeEmpty())
				Expect(f.Imports.ByPkgPath).NotTo(HaveKey(pkgPath))
				b, err := f.Generate(true)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("Do(arg1 Thing) (Thing, error)"))
				Expect(string(b)).To(ContainSubstring("var _ Widget = new(FakeWidget)"))
				Expect(string(b)).NotTo(ContainSubstring("samepackage."))
			})

			it("keeps the fake of an unexported interface unexported and asserts that it implements it", func() {
				f, err = NewFake(InterfaceOrFunction, "gadget", pkgPath, "FakeGadget", "samepackage", "", "", &Cache{}, WithDestinationDir(dir))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Name).To(Equal("fakeGadget"))
				b, err := f.Generate(false)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("type fakeGadget struct"))
				Expect(string(b)).To(ContainSubstring("var _ gadget = new(fakeGadget)"))
				Expect(string(b)).NotTo(ContainSubstring("FakeGadget"))
			})

			it("uses a name given with -fake-name as it is, even for an unexported interface", func() {
				f, err = NewFake(InterfaceOrFunction, "gadget", pkgPath, "GadgetDouble", "samepackage", "", "", &Cache{}, WithDestinationDir(dir), WithExplicitName())
				Expect(err).NotTo(HaveOccurred())
				Expect(f.Name).To(Equal("GadgetDouble"))
				b, err := f.Generate(false)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("var _ gadget = new(GadgetDouble)"))
			})

			it("names the package after the target's package, not its directory", func() {
				hyphenDir, err := filepath.Abs(filepath.Join("..", "fixtures", "go-hyphenpackage"))
				Expect(err).NotTo(HaveOccurred())
				f, err = NewFake(InterfaceOrFunction, "Hyphenated", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/go-hyphenpackage", "FakeHyphenated", "gohyphenpackage", "", "", &Cache{}, WithDestinationDir(hyphenDir))
				Expect(err).NotTo(HaveOccurred())
				Expect(f.DestinationPackage).To(Equal("hyphenpackage"))
				b, err := f.Generate(false)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(ContainSubstring("package hyphenpackage\n"))
			})

			when("the fake goes into the external test package", func() {
				it("names the package <package>_test and imports the target package", func() {
					f, err = NewFake(InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "samepackage_test", "", "", &Cache{}, WithDestinationDir(dir), WithTestPackage())
					Expect(err).NotTo(HaveOccurred())
					Expect(f.DestinationPackage).To(Equal("samepackage_test"))
					Expect(f.TargetAlias).To(Equal("samepackage"))
					Expect(f.Imports.ByPkgPath).To(HaveKey(pkgPath))
					b, err := f.Generate(true)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(b)).To(ContainSubstring("package samepackage_test\n"))
					Expect(string(b)).To(ContainSubstring("Do(arg1 samepackage.Thing) (samepackage.Thing, error)"))
					Expect(string(b)).To(ContainSubstring("var _ samepackage.Widget = new(FakeWidget)"))
				})

				it("refuses an unexported interface, which the test package cannot see", func() {
					f, err = NewFake(InterfaceOrFunction, "gadget", pkgPath, "FakeGadget", "samepackage_test", "", "", &Cache{}, WithDestinationDir(dir), WithTestPackage())
					Expect(err).To(MatchError(And(ContainSubstring("gadget"), ContainSubstring("samepackage_test"), ContainSubstring("unexported"))))
				})

				it("uses the name of the package in the destination directory, not the directory name", func() {
					hyphenDir, err := filepath.Abs(filepath.Join("..", "fixtures", "go-hyphenpackage"))
					Expect(err).NotTo(HaveOccurred())
					f, err = NewFake(InterfaceOrFunction, "WriteCloser", "io", "FakeWriteCloser", "gohyphenpackage_test", "", "", &Cache{}, WithDestinationDir(hyphenDir), WithTestPackage())
					Expect(err).NotTo(HaveOccurred())
					Expect(f.DestinationPackage).To(Equal("hyphenpackage_test"))
					Expect(f.TargetAlias).To(Equal("io"))
					Expect(f.Imports.ByPkgPath).To(HaveKey("io"))
				})

				it("keeps the given name when the destination directory has no Go files", func() {
					f, err = NewFake(InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "other_test", "", "", &Cache{}, WithDestinationDir(t.TempDir()), WithTestPackage())
					Expect(err).NotTo(HaveOccurred())
					Expect(f.DestinationPackage).To(Equal("other_test"))
					Expect(f.TargetAlias).To(Equal("samepackage"))
				})

				it("fakes an interface from a third-party module into the test package", func() {
					testDir, err := filepath.Abs(filepath.Join("..", "fixtures", "externaltest"))
					Expect(err).NotTo(HaveOccurred())
					f, err = NewFake(InterfaceOrFunction, "GomegaMatcher", "github.com/onsi/gomega/types", "FakeGomegaMatcher", "externaltest_test", "", "", &Cache{}, WithDestinationDir(testDir), WithTestPackage())
					Expect(err).NotTo(HaveOccurred())
					Expect(f.DestinationPackage).To(Equal("externaltest_test"))
					Expect(f.TargetAlias).To(Equal("types"))
					Expect(f.Imports.ByPkgPath).To(HaveKey("github.com/onsi/gomega/types"))
					b, err := f.Generate(true)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(b)).To(ContainSubstring("package externaltest_test\n"))
					Expect(string(b)).To(ContainSubstring("var _ types.GomegaMatcher = new(FakeGomegaMatcher)"))
				})
			})

			when("the destination only shares the target's package name", func() {
				it("still imports the target package", func() {
					other := t.TempDir()
					f, err = NewFake(InterfaceOrFunction, "Widget", pkgPath, "FakeWidget", "samepackage", "", "", &Cache{}, WithDestinationDir(other))
					Expect(err).NotTo(HaveOccurred())
					Expect(f.TargetAlias).To(Equal("samepackage"))
					Expect(f.Imports.ByPkgPath).To(HaveKey(pkgPath))
					b, err := f.Generate(false)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(b)).To(ContainSubstring("var _ samepackage.Widget = new(FakeWidget)"))
				})

				it("does not assert an unexported interface, and keeps the fake exported", func() {
					other := t.TempDir()
					f, err = NewFake(InterfaceOrFunction, "gadget", pkgPath, "FakeGadget", "samepackage", "", "", &Cache{}, WithDestinationDir(other))
					Expect(err).NotTo(HaveOccurred())
					Expect(f.Name).To(Equal("FakeGadget"))
					b, err := f.Generate(false)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(b)).NotTo(ContainSubstring("var _ "))
				})
			})
		})

		when("the target is a function that exists", func() {
			it("succeeds", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "HandlerFunc", "net/http", "FakeHandlerFunc", "httpfakes", "", "", c)
				Expect(err).NotTo(HaveOccurred())

				Expect(f).NotTo(BeNil())
				Expect(f.TargetAlias).To(Equal("http"))
				Expect(f.TargetName).To(Equal("HandlerFunc"))
				Expect(f.TargetPackage).To(Equal("net/http"))
				Expect(f.Name).To(Equal("FakeHandlerFunc"))
				Expect(f.Mode).To(Equal(InterfaceOrFunction))
				Expect(f.DestinationPackage).To(Equal("httpfakes"))
				Expect(f.Imports).To(BeEquivalentTo(Imports{
					ByAlias: map[string]Import{
						"http": {Alias: "http", PkgPath: "net/http"},
						"sync": {Alias: "sync", PkgPath: "sync"},
					},
					ByPkgPath: map[string]Import{
						"net/http": {Alias: "http", PkgPath: "net/http"},
						"sync":     {Alias: "sync", PkgPath: "sync"},
					},
				}))
				Expect(f.Function).NotTo(BeZero())
				Expect(f.Packages).NotTo(BeNil())
				Expect(f.Package).NotTo(BeNil())
				Expect(f.Methods).To(HaveLen(0))
				Expect(f.Function.Name).To(Equal("HandlerFunc"))
				Expect(f.Function.Params).To(HaveLen(2))
				Expect(f.Function.Returns).To(BeEmpty())
			})
		})
	})

	when("manually constructing a fake", func() {
		it.Before(func() {
			f = &Fake{Imports: newImports()}
		})

		when("duplicate import package names are added", func() {
			it.Before(func() {
				f.Imports.Add("sync", "sync")
				f.Imports.Add("sync", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/sync")
				f.Imports.Add("sync", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/othersync")
			})

			it("all packages have unique aliases", func() {
				Expect(f.Imports).To(BeEquivalentTo(Imports{
					ByAlias: map[string]Import{
						"sync":  {Alias: "sync", PkgPath: "sync"},
						"synca": {Alias: "synca", PkgPath: "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/sync"},
						"syncb": {Alias: "syncb", PkgPath: "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/othersync"},
					},
					ByPkgPath: map[string]Import{
						"sync": {Alias: "sync", PkgPath: "sync"},
						"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/sync":      {Alias: "synca", PkgPath: "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/sync"},
						"github.com/maxbrunsfeld/counterfeiter/v6/fixtures/othersync": {Alias: "syncb", PkgPath: "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/othersync"},
					},
				}))
			})
		})

		when("inspecting the target", func() {
			when("the target is not set", func() {
				it("IsInterface() is false", func() {
					Expect(f.IsInterface()).To(BeFalse())
				})

				it("IsFunction() is false", func() {
					Expect(f.IsFunction()).To(BeFalse())
				})
			})

			when("the target is an interface", func() {
				it.Before(func() {
					f.Mode = InterfaceOrFunction
					f.TargetPackage = "os"
					f.TargetName = "FileInfo"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).NotTo(HaveOccurred())
					err = f.findPackage()
					Expect(err).NotTo(HaveOccurred())
				})

				it("IsInterface() is true", func() {
					Expect(f.IsInterface()).To(BeTrue())
				})

				it("IsFunction() is false", func() {
					Expect(f.IsFunction()).To(BeFalse())
				})
			})

			when("the target is a function", func() {
				it.Before(func() {
					f.Mode = InterfaceOrFunction
					f.TargetPackage = "net/http"
					f.TargetName = "HandlerFunc"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).NotTo(HaveOccurred())
					err = f.findPackage()
					Expect(err).NotTo(HaveOccurred())
				})

				it("IsInterface() is false", func() {
					Expect(f.IsInterface()).To(BeFalse())
				})

				it("IsFunction() is true", func() {
					Expect(f.IsFunction()).To(BeTrue())
				})
			})

			when("the target is a struct", func() {
				it.Before(func() {
					f.Mode = InterfaceOrFunction
					f.TargetPackage = "net/http"
					f.TargetName = "Client"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).NotTo(HaveOccurred())
					err = f.findPackage()
					Expect(err).To(HaveOccurred())
				})

				it("is not a function", func() {
					Expect(f.IsFunction()).To(BeFalse())
				})

				it("is not an interface", func() {
					Expect(f.IsInterface()).To(BeFalse())
				})
			})
		})

		when("in interface mode", func() {
			it.Before(func() {
				f.Mode = InterfaceOrFunction
			})

			when("targeting the os.FileInfo interface", func() {
				it.Before(func() {
					f.TargetPackage = "os"
					f.TargetName = "FileInfo"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		when("in package mode", func() {
			it.Before(func() {
				f.Mode = Package
			})

			when("targeting a nonexistent package", func() {
				it("returns an error", func() {
					f.TargetPackage = "counterfeiternonexistentpackage"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).To(HaveOccurred())
				})
			})

			when("targeting the os package", func() {
				it.Before(func() {
					f.TargetPackage = "os"
					c := &Cache{}
					err := f.loadPackages(c, "")
					Expect(err).NotTo(HaveOccurred())
				})

				it("can load packages", func() {
					Expect(len(f.Packages)).To(BeNumerically(">=", 1))
					Expect(f.Packages[0].Name).To(Equal("os"))
				})

				it("can find the package with the os package path", func() {
					err := f.findPackage()
					Expect(err).NotTo(HaveOccurred())
					Expect(f.Package).NotTo(BeNil())
					Expect(f.Package).To(Equal(f.Packages[0]))
				})

				it("skips invalid packages", func() {
					var p []*packages.Package
					empty := &packages.Package{}
					p = append(p, empty)
					p = append(p, f.Packages...)
					f.Packages = p
					err := f.findPackage()
					Expect(err).NotTo(HaveOccurred())
					Expect(f.Package).NotTo(BeNil())
					Expect(f.Package).To(Equal(f.Packages[1]))
				})

				it("can identify the method set for the package", func() {
					err := f.findPackage()
					Expect(err).NotTo(HaveOccurred())
					methods := packageMethodSet(f.Package)
					Expect(len(methods)).To(BeNumerically(">=", 51)) // yes, this is crazy because go 1.11 added a function
				})

				it("can load the methods", func() {
					err := f.findPackage()
					Expect(err).NotTo(HaveOccurred())
					err = f.loadMethods()
					Expect(err).NotTo(HaveOccurred())
					Expect(len(f.Methods)).To(BeNumerically(">=", 51)) // yes, this is crazy because go 1.11 added a function
					Expect(len(f.Imports.ByAlias)).To(Equal(3))
				})
			})
		})

		when("working with imports", func() {
			when("there are no imports", func() {
				it("returns an empty alias map", func() {
					Expect(f.Imports.ByAlias).To(BeEmpty())
				})

				it("turns a vendor path into the correct import", func() {
					i := f.Imports.Add("apackage", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/vendored/vendor/apackage")
					Expect(i.Alias).To(Equal("apackage"))
					Expect(i.PkgPath).To(Equal("apackage"))

					i = f.Imports.Add("anotherpackage", "vendor/anotherpackage")
					Expect(i.Alias).To(Equal("anotherpackage"))
					Expect(i.PkgPath).To(Equal("anotherpackage"))
				})
			})

			when("there is a single import", func() {
				it.Before(func() {
					f.Imports.Add("os", "os")
				})

				it("is present in the map", func() {
					Expect(f.Imports).To(BeEquivalentTo(Imports{
						ByAlias: map[string]Import{
							"os": {Alias: "os", PkgPath: "os"},
						},
						ByPkgPath: map[string]Import{
							"os": {Alias: "os", PkgPath: "os"},
						},
					}))
				})

				it("returns the existing imports if there is a path match", func() {
					i := f.Imports.Add("aliasedos", "os")
					Expect(i.Alias).To(Equal("os"))
					Expect(i.PkgPath).To(Equal("os"))
					Expect(f.Imports).To(BeEquivalentTo(Imports{
						ByAlias: map[string]Import{
							"os": {Alias: "os", PkgPath: "os"},
						},
						ByPkgPath: map[string]Import{
							"os": {Alias: "os", PkgPath: "os"},
						},
					}))
				})
			})
		})
	})

	when("helper functions", func() {
		when("unexport()", func() {
			it("is a no-op on an empty string", func() {
				Expect(unexport("")).To(Equal(""))
				Expect(unexport(" ")).To(Equal(""))
			})

			it("makes the first letter lowercase", func() {
				Expect(unexport("TheExportedThing")).To(Equal("theExportedThing"))
			})

			it("leaves unexported things unchanged", func() {
				Expect(unexport("theUnexportedThing")).To(Equal("theUnexportedThing"))
			})
		})

		when("isBuildTranscript()", func() {
			it("recognises the compiler output go list -export attaches to a package that failed to build", func() {
				Expect(isBuildTranscript(packages.Error{Kind: packages.ListError, Msg: "# example.com/widgets\n./fake_widget.go:112:16: missing method Undo"})).To(BeTrue())
			})

			it("leaves positioned and non-build errors alone", func() {
				Expect(isBuildTranscript(packages.Error{Kind: packages.TypeError, Pos: "/a/fake_widget.go:112:16", Msg: "missing method Undo"})).To(BeFalse())
				Expect(isBuildTranscript(packages.Error{Kind: packages.ListError, Pos: "/a/widgets.go:5:2", Msg: "no required module provides package x.invalid/nope"})).To(BeFalse())
				Expect(isBuildTranscript(packages.Error{Kind: packages.ParseError, Msg: "# example.com/widgets"})).To(BeFalse())
			})
		})

		when("hasInvalidType()", func() {
			invalid := types.Typ[types.Invalid]
			str := types.Typ[types.String]
			named := func(underlying types.Type) *types.Named {
				return types.NewNamed(types.NewTypeName(0, nil, "Named", nil), underlying, nil)
			}

			it("is true for a type the loader could not resolve", func() {
				Expect(hasInvalidType(invalid)).To(BeTrue())
			})

			it("is false for valid types", func() {
				Expect(hasInvalidType(nil)).To(BeFalse())
				Expect(hasInvalidType(str)).To(BeFalse())
				Expect(hasInvalidType(types.NewPointer(named(str)))).To(BeFalse())
			})

			it("looks through composite types", func() {
				Expect(hasInvalidType(types.NewPointer(invalid))).To(BeTrue())
				Expect(hasInvalidType(types.NewSlice(invalid))).To(BeTrue())
				Expect(hasInvalidType(types.NewArray(invalid, 2))).To(BeTrue())
				Expect(hasInvalidType(types.NewChan(types.SendRecv, invalid))).To(BeTrue())
				Expect(hasInvalidType(types.NewMap(str, invalid))).To(BeTrue())
				Expect(hasInvalidType(types.NewMap(invalid, str))).To(BeTrue())
				Expect(hasInvalidType(types.NewStruct([]*types.Var{types.NewField(0, nil, "f", invalid, false)}, nil))).To(BeTrue())
				Expect(hasInvalidType(types.NewSignatureType(nil, nil, nil, types.NewTuple(types.NewParam(0, nil, "p", invalid)), nil, false))).To(BeTrue())
				Expect(hasInvalidType(types.NewSignatureType(nil, nil, nil, nil, types.NewTuple(types.NewParam(0, nil, "r", invalid)), false))).To(BeTrue())
			})

			it("does not look through a named type, which prints by name", func() {
				Expect(hasInvalidType(named(invalid))).To(BeFalse())
				Expect(hasInvalidType(types.NewPointer(named(invalid)))).To(BeFalse())
			})
		})

		when("isExported()", func() {
			it("returns false for an empty string", func() {
				Expect(isExported("")).To(BeFalse())
				Expect(isExported(" ")).To(BeFalse())
			})

			it("returns true when the first rune is upper case", func() {
				Expect(isExported("Identifier")).To(BeTrue())
				Expect(isExported("Ʊpsilon")).To(BeTrue())
			})

			it("returns false when the first rune not upper case", func() {
				Expect(isExported("identifier")).To(BeFalse())
				Expect(isExported("ʊpsilon")).To(BeFalse())
			})
		})
	})
}
