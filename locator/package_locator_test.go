package locator_test

import (
	"go/ast"
	"path"
	"runtime"

	"github.com/maxbrunsfeld/counterfeiter/model"

	. "github.com/maxbrunsfeld/counterfeiter/locator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Locator", func() {
	Describe("finding functions in a package", func() {
		var packageToInterfacify *model.PackageToInterfacify
		var err error

		JustBeforeEach(func() {
			packageToInterfacify, err = GetFunctionsFromDirectory("os", path.Join(runtime.GOROOT(), "src/os"))
		})

		Context("when the package exists", func() {
			It("should have the correct name", func() {
				Expect(packageToInterfacify.Name).To(Equal("os"))
			})

			It("should have the correct import path", func() {
				Expect(packageToInterfacify.ImportPath).To(HavePrefix(runtime.GOROOT()))
				Expect(packageToInterfacify.ImportPath).To(HaveSuffix("src/os"))
			})

			It("should have the correct methods", func() {
				Expect(len(packageToInterfacify.Funcs)).To(BeNumerically(">", 0))

				var findProcessMethod *ast.FuncDecl
				for _, funcNode := range packageToInterfacify.Funcs {
					if funcNode.Name.Name == "FindProcess" {
						findProcessMethod = funcNode
						break
					}
				}

				Expect(findProcessMethod).ToNot(BeNil())

				// Expect(packageToInterfacify.Methods[1].Field.Names[0].Name).To(Equal("DoNothing"))
				// Expect(packageToInterfacify.Methods[1].Imports).To(HaveLen(1))
				// Expect(packageToInterfacify.Methods[2].Field.Names[0].Name).To(Equal("DoASlice"))
				// Expect(packageToInterfacify.Methods[2].Imports).To(HaveLen(1))
				// Expect(packageToInterfacify.Methods[3].Field.Names[0].Name).To(Equal("DoAnArray"))
				// Expect(packageToInterfacify.Methods[3].Imports).To(HaveLen(1))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		// Context("when it does not exist", func() {
		// 	BeforeEach(func() {
		// 		interfaceName = "GARBAGE"
		// 	})

		// 	It("returns an error", func() {
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
	})

	Describe("one package, 2 files, duplicate imports", func() {})

	Describe("finding an exported function with renamed imports", func() {})

	Describe("finding an interface with dot imports", func() {})
})
