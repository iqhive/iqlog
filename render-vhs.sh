#!/usr/bin/env bash
# Render the README VCR animations with vhs.
#
# Each docs/*.tape builds its output from the real library (see
# docs/hero-demo) so the GIFs can never drift from actual iqlog output.
#
# vhs drives a headless Chromium. If /usr/bin/chromium-browser is a broken
# snap stub (common on Ubuntu), the script falls back to the Playwright
# Chromium cache and passes --no-sandbox automatically.
set -euo pipefail

cd "$(dirname "$0")"

# Locate a working Chromium for vhs (rod looks for chromium-browser / chrome).
if [ -z "${ROD_BROWSER_BIN:-}" ] && ! (command -v chromium-browser >/dev/null && chromium-browser --version >/dev/null 2>&1); then
    for candidate in \
        "$HOME/.cache/ms-playwright"/chromium-*/chrome-linux64/chrome \
        "$HOME/.cache/ms-playwright"/chromium-*/chrome-linux/chrome; do
        if [ -x "$candidate" ]; then
            mkdir -p .vhs-bin
            cat > .vhs-bin/chromium-browser <<SHIM
#!/bin/sh
exec "$candidate" --no-sandbox --disable-dev-shm-usage "\$@"
SHIM
            chmod +x .vhs-bin/chromium-browser
            PATH="$(pwd)/.vhs-bin:$PATH"
            echo "Using fallback Chromium: $candidate"
            break
        fi
    done
fi

command -v vhs >/dev/null || { echo "error: vhs is not installed" >&2; exit 1; }

tapes=("$@")
if [ ${#tapes[@]} -eq 0 ]; then
    tapes=(docs/*.tape)
fi

for tape in "${tapes[@]}"; do
    echo "Rendering $tape..."
    vhs "$tape"
done

echo "Done:"
ls -lh docs/*.gif
