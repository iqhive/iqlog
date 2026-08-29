# iqlog — "Format Switch" (README VCR: CONSOLE_TO_JSON)
# Render: PATH=/tmp/opencode/bin:$PATH vhs docs/console_to_json.tape
Output docs/console_to_json.gif
Set FontSize 15
Set Width 820
Set Height 480
Set Theme { "name": "FORMAT", "black": "#0c0e0c", "red": "#ff6b6b", "green": "#72f1a1", "yellow": "#f1b85b", "blue": "#67d9e9", "magenta": "#f1b85b", "cyan": "#67d9e9", "white": "#ebebe1", "brightBlack": "#6e7669", "brightRed": "#ff6b6b", "brightGreen": "#72f1a1", "brightYellow": "#f1b85b", "brightBlue": "#67d9e9", "brightMagenta": "#f1b85b", "brightCyan": "#67d9e9", "brightWhite": "#ebebe1", "background": "#0c0e0c", "foreground": "#d8d1b5", "cursor": "#f1b85b" }
Set Padding 24
Set LoopOffset 0
Set TypingSpeed 60ms
Set Shell bash

Hide
Type 'clear' Enter
Sleep 100ms
Type 'cd docs/demo-format && go build -o /tmp/iqlog-demo-format . && cd ../..' Enter
Sleep 1500ms
Show

Type '/tmp/iqlog-demo-format' Enter
Sleep 2500ms

Hide
Type 'clear' Enter
Sleep 400ms
Show
