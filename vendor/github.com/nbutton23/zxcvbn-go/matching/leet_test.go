package matching

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeetCanCreateSubstitutionMapsFromTable(t *testing.T) {
	table01 := map[string][]string{
		"a": []string{"@"},
		"b": []string{"8"},
		"g": []string{"6"},
	}

	table02 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
	}

	table03 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
	}

	table04 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
	}

	table05 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"l": []string{"!", "1", "|", "7"},
	}

	tests := []struct {
		name             string
		table            map[string][]string
		expectedListSize int
	}{
		{"Empty map generates an empty substitution map", map[string][]string{}, 0},
		{"Table with single values for every key returns only one substititution map", table01, 1},
		{"Table with two values on one key returns two substititution maps", table02, 2},
		{"Table with two values on two keys returns four substititution maps", table03, 4},
		{"Table should generate a substititution map with exponential variation according to table values (2*2*3 = 12)", table04, 12},
		{"Table should generate a substititution map with exponential variation according to table values (2*2*4 = 16)", table05, 16},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := createSubstitutionsMapsFromTable(test.table)
			assert.Equal(t, test.expectedListSize, len(result))
		})
	}
}

func TestSubstitutionWorksProperly(t *testing.T) {
	tests := []struct {
		name         string
		table        map[string]string
		word         string
		expectedWord string
	}{
		{"Word generated properly using substitution map", map[string]string{}, "password", "password"},
		{"Word generated properly using substitution map", map[string]string{}, "p@ssword", "p@ssword"},
		{"Word generated properly using substitution map", map[string]string{}, "p@$$w0rd", "p@$$w0rd"},
		{"Word generated properly using substitution map", map[string]string{}, "p@$$w0rD", "p@$$w0rD"},
		{"Word generated properly using substitution map", map[string]string{"a": "4"}, "p@ssword", "p@ssword"},
		{"Word generated properly using substitution map", map[string]string{"a": "@"}, "p@ssword", "password"},
		{"Word generated properly using substitution map", map[string]string{"a": "@"}, "p@$$word", "pa$$word"},
		{"Word generated properly using substitution map", map[string]string{"a": "@", "s": "$"}, "p@$$word", "password"},
		{"Word generated properly using substitution map", map[string]string{"a": "@", "s": "$", "o": "0"}, "p@$$w0rd", "password"},
		{"Word generated properly using substitution map", map[string]string{"a": "@", "s": "$", "o": "0"}, "p@$$w0rD", "passworD"},
		{"Word generated properly using substitution map", map[string]string{"i": "|"}, "|1|1|1|1|1|1|1|1|", "i1i1i1i1i1i1i1i1i"},
		{"Word generated properly using substitution map", map[string]string{"i": "1"}, "|1|1|1|1|1|1|1|1|", "|i|i|i|i|i|i|i|i|"},
		{"Word generated properly using substitution map", map[string]string{"i": "|", "l": "1"}, "|1|1|1|1|1|1|1|1|", "ilililililililili"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedWord, createWordForSubstitutionMap(test.word, test.table))
		})
	}
}

func TestLeetCanListConflictsOnTable(t *testing.T) {
	mapWithoutConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
	}

	mapWithOneConflictInTwoKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|"},
	}

	mapWithTwoConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
	}

	mapWithOneConflictInThreeKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1"},
	}

	regularMap := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"c": []string{"(", "{", "[", "<"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
		"o": []string{"0"},
		"s": []string{"$", "5"},
		"t": []string{"+", "7"},
		"x": []string{"%"},
		"z": []string{"2"},
	}

	tests := []struct {
		name         string
		table        map[string][]string
		expectedList []string
	}{
		{"Empty map generates an empty conflicts list", map[string][]string{}, []string{}},
		{"Map without conflicts generates an empty conflicts list", mapWithoutConflicts, []string{}},
		{"Map with one conflict generates the conflicts list properly", mapWithOneConflictInTwoKeys, []string{"1"}},
		{"Map with two conflicts generates the conflicts list properly", mapWithTwoConflicts, []string{"1", "|"}},
		{"Map with one conflict generates the conflicts list properly even for three keys", mapWithOneConflictInThreeKeys, []string{"1"}},
		{"Regular map generates the conflicts list properly", regularMap, []string{"1", "|", "7"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := retrieveConflictsListFromTable(test.table)

			// these sorts are necessary to make sure that the comparison will happen as expected
			sort.Strings(test.expectedList)
			sort.Strings(result)

			assert.Equal(t, test.expectedList, result)
		})
	}
}

