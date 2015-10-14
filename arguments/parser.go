package arguments

import (
	"path/filepath"
	"regexp"
	"strings"
)

type ArgumentParser interface {
	ParseArguments(...string) ParsedArguments
}

func NewArgumentParser(
	failHandler FailHandler,
	currentWorkingDir CurrentWorkingDir,
	symlinkEvaler SymlinkEvaler,
	fileStatReader FileStatReader,
) ArgumentParser {
	return argumentParser{
		failHandler:       failHandler,
		currentWorkingDir: currentWorkingDir,
		symlinkEvaler:     symlinkEvaler,
		fileStatReader:    fileStatReader,
	}
}

func (argParser argumentParser) ParseArguments(args ...string) ParsedArguments {
	sourcePackageDir := argParser.getSourceDir(args[0])
	fakeImplName := getFakeName(args[1], *fakeNameFlag)
	outputPath := argParser.getOutputPath(sourcePackageDir, fakeImplName, *outputPathFlag)

	return ParsedArguments{
		SourcePackageDir: sourcePackageDir,
		OutputPath:       outputPath,

		InterfaceName: args[1],
		FakeImplName:  fakeImplName,

		PrintToStdOut: len(args) == 3 && args[2] == "-",
	}
}

type argumentParser struct {
	failHandler       FailHandler
	currentWorkingDir CurrentWorkingDir
	symlinkEvaler     SymlinkEvaler
	fileStatReader    FileStatReader
}

type ParsedArguments struct {
	SourcePackageDir string // abs path to the dir containing the interface to fake
	OutputPath       string // path to write the fake file to

	InterfaceName string // the interface to counterfeit
	FakeImplName  string // the name of the struct implementing the given interface

	PrintToStdOut bool
}

func getFakeName(interfaceName, arg string) string {
	if arg == "" {
		return "Fake" + interfaceName
	} else {
		return arg
	}
}

var camelRegexp = regexp.MustCompile("([a-z])([A-Z])")

func (argParser argumentParser) getOutputPath(sourceDir, fakeName, arg string) string {
	if arg == "" {
		snakeCaseName := strings.ToLower(camelRegexp.ReplaceAllString(fakeName, "${1}_${2}"))
		return filepath.Join(sourceDir, "fakes", snakeCaseName+".go")
	} else {
		if !filepath.IsAbs(arg) {
			arg = filepath.Join(argParser.currentWorkingDir(), arg)
		}
		return arg
	}
}

func (argParser argumentParser) getSourceDir(arg string) string {
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

type FailHandler func(string, ...interface{})
