package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega/matchers"
)

var _ = Describe("BeFalse", func() {
	It("should handle true and false correctly", func() {
		Ω(true).ShouldNot(BeFalse())
		Ω(false).Should(BeFalse())
	})

	It("should only support booleans", func() {
		success, err := (&BeFalseMatcher{}).Match("foo")
		Ω(success).Should(BeFalse())
		Ω(err).Should(HaveOccurred())
	})
})
