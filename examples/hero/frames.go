package main

import (
	"bytes"
	"fmt"

	"strings"
	"time"

	"github.com/iqhive/banner"
	"github.com/iqhive/iqlog"
)

// All styling, wordmark layout, opening animation, SVG and playback live in banner.
// This file owns project text, deterministic demonstrations and reading holds.
type frame = banner.Frame

const (
	cols        = banner.Cols
	rows        = banner.Rows
	plain       = banner.Plain
	dim         = banner.Dim
	muted       = banner.Muted
	bright      = banner.Bright
	lit         = banner.Literal
	hot         = banner.Changed
	on          = banner.On
	off         = banner.Off
	bar         = banner.Bar
	prompt      = banner.Prompt
	caret       = banner.Caret
	full        = banner.Full
	shade       = banner.Shade
	projectName = "iqlog"
	tagline     = "Typed events for structured Go logs."
	subline     = "lightning fast · reflection-free typed fields · console + JSON + slog"
	install     = "go get github.com/iqhive/iqlog"
)

var points = []string{"lightning fast", "reflection-free typed fields", "console + JSON + slog"}

type film struct {
	cur, card frame
	frames    []frame
}

func (m *film) cut(ms int) {
	f := m.cur
	f.Hold = time.Duration(ms) * time.Millisecond
	m.frames = append(m.frames, f)
}

// build executes the project fixtures. The seed affects decorative digits;
// iqhash also uses it for the real demonstrated digests.
func build(seed uint64) []frame {
	m := &film{}
	opening, card, err := banner.Opening(banner.Card{
		Word: projectName, Tagline: tagline, Points: points, Install: install,
		Theme:    banner.IQCapsTheme(),
		Wordmark: banner.WordmarkOptions{Seed: seed}, ReadHold: 2500 * time.Millisecond,
		TaglineReveal: banner.TaglineWithReveal, PointInterval: 200 * time.Millisecond,
	})
	if err != nil {
		panic(err)
	}
	m.frames, m.card = opening, card
	m.structured()
	m.levels()
	m.cur = m.card
	m.cut(2500)
	m.cur = frame{HiveLogo: true}
	m.cur.Center(14, "IQ Hive", bright)
	m.cur.Center(16, "iqhive.com", muted)
	m.cut(1000)
	// Keep all scenes and quick motion; distribute extra reading time among
	// settled demonstration results and the final project card.
	extra := 25*time.Second - banner.Duration(m.frames)
	if extra < 0 {
		panic("hero: scene timing exceeds 25 seconds")
	}
	var holds []int
	for i := len(opening); i < len(m.frames)-1; i++ {
		if m.frames[i].Hold >= time.Second {
			holds = append(holds, i)
		}
	}
	for n, i := range holds {
		share := extra / time.Duration(len(holds)-n)
		m.frames[i].Hold += share
		extra -= share
	}
	return m.frames
}

// animation keeps generator failures visible instead of writing guessed output.
func animation(seed uint64) (a banner.Animation, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("hero fixture: %v", p)
		}
	}()
	frames := build(seed)
	a = banner.Animation{Frames: frames, StaticFrame: len(frames) - 2, Theme: banner.IQCapsTheme(),
		Title:       projectName + " — README demonstration",
		Description: "Typed fields produce a real JSON record; INFO and DEBUG gates emit different record counts.",
		Command:     "make hero",
	}
	return a, a.Validate()
}

// readyRecord calls the same chain displayed on screen. JSONTimeDisabled and
// DisableColor make the fixture independent of wall time and terminal state.
func readyRecord() string {
	var out bytes.Buffer
	log := iqlog.MustNew(iqlog.Config{Format: iqlog.FormatJSON, Writer: &out, DisableColor: true})
	defer func() {
		if err := log.Close(); err != nil {
			panic(err)
		}
	}()
	log.InfoEvent().Str("service", "api").Int("port", 8080).Bool("tls", true).Msg("ready")
	return strings.TrimSuffix(out.String(), "\n")
}

// typed draws code character by character, with a visible insertion cursor.
func (m *film) typed(x, y int, code string, ms int) {
	for i := 0; i <= len(code); i += 3 {
		m.cur.Clear(y, y+1)
		end := min(i, len(code))
		m.cur.Text(x, y, code[:end], lit)
		m.cur.Text(x+end, y, "▏", caret)
		m.cut(ms)
	}
	m.cur.Clear(y, y+1)
	m.cur.Text(x, y, code, lit)
}

