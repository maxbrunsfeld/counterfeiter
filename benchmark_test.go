package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"
	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func BenchmarkWithoutCache(b *testing.B) {
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

	cache := &generator.FakeCache{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		doGenerate(workingDir, args, cache)
	}
}

func BenchmarkWithCache(b *testing.B) {
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

	cache := &generator.Cache{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		doGenerate(workingDir, args, cache)
	}
}