func TestLeetCanListKeysWithSpecificValueOnTable(t *testing.T) {
	mapWithoutConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
	}

	mapWithOneConflictInTwoKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|"},
	}

	mapWithTwoConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
	}

	mapWithOneConflictInThreeKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1"},
	}

	regularMap := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"c": []string{"(", "{", "[", "<"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
		"o": []string{"0"},
		"s": []string{"$", "5"},
		"t": []string{"+", "7"},
		"x": []string{"%"},
		"z": []string{"2"},
	}

	tests := []struct {
		name         string
		table        map[string][]string
		valueToFind  string
		expectedList []string
	}{
		{"Empty map generates an empty list", map[string][]string{}, "@", []string{}},
		{"Map without conflicts returns only one representation for leet char", mapWithoutConflicts, "@", []string{"a"}},
		{"Map without conflicts returns no representation for unknown leet char", mapWithoutConflicts, "&", []string{}},
		{"Map with one conflict generates the list properly for conflicting value", mapWithOneConflictInTwoKeys, "1", []string{"i", "l"}},
		{"Map with one conflict generates the list properly for non conflicting value", mapWithOneConflictInTwoKeys, "|", []string{"l"}},
		{"Map with two conflicts generates the list properly for conflicting value (1)", mapWithTwoConflicts, "1", []string{"i", "l"}},
		{"Map with two conflicts generates the list properly for conflicting value (|)", mapWithTwoConflicts, "|", []string{"i", "l"}},
		{"Map with two conflicts generates the list properly for non conflicting value in conflicting key (i)", mapWithTwoConflicts, "!", []string{"i"}},
		{"Map with two conflicts generates the list properly for non conflicting value in conflicting key (l)", mapWithTwoConflicts, "7", []string{"l"}},
		{"Map with one conflict generates the list properly even for three keys", mapWithOneConflictInThreeKeys, "1", []string{"i", "l", "t"}},
		{"Regular map generates the list properly for conflicting value (|)", regularMap, "|", []string{"i", "l"}},
		{"Regular map generates the list properly for conflicting value (7)", regularMap, "7", []string{"l", "t"}},
		{"Regular map generates the list properly for non conflicting value", regularMap, "@", []string{"a"}},
		{"Regular map generates the list properly for unknown value", regularMap, "&", []string{}},
		{"Regular map generates the list properly for empty value", regularMap, "", []string{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := retrieveListOfKeysWithSpecificValueFromTable(test.table, test.valueToFind)

			// these sorts are necessary to make sure that the comparison will happen as expected
			sort.Strings(test.expectedList)
			sort.Strings(result)

			assert.Equal(t, test.expectedList, result)
		})
	}
}

func TestLeetCanCreateDifferentTablesForConflictingChar(t *testing.T) {
	mapWithoutConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
	}

	mapWithTwoConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
	}

	mapWithOneConflictInThreeKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1"},
	}

	tests := []struct {
		name                   string
		table                  map[string][]string
		leetChar               string
		expectedListOfMapsSize int
	}{
		{"Empty map must return no map on result", map[string][]string{}, "", 0},
		{"Map without conflicts generates the same map on result", mapWithoutConflicts, "4", 1},
		{"Map without conflicts generates no map on result for unknown char", mapWithoutConflicts, "&", 0},
		{"Map with two conflicts generates two maps using conflicting char", mapWithTwoConflicts, "|", 2},
		{"Map with two conflicts generates the same map using non conflicting char", mapWithTwoConflicts, "8", 1},
		{"Map with two conflicts generates no map using not existing char", mapWithTwoConflicts, "2", 0},
		{"Map with one conflict in three keys generates three maps", mapWithOneConflictInThreeKeys, "1", 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedListOfMapsSize, len(createDifferentMapsForLeetChar(test.table, test.leetChar)))
		})
	}
}

func TestLeetCanCreateTablesWithoutConflict(t *testing.T) {
	mapWithoutConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"e": []string{"3"},
		"g": []string{"6", "9"},
	}

	mapWithTwoConflicts := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
	}

	mapWithOneConflictInThreeKeys := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1"},
	}

	mapThatGeneratesTwelveOtherMaps := map[string][]string{
		"a": []string{"@", "4", "9"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1", "7"},
	}

	tests := []struct {
		name                   string
		table                  map[string][]string
		expectedListOfMapsSize int
	}{
		{"Empty map must return only the same map on result", map[string][]string{}, 1},
		{"Map without conflicts generates the same map on result", mapWithoutConflicts, 1},
		{"Map with two conflicts generates four maps", mapWithTwoConflicts, 4},
		{"Map with one conflict in three keys generates three maps", mapWithOneConflictInThreeKeys, 3},
		{"An specific map should generate twelve other maps", mapThatGeneratesTwelveOtherMaps, 12},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedListOfMapsSize, len(createListOfMapsWithoutConflicts(test.table)))
		})
	}
}

