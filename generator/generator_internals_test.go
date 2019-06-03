package generator

import (
	"io/ioutil"
	"log"
	"testing"

	reporter "github.com/joefitzgerald/rainbow-reporter"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"golang.org/x/tools/go/packages"
)

func TestGenerator(t *testing.T) {
	log.SetOutput(ioutil.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	spec.Run(t, "Generator", testGenerator, spec.Report(reporter.Rainbow{}))
}

func testGenerator(t *testing.T, when spec.G, it spec.S) {
	var (
		f   *Fake
		err error
	)

	it.Before(func() {
		RegisterTestingT(t)
	})

	when("constructing a fake with NewFake()", func() {
		when("the target is a nonexistent package", func() {
			it("errors", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "nonexistentpackage", "FakeNonExistent", "nonexistentpackagefakes", "", c)
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is a package with a nonexistent interface", func() {
			it("errors", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "os", "FakeNonExistent", "osfakes", "", c)
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is an interface that exists", func() {
			it("succeeds", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "FileInfo", "os", "FakeFileInfo", "osfakes", "", c)
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
					},
					ByPkgPath: map[string]Import{
						"os":   {Alias: "os", PkgPath: "os"},
						"sync": {Alias: "sync", PkgPath: "sync"},
						"time": {Alias: "time", PkgPath: "time"},
					},
				}))
				Expect(f.Function).To(BeZero())
				Expect(f.Packages).NotTo(BeNil())
				Expect(f.Package).NotTo(BeNil())
				Expect(f.Methods).To(HaveLen(6))
			})
		})

		when("the target is a function that exists", func() {
			it("succeeds", func() {
				c := &Cache{}
				f, err = NewFake(InterfaceOrFunction, "HandlerFunc", "net/http", "FakeHandlerFunc", "httpfakes", "", c)
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
					Expect(len(methods)).To(BeNumerically("<=", 53))
				})

				it("can load the methods", func() {
					err := f.findPackage()
					Expect(err).NotTo(HaveOccurred())
					f.loadMethods()
					Expect(len(f.Methods)).To(BeNumerically(">=", 51)) // yes, this is crazy because go 1.11 added a function
					Expect(len(f.Methods)).To(BeNumerically("<=", 53))
					Expect(len(f.Imports.ByAlias)).To(Equal(2))
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
