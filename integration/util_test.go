package integration_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"io/fs"

	. "github.com/onsi/gomega"
)

func WriteOutput(b []byte, file string) {
	_ = os.MkdirAll(filepath.Dir(file), 0700)
	_ = os.WriteFile(file, b, fs.FileMode(0600))
}

// RunVet type-checks the module including its test files.
func RunVet(baseDir string) {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = baseDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
	}
	Expect(err).NotTo(HaveOccurred())
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
