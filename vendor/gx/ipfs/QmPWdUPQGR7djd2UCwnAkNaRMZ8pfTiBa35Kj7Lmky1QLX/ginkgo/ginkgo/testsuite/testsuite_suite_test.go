package testsuite_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func TestTestsuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testsuite Suite")
}
