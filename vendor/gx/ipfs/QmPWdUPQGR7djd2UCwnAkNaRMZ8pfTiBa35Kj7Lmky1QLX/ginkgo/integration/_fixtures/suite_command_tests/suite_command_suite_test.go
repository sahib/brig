package suite_command_test

import (
	. "github.com/onsi/gomega"
	. "gx/ipfs/QmPWdUPQGR7djd2UCwnAkNaRMZ8pfTiBa35Kj7Lmky1QLX/ginkgo"

	"testing"
)

func TestSuiteCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite Command Suite")
}
