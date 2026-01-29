# ReserveChain Devnet v1.0.0

This is an end-to-end devnet for ReserveChain, including:

- **Go backend node**
  - PoW-based devnet mining
  - Basic chain engine, mempool, difficulty adjustment
  - HTTP/WS APIs for wallet, vault, explorer and workstation
  - Follower + peer sync (HTTP-based P2P v1)
- **PHP + HTML/CSS/JS website**
  - Marketing / landing site
  - Documentation, governance, security and FAQ sections
  - Integrated workstation shell (single-page style app)
- **Workstation**
  - Sidebar-driven SPA-style UI
  - Wallet, vault, treasury, explorer and risk views
  - Hooks ready for on-chain modules (trading, tiers, PoP, etc.)
- **Runtime + scripts**
  - `runtime/` for devnet DB and external tools
  - `scripts/` for helper start/reset scripts (Windows and POSIX)
- **Database**
  - `database/schema.sql` with chain, wallet, vault and operator scaffolding

## Layout

- `cmd/node` — main Go node entrypoint
- `internal/core` — chain engine, blocks, mempool, work scoring
- `internal/net` — HTTP/WS server, follower + peer sync
- `internal/store` — SQLite persistence helpers
- `config/` — devnet YAML configuration
- `database/schema.sql` — SQLite schema for devnet
- `public/index.php` — main marketing + shell entry
- `public/sections/` — marketing + informational sections
- `public/workstation/` — workstation views (wallet, vault, treasury, etc.)
- `public/trading-terminal/` — placeholder for the trading terminal popup
- `public/assets/css/` — `site.css`, `workstation.css`
- `public/assets/js/` — site + workstation JS modules
- `runtime/` — devnet DB and runtime artifacts
- `scripts/` — platform helper scripts

## Running the node

From the project root:

```bash
go mod tidy
go run ./cmd/node
```

The node will start an HTTP/WS server on the configured port (default `:8080` in devnet).

## Running the PHP website + workstation

From the project root:

```bash
cd public
php -S 127.0.0.1:8000
```

Then open:

- `http://127.0.0.1:8000/` for the marketing site
- Workstation entries via the navigation (ACCOUNT / NETWORK / PRIVACY, etc.)
- Trading terminal launcher under the appropriate section

## Notes

- This is a **devnet** build intended for experimentation.
- Consensus, economics and node work scoring are under active iteration.
- New builds will bump the version number (v1.0.1, v1.0.2, ...) and update this README and the file tracker accordingly.


## Structure

### Web entrypoints

- `site/` — Friendly entrypoint for the marketing site. Internally forwards to `public/index.php`.
- `workstation/` — Friendly entrypoint for the workstation UI. Internally forwards to `public/workstation/index.php`.
- `public/` — Actual web root containing assets, sections, workstation, and trading terminal.

### Chain & runtime

- `cmd/node` — Go node entrypoint (`go run ./cmd/node`).
- `internal/` — Core chain logic (consensus, vaults, tiers, work scoring, etc.).
- `runtime/` — Local data, paths configuration, SQLite runtime.
- `database/schema.sql` — Database schema used by the PHP APIs and analytics.

### Frontend modes

- **Mode C (current)** — PHP + JS workstation living under `public/` and exposed via `workstation/`.
- **Mode B (planned)** — React + TypeScript SPA workstation to be scaffolded under `apps/workstation/` in a future build.


### Workstation SPA (v2.0.0-pre1)
- React + TypeScript SPA scaffolded in `apps/workstation/`.
- Uses Vite, Tailwind, and Zustand; currently renders placeholder panels matching the legacy PHP workstation sidebar.


### Workstation SPA Theme
- Default theme is dark with an extra-dim variant optimized for long trading sessions.
- Sidebar sections: Finance, Network, Vaults, Privacy, Account. Operator flows will live in a separate `Reserve Operator` app.



## Workstation Portal (Static)

The Workstation Portal SPA is served **statically by the Go node** at:

- `/workstation/`

Build it once:

```bash
./build.sh
```

Then start your node normally. The node will serve the built assets from:

- `apps/workstation_portal/dist`

You can override the dist folder with:

```bash
export RESERVECHAIN_WORKSTATION_DIST=/absolute/path/to/dist
```
