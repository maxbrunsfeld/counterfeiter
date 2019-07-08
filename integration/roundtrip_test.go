package integration_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func runTests(useGopath bool, t *testing.T, when spec.G, it spec.S) {
	log.SetOutput(ioutil.Discard) // Comment this out to see verbose log output
	log.SetFlags(log.Llongfile)
	var (
		baseDir             string
		relativeDir         string
		originalGopath      string
		originalGo111module string
		testDir             string
		copyDirFunc         func()
		copyFileFunc        func(name string)
		initModuleFunc      func()
		writeToTestData     bool
	)

	name := "working with a GOPATH"
	if !useGopath {
		name = "working with a module"
	}

	it.Before(func() {
		RegisterTestingT(t)
		originalGo111module = os.Getenv("GO111MODULE")
		if useGopath {
			os.Setenv("GO111MODULE", "off")
		} else {
			os.Setenv("GO111MODULE", "on")
		}
		originalGopath = os.Getenv("GOPATH")
		var err error
		testDir, err = ioutil.TempDir("", "counterfeiter-integration")
		Expect(err).NotTo(HaveOccurred())
		if useGopath {
			os.Setenv("GOPATH", testDir)
		} else {
			os.Unsetenv("GOPATH")
		}

		if useGopath {
			baseDir = filepath.Join(testDir, "src", "github.com", "maxbrunsfeld", "counterfeiter", "v6", "fixtures")
		} else {
			baseDir = testDir
		}

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
		initModuleFunc = func() {
			copyFileFunc("blank.go")
			err := ioutil.WriteFile(filepath.Join(baseDir, "go.mod"), []byte("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures"), 0755)
			Expect(err).ToNot(HaveOccurred())
		}
		// Set this to true to write the output of tests to the testdata/output
		// directory 🙃 happy debugging!
		// writeToTestData = true
	})

	it.After(func() {
		if originalGo111module != "" {
			os.Setenv("GO111MODULE", originalGo111module)
		} else {
			os.Unsetenv("GO111MODULE")
		}
		if originalGopath != "" {
			os.Setenv("GOPATH", originalGopath)
		} else {
			os.Unsetenv("GOPATH")
		}
		if baseDir == "" {
			return
		}
		err := os.RemoveAll(testDir)
		Expect(err).ToNot(HaveOccurred())
	})

	when("generating a fake for stdlib interfaces", func() {
		it("succeeds", func() {
			initModuleFunc()
			cache := &generator.FakeCache{}
			f, err := generator.NewFake(generator.InterfaceOrFunction, "WriteCloser", "io", "FakeWriteCloser", "custom", baseDir, cache)
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true) // Flip to false to see output if goimports fails
			Expect(err).NotTo(HaveOccurred())
			if writeToTestData {
				WriteOutput(b, filepath.Join("testdata", "output", "write_closer", "actual.go"))
			}
			WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_write_closer.go"))
			RunBuild(baseDir)
			b2, err := ioutil.ReadFile(filepath.Join("testdata", "expected_fake_writecloser.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b2)).To(Equal(string(b)))
		})
	})

	when("generating an interface for a package", func() {
		it("succeeds", func() {
			initModuleFunc()
			cache := &generator.FakeCache{}
			f, err := generator.NewFake(generator.Package, "", "os", "Os", "custom", baseDir, cache)
			Expect(err).NotTo(HaveOccurred())
			b, err := f.Generate(true) // Flip to false to see output if goimports fails
			Expect(err).NotTo(HaveOccurred())
			if writeToTestData {
				WriteOutput(b, filepath.Join("testdata", "output", "package_mode", "actual.go"))
			}
			WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_os.go"))
			RunBuild(baseDir)
		})
	})

	when(name, func() {
		t := func(interfaceName string, filename string, subDir string, files ...string) {
			when("working with "+filename, func() {
				it.Before(func() {
					if subDir != "" {
						baseDir = filepath.Join(baseDir, subDir)
						relativeDir = filepath.Join(relativeDir, subDir)
					}
					log.Println(testDir)
					copyFileFunc(filename)
					for i := range files {
						copyFileFunc(files[i])
					}
				})

				it("succeeds", func() {
					suffix := strings.Replace(subDir, "\\", "/", -1)
					if suffix != "" {
						suffix = "/" + suffix
					}
					if !useGopath {
						WriteOutput([]byte(fmt.Sprintf("module github.com/maxbrunsfeld/counterfeiter/v6/fixtures%s\n", suffix)), filepath.Join(baseDir, "go.mod"))
					}
					cache := &generator.FakeCache{}
					f, err := generator.NewFake(generator.InterfaceOrFunction, interfaceName, fmt.Sprintf("github.com/maxbrunsfeld/counterfeiter/v6/fixtures%s", suffix), "Fake"+interfaceName, "fixturesfakes", baseDir, cache)
					Expect(err).NotTo(HaveOccurred())
					b, err := f.Generate(true) // Flip to false to see output if goimports fails
					Expect(err).NotTo(HaveOccurred())
					if writeToTestData {
						WriteOutput(b, filepath.Join("testdata", "output", strings.Replace(filename, ".go", "", -1), "actual.go"))
					}
					WriteOutput(b, filepath.Join(baseDir, "fixturesfakes", "fake_"+filename))
					RunBuild(baseDir)
				})
			})
		}
		t("SomethingElse", "compound_return.go", "")
		t("DotImports", "dot_imports.go", "")
		t("EmbedsInterfaces", "embeds_interfaces.go", "", filepath.Join("another_package", "types.go"))
		t("AliasedInterface", "aliased_interfaces.go", "", filepath.Join("another_package", "types.go"))
		t("HasImports", "has_imports.go", "")
		t("HasOtherTypes", "has_other_types.go", "", "other_types.go")
		t("HasVarArgs", "has_var_args.go", "")
		t("HasVarArgsWithLocalTypes", "has_var_args.go", "")
		t("ImportsGoHyphenPackage", "imports_go_hyphen_package.go", "", filepath.Join("go-hyphenpackage", "fixture.go"))
		t("FirstInterface", "multiple_interfaces.go", "")
		t("SecondInterface", "multiple_interfaces.go", "")
		t("InlineStructParams", "inline_struct_params.go", "")
		t("RequestFactory", "request_factory.go", "")
		t("ReusesArgTypes", "reuses_arg_types.go", "")
		t("SomethingWithForeignInterface", "something_remote.go", "", filepath.Join("aliased_package", "in_aliased_package.go"))
		t("Something", "something.go", "")
		t("SomethingFactory", "typed_function.go", "")
		t("SyncSomething", "interface.go", "sync")

		when("working with duplicate packages", func() {
			t := func(interfaceName string, offset string, fakePackageName string) {
				when("working with "+interfaceName, func() {
					it.Before(func() {
						if useGopath {
							baseDir = filepath.Join(baseDir, "dup_packages")
						}
						relativeDir = filepath.Join(relativeDir, "dup_packages")
						copyDirFunc()
					})

					it("succeeds", func() {
						pkgPath := "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/dup_packages"
						if offset != "" {
							pkgPath = pkgPath + "/" + offset
						}
						cache := &generator.FakeCache{}
						f, err := generator.NewFake(generator.InterfaceOrFunction, interfaceName, pkgPath, "Fake"+interfaceName, fakePackageName, baseDir, cache)
						Expect(err).NotTo(HaveOccurred())
						b, err := f.Generate(false) // Flip to false to see output if goimports fails
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
