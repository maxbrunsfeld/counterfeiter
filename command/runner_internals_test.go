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
				matches: false,
			},
			{
				input:   "//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Intf",
				matches: false,
			},
			{
				input:   "//counterfeiter:generate . Intf",
				matches: true,
				args:    []string{"counterfeiter", ".", "Intf"},
			},
			{
				input:   "//go:generate stringer -type=Enum",
				matches: false,
			},
		}
	})

	it("splits args correctly", func() {
		Expect(stringToArgs(". Intf")).To(ConsistOf([]string{"counterfeiter", ".", "Intf"}))
		Expect(stringToArgs("    .    Intf     ")).To(ConsistOf([]string{"counterfeiter", ".", "Intf"}))
	})

	it("matches lines appropriately", func() {
		for _, c := range cases {
			result, ok := matchForString(c.input)
			if c.matches {
				Expect(ok).To(BeTrue(), c.input)
				Expect(result).To(ConsistOf(c.args), c.input)
			} else {
				Expect(ok).To(BeFalse())
			}
		}
	})
}
