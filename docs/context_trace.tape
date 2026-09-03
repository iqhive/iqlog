# iqlog: "Request Signal Trace" (README VCR: CONTEXT_TRACE)
# Render: PATH=/tmp/opencode/bin:$PATH vhs docs/context_trace.tape
Output docs/context_trace.gif
Set FontSize 14
Set Width 860
Set Height 500
Set Theme { "name": "TRACE", "black": "#0a0e1a", "red": "#ff6b6b", "green": "#72f1a1", "yellow": "#f1b85b", "blue": "#96aaff", "magenta": "#96aaff", "cyan": "#67d9e9", "white": "#ebebe1", "brightBlack": "#5a6478", "brightRed": "#ff6b6b", "brightGreen": "#72f1a1", "brightYellow": "#f1b85b", "brightBlue": "#96aaff", "brightMagenta": "#96aaff", "brightCyan": "#67d9e9", "brightWhite": "#ebebe1", "background": "#0a0e1a", "foreground": "#d8d1b5", "cursor": "#72f1a1" }
Set Padding 24
Set LoopOffset 0
Set TypingSpeed 50ms
Set Shell bash

Hide
Type 'clear' Enter
Sleep 100ms
Type 'cd docs/demo-trace && go build -o /tmp/iqlog-demo-trace . && cd ../..' Enter
Sleep 1500ms
Show

Type '/tmp/iqlog-demo-trace' Enter
Sleep 2500ms

Hide
Type 'clear' Enter
Sleep 400ms
Show