func (m *film) structured() {
	m.cur = frame{}
	m.cur.Text(3, 1, "01 / typed fields → one JSON record", muted)
	m.cur.Text(3, 3, "// Format: iqlog.FormatJSON, Writer: &out", muted)
	m.typed(3, 5, "log.InfoEvent().", 60)
	m.typed(5, 6, `Str("service", "api").Int("port", 8080).`, 60)
	m.typed(5, 7, `Bool("tls", true).Msg("ready")`, 60)
	m.cut(250)
	record := readyRecord()
	m.cur.Text(3, 10, "↓ encoded by iqlog", muted)

	for i, at := range []int{strings.Index(record, `,"port"`), strings.Index(record, `,"tls"`), strings.Index(record, `,"message"`), len(record)} {
		m.cur.Clear(12, 13)
		m.cur.Text(3, 12, record[:at], plain)
		m.cur.Text(3, 14, strings.Repeat(string(full), 8*(i+1)), bar)
		if i == 3 {
			m.cur.Text(38, 14, "1 record · valid JSON", bright)
		}
		m.cur.Text(3, 16, "strings stay strings · integers and booleans stay typed", muted)
		if i == 3 {
			m.cut(1600)
		} else {
			m.cut(180)
		}
	}
}

var demoLevels = []iqlog.Level{iqlog.LevelDebug, iqlog.LevelInfo, iqlog.LevelWarn, iqlog.LevelError}

var demoMessages = []string{"cache probe", "listening", "pool low", "upstream failed"}

// gateRecords emits each fixture through one logger, then changes its real
// level gate using SetConfig. An empty captured record means it was filtered.
func gateRecords() [2][4]string {
	var result [2][4]string
	var out bytes.Buffer
	log := iqlog.MustNew(iqlog.Config{Writer: &out, Level: iqlog.LevelInfo, DisableColor: true})
	defer func() {
		if err := log.Close(); err != nil {
			panic(err)
		}
	}()
	for pass, gate := range []iqlog.Level{iqlog.LevelInfo, iqlog.LevelDebug} {
		cfg := log.Config()
		cfg.Level = gate
		if err := log.SetConfig(cfg); err != nil {
			panic(err)
		}
		for i, level := range demoLevels {
			out.Reset()
			log.Log(level, demoMessages[i])
			result[pass][i] = strings.TrimSuffix(out.String(), "\n")
		}
	}
	return result
}

func (m *film) levels() {
	records := gateRecords()
	for pass, gate := range []string{"Info", "Debug"} {
		m.cur = frame{}
		m.cur.Text(3, 1, "02 / change the level, change what gets through", muted)
		m.cur.Text(3, 3, "cfg := log.Config()", plain)
		m.cur.Text(3, 4, "cfg.Level = iqlog.Level"+gate, hot)
		m.cur.Text(3, 5, "log.SetConfig(cfg)", plain)
		m.cur.Text(3, 7, "INPUT", muted)
		m.cur.Text(27, 7, "GATE", muted)
		m.cur.Text(42, 7, "OUTPUT / console", muted)
		emitted, filtered := 0, 0
		for i, name := range []string{"DEBUG", "INFO", "WARN", "ERROR"} {
			y := 9 + 2*i
			m.cur.Text(3, y, name, plain)
			m.cur.Text(27, y, strings.ToUpper(gate), muted)
			for pos := range 3 {
				m.cur.Text(12, y, strings.Repeat(" ", 12), plain)
				m.cur.Text(12+4*pos, y, string(full)+string(full), on)
				m.cut(90)
			}
			m.cur.Text(12, y, strings.Repeat(" ", 12), plain)
			if records[pass][i] == "" {
				filtered++
				m.cur.Text(42, y, "filtered", dim)
				m.cur.Text(35, y, "×", muted)
			} else {
				emitted++
				m.cur.Text(35, y, "→", prompt)
				m.cur.Text(42, y, records[pass][i], plain)
			}
			m.cur.Clear(18, 19)
			m.cur.Text(3, 18, fmt.Sprintf("%d submitted   %d emitted   %d filtered", i+1, emitted, filtered), bright)
			m.cut(80)
		}
		if pass == 0 {
			m.cut(650)
		} else {
			m.cut(1050)
		}
	}
}
