#!/usr/bin/env bash
#
# ReserveChain DevNet — Raspberry Pi 4 (Ubuntu Server) installer
#
# Target: Raspberry Pi 4 Model B, 8GB RAM, Ubuntu Server (64‑bit)
# This script:
#   - Installs basic dependencies (git, build tools, PHP, SQLite)
#   - Prepares runtime/ wrappers so the dev scripts can find Go/PHP/SQLite
#   - Does NOT overwrite any existing Go installation
#
set -euo pipefail

echo "============================================================"
echo " ReserveChain DevNet — Raspberry Pi 4 installer"
echo "============================================================"
echo

# Verify we are in project root (very simple heuristic).
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
  echo "This script must be run from the ReserveChain project root."
  echo "Example:"
  echo "  cd ~/ReserveChain-Devnet"
  echo "  bash install/linux/pi/install.sh"
  exit 1
fi

# Basic system info
ARCH=$(uname -m || echo "unknown")
OS=$(uname -s || echo "unknown")
echo "Detected OS:   ${OS}"
echo "Detected arch: ${ARCH}"
echo

if [ "${OS}" != "Linux" ]; then
  echo "Warning: this script is intended for Ubuntu Server on Raspberry Pi."
fi

echo "[1/4] Updating apt package index..."
sudo apt-get update -y

echo "[2/4] Installing base packages (git, build-essential, PHP, SQLite)..."
sudo apt-get install -y git build-essential php-cli php-common sqlite3 ca-certificates curl

echo
echo "[3/4] Preparing runtime/ directory structure..."
mkdir -p runtime/go/bin
mkdir -p runtime/php
mkdir -p runtime/sqlite

# Wire Go: prefer system Go if already installed.
if command -v go >/dev/null 2>&1; then
  echo "Found system Go at: $(command -v go)"
  ln -sf "$(command -v go)" runtime/go/bin/go
else
  echo "No system Go installation detected."
  echo "You can:"
  echo "  - Install Go via apt:   sudo apt-get install golang-go"
  echo "  - Or from go.dev (arm64 tarball), then symlink go into runtime/go/bin/go"
  echo
  echo "For now, creating a placeholder script that will error clearly if used."
  cat > runtime/go/bin/go << 'EOF'
#!/usr/bin/env bash
echo "Go compiler not found. Please install Go and update runtime/go/bin/go."
exit 1
EOF
  chmod +x runtime/go/bin/go
fi

# Wire PHP: wrapper to the system php binary.
if command -v php >/dev/null 2>&1; then
  PHP_BIN="$(command -v php)"
  echo "Using PHP at: ${PHP_BIN}"
  cat > runtime/php/php << EOF
#!/usr/bin/env bash
exec "${PHP_BIN}" "\$@"
EOF
  chmod +x runtime/php/php
else
  echo "WARNING: php not found in PATH. Please install php-cli."
fi

# Wire SQLite: wrapper to sqlite3.
if command -v sqlite3 >/dev/null 2>&1; then
  SQLITE_BIN="$(command -v sqlite3)"
  echo "Using sqlite3 at: ${SQLITE_BIN}"
  cat > runtime/sqlite/sqlite3 << EOF
#!/usr/bin/env bash
exec "${SQLITE_BIN}" "\$@"
EOF
  chmod +x runtime/sqlite/sqlite3
else
  echo "WARNING: sqlite3 not found in PATH. Please install sqlite3."
fi

echo
echo "[4/4] Marking helper scripts as executable..."
chmod +x scripts/*.sh || true

echo
echo "============================================================"
echo " Install complete (base setup)."
echo
echo "Next steps on Raspberry Pi:"
echo "  1) Start the DevNet node:"
echo "       ./scripts/start_node.sh"
echo "  2) Start the website/workstation:"
echo "       ./scripts/start_website.sh"
echo "  3) Or start both:"
echo "       ./scripts/start_all.sh"
echo
echo "If Go was not installed, install it and update runtime/go/bin/go accordingly."
echo "============================================================"
