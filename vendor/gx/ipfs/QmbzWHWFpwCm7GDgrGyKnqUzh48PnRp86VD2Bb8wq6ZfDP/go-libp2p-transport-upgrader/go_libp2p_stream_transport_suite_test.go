package stream_test

import (
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"
	. "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega"

	"testing"
)

func TestGoLibp2pStreamTransport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoLibp2pStreamTransport Suite")
}
