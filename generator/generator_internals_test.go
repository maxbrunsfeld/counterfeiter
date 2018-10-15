package generator

import (
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/tools/go/packages"
)

func TestGenerator(t *testing.T) {
	// log.SetOutput(ioutil.Discard) // Comment this out to see verbose log output
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

	when("constructing a fake with NewFake()", func() {
		when("the target is a nonexistent package", func() {
			it("errors", func() {
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "nonexistentpackage", "FakeNonExistent", "nonexistentpackagefakes", "")
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is a package with a nonexistent interface", func() {
			it("errors", func() {
				f, err = NewFake(InterfaceOrFunction, "NonExistent", "os", "FakeNonExistent", "osfakes", "")
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeNil())
			})
		})

		when("the target is an interface that exists", func() {
			it("succeeds", func() {
				f, err = NewFake(InterfaceOrFunction, "FileInfo", "os", "FakeFileInfo", "osfakes", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).NotTo(BeNil())
				Expect(f.TargetAlias).To(Equal("os"))
				Expect(f.TargetName).To(Equal("FileInfo"))
				Expect(f.TargetPackage).To(Equal("os"))
				Expect(f.Name).To(Equal("FakeFileInfo"))
				Expect(f.Mode).To(Equal(InterfaceOrFunction))
				Expect(f.DestinationPackage).To(Equal("osfakes"))
				Expect(f.Imports).To(HaveLen(3))
				Expect(f.Imports).To(ConsistOf(
					Import{Alias: "os", Path: "os"},
					Import{Alias: "sync", Path: "sync"},
					Import{Alias: "time", Path: "time"},
				))
				Expect(f.Function).To(BeZero())
				Expect(f.Packages).NotTo(BeNil())
				Expect(f.Package).NotTo(BeNil())
				Expect(f.Methods).To(HaveLen(6))
			})
		})

		when("the target is a function that exists", func() {
			it("succeeds", func() {
				f, err = NewFake(InterfaceOrFunction, "HandlerFunc", "net/http", "FakeHandlerFunc", "httpfakes", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(f).NotTo(BeNil())
				Expect(f.TargetAlias).To(Equal("http"))
				Expect(f.TargetName).To(Equal("HandlerFunc"))
				Expect(f.TargetPackage).To(Equal("net/http"))
				Expect(f.Name).To(Equal("FakeHandlerFunc"))
				Expect(f.Mode).To(Equal(InterfaceOrFunction))
				Expect(f.DestinationPackage).To(Equal("httpfakes"))
				Expect(f.Imports).To(HaveLen(2))
				Expect(f.Imports).To(ConsistOf(
					Import{Alias: "http", Path: "net/http"},
					Import{Alias: "sync", Path: "sync"},
				))
				Expect(f.Function).NotTo(BeZero())
				Expect(f.Packages).NotTo(BeNil())
				Expect(f.Package).NotTo(BeNil())
				Expect(f.Methods).To(HaveLen(0))
				Expect(f.Function.Name).To(Equal("HandlerFunc"))
				Expect(f.Function.FakeName).To(Equal("FakeHandlerFunc"))
				Expect(f.Function.Params).To(HaveLen(2))
				Expect(f.Function.Returns).To(BeEmpty())
			})
		})
	})

	when("manually constructing a fake", func() {
		it.Before(func() {
			f = &Fake{}
		})

		when("there are imports", func() {
			it.Before(func() {
				f.AddImport("sync", "sync")
				f.AddImport("sync", "github.com/maxbrunsfeld/counterfeiter/fixtures/sync")
				f.AddImport("sync", "github.com/maxbrunsfeld/counterfeiter/fixtures/othersync")
			})

			it("always leaves the built-in sync in position 0", func() {
				f.sortImports()
				Expect(f.Imports[0].Alias).To(Equal("sync"))
				Expect(f.Imports[0].Path).To(Equal("sync"))
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
					err := f.loadPackages()
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
					err := f.loadPackages()
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
					err := f.loadPackages()
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
					err := f.loadPackages()
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
					err := f.loadPackages()
					Expect(err).To(HaveOccurred())
				})
			})

			when("targeting the os package", func() {
				it.Before(func() {
					f.TargetPackage = "os"
					err := f.loadPackages()
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
					Expect(len(f.Imports)).To(Equal(2))
				})
			})
		})

		when("working with imports", func() {
			when("there are no imports", func() {
				it("returns an empty alias map", func() {
					m := f.aliasMap()
					Expect(m).To(BeEmpty())
				})

				it("turns a vendor path into the correct import", func() {
					i := f.AddImport("apackage", "github.com/maxbrunsfeld/counterfeiter/fixtures/vendored/vendor/apackage")
					Expect(i.Alias).To(Equal("apackage"))
					Expect(i.Path).To(Equal("apackage"))

					i = f.AddImport("anotherpackage", "vendor/anotherpackage")
					Expect(i.Alias).To(Equal("anotherpackage"))
					Expect(i.Path).To(Equal("anotherpackage"))
				})
			})

			when("there is a single import", func() {
				it.Before(func() {
					f.AddImport("os", "os")
				})

				it("is present in the map", func() {
					expected := Import{Alias: "os", Path: "os"}
					m := f.aliasMap()
					Expect(m).To(HaveLen(1))
					Expect(m).To(HaveKeyWithValue("os", []Import{expected}))
				})

				it("returns the existing imports if there is a path match", func() {
					i := f.AddImport("aliasedos", "os")
					Expect(i.Alias).To(Equal("os"))
					Expect(i.Path).To(Equal("os"))
					Expect(f.Imports).To(HaveLen(1))
					Expect(f.Imports[0].Alias).To(Equal("os"))
					Expect(f.Imports[0].Path).To(Equal("os"))
				})
			})

			when("there are imports", func() {
				it.Before(func() {
					f.Imports = []Import{
						Import{
							Alias: "dup_packages",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages",
						},
						Import{
							Alias: "foo",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
						},
						Import{
							Alias: "foo",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
						},
						Import{
							Alias: "sync",
							Path:  "sync",
						},
					}
				})

				it("collects duplicates", func() {
					m := f.aliasMap()
					Expect(m).To(HaveLen(3))
					Expect(m).To(HaveKey("dup_packages"))
					Expect(m).To(HaveKey("sync"))
					Expect(m).To(HaveKey("foo"))
					Expect(m["foo"]).To(ConsistOf(
						Import{
							Alias: "foo",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
						},
						Import{
							Alias: "foo",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
						},
					))
				})

				it("disambiguates aliases", func() {
					m := f.aliasMap()
					Expect(m).To(HaveLen(3))
					f.disambiguateAliases()
					m = f.aliasMap()
					Expect(m).To(HaveLen(4))
					Expect(m["fooa"]).To(ConsistOf(Import{
						Alias: "fooa",
						Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
					}))
				})

				when("there is a package named sync", func() {
					it.Before(func() {
						f.Imports = []Import{
							Import{
								Alias: "sync",
								Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/othersync",
							},
							Import{
								Alias: "sync",
								Path:  "sync",
							},
							Import{
								Alias: "sync",
								Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/sync",
							},
						}
					})

					it("preserves the stdlib sync alias", func() {
						m := f.aliasMap()
						Expect(m).To(HaveLen(1))
						f.disambiguateAliases()
						m = f.aliasMap()
						Expect(m).To(HaveLen(3))
						Expect(m["sync"]).To(ConsistOf(Import{
							Alias: "sync",
							Path:  "sync",
						}))
						Expect(m["syncb"]).To(ConsistOf(Import{
							Alias: "syncb",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/sync",
						}))
						Expect(m["synca"]).To(ConsistOf(Import{
							Alias: "synca",
							Path:  "github.com/maxbrunsfeld/counterfeiter/fixtures/othersync",
						}))
					})
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

		when("export()", func() {
			it("is a no-op on an empty string", func() {
				Expect(export("")).To(Equal(""))
				Expect(export(" ")).To(Equal(""))
			})

			it("makes the first letter uppercase", func() {
				Expect(export("theUnexportedThing")).To(Equal("TheUnexportedThing"))
			})

			it("leaves exported things unchanged", func() {
				Expect(export("TheExportedThing")).To(Equal("TheExportedThing"))
			})
		})
	})
}
