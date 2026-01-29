#!/usr/bin/env bash
set -euo pipefail

# Auto-update script for ReserveChain Pi Main Server
# - pulls latest deployment repo
# - rebuilds node + workstation
# - refreshes web/marketing
# - restarts services

ROOT="/opt/reservechain"
REPO_DIR="$ROOT/repo"
BIN="$ROOT/bin"
WEB="$ROOT/web"
LOG="$ROOT/runtime/logs/update.log"

echo "=== UPDATE $(date) ===" >> "$LOG"

cd "$REPO_DIR"
git fetch origin
git reset --hard origin/main

# Refresh marketing webroot from repo
rm -rf "$WEB/marketing"
mkdir -p "$WEB/marketing"
cp -a "$REPO_DIR/web/marketing/." "$WEB/marketing/"

# Build node
cd "$REPO_DIR/core/node"
go mod tidy
mkdir -p "$BIN/tmp"
go build -o "$BIN/tmp/reservechain-node" ./cmd/node

# Atomic swap
mv "$BIN/reservechain-node" "$BIN/reservechain-node.prev" 2>/dev/null || true
mv "$BIN/tmp/reservechain-node" "$BIN/reservechain-node"
chmod 0755 "$BIN/reservechain-node"

# Build workstation (always rebuild so web dirs are precompiled)
if [[ -d "$REPO_DIR/core/workstation-src/workstation_portal" ]]; then
  cd "$REPO_DIR/core/workstation-src/workstation_portal"
  npm install
  npm run build
  rm -rf "$WEB/workstation"
  mkdir -p "$WEB/workstation"
  cp -a dist/. "$WEB/workstation/"
fi

systemctl restart reservechain-node
systemctl reload caddy

echo "Update complete" >> "$LOG"
