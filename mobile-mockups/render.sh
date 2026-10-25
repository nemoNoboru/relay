#!/usr/bin/env bash
# Render mockup screens to PNG (780x1688, a 390x844 phone at 2x).
# Usage: ./render.sh [screens/NN-name.html ...]
set -euo pipefail
cd "$(dirname "$0")"
mkdir -p png fonts
# Colour emoji for hosts that have none (25MB, not committed; see shared.css).
if [ ! -s fonts/NotoColorEmoji.ttf ]; then
  curl -sSL -o fonts/NotoColorEmoji.ttf "https://fonts.gstatic.com/s/notocoloremoji/v47/Yq6P-KqIXTD0t4D9z1ESnKM3-HpFab4.ttf"
fi
if [ $# -gt 0 ]; then files=("$@"); else files=(screens/*.html); fi
for f in "${files[@]}"; do
  name=$(basename "$f" .html)
  out="$PWD/png/$name.png"
  # Headless chromium's viewport is 87px shorter than --window-size, so ask
  # for 844+87 and trim the blank strip below the phone.
  chromium --headless --disable-gpu --hide-scrollbars --force-device-scale-factor=2 \
    --virtual-time-budget=4000 --window-size=390,931 --screenshot="$out" "file://$(realpath "$f")" >/dev/null 2>&1
  node crop.js "$out" 844 390
  echo "png/$name.png"
done
