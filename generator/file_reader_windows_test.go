// +build windows

package generator_test

const (
	relFile    = "file.ext"
	absFile    = "c:\\file.ext"
	workingDir = "c:\\some\\dir"

	relFileUp      = "..\\file.ext"
	expectedFileUp = "c:\\some\\file.ext"
)
