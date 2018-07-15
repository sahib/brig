package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega/matchers"
)

var _ = Describe("BeTrue", func() {
	It("should handle true and false correctly", func() {
		Ω(true).Should(BeTrue())
		Ω(false).ShouldNot(BeTrue())
	})

	It("should only support booleans", func() {
		success, err := (&BeTrueMatcher{}).Match("foo")
		Ω(success).Should(BeFalse())
		Ω(err).Should(HaveOccurred())
	})
})
