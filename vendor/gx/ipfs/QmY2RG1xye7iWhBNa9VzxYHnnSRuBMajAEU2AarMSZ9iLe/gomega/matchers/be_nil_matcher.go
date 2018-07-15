package matchers

import "gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega/format"

type BeNilMatcher struct {
}

func (matcher *BeNilMatcher) Match(actual interface{}) (success bool, err error) {
	return isNil(actual), nil
}

func (matcher *BeNilMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be nil")
}

func (matcher *BeNilMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be nil")
}
