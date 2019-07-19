package command

import (
	"log"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRunner(t *testing.T) {
	spec.Run(t, "Regexp", testRegexp, spec.Report(report.Terminal{}))
}

type Case struct {
	input   string
	matches bool
	args    []string
}

func testRegexp(t *testing.T, when spec.G, it spec.S) {
	var cases []Case

	it.Before(func() {
		RegisterTestingT(t)
		log.SetFlags(log.Llongfile)
		cases = []Case{
			{
				input:   "//go:generate counterfeiter . Intf",
				matches: true,
				args:    []string{".", "Intf"},
			},
			{
				input:   "//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Intf",
				matches: true,
				args:    []string{".", "Intf"},
			},
			{
				input:   "//counterfeiter:generate . Intf",
				matches: true,
				args:    []string{".", "Intf"},
			},
			{
				input:   "//go:generate  stringer -type=Enum",
				matches: false,
				args:    []string{".", "Intf"},
			},
		}
	})

	it.Focus("splits args correctly", func() {
		Expect(stringToArgs(". Intf")).To(ConsistOf([]string{"counterfeiter", ".", "Intf"}))
		Expect(stringToArgs("    .    Intf     ")).To(ConsistOf([]string{"counterfeiter", ".", "Intf"}))
	})

	it("matches lines appropriately", func() {
		for _, c := range cases {
			result := matchForString(c.input)
			if c.matches {
				Expect(result).NotTo(BeNil(), c.input)
				Expect(result.args).To(ConsistOf(c.args))
			} else {
				Expect(result).To(BeNil(), c.input)
			}
		}
	})
}
