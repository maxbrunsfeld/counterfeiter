package arguments

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/maxbrunsfeld/counterfeiter/locator"
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
	interfaceLocator locator.InterfaceLocator,
) ArgumentParser {
	return &argumentParser{
		ui:                ui,
		failHandler:       failHandler,
		currentWorkingDir: currentWorkingDir,
		symlinkEvaler:     symlinkEvaler,
		fileStatReader:    fileStatReader,
		interfaceLocator:  interfaceLocator,
	}
}

func (argParser *argumentParser) ParseArguments(args ...string) ParsedArguments {
	sourcePackageDir := argParser.getSourceDir(args[0])

	var interfaceName string

	if len(args) > 1 {
		interfaceName = args[1]
	} else {
		interfaceName = argParser.PromptUserForInterfaceName(sourcePackageDir)
	}

	fakeImplName := getFakeName(interfaceName, *fakeNameFlag)

	outputPath := argParser.getOutputPath(
		sourcePackageDir,
		fakeImplName,
		*outputPathFlag,
	)

	return ParsedArguments{
		SourcePackageDir: sourcePackageDir,
		OutputPath:       outputPath,

		InterfaceName: interfaceName,
		FakeImplName:  fakeImplName,

		PrintToStdOut: any(args, "-"),
	}
}

func (parser *argumentParser) PromptUserForInterfaceName(filepath string) string {
	if !(parser.ui.TerminalIsTTY()) {
		parser.ui.WriteLine("Cowardly refusing to prompt user for an interface name in a non-tty environment")
		parser.failHandler("Perhaps you meant to invoke counterfeiter with more than one argument?")
		return ""
	}

	parser.ui.WriteLine("Which interface to counterfeit?")

	interfacesInPackage := parser.interfaceLocator.GetInterfacesFromFilePath(filepath)

	for i, interfaceName := range interfacesInPackage {
		parser.ui.WriteLine(fmt.Sprintf("%d. %s", i+1, interfaceName))
	}
	parser.ui.WriteLine("")

	response := parser.ui.ReadLineFromStdin()
	parsedResponse, err := strconv.ParseInt(response, 10, 64)
	if err != nil {
		parser.failHandler("Unknown option '%s'", response)
		return ""
	}

	option := int(parsedResponse - 1)
	if option < 0 || option >= len(interfacesInPackage) {
		parser.failHandler("Unknown option '%s'", response)
		return ""
	}

	return interfacesInPackage[option]
}

type argumentParser struct {
	ui                terminal.UI
	interfaceLocator  locator.InterfaceLocator
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

func (argParser *argumentParser) getOutputPath(sourceDir, fakeName, arg string) string {
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

type FailHandler func(string, ...interface{})
