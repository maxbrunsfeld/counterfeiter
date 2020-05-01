// +build !windows

package generator_test

const (
	relFile    = "file.ext"
	absFile    = "/file.ext"
	workingDir = "/some/dir"

	relFileUp      = "../file.ext"
	expectedFileUp = "/some/file.ext"
)
