package iqlog

import (
	"encoding/binary"
	"math/rand"
	"testing"
)

// jsonNeedsEscapingReference is the byte-at-a-time definition the word scan
// must agree with.
func jsonNeedsEscapingReference(b []byte) bool {
	for _, c := range b {
		if jsonEscapeTable[c] {
			return true
		}
	}
	return false
}

// The word test must be exact: every byte value in every lane, with the
// other lanes clean, must be classified exactly as the table classifies it.
func TestJSONWordNeedsEscapingExhaustiveLanes(t *testing.T) {
	for lane := 0; lane < 8; lane++ {
		for c := 0; c < 256; c++ {
			var word [8]byte
			for i := range word {
				word[i] = 'a'
			}
			word[lane] = byte(c)
			got := jsonWordNeedsEscaping(binary.LittleEndian.Uint64(word[:]))
			if want := jsonEscapeTable[c]; got != want {
				t.Fatalf("lane %d byte %#x: word test = %v, table = %v", lane, c, got, want)
			}
		}
	}
}

// Random words with several suspect bytes, including ones that would
// borrow across lanes, and random slices of every length around the word
// boundary must agree with the reference scan.
func TestJSONNeedsEscapingMatchesReference(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	pool := []byte{0x00, 0x01, 0x1f, 0x20, 0x21, '"', '#', '[', '\\', ']', 'a', 'z', 0x7f, 0x80, 0xc2, 0xff}
	for i := 0; i < 200000; i++ {
		var word [8]byte
		for j := range word {
			if r.Intn(4) == 0 {
				word[j] = pool[r.Intn(len(pool))]
			} else {
				word[j] = byte(r.Intn(256))
			}
		}
		got := jsonWordNeedsEscaping(binary.LittleEndian.Uint64(word[:]))
		if want := jsonNeedsEscapingReference(word[:]); got != want {
			t.Fatalf("word %x: word test = %v, reference = %v", word, got, want)
		}
	}
	for i := 0; i < 20000; i++ {
		b := make([]byte, r.Intn(40))
		for j := range b {
			if r.Intn(16) == 0 {
				b[j] = pool[r.Intn(len(pool))]
			} else {
				b[j] = 'a' + byte(r.Intn(26))
			}
		}
		if got, want := jsonNeedsEscaping(b), jsonNeedsEscapingReference(b); got != want {
			t.Fatalf("%q: scan = %v, reference = %v", b, got, want)
		}
	}
}

func BenchmarkJSONNeedsEscaping(b *testing.B) {
	msg := []byte("The race was 16 minutes and 32 seconds long")
	b.Run("word", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if jsonNeedsEscaping(msg) {
				b.Fatal("clean message reported")
			}
		}
	})
	b.Run("table", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if jsonKeyNeedsEscaping(unsafeString(msg)) {
				b.Fatal("clean message reported")
			}
		}
	})
}
