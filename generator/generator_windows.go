// +build windows

package generator

// the filename provided to golang.org/x/tools/imports.Process
// cannot be empty on the windows platform, apparently
func tempProcessFilename() string {
	return "counterfeiter_temp_process_file"
}