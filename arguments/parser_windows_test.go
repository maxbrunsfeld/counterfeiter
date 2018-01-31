// +build windows

package arguments

import (
	"os"
	"path/filepath"
	"time"

	"github.com/maxbrunsfeld/counterfeiter/terminal/terminalfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parsing arguments (for windows)", func() {
	var subject ArgumentParser
	var parsedArgs ParsedArguments
	var args []string

	var fail FailHandler
	var cwd CurrentWorkingDir
	var symlinkEvaler SymlinkEvaler
	var fileStatReader FileStatReader

	var ui *terminalfakes.FakeUI

	var failWasCalled bool
	var failWasCalledWithMessage string
	var failWasCalledWithArgs []interface{}

	JustBeforeEach(func() {
		subject = NewArgumentParser(
			fail,
			cwd,
			symlinkEvaler,
			fileStatReader,
			ui,
		)
		parsedArgs = subject.ParseArguments(args...)
	})

	BeforeEach(func() {
		*packageFlag = false
		failWasCalled = false
		*outputPathFlag = ""
		fail = func(msg string, args ...interface{}) {
			failWasCalled = true
			failWasCalledWithMessage = msg
			failWasCalledWithArgs = args

		}
		cwd = func() string {
			return "C:\\Users\\test-user\\workspace"
		}

		ui = new(terminalfakes.FakeUI)

		symlinkEvaler = func(input string) (string, error) {
			return input, nil
		}
		fileStatReader = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

	Describe("when a single argument is provided with the output directory", func() {
		BeforeEach(func() {
			*outputPathFlag = "C:\\tmp\\foo"
			args = []string{"io.Writer"}
		})

		It("copies the provided output path into the result", func() {
			Expect(parsedArgs.OutputPath).To(Equal("C:\\tmp\\foo"))
		})
	})

	Describe("when two arguments are provided", func() {
		BeforeEach(func() {
			args = []string{"my\\specialpackage", "MySpecialInterface"}
		})

		It("snake cases the filename for the output directory", func() {
			Expect(parsedArgs.OutputPath).To(Equal(
				filepath.Join(
					parsedArgs.SourcePackageDir,
					"specialpackagefakes",
					"fake_my_special_interface.go",
				),
			))
		})

		Describe("the source directory", func() {
			It("should be an absolute path", func() {
				Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
			})
		})
	})

	Describe("when three arguments are provided", func() {
		Context("and the third one is '-'", func() {
			BeforeEach(func() {
				args = []string{"my/mypackage", "MySpecialInterface", "-"}
			})

			It("snake cases the filename for the output directory", func() {
				Expect(parsedArgs.OutputPath).To(Equal(
					filepath.Join(
						parsedArgs.SourcePackageDir,
						"mypackagefakes",
						"fake_my_special_interface.go",
					),
				))
			})

			Describe("the source directory", func() {
				It("should be an absolute path", func() {
					Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
				})
			})
		})
	})
})

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
