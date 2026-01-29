#!/usr/bin/env bash
set -euo pipefail

# -----------------------------------------------------------------------------
#  ReserveChain DevNet - Start All (Linux/macOS)
# -----------------------------------------------------------------------------

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "==============================================="
echo "  ReserveChain DevNet - Start All"
echo "==============================================="
echo "  This will launch node + website."
echo "==============================================="
echo

# Start node in background terminal if available, else just background
bash "$SCRIPT_DIR/start_node.sh" &
sleep 2
bash "$SCRIPT_DIR/start_website.sh"
