package counterfeiter_test

import (
	"github.com/maxbrunsfeld/counterfeiter/fixtures/fakes"
	"github.com/maxbrunsfeld/counterfeiter/fixtures/interfaces"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
  "errors"
)

var _ = Describe("Generated fakes", func() {
  It("implements the interface", func() {
    var fake interfaces.SomeInterface
    fake = fakes.NewFakeSomeInterface()
    Expect(fake).NotTo(BeNil())
  })

  It("can have its behavior configured", func() {
    var fake *fakes.FakeSomeInterface
    fake = fakes.NewFakeSomeInterface()

    fake.Method1_ = func(arg1 string, arg2 uint64) error {
      Expect(arg1).To(Equal("stuff"))
      Expect(arg2).To(Equal(5))
      return errors.New("hi")
    }

    ret := fake.Method1("stuff", 5)

    Expect(ret).To(Equal(errors.New("hi")))
  })
})
