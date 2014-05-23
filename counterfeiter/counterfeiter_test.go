package counterfeiter_test

import (
	. "github.com/maxbrunsfeld/counterfeiter/counterfeiter"

	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Counterfeiter", func() {
	Describe("Generate", func() {
		It("generates a fake implementation of a struct", func() {
			dir, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			code, err := Generate(
				filepath.Join(dir, "../fixtures/interfaces"),
				"SomeInterface",
				"fakes",
				"FakeSomeInterface",
			)
			Expect(err).NotTo(HaveOccurred())

			expectedCode, err := ioutil.ReadFile(filepath.Join(dir, "../fixtures/fakes/fake_some_interface.go"))
			Expect(err).NotTo(HaveOccurred())

			Expect(normalizeWhitespace(code)).To(Equal(normalizeWhitespace(string(expectedCode))))
		})
	})
})

func normalizeWhitespace(input string) string {
	var spaceRegexp = regexp.MustCompile("([^\t \n])[\t ]+")
	input = spaceRegexp.ReplaceAllString(input, "$1 ")
	return strings.TrimSpace(input)
}
