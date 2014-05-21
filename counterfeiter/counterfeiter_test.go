package counterfeiter_test

import (
	. "github.com/maxbrunsfeld/counterfeiter/counterfeiter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
  "regexp"
	"path/filepath"
	"strings"
)

var _ = Describe("Counterfeiter", func() {
	Describe("Generate", func() {
    table := [][]string{
      {
        "handles full package paths",

				"github.com/maxbrunsfeld/counterfeiter/fixtures/interfaces",
				"SomeInterface",
        "fakes",
        "fake_some_interface.go",
      },
      {
        "handles relative package paths",

				"./fixtures/interfaces",
				"SomeInterface",
        "fakes",
        "fake_some_interface.go",
      },
    }

    for _, row := range table {
      description := row[0]
      packagePath := row[1]
      interfaceName := row[2]
      fakePackageName := row[3]
      expectedOutputPath := row[4]

      It(description, func() {
        code, err := Generate(packagePath, interfaceName, fakePackageName)
        Expect(err).NotTo(HaveOccurred())

        fakeFile, err := os.Open(fixturePath(expectedOutputPath))
        Expect(err).NotTo(HaveOccurred())

        expectedCode, err := ioutil.ReadAll(fakeFile)
        Expect(err).NotTo(HaveOccurred())

        Expect(normalizeWhitespace(code)).To(Equal(normalizeWhitespace(string(expectedCode))))
      })
    }
	})
})

func fixturePath(basename string) string {
	gopath := os.Getenv("GOPATH")
	firstGopath := strings.Split(gopath, ":")[0]
	return filepath.Join(
		firstGopath,
		"src/github.com/maxbrunsfeld/counterfeiter/fixtures/fakes",
		basename,
	)
}

var tabRegexp = regexp.MustCompile("\t+")
var spaceRegexp = regexp.MustCompile(" +")

func normalizeWhitespace(input string) string {
  input = tabRegexp.ReplaceAllString(input, " ")
  input = spaceRegexp.ReplaceAllString(input, " ")
  return strings.TrimSpace(input)
}
