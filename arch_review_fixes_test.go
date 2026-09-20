package iqlog

import (
	"io"
	"testing"
)

// R-01 regression: every writer replacement must clear native state, so the
// config flag and active routing never disagree. On Linux Config{NativeLog:true}
// cannot be built via SetConfig (prepareNativeLog returns unsupported), so the
// existing internal fake is required — mirroring native_log_test.go.
func TestArchReviewSameWriterReplacementClearsNativeLogFlag(t *testing.T) {
	l := MustNew(Config{Writer: io.Discard})
	fake := &fakeNativeWriter{}
	installFakeNative(l, fake)
	l.updateConfig(func(cfg *loggerConfig) { cfg.nativeLog = true })
	if !l.Config().NativeLog {
		t.Fatal("precondition failed: NativeLog flag not set")
	}
	if err := l.SetWriter(l.GetWriter()); err != nil {
		t.Fatal(err)
	}
	if l.Config().NativeLog {
		t.Error("SetWriter(GetWriter()) left NativeLog set in the configuration")
	}
}
