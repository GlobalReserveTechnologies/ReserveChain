#!/usr/bin/env bash
set -euo pipefail

echo "[1/2] Building Workstation Portal (static)..."
cd "$(dirname "$0")/apps/workstation_portal"
if command -v npm >/dev/null 2>&1; then
  npm install
  npm run build
else
  echo "npm not found. Please install Node.js + npm to build the workstation portal."
  exit 1
fi

echo "[2/2] Workstation build complete."
echo "Portal will be served by the node at: /workstation/"
echo "If you need a custom dist path, set RESERVECHAIN_WORKSTATION_DIST"
