package gstruct_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gstruct Suite")
}
