package arguments

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/maxbrunsfeld/counterfeiter/terminal/terminalfakes"

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

	var ui *terminalfakes.FakeUI

	var failWasCalled bool

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
		fail = func(_ string, _ ...interface{}) { failWasCalled = true }
		cwd = func() string {
			return "/home/test-user/workspace"
		}

		ui = new(terminalfakes.FakeUI)

		symlinkEvaler = func(input string) (string, error) {
			return input, nil
		}
		fileStatReader = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

	Describe("when the -p flag is provided", func() {
		BeforeEach(func() {
			args = []string{"os"}
			*packageFlag = true
		})

		It("doesn't parse extraneous arguments", func() {
			Expect(parsedArgs.InterfaceName).To(Equal(""))
			Expect(parsedArgs.FakeImplName).To(Equal(""))
		})

		Context("given a stdlib package", func() {
			It("sets arguments as expected", func() {
				Expect(parsedArgs.SourcePackageDir).To(Equal(path.Join(runtime.GOROOT(), "src/os")))
				Expect(parsedArgs.OutputPath).To(Equal(path.Join(cwd(), "osshim")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("osshim"))
			})
		})

		Context("given a relative path to a path to a package", func() {})
	})

	Describe("when a single argument is provided", func() {
		BeforeEach(func() {
			args = []string{"someonesinterfaces.AnInterface"}
		})

		It("indicates to not print to stdout", func() {
			Expect(parsedArgs.PrintToStdOut).To(BeFalse())
		})

		It("provides a name for the fake implementing the interface", func() {
			Expect(parsedArgs.FakeImplName).To(Equal("FakeAnInterface"))
		})

		It("provides a path for the interface source", func() {
			Expect(parsedArgs.ImportPath).To(Equal("someonesinterfaces"))
		})

		It("treats the last segment as the interface to counterfeit", func() {
			Expect(parsedArgs.InterfaceName).To(Equal("AnInterface"))
		})

		It("snake cases the filename for the output directory", func() {
			Expect(parsedArgs.OutputPath).To(Equal(
				filepath.Join(
					cwd(),
					"workspacefakes",
					"fake_an_interface.go",
				),
			))
		})
	})

	Describe("when two arguments are provided", func() {
		BeforeEach(func() {
			args = []string{"my/my5package", "MySpecialInterface"}
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
					"my5packagefakes",
					"fake_my_special_interface.go",
				),
			))
		})

		It("specifies the destination package name", func() {
			Expect(parsedArgs.DestinationPackageName).To(Equal("my5packagefakes"))
		})

		Context("when the interface is unexported", func() {
			BeforeEach(func() {
				args = []string{"my/mypackage", "mySpecialInterface"}
			})

			It("fixes up the fake name to be TitleCase", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
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
				BeforeEach(func() {
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
				args = []string{"my/mypackage", "MySpecialInterface", "-"}
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
						"mypackagefakes",
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
				args = []string{"my/mypackage", "MySpecialInterface", "WHOOPS"}
			})

			It("indicates to not print to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})
		})
	})

	Context("when the output dir contains characters inappropriate for a package name", func() {
		BeforeEach(func() {
			args = []string{"@my-special-package[]{}", "MySpecialInterface"}
		})

		It("should choose a valid package name", func() {
			Expect(parsedArgs.DestinationPackageName).To(Equal("myspecialpackagefakes"))
		})
	})

	Context("when the output dir contains underscores in package name", func() {
		BeforeEach(func() {
			args = []string{"fake_command_runner", "MySpecialInterface"}
		})

		It("should ensure underscores are in the package name", func() {
			Expect(parsedArgs.DestinationPackageName).To(Equal("fake_command_runnerfakes"))
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
