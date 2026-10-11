#!/usr/bin/env bash
# Regenerate the HTML pages from their Markdown sources.
#
#   docs/vision.md      -> manifesto.html
#   docs/technicals.md  -> technicals.html
#
#   ./tools/sync.sh           # rebuild
#   ./tools/sync.sh --check   # exit 1 if out of date (CI / pre-commit)
set -euo pipefail
cd "$(dirname "$0")/.."
exec python3 tools/build.py "$@"
