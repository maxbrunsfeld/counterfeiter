package arguments

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUsage(t *testing.T) {
	spec.Run(t, "Usage", testUsage, spec.Report(report.Terminal{}))
}

func testUsage(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("is reproduced verbatim in the README's command reference", func() {
		readme, err := os.ReadFile("../README.md")
		Expect(err).NotTo(HaveOccurred())

		_, section, found := strings.Cut(string(readme), "## Command reference\n")
		Expect(found).To(BeTrue(), "the README has no 'Command reference' section")
		_, block, found := strings.Cut(section, "```text\n")
		Expect(found).To(BeTrue(), "the command reference has no ```text block")
		block, _, found = strings.Cut(block, "\n```")
		Expect(found).To(BeTrue(), "the ```text block is not closed")

		Expect(block).To(Equal(strings.TrimSpace(usage)), "README command reference differs from the -help text in arguments/usage.go")
	})
}
