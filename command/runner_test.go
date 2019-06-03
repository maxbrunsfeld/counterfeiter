package command_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/command"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRunner(t *testing.T) {
	spec.Run(t, "Runner", testRunner, spec.Report(report.Terminal{}))
}

func testRunner(t *testing.T, when spec.G, it spec.S) {
	reset := func() {
		os.Unsetenv("DOLLAR")
		os.Unsetenv("GOFILE")
		os.Unsetenv("GOLINE")
		os.Unsetenv("GOPACKAGE")
	}

	it.Before(func() {
		RegisterTestingT(t)
		reset()
		log.SetFlags(log.Llongfile)
	})

	it.After(func() {
		reset()
	})

	when("counterfeiter has been invoked directly", func() {
		it.Before(func() {
		})

		it("creates an invocation", func() {
			i, err := command.Detect(filepath.Join(".", "..", "fixtures"), []string{"counterfeiter", ".", "AliasedInterface"}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(i).NotTo(BeNil())
			Expect(i).To(HaveLen(1))
			Expect(i[0].Args).To(HaveLen(3))
			Expect(i[0].Args[1]).To(Equal("."))
			Expect(i[0].Args[2]).To(Equal("AliasedInterface"))
		})
	})

	when("counterfeiter is invoked in generate mode", func() {
		it.Before(func() {
			os.Unsetenv("DOLLAR")
			os.Unsetenv("GOFILE")
			os.Unsetenv("GOLINE")
			os.Unsetenv("GOPACKAGE")
		})

		it("creates invocations", func() {
			i, err := command.Detect(filepath.Join(".", "..", "fixtures"), []string{"counterfeiter", ".", "AliasedInterface"}, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(i).NotTo(BeNil())
			Expect(len(i)).To(Equal(17))
			Expect(i[0].File).To(Equal("aliased_interfaces.go"))
			Expect(i[0].Line).To(Equal(7))
			Expect(i[0].Args).To(HaveLen(3))
			Expect(i[0].Args[0]).To(Equal("counterfeiter"))
			Expect(i[0].Args[1]).To(Equal("."))
			Expect(i[0].Args[2]).To(Equal("AliasedInterface"))
		})
	})

	when("counterfeiter has been invoked by go generate", func() {
		it.Before(func() {
			os.Setenv("DOLLAR", "$")
			os.Setenv("GOFILE", "aliased_interfaces.go")
			os.Setenv("GOLINE", "5")
			os.Setenv("GOPACKAGE", "fixtures")
		})

		it("creates invocations but does not include generate mode as an invocation", func() {
			i, err := command.Detect(filepath.Join(".", "..", "fixtures"), []string{"counterfeiter", ".", "AliasedInterface"}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(i).NotTo(BeNil())
			Expect(len(i)).To(Equal(1))
			Expect(i[0].File).To(Equal("aliased_interfaces.go"))
			Expect(i[0].Line).To(Equal(5))
			Expect(i[0].Args).To(HaveLen(3))
			Expect(i[0].Args[0]).To(Equal("counterfeiter"))
			Expect(i[0].Args[1]).To(Equal("."))
			Expect(i[0].Args[2]).To(Equal("AliasedInterface"))
		})

		when("there is a mismatch in the file name", func() {
			it.Before(func() {
				os.Setenv("GOFILE", "some_other_file.go")
			})

			it("has no invocations", func() {
				i, err := command.Detect(filepath.Join(".", "..", "fixtures"), []string{"counterfeiter", ".", "AliasedInterface"}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(i).To(HaveLen(0))
			})
		})

		when("there is a mismatch in the line number", func() {
			it.Before(func() {
				os.Setenv("GOLINE", "100")
			})

			it("has no invocations", func() {
				i, err := command.Detect(filepath.Join(".", "..", "fixtures"), []string{"counterfeiter", ".", "AliasedInterface"}, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(i).To(HaveLen(0))
			})
		})
	})
}
