package generator_test

import (
	"path/filepath"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestInterfaceGenerator(t *testing.T) {
	spec.Run(t, "InterfaceGenerator", testInterfaceGenerator, spec.Report(report.Terminal{}))
}

func testInterfaceGenerator(t *testing.T, when spec.G, it spec.S) {
	var (
		subject          generator.InterfaceGenerator
		fakeFileContents string
		err              error
	)

	it.Before(func() {
		RegisterTestingT(t)
		fixturePath := filepath.Join("..", "fixtures", "packagegen", "apackage")
		model, _ := locator.GetFunctionsFromDirectory("ostest", fixturePath)

		subject = generator.InterfaceGenerator{
			Model:                  model,
			Package:                fixturePath,
			DestinationPackageName: "osshim",
			DestinationInterface:   "Os",
		}
		fakeFileContents, err = subject.GenerateInterface()
	})

	it("should not fail", func() {
		Expect(err).ToNot(HaveOccurred())
	})

	it("correctly names the package", func() {
		Expect(fakeFileContents).To(ContainSubstring("package osshim"))
	})

	it("correctly names the interface", func() {
		Expect(fakeFileContents).To(ContainSubstring("type Os interface {"))
	})

	it("should produce a correct function prototype", func() {
		Expect(fakeFileContents).To(ContainSubstring("MkdirAll(path string, perm os.FileMode) error"))
	})

	it("should import the appropriate packages", func() {
		Expect(fakeFileContents).To(ContainSubstring(`"os"`))
		Expect(fakeFileContents).To(ContainSubstring(`"time"`))
	})

	it("should produce the correct file contents", func() {
		Expect(fakeFileContents).To(ContainSubstring(expectedOutput))
	})

	it("should produce a go generate comment", func() {
		Expect(fakeFileContents).To(ContainSubstring("//go:generate counterfeiter -o ostest_fake/fake_ostest.go . Os"))
	})
}

const expectedOutput string = `
type Os interface {
	FindProcess(pid int) (*os.Process, error)
	Hostname() (name string, err error)
	Expand(s string, mapping func(string) string) string
	Clearenv()
	Environ() []string
	Chtimes(name string, atime time.Time, mtime time.Time) error
	MkdirAll(path string, perm os.FileMode) error
	Exit(code int)
	Fictional(lol ...string)
}`
