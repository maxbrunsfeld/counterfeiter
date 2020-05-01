package main

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/arguments"
	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func BenchmarkDoGenerate(b *testing.B) {
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

	caches := map[string]struct {
		cache        generator.Cacher
		headerReader generator.FileReader
	}{
		"without caches": {
			cache:        &generator.FakeCache{},
			headerReader: &generator.SimpleFileReader{},
		},
		"with caches": {
			cache:        &generator.Cache{},
			headerReader: &generator.CachedFileReader{},
		},
	}

	headers := map[string]string{
		"without headerfile": "",
		"with headerfile":    "headers/default.header.go.txt",
	}

	for name, caches := range caches {
		caches := caches
		b.Run(name, func(b *testing.B) {
			for name, headerFile := range headers {
				headerFile := headerFile
				b.Run(name, func(b *testing.B) {
					args.HeaderFile = headerFile
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						if _, err := doGenerate(workingDir, args, caches.cache, caches.headerReader); err != nil {
							b.Errorf("Expected doGenerate not to return an error, got %v", err)
						}
					}
				}) // b.Run for headerFiles
			}
		}) // b.Run for caches
	}
}
