# iqlog: README hero: one log call, one colored console record
# Render: vhs docs/hero.tape
#
# The output line is produced by the real library at record time:
# docs/hero-demo is a tiny module against this repository. The tape builds it
# (hidden), then runs it on the recording TTY. iqlog detects the terminal and
# emits its own ANSI color codes, which VHS renders as color changes.
Output docs/hero.gif
Set FontSize 18
Set Width 820
Set Height 220
Set Theme { "name": "VCR", "black": "#090b0a", "red": "#ff6b6b", "green": "#72f1a1", "yellow": "#f1b85b", "blue": "#67a9f1", "magenta": "#72f1a1", "cyan": "#67d9e9", "white": "#d8d1b5", "brightBlack": "#6e7669", "brightRed": "#ff6b6b", "brightGreen": "#72f1a1", "brightYellow": "#f1b85b", "brightBlue": "#67a9f1", "brightMagenta": "#72f1a1", "brightCyan": "#67d9e9", "brightWhite": "#ebebe1", "background": "#090b0a", "foreground": "#d8d1b5", "cursor": "#72f1a1" }
Set Padding 28
Set LoopOffset 0
Set TypingSpeed 65ms
Set Shell bash

Hide
Type 'clear' Enter
Sleep 100ms
Type 'cd docs/hero-demo && go build -o /tmp/iqlog-hero-demo . && cd ../..' Enter
Sleep 2000ms
Show

# the log call you write (display only, cancel it, don't run it)
Type '# log.InfoEvent().Str("service", "checkout").Int("port", 8080).Bool("tls", true).Msg("listening")' Enter
Sleep 400ms

# run it: the library emits its own ANSI color codes onto the recording TTY
Type '/tmp/iqlog-hero-demo' Enter
Sleep 2000ms

Hide
Type 'clear' Enter
Sleep 300ms
Show
