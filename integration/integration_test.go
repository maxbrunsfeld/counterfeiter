package integration_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"testing"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var (
	tmpDir              string
	pathToCounterfeiter string
)

func TestMain(m *testing.M) {
	var err error
	pathToCounterfeiter, err = gexec.Build("github.com/maxbrunsfeld/counterfeiter")
	if err != nil {
		panic(err)
	}

	result := m.Run()
	gexec.CleanupBuildArtifacts()
	os.Exit(result)
}

func TestCounterfeiter(t *testing.T) {
	spec.Run(t, "Counterfeiter", testCounterfeiter, spec.Report(report.Terminal{}))
}

func testCounterfeiter(t *testing.T, when spec.G, it spec.S) {
	var pathToCLI string

	tmpPath := func(destination string) string {
		return filepath.Join(tmpDir, "src", destination)
	}

	copyIn := func(fixture string, directory string) {
		fixturesPath := filepath.Join(directory, "fixtures")
		err := os.MkdirAll(fixturesPath, 0777)
		Expect(err).ToNot(HaveOccurred())

		filepath.Walk(filepath.Join("..", "fixtures", fixture), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			base := filepath.Base(path)

			fileHandle, err := os.Open(path)
			Expect(err).ToNot(HaveOccurred())
			defer fileHandle.Close()

			dst, err := os.Create(filepath.Join(fixturesPath, base))
			Expect(err).ToNot(HaveOccurred())
			defer dst.Close()

			_, err = io.Copy(dst, fileHandle)
			Expect(err).ToNot(HaveOccurred())
			return nil
		})
	}

	startCounterfeiter := func(workingDir string, fixtureName string, otherArgs ...string) *gexec.Session {
		fakeGoPathDir := filepath.Dir(filepath.Dir(workingDir))
		absPath, _ := filepath.Abs(fakeGoPathDir)
		absPathWithSymlinks, _ := filepath.EvalSymlinks(absPath)

		fixturePath := filepath.Join("fixtures", fixtureName)
		args := append([]string{fixturePath}, otherArgs...)
		cmd := exec.Command(pathToCounterfeiter, args...)
		cmd.Dir = workingDir
		cmd.Env = []string{"GOPATH=" + absPathWithSymlinks}
		outWriter := &bytes.Buffer{}
		errWriter := &bytes.Buffer{}
		session, err := gexec.Start(cmd, outWriter, errWriter)
		Expect(err).ToNot(HaveOccurred())
		return session
	}

	startCounterfeiterWithoutFixture := func(workingDir string, args ...string) *gexec.Session {
		fakeGoPathDir := filepath.Dir(filepath.Dir(workingDir))
		absPath, _ := filepath.Abs(fakeGoPathDir)
		absPathWithSymlinks, _ := filepath.EvalSymlinks(absPath)

		cmd := exec.Command(pathToCounterfeiter, args...)
		cmd.Dir = workingDir
		cmd.Env = []string{
			"GOPATH=" + absPathWithSymlinks,
			"GOROOT=" + os.Getenv("GOROOT"),
		}
		outWriter := &bytes.Buffer{}
		errWriter := &bytes.Buffer{}
		session, err := gexec.Start(cmd, outWriter, errWriter)
		Expect(err).ToNot(HaveOccurred())
		return session
	}

	it.Before(func() {
		RegisterTestingT(t)
		var err error
		tmpDir, err = ioutil.TempDir("", "counterfeiter-integration")
		Expect(err).ToNot(HaveOccurred())
		pathToCLI = tmpPath("counterfeiter")
	})

	it.After(func() {
		if tmpDir != "" {
			err := os.RemoveAll(tmpDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	it("can generate a fake for a typed function", func() {
		copyIn("typed_function.go", pathToCLI)

		session := startCounterfeiter(pathToCLI, "typed_function.go", "SomethingFactory")

		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Wrote `FakeSomethingFactory"))

		generatedFakePath := filepath.Join(pathToCLI, "fixtures", "fixturesfakes", "fake_something_factory.go")
		Expect(generatedFakePath).To(BeARegularFile())

		expectedOutputPath := "../fixtures/expected_output/fake_something_factory.example"
		expectedContents, err := ioutil.ReadFile(expectedOutputPath)
		Expect(err).ToNot(HaveOccurred())

		actualContents, err := ioutil.ReadFile(generatedFakePath)
		Expect(err).ToNot(HaveOccurred())

		// assert file content matches what we expect
		Expect(string(actualContents)).To(Equal(string(expectedContents)))
	})

	it("can generate a fake for a internal interface, on a provided path", func() {
		os.MkdirAll(filepath.Join(pathToCLI, "src", "counterfeiter"), 0777)

		session := startCounterfeiterWithoutFixture(pathToCLI, "-o", pathToCLI+"/custom/fake_write_closer.go", "io.WriteCloser")
		Eventually(session).Should(gexec.Exit(0))
		Expect(session).To(gbytes.Say("Wrote `FakeWriteCloser`"))

		generatedFakePath := filepath.Join(pathToCLI, "custom", "fake_write_closer.go")
		Expect(generatedFakePath).To(BeARegularFile())

		expectedOutputPath := "../fixtures/expected_output/fake_write_closer.example"
		expectedContents, err := ioutil.ReadFile(expectedOutputPath)
		Expect(err).ToNot(HaveOccurred())

		actualContents, err := ioutil.ReadFile(generatedFakePath)
		Expect(err).ToNot(HaveOccurred())

		// assert file content matches what we expect
		Expect(string(actualContents)).To(Equal(string(expectedContents)))
	})

	when("when given a single argument", func() {
		it.Before(func() {
			copyIn("other_types.go", pathToCLI)
			copyIn("something.go", tmpPath("otherrepo.com"))
		})

		it("writes a fake for the fully qualified interface that is provided in the argument", func() {
			session := startCounterfeiterWithoutFixture(pathToCLI, "otherrepo.com/fixtures.Something")

			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Expect(output).To(ContainSubstring("Wrote `FakeSomething`"))
		})
	})

	when("when given two arguments", func() {
		it.Before(func() {
			copyIn("something.go", pathToCLI)
		})

		it("writes a fake for the given interface from the provided file", func() {
			session := startCounterfeiter(pathToCLI, "something.go", "Something")

			Eventually(session).Should(gexec.Exit(0))
			output := string(session.Out.Contents())

			Expect(output).To(ContainSubstring("Wrote `FakeSomething`"))
		})
	})

	when("when provided three arguments", func() {
		it.Before(func() {
			copyIn("something.go", pathToCLI)
		})

		it("writes the fake to stdout", func() {
			session := startCounterfeiter(pathToCLI, "something.go", "Something", "-")

			Eventually(session).Should(gexec.Exit(0))
			stdout := string(session.Out.Contents())
			stderr := string(session.Err.Contents())

			Expect(stdout).To(ContainSubstring("// Code generated by counterfeiter. DO NOT EDIT."))
			Expect(stdout).NotTo(ContainSubstring("Wrote `FakeSomething"))

			Expect(stderr).To(ContainSubstring("Wrote `FakeSomething"))
		})
	})
}
