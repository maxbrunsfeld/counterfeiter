package locator_test

import (
	"path"
	"runtime"

	"github.com/maxbrunsfeld/counterfeiter/model"

	. "github.com/maxbrunsfeld/counterfeiter/locator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Locator", func() {
	Describe("finding functions in a package", func() {
		var model *model.PackageToInterfacify
		// var err error

		JustBeforeEach(func() {
			model, _ = GetFunctionsFromDirectory("os", path.Join(runtime.GOROOT(), "src/os"))
		})

		Context("when the package exists", func() {
			It("should have the correct name", func() {
				Expect(model.Name).To(Equal("os"))
			})

			// It("should have the correct package name", func() {
			// 	Expect(model.PackageName).To(Equal("fixtures"))
			// })

			It("should have the correct import path", func() {
				Expect(model.ImportPath).To(HavePrefix(runtime.GOROOT()))
				Expect(model.ImportPath).To(HaveSuffix("src/os"))
			})

			It("should have the correct methods", func() {
				Expect(model.Funcs).To(HaveLen(49))
				Expect(model.Funcs[0].Name.Name).To(Equal("FindProcess"))
				// Expect(model.Methods[1].Field.Names[0].Name).To(Equal("DoNothing"))
				// Expect(model.Methods[1].Imports).To(HaveLen(1))
				// Expect(model.Methods[2].Field.Names[0].Name).To(Equal("DoASlice"))
				// Expect(model.Methods[2].Imports).To(HaveLen(1))
				// Expect(model.Methods[3].Field.Names[0].Name).To(Equal("DoAnArray"))
				// Expect(model.Methods[3].Imports).To(HaveLen(1))
			})

			// It("does not return an error", func() {
			// 	Expect(err).ToNot(HaveOccurred())
			// })
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
