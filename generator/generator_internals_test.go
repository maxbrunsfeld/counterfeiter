package generator

import (
	"io/ioutil"
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestGenerator(t *testing.T) {
	log.SetOutput(ioutil.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	spec.Run(t, "Generator", testGenerator, spec.Report(report.Terminal{}))
}

func testGenerator(t *testing.T, when spec.G, it spec.S) {
	var f *Fake

	it.Before(func() {
		RegisterTestingT(t)
		f = &Fake{}
	})

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
	})
}
