package arguments

import (
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/maxbrunsfeld/counterfeiter/terminal"
)

type ArgumentParser interface {
	ParseArguments(...string) ParsedArguments
}

func NewArgumentParser(
	failHandler FailHandler,
	currentWorkingDir CurrentWorkingDir,
	symlinkEvaler SymlinkEvaler,
	fileStatReader FileStatReader,
	ui terminal.UI,
) ArgumentParser {
	return &argumentParser{
		ui:                ui,
		failHandler:       failHandler,
		currentWorkingDir: currentWorkingDir,
		symlinkEvaler:     symlinkEvaler,
		fileStatReader:    fileStatReader,
	}
}

func (argParser *argumentParser) ParseArguments(args ...string) ParsedArguments {
	if *packageFlag {
		return argParser.parsePackageArgs(args...)
	} else {
		return argParser.parseInterfaceArgs(args...)
	}
}

func (argParser *argumentParser) parseInterfaceArgs(args ...string) ParsedArguments {
	var interfaceName string
	var outputPathFlagValue string
	var rootDestinationDir string
	var sourcePackageDir string
	var importPath string

	if len(args) > 1 {
		interfaceName = args[1]
		outputPathFlagValue = *outputPathFlag
		sourcePackageDir = argParser.getSourceDir(args[0])
		rootDestinationDir = sourcePackageDir
	} else {
		fullyQualifiedInterface := strings.Split(args[0], ".")
		interfaceName = fullyQualifiedInterface[len(fullyQualifiedInterface)-1]
		rootDestinationDir = argParser.currentWorkingDir()
		importPath = strings.Join(fullyQualifiedInterface[:len(fullyQualifiedInterface)-1], ".")
	}

	fakeImplName := getFakeName(interfaceName, *fakeNameFlag)

	outputPath := argParser.getOutputPath(
		rootDestinationDir,
		fakeImplName,
		outputPathFlagValue,
	)

	packageName := restrictToValidPackageName(filepath.Base(filepath.Dir(outputPath)))

	return ParsedArguments{
		GenerateInterfaceAndShimFromPackageDirectory: false,
		SourcePackageDir:                             sourcePackageDir,
		OutputPath:                                   outputPath,
		ImportPath:                                   importPath,

		InterfaceName:          interfaceName,
		DestinationPackageName: packageName,
		FakeImplName:           fakeImplName,

		PrintToStdOut: any(args, "-"),
	}
}

func (argParser *argumentParser) parsePackageArgs(args ...string) ParsedArguments {
	dir := argParser.getPackageDir(args[0])

	packageName := path.Base(dir) + "shim"

	var outputPath string
	if *outputPathFlag != "" {
		// TODO: sensible checking of dirs and symlinks
		outputPath = *outputPathFlag
	} else {
		outputPath = path.Join(argParser.currentWorkingDir(), packageName)
	}

	return ParsedArguments{
		GenerateInterfaceAndShimFromPackageDirectory: true,
		SourcePackageDir:                             dir,
		OutputPath:                                   outputPath,

		DestinationPackageName: packageName,

		PrintToStdOut: any(args, "-"),
	}
}

type argumentParser struct {
	ui                terminal.UI
	failHandler       FailHandler
	currentWorkingDir CurrentWorkingDir
	symlinkEvaler     SymlinkEvaler
	fileStatReader    FileStatReader
}

type ParsedArguments struct {
	GenerateInterfaceAndShimFromPackageDirectory bool

	SourcePackageDir string // abs path to the dir containing the interface to fake
	ImportPath       string // import path to the package containing the interface to fake
	OutputPath       string // path to write the fake file to

	DestinationPackageName string // often the base-dir for OutputPath but must be a valid package name

	InterfaceName string // the interface to counterfeit
	FakeImplName  string // the name of the struct implementing the given interface

	PrintToStdOut bool
}

func fixupUnexportedNames(interfaceName string) string {
	asRunes := []rune(interfaceName)
	if len(asRunes) == 0 || !unicode.IsLower(asRunes[0]) {
		return interfaceName
	}
	asRunes[0] = unicode.ToUpper(asRunes[0])
	return string(asRunes)
}

func getFakeName(interfaceName, arg string) string {
	if arg == "" {
		interfaceName = fixupUnexportedNames(interfaceName)
		return "Fake" + interfaceName
	} else {
		return arg
	}
}

var camelRegexp = regexp.MustCompile("([a-z])([A-Z])")

func (argParser *argumentParser) getOutputPath(rootDestinationDir, fakeName, arg string) string {
	if arg == "" {
		snakeCaseName := strings.ToLower(camelRegexp.ReplaceAllString(fakeName, "${1}_${2}"))
		return filepath.Join(rootDestinationDir, packageNameForPath(rootDestinationDir), snakeCaseName+".go")
	} else {
		if !filepath.IsAbs(arg) {
			arg = filepath.Join(argParser.currentWorkingDir(), arg)
		}
		return arg
	}
}

func packageNameForPath(pathToPackage string) string {
	_, packageName := filepath.Split(pathToPackage)
	return packageName + "fakes"
}

func (argParser *argumentParser) getPackageDir(arg string) string {
	if filepath.IsAbs(arg) {
		return arg
	}

	pathToCheck := path.Join(runtime.GOROOT(), "src", arg)

	stat, err := argParser.fileStatReader(pathToCheck)
	if err != nil {
		argParser.failHandler("No such file or directory '%s'", arg)
	}
	if !stat.IsDir() {
		argParser.failHandler("No such file or directory '%s'", arg) // TODO: for now?
	}

	return pathToCheck
}

func (argParser *argumentParser) getSourceDir(arg string) string {
	if !filepath.IsAbs(arg) {
		arg = filepath.Join(argParser.currentWorkingDir(), arg)
	}

	arg, _ = argParser.symlinkEvaler(arg)
	stat, err := argParser.fileStatReader(arg)
	if err != nil {
		argParser.failHandler("No such file or directory '%s'", arg)
	}

	if !stat.IsDir() {
		return filepath.Dir(arg)
	} else {
		return arg
	}
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

type FailHandler func(string, ...interface{})
