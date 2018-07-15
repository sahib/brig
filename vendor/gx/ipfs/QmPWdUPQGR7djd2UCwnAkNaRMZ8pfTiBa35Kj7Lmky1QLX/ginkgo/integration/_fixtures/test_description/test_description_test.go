package test_description_test

import (
	"fmt"

	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
)

var _ = Describe("TestDescription", func() {
	It("should pass", func() {
		Ω(true).Should(BeTrue())
	})

	It("should fail", func() {
		Ω(true).Should(BeFalse())
	})

	AfterEach(func() {
		description := CurrentGinkgoTestDescription()
		fmt.Printf("%s:%t\n", description.FullTestText, description.Failed)
	})
})