func TestStringSliceContainsValueFunction(t *testing.T) {
	tests := []struct {
		name           string
		slice          []string
		value          string
		expectedResult bool
	}{
		{"Empty slice contains no value", []string{}, "a", false},
		{"Empty slice contains no value, even if it is empty string", []string{}, "", false},
		{"Empty string is a valid value", []string{""}, "", true},
		{"Empty string is a valid value among others", []string{"0", "1", "", "a", "b"}, "", true},
		{"Empty string can not be found among others if it is not on list", []string{"0", "1", "a", "b"}, "", false},
		{"Slice with one value contains the stored value", []string{"a"}, "a", true},
		{"Slice with one value does not contain wrong value", []string{"a"}, "b", false},
		{"Slice with many items contains the stored value", []string{"a", "b", "c", "d", "0"}, "c", true},
		{"Slice with many items does not contain wrong value", []string{"a", "b", "c", "d", "0"}, "2", false},
		{"Slice with many items does not contain a value that looks like one that is stored", []string{"a", "b", "c", "d", "0"}, "C", false},
		{"Slice with many items does not contain a value that includes an space character", []string{"a", "b", "c", "d", "0"}, "c ", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedResult, stringSliceContainsValue(test.slice, test.value))
		})
	}
}

func TestCopyMapRemovingSameValueFromOtherKeysFunction(t *testing.T) {
	map01 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|"},
	}

	expectedMap01 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"|"},
	}

	map02 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!", "|"},
		"l": []string{"1", "|", "7"},
	}

	expectedMap02 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
	}

	map03 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"1", "!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+", "1"},
	}

	expectedMap03 := map[string][]string{
		"a": []string{"@", "4"},
		"b": []string{"8"},
		"g": []string{"6", "9"},
		"i": []string{"!"},
		"l": []string{"1", "|", "7"},
		"t": []string{"+"},
	}

	tests := []struct {
		name          string
		table         map[string][]string
		keyToFix      string
		valueToFix    string
		expectedTable map[string][]string
	}{
		{"Copy of map removing same value from other keys - test 01", map01, "i", "1", expectedMap01},
		{"Copy of map removing same value from other keys - test 02", map02, "l", "|", expectedMap02},
		{"Copy of map removing same value from other keys - test 03", map03, "l", "1", expectedMap03},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedTable, copyMapRemovingSameValueFromOtherKeys(test.table, test.keyToFix, test.valueToFix))
		})
	}
}

func TestLeetSubTable(t *testing.T) {
	subs := relevantL33tSubtable("password")
	assert.Len(t, subs, 0, "password should produce no leet subs")

	subs = relevantL33tSubtable("p4ssw0rd")
	assert.Len(t, subs, 2, "p4ssw0rd should produce 2 subs")

	subs = relevantL33tSubtable("1eet")
	assert.Len(t, subs, 2, "1eet should produce 2 subs")
	assert.Equal(t, subs["i"][0], "1")
	assert.Equal(t, subs["l"][0], "1")

	subs = relevantL33tSubtable("4pple@pple")
	assert.Len(t, subs, 1, "4pple@pple should produce 1 subs")
	assert.Len(t, subs["a"], 2)
}

func TestPermutationsOfLeetSubstitution(t *testing.T) {
	tests := []struct {
		name          string
		word          string
		expectedWords []string
	}{
		{"Permutation returns the expected list", "1337", []string{"1eel", "ieel", "ieet", "lee7", "leet"}},
		{"Permutation returns the expected list", "l33t", []string{"leet"}},
		{"Permutation returns the expected list", "password", []string{}},
		{"Permutation returns the expected list", "p@ssword", []string{"password"}},
		{"Permutation returns the expected list", "p@$$word", []string{"password"}},
		{"Permutation returns the expected list", "p@$$w0rd", []string{"password"}},
		{"Permutation returns the expected list", "p@4a$$w0rd", []string{"pa4assword", "p@aassword"}},
		{"Permutation returns the expected list", "|1|1|1|", []string{"i1i1i1i", "l1l1l1l", "|i|i|i|", "|l|l|l|", "lililil", "ililili"}},
		{"Permutation returns the expected list", "|1|1|@", []string{"i1i1ia", "l1l1la", "|i|i|a", "|l|l|a", "lilila", "ililia"}},
		{"Permutation returns the expected list", "1|1|@7", []string{"1i1ial", "1i1iat", "1l1la7", "1l1lat", "1|1|al", "ilila7", "ililat", "i|i|al", "i|i|at", "lilia7", "liliat", "l|l|a7", "l|l|at"}},
		{"Permutation returns the expected list", "1|1|@74", []string{"1i1ial4", "1i1iat4", "1l1la74", "1l1lat4", "1|1|al4", "ilila74", "ililat4", "i|i|al4", "i|i|at4", "lilia74", "liliat4", "l|l|a74", "l|l|at4", "1i1i@la", "1i1i@ta", "1l1l@7a", "1l1l@ta", "1|1|@la", "ilil@7a", "ilil@ta", "i|i|@la", "i|i|@ta", "lili@7a", "lili@ta", "l|l|@7a", "l|l|@ta"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getPermutations(test.word)

			// these sorts are necessary to make sure that the comparison will happen as expected
			sort.Strings(result)
			sort.Strings(test.expectedWords)

			assert.Equal(t, test.expectedWords, result)
		})
	}
}

