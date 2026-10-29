package zxcvbn

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	corpusFile = "testdata/corpus.txt"

	// Set to 1 to regenerate the golden files instead of checking against them.
	writeGoldenEnvVar = "ZXCVBN_WRITE_GOLDEN"

	// Entropy and CrackTime may differ in the last bits across architectures and
	// toolchains (e.g. FMA on arm64).
	floatTolerance = 1e-9
)

// parityUserInputs exercises the user_inputs dictionary: case duplicates, non-ASCII
// words whose lowercase form has a different byte length, and invalid UTF-8.
var parityUserInputs = []string{"Horse", "horse", "ÇA", "ça", "straße", "İstanbul", "\xff", "a\xffb", "Ünïcødé", "pass", "PASSWORD", "x"}

var parityPasses = []struct {
	name       string
	golden     string
	userInputs []string
}{
	{"NoUserInputs", "testdata/parity_golden.json", nil},
	{"UserInputs", "testdata/parity_golden_user_inputs.json", parityUserInputs},
}

// goldenResult is the part of a PasswordStrength result that is pinned: all of it but
// the wall-clock CalcTime.
type goldenResult struct {
	Password         quoted
	Entropy          float64
	MatchSequence    []goldenMatch
	CrackTime        float64
	CrackTimeDisplay string
	Score            int
}

type goldenMatch struct {
	Pattern        string
	I, J           int
	Token          quoted
	DictionaryName string
	Entropy        float64
}

// quoted is stored Go-quoted, so that invalid UTF-8 survives JSON.
type quoted string

func (q quoted) MarshalText() ([]byte, error) {
	return []byte(strconv.Quote(string(q))), nil
}

func (q *quoted) UnmarshalText(b []byte) error {
	s, err := strconv.Unquote(string(b))
	*q = quoted(s)
	return err
}

func loadCorpus(tb testing.TB) []string {
	tb.Helper()
	raw, err := os.ReadFile(corpusFile)
	if err != nil {
		tb.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
}

// parityResult runs PasswordStrength. Spatial mac_keypad matches are recorded as
// keypad: see TestParity.
func parityResult(password string, userInputs []string) goldenResult {
	r := PasswordStrength(password, userInputs)
	g := goldenResult{
		Password:         quoted(r.Password),
		Entropy:          r.Entropy,
		CrackTime:        r.CrackTime,
		CrackTimeDisplay: r.CrackTimeDisplay,
		Score:            r.Score,
	}
	for _, m := range r.MatchSequence {
		if m.Pattern == "spatial" && m.DictionaryName == "mac_keypad" {
			m.DictionaryName = "keypad"
		}
		g.MatchSequence = append(g.MatchSequence, goldenMatch{m.Pattern, m.I, m.J, quoted(m.Token), m.DictionaryName, m.Entropy})
	}
	return g
}

// TestParity pins PasswordStrength over the corpus, without and with user inputs, to
// the results of upstream master 5e86766.
//
// PasswordStrength is not deterministic: the leet matcher iterates maps, and on master
// the dictionary matchers are also built in random map order. Both feed an unstable
// sort, so a spatial match that ties between the keypad and mac_keypad graphs reports
// whichever graph that order favours. Spatial keypad and mac_keypad matches are
// therefore both recorded as keypad. Every other field must match exactly, floats
// within floatTolerance.
//
// To regenerate, check out the commit that added this test (it runs master's matchers)
// and run, on arm64 (amd64 gives ULP-different floats, which floatTolerance covers):
//
//	ZXCVBN_WRITE_GOLDEN=1 go test -run '^TestParity$' .
//
// Regenerating on a later commit records that commit's own results instead.
func TestParity(t *testing.T) {
	corpus := loadCorpus(t)
	if os.Getenv(writeGoldenEnvVar) == "1" {
		writeGolden(t, corpus)
		return
	}

	for _, pass := range parityPasses {
		pass := pass
		t.Run(pass.name, func(t *testing.T) {
			golden := readGolden(t, pass.golden)
			if len(golden) != len(corpus) {
				t.Fatalf("%s has %d entries, corpus %d", pass.golden, len(golden), len(corpus))
			}
			for i, pw := range corpus {
				got := parityResult(pw, pass.userInputs)
				if !sameResult(golden[i], got) {
					gotJSON, _ := json.Marshal(got)
					wantJSON, _ := json.Marshal(golden[i])
					t.Errorf("%q:\n got  %s\n want %s", pw, gotJSON, wantJSON)
				}
			}
		})
	}
}

// Golden files hold one JSON result per corpus line.
func writeGolden(t *testing.T, corpus []string) {
	for _, pass := range parityPasses {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		for _, pw := range corpus {
			if err := enc.Encode(parityResult(pw, pass.userInputs)); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(pass.golden, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func readGolden(t *testing.T, path string) []goldenResult {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var golden []goldenResult
	for _, line := range bytes.Split(bytes.TrimSuffix(raw, []byte("\n")), []byte("\n")) {
		var r goldenResult
		if err := json.Unmarshal(line, &r); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		golden = append(golden, r)
	}
	return golden
}

func sameResult(a, b goldenResult) bool {
	if a.Password != b.Password || a.CrackTimeDisplay != b.CrackTimeDisplay || a.Score != b.Score ||
		!closeTo(a.Entropy, b.Entropy) || !closeTo(a.CrackTime, b.CrackTime) ||
		len(a.MatchSequence) != len(b.MatchSequence) {
		return false
	}
	for i, m := range a.MatchSequence {
		n := b.MatchSequence[i]
		if m.Pattern != n.Pattern || m.I != n.I || m.J != n.J || m.Token != n.Token ||
			m.DictionaryName != n.DictionaryName || !closeTo(m.Entropy, n.Entropy) {
			return false
		}
	}
	return true
}

func closeTo(a, b float64) bool {
	return a == b || math.Abs(a-b) <= floatTolerance*math.Max(math.Abs(a), math.Abs(b))
}
