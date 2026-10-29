package frequency

import (
	"reflect"
	"testing"
)

func TestNames(t *testing.T) {
	want := []string{"MaleNames", "FemaleNames", "Surname", "English", "Passwords"}
	if got := Names(); !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

func TestLookup(t *testing.T) {
	for _, tc := range []struct {
		list, word string
		rank       int
		ok         bool
	}{
		{"Passwords", "password", 1, true},
		{"Passwords", "PassWord", 1, true},
		{"English", "you", 1, true},
		{"Passwords", "passwordx", 0, false},
		{"Passwords", "", 0, false},
		{"Unknown", "password", 0, false},
	} {
		rank, ok := Lookup(tc.list, tc.word)
		if rank != tc.rank || ok != tc.ok {
			t.Errorf("Lookup(%q, %q) = %d, %v, want %d, %v", tc.list, tc.word, rank, ok, tc.rank, tc.ok)
		}
	}
}
