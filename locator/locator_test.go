package locator_test

import (
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
				Expect(model.Methods).To(HaveLen(3))
				Expect(model.Methods[0].Names[0].Name).To(Equal("DoThings"))
				Expect(model.Methods[1].Names[0].Name).To(Equal("DoNothing"))
				Expect(model.Methods[2].Names[0].Name).To(Equal("DoASlice"))
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
				Expect(model.Methods[0].Names[0].Name).To(Equal("RequestFactory"))
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
})
