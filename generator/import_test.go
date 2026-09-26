package generator

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestImport(t *testing.T) {
	spec.Run(t, "Import", testImport, spec.Report(report.Terminal{}))
}

func testImport(t *testing.T, when spec.G, it spec.S) {
	g := NewWithT(t)

	when("printing an import", func() {
		it("leaves out the alias for a stdlib package", func() {
			g.Expect(Import{Alias: "os", PkgPath: "os"}.String()).To(Equal(`"os"`))
		})

		it("leaves out the alias when it matches the base name", func() {
			g.Expect(Import{Alias: "foo", PkgPath: "example.com/goo/foo"}.String()).To(Equal(`"example.com/goo/foo"`))
		})

		it("prints a custom alias", func() {
			g.Expect(Import{Alias: "thinga", PkgPath: "example.com/go-thing"}.String()).To(Equal(`thinga "example.com/go-thing"`))
		})
	})
}
