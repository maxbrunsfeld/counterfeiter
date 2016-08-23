package generator_test

import (
	"github.com/maxbrunsfeld/counterfeiter/locator"

	. "github.com/maxbrunsfeld/counterfeiter/generator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generator", func() {
	var subject ShimGenerator

	BeforeEach(func() {
		model, _ := locator.GetInterfaceFromFilePath("Os", "../fixtures/packagegen/package_gen.go")

		subject = ShimGenerator{
			Model:         *model,
			SourcePackage: "os",
			StructName:    "OsShim",
			PackageName:   "osshim",
		}
	})

	Describe("generating a shim for a package", func() {
		var fileContents string
		var err error

		BeforeEach(func() {
			fileContents, err = subject.GenerateReal()
		})

		It("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have a package", func() {
			Expect(fileContents).To(ContainSubstring("package osshim"))
		})

		It("should define an empty shim struct", func() {
			Expect(fileContents).To(ContainSubstring("type OsShim struct{}"))
		})

		It("should attach a real function to the empty struct", func() {
			Expect(fileContents).To(ContainSubstring("func (sh *OsShim) MkdirAll("))
		})

		It("should attach a real function parameters", func() {
			Expect(fileContents).To(ContainSubstring("MkdirAll(path string, perm os.FileMode) error {"))
		})

		It("should not return when it should not return", func() {
			Expect(fileContents).To(ContainSubstring(`
func (sh *OsShim) Exit(code int) {
	os.Exit(code)
}`))
		})

		It("it should handle variadics without exploding", func() {
			Expect(fileContents).To(ContainSubstring("Fictional(lol ...string) {"))
			Expect(fileContents).To(ContainSubstring("os.Fictional(lol...)"))
		})

		It("should attach a real function to the empty struct", func() {
			Expect(fileContents).To(ContainSubstring("return os.MkdirAll(path, perm)"))
		})
	})
})
