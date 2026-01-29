#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
#  ReserveChain DevNet - Node Launcher (Linux/macOS)
#  Starts a single DevNet node using the bundled Go runtime.
# -----------------------------------------------------------------------------

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo "==============================================="
echo "  ReserveChain DevNet - Node Starter (Unix)"
echo "==============================================="
echo "  Project root : $ROOT"
echo "  Go runtime   : $ROOT/runtime/go/bin/go"
echo "  Config file  : $ROOT/config/devnet.yaml"
echo "-----------------------------------------------"
echo "  Press Ctrl+C to stop the node cleanly."
echo "==============================================="
echo

if [ ! -x "$ROOT/runtime/go/bin/go" ]; then
  echo "[ERROR] Go runtime not found or not executable at:"
  echo "        $ROOT/runtime/go/bin/go"
  echo
  exit 1
fi

if [ ! -f "$ROOT/config/devnet.yaml" ]; then
  echo "[ERROR] Config file not found:"
  echo "        $ROOT/config/devnet.yaml"
  echo
  exit 1
fi

cd "$ROOT"
"$ROOT/runtime/go/bin/go" run ./cmd/node
