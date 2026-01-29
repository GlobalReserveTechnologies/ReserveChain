# ReserveChain Pi Main Server (Ubuntu)

This repo is a **deployment repo** for running ReserveChain as a **single Pi main server + seed** on **Ubuntu Server (ARM64)**.

It includes:
- Marketing PHP site served from `web/marketing`
- Workstation SPA served from `web/workstation` (placeholder in repo; rebuilt on install/update)
- Go node source in `core/node` (rebuilt on install/update)
- Workstation source in `core/workstation-src/workstation_portal` (rebuilt on install/update)
- Caddy + PHP-FPM config + systemd units
- Auto-updates via systemd timer (`daily`)

## Install (fresh Ubuntu Pi)

```bash
sudo REPO_URL="https://github.com/YOU/ReserveChain-Pi-MainServer.git" bash scripts/installer.sh
```

Optionally override LAN subnet:
```bash
sudo REPO_URL="..." LAN_SUBNET="192.168.1.0/24" bash scripts/installer.sh
```

## URLs (LAN)

- `https://<pi-ip>/` marketing site
- `https://<pi-ip>/workstation/` workstation UI
- `https://<pi-ip>/api/*` node API (proxied to localhost)

## Notes

Caddy uses `tls internal` so your browser will show a warning until you trust the local CA.
