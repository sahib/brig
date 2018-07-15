package fail_fixture_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func TestFail_fixture(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Skip_fixture Suite")
}
