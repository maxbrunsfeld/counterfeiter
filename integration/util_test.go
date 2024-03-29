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
