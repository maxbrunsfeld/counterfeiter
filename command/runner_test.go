package command

import (
	"os"
	"testing"

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
	})

	it.After(func() {
		reset()
	})

	when("counterfeiter has been invoked directly", func() {
		it.Before(func() {
		})

		it.Focus("creates an invocation", func() {
			i, err := invocations([]string{"counterfeiter", ".", "AliasedInterface"})
			Expect(err).NotTo(HaveOccurred())
			Expect(i).NotTo(BeNil())
			Expect(i).To(HaveLen(1))
		})
	})

	when("counterfeiter has been invoked by go generate", func() {
		it.Before(func() {
			os.Setenv("DOLLAR", "$")
			os.Setenv("GOFILE", "aliased_interfaces.go")
			os.Setenv("GOLINE", "5")
			os.Setenv("GOPACKAGE", "fixtures")
		})

		it("creates an invocation", func() {

		})
	})
}
