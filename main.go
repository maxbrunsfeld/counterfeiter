package main

import (
	"flag"
	"fmt"
	"github.com/maxbrunsfeld/counterfeiter/counterfeiter"
	"os"
)

var fakePackage = flag.String(
	"fakePackage",
	"fakes",
	"The package name for the generated fake",
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		fmt.Println("Usage - counterfeiter PACKAGE_NAME INTERFACE_NAME [ --fakePackage = FAKE_PACKAGE_NAME ]")
		os.Exit(1)
	}

	code, err := counterfeiter.Generate(args[0], args[1], *fakePackage)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(code)
}

