//go:build !windows

package arguments_test

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParsingArguments(t *testing.T) {
	spec.Run(t, "ParsingArguments", testParsingArguments, spec.Report(report.Terminal{}))
}

func testParsingArguments(t *testing.T, when spec.G, it spec.S) {
	var (
		err        error
		parsedArgs *arguments.ParsedArguments
		args       []string
		workingDir string
		evaler     arguments.Evaler
		stater     arguments.Stater
	)

	justBefore := func() {
		parsedArgs, err = arguments.New(args, workingDir, evaler, stater)
	}

	it.Before(func() {
		RegisterTestingT(t)
		log.SetOutput(io.Discard)
		workingDir = "/home/test-user/workspace"

		evaler = func(input string) (string, error) {
			return input, nil
		}
		stater = func(filename string) (os.FileInfo, error) {
			return fakeFileInfo(filename, true), nil
		}
	})

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

	when("when a single argument is provided followed by '-'", func() {
		it.Before(func() {
			args = []string{"counterfeiter", "io.WriteCloser", "-"}
			justBefore()
		})

		it("still treats the argument as a fully qualified interface", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedArgs.PackagePath).To(Equal("io"))
			Expect(parsedArgs.InterfaceName).To(Equal("WriteCloser"))
			Expect(parsedArgs.FakeImplName).To(Equal("FakeWriteCloser"))
			Expect(parsedArgs.SourcePackageDir).To(BeEmpty())
		})

		it("indicates that the fake should be printed to stdout", func() {
			Expect(parsedArgs.PrintToStdOut).To(BeTrue())
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

	when("when '-fake-name-template' is used", func() {
		it.Before(func() {
			args = []string{"counterfeiter", "-fake-name-template", "The{{.TargetName}}Imposter", "my/mypackage", "mySpecialInterface"}
			justBefore()
		})

		it("names the fake by evaluating the template against the target name", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedArgs.FakeImplName).To(Equal("TheMySpecialInterfaceImposter"))
		})

		it("snake cases the templated name for the output file", func() {
			Expect(parsedArgs.OutputPath).To(Equal(
				filepath.Join(
					parsedArgs.SourcePackageDir,
					"mypackagefakes",
					"the_my_special_interface_imposter.go",
				),
			))
		})

		when("'-fake-name' is also given", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-fake-name", "Explicit", "-fake-name-template", "The{{.TargetName}}Imposter", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("prefers the explicit fake name", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.FakeImplName).To(Equal("Explicit"))
			})
		})

		when("the template references an unknown variable", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-fake-name-template", "{{.Nope}}", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("returns an error naming the flag", func() {
				Expect(err).To(MatchError(ContainSubstring("-fake-name-template")))
			})
		})

		when("the template does not parse", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-fake-name-template", "{{.TargetName", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("returns an error naming the flag", func() {
				Expect(err).To(MatchError(ContainSubstring("-fake-name-template")))
			})
		})
	})

	when("when '-test' is used", func() {
		when("with a source path and an interface", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-test", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("records that the fake goes into the external test package", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.TestPackage).To(BeTrue())
			})

			it("writes a test file into the working directory, next to the tests that use it", func() {
				Expect(parsedArgs.SourcePackageDir).To(Equal(filepath.Join(workingDir, "my", "mypackage")))
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "fake_my_special_interface_test.go")))
			})

			it("names the destination package after the working directory with a _test suffix", func() {
				Expect(parsedArgs.DestinationPackageName).To(Equal("workspace_test"))
			})
		})

		when("with a fully qualified interface", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-test", "io.WriteCloser"}
				justBefore()
			})

			it("writes a test file into the working directory", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "fake_write_closer_test.go")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("workspace_test"))
			})
		})

		when("with an output directory", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-test", "-o", "other", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("writes a test file into that directory", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "other", "fake_my_special_interface_test.go")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("other_test"))
			})
		})

		when("with an output file that is a test file", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-test", "-o", "thing_test.go", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("keeps the file name", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "thing_test.go")))
			})
		})

		when("with an output file that is not a test file", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-test", "-o", "thing.go", "my/mypackage", "MySpecialInterface"}
				justBefore()
			})

			it("returns an error, since Go only compiles a _test package from test files", func() {
				Expect(err).To(MatchError(And(ContainSubstring("-test"), ContainSubstring("_test.go"))))
			})
		})

		when("together with '-p'", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-p", "-test", "os"}
				justBefore()
			})

			it("returns an error", func() {
				Expect(err).To(MatchError(ContainSubstring("-test")))
			})
		})
	})

	when("when '-generate' is used", func() {
		it.Before(func() {
			args = []string{"counterfeiter", "-generate", "-o", "fake", "-fake-name-template", "{{.TargetName}}", "-header", "generic.txt", "-q", "-test"}
			justBefore()
		})

		it("keeps the flags as defaults for the directives", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedArgs.GenerateMode).To(BeTrue())
			Expect(parsedArgs.OutputPath).To(Equal("fake"))
			Expect(parsedArgs.FakeNameTemplate).To(Equal("{{.TargetName}}"))
			Expect(parsedArgs.HeaderFile).To(Equal("generic.txt"))
			Expect(parsedArgs.Quiet).To(BeTrue())
			Expect(parsedArgs.TestPackage).To(BeTrue())
		})

		when("the fake name template does not parse", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-generate", "-fake-name-template", "{{.TargetName"}
				justBefore()
			})

			it("returns an error naming the flag", func() {
				Expect(err).To(MatchError(ContainSubstring("-fake-name-template")))
			})
		})
	})

	when("when defaults from a '-generate' invocation are supplied", func() {
		var defaults *arguments.ParsedArguments

		it.Before(func() {
			defaults, err = arguments.New(
				[]string{"counterfeiter", "-generate", "-o", "fake", "-fake-name-template", "{{.TargetName}}", "-header", "generic.txt", "-q"},
				workingDir, evaler, stater,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		justBeforeWithDefaults := func() {
			parsedArgs, err = arguments.New(args, workingDir, evaler, stater, arguments.WithDefaults(defaults))
		}

		when("the directive has no flags of its own", func() {
			it.Before(func() {
				args = []string{"counterfeiter", ".", "mySpecialInterface"}
				justBeforeWithDefaults()
			})

			it("uses the defaults", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.FakeImplName).To(Equal("MySpecialInterface"))
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "fake", "my_special_interface.go")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("fake"))
				Expect(parsedArgs.HeaderFile).To(Equal("generic.txt"))
				Expect(parsedArgs.Quiet).To(BeTrue())
			})
		})

		when("the directive has flags of its own", func() {
			it.Before(func() {
				args = []string{"counterfeiter", "-o", "other", "-fake-name", "Other", "-header", "specific.txt", ".", "mySpecialInterface"}
				justBeforeWithDefaults()
			})

			it("prefers the directive's flags", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.FakeImplName).To(Equal("Other"))
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "other", "other.go")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("other"))
				Expect(parsedArgs.HeaderFile).To(Equal("specific.txt"))
			})
		})

		when("the defaults ask for the external test package", func() {
			it.Before(func() {
				defaults, err = arguments.New([]string{"counterfeiter", "-generate", "-test"}, workingDir, evaler, stater)
				Expect(err).NotTo(HaveOccurred())
				args = []string{"counterfeiter", ".", "MySpecialInterface"}
				justBeforeWithDefaults()
			})

			it("generates the directive's fake into the external test package", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.TestPackage).To(BeTrue())
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "fake_my_special_interface_test.go")))
				Expect(parsedArgs.DestinationPackageName).To(Equal("workspace_test"))
			})
		})

		when("the defaults are nil", func() {
			it.Before(func() {
				defaults = nil
				args = []string{"counterfeiter", ".", "mySpecialInterface"}
				justBeforeWithDefaults()
			})

			it("behaves as if no defaults were supplied", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedArgs.FakeImplName).To(Equal("FakeMySpecialInterface"))
				Expect(parsedArgs.OutputPath).To(Equal(filepath.Join(workingDir, "workspacefakes", "fake_my_special_interface.go")))
				Expect(parsedArgs.HeaderFile).To(BeEmpty())
				Expect(parsedArgs.Quiet).To(BeFalse())
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
