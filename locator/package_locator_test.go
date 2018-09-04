package locator_test

import (
	"go/ast"
	"path"
	"runtime"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/locator"
	"github.com/maxbrunsfeld/counterfeiter/model"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestPackageLocator(t *testing.T) {
	spec.Run(t, "PackageLocator", testPackageLocator, spec.Report(report.Terminal{}))
}

func testPackageLocator(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("finding functions in a package", func() {
		var packageToInterfacify *model.PackageToInterfacify
		var err error

		it.Before(func() {
			packageToInterfacify, err = locator.GetFunctionsFromDirectory("os", path.Join(runtime.GOROOT(), "src/os"))
		})

		when("when the package exists", func() {
			it("should have the correct name", func() {
				Expect(packageToInterfacify.Name).To(Equal("os"))
			})

			it("should have the correct import path", func() {
				Expect(packageToInterfacify.ImportPath).To(HavePrefix(runtime.GOROOT()))
				Expect(packageToInterfacify.ImportPath).To(HaveSuffix("src/os"))
			})

			it("should have the correct methods", func() {
				Expect(len(packageToInterfacify.Funcs)).To(BeNumerically(">", 0))

				var findProcessMethod *ast.FuncDecl
				for _, funcNode := range packageToInterfacify.Funcs {
					if funcNode.Name.Name == "FindProcess" {
						findProcessMethod = funcNode
						break
					}
				}

				Expect(findProcessMethod).ToNot(BeNil())
			})

			it("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		when("one package, 2 files, duplicate imports", func() {})

		when("finding an exported function with renamed imports", func() {})

		when("finding an interface with dot imports", func() {})
	})
}
