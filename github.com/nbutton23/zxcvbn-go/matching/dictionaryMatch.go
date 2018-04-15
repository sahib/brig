package matching

import (
	"strings"

	"github.com/nbutton23/zxcvbn-go/entropy"
	"github.com/nbutton23/zxcvbn-go/match"
)

func buildDictMatcher(dictName string, rankedDict map[string]int) func(password string) []match.Match {
	return func(password string) []match.Match {
		matches := dictionaryMatch(password, dictName, rankedDict)
		for _, v := range matches {
			v.DictionaryName = dictName
		}
		return matches
	}

}

func dictionaryMatch(password string, dictionaryName string, rankedDict map[string]int) []match.Match {
	// length := len(password)
	var results []match.Match
	pwLower := strings.ToLower(password)

	// for i := 0; i < 0; i++ {
	// 	for j := i; j < length; j++ {
	// word := pwLower
	if val, ok := rankedDict[pwLower]; ok {
		matchDic := match.Match{Pattern: "dictionary",
			DictionaryName: dictionaryName,
			I:              0,
			J:              len(pwLower),
			Token:          pwLower,
		}
		matchDic.Entropy = entropy.DictionaryEntropy(matchDic, float64(val))

		results = append(results, matchDic)
	}
	//	}
	// }
	return results
}

func buildRankedDict(unrankedList []string) map[string]int {

	result := make(map[string]int)

	for i, v := range unrankedList {
		result[strings.ToLower(v)] = i + 1
	}

	return result
}
