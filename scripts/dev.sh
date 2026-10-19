#!/usr/bin/env bash
# Build the web UI and the Relay binary, then run a local node.
#
#   ./scripts/dev.sh
#   NODE=my-group DATA=./data LISTEN=:8080 ./scripts/dev.sh
#   ADMIN_TOKEN=secret ./scripts/dev.sh
#
# This is the no-make fallback for `make dev`.
set -euo pipefail
cd "$(dirname "$0")/.."

NODE="${NODE:-my-group}"
DATA="${DATA:-./data}"
LISTEN="${LISTEN:-:8080}"

if [ ! -d web/node_modules ]; then
  echo "installing web dependencies"
  (cd web && npm ci --include=dev)
fi

echo "building web UI"
(cd web && npm run build)
touch web/dist/.gitkeep

echo "building relay"
go build -o relay ./cmd/relay

mkdir -p "$DATA"
args=(node --name "$NODE" --data "$DATA" --listen "$LISTEN")
if [ -n "${ADMIN_TOKEN:-}" ]; then
  args+=(--admin-token "$ADMIN_TOKEN")
fi

echo "starting relay on $LISTEN (node: $NODE, data: $DATA)"
exec ./relay "${args[@]}"
