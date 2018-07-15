package testingtsupport_test

import (
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"

	"testing"
)

func TestTestingT(t *testing.T) {
	RegisterTestingT(t)
	Ω(true).Should(BeTrue())
}
