// +build windows

package arguments_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParsingArguments(t *testing.T) {
	spec.Run(t, "ParsingArguments (Windows)", testParsingArguments, spec.Report(report.Terminal{}))
}

func testParsingArguments(t *testing.T, when spec.G, it spec.S) {
	var (
		err error
		parsedArgs *arguments.ParsedArguments
		args []string
		workingDir string
		evaler arguments.Evaler
		stater arguments.Stater
	)

	justBefore := func() {
		parsedArgs, err = arguments.New(args, workingDir, evaler, stater)
	}

	it.Before(func() {
		RegisterTestingT(t)
		log.SetOutput(ioutil.Discard)
		workingDir = "C:\\Users\\test-user\\workspace"

		evaler = func(input string) (string, error) {
			return input, nil
		}
		stater = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

	when("when a single argument is provided with the output directory", func() {
		it.Before(func() {
			args = []string{"counterfeiter", "-o", "C:\\tmp\\foo", "io.Writer"}
			justBefore()
		})

		it("copies the provided output path into the result", func() {
			Expect(parsedArgs.OutputPath).To(Equal("C:\\tmp\\foo"))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	when("when two arguments are provided", func() {
		it.Before(func() {
			args = []string{"counterfeiter", "my\\specialpackage", "MySpecialInterface"}
			justBefore()
		})

		it("snake cases the filename for the output directory", func() {
			Expect(parsedArgs.OutputPath).To(Equal(
				filepath.Join(
					parsedArgs.SourcePackageDir,
					"specialpackagefakes",
					"fake_my_special_interface.go",
				),
			))
			Expect(err).NotTo(HaveOccurred())
		})

		when("the source directory", func() {
			it("should be an absolute path", func() {
				Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	when("when three arguments are provided", func() {
		when("and the third one is '-'", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "my/mypackage", "MySpecialInterface", "-"}
				justBefore()
			})

			it("snake cases the filename for the output directory", func() {
				Expect(parsedArgs.OutputPath).To(Equal(
					filepath.Join(
						parsedArgs.SourcePackageDir,
						"mypackagefakes",
						"fake_my_special_interface.go",
					),
				))
				Expect(err).NotTo(HaveOccurred())
			})

			when("the source directory", func() {
				it("should be an absolute path", func() {
					Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
}

func fakeFileInfo(filename string, isDir bool) os.FileInfo {
	return testFileInfo{name: filename, isDir: isDir}
}

type testFileInfo struct {
	name  string
	isDir bool
}

func (testFileInfo testFileInfo) Name() string {
	return testFileInfo.name
}

func (testFileInfo testFileInfo) IsDir() bool {
	return testFileInfo.isDir
}

func (testFileInfo testFileInfo) Size() int64 {
	return 0
}

func (testFileInfo testFileInfo) Mode() os.FileMode {
	return 0
}

func (testFileInfo testFileInfo) ModTime() time.Time {
	return time.Now()
}

func (testFileInfo testFileInfo) Sys() interface{} {
	return nil
}
