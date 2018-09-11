package generator_test

import (
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestRealGenerator(t *testing.T) {
	//spec.Run(t, "RealGenerator", testRealGenerator, spec.Report(report.Terminal{}))
}

func testRealGenerator(t *testing.T, when spec.G, it spec.S) {
	var subject generator.ShimGenerator

	it.Before(func() {
		RegisterTestingT(t)
		model, _ := locator.GetInterfaceFromFilePath("Os", "../fixtures/packagegen/package_gen.go")

		subject = generator.ShimGenerator{
			Model:         *model,
			SourcePackage: "os",
			StructName:    "OsShim",
			PackageName:   "osshim",
		}
	})

	when("generating a shim for a package", func() {
		var fileContents string
		var err error

		it.Before(func() {
			fileContents, err = subject.GenerateReal()
		})

		it("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		it("should have a package", func() {
			Expect(fileContents).To(ContainSubstring("package osshim"))
		})

		it("should define an empty shim struct", func() {
			Expect(fileContents).To(ContainSubstring("type OsShim struct{}"))
		})

		it("should attach a real function to the empty struct", func() {
			Expect(fileContents).To(ContainSubstring("func (sh *OsShim) MkdirAll("))
		})

		it("should attach a real function parameters", func() {
			Expect(fileContents).To(ContainSubstring("MkdirAll(path string, perm os.FileMode) error {"))
		})

		it("should not return when it should not return", func() {
			Expect(fileContents).To(ContainSubstring(`
func (sh *OsShim) Exit(code int) {
	os.Exit(code)
}`))
		})

		it("it should handle variadics without exploding", func() {
			Expect(fileContents).To(ContainSubstring("Fictional(lol ...string) {"))
			Expect(fileContents).To(ContainSubstring("os.Fictional(lol...)"))
		})

		it("should attach a real function to the empty struct", func() {
			Expect(fileContents).To(ContainSubstring("return os.MkdirAll(path, perm)"))
		})
	})
}
