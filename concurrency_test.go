package main_test

import (
	"io"
	"log"
	"path/filepath"
	"sync"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestConcurrency(t *testing.T) {
	log.SetOutput(io.Discard)
	spec.Run(t, "Concurrency", testConcurrency, spec.Report(report.Terminal{}))
}

// These specs run every public entry point of the generator from several
// goroutines at once. They only mean something under the race detector, which
// is why they live in the root package: it is the one that CI always runs with
// -race.
func testConcurrency(t *testing.T, when spec.G, it spec.S) {
	g := NewWithT(t)

	type target struct {
		mode        generator.FakeMode
		name        string
		pkg         string
		fake        string
		destination string
	}
	targets := []target{
		{generator.InterfaceOrFunction, "Something", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures", "FakeSomething", "fixturesfakes"},
		{generator.InterfaceOrFunction, "SomethingFactory", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures", "FakeSomethingFactory", "fixturesfakes"},
		{generator.InterfaceOrFunction, "GenericInterface", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/genericinterface", "FakeGenericInterface", "genericinterfacefakes"},
		{generator.Package, "", "github.com/maxbrunsfeld/counterfeiter/v6/fixtures/packagemode", "Packagemode", "packagemodeshim"},
	}
	generate := func(tt target, cache generator.Cacher) ([]byte, error) {
		f, err := generator.NewFake(tt.mode, tt.name, tt.pkg, tt.fake, tt.destination, "", "", cache)
		if err != nil {
			return nil, err
		}
		return f.Generate(true)
	}

	it("generates every kind of fake from many goroutines sharing one cache", func() {
		expected := make([]string, len(targets))
		for i, tt := range targets {
			b, err := generate(tt, &generator.FakeCache{})
			g.Expect(err).NotTo(HaveOccurred())
			expected[i] = string(b)
		}

		const workers = 4
		cache := &generator.Cache{}
		results := make([][]byte, workers*len(targets))
		errs := make([]error, len(results))
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			for i, tt := range targets {
				wg.Add(1)
				go func(slot int, tt target) {
					defer wg.Done()
					results[slot], errs[slot] = generate(tt, cache)
				}(w*len(targets)+i, tt)
			}
		}
		wg.Wait()

		for slot := range results {
			g.Expect(errs[slot]).NotTo(HaveOccurred())
			g.Expect(string(results[slot])).To(Equal(expected[slot%len(targets)]))
		}
	})

	it("reads header files from many goroutines sharing one reader", func() {
		dir, err := filepath.Abs(filepath.Join("fixtures", "headers"))
		g.Expect(err).NotTo(HaveOccurred())
		files := []string{"default.header.go.txt", "specific.header.go.txt"}
		expected := make([]string, len(files))
		for i, file := range files {
			expected[i], err = (&generator.SimpleFileReader{}).Get(dir, file)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(expected[i]).NotTo(BeEmpty())
		}

		const workers = 4
		reader := &generator.CachedFileReader{}
		results := make([]string, workers*len(files))
		errs := make([]error, len(results))
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			for i, file := range files {
				wg.Add(1)
				go func(slot int, file string) {
					defer wg.Done()
					results[slot], errs[slot] = reader.Get(dir, file)
				}(w*len(files)+i, file)
			}
		}
		wg.Wait()

		for slot := range results {
			g.Expect(errs[slot]).NotTo(HaveOccurred())
			g.Expect(results[slot]).To(Equal(expected[slot%len(files)]))
		}
	})
}
