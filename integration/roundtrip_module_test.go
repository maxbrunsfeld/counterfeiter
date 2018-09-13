// +build go1.11

package integration_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestRoundTripAsModule(t *testing.T) {
	spec.Run(t, "RoundTripAsModule", testRoundTripAsModule, spec.Report(report.Terminal{}))
}

func testRoundTripAsModule(t *testing.T, when spec.G, it spec.S) {
	it("is here so that you can comment out the runTests function below when focusing tests", func() {})
	runTests(false, t, when, it)
}
