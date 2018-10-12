package integration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRoundTrip(t *testing.T) {
	spec.Run(t, "RoundTrip", testRoundTrip, spec.Report(report.Terminal{}))
}

func testRoundTrip(t *testing.T, when spec.G, it spec.S) {
	it("is here so that you can comment out the runTests function below when focusing tests", func() {})
	runTests(true, t, when, it)
}

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
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
	}
	Expect(err).NotTo(HaveOccurred())
}
