package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"

	"github.com/maxbrunsfeld/counterfeiter/arguments"
	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"
	"github.com/maxbrunsfeld/counterfeiter/terminal"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fail("%s", usage)
		return
	}

	argumentParser := arguments.NewArgumentParser(
		fail,
		cwd,
		filepath.EvalSymlinks,
		os.Stat,
		terminal.NewUI(),
		locator.NewInterfaceLocator(),
	)
	parsedArgs := argumentParser.ParseArguments(args...)

	interfaceName := parsedArgs.InterfaceName
	fakeName := parsedArgs.FakeImplName
	sourceDir := parsedArgs.SourcePackageDir
	outputPath := parsedArgs.OutputPath
	destinationPackage := parsedArgs.DestinationPackageName

	iface, err := locator.GetInterfaceFromFilePath(interfaceName, sourceDir)
	if err != nil {
		fail("%v", err)
	}

	code, err := generator.CodeGenerator{
		Model:       *iface,
		StructName:  fakeName,
		PackageName: destinationPackage,
	}.GenerateFake()

	if err != nil {
		fail("%v", err)
	}

	newCode, err := format.Source([]byte(code))
	code = string(newCode)

	if parsedArgs.PrintToStdOut {
		fmt.Println(code)
	} else {
		os.MkdirAll(filepath.Dir(outputPath), 0777)
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
		Path to the file or directory for the generated fakes.
		This also determines the package name that will be used.
		By default, the generated fakes will be generated in
		the package "xyzfakes" which is nested in package "xyz",
		where "xyz" is the name of referenced package.

	example:
		# writes "FakeMyInterface" to ./fakes/fake_my_interface.go
		counterfeiter -o ./fakes MyInterface ./mypackage

	--fake-name
		Name of the fake struct to generate. By default, 'Fake' will
		be prepended to the name of the original interface.

	example:
		# writes "CoolThing" to ./fakes/cool_thing.go
		counterfeiter --fake-name CoolThing MyInterface ./mypackage
`
