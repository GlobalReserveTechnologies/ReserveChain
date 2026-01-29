#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
#  ReserveChain DevNet - Website / Workstation Launcher (Linux/macOS)
# -----------------------------------------------------------------------------

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

HOST="127.0.0.1"
PORT="8090"

echo "==============================================="
echo "  ReserveChain DevNet - Website Starter"
echo "==============================================="
echo "  Project root : $ROOT"
echo "  PHP runtime  : $ROOT/runtime/php/php"
echo "  Doc root     : $ROOT/public"
echo "  URL          : http://$HOST:$PORT/"
echo "-----------------------------------------------"
echo "  Press Ctrl+C to stop the server."
echo "==============================================="
echo

if [ ! -x "$ROOT/runtime/php/php" ]; then
  echo "[ERROR] PHP runtime not found or not executable at:"
  echo "        $ROOT/runtime/php/php"
  echo
  exit 1
fi

if [ ! -d "$ROOT/public" ]; then
  echo "[ERROR] Public web root not found:"
  echo "        $ROOT/public"
  echo
  exit 1
fi

cd "$ROOT"
"$ROOT/runtime/php/php" -S "$HOST:$PORT" -t public
