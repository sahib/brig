package format_test

import (
	. "github.com/onsi/ginkgo"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"

	"testing"
)

func TestFormat(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Format Suite")
}
