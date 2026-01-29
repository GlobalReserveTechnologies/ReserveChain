
-- ReserveChain DevNet Database Schema (SQLite variant)
-- Engine: SQLite for DevNet / Workstation
-- For mainnet, a PostgreSQL-optimized variant should be generated from this base.

PRAGMA foreign_keys = ON;

----------------------------------------------------------------------
-- Core identity and accounts
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    external_id     TEXT UNIQUE,          -- e.g. auth provider id
    email           TEXT UNIQUE,
    display_name    TEXT,
    status          TEXT DEFAULT 'active', -- active / disabled / locked
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME
);

CREATE TABLE IF NOT EXISTS wallets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER,
    wallet_public   TEXT NOT NULL,        -- on-chain public key / address
    label           TEXT,
    is_primary      INTEGER DEFAULT 0,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_wallets_user ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallets_pub ON wallets(wallet_public);

----------------------------------------------------------------------
-- Vaults and balances
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS vaults (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id            TEXT UNIQUE NOT NULL,    -- logical id used by app (e.g. v1, v2)
    owner_wallet_id     INTEGER,                 -- non-custodial owner wallet
    label               TEXT,
    type                TEXT DEFAULT 'single',   -- single / multisig
    visibility_mode     TEXT DEFAULT 'A',        -- A/B/C/D per UI
    pnl_settlement_mode TEXT DEFAULT 'source',   -- source / target / etc.
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME,
    FOREIGN KEY (owner_wallet_id) REFERENCES wallets(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vaults_vault_id ON vaults(vault_id);

CREATE TABLE IF NOT EXISTS vault_balances (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id    INTEGER NOT NULL,
    asset_code  TEXT NOT NULL,      -- e.g. GRC, USD, BTC
    balance     REAL NOT NULL DEFAULT 0,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_balances_vault_asset
    ON vault_balances(vault_id, asset_code);

CREATE TABLE IF NOT EXISTS vault_yield_policies (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    vault_id        INTEGER NOT NULL,
    duration_tier   TEXT NOT NULL DEFAULT 'medium', -- short / medium / long
    sources         TEXT NOT NULL DEFAULT 'external,internal', -- CSV of sources
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_yield_policy_vault
    ON vault_yield_policies(vault_id);

CREATE TABLE IF NOT EXISTS vault_transfers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    from_vault_id   INTEGER,
    to_vault_id     INTEGER,
    asset_code      TEXT NOT NULL,
    amount          REAL NOT NULL,
    tx_ref          TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_vault_id) REFERENCES vaults(id) ON DELETE SET NULL,
    FOREIGN KEY (to_vault_id) REFERENCES vaults(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_vault_transfers_from ON vault_transfers(from_vault_id);
CREATE INDEX IF NOT EXISTS idx_vault_transfers_to   ON vault_transfers(to_vault_id);

----------------------------------------------------------------------
-- Stealth addresses (non-custodial privacy rail)
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS stealth_addresses (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    wallet_id       INTEGER NOT NULL,
    stealth_address TEXT NOT NULL,
    label           TEXT,
    is_active       INTEGER DEFAULT 1,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stealth_wallet ON stealth_addresses(wallet_id);

----------------------------------------------------------------------
-- Margin, orders, and positions
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS margin_sessions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      TEXT UNIQUE NOT NULL,
    user_id         INTEGER,
    wallet_id       INTEGER,
    vault_id        INTEGER,          -- collateral source
    status          TEXT NOT NULL,    -- open / closed / liquidated
    base_asset      TEXT NOT NULL,    -- e.g. BTC
    quote_asset     TEXT NOT NULL,    -- e.g. USDT
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL,
    FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_margin_sessions_user   ON margin_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_margin_sessions_wallet ON margin_sessions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_margin_sessions_vault  ON margin_sessions(vault_id);

CREATE TABLE IF NOT EXISTS orders (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id          TEXT UNIQUE NOT NULL,
    session_id        TEXT,             -- link to margin_sessions.session_id
    user_id           INTEGER,
    wallet_id         INTEGER,
    side              TEXT NOT NULL,    -- buy / sell
    type              TEXT NOT NULL,    -- market / limit / stop
    base_asset        TEXT NOT NULL,
    quote_asset       TEXT NOT NULL,
    quantity          REAL NOT NULL,
    price             REAL,
    status            TEXT NOT NULL,    -- open / filled / cancelled / rejected
    time_in_force     TEXT,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user       ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_session    ON orders(session_id);
CREATE INDEX IF NOT EXISTS idx_orders_status     ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_pair       ON orders(base_asset, quote_asset);

CREATE TABLE IF NOT EXISTS positions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id       TEXT UNIQUE NOT NULL,
    session_id        TEXT,
    user_id           INTEGER,
    wallet_id         INTEGER,
    vault_id          INTEGER,
    base_asset        TEXT NOT NULL,
    quote_asset       TEXT NOT NULL,
    side              TEXT NOT NULL,     -- long / short
    quantity          REAL NOT NULL,
    entry_price       REAL NOT NULL,
    liquidation_price REAL,
    pnl_realized      REAL DEFAULT 0,
    pnl_unrealized    REAL DEFAULT 0,
    status            TEXT NOT NULL,     -- open / closed / liquidated
    opened_at         DATETIME DEFAULT CURRENT_TIMESTAMP,
    closed_at         DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL,
    FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_positions_user   ON positions(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_wallet ON positions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_positions_vault  ON positions(vault_id);
CREATE INDEX IF NOT EXISTS idx_positions_pair   ON positions(base_asset, quote_asset);

----------------------------------------------------------------------
-- Reserve engine, tiers, and coverage
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS reserve_tiers (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    code            TEXT UNIQUE NOT NULL,  -- T1 / T2 / T3 / T4
    name            TEXT NOT NULL,         -- e.g. "High Liquidity"
    description     TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reserve_snapshots (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    snapshot_time   DATETIME NOT NULL,
    total_usd       REAL NOT NULL,
    tier_code       TEXT NOT NULL,
    notional_usd    REAL NOT NULL,
    notes           TEXT,
    FOREIGN KEY (tier_code) REFERENCES reserve_tiers(code) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reserve_snapshots_time ON reserve_snapshots(snapshot_time);

CREATE TABLE IF NOT EXISTS redemption_queue (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id      TEXT UNIQUE NOT NULL,
    user_id         INTEGER,
    wallet_id       INTEGER,
    vault_id        INTEGER,
    asset_code      TEXT NOT NULL,
    amount          REAL NOT NULL,
    status          TEXT NOT NULL,       -- pending / processing / completed / cancelled
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed_at    DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL,
    FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_redemption_status ON redemption_queue(status);

----------------------------------------------------------------------
-- Governance and monetary policy
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS governance_proposals (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    proposal_id     TEXT UNIQUE NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    status          TEXT NOT NULL,    -- draft / active / passed / rejected / executed
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME
);

CREATE TABLE IF NOT EXISTS governance_votes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    proposal_id     TEXT NOT NULL,
    voter_wallet_id INTEGER NOT NULL,
    choice          TEXT NOT NULL,    -- yes / no / abstain
    weight          REAL NOT NULL,    -- voting weight
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (proposal_id) REFERENCES governance_proposals(proposal_id) ON DELETE CASCADE,
    FOREIGN KEY (voter_wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_votes_proposal ON governance_votes(proposal_id);

CREATE TABLE IF NOT EXISTS issuance_schedule (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    epoch           INTEGER NOT NULL,
    planned_issuance REAL NOT NULL,
    actual_issuance  REAL,
    corridor_low     REAL,
    corridor_high    REAL,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_issuance_epoch ON issuance_schedule(epoch);

CREATE TABLE IF NOT EXISTS validators (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    validator_id    TEXT UNIQUE NOT NULL,
    wallet_id       INTEGER,
    stake_amount    REAL NOT NULL DEFAULT 0,
    status          TEXT NOT NULL,      -- active / inactive / jailed
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    FOREIGN KEY (wallet_id) REFERENCES wallets(id) ON DELETE SET NULL
);

----------------------------------------------------------------------
-- Audit and global settings
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS audit_events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type     TEXT NOT NULL,       -- vault / order / position / user / system
    entity_id       TEXT,
    event_type      TEXT NOT NULL,       -- created / updated / deleted / policy_change / etc.
    payload         TEXT,                -- JSON blob
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    actor           TEXT                 -- who performed (user id, system, etc.)
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_events(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_time   ON audit_events(created_at);

CREATE TABLE IF NOT EXISTS settings (
    key             TEXT PRIMARY KEY,
    value           TEXT,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);


-- =============================================
-- Tier & Earn extension
-- =============================================

-- ReserveChain Devnet – Tier & Earn Schema

CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    wallet_address  TEXT NOT NULL UNIQUE,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_tiers (
    user_id            INTEGER PRIMARY KEY,
    tier               TEXT NOT NULL,   -- 'core' | 'elite' | 'executive' | 'express'
    status             TEXT NOT NULL,   -- 'active' | 'grace' | 'frozen'
    billing_cycle      TEXT NOT NULL,   -- 'monthly' | 'quarterly' | 'yearly'
    current_renewal_tx TEXT,
    renew_expires_at   DATETIME NOT NULL,
    grace_ends_at      DATETIME,
    created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_tier_runtime (
    user_id             INTEGER PRIMARY KEY,
    collateral_mode     TEXT NOT NULL,   -- 'isolation' | 'protection'
    haircut_applied     REAL NOT NULL,
    margin_limit        REAL NOT NULL,
    pop_multiplier      REAL NOT NULL,
    staking_multiplier  REAL NOT NULL,
    last_trading_epoch  DATETIME,
    last_treasury_epoch DATETIME,
    last_staking_epoch  DATETIME,
    last_renewal_epoch  DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS earn_credits (
    user_id        INTEGER PRIMARY KEY,
    balance_grc    REAL NOT NULL DEFAULT 0,
    last_update_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS earn_ledger (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id        INTEGER NOT NULL,
    source_type    TEXT NOT NULL,   -- 'capital' | 'flow' | 'risk' | 'bonus'
    amount_grc     REAL NOT NULL,
    trading_epoch  INTEGER,
    treasury_epoch INTEGER,
    staking_epoch  INTEGER,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS tier_payments (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id             INTEGER NOT NULL,
    tier_before         TEXT NOT NULL,
    tier_after          TEXT NOT NULL,
    billing_cycle       TEXT NOT NULL,
    amount_grc          REAL NOT NULL,
    earn_applied_grc    REAL NOT NULL,
    stake_discount_grc  REAL NOT NULL,
    surplus_to_time_grc REAL NOT NULL,
    tx_hash             TEXT NOT NULL,
    created_at          DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

----------------------------------------------------------------------
-- Chain log (blocks + transactions) for DevNet
-- Canonical chain state is in the Go engine; this log is used for
-- replay and indexing. For mainnet, a PostgreSQL variant should
-- mirror this structure.
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS chain_blocks (
    height      INTEGER PRIMARY KEY,
    hash        TEXT NOT NULL,
    prev_hash   TEXT,
    timestamp   DATETIME NOT NULL,
    tx_type     TEXT NOT NULL,
    nonce       INTEGER NOT NULL,
    difficulty  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS chain_tx (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    block_height  INTEGER NOT NULL,
    tx_hash       TEXT NOT NULL,
    tx_type       TEXT NOT NULL,
    body_json     TEXT NOT NULL,
    FOREIGN KEY (block_height) REFERENCES chain_blocks(height) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chain_tx_hash  ON chain_tx(tx_hash);
CREATE INDEX IF NOT EXISTS idx_chain_tx_block       ON chain_tx(block_height);

-- Optional typed transaction tables for fast queries. These are
-- non-canonical projections over chain_tx and can be rebuilt from
-- the chain log if needed.

CREATE TABLE IF NOT EXISTS chain_tx_transfer (
    tx_id        INTEGER PRIMARY KEY,
    from_address TEXT NOT NULL,
    to_address   TEXT NOT NULL,
    asset        TEXT NOT NULL,
    amount       REAL NOT NULL,
    FOREIGN KEY (tx_id) REFERENCES chain_tx(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS chain_tx_vault (
    tx_id        INTEGER PRIMARY KEY,
    vault_id     TEXT NOT NULL,
    op_type      TEXT NOT NULL, -- CREATE / DEPOSIT / WITHDRAW / TRANSFER
    from_vault   TEXT,
    to_vault     TEXT,
    asset        TEXT,
    amount       REAL,
    FOREIGN KEY (tx_id) REFERENCES chain_tx(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS chain_tx_tier (
    tx_id        INTEGER PRIMARY KEY,
    wallet_addr  TEXT NOT NULL,
    tier_code    TEXT NOT NULL,
    periods      INTEGER NOT NULL,
    amount_grc   REAL NOT NULL,
    FOREIGN KEY (tx_id) REFERENCES chain_tx(id) ON DELETE CASCADE
);


----------------------------------------------------------------------
-- Node operators: work accounting and rewards
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS operator_nodes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id         TEXT UNIQUE NOT NULL,      -- logical node identifier
    hardware_json   TEXT,                      -- declared hardware profile + benchmark summary
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS operator_epoch_stats (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id             TEXT NOT NULL,         -- link to operator_nodes.node_id
    epoch_start         DATETIME NOT NULL,
    epoch_end           DATETIME NOT NULL,
    consensus_work      REAL NOT NULL DEFAULT 0,   -- A
    network_work        REAL NOT NULL DEFAULT 0,   -- B
    storage_work        REAL NOT NULL DEFAULT 0,   -- C
    service_work        REAL NOT NULL DEFAULT 0,   -- D
    hardware_cap        REAL NOT NULL DEFAULT 0,   -- capacity ceiling based on hybrid profiling
    work_score_raw      REAL NOT NULL DEFAULT 0,
    work_score_final    REAL NOT NULL DEFAULT 0,
    rewards_grc         REAL NOT NULL DEFAULT 0,
    metadata_json       TEXT                     -- optional extra detail
);

CREATE INDEX IF NOT EXISTS idx_operator_epoch_node ON operator_epoch_stats(node_id, epoch_start);


----------------------------------------------------------------------
-- RSX staking + validators (PoS) and Proof-of-Participation (PoP)
----------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS rsx_validators (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    validator_id     TEXT UNIQUE NOT NULL,
    operator_wallet  TEXT NOT NULL,            -- canonical wallet id (rc:... or evm:...)
    commission_bps   INTEGER NOT NULL DEFAULT 500, -- 5% default
    status           TEXT NOT NULL DEFAULT 'active', -- active / jailed / disabled
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME
);

CREATE INDEX IF NOT EXISTS idx_rsx_validators_wallet ON rsx_validators(operator_wallet);

CREATE TABLE IF NOT EXISTS rsx_stakes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    staker_wallet   TEXT NOT NULL,             -- canonical wallet id
    validator_id    TEXT NOT NULL,
    amount_rsx      REAL NOT NULL DEFAULT 0,
    lock_until_epoch INTEGER NOT NULL DEFAULT 0,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    UNIQUE(staker_wallet, validator_id)
);

CREATE INDEX IF NOT EXISTS idx_rsx_stakes_validator ON rsx_stakes(validator_id);
CREATE INDEX IF NOT EXISTS idx_rsx_stakes_staker ON rsx_stakes(staker_wallet);

-- Node registry for PoP payouts (operators may run many nodes).
CREATE TABLE IF NOT EXISTS pop_nodes (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id         TEXT UNIQUE NOT NULL,
    operator_wallet TEXT NOT NULL,            -- canonical wallet id
    role            TEXT,
    tx_hash         TEXT, -- optional: chain tx hash for idempotent replay
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME,
    UNIQUE(tx_hash)
);


CREATE INDEX IF NOT EXISTS idx_pop_nodes_operator ON pop_nodes(operator_wallet);

CREATE TABLE IF NOT EXISTS pop_node_caps (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    node_id         TEXT NOT NULL,
    cpu_score       REAL NOT NULL DEFAULT 1,
    ram_score       REAL NOT NULL DEFAULT 1,
    storage_score   REAL NOT NULL DEFAULT 1,
    bandwidth_score REAL NOT NULL DEFAULT 1,
    tx_hash         TEXT, -- optional: chain tx hash for idempotent replay
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(node_id),
    UNIQUE(tx_hash)
);


CREATE TABLE IF NOT EXISTS pop_epoch_metrics (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    epoch           INTEGER NOT NULL,
    node_id         TEXT NOT NULL,
    uptime_score    REAL NOT NULL DEFAULT 0,
    requests_served REAL NOT NULL DEFAULT 0,
    blocks_relayed  REAL NOT NULL DEFAULT 0,
    storage_io      REAL NOT NULL DEFAULT 0,
    latency_score   REAL NOT NULL DEFAULT 0,
    tx_hash         TEXT, -- optional: chain tx hash for idempotent replay
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tx_hash)
);

CREATE INDEX IF NOT EXISTS idx_pop_metrics_epoch ON pop_epoch_metrics(epoch);
CREATE INDEX IF NOT EXISTS idx_pop_metrics_node ON pop_epoch_metrics(node_id);
CREATE INDEX IF NOT EXISTS idx_pop_metrics_txhash ON pop_epoch_metrics(tx_hash);

CREATE TABLE IF NOT EXISTS epoch_payouts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    epoch           INTEGER NOT NULL,
    kind            TEXT NOT NULL,            -- 'stake' or 'pop' or 'treasury'
    recipient       TEXT NOT NULL,            -- canonical wallet id or 'treasury'
    asset_code      TEXT NOT NULL,            -- GRC / USDR
    amount          REAL NOT NULL DEFAULT 0,
    meta_json       TEXT,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_epoch_payouts_epoch ON epoch_payouts(epoch);
CREATE INDEX IF NOT EXISTS idx_epoch_payouts_recipient ON epoch_payouts(recipient);


-- Epoch payout commitment transactions (auditable commitments to payout ledger)
CREATE TABLE IF NOT EXISTS epoch_payout_commits (
    epoch INTEGER NOT NULL,
    tx_hash TEXT NOT NULL,
    author TEXT NOT NULL,
    payout_hash TEXT NOT NULL,
    num_payouts INTEGER NOT NULL,
    stake_budget_grc REAL NOT NULL,
    pop_budget_grc REAL NOT NULL,
    treasury_budget_grc REAL NOT NULL,
    created_at TEXT NOT NULL,
    UNIQUE(tx_hash)
);
CREATE INDEX IF NOT EXISTS idx_epoch_payout_commits_epoch ON epoch_payout_commits(epoch);


-- -------------------- slashing / evidence --------------------
CREATE TABLE IF NOT EXISTS slashing_events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    epoch           INTEGER NOT NULL,
    subject_type    TEXT NOT NULL,          -- 'validator' | 'pop_node' | 'wallet'
    subject_id      TEXT NOT NULL,          -- validator_id | node_id | canonical wallet id
    severity        TEXT NOT NULL,          -- 'info' | 'warn' | 'penalty' | 'critical'
    score           REAL NOT NULL DEFAULT 0,
    penalty_factor  REAL NOT NULL DEFAULT 0, -- 0..1, applied to rewards for the epoch (soft slashing)
    reason_code     TEXT NOT NULL,
    reason_detail   TEXT NOT NULL,
    evidence_json   TEXT NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'pending', -- 'pending' | 'applied' | 'dismissed'
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    applied_at      DATETIME
);
CREATE INDEX IF NOT EXISTS idx_slashing_events_epoch ON slashing_events(epoch);
CREATE INDEX IF NOT EXISTS idx_slashing_events_subject ON slashing_events(subject_type, subject_id);
