package iqlog

import (
	"math/rand"
	"testing"
	"time"
)

var testTime time.Time

func init() {
	testTime = time.Now()
}

func BenchmarkTimeAddTimeConsoleInPlaceCopy(b *testing.B) {
	b.StopTimer()
	byteSlice := [1024]byte{}
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		addTimeConsoleInPlaceCopy(timeNow, byteSlice[:])
	}
}

// commented out because it is slightly slower (+5% ns) than addTimeJSONInPlaceCopy
// func BenchmarkTimeAddTimeConsoleInPlaceOverwrite(b *testing.B) {
// 	b.StopTimer()
// 	byteSlice := [1024]byte{}
// 	timeNow := testTime
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		AddTimeConsoleInPlaceOverwrite(timeNow, byteSlice[:])
// 	}
// }

// func BenchmarkTimeConsoleAppendFormat(b *testing.B) {
// 	b.StopTimer()
// 	byteSlice := make([]byte, 0)
// 	timeNow := testTime
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		timeNow.AppendFormat(byteSlice, "[2006-01-02T15:04:05.999999] ")
// 	}
// }

func TestAddTimeConsoleInPlaceCopy(t *testing.T) {

	byteSlice := [1024]byte{}

	timeNow := testTime
	addTimeConsoleInPlaceCopy(timeNow, byteSlice[:])
	str := string(byteSlice[0:29])
	// the encoder always emits a fixed six-digit fraction, so compare against
	// the zero-padded format (".999999" trims trailing zeros and is flaky)
	if str != timeNow.Format("[2006-01-02T15:04:05.000000] ") {
		t.Errorf("addTimeConsoleInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("[2006-01-02T15:04:05.000000] "))
	}
}

func BenchmarkTimeAddTimeJSONInPlaceCopy(b *testing.B) {
	b.StopTimer()
	byteSlice := [1024]byte{}
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		addTimeJSONInPlaceCopy(timeNow, byteSlice[:])
	}
}

//
// commented out because it is slightly slower (+5% ns) than addTimeJSONInPlaceCopy
//
// func BenchmarkTimeAddTimeJSONInPlaceOverwrite(b *testing.B) {
// 	b.StopTimer()
// 	byteSlice := [1024]byte{}
// 	timeNow := testTime
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		AddTimeJSONInPlaceOverwrite(timeNow, byteSlice[:])
// 	}
// }

// func BenchmarkTimeTimeJSONAppendFormat(b *testing.B) {
// 	b.StopTimer()
// 	byteSlice := make([]byte, 0)
// 	timeNow := testTime
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		timeNow.AppendFormat(byteSlice, "{\"time\":\"2006-01-02T15:04:05.999999\",")
// 	}
// }

// func BenchmarkTimeTimeJSONAppendFormat2(b *testing.B) {
// 	b.StopTimer()
// 	byteSlice := make([]byte, 100)
// 	timeNow := testTime
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		timeNow.AppendFormat(byteSlice, "{\"time\":\"2006-01-02T15:04:05.999999\",")
// 	}
// }

// func TestAddTimeConsoleInPlaceOverwrite(t *testing.T) {

// 	byteSlice := [1024]byte{}

// 	timeNow := testTime
// 	AddTimeConsoleInPlaceOverwrite(timeNow, byteSlice[:])
// 	str := string(byteSlice[0:29])
// 	if str != timeNow.Format("[2006-01-02T15:04:05.999999] ") {
// 		t.Errorf("AddTimeConsoleInPlaceOverwrite() = |%v|, want |%v|", str, timeNow.Format("[2006-01-02T15:04:05.999999] "))
// 	}
// }

// func TestAddTimeJSONInPlaceCopySlow(t *testing.T) {

// 	byteSlice := [1024]byte{}

// 	timeNow := testTime
// 	AddTimeJSONInPlaceCopySlow(timeNow, byteSlice[:])
// 	str := string(byteSlice[0:37])
// 	if str != timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\",") {
// 		t.Errorf("TestAddTimeJSONInPlaceCopySlow() = |%v|, want |%v|", str, timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\","))
// 	}
// }

