package iqlog

import (
	"os/exec"
	"strings"
	"testing"
)

// The disabled path of the gated constructors is a load and a compare in the
// caller's frame only while the compiler inlines them. That depends on their
// inlining cost staying within budget, which is easy to lose in a refactor,
// so this asks the compiler directly. The same holds for Enabled, which the
// slog handler calls on every record.
func TestGateConstructorsInline(t *testing.T) {
	if testing.Short() {
		t.Skip("rebuilds the package with -gcflags=-m")
	}
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not found")
	}
	out, err := exec.Command(goTool, "build", "-gcflags=-m", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go build -gcflags=-m: %v\n%s", err, out)
	}
	report := string(out)
	for _, fn := range []string{
		"(*Logger).TraceEvent",
		"(*Logger).DebugEvent",
		"(*Logger).InfoEvent",
		"(*Logger).WarnEvent",
		"(*Logger).ErrorEvent",
		"(*Logger).Enabled",
	} {
		if !strings.Contains(report, "can inline "+fn+"\n") && !strings.Contains(report, "can inline "+fn+" ") {
			t.Errorf("%s is no longer inlinable; its disabled path now costs a call", fn)
		}
	}
}
