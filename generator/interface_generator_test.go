package generator_test

import (
	"path"

	. "github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Interface Generator", func() {
	var (
		subject          InterfaceGenerator
		fakeFileContents string
		err              error
	)

	BeforeEach(func() {
		model, _ := locator.GetFunctionsFromDirectory("ostest", path.Join("../fixtures/", "packagegen", "apackage"))

		subject = InterfaceGenerator{
			Model:                  model,
			Package:                path.Join("../fixtures/", "packagegen", "apackage"),
			DestinationPackageName: "osshim",
			DestinationInterface:   "Os",
		}
		fakeFileContents, err = subject.GenerateInterface()
	})

	It("should not fail", func() {
		Expect(err).ToNot(HaveOccurred())
	})

	It("correctly names the package", func() {
		Expect(fakeFileContents).To(ContainSubstring("package osshim"))
	})

	It("correctly names the interface", func() {
		Expect(fakeFileContents).To(ContainSubstring("type Os interface {"))
	})

	It("should produce a correct function prototype", func() {
		Expect(fakeFileContents).To(ContainSubstring("MkdirAll(path string, perm os.FileMode) error"))
	})

	It("should import the appropriate packages", func() {
		Expect(fakeFileContents).To(ContainSubstring(`"os"`))
		Expect(fakeFileContents).To(ContainSubstring(`"time"`))
	})

	It("should produce the correct file contents", func() {
		Expect(fakeFileContents).To(ContainSubstring(expectedOutput))
	})

	It("should produce a go generate comment", func() {
		Expect(fakeFileContents).To(ContainSubstring("//go:generate counterfeiter -o ostest_fake/fake_ostest.go . Os"))
	})
})

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
