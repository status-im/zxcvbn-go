package zxcvbn

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

const dictHeapChildEnv = "ZXCVBN_DICT_HEAP_CHILD"

// Dictionaries are constant data: nothing they hold may live on the heap, neither after
// package init nor after PasswordStrength ran. Runs in a child process so
// GODEBUG=memprofilerate=1 records init-time allocations.
func TestDictionariesNotOnHeap(t *testing.T) {
	if os.Getenv(dictHeapChildEnv) == "1" {
		assertNoLiveAllocations(t, "zxcvbn-go/frequency.", "zxcvbn-go/internal/dictionary.")
		PasswordStrength("correcthorsebatterystaple", []string{"horse"})
		assertNoLiveAllocations(t, "zxcvbn-go/frequency.", "zxcvbn-go/internal/dictionary.", "matching.dictionaryMatch", "matching.buildRankedDict")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestDictionariesNotOnHeap$", "-test.v")
	cmd.Env = append(os.Environ(), dictHeapChildEnv+"=1", "GODEBUG=memprofilerate=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child failed: %v\n%s", err, out)
	}
}

func assertNoLiveAllocations(t *testing.T, frames ...string) {
	t.Helper()
	runtime.GC()
	runtime.GC()

	var records []runtime.MemProfileRecord
	n, _ := runtime.MemProfile(nil, false)
	for {
		records = make([]runtime.MemProfileRecord, n+50)
		var ok bool
		n, ok = runtime.MemProfile(records, false)
		if ok {
			records = records[:n]
			break
		}
	}

	live := make(map[string]int64)
	for _, r := range records {
		fs := runtime.CallersFrames(r.Stack())
		for {
			f, more := fs.Next()
			for _, name := range frames {
				if strings.Contains(f.Function, name) {
					live[f.Function] += r.InUseBytes()
				}
			}
			if !more {
				break
			}
		}
	}
	for name, bytes := range live {
		if bytes > 0 {
			t.Errorf("%d live bytes allocated by %s", bytes, name)
		}
	}
}
