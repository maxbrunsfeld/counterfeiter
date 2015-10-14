package arguments_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/maxbrunsfeld/counterfeiter/arguments"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parsing arguments", func() {
	var subject ArgumentParser
	var parsedArgs ParsedArguments
	var args []string

	var fail FailHandler
	var cwd CurrentWorkingDir
	var symlinkEvaler SymlinkEvaler
	var fileStatReader FileStatReader

	JustBeforeEach(func() {
		subject = NewArgumentParser(fail, cwd, symlinkEvaler, fileStatReader)
		parsedArgs = subject.ParseArguments(args...)
	})

	BeforeEach(func() {
		fail = func(_ string, _ ...interface{}) {}
		cwd = func() string {
			return "/home/test-user/workspace"
		}

		symlinkEvaler = func(input string) (string, error) {
			return input, nil
		}
		fileStatReader = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

	Describe("when two arguments are provided", func() {
		BeforeEach(func() {
			args = []string{"some/path", "MySpecialInterface"}
		})

		It("indicates to not print to stdout", func() {
			Expect(parsedArgs.PrintToStdOut).To(BeFalse())
		})

		It("provides a name for the fake implementing the interface", func() {
			Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
		})

		It("treats the second argument as the interface to counterfeit", func() {
			Expect(parsedArgs.InterfaceName).To(Equal("MySpecialInterface"))
		})

		It("snake cases the filename for the output directory", func() {
			Expect(parsedArgs.OutputPath).To(Equal(
				filepath.Join(
					parsedArgs.SourcePackageDir,
					"fakes",
					"fake_my_special_interface.go",
				),
			))
		})

		Describe("the source directory", func() {
			It("should be an absolute path", func() {
				Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
			})

			Context("when the first arg is a path to a file", func() {
				BeforeEach(func() {
					fileStatReader = func(filename string) (os.FileInfo, error) {
						return fakeFileInfo(filename, false), nil
					}
				})

				It("should be the directory containing the file", func() {
					Expect(parsedArgs.SourcePackageDir).ToNot(ContainSubstring("something.go"))
				})
			})

			Context("when the file stat cannot be read", func() {
				var failWasCalled bool

				BeforeEach(func() {
					fail = func(_ string, _ ...interface{}) { failWasCalled = true }
					fileStatReader = func(_ string) (os.FileInfo, error) {
						return fakeFileInfo("", false), errors.New("submarine-shoutout")
					}
				})

				It("should call its fail handler", func() {
					Expect(failWasCalled).To(BeTrue())
				})
			})
		})
	})

	Describe("when three arguments are provided", func() {
		Context("and the third one is '-'", func() {
			BeforeEach(func() {
				args = []string{"some/path", "MySpecialInterface", "-"}
			})

			It("treats the second argument as the interface to counterfeit", func() {
				Expect(parsedArgs.InterfaceName).To(Equal("MySpecialInterface"))
			})

			It("provides a name for the fake implementing the interface", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
			})

			It("indicates that the fake should be printed to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeTrue())
			})

			It("snake cases the filename for the output directory", func() {
				Expect(parsedArgs.OutputPath).To(Equal(
					filepath.Join(
						parsedArgs.SourcePackageDir,
						"fakes",
						"fake_my_special_interface.go",
					),
				))
			})

			Describe("the source directory", func() {
				It("should be an absolute path", func() {
					Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
				})

				Context("when the first arg is a path to a file", func() {
					BeforeEach(func() {
						fileStatReader = func(filename string) (os.FileInfo, error) {
							return fakeFileInfo(filename, false), nil
						}
					})

					It("should be the directory containing the file", func() {
						Expect(parsedArgs.SourcePackageDir).ToNot(ContainSubstring("something.go"))
					})
				})
			})
		})

		Context("and the third one is some random input", func() {
			BeforeEach(func() {
				args = []string{"some/path", "MySpecialInterface", "WHOOPS"}
			})

			It("indicates to not print to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})
		})
	})
})

func TestCounterfeiterCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Argument Parser Suite")
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
