//go:build !windows
// +build !windows

package arguments_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/onsi/gomega/gbytes"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParsingArguments(t *testing.T) {
	spec.Run(t, "ParseGenerateMode", testParseGenerateMode, spec.Report(report.Terminal{}))
	spec.Run(t, "ParsingArguments", testNew, spec.Report(report.Terminal{}))
}

func testParseGenerateMode(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
		log.SetOutput(ioutil.Discard)
	})

	when("-generate is given", func() {
		it("returns true", func() {
			generateMode, generateArgs, err := arguments.ParseGenerateMode([]string{"counterfeiter",
				"-generate",
				"-o", "fake",
				"-fake-name-template", `The{{.TargetName}}Imposter`,
				"-header", "my-header",
				"-q",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateMode).To(BeTrue())
			Expect(generateArgs).NotTo(BeNil())
			Expect(generateArgs.OutputPath).To(Equal("fake"))
			Expect(generateArgs.Header).To(Equal("my-header"))
			Expect(generateArgs.Quiet).To(BeTrue())

			Expect(generateArgs.FakeNameTemplate).NotTo(BeNil())
			nameWriter := gbytes.NewBuffer()
			Expect(
				generateArgs.FakeNameTemplate.Execute(nameWriter, struct{ TargetName string }{"MyType"}),
			).To(Succeed())
			Expect(string(nameWriter.Contents())).To(Equal("TheMyTypeImposter"))
		})

		when("the fake-name-template is invalid", func() {
			it("errors", func() {
				_, _, err := arguments.ParseGenerateMode([]string{"counterfeiter",
					"-generate",
					"-fake-name-template", `{{panic "boom"}}`,
				})
				Expect(err).To(MatchError(ContainSubstring("fake-name-template")))
			})
		})
	})

	when("-generate is not given", func() {
		it("returns false and nil", func() {
			generateMode, generateArgs, err := arguments.ParseGenerateMode([]string{"counterfeiter",
				"-o", "fake",
				"-fake-name", "Bob",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateMode).To(BeFalse())
			Expect(generateArgs).To(BeNil())
		})
	})

	when("no args are given", func() {
		it("returns false and nil", func() {
			generateMode, generateArgs, err := arguments.ParseGenerateMode([]string{"counterfeiter"})
			Expect(err).NotTo(HaveOccurred())
			Expect(generateMode).To(BeFalse())
			Expect(generateArgs).To(BeNil())
		})
	})

	when("unknown flags are given", func() {
		it("returns an error", func() {
			generateMode, generateArgs, err := arguments.ParseGenerateMode([]string{"counterfeiter", "-generate", "-no-such-flag"})
			Expect(err).To(HaveOccurred())
			Expect(generateMode).To(BeFalse())
			Expect(generateArgs).To(BeNil())
		})
	})

	when("-help is given", func() {
		it("returns an error", func() {
			generateMode, generateArgs, err := arguments.ParseGenerateMode([]string{"counterfeiter", "-generate", "-help"})
			Expect(err).To(HaveOccurred())
			Expect(generateMode).To(BeFalse())
			Expect(generateArgs).To(BeNil())

			generateMode, generateArgs, err = arguments.ParseGenerateMode([]string{"counterfeiter", "-help"})
			Expect(err).To(HaveOccurred())
			Expect(generateMode).To(BeFalse())
			Expect(generateArgs).To(BeNil())
		})
	})
}

func testNew(t *testing.T, when spec.G, it spec.S) {
	var (
		err        error
		parsedArgs *arguments.ParsedArguments
		args       []string
		workingDir string
		evaler     arguments.Evaler
		stater     arguments.Stater
	)

	it.Before(func() {
		RegisterTestingT(t)
		log.SetOutput(ioutil.Discard)
		workingDir = "/home/test-user/workspace"

		evaler = func(input string) (string, error) {
			return input, nil
		}
		stater = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

	when("not in generate mode", func() {
		justBefore := func() {
			parsedArgs, err = arguments.New(args, workingDir, nil, evaler, stater)
		}

		when("when the -p flag is provided", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-p", "os"}
				justBefore()
			})

			it("doesn't parse extraneous arguments", func() {
				Expect(err).To(Succeed())
				Expect(parsedArgs.GenerateInterfaceAndShimFromPackageDirectory).To(BeTrue())
				Expect(parsedArgs.InterfaceName).To(Equal(""))
				Expect(parsedArgs.FakeImplName).To(Equal("Os"))
			})

			when("given a stdlib package", func() {
				it("sets arguments as expected", func() {
					Expect(parsedArgs.SourcePackageDir).To(Equal("os"))
					Expect(parsedArgs.OutputPath).To(Equal(path.Join(workingDir, "osshim", "os.go")))
					Expect(parsedArgs.DestinationPackageName).To(Equal("osshim"))
				})
			})
		})

		when("when a single argument is provided", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
				justBefore()
			})

			it("sets PrintToStdOut to false", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})

			it("provides a name for the fake implementing the interface", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeAnInterface"))
			})

			it("provides a path for the interface source", func() {
				Expect(parsedArgs.PackagePath).To(Equal("someonesinterfaces"))
			})

			it("treats the last segment as the interface to counterfeit", func() {
				Expect(parsedArgs.InterfaceName).To(Equal("AnInterface"))
			})

			it("snake cases the filename for the output directory", func() {
				Expect(parsedArgs.OutputPath).To(Equal(
					filepath.Join(
						workingDir,
						"workspacefakes",
						"fake_an_interface.go",
					),
				))
			})
		})

		when("when a single argument is provided with the output directory", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-o", "/tmp/foo", "io.Writer"}
				justBefore()
			})

			it("indicates to not print to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})

			it("provides a name for the fake implementing the interface", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeWriter"))
			})

			it("provides a path for the interface source", func() {
				Expect(parsedArgs.PackagePath).To(Equal("io"))
			})

			it("treats the last segment as the interface to counterfeit", func() {
				Expect(parsedArgs.InterfaceName).To(Equal("Writer"))
			})

			it("copies the provided output path into the result", func() {
				Expect(parsedArgs.OutputPath).To(Equal("/tmp/foo/fake_writer.go"))
			})
		})

		when("when a single argument is provided with the output file", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-o", "/tmp/foo/fake_foo.go", "io.Writer"}
				justBefore()
			})

			it("indicates to not print to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})

			it("provides a name for the fake implementing the interface", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeWriter"))
			})

			it("provides a path for the interface source", func() {
				Expect(parsedArgs.PackagePath).To(Equal("io"))
			})

			it("treats the last segment as the interface to counterfeit", func() {
				Expect(parsedArgs.InterfaceName).To(Equal("Writer"))
			})

			it("copies the provided output path into the result", func() {
				Expect(parsedArgs.OutputPath).To(Equal("/tmp/foo/fake_foo.go"))
			})
		})

		when("when two arguments are provided", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "my/my5package", "MySpecialInterface"}
				justBefore()
			})

			it("indicates to not print to stdout", func() {
				Expect(parsedArgs.PrintToStdOut).To(BeFalse())
			})

			it("provides a name for the fake implementing the interface", func() {
				Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
			})

			it("treats the second argument as the interface to counterfeit", func() {
				Expect(parsedArgs.InterfaceName).To(Equal("MySpecialInterface"))
			})

			it("snake cases the filename for the output directory", func() {
				Expect(parsedArgs.OutputPath).To(Equal(
					filepath.Join(
						parsedArgs.SourcePackageDir,
						"my5packagefakes",
						"fake_my_special_interface.go",
					),
				))
			})

			it("specifies the destination package name", func() {
				Expect(parsedArgs.DestinationPackageName).To(Equal("my5packagefakes"))
			})

			when("when the interface is unexported", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "my/mypackage", "mySpecialInterface"}
					justBefore()
				})

				it("fixes up the fake name to be TitleCase", func() {
					Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
				})

				it("snake cases the filename for the output directory", func() {
					Expect(parsedArgs.OutputPath).To(Equal(
						filepath.Join(
							parsedArgs.SourcePackageDir,
							"mypackagefakes",
							"fake_my_special_interface.go",
						),
					))
				})
			})

			when("the source directory", func() {
				it("should be an absolute path", func() {
					Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
				})

				when("when the first arg is a path to a file", func() {
					it.Before(func() {
						stater = func(filename string) (os.FileInfo, error) {
							return fakeFileInfo(filename, false), nil
						}
						justBefore()
					})

					it("should be the directory containing the file", func() {
						Expect(parsedArgs.SourcePackageDir).ToNot(ContainSubstring("something.go"))
					})
				})

				when("when evaluating symlinks fails", func() {
					it.Before(func() {
						evaler = func(input string) (string, error) {
							return "", errors.New("aww shucks")
						}
						justBefore()
					})

					it("should return an error with a useful message", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal(fmt.Sprintf("No such file/directory/package [%s]: aww shucks", path.Join(workingDir, "my/my5package"))))
					})
				})

				when("when the file stat cannot be read", func() {
					it.Before(func() {
						stater = func(_ string) (os.FileInfo, error) {
							return fakeFileInfo("", false), errors.New("submarine-shoutout")
						}
						justBefore()
					})

					it("should return an error with a useful message", func() {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(Equal(fmt.Sprintf("No such file/directory/package [%s]: submarine-shoutout", path.Join(workingDir, "my/my5package"))))
					})
				})
			})
		})

		when("when the output dir contains characters inappropriate for a package name", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "@my-special-package[]{}", "MySpecialInterface"}
				justBefore()
			})

			it("should choose a valid package name", func() {
				Expect(parsedArgs.DestinationPackageName).To(Equal("myspecialpackagefakes"))
			})
		})

		when("when three arguments are provided", func() {
			when("and the third one is '-'", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "my/mypackage", "MySpecialInterface", "-"}
					justBefore()
				})

				it("treats the second argument as the interface to counterfeit", func() {
					Expect(parsedArgs.InterfaceName).To(Equal("MySpecialInterface"))
				})

				it("provides a name for the fake implementing the interface", func() {
					Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
				})

				it("indicates that the fake should be printed to stdout", func() {
					Expect(parsedArgs.PrintToStdOut).To(BeTrue())
				})

				it("snake cases the filename for the output directory", func() {
					Expect(parsedArgs.OutputPath).To(Equal(
						filepath.Join(
							parsedArgs.SourcePackageDir,
							"mypackagefakes",
							"fake_my_special_interface.go",
						),
					))
				})

				when("the source directory", func() {
					it("should be an absolute path", func() {
						Expect(filepath.IsAbs(parsedArgs.SourcePackageDir)).To(BeTrue())
					})

					when("when the first arg is a path to a file", func() {
						it.Before(func() {
							stater = func(filename string) (os.FileInfo, error) {
								return fakeFileInfo(filename, false), nil
							}
						})

						it("should be the directory containing the file", func() {
							Expect(parsedArgs.SourcePackageDir).ToNot(ContainSubstring("something.go"))
						})
					})
				})
			})

			when("and the third one is some random input", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "my/mypackage", "MySpecialInterface", "WHOOPS"}
					justBefore()
				})

				it("indicates to not print to stdout", func() {
					Expect(parsedArgs.PrintToStdOut).To(BeFalse())
				})
			})
		})

		when("when the output dir contains underscores in package name", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "fake_command_runner", "MySpecialInterface"}
				justBefore()
			})

			it("should ensure underscores are in the package name", func() {
				Expect(parsedArgs.DestinationPackageName).To(Equal("fake_command_runnerfakes"))
			})
		})

		when("when '-header' is used", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-header", "some/header/file", "some.interface"}
				justBefore()
			})

			it("sets the HeaderFile attribute on the parsedArgs struct", func() {
				Expect(parsedArgs.HeaderFile).To(Equal("some/header/file"))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	when("in generate mode", func() {
		var generateArgs arguments.GenerateArgs

		it.Before(func() {
			generateArgs = arguments.GenerateArgs{}
		})

		justBefore := func() {
			parsedArgs, err = arguments.New(args, workingDir, &generateArgs, evaler, stater)
		}

		when("generate was called with -o", func() {
			it.Before(func() {
				generateArgs.OutputPath = "generate-output"
			})

			when("the invocation specified -o also", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "-o", "output", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("chooses the invocation's output path", func() {
					Expect(parsedArgs.OutputPath).To(Equal(
						filepath.Join(
							workingDir,
							"output",
							"fake_an_interface.go",
						),
					))
				})
			})

			when("the invocation did not specify -o", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("chooses the generate call's output path", func() {
					Expect(parsedArgs.OutputPath).To(Equal(
						filepath.Join(
							workingDir,
							"generate-output",
							"fake_an_interface.go",
						),
					))
				})
			})
		})

		when("generate was called with -q", func() {
			it.Before(func() {
				generateArgs.Quiet = true
			})

			when("the invocation specified -q also", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "-q", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("is quiet", func() {
					Expect(parsedArgs.Quiet).To(BeTrue())
				})
			})

			when("the invocation did not specify -q", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("is quiet", func() {
					Expect(parsedArgs.Quiet).To(BeTrue())
				})
			})
		})

		when("generate was called with -header", func() {
			it.Before(func() {
				generateArgs.Header = "generate-header"
			})

			when("the invocation specified -header also", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "-header", "header", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("chooses the invocation's header", func() {
					Expect(parsedArgs.HeaderFile).To(Equal("header"))
				})
			})

			when("the invocation did not specify -header", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("chooses the generate call's header", func() {
					Expect(parsedArgs.HeaderFile).To(Equal("generate-header"))
				})
			})
		})

		when("generate was called with -fake-name-template", func() {
			it.Before(func() {
				generateArgs.FakeNameTemplate, err = template.New("test").Parse("The{{.TargetName}}Imposter")
				Expect(err).NotTo(HaveOccurred())
			})

			when("the invocation specified -fake-name", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "-fake-name", "FakestFake", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("chooses the invocation's fake name", func() {
					Expect(parsedArgs.FakeImplName).To(Equal("FakestFake"))
				})
			})

			when("the invocation did not specify -fake-name", func() {
				it.Before(func() {
					args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("uses the fake-name-template to generate a fake name", func() {
					Expect(parsedArgs.FakeImplName).To(Equal("TheAnInterfaceImposter"))
				})
			})

			when("the template is invalid", func() {
				it.Before(func() {
					generateArgs.FakeNameTemplate, err = template.New("test").Parse("{{.NoSuchField}}")
					Expect(err).NotTo(HaveOccurred())

					args = []string{"counterfeiter", "someonesinterfaces.AnInterface"}
					justBefore()
				})

				it("errors", func() {
					Expect(err).To(MatchError(ContainSubstring("fake-name-template")))
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
