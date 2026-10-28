#!/usr/bin/env sh
# Install the Relay node binary.
#
# One command:
#
#   curl --proto '=https' --tlsv1.2 -sSfL \
#     https://raw.githubusercontent.com/mansus-org/relay/main/install.sh | sh
#
# Downloads a prebuilt binary for your platform from GitHub Releases and
# installs it to ~/.local/bin (override with RELAY_INSTALL_DIR). When no
# release asset matches — e.g. before the first tagged release — it falls
# back to building from source if Git, Go, Node, npm, and make are present.
#
# Configuration via environment variables:
#   RELAY_INSTALL_DIR   directory to install into (default: ~/.local/bin,
#                       or /usr/local/bin when run as root)
#   RELAY_VERSION       release tag to install (default: latest)
set -eu

REPO="mansus-org/relay"

log() { printf '%s\n' "$*" >&2; }
die() { log "error: $*"; exit 1; }

# --- platform detection -----------------------------------------------------
os="$(uname -s)"
case "$os" in
  Linux)  os="linux" ;;
  Darwin) os="darwin" ;;
  *)      die "unsupported OS: $os (build from source instead, or run the container image)" ;;
esac
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) die "unsupported architecture: $arch (build from source instead, or run the container image)" ;;
esac

if [ "$(id -u)" = "0" ]; then
  INSTALL_DIR="${RELAY_INSTALL_DIR:-/usr/local/bin}"
else
  INSTALL_DIR="${RELAY_INSTALL_DIR:-$HOME/.local/bin}"
fi
VERSION="${RELAY_VERSION:-latest}"

# --- helpers -----------------------------------------------------------------
download() { # URL OUT
  url="$1"; out="$2"
  if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$out"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$out" "$url"
  else
    die "need curl or wget to download"
  fi
}

checksum_ok() { # FILE SHA256FILE
  file="$1"; sumfile="$2"
  expected="$(awk '{print $1}' "$sumfile" 2>/dev/null)" || return 1
  [ -n "$expected" ] || return 1
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "$file" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "$file" | awk '{print $1}')"
  else
    return 0 # no checksum tool available; skip verification
  fi
  [ "$actual" = "$expected" ]
}

build_from_source() {
  for tool in git make go node npm; do
    command -v "$tool" >/dev/null 2>&1 || die "no release binary for ${os}/${arch}, and '$tool' is missing (needed to build from source; or install via Docker)"
  done
  log "no release binary for ${os}/${arch} — building from source"
  git clone --depth 1 "https://github.com/${REPO}.git" "$tmp/src"
  (cd "$tmp/src" && make build)
  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$tmp/src/relay" "$INSTALL_DIR/relay"
}

# --- main --------------------------------------------------------------------
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

asset="relay-${os}-${arch}.tar.gz"
if [ "$VERSION" = "latest" ]; then
  base="https://github.com/${REPO}/releases/latest/download"
else
  base="https://github.com/${REPO}/releases/download/${VERSION}"
fi

if download "$base/$asset" "$tmp/$asset" 2>/dev/null; then
  if download "$base/$asset.sha256" "$tmp/$asset.sha256" 2>/dev/null; then
    if ! checksum_ok "$tmp/$asset" "$tmp/$asset.sha256"; then
      die "checksum mismatch for $asset"
    fi
  fi
  tar -xzf "$tmp/$asset" -C "$tmp"
  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$tmp/relay" "$INSTALL_DIR/relay"
else
  build_from_source
fi

"$INSTALL_DIR/relay" version || die "installed binary failed to run: $INSTALL_DIR/relay"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) log "note: add $INSTALL_DIR to your PATH to run 'relay'" ;;
esac
log "installed relay to $INSTALL_DIR/relay"
