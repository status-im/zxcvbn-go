package dictionary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var listFiles = map[string]string{
	"MaleNames":   "MaleNames.json",
	"FemaleNames": "FemaleNames.json",
	"Surname":     "Surnames.json",
	"English":     "English.json",
	"Passwords":   "Passwords.json",
}

// Table integrity only: the generated tables hold exactly the lowercased words of
// lists/*.json, sorted, each at the rank (index+1, last duplicate wins) the map-based
// dictionaries gave it. It re-derives what gen does, so matching parity with master is
// pinned separately, by TestParity in the root package.
func TestTablesMatchLists(t *testing.T) {
	for i := range All {
		d := &All[i]
		raw, err := os.ReadFile(filepath.Join("lists", listFiles[d.Name]))
		if err != nil {
			t.Fatal(err)
		}
		var list struct{ List []string }
		if err := json.Unmarshal(raw, &list); err != nil {
			t.Fatal(err)
		}
		want := make(map[string]int)
		for i, w := range list.List {
			want[strings.ToLower(w)] = i + 1
		}

		if d.Len() != len(want) {
			t.Fatalf("%s: %d words, want %d", d.Name, d.Len(), len(want))
		}
		for w, rank := range want {
			k := d.LowerBound(w, 0)
			if k == d.Len() || d.Word(k) != w {
				t.Fatalf("%s: %q not found", d.Name, w)
			}
			if d.Rank(k) != rank {
				t.Fatalf("%s: %q rank %d, want %d", d.Name, w, d.Rank(k), rank)
			}
		}
		for k := 1; k < d.Len(); k++ {
			if d.Word(k-1) >= d.Word(k) {
				t.Fatalf("%s: %q before %q", d.Name, d.Word(k-1), d.Word(k))
			}
		}
	}
}

func TestLowerBoundMisses(t *testing.T) {
	d := &All[3]
	if k := d.LowerBound("", 0); k != 0 {
		t.Errorf("empty: %d", k)
	}
	if k := d.LowerBound("\xff", 0); k != d.Len() {
		t.Errorf("past end: %d", k)
	}
	k := d.LowerBound("passwordx", 0)
	if k < d.Len() && d.Word(k) == "passwordx" {
		t.Errorf("unexpected hit")
	}
	if k > 0 && d.Word(k-1) >= "passwordx" {
		t.Errorf("not a lower bound: %q", d.Word(k-1))
	}
}

func TestLookupIsAllocationFree(t *testing.T) {
	d := &All[3]
	allocs := testing.AllocsPerRun(100, func() {
		k := d.LowerBound("horse", 0)
		_ = d.Word(k) == "horse" && d.Rank(k) > 0
	})
	if allocs != 0 {
		t.Errorf("%v allocs per lookup", allocs)
	}
}
