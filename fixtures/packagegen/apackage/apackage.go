package ostest

import (
	"fmt"
	"os"
	"time"
)

func FindProcess(pid int) (*os.Process, error) {
	return os.FindProcess(pid)
}

func Hostname() (name string, err error) {
	return os.Hostname()
}

func Expand(s string, mapping func(string) string) string {
	return os.Expand(s, mapping)
}

func Clearenv() {
	os.Clearenv()
}

func Environ() []string {
	return os.Environ()
}

func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}

func MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func Exit(code int) {
	os.Exit(code)
}

func Fictional(lol ...string) {
	fmt.Printf("%#v", lol)
}
