package matchers

import (
	"fmt"
	"os"

	"gx/ipfs/QmY2RG1xye7iWhBNa9VzxYHnnSRuBMajAEU2AarMSZ9iLe/gomega/format"
)

type BeAnExistingFileMatcher struct {
	expected interface{}
}

func (matcher *BeAnExistingFileMatcher) Match(actual interface{}) (success bool, err error) {
	actualFilename, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("BeAnExistingFileMatcher matcher expects a file path")
	}

	if _, err = os.Stat(actualFilename); err != nil {
		switch {
		case os.IsNotExist(err):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func (matcher *BeAnExistingFileMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("to exist"))
}

func (matcher *BeAnExistingFileMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, fmt.Sprintf("not to exist"))
}