func TestPermutationsLenOfLeetSubstitutions(t *testing.T) {
	// scenarios with just one possible substitution must return only one human readable value
	checkPermutationsLen(t, "p4ssw0rd", 1)
	checkPermutationsLen(t, "p4$sw0rd", 1)
	checkPermutationsLen(t, "p4$$w0rd", 1)
	checkPermutationsLen(t, "p@$$w0rd", 1)
	checkPermutationsLen(t, "l33t", 1)
	checkPermutationsLen(t, "p@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rdp@$$w0rd", 1)

	// number of variations are exponential if has more than one representation for same char
	checkPermutationsLen(t, "@", 1)
	checkPermutationsLen(t, "@4", 2)
	checkPermutationsLen(t, "@4(", 2)
	checkPermutationsLen(t, "@4({", 2*2)
	checkPermutationsLen(t, "@4({[", 2*3)
	checkPermutationsLen(t, "@4({[6", 2*3)
	checkPermutationsLen(t, "@4({[69", 2*3*2)

	// scenarios with no substitutions must return no result
	checkPermutationsLen(t, "test some good pass with this", 0)
	checkPermutationsLen(t, "no substitution should be made here", 0)
	checkPermutationsLen(t, "no substitution even with > . , ] } Âº # &", 0)
	checkPermutationsLen(t, "no SUBSTITUTION even with > . , ] } Âº # &", 0)

	// special characters without conflic should be replaced only by one human readable character
	checkPermutationsLen(t, "@@@@@@333333+++++(((((", 1)
	checkPermutationsLen(t, "@@@@@@333333+++++{{{{{", 1)
	checkPermutationsLen(t, "@@@@@@333333+++++[[[[[", 1)

	// if there is a conflicting character, consider all available options (but only one per time)
	checkPermutationsLen(t, "this_is_[[{{((hanging", 3)
	checkPermutationsLen(t, "p@$$w0rd 4", 2)
	checkPermutationsLen(t, "1337", 5)
	checkPermutationsLen(t, "7331", 5)
	checkPermutationsLen(t, "%6|a(+|<91%{s91![go0li0soe||", 56)                                                                      // it was an out of memory scenarion
	checkPermutationsLen(t, "|1|1|1|1|1|1|1|1|1|1|1|1|1|1|1|1|1|1|", 6)                                                              // it was an out of memory scenarion
	checkPermutationsLen(t, "7|1!7|1!7|1!7|1!7|1!69", 28)                                                                            // it was an out of memory scenarion
	checkPermutationsLen(t, "4@8({[<3691!|1|70$5+7%24@8({[<36", 544)                                                                 // the worst scenario
	checkPermutationsLen(t, "4@8({[<3691!|1|70$5+7%24@8({[<364@8({[<3691!|1|70$5+7%24@8({[<364@8({[<3691!|1|70$5+7%24@8({[<36", 544) // the worst scenario x3 times
	checkPermutationsLen(t, "p4$$w0rd @1!|", 14)
	checkPermutationsLen(t, "m&u]z^ou;\\!0o7t*x}uo)[s%kb618h#'gks|z\\!l3%8:z>pcq=!5w%\"%gs~@]5as`g&.'\\z4/\\`.nz$>yck.!%twu{})|%x8)$6\"", 112)
}

func checkPermutationsLen(t *testing.T, password string, permutationsLen int) {
	permutations := getPermutations(password)
	assert.Len(t, permutations, permutationsLen)
}

func TestLeet(t *testing.T) {
	password := "1337"
	matches := l33tMatch(password)
	bytes, _ := json.Marshal(matches)
	fmt.Println(string(bytes))
	fmt.Println(matches[0].J)
}
