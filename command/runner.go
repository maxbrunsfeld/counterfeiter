package command

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
)

func Run() {

}

type invocation struct {
	file  string
	line  int
	args  []string
	flags *flag.FlagSet
}

func invokedByGoGenerate() bool {
	return os.Getenv("DOLLAR") == "$"
}

func invocations(args []string) ([]invocation, error) {
	if !invokedByGoGenerate() {
		i := invocation{
			args:  args,
			flags: flag.NewFlagSet(args[0], flag.ContinueOnError),
		}
		i.flags.Parse(args[1:])
		if len(i.flags.Args()) < 1 {
			return nil, errors.New("at least one argument must be supplied")
		}
		return []invocation{i}, nil
	}
	var result []invocation
	fmt.Println("invoked by go generate")
	// Find all the go files
	dir := filepath.Dir(os.Getenv("GOFILE"))
	pkg, err := build.ImportDir(dir, build.IgnoreVendor)
	if err != nil {
		log.Fatal(err)
	}

	gofiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.CgoFiles)+len(pkg.TestGoFiles)+len(pkg.XTestGoFiles))
	gofiles = append(gofiles, pkg.GoFiles...)
	gofiles = append(gofiles, pkg.CgoFiles...)
	gofiles = append(gofiles, pkg.TestGoFiles...)
	gofiles = append(gofiles, pkg.XTestGoFiles...)

	// Find all the generate statements

	firstComment := false

	if !firstComment {
		log.Printf("GOFILE: %s\n", os.Getenv("GOFILE"))
		log.Printf("GOLINE: %s\n", os.Getenv("GOLINE"))
		log.Printf("GOPACKAGE: %s\n", os.Getenv("GOPACKAGE"))
		os.Exit(0) // Bail out if we're not the first comment in the first file with a comment
	}
	return result, nil
}
