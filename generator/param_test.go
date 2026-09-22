package generator

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestParams(t *testing.T) {
	spec.Run(t, "Params", testParams, spec.Report(report.Terminal{}))
}

func testParams(t *testing.T, when spec.G, it spec.S) {
	g := NewWithT(t)

	when("listing the params as the fields of the recorded arguments struct", func() {
		it("returns nothing when there are no params", func() {
			g.Expect(Params{}.AsFieldsWithPrefix("argsForCall.")).To(Equal(""))
		})

		it("exports each param name and adds the prefix", func() {
			params := Params{{Name: "arg1", Type: "string"}}
			g.Expect(params.AsFieldsWithPrefix("argsForCall.")).To(Equal("argsForCall.Arg1"))
			g.Expect(params.AsFieldsWithPrefix("")).To(Equal("Arg1"))
		})

		it("separates several params with commas, variadic ones included", func() {
			params := Params{
				{Name: "arg1", Type: "int"},
				{Name: "arg2", Type: "...string", IsVariadic: true, IsSlice: true},
			}
			g.Expect(params.AsFieldsWithPrefix("fake.argsForCall[i].")).To(Equal("fake.argsForCall[i].Arg1, fake.argsForCall[i].Arg2"))
		})
	})
}
