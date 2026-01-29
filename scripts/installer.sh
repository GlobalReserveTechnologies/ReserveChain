#!/usr/bin/env bash
set -euo pipefail

# ReserveChain Pi Main Server installer (Ubuntu Server ARM64)
# - installs deps (Go, Node, PHP-FPM, Caddy)
# - builds node + workstation
# - installs services + auto-updates
# - LAN-only firewall defaults

REPO_URL="${REPO_URL:-__REPO_URL__}"
LAN_SUBNET="${LAN_SUBNET:-192.168.1.0/24}"

ROOT="/opt/reservechain"
REPO_DIR="$ROOT/repo"
BIN="$ROOT/bin"
WEB="$ROOT/web"
RUNTIME="$ROOT/runtime"
SCRIPTS="$ROOT/scripts"

USER="reservechain"

GO_VERSION="1.22.6"
GO_TGZ="go${GO_VERSION}.linux-arm64.tar.gz"

echo "=== ReserveChain Installer (Main Server + Seed) ==="
echo "Repo: $REPO_URL"
echo "LAN:  $LAN_SUBNET"

if [[ "$REPO_URL" == "__REPO_URL__" ]]; then
  echo "ERROR: REPO_URL is not set. Run like:"
  echo "  sudo REPO_URL='https://github.com/YOU/ReserveChain-Pi-MainServer.git' bash installer.sh"
  exit 1
fi

sudo apt update
sudo apt upgrade -y
sudo apt install -y \
  git curl ca-certificates build-essential sqlite3 \
  ufw fail2ban \
  caddy \
  php-fpm php-cli php-sqlite3 php-curl php-mbstring php-xml \
  nodejs npm

sudo systemctl enable --now fail2ban

# Create service user
if ! id "$USER" >/dev/null 2>&1; then
  sudo adduser --system --group --home "$ROOT" "$USER"
fi

sudo mkdir -p "$ROOT" "$BIN" "$WEB" "$RUNTIME/data" "$RUNTIME/logs" "$SCRIPTS"
sudo chown -R "$USER:$USER" "$ROOT"

# Install Go (ARM64)
if ! command -v go >/dev/null 2>&1; then
  cd /tmp
  curl -LO "https://go.dev/dl/${GO_TGZ}"
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf "${GO_TGZ}"
  rm "${GO_TGZ}"
  echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null
fi
export PATH="$PATH:/usr/local/go/bin"

# Clone deployment repo
if [[ ! -d "$REPO_DIR/.git" ]]; then
  sudo -u "$USER" git clone --depth=1 "$REPO_URL" "$REPO_DIR"
else
  sudo -u "$USER" bash -lc "cd '$REPO_DIR' && git fetch origin && git reset --hard origin/main"
fi

# Install scripts
sudo cp "$REPO_DIR/scripts/update.sh" "$SCRIPTS/update.sh"
sudo chown "$USER:$USER" "$SCRIPTS/update.sh"
sudo chmod +x "$SCRIPTS/update.sh"

# Sync web roots from repo (marketing + placeholder workstation)
sudo rm -rf "$WEB/marketing" "$WEB/workstation"
sudo mkdir -p "$WEB/marketing" "$WEB/workstation"
sudo cp -a "$REPO_DIR/web/marketing/." "$WEB/marketing/"
sudo cp -a "$REPO_DIR/web/workstation/." "$WEB/workstation/"
sudo chown -R "$USER:$USER" "$WEB"

# Build node
sudo -u "$USER" bash -lc "cd '$REPO_DIR/core/node' && go mod tidy && go build -o '$BIN/reservechain-node' ./cmd/node"
sudo chown "$USER:$USER" "$BIN/reservechain-node"
sudo chmod 0755 "$BIN/reservechain-node"

# Build workstation (replaces placeholder)
if [[ -d "$REPO_DIR/core/workstation-src/workstation_portal" ]]; then
  sudo -u "$USER" bash -lc "cd '$REPO_DIR/core/workstation-src/workstation_portal' && npm install && npm run build"
  sudo rm -rf "$WEB/workstation"
  sudo mkdir -p "$WEB/workstation"
  sudo cp -a "$REPO_DIR/core/workstation-src/workstation_portal/dist/." "$WEB/workstation/"
  sudo chown -R "$USER:$USER" "$WEB/workstation"
fi

# PHP-FPM socket symlink (Ubuntu version dependent)
PHP_SOCK="$(ls /run/php/php*-fpm.sock 2>/dev/null | head -n 1 || true)"
if [[ -n "$PHP_SOCK" ]]; then
  sudo ln -sf "$PHP_SOCK" /run/php/php-fpm.sock
fi
sudo systemctl enable --now php8.1-fpm 2>/dev/null || true
sudo systemctl enable --now php8.2-fpm 2>/dev/null || true

# Install Caddy config
sudo cp "$REPO_DIR/system/caddy/Caddyfile" /etc/caddy/Caddyfile

# Install systemd units
sudo cp "$REPO_DIR/system/systemd/"*.service /etc/systemd/system/
sudo cp "$REPO_DIR/system/systemd/"*.timer /etc/systemd/system/

sudo systemctl daemon-reload
sudo systemctl enable --now reservechain-node
sudo systemctl enable --now reservechain-update.timer
sudo systemctl enable --now caddy

sudo systemctl restart reservechain-node
sudo systemctl restart caddy

# Firewall (LAN-only)
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from "$LAN_SUBNET" to any port 22 proto tcp
sudo ufw allow from "$LAN_SUBNET" to any port 80 proto tcp
sudo ufw allow from "$LAN_SUBNET" to any port 443 proto tcp
sudo ufw --force enable

echo ""
echo "✅ ReserveChain installed."
echo "Open (LAN):"
echo "  https://<PI-IP>/"
echo "  https://<PI-IP>/workstation/"
echo ""
echo "Logs:"
echo "  journalctl -u reservechain-node -f"
echo "  journalctl -u caddy -f"
