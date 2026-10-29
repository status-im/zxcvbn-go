package matching

import (
	"strings"
	"unicode/utf8"

	"github.com/status-im/zxcvbn-go/entropy"
	"github.com/status-im/zxcvbn-go/internal/dictionary"
	"github.com/status-im/zxcvbn-go/match"
)

func buildDictMatcher(dict *dictionary.Dictionary) func(password string) []match.Match {
	return func(password string) []match.Match {
		return dictionaryMatch(password, dict)
	}
}

func dictionaryMatch(password string, dict *dictionary.Dictionary) []match.Match {
	var results []match.Match
	pwLower := strings.ToLower(password)
	var boundsBuf [64]int
	bounds := runeBounds(pwLower, boundsBuf[:0])
	length := len(bounds) - 1
	var pwRunes []rune

	for i := 0; i < length; i++ {
		lo := 0
		for j := i; j < length; j++ {
			word := pwLower[bounds[i]:bounds[j+1]]
			// Words sharing the prefix are contiguous from the lower bound; none left means
			// no longer substring can match either.
			k := dict.LowerBound(word, lo)
			if k == dict.Len() || !strings.HasPrefix(dict.Word(k), word) {
				break
			}
			lo = k
			if dict.Word(k) == word {
				results = appendDictionaryMatch(results, password, &pwRunes, dict.Name, i, j, dict.Rank(k))
			}
		}
	}

	return results
}

func userInputsMatch(password string, rankedDict map[string]int) []match.Match {
	var results []match.Match
	pwLower := strings.ToLower(password)
	var boundsBuf [64]int
	bounds := runeBounds(pwLower, boundsBuf[:0])
	length := len(bounds) - 1
	var pwRunes []rune

	for i := 0; i < length; i++ {
		for j := i; j < length; j++ {
			if val, ok := rankedDict[pwLower[bounds[i]:bounds[j+1]]]; ok {
				results = appendDictionaryMatch(results, password, &pwRunes, "user_inputs", i, j, val)
			}
		}
	}

	return results
}

// runeBounds appends the byte offset of every rune of s, as []rune(s) splits it, and len(s).
func runeBounds(s string, bounds []int) []int {
	for i := 0; i < len(s); {
		bounds = append(bounds, i)
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
	}
	return append(bounds, len(s))
}

// I and J index runes of the lowercased password, which are the same runes of the
// original password: strings.ToLower maps each rune to exactly one rune and each invalid
// byte to one U+FFFD, as []rune does. Token keeps the original case.
func appendDictionaryMatch(results []match.Match, password string, pwRunes *[]rune, dictionaryName string, i, j, rank int) []match.Match {
	if *pwRunes == nil {
		*pwRunes = []rune(password)
	}
	matchDic := match.Match{Pattern: "dictionary",
		DictionaryName: dictionaryName,
		I:              i,
		J:              j,
		Token:          string((*pwRunes)[i : j+1]),
	}
	matchDic.Entropy = entropy.DictionaryEntropy(matchDic, float64(rank))
	return append(results, matchDic)
}

func buildRankedDict(unrankedList []string) map[string]int {

	result := make(map[string]int)

	for i, v := range unrankedList {
		result[strings.ToLower(v)] = i + 1
	}

	return result
}
