package generator_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestTypeFor(t *testing.T) {
	spec.Run(t, "TypeFor", testTypeFor, spec.Report(report.Terminal{}))
}

func testTypeFor(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("generates types for basic types", func() {

	})
}
