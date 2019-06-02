package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"
)

func BenchmarkSingleRun(b *testing.B) {
	b.StopTimer()
	workingDir, err := filepath.Abs(filepath.Join(".", "fixtures"))
	if err != nil {
		b.Fatal(err)
	}
	log.SetOutput(ioutil.Discard)

	args := &arguments.ParsedArguments{
		GenerateInterfaceAndShimFromPackageDirectory: false,
		SourcePackageDir:       workingDir,
		PackagePath:            workingDir,
		OutputPath:             filepath.Join(workingDir, "fixturesfakes", "fake_something.go"),
		DestinationPackageName: "fixturesfakes",
		InterfaceName:          "Something",
		FakeImplName:           "FakeSomething",
		PrintToStdOut:          false,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		doGenerate(workingDir, args)
	}
}
