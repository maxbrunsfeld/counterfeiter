// +build windows

package arguments

import (
	"os"
	"path/filepath"
	"time"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/terminal/terminalfakes"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParsingArguments(t *testing.T) {
	spec.Run(t, "ParsingArguments (Windows)", testParsingArguments, spec.Report(report.Terminal{}))
}

func testParsingArguments(t *testing.T, when spec.G, it spec.S) {
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

	justBefore := func() {
		subject = NewArgumentParser(
			fail,
			cwd,
			symlinkEvaler,
			fileStatReader,
			ui,
		)
		parsedArgs = subject.ParseArguments(args...)
	}

	it.Before(func() {
		RegisterTestingT(t)
		*packageFlag = false
		failWasCalled = false
		failWasCalledWithMessage = ""
		failWasCalledWithArgs = []interface{}{}
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

	when("when a single argument is provided with the output directory", func() {
		it.Before(func() {
			*outputPathFlag = "C:\\tmp\\foo"
			args = []string{"io.Writer"}
			justBefore()
		})

		it("copies the provided output path into the result", func() {
			Expect(parsedArgs.OutputPath).To(Equal("C:\\tmp\\foo"))
			Expect(failWasCalled).To(BeFalse())
		})
	})

	when("when two arguments are provided", func() {
		it.Before(func() {
			args = []string{"my\\specialpackage", "MySpecialInterface"}
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
			Expect(failWasCalled).To(BeFalse())
		})

		when("the source directory", func() {
			it("should be an absolute path", func() {
				Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
				Expect(failWasCalled).To(BeFalse())
			})
		})
	})

	when("when three arguments are provided", func() {
		when("and the third one is '-'", func() {
			it.Before(func() {
				args = []string{"my/mypackage", "MySpecialInterface", "-"}
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
				Expect(failWasCalled).To(BeFalse())
			})

			when("the source directory", func() {
				it("should be an absolute path", func() {
					Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
					Expect(failWasCalled).To(BeFalse())
					Expect(failWasCalledWithMessage).To(BeZero())
					Expect(failWasCalledWithArgs).To(HaveLen(0))
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
