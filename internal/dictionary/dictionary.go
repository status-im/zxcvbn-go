// Package dictionary holds the ranked frequency lists as constant data, so matching
// against them allocates nothing and keeps nothing on the heap.
package dictionary

//go:generate go run ./gen

// Byte widths of the little-endian integers in Dictionary.ends (and firsts) and
// Dictionary.ranks.
const (
	endWidth  = 3
	rankWidth = 2
)

// Dictionary is a ranked word list. Words are lowercase and sorted bytewise.
type Dictionary struct {
	Name  string
	words string // all words concatenated
	ends  string // end offset of each word in words
	ranks string // rank of each word, 1 being the most frequent
	// firsts holds, for each byte value c and then 256, the index of the first word
	// starting with a byte >= c.
	firsts string
}

// Len returns the number of words.
func (d *Dictionary) Len() int {
	return len(d.ranks) / rankWidth
}

// Word returns the k-th word in sort order.
func (d *Dictionary) Word(k int) string {
	start := 0
	if k > 0 {
		start = d.end(k - 1)
	}
	return d.words[start:d.end(k)]
}

// Rank returns the rank of the k-th word.
func (d *Dictionary) Rank(k int) int {
	o := k * rankWidth
	return int(d.ranks[o]) | int(d.ranks[o+1])<<8
}

// LowerBound returns the index of the first word >= s at or after index lo, or Len().
func (d *Dictionary) LowerBound(s string, lo int) int {
	hi := d.Len()
	if len(s) > 0 {
		if first := d.first(int(s[0])); first > lo {
			lo = first
		}
		hi = d.first(int(s[0]) + 1)
	}
	for lo < hi {
		mid := int(uint(lo+hi) >> 1)
		if d.Word(mid) < s {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// Lookup returns the rank of word, which must be lowercase.
func (d *Dictionary) Lookup(word string) (rank int, ok bool) {
	if k := d.LowerBound(word, 0); k < d.Len() && d.Word(k) == word {
		return d.Rank(k), true
	}
	return 0, false
}

func (d *Dictionary) end(k int) int {
	return uint24(d.ends, k)
}

func (d *Dictionary) first(c int) int {
	return uint24(d.firsts, c)
}

func uint24(table string, k int) int {
	o := k * endWidth
	return int(table[o]) | int(table[o+1])<<8 | int(table[o+2])<<16
}
