package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/iqhive/banner"
	"github.com/iqhive/iqlog"
)

// A parseable, complete record must appear verbatim in a frame. Level
// counts are compared to independently executed public API calls.
func TestFramesAreComputed(t *testing.T) {
	frames := build(0)
	var all strings.Builder
	for i := range frames {
		all.WriteString(banner.PlainText(&frames[i]))
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(readyRecord()), &record); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"level": "INFO", "service": "api", "port": float64(8080), "tls": true, "message": "ready"}
	if !reflect.DeepEqual(record, want) {
		t.Fatalf("record = %#v", record)
	}
	if !strings.Contains(all.String(), readyRecord()) {
		t.Fatal("no frame displays the entire JSON record")
	}
	for _, gate := range []iqlog.Level{iqlog.LevelInfo, iqlog.LevelDebug} {
		var out bytes.Buffer
		log := iqlog.MustNew(iqlog.Config{Writer: &out, Level: gate, DisableColor: true})
		for i, level := range demoLevels {
			log.Log(level, demoMessages[i])
		}
		if err := log.Close(); err != nil {
			t.Fatal(err)
		}
		n := strings.Count(out.String(), "\n")
		count := fmt.Sprintf("4 submitted   %d emitted   %d filtered", n, 4-n)
		if !strings.Contains(all.String(), count) {
			t.Errorf("no frame shows %q", count)
		}
		for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
			if !strings.Contains(all.String(), line) {
				t.Errorf("missing actual record %q", line)
			}
		}
	}
}

func TestHeroArtifact(t *testing.T) {
	a, err := animation(0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := animation(0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, again) {
		t.Fatal("generation is not deterministic")
	}
	if banner.Duration(a.Frames) != 25*time.Second {
		t.Fatal("loop must be 25 seconds")
	}
	if a.StaticFrame != len(a.Frames)-2 || !a.Frames[len(a.Frames)-1].HiveLogo {
		t.Fatal("final cards")
	}
	for _, want := range []string{tagline, subline, install} {
		if !strings.Contains(banner.PlainText(&a.Frames[a.StaticFrame]), want) {
			t.Fatalf("static card missing %q", want)
		}
	}
	var svg bytes.Buffer
	if err := banner.WriteSVG(&svg, a); err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("../../docs/hero.svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(svg.Bytes(), committed) {
		t.Fatal("stale hero: run make hero-svg from the repository root")
	}
	if svg.Len() >= 500*1024 {
		t.Fatal("hero exceeds 500 KiB")
	}
	decoder := xml.NewDecoder(bytes.NewReader(svg.Bytes()))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(svg.String(), "<style>") || !strings.Contains(svg.String(), "prefers-reduced-motion") {
		t.Fatal("SVG animation/fallback missing")
	}
	var out, errs bytes.Buffer
	if run(nil, &out, &errs) != 0 {
		t.Fatal(errs.String())
	}
	if out.String() != banner.PlainText(&a.Frames[a.StaticFrame]) {
		t.Fatal("redirected output is not static card")
	}
}

// The theme and raw frames must agree, including the sans-serif prefix.
func TestDefaultIQCapsWordmark(t *testing.T) {
	a, err := animation(0)
	if err != nil {
		t.Fatal(err)
	}
	if a.Theme.Name != "iqcaps" || !a.Theme.SansSerifI {
		t.Fatal("expected iqcaps with a serif-free I")
	}
	layout, err := banner.LayoutWordmark(strings.ToUpper(projectName), banner.WordmarkOptions{SansSerifI: true})
	if err != nil || layout.Gap < 1 {
		t.Fatal("uppercase layout does not fit with a gap", err)
	}
	card := a.Frames[a.StaticFrame]
	for _, c := range layout.Cells {
		cell := card.Cells[c.Y][c.X]
		if cell.Rune != c.Digit || cell.Style < banner.Mark {
			t.Fatal("missing uppercase glyph")
		}
		if (c.Letter < 2) != (cell.Style < banner.Mark+16) {
			t.Fatal("warm prefix and cool suffix styles overlap")
		}
	}
	actual := 0
	for _, row := range card.Cells {
		for _, c := range row {
			if c.Style >= banner.Mark {
				actual++
			}
		}
	}
	if actual != len(layout.Cells) {
		t.Fatal("unexpected wordmark cells")
	}
}