func TestAddTimeJSONInPlaceCopy(t *testing.T) {

	byteSlice := [1024]byte{}

	timeNow := testTime
	addTimeJSONInPlaceCopy(timeNow, byteSlice[:])
	str := string(byteSlice[0:37])
	// the encoder always emits a fixed six-digit fraction, so compare against
	// the zero-padded format (".999999" trims trailing zeros and is flaky)
	if str != timeNow.Format("{\"time\":\"2006-01-02T15:04:05.000000\",") {
		t.Errorf("TestAddTimeJSONInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("{\"time\":\"2006-01-02T15:04:05.000000\","))
	}
}

//
// commented out because it is slightly slower (+5% ns) than addTimeJSONInPlaceCopy
//
// func TestAddTimeJSONInPlaceOverwrite(t *testing.T) {
//
// 	byteSlice := [1024]byte{}
//
// 	timeNow := testTime
// 	AddTimeJSONInPlaceOverwrite(timeNow, byteSlice[:])
// 	str := string(byteSlice[0:37])
// 	if str != timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\",") {
// 		t.Errorf("addTimeConsoleInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\","))
// 	}
//
// }

// the fixed-width timestamp encoders must agree with time.Format
func TestAuditTimestampEncodersMatchFormat(t *testing.T) {
	check := func(tm time.Time) {
		t.Helper()
		var c [64]byte
		addTimeConsoleInPlaceCopy(tm, c[:])
		if got, want := string(c[:29]), tm.Format("[2006-01-02T15:04:05.000000] "); got != want {
			t.Errorf("console: time %v -> %q, want %q", tm, got, want)
		}
		var j [64]byte
		addTimeJSONInPlaceCopy(tm, j[:])
		if got, want := string(j[:37]), tm.Format(`{"time":"2006-01-02T15:04:05.000000",`); got != want {
			t.Errorf("json: time %v -> %q, want %q", tm, got, want)
		}
	}
	// boundaries
	for _, tm := range []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 12, 31, 23, 59, 59, 999999000, time.UTC),
		time.Date(2000, 2, 29, 12, 0, 0, 0, time.UTC),
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
		time.Date(1000, 10, 10, 10, 10, 10, 100000, time.UTC),
		time.Date(2026, 10, 20, 20, 20, 20, 202020000, time.UTC),
	} {
		check(tm)
	}
	// every month/day/hour/minute/second decade boundary
	for mo := 1; mo <= 12; mo++ {
		for _, d := range []int{1, 9, 10, 19, 20, 28} {
			for _, h := range []int{0, 9, 10, 19, 20, 23} {
				check(time.Date(2026, time.Month(mo), d, h, h%60, (h*7)%60, h*1000, time.UTC))
			}
		}
	}
	r := rand.New(rand.NewSource(3))
	for i := 0; i < 200000; i++ {
		tm := time.Unix(r.Int63n(253402300799), r.Int63n(1e9)).UTC()
		check(tm)
	}
}

// the per-second timestamp cache must track a moving clock
func TestAuditJSONTimestampCacheAcrossSeconds(t *testing.T) {
	var now time.Time
	sb := &syncBuffer{}
	l := MustNew(Config{Format: FormatJSON, Writer: sb, Level: LevelInfo,
		JSONTimeMode: JSONTimeUTC, Now: func() time.Time { return now }})
	base := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	for i := 0; i < 5000; i++ {
		now = base.Add(time.Duration(i) * 371 * time.Millisecond)
		sb.Reset()
		l.Info("m")
		want := now.UTC().Format(`{"time":"2006-01-02T15:04:05.000000Z"`)
		if got := sb.String(); len(got) < len(want) || got[:len(want)] != want {
			t.Fatalf("i=%d now=%v\n got %q\nwant prefix %q", i, now, got, want)
		}
	}
}
