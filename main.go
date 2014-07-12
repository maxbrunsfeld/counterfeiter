package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"
)

var usage = `
USAGE
	counterfeiter
		[-o <output-path>] [--fake-name <fake-name>]
		<source-path> <interface-name> [-]

ARGUMENTS
	source-path
		Path to the file or directory containing the interface to fake

	interface-name
		Name of the interface to fake

	'-' argument
		Write code to standard out instead of to a file

OPTIONS
	-o
		Path to the file or directory to which code should be written.
		This also determines the package name that will be used.
		By default, code will be written to a 'fakes' directory inside
		of the directory containing the original interface.
	
	--fake-name
		Name of the fake struct to generate. By default, 'Fake' will
		be prepended to the name of the original interface.
`

var outputPathFlag = flag.String(
	"o",
	"",
	"The file or directory to which the generated fake will be written",
)

var fakeNameFlag = flag.String(
	"fake-name",
	"",
	"The name of the fake struct",
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fail("%s", usage)
	}

	sourceDir := getSourceDir(args[0])
	interfaceName := args[1]
	fakeName := getFakeName(interfaceName, *fakeNameFlag)
	outputPath := getOutputPath(sourceDir, fakeName, *outputPathFlag)
	outputDir := filepath.Dir(outputPath)
	fakePackageName := filepath.Base(outputDir)
	shouldPrintToStdout := len(args) >= 3 && args[2] == "-"

	iface, err := locator.GetInterfaceFromFilePath(interfaceName, sourceDir)
	if err != nil {
		fail("%v", err)
	}

	code, err := generator.CodeGenerator{
		Model:       *iface,
		StructName:  fakeName,
		PackageName: fakePackageName,
	}.GenerateFake()

	if err != nil {
		fail("%v", err)
	}

	if shouldPrintToStdout {
		fmt.Println(code)
	} else {
		os.MkdirAll(outputDir, 0777)
		file, err := os.Create(outputPath)
		if err != nil {
			fail("Couldn't create fake file - %v", err)
		}

		_, err = file.WriteString(code)
		if err != nil {
			fail("Couldn't write to fake file - %v", err)
		}

		rel, err := filepath.Rel(cwd(), outputPath)
		if err != nil {
			fail("%v", err)
		}

		fmt.Printf("Wrote `%s` to `%s`\n", fakeName, rel)
	}
}

func getSourceDir(arg string) string {
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(cwd(), arg)
	}

	stat, err := os.Stat(arg)
	if err != nil {
		fail("No such file or directory '%s'", arg)
	}

	if !stat.IsDir() {
		return filepath.Dir(arg)
	} else {
		return arg
	}
}

func getOutputPath(sourceDir, fakeName, arg string) string {
	if arg == "" {
		return filepath.Join(sourceDir, "fakes", snakeCase(fakeName)+".go")
	} else {
		if !filepath.IsAbs(arg) {
			arg = filepath.Join(cwd(), arg)
		}
		return arg
	}
}

func getFakeName(interfaceName, arg string) string {
	if arg == "" {
		return "Fake" + interfaceName
	} else {
		return arg
	}
}

func snakeCase(input string) string {
	camelRegexp := regexp.MustCompile("([a-z])([A-Z])")
	return strings.ToLower(camelRegexp.ReplaceAllString(input, "${1}_${2}"))
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		fail("Error - couldn't determine current working directory")
	}
	return dir
}

func fail(s string, args ...interface{}) {
	fmt.Printf(s+"\n", args...)
	os.Exit(1)
}
