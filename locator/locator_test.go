package locator_test

import (
	"fmt"
	"go/ast"
	"log"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/model"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/locator"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestLocator(t *testing.T) {
	log.SetFlags(log.Llongfile)
	spec.Run(t, "Locator", testLocator, spec.Report(report.Terminal{}))
}

func testLocator(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("finding a named interface in a file", func() {
		var interfaceName string
		var model *model.InterfaceToFake
		var err error

		when("when it exists", func() {
			it.Before(func() {
				interfaceName = "Something"
				model, err = locator.GetInterfaceFromFilePath(interfaceName, "../fixtures/something.go")
				if err != nil {
					fmt.Println(err)
				}
			})

			it("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			it("should have the correct name", func() {
				Expect(model.Name).To(Equal("Something"))
			})

			it("should have the correct package name", func() {
				Expect(model.PackageName).To(Equal("fixtures"))
			})

			it("should have the correct import path", func() {
				// Make the code testable even in forked repos :)
				// e.g.: you fork counterfeiter to make a change,
				//       the repo is now github.com/pizzabandit/counterfeiter
				//       you should expect these assertions to still pass
				Expect(model.ImportPath).To(MatchRegexp("^github\\.com/[^/]+/counterfeiter/fixtures$"))
			})

			it("should have the correct methods", func() {
				Expect(model.Methods).To(HaveLen(4))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("DoThings"))
				Expect(model.Methods[0].Imports).To(HaveLen(1))
				Expect(model.Methods[1].Field.Names[0].Name).To(Equal("DoNothing"))
				Expect(model.Methods[1].Imports).To(HaveLen(1))
				Expect(model.Methods[2].Field.Names[0].Name).To(Equal("DoASlice"))
				Expect(model.Methods[2].Imports).To(HaveLen(1))
				Expect(model.Methods[3].Field.Names[0].Name).To(Equal("DoAnArray"))
				Expect(model.Methods[3].Imports).To(HaveLen(1))
			})
		})

		when("when it does not exist", func() {
			it.Before(func() {
				interfaceName = "GARBAGE"
				model, err = locator.GetInterfaceFromFilePath(interfaceName, "../fixtures/something.go")
			})

			it("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	when("finding an interface described by a named function from a file", func() {
		var interfaceName string
		var model *model.InterfaceToFake
		var err error

		when("when it exists", func() {
			it.Before(func() {
				interfaceName = "RequestFactory"
				model, err = locator.GetInterfaceFromFilePath(interfaceName, "../fixtures/request_factory.go")
			})

			it("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("RequestFactory"))
				Expect(model.RepresentedByInterface).To(BeFalse())
			})

			it("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("RequestFactory"))
				Expect(model.Methods[0].Imports).To(HaveLen(1))
			})

			it("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		when("when it does not exist", func() {
			it.Before(func() {
				interfaceName = "Whoops!"
				model, err = locator.GetInterfaceFromFilePath(interfaceName, "../fixtures/request_factory.go")
			})

			it("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	when("finding an interface with duplicate imports", func() {
		var model *model.InterfaceToFake
		var err error

		it.Before(func() {
			model, err = locator.GetInterfaceFromFilePath("AB", "../fixtures/dup_packages/dup_packagenames.go")
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("AB"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		it("should have methods", func() {
			Expect(model.Methods).To(HaveLen(4))

			Expect(model.Methods[0].Field.Names[0].Name).To(Equal("A"))
			Expect(collectImports(model.Methods[0].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))

			Expect(model.Methods[2].Field.Names[0].Name).To(Equal("B"))
			Expect(collectImports(model.Methods[2].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))

			// NOTE: we don't have to recurse into embedded interfaces because
			// we use typeutil.IntuitiveMethodSet
			// but this affects our imports because we're finding imports for the
			// file where the interface is embedded

			Expect(model.Methods[1].Field.Names[0].Name).To(Equal("FromA"))
			Expect(collectImports(model.Methods[1].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo"))

			Expect(model.Methods[3].Field.Names[0].Name).To(Equal("FromB"))
			Expect(collectImports(model.Methods[3].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo"))
		})
	})

	when("finding an interface with duplicate indirect imports", func() {
		var model *model.InterfaceToFake
		var err error

		it.Before(func() {
			model, err = locator.GetInterfaceFromFilePath("DupAB", "../fixtures/dup_packages/dupAB.go")
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("DupAB"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		it("should have methods", func() {
			Expect(model.Methods).To(HaveLen(2))
			Expect(model.Methods[0].Field.Names[0].Name).To(Equal("A"))
			Expect(collectImports(model.Methods[0].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
			Expect(model.Methods[1].Field.Names[0].Name).To(Equal("B"))
			Expect(collectImports(model.Methods[1].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
		})
	})

	when("finding an interface with multiple duplicate indirect imports", func() {
		var model *model.InterfaceToFake
		var err error

		it.Before(func() {
			model, err = locator.GetInterfaceFromFilePath("MultiAB", "../fixtures/dup_packages/foo/multi_import.go")
			Expect(err).NotTo(HaveOccurred())
		})

		it.Focus("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("MultiAB"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		it.Focus("should have methods", func() {
			Expect(model.Methods).To(HaveLen(3))
			Expect(model.Methods[0].Field.Names[0].Name).To(Equal("Mine"))
			Expect(collectImports(model.Methods[0].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
			))
			Expect(model.Methods[1].Field.Names[0].Name).To(Equal("FromA"))
			Expect(collectImports(model.Methods[1].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/foo",
			))
			Expect(model.Methods[2].Field.Names[0].Name).To(Equal("FromB"))
			Expect(collectImports(model.Methods[2].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/foo",
			))
		})
	})

	when("finding an interface with dot imports", func() {
		var model *model.InterfaceToFake
		var err error

		it.Before(func() {
			model, err = locator.GetInterfaceFromFilePath("DotImports", "../fixtures/dot_imports.go")
			Expect(err).NotTo(HaveOccurred())
		})

		it("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("DotImports"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		it("should have a single method", func() {
			Expect(model.Methods).To(HaveLen(1))
		})
	})

	when("finding an interface in vendored code", func() {
		var model *model.InterfaceToFake
		var err error

		when("when the vendor dir is in the same directory", func() {
			it.Before(func() {
				model, err = locator.GetInterfaceFromFilePath("FooInterface", filepath.Join("..", "fixtures", "vendored", "foo.go"))
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("FooInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			it("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("FooVendor"))
			})

			it("should strip vendor from the import path", func() {
				Expect(model.Methods[0].Imports).To(HaveLen(1))
				Expect(model.Methods[0].Imports["xyz123"].Path.Value).To(Equal(`"apackage"`))
			})
		})

		when("when the vendor dir is in a parent directory", func() {
			it.Before(func() {
				model, err = locator.GetInterfaceFromFilePath("BazInterface", "../fixtures/vendored/baz/baz.go")
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("BazInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			it("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("FooVendor"))
			})
		})

		when("when the vendor code shadows a higher level", func() {
			it.Before(func() {
				model, err = locator.GetInterfaceFromFilePath("BarInterface", "../fixtures/vendored/bar/bar.go")
				Expect(err).NotTo(HaveOccurred())
			})

			it("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("BarInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			it("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("BarVendor"))
			})
		})
	})
}

func collectImports(specs map[string]*ast.ImportSpec) []string {
	imports := []string{}
	for _, v := range specs {
		alias := ""
		if v.Name == nil {
			alias = path.Base(strings.Trim(v.Path.Value, `"`))
		} else {
			alias = v.Name.Name
		}
		log.Printf("import %s \"%s\"", alias, v.Path.Value)
		s, err := strconv.Unquote(v.Path.Value)
		Expect(err).NotTo(HaveOccurred())
		imports = append(imports, s)
	}
	return imports
}
