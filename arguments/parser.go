package arguments

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

type flagsForGenerate struct {
	FakeNameTemplate *string
}

func (f *flagsForGenerate) RegisterFlags(fs *flag.FlagSet) {
	f.FakeNameTemplate = fs.String(
		"fake-name-template",
		"",
		`A template for the names of the fake structs in a generate call. Example: "The{{.TargetName}}Imposter"`,
	)
}

type flagsForNonGenerate struct {
	FakeName *string
	Package  *bool
}

func (f *flagsForNonGenerate) RegisterFlags(fs *flag.FlagSet) {
	f.FakeName = fs.String(
		"fake-name",
		"",
		"The name of the fake struct",
	)

	f.Package = fs.Bool(
		"p",
		false,
		"Whether or not to generate a package shim",
	)
}

type sharedFlags struct {
	Generate   *bool
	OutputPath *string
	Header     *string
	Quiet      *bool
	Help       *bool
}

func (f *sharedFlags) RegisterFlags(fs *flag.FlagSet) {
	f.OutputPath = fs.String(
		"o",
		"",
		"The file or directory to which the generated fake will be written",
	)

	f.Generate = fs.Bool(
		"generate",
		false,
		"Identify all //counterfeiter:generate directives in the current working directory and generate fakes for them",
	)

	f.Header = fs.String(
		"header",
		"",
		"A path to a file that should be used as a header for the generated fake",
	)

	f.Quiet = fs.Bool(
		"q",
		false,
		"Suppress status statements",
	)

	f.Help = fs.Bool(
		"help",
		false,
		"Display this help",
	)
}

type allFlags struct {
	flagsForGenerate
	flagsForNonGenerate
	sharedFlags
}

func (f *allFlags) RegisterFlags(fs *flag.FlagSet) {
	f.flagsForGenerate.RegisterFlags(fs)
	f.flagsForNonGenerate.RegisterFlags(fs)
	f.sharedFlags.RegisterFlags(fs)
}

type standardFlags struct {
	flagsForNonGenerate
	sharedFlags
}

func (f *standardFlags) RegisterFlags(fs *flag.FlagSet) {
	f.flagsForNonGenerate.RegisterFlags(fs)
	f.sharedFlags.RegisterFlags(fs)
}

type GenerateArgs struct {
	OutputPath       string
	FakeNameTemplate *template.Template
	Header           string
	Quiet            bool
}

func ParseGenerateMode(args []string) (bool, *GenerateArgs, error) {
	if len(args) == 0 {
		return false, nil, errors.New("argument parsing requires at least one argument")
	}

	fs := flag.NewFlagSet("counterfeiter", flag.ContinueOnError)
	flags := new(allFlags)
	flags.RegisterFlags(fs)

	err := fs.Parse(args[1:])
	if err != nil {
		return false, nil, err
	}

	if *flags.Help {
		return false, nil, errors.New(usage)
	}
	if !*flags.Generate {
		return false, nil, nil
	}

	fakeNameTemplate, err := template.New("counterfeiter").Parse(*flags.FakeNameTemplate)
	if err != nil {
		return false, nil, fmt.Errorf("error parsing fake-name-template: %w", err)
	}

	return true, &GenerateArgs{
		OutputPath:       *flags.OutputPath,
		FakeNameTemplate: fakeNameTemplate,
		Header:           *flags.Header,
		Quiet:            *flags.Quiet,
	}, nil
}

type FakeNameTemplateArg struct {
	TargetName string
}

func New(args []string, workingDir string, generateArgs *GenerateArgs, evaler Evaler, stater Stater) (*ParsedArguments, error) {
	if len(args) == 0 {
		return nil, errors.New("argument parsing requires at least one argument")
	}

	fs := flag.NewFlagSet("counterfeiter", flag.ContinueOnError)
	flags := new(standardFlags)
	flags.RegisterFlags(fs)

	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	if len(fs.Args()) == 0 && !*flags.Generate {
		return nil, errors.New(usage)
	}

	header := *flags.Header
	outputPath := *flags.OutputPath
	quiet := *flags.Quiet
	fakeName := *flags.FakeName

	if generateArgs != nil {
		header = or(header, generateArgs.Header)
		quiet = quiet || generateArgs.Quiet

	}

	result := &ParsedArguments{
		PrintToStdOut: any(args, "-"),
		GenerateInterfaceAndShimFromPackageDirectory: *flags.Package,
		HeaderFile: header,
		Quiet:      quiet,
	}

	err = result.parseSourcePackageDir(*flags.Package, workingDir, evaler, stater, fs.Args())
	if err != nil {
		return nil, err
	}
	result.parseInterfaceName(*flags.Package, fs.Args())

	if generateArgs != nil {
		outputPath = or(outputPath, generateArgs.OutputPath)

		if fakeName == "" && generateArgs.FakeNameTemplate != nil {
			fakeNameWriter := new(bytes.Buffer)
			err = generateArgs.FakeNameTemplate.Execute(
				fakeNameWriter,
				FakeNameTemplateArg{TargetName: fixupUnexportedNames(result.InterfaceName)},
			)
			if err != nil {
				return nil, fmt.Errorf("error evaluating fake-name-template: %w", err)
			}
			fakeName = fakeNameWriter.String()
		}
	}
	result.parseFakeName(*flags.Package, fakeName, fs.Args())
	result.parseOutputPath(*flags.Package, workingDir, outputPath, fs.Args())
	result.parseDestinationPackageName(*flags.Package, fs.Args())
	result.parsePackagePath(*flags.Package, fs.Args())
	return result, nil
}

