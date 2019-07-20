package main

import (
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"runtime/pprof"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"
	"github.com/maxbrunsfeld/counterfeiter/v6/command"
	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func main() {
	debug.SetGCPercent(-1)

	if err := run(); err != nil {
		fail("%v", err)
	}
}

func run() error {
	profile := os.Getenv("COUNTERFEITER_PROFILE") != ""
	if profile {
		p, err := filepath.Abs(filepath.Join(".", "counterfeiter.profile"))
		if err != nil {
			return err
		}
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		fmt.Printf("Profile: %s\n", p)
		defer pprof.StopCPUProfile()
	}

	log.SetFlags(log.Lshortfile)
	if !isDebug() {
		log.SetOutput(ioutil.Discard)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.New("Error - couldn't determine current working directory")
	}

	var cache generator.Cacher
	if disableCache() {
		cache = &generator.FakeCache{}
	} else {
		cache = &generator.Cache{}
	}
	var invocations []command.Invocation
	var args *arguments.ParsedArguments
	args, _ = arguments.New(os.Args, cwd, filepath.EvalSymlinks, os.Stat)
	generateMode := false
	if args != nil {
		generateMode = args.GenerateMode
	}
	invocations, err = command.Detect(cwd, os.Args, generateMode)
	if err != nil {
		return err
	}

	for i := range invocations {
		a, err := arguments.New(invocations[i].Args, cwd, filepath.EvalSymlinks, os.Stat)
		if err != nil {
			return err
		}
		err = generate(cwd, a, cache)
		if err != nil {
			return err
		}
	}
	return nil
}

func isDebug() bool {
	return os.Getenv("COUNTERFEITER_DEBUG") != ""
}

func disableCache() bool {
	return os.Getenv("COUNTERFEITER_DISABLECACHE") != ""
}

func generate(workingDir string, args *arguments.ParsedArguments, cache generator.Cacher) error {
	if err := reportStarting(workingDir, args.OutputPath, args.FakeImplName); err != nil {
		return err
	}

	b, err := doGenerate(workingDir, args, cache)
	if err != nil {
		return err
	}

	if err := printCode(b, args.OutputPath, args.PrintToStdOut); err != nil {
		return err
	}
	fmt.Fprint(os.Stderr, "Done\n")
	return nil
}

func doGenerate(workingDir string, args *arguments.ParsedArguments, cache generator.Cacher) ([]byte, error) {
	mode := generator.InterfaceOrFunction
	if args.GenerateInterfaceAndShimFromPackageDirectory {
		mode = generator.Package
	}
	f, err := generator.NewFake(mode, args.InterfaceName, args.PackagePath, args.FakeImplName, args.DestinationPackageName, workingDir, cache)
	if err != nil {
		return nil, err
	}
	return f.Generate(true)
}

func printCode(code []byte, outputPath string, printToStdOut bool) error {
	formattedCode, err := format.Source(code)
	if err != nil {
		return err
	}

	if printToStdOut {
		fmt.Println(string(formattedCode))
		return nil
	}
	os.MkdirAll(filepath.Dir(outputPath), 0777)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Couldn't create fake file - %v", err)
	}

	_, err = file.Write(formattedCode)
	if err != nil {
		return fmt.Errorf("Couldn't write to fake file - %v", err)
	}
	return nil
}

func reportStarting(workingDir string, outputPath, fakeName string) error {
	rel, err := filepath.Rel(workingDir, outputPath)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Writing `%s` to `%s`... ", fakeName, rel)
	if isDebug() {
		msg = msg + "\n"
	}
	fmt.Fprint(os.Stderr, msg)
	return nil
}

func fail(s string, args ...interface{}) {
	fmt.Printf("\n"+s+"\n", args...)
	os.Exit(1)
}
