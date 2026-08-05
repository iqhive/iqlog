package iqlog

import (
	"bytes"
	"testing"
	"time"
)

var testTime time.Time

func init() {
	testTime = time.Now()
}

func BenchmarkTimeAddTimeConsoleToBuffer(b *testing.B) {
	b.StopTimer()
	buffer := bytes.NewBuffer(make([]byte, 1000))
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		AddTimeConsoleToBuffer(timeNow, buffer)
	}
}

func BenchmarkTimeAddTimeConsoleAppend(b *testing.B) {
	b.StopTimer()
	emptySlice := make([]byte, 0)
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		emptySlice = emptySlice[:0]
		AddTimeConsoleAppend(timeNow, &emptySlice)
	}
}

func BenchmarkTimeAddTimeConsoleInPlaceCopy(b *testing.B) {
	b.StopTimer()
	byteSlice := [1024]byte{}
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		AddTimeConsoleInPlaceCopy(timeNow, byteSlice[:])
	}
}

// commented out because it is slightly slower (+5% ns) than AddTimeJSONInPlaceCopy
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
	AddTimeConsoleInPlaceCopy(timeNow, byteSlice[:])
	str := string(byteSlice[0:29])
	// the encoder always emits a fixed six-digit fraction, so compare against
	// the zero-padded format (".999999" trims trailing zeros and is flaky)
	if str != timeNow.Format("[2006-01-02T15:04:05.000000] ") {
		t.Errorf("AddTimeConsoleInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("[2006-01-02T15:04:05.000000] "))
	}
}

func BenchmarkTimeAddTimeJSONInPlaceCopy(b *testing.B) {
	b.StopTimer()
	byteSlice := [1024]byte{}
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		AddTimeJSONInPlaceCopy(timeNow, byteSlice[:])
	}
}

//
// commented out because it is slightly slower (+5% ns) than AddTimeJSONInPlaceCopy
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

func BenchmarkTimeAddTimeJSONToBuffer(b *testing.B) {
	b.StopTimer()
	buffer := bytes.NewBuffer(make([]byte, 1000))
	timeNow := testTime
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		AddTimeJSONToBuffer(timeNow, buffer)
	}
}

func BenchmarkTimeAddTimeJSONAppend(b *testing.B) {
	b.StopTimer()
	emptySlice := make([]byte, 0)
	timeNow := testTime
	b.StartTimer()
	_ = emptySlice
	_ = timeNow
	for i := 0; i < b.N; i++ {
		// emptySlice = make([]byte, 0)
		emptySlice = emptySlice[:0]
		AddTimeJSONAppend(timeNow, &emptySlice)
	}
}

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
	AddTimeJSONInPlaceCopy(timeNow, byteSlice[:])
	str := string(byteSlice[0:37])
	// the encoder always emits a fixed six-digit fraction, so compare against
	// the zero-padded format (".999999" trims trailing zeros and is flaky)
	if str != timeNow.Format("{\"time\":\"2006-01-02T15:04:05.000000\",") {
		t.Errorf("TestAddTimeJSONInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("{\"time\":\"2006-01-02T15:04:05.000000\","))
	}
}

//
// commented out because it is slightly slower (+5% ns) than AddTimeJSONInPlaceCopy
//
// func TestAddTimeJSONInPlaceOverwrite(t *testing.T) {
//
// 	byteSlice := [1024]byte{}
//
// 	timeNow := testTime
// 	AddTimeJSONInPlaceOverwrite(timeNow, byteSlice[:])
// 	str := string(byteSlice[0:37])
// 	if str != timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\",") {
// 		t.Errorf("AddTimeConsoleInPlaceCopy() = |%v|, want |%v|", str, timeNow.Format("{\"time\":\"2006-01-02T15:04:05.999999\","))
// 	}
//
// }
