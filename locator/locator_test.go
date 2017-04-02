package locator_test

import (
	"go/ast"
	"strconv"

	"github.com/maxbrunsfeld/counterfeiter/model"

	. "github.com/maxbrunsfeld/counterfeiter/locator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Locator", func() {
	Describe("finding a named interface in a file", func() {
		var interfaceName string
		var model *model.InterfaceToFake
		var err error

		JustBeforeEach(func() {
			model, err = GetInterfaceFromFilePath(interfaceName, "../fixtures/something.go")
		})

		Context("when it exists", func() {
			BeforeEach(func() {
				interfaceName = "Something"
			})

			It("should have the correct name", func() {
				Expect(model.Name).To(Equal("Something"))
			})

			It("should have the correct package name", func() {
				Expect(model.PackageName).To(Equal("fixtures"))
			})

			It("should have the correct import path", func() {
				Expect(model.ImportPath).To(HavePrefix("github.com"))
				Expect(model.ImportPath).To(HaveSuffix("counterfeiter/fixtures"))
			})

			It("should have the correct methods", func() {
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

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when it does not exist", func() {
			BeforeEach(func() {
				interfaceName = "GARBAGE"
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("finding an interface described by a named function from a file", func() {
		var interfaceName string
		var model *model.InterfaceToFake
		var err error

		JustBeforeEach(func() {
			model, err = GetInterfaceFromFilePath(interfaceName, "../fixtures/request_factory.go")
		})

		Context("when it exists", func() {
			BeforeEach(func() {
				interfaceName = "RequestFactory"
			})

			It("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("RequestFactory"))
				Expect(model.RepresentedByInterface).To(BeFalse())
			})

			It("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("RequestFactory"))
				Expect(model.Methods[0].Imports).To(HaveLen(1))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when it does not exist", func() {
			BeforeEach(func() {
				interfaceName = "Whoops!"
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("finding an interface with duplicate imports", func() {
		var model *model.InterfaceToFake
		var err error

		JustBeforeEach(func() {
			model, err = GetInterfaceFromFilePath("AB", "../fixtures/dup_packages/dup_packagenames.go")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("AB"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		It("should have methods", func() {
			Expect(model.Methods).To(HaveLen(4))
			Expect(model.Methods[0].Field.Names[0].Name).To(Equal("A"))
			Expect(collectImports(model.Methods[0].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
			Expect(model.Methods[1].Field.Names[0].Name).To(Equal("FromA"))
			Expect(collectImports(model.Methods[1].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1"))
			Expect(model.Methods[2].Field.Names[0].Name).To(Equal("B"))
			Expect(collectImports(model.Methods[2].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
			Expect(model.Methods[3].Field.Names[0].Name).To(Equal("FromB"))
			Expect(collectImports(model.Methods[3].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1"))
		})
	})

	Describe("finding an interface with duplicate indirect imports", func() {
		var model *model.InterfaceToFake
		var err error

		JustBeforeEach(func() {
			model, err = GetInterfaceFromFilePath("DupAB", "../fixtures/dup_packages/dupAB.go")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("DupAB"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		It("should have methods", func() {
			Expect(model.Methods).To(HaveLen(2))
			Expect(model.Methods[0].Field.Names[0].Name).To(Equal("A"))
			Expect(collectImports(model.Methods[0].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/a/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
			Expect(model.Methods[1].Field.Names[0].Name).To(Equal("B"))
			Expect(collectImports(model.Methods[1].Imports)).To(ConsistOf(
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages/b/v1",
				"github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"))
		})
	})

	Describe("finding an interface with dot imports", func() {
		var model *model.InterfaceToFake
		var err error

		JustBeforeEach(func() {
			model, err = GetInterfaceFromFilePath("DotImports", "../fixtures/dot_imports.go")
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a model representing the named function alias", func() {
			Expect(model.Name).To(Equal("DotImports"))
			Expect(model.RepresentedByInterface).To(BeTrue())
		})

		It("should have a single method", func() {
			Expect(model.Methods).To(HaveLen(1))
			// Expect(model.Methods[0].Names[0].Name).To(Equal("DoThings"))
		})
	})

	Describe("finding an interface in vendored code", func() {
		var model *model.InterfaceToFake
		var err error

		Context("when the vendor dir is in the same directory", func() {
			JustBeforeEach(func() {
				model, err = GetInterfaceFromFilePath("FooInterface", "../fixtures/vendored/foo.go")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("FooInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			It("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("FooVendor"))
			})
		})

		Context("when the vendor dir is in a parent directory", func() {
			JustBeforeEach(func() {
				model, err = GetInterfaceFromFilePath("BazInterface", "../fixtures/vendored/baz/baz.go")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("BazInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			It("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("FooVendor"))
			})
		})

		Context("when the vendor code shadows a higher level", func() {
			JustBeforeEach(func() {
				model, err = GetInterfaceFromFilePath("BarInterface", "../fixtures/vendored/bar/bar.go")
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a model representing the named function alias", func() {
				Expect(model.Name).To(Equal("BarInterface"))
				Expect(model.RepresentedByInterface).To(BeTrue())
			})

			It("should have a single method", func() {
				Expect(model.Methods).To(HaveLen(1))
				Expect(model.Methods[0].Field.Names[0].Name).To(Equal("BarVendor"))
			})
		})
	})
})

func collectImports(specs map[string]*ast.ImportSpec) []string {
	imports := []string{}
	for _, v := range specs {
		s, err := strconv.Unquote(v.Path.Value)
		Expect(err).NotTo(HaveOccurred())
		imports = append(imports, s)
	}
	return imports
}
