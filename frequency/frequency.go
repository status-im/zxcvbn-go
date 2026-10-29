// Package frequency gives read-only access to the ranked frequency lists the
// dictionary matchers use.
package frequency

import (
	"strings"

	"github.com/status-im/zxcvbn-go/internal/dictionary"
)

// Names returns the names of the frequency lists, in the order the dictionary matchers
// run.
func Names() []string {
	names := make([]string, len(dictionary.All))
	for i := range dictionary.All {
		names[i] = dictionary.All[i].Name
	}
	return names
}

// Lookup returns the rank of word, lowercased, in the named list, 1 being the most
// frequent.
func Lookup(list, word string) (rank int, ok bool) {
	for i := range dictionary.All {
		if d := &dictionary.All[i]; d.Name == list {
			return d.Lookup(strings.ToLower(word))
		}
	}
	return 0, false
}
