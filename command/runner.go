package command

import (
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func Run(cwd string, args []string) ([]Invocation, error) {
	if invokedByGoGenerate() {
		return invocations(cwd, args)
	}
	i, err := NewInvocation("", 0, args)
	if err != nil {
		return nil, err
	}
	return []Invocation{i}, nil
}

func (i *Invocation) Validate() error {
	if len(i.Flags.Args()) < 1 {
		return errors.New("at least one argument must be supplied")
	}
	return nil
}

type Invocation struct {
	File                                         string
	Line                                         int
	Args                                         []string
	Flags                                        *flag.FlagSet
	FakeNameFlag                                 *string
	OutputPathFlag                               *string
	PackageModeFlag                              *bool
	GenerateInterfaceAndShimFromPackageDirectory bool

	SourcePackageDir string // abs path to the dir containing the interface to fake
	PackagePath      string // package path to the package containing the interface to fake
	OutputPath       string // path to write the fake file to

	DestinationPackageName string // often the base-dir for OutputPath but must be a valid package name

	InterfaceName string // the interface to counterfeit
	FakeImplName  string // the name of the struct implementing the given interface

	PrintToStdOut bool
}

func NewInvocation(file string, line int, args []string) (Invocation, error) {
	if len(args) < 1 {
		return Invocation{}, fmt.Errorf("%s:%v an invocation of counterfeiter must have arguments", file, line)
	}
	i := Invocation{
		File: file,
		Line: line,
		Args: args,
	}
	i.Flags = flag.NewFlagSet("counterfeiter", flag.ContinueOnError)
	i.FakeNameFlag = i.Flags.String("fake-name", "", "The name of the fake struct")
	i.OutputPathFlag = i.Flags.String("o", "", "The file or directory to which the generated fake will be written")
	i.PackageModeFlag = i.Flags.Bool("p", false, "whether or not to generate a package shim")
	err := i.Flags.Parse(i.Args[1:])
	if err != nil {
		return Invocation{}, err
	}
	if len(i.Flags.Args()) < 1 {
		return Invocation{}, fmt.Errorf("%s:%v an invocation of counterfeiter must have arguments", file, line)
	}
	return i, nil
}

func invokedByGoGenerate() bool {
	return os.Getenv("DOLLAR") == "$"
}

func invocations(cwd string, args []string) ([]Invocation, error) {
	var result []Invocation
	// Find all the go files
	pkg, err := build.ImportDir(cwd, build.IgnoreVendor)
	if err != nil {
		return nil, err
	}

	gofiles := make([]string, 0, len(pkg.GoFiles)+len(pkg.CgoFiles)+len(pkg.TestGoFiles)+len(pkg.XTestGoFiles))
	gofiles = append(gofiles, pkg.GoFiles...)
	gofiles = append(gofiles, pkg.CgoFiles...)
	gofiles = append(gofiles, pkg.TestGoFiles...)
	gofiles = append(gofiles, pkg.XTestGoFiles...)
	sort.Strings(gofiles)
	// Find all the generate statements
	line, err := strconv.Atoi(os.Getenv("GOLINE"))
	if err != nil {
		return nil, err
	}
	for i := range gofiles {
		i, err := open(cwd, gofiles[i])
		if err != nil {
			return nil, err
		}
		result = append(result, i...)
		if len(result) > 0 && result[0].File != os.Getenv("GOFILE") {
			return nil, nil
		}
		if len(result) > 0 && result[0].Line != line {
			return nil, nil
		}
	}

	return result, nil
}

var re = regexp.MustCompile(`(?mi)^//go:generate counterfeiter(\.exe)? (.*)$`)

func open(dir string, file string) ([]Invocation, error) {
	str, err := ioutil.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(str), "\n")

	var result []Invocation
	line := 0
	for i := range lines {
		line++
		match := re.FindStringSubmatch(lines[i])
		if match == nil {
			continue
		}

		inv, err := NewInvocation(file, line, stringToArgs(match[2]))
		if err != nil {
			return nil, err
		}
		result = append(result, inv)
	}

	return result, nil
}

func stringToArgs(s string) []string {
	a := strings.Split(s, " ")
	result := []string{
		"counterfeiter",
	}
	for i := range a {
		item := strings.TrimSpace(a[i])
		if item == "" {
			continue
		}
		result = append(result, item)
	}
	return result
}
