package acctest

import (
	"fmt"
	"testing"
)

func SkipNotImplemented(t *testing.T) {
	t.Skip("not implemented")
}

func SkipWhenOSEquals(t *testing.T, target Target, ids ...string) {
	for _, id := range ids {
		if id == target.Os.Id {
			t.Skip(fmt.Sprintf("test not relevant for os %q", target.Os.Id))
		}
	}
}

func SkipWhenOSNotEquals(t *testing.T, target Target, ids ...string) {
	for _, id := range ids {
		if id == target.Os.Id {
			return
		}
	}

	t.Skip(fmt.Sprintf("test not relevant for os %q", target.Os.Id))
}

func RelevantForOS(t *testing.T, target Target, ids ...string) {
	SkipWhenOSNotEquals(t, target, ids...)
}
