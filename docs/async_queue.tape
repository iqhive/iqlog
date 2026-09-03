# iqlog: "Buffer Under Pressure" (README VCR: ASYNC_QUEUE)
# Render: PATH=/tmp/opencode/bin:$PATH vhs docs/async_queue.tape
Output docs/async_queue.gif
Set FontSize 14
Set Width 880
Set Height 520
Set Theme { "name": "QUEUE", "black": "#101112", "red": "#ff6b6b", "green": "#72f1a1", "yellow": "#f1b85b", "blue": "#67d9e9", "magenta": "#f1b85b", "cyan": "#67d9e9", "white": "#ebebe1", "brightBlack": "#6e7669", "brightRed": "#ff6b6b", "brightGreen": "#72f1a1", "brightYellow": "#f1b85b", "brightBlue": "#67d9e9", "brightMagenta": "#f1b85b", "brightCyan": "#67d9e9", "brightWhite": "#ebebe1", "background": "#101112", "foreground": "#d8d1b5", "cursor": "#67d9e9" }
Set Padding 24
Set LoopOffset 0
Set TypingSpeed 45ms
Set Shell bash

Hide
Type 'clear' Enter
Sleep 100ms
Type 'cd docs/demo-queue && go build -o /tmp/iqlog-demo-queue . && cd ../..' Enter
Sleep 1500ms
Show

# segment 1: OVERFLOW BLOCK
Type 'echo "WRITER ASYNC // OVERFLOW BLOCK"' Enter
Sleep 300ms
Type '/tmp/iqlog-demo-queue block' Enter
Sleep 1500ms

# segment 2: OVERFLOW DROP
Type 'clear' Enter
Type 'echo "WRITER ASYNC // OVERFLOW DROP"' Enter
Sleep 300ms
Type '/tmp/iqlog-demo-queue drop' Enter
Sleep 1500ms

# segment 3: OVERFLOW SYNC
Type 'clear' Enter
Type 'echo "WRITER ASYNC // OVERFLOW SYNC"' Enter
Sleep 300ms
Type '/tmp/iqlog-demo-queue sync' Enter
Sleep 1500ms

# comparison card
Type 'clear' Enter
Sleep 200ms
Type 'echo "BLOCK = WAIT     DROP = COUNT LOSS     SYNC = CALLER I/O"' Enter
Sleep 1500ms
Type '' Enter

Hide
Type 'clear' Enter
Sleep 400ms
Show
