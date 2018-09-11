package integration_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/generator"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func WriteOutput(b []byte, file string) {
	os.MkdirAll(filepath.Dir(file), 0700)
	ioutil.WriteFile(file, b, 0600)
}

func RunBuild(baseDir string) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = baseDir
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		log.Println(stdout.String())
		log.Println(stderr.String())
	}
	Expect(err).NotTo(HaveOccurred())
}

func TestRoundTrip(t *testing.T) {
	log.SetOutput(ioutil.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	spec.Run(t, "RoundTrip", testRoundTrip, spec.Report(report.Terminal{}))
}

func testRoundTrip(t *testing.T, when spec.G, it spec.S) {
	var (
		baseDir         string
		relativeDir     string
		originalGopath  string
		gopath          string
		copyDirFunc     func()
		copyFileFunc    func(name string)
		writeToTestData bool
	)

	it.Before(func() {
		RegisterTestingT(t)
		originalGopath = os.Getenv("GOPATH")
		var err error
		gopath, err = ioutil.TempDir("", "counterfeiter-integration")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("GOPATH", gopath)
		baseDir = filepath.Join(gopath, "src", "github.com", "maxbrunsfeld", "counterfeiter", "fixtures")
		err = os.MkdirAll(baseDir, 0777)
		Expect(err).ToNot(HaveOccurred())
		relativeDir = filepath.Join("..", "fixtures")
		copyDirFunc = func() {
			err = os.MkdirAll(baseDir, 0777)
			Expect(err).ToNot(HaveOccurred())
			err = Copy(relativeDir, baseDir)
			Expect(err).ToNot(HaveOccurred())
		}
		copyFileFunc = func(name string) {
			dir := baseDir
			d := filepath.Dir(name)
			if d != "." {
				dir = filepath.Join(dir, d)
			}

			err = os.MkdirAll(dir, 0777)
			Expect(err).ToNot(HaveOccurred())
			b, err := ioutil.ReadFile(filepath.Join(relativeDir, name))
			Expect(err).ToNot(HaveOccurred())
			err = ioutil.WriteFile(filepath.Join(baseDir, name), b, 0755)
			Expect(err).ToNot(HaveOccurred())
		}
		// Set this to true to write the output of tests to the testdata/output
		// directory 🙃 happy debugging!
		writeToTestData = true
	})

	it.After(func() {
		if originalGopath != "" {
			os.Setenv("GOPATH", originalGopath)
		} else {
			os.Unsetenv("GOPATH")
		}
		if baseDir == "" {
			return
		}
		err := os.RemoveAll(gopath)
		Expect(err).ToNot(HaveOccurred())
	})

	when("generating a fake for stdlib interfaces", func() {
		it("succeeds", func() {
			f, err := generator.NewFake("WriteCloser", "io", "FakeWriteCloser", "custom")
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true)
			Expect(err).NotTo(HaveOccurred())
			if writeToTestData {
				WriteOutput(b, filepath.Join("testdata", "output", "write_closer", "actual.go"))
			}
			WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_write_closer.go"))
			RunBuild(baseDir)
			b2, err := ioutil.ReadFile(filepath.Join("testdata", "expected_fake_write_closer.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b2)).To(Equal(string(b)))
		})
	})

	when("working with a GOPATH", func() {
		t := func(interfaceName string, filename string, files ...string) {
			when("working with "+filename, func() {
				it.Before(func() {
					copyFileFunc(filename)
					for i := range files {
						copyFileFunc(files[i])
					}
				})

				it("succeeds", func() {
					f, err := generator.NewFake(interfaceName, "github.com/maxbrunsfeld/counterfeiter/fixtures", "Fake"+interfaceName, "fixturesfakes")
					Expect(err).NotTo(HaveOccurred())
					b, err := f.Generate(true)
					Expect(err).NotTo(HaveOccurred())
					if writeToTestData {
						WriteOutput(b, filepath.Join("testdata", "output", strings.Replace(filename, ".go", "", -1), "actual.go"))
					}
					WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_"+filename))
					RunBuild(baseDir)
				})
			})
		}
		t("SomethingElse", "compound_return.go")
		t("DotImports", "dot_imports.go")
		t("EmbedsInterfaces", "embeds_interfaces.go", filepath.Join("another_package", "types.go"))
		t("HasImports", "has_imports.go")
		t("HasOtherTypes", "has_other_types.go", "other_types.go")
		t("HasVarArgs", "has_var_args.go")
		t("HasVarArgsWithLocalTypes", "has_var_args.go")
		t("ImportsGoHyphenPackage", "imports_go_hyphen_package.go", filepath.Join("go-hyphenpackage", "fixture.go"))
		t("FirstInterface", "multiple_interfaces.go")
		t("SecondInterface", "multiple_interfaces.go")
		// t("RequestFactory", "request_factory.go")
		t("ReusesArgTypes", "reuses_arg_types.go")
		t("Something", "something.go")
		// t("SomethingFactory", "typed_function.go")

		when("working with duplicate packages", func() {
			t := func(interfaceName string, offset string, fakePackageName string) {
				when("working with "+interfaceName, func() {
					it.Before(func() {
						baseDir = filepath.Join(baseDir, "dup_packages")
						relativeDir = filepath.Join(relativeDir, "dup_packages")
						copyDirFunc()
					})

					it("succeeds", func() {
						pkgPath := "github.com/maxbrunsfeld/counterfeiter/fixtures/dup_packages"
						if offset != "" {
							pkgPath = pkgPath + "/" + offset
						}
						f, err := generator.NewFake(interfaceName, pkgPath, "Fake"+interfaceName, fakePackageName)
						Expect(err).NotTo(HaveOccurred())
						b, err := f.Generate(true)
						Expect(err).NotTo(HaveOccurred())
						if writeToTestData {
							WriteOutput(b, filepath.Join("testdata", "output", "dup_"+strings.ToLower(interfaceName), "actual.go"))
						}
						WriteOutput(b, filepath.Join(baseDir, offset, fakePackageName, "fake_"+strings.ToLower(interfaceName)+".go"))
						RunBuild(filepath.Join(baseDir, offset, fakePackageName))
					})
				})
			}

			t("MultiAB", "foo", "foofakes")
			t("AliasV1", "", "dup_packagesfakes")
		})
	})
}