func or(opts ...string) string {
	for _, s := range opts {
		if s != "" {
			return s
		}
	}
	return ""
}

func (a *ParsedArguments) PrettyPrint() {
	b, _ := json.Marshal(a)
	fmt.Println(string(b))
}

func (a *ParsedArguments) parseInterfaceName(packageMode bool, args []string) {
	if packageMode {
		a.InterfaceName = ""
		return
	}
	if len(args) == 1 {
		fullyQualifiedInterface := strings.Split(args[0], ".")
		a.InterfaceName = fullyQualifiedInterface[len(fullyQualifiedInterface)-1]
	} else {
		a.InterfaceName = args[1]
	}
}

func (a *ParsedArguments) parseSourcePackageDir(packageMode bool, workingDir string, evaler Evaler, stater Stater, args []string) error {
	if packageMode {
		a.SourcePackageDir = args[0]
		return nil
	}
	if len(args) <= 1 {
		return nil
	}
	s, err := getSourceDir(args[0], workingDir, evaler, stater)
	if err != nil {
		return err
	}
	a.SourcePackageDir = s
	return nil
}

func (a *ParsedArguments) parseFakeName(packageMode bool, fakeName string, args []string) {
	if packageMode {
		a.parsePackagePath(packageMode, args)
		a.FakeImplName = strings.ToUpper(path.Base(a.PackagePath))[:1] + path.Base(a.PackagePath)[1:]
		return
	}
	if fakeName == "" {
		fakeName = "Fake" + fixupUnexportedNames(a.InterfaceName)
	}
	a.FakeImplName = fakeName
}

func (a *ParsedArguments) parseOutputPath(packageMode bool, workingDir string, outputPath string, args []string) {
	outputPathIsFilename := false
	if strings.HasSuffix(outputPath, ".go") {
		outputPathIsFilename = true
	}
	snakeCaseName := strings.ToLower(camelRegexp.ReplaceAllString(a.FakeImplName, "${1}_${2}"))

	if outputPath != "" {
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(workingDir, outputPath)
		}
		a.OutputPath = outputPath
		if !outputPathIsFilename {
			a.OutputPath = filepath.Join(a.OutputPath, snakeCaseName+".go")
		}
		return
	}

	if packageMode {
		a.parseDestinationPackageName(packageMode, args)
		a.OutputPath = path.Join(workingDir, a.DestinationPackageName, snakeCaseName+".go")
		return
	}

	d := workingDir
	if len(args) > 1 {
		d = a.SourcePackageDir
	}
	a.OutputPath = filepath.Join(d, packageNameForPath(d), snakeCaseName+".go")
}

func (a *ParsedArguments) parseDestinationPackageName(packageMode bool, args []string) {
	if packageMode {
		a.parsePackagePath(packageMode, args)
		a.DestinationPackageName = path.Base(a.PackagePath) + "shim"
		return
	}

	a.DestinationPackageName = restrictToValidPackageName(filepath.Base(filepath.Dir(a.OutputPath)))
}

func (a *ParsedArguments) parsePackagePath(packageMode bool, args []string) {
	if packageMode {
		a.PackagePath = args[0]
		return
	}
	if len(args) == 1 {
		fullyQualifiedInterface := strings.Split(args[0], ".")
		a.PackagePath = strings.Join(fullyQualifiedInterface[:len(fullyQualifiedInterface)-1], ".")
	} else {
		a.InterfaceName = args[1]
	}

	if a.PackagePath == "" {
		a.PackagePath = a.SourcePackageDir
	}
}

type ParsedArguments struct {
	GenerateInterfaceAndShimFromPackageDirectory bool

	SourcePackageDir string // abs path to the dir containing the interface to fake
	PackagePath      string // package path to the package containing the interface to fake
	OutputPath       string // path to write the fake file to

	DestinationPackageName string // often the base-dir for OutputPath but must be a valid package name

	InterfaceName string // the interface to counterfeit
	FakeImplName  string // the name of the struct implementing the given interface

	PrintToStdOut bool
	Quiet         bool

	HeaderFile string
}

func fixupUnexportedNames(interfaceName string) string {
	asRunes := []rune(interfaceName)
	if len(asRunes) == 0 || !unicode.IsLower(asRunes[0]) {
		return interfaceName
	}
	asRunes[0] = unicode.ToUpper(asRunes[0])
	return string(asRunes)
}

var camelRegexp = regexp.MustCompile("([a-z])([A-Z])")

func packageNameForPath(pathToPackage string) string {
	_, packageName := filepath.Split(pathToPackage)
	return packageName + "fakes"
}

func getSourceDir(path string, workingDir string, evaler Evaler, stater Stater) (string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(workingDir, path)
	}

	evaluatedPath, err := evaler(path)
	if err != nil {
		return "", fmt.Errorf("No such file/directory/package [%s]: %v", path, err)
	}

	stat, err := stater(evaluatedPath)
	if err != nil {
		return "", fmt.Errorf("No such file/directory/package [%s]: %v", path, err)
	}

	if !stat.IsDir() {
		return filepath.Dir(path), nil
	}
	return path, nil
}

func any(slice []string, needle string) bool {
	for _, str := range slice {
		if str == needle {
			return true
		}
	}

	return false
}

func restrictToValidPackageName(input string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		} else {
			return -1
		}
	}, input)
}
