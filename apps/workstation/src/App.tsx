import React from 'react';
import {{ Routes, Route, Navigate, Link, useLocation }} from 'react-router-dom';

type PanelDef = {{
  id: string;
  label: string;
}};

type FinanceGroupDef = {{
  id: string;
  label: string;
  panels: PanelDef[];
}};

type SectionDef = {{
  id: string;
  label: string;
  panels: PanelDef[];
}};

const FINANCE_GROUPS: FinanceGroupDef[] = [
      { id: 'trading', label: 'Trading', panels: [
        { id: 'trading-terminal', label: '' },
        { id: 'trading-swap', label: '' },
        { id: 'defi-orders-fills', label: '' },
        { id: 'defi-activity', label: '' }
      ] },
      { id: 'account', label: 'Account', panels: [
        { id: 'defi-wallet', label: '' },
        { id: 'defi-portfolio', label: '' },
        { id: 'defi-deposit-withdraw', label: '' },
        { id: 'defi-transfers', label: '' }
      ] },
      { id: 'defi', label: 'DeFi', panels: [
        { id: 'defi-lending', label: '' },
        { id: 'defi-staking', label: '' },
        { id: 'defi-liquidity', label: '' }
      ] }
];

const OTHER_SECTIONS: SectionDef[] = [
  { id: 'network', label: 'Network', panels: [
      { id: 'network-explorer', label: 'Explorer (coming soon)' },
      { id: 'treasury-overview', label: 'Treasury / Reserves' },
      { id: 'network-coverage', label: 'Reserve Coverage' },
      { id: 'treasury-mints', label: 'Mint Queue' },
      { id: 'treasury-redemptions', label: 'Redemption Queue' },
      { id: 'treasury-grc-policy', label: 'GRC Issuance Policy' },
      { id: 'network-epochs', label: 'Epoch Control' },
      { id: 'network-history-supply', label: 'Supply Over Time' },
      { id: 'network-history-equity', label: 'Equity Over Time' },
      { id: 'network-history-grc', label: 'GRC Policy Output' },
      { id: 'network-staking', label: 'Staking (RSX)' },
      { id: 'network-pop', label: 'Operator PoP Rewards' },
      { id: 'network-benchmarks', label: 'Benchmarks (coming soon)' },
      { id: 'network-health', label: 'Network Health (coming soon)' }
  ] },  { id: 'vaults', label: 'Vaults', panels: [
      { id: 'vault-dashboard', label: '' },
      { id: 'vault-stealth', label: '' },
      { id: 'vault-policies', label: '' },
      { id: 'vault-roles', label: '' },
      { id: 'vault-analytics', label: '' },
      { id: 'vault-audit', label: '' },
      { id: 'vault-settings', label: '' }
  ] },
  { id: 'privacy', label: 'Privacy', panels: [
      { id: 'privacy-private-amm', label: '' },
      { id: 'privacy-private-lending', label: '' },
      { id: 'privacy-audit-proofs', label: '' }
  ] },
  { id: 'account', label: 'Account', panels: [
      { id: 'account-profile', label: '' },
      { id: 'account-api-keys', label: '' },
      { id: 'account-security', label: '' },
      { id: 'account-tier-billing', label: '' }
  ] }
];

type FinanceOpenState = {{
  [groupId: string]: boolean;
}};

const FINANCE_STATE_KEY = 'rc_finance_sidebar_state_v1';

function loadInitialFinanceState(): FinanceOpenState {{
  try {{
    const raw = window.localStorage.getItem(FINANCE_STATE_KEY);
    if (!raw) return {{ trading: true, account: true, defi: true }};
    const parsed = JSON.parse(raw);
    return {{
      trading: parsed.trading ?? true,
      account: parsed.account ?? true,
      defi: parsed.defi ?? true,
    }};
  }} catch {{
    return {{ trading: true, account: true, defi: true }};
  }}
}}

function panelIdToPath(id: string): string {{
  if (id.startsWith('trading-') || id.startsWith('defi-')) return '/finance/' + id;
  if (id.startsWith('vault-')) return '/vaults/' + id.replace('vault-', '');
  if (id.startsWith('privacy-')) return '/privacy/' + id.replace('privacy-', '');
  if (id.startsWith('network-') || id === 'treasury-overview') return '/network/' + id;
  if (id.startsWith('account-')) return '/account/' + id.replace('account-', '');
  return '/' + id;
}}

const PanelPlaceholder: React.FC<{ id: string; label: string }> = ({ id, label }) => {
  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">{label}</h1>
        <span className="rc-ws-panel-subtitle">SPA panel <code>{id}</code></span>
      </div>
      <div className="rc-ws-panel-body">
        <p>This panel is wired in the SPA shell and ready to be connected to live ReserveChain data.</p>
      </div>
    </div>
  );
};


const TreasuryOverviewPanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [snap, setSnap] = React.useState<any | null>(null);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/treasury');
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = await res.json();
        if (!cancelled) {
          setSnap(data);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load treasury snapshot');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();

    const id = window.setInterval(load, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };




const GRCPolicyPanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [sig, setSig] = React.useState<any | null>(null);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/grc-issuance');
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = await res.json();
        if (!cancelled) {
          setSig(data);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load GRC issuance signal');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    const id = window.setInterval(load, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    });
  };

  const modeLabel = (mode: string | undefined) => {
    if (!mode) return 'Hold';
    switch (mode) {
      case 'issue': return 'Issue (Expansion)';
      case 'constrict': return 'Constrict (Burn)';
      default: return 'Hold';
    }
  };

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">GRC Issuance Policy</h1>
        <div className="flex items-center gap-2 text-xs opacity-70">
          <span>Epoch</span>
          <span className="rc-pill rc-pill--slate">
            {sig?.Epoch ?? sig?.epoch ?? '–'}
          </span>
        </div>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading GRC issuance signal…</div>}
        {error && (
          <div className="rc-error">
            <div className="font-semibold mb-1">Error</div>
            <div className="text-xs opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
              <div className="rc-pill rc-pill--emerald flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Mode
                </span>
                <span className="text-sm font-semibold mt-1">
                  {modeLabel(sig?.Mode ?? sig?.mode)}
                </span>
              </div>
              <div className="rc-pill rc-pill--amber flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Recommended Δ GRC
                </span>
                <span className="text-sm font-semibold mt-1">
                  {fmt(sig?.RecommendedDelta ?? sig?.recommended_delta ?? 0, 4)} GRC
                </span>
              </div>
              <div className="rc-pill rc-pill--purple flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Target Supply
                </span>
                <span className="text-sm font-semibold mt-1">
                  {fmt(sig?.TargetGRCSupply ?? sig?.target_grc_supply ?? 0, 4)} GRC
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title">Treasury Snapshot</h2>
                </div>
                <div className="rc-card-body text-sm">
                  <div className="flex justify-between py-1">
                    <span className="opacity-70">Reserve Assets (USD)</span>
                    <span className="font-mono">{fmt(sig?.ReserveAssetsUSD ?? sig?.reserve_assets_usd ?? 0)}</span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="opacity-70">Equity (USD)</span>
                    <span className="font-mono">{fmt(sig?.EquityUSD ?? sig?.equity_usd ?? 0)}</span>
                  </div>
                  <div className="flex justify-between py-1">
                    <span className="opacity-70">Current GRC Supply</span>
                    <span className="font-mono">{fmt(sig?.CurrentGRCSupply ?? sig?.current_grc_supply ?? 0, 4)} GRC</span>
                  </div>
                </div>
              </div>
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title">Policy Commentary</h2>
                </div>
                <div className="rc-card-body text-xs opacity-80 space-y-2">
                  <p>
                    The hybrid policy combines corridor coverage, treasury equity, and
                    net USDR demand to produce a single issuance signal each epoch.
                  </p>
                  <p>
                    Positive signal &rarr; expansion (issuance). Negative signal &rarr; constriction (burn).
                    A near-zero signal results in a hold (no change to GRC supply).
                  </p>
                  <p>
                    This panel is read-only. Actual issuance and burns are applied when
                    the operator settles a mainnet epoch.
                  </p>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};



const StakingRewardsPanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [history, setHistory] = React.useState<any[]>([]);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/mainnet-history?limit=64');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setHistory(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load mainnet history');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    const id = window.setInterval(load, 8000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    });
  };

  const epochs = history.map(h => h.Epoch ?? h.epoch ?? 0);
  const stakeUSD = history.map(h => h.StakePortionUSD ?? h.stake_portion_usd ?? 0);
  const last = history.length > 0 ? history[history.length - 1] : null;
  const lastEpoch = last ? (last.Epoch ?? last.epoch ?? null) : null;
  const lastStakeUSD = last ? (last.StakePortionUSD ?? last.stake_portion_usd ?? null) : null;
  const lastStakeShare = last ? (last.StakeShare ?? last.stake_share ?? null) : null;

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Staking Rewards (RSX)</h1>
        <div className="flex items-center gap-2 text-xs opacity-70">
          <span>Epoch</span>
          <span className="rc-pill rc-pill--slate">{lastEpoch ?? '–'}</span>
        </div>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading staking rewards…</div>}
        {error && (
          <div className="rc-error mb-4">
            <div className="font-semibold mb-1">Error</div>
            <div className="text-xs opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="rc-card">
              <div className="rc-card-header">
                <h2 className="rc-card-title">Latest Epoch</h2>
              </div>
              <div className="rc-card-body text-sm space-y-1.5">
                <div className="flex justify-between">
                  <span className="opacity-70">Epoch</span>
                  <span className="font-mono">{lastEpoch ?? '–'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Staking Reward Pool</span>
                  <span className="font-mono">${fmt(lastStakeUSD ?? 0, 2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Share of Total Rewards</span>
                  <span className="font-mono">
                    {lastStakeShare !== null ? fmt((lastStakeShare as number) * 100, 2) + '%' : '–'}
                  </span>
                </div>
              </div>
            </div>
            <div className="rc-card md:col-span-2">
              <div className="rc-card-header flex items-center justify-between">
                <h2 className="rc-card-title">Staking Reward History</h2>
                <span className="text-[0.65rem] opacity-70">USD equivalent per epoch</span>
              </div>
              <div className="rc-card-body">
                {epochs.length === 0 ? (
                  <div className="text-xs opacity-70">No epochs settled yet.</div>
                ) : (
                  <div className="text-xs opacity-70">
                    Latest {epochs.length} epochs recorded. Charts and per-validator breakdowns
                    will appear here as the node-level staking model is wired in.
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const PoPRewardsPanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [history, setHistory] = React.useState<any[]>([]);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/mainnet-history?limit=64');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setHistory(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load mainnet history');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    const id = window.setInterval(load, 8000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    });
  };

  const epochs = history.map(h => h.Epoch ?? h.epoch ?? 0);
  const popUSD = history.map(h => h.PoPPortionUSD ?? h.pop_portion_usd ?? 0);
  const last = history.length > 0 ? history[history.length - 1] : null;
  const lastEpoch = last ? (last.Epoch ?? last.epoch ?? null) : null;
  const lastPoPUSD = last ? (last.PoPPortionUSD ?? last.pop_portion_usd ?? null) : null;
  const lastPoPShare = last ? (last.PoPShare ?? last.pop_share ?? null) : null;
  const lastStableUSD = last ? (last.PoPStablePortionUSD ?? last.pop_stable_portion_usd ?? null) : null;
  const lastVolUSD = last ? (last.PoPVolPortionUSD ?? last.pop_vol_portion_usd ?? null) : null;
  const lastStableShare = last ? (last.PoPStableShare ?? last.pop_stable_share ?? null) : null;
  const lastVolShare = last ? (last.PoPVolShare ?? last.pop_vol_share ?? null) : null;

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Operator Rewards (PoP)</h1>
        <div className="flex items-center gap-2 text-xs opacity-70">
          <span>Epoch</span>
          <span className="rc-pill rc-pill--slate">{lastEpoch ?? '–'}</span>
        </div>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading operator rewards…</div>}
        {error && (
          <div className="rc-error mb-4">
            <div className="font-semibold mb-1">Error</div>
            <div className="text-xs opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="rc-card">
              <div className="rc-card-header">
                <h2 className="rc-card-title">Latest Epoch</h2>
              </div>
              <div className="rc-card-body text-sm space-y-1.5">
                <div className="flex justify-between">
                  <span className="opacity-70">Epoch</span>
                  <span className="font-mono">{lastEpoch ?? '–'}</span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Operator Reward Pool</span>
                  <span className="font-mono">${fmt(lastPoPUSD ?? 0, 2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Share of Total Rewards</span>
                  <span className="font-mono">
                    {lastPoPShare !== null ? fmt((lastPoPShare as number) * 100, 2) + '%' : '–'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Stable (USDR) Portion</span>
                  <span className="font-mono">
                    {lastStableUSD !== null ? `$${fmt(lastStableUSD, 2)}` : '–'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="opacity-70">Volatile (GRC) Portion</span>
                  <span className="font-mono">
                    {lastVolUSD !== null ? `$${fmt(lastVolUSD, 2)}` : '–'}
                  </span>
                </div>
              </div>
            </div>
            <div className="rc-card md:col-span-2">
              <div className="rc-card-header flex items-center justify-between">
                <h2 className="rc-card-title">Operator Reward History</h2>
                <span className="text-[0.65rem] opacity-70">USD equivalent per epoch</span>
              </div>
              <div className="rc-card-body">
                {epochs.length === 0 ? (
                  <div className="text-xs opacity-70">No epochs settled yet.</div>
                ) : (
                  <div className="text-xs opacity-70">
                    Latest {epochs.length} epochs recorded. Per-node PoP scores and payouts
                    will be displayed here once the node-level work accounting is wired.
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};


const EpochControlPanel: React.FC = () => {
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [state, setState] = React.useState<any | null>(null);
  const [prevState, setPrevState] = React.useState<any | null>(null);
  const [lastDelta, setLastDelta] = React.useState<any | null>(null);
  const [settling, setSettling] = React.useState(false);
  const [autoMode, setAutoMode] = React.useState(false);
  const [autoIntervalSec, setAutoIntervalSec] = React.useState(10);
  const [verboseOpen, setVerboseOpen] = React.useState(false);

  React.useEffect(() => {
    let cancelled = false;

    async function loadState() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/mainnet-state');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setState(data);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load mainnet state');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    loadState();
    const id = window.setInterval(loadState, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  React.useEffect(() => {
    if (!autoMode) return;
    let cancelled = false;

    const tick = async () => {
      if (cancelled) return;
      await handleSettle(true);
    };

    const id = window.setInterval(tick, Math.max(5, autoIntervalSec) * 1000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [autoMode, autoIntervalSec]);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals
    });
  };

  const handleSettle = async (fromAuto: boolean = false) => {
    if (settling) return;
    setSettling(true);
    setError(null);
    try {
      const before = state;
      const res = await fetch('/econ/settle-mainnet-epoch', {
        method: 'POST'
      });
      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
      }
      const after = await res.json();
      setPrevState(before);
      setState(after);

      if (before && after) {
        const dEpoch =
          (after.Epoch ?? after.epoch ?? 0) -
          (before.Epoch ?? before.epoch ?? 0);
        const beforeSupply = before.Supply ?? before.supply ?? {};
        const afterSupply = after.Supply ?? after.supply ?? {};
        const dUSDR =
          (afterSupply.USDR ?? afterSupply.usdr ?? 0) -
          (beforeSupply.USDR ?? beforeSupply.usdr ?? 0);
        const dGRC =
          (afterSupply.GRC ?? afterSupply.grc ?? 0) -
          (beforeSupply.GRC ?? beforeSupply.grc ?? 0);

        const beforeEq = (before.Equity ?? before.equity ?? {}).EquityUSD ?? (before.Equity ?? before.equity ?? {}).equity_usd ?? 0;
        const afterEq = (after.Equity ?? after.equity ?? {}).EquityUSD ?? (after.Equity ?? after.equity ?? {}).equity_usd ?? 0;
        const dEq = afterEq - beforeEq;

        setLastDelta({
          dEpoch,
          dUSDR,
          dGRC,
          dEquityUSD: dEq,
          fromAuto
        });
      } else {
        setLastDelta(null);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to settle epoch');
    } finally {
      setSettling(false);
    }
  };

  const epoch = state?.Epoch ?? state?.epoch ?? null;
  const equity = (state?.Equity ?? state?.equity ?? {}).EquityUSD ?? (state?.Equity ?? state?.equity ?? {}).equity_usd ?? null;
  const supply = state?.Supply ?? state?.supply ?? {};
  const usdr = supply.USDR ?? supply.usdr ?? null;
  const grc = supply.GRC ?? supply.grc ?? null;

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Epoch Control</h1>
        <div className="flex items-center gap-2 text-xs opacity-70">
          <span>Epoch</span>
          <span className="rc-pill rc-pill--slate">
            {epoch ?? '–'}
          </span>
          <span className="ml-2 rc-badge rc-badge--outline">
            DevNet / Simulation auto mode only &mdash; mainnet uses on-chain clocks.
          </span>
        </div>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading mainnet state…</div>}
        {error && (
          <div className="rc-error mb-4">
            <div className="font-semibold mb-1">Error</div>
            <div className="text-xs opacity-80">{error}</div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="rc-card">
            <div className="rc-card-header">
              <h2 className="rc-card-title">Supply Snapshot</h2>
            </div>
            <div className="rc-card-body text-sm">
              <div className="flex justify-between py-1">
                <span className="opacity-70">USDR Supply</span>
                <span className="font-mono">{fmt(usdr ?? 0, 2)}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="opacity-70">GRC Supply</span>
                <span className="font-mono">{fmt(grc ?? 0, 4)}</span>
              </div>
              <div className="flex justify-between py-1">
                <span className="opacity-70">Equity (USD)</span>
                <span className="font-mono">{fmt(equity ?? 0, 2)}</span>
              </div>
            </div>
          </div>
          <div className="rc-card">
            <div className="rc-card-header">
              <h2 className="rc-card-title">Epoch Actions</h2>
            </div>
            <div className="rc-card-body text-sm space-y-3">
              <button
                className="rc-btn rc-btn--primary w-full"
                onClick={() => handleSettle(false)}
                disabled={settling}
              >
                {settling ? 'Settling…' : 'Settle Epoch Now'}
              </button>
              <div className="flex items-center gap-2 text-xs opacity-80">
                <span>Auto-settle every</span>
                <input
                  type="number"
                  min={5}
                  max={300}
                  value={autoIntervalSec}
                  onChange={e => setAutoIntervalSec(Number(e.target.value) || 10)}
                  className="rc-input rc-input--xs w-16"
                />
                <span>seconds</span>
                <label className="flex items-center gap-1 ml-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={autoMode}
                    onChange={e => setAutoMode(e.target.checked)}
                  />
                  <span>Enable auto (DevNet only)</span>
                </label>
              </div>
            </div>
          </div>
          <div className="rc-card">
            <div className="rc-card-header">
              <h2 className="rc-card-title">Last Epoch Summary</h2>
            </div>
            <div className="rc-card-body text-xs opacity-80 space-y-1">
              {lastDelta ? (
                <>
                  <div className="flex justify-between">
                    <span>Δ Epoch</span>
                    <span className="font-mono">{lastDelta.dEpoch}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Δ USDR Supply</span>
                    <span className="font-mono">{fmt(lastDelta.dUSDR, 2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Δ GRC Supply</span>
                    <span className="font-mono">{fmt(lastDelta.dGRC, 4)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Δ Equity (USD)</span>
                    <span className="font-mono">{fmt(lastDelta.dEquityUSD, 2)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Origin</span>
                    <span className="font-mono">{lastDelta.fromAuto ? 'Auto' : 'Manual'}</span>
                  </div>
                </>
              ) : (
                <div>No epoch has been settled from this console yet.</div>
              )}
            </div>
          </div>
        </div>

        <div className="rc-card">
          <div className="rc-card-header flex items-center justify-between">
            <h2 className="rc-card-title text-sm">Verbose Details</h2>
            <button
              className="rc-link text-xs"
              onClick={() => setVerboseOpen(v => !v)}
            >
              {verboseOpen ? 'Hide' : 'Show'} before / after snapshots
            </button>
          </div>
          {verboseOpen && (
            <div className="rc-card-body grid grid-cols-1 md:grid-cols-2 gap-4 text-[0.65rem] font-mono whitespace-pre overflow-x-auto max-h-80">
              <div>
                <div className="font-semibold mb-1">Before (prevState)</div>
                <pre className="rc-code-block">
{JSON.stringify(prevState, null, 2)}
                </pre>
              </div>
              <div>
                <div className="font-semibold mb-1">After (state)</div>
                <pre className="rc-code-block">
{JSON.stringify(state, null, 2)}
                </pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};


const HistoryChart: React.FC<{ label: string; series: { x: number; y: number }[]; yLabel: string }> = ({ label, series, yLabel }) => {
  if (!series.length) {
    return (
      <div className="rc-card">
        <div className="rc-card-header">
          <h2 className="rc-card-title text-sm">{label}</h2>
        </div>
        <div className="rc-card-body text-xs opacity-80">
          No epoch history available yet. Settle a few epochs to populate this chart.
        </div>
      </div>
    );
  }

  const xs = series.map(p => p.x);
  const ys = series.map(p => p.y);
  const minX = Math.min(...xs);
  const maxX = Math.max(...xs);
  const minY = Math.min(...ys);
  const maxY = Math.max(...ys);

  const width = 400;
  const height = 120;
  const padLeft = 32;
  const padRight = 8;
  const padTop = 10;
  const padBottom = 18;

  const spanX = maxX - minX || 1;
  const spanY = maxY - minY || 1;

  const toSvgX = (x: number) =>
    padLeft + ((x - minX) / spanX) * (width - padLeft - padRight);
  const toSvgY = (y: number) =>
    height - padBottom - ((y - minY) / spanY) * (height - padTop - padBottom);

  const pathD = series
    .map((p, idx) => {
      const cx = toSvgX(p.x);
      const cy = toSvgY(p.y);
      return `${idx === 0 ? 'M' : 'L'} ${cx} ${cy}`;
    })
    .join(' ');

  return (
    <div className="rc-card">
      <div className="rc-card-header flex items-center justify-between">
        <h2 className="rc-card-title text-sm">{label}</h2>
        <span className="text-[0.65rem] opacity-70">{yLabel}</span>
      </div>
      <div className="rc-card-body">
        <svg viewBox={`0 0 ${width} ${height}`} className="w-full h-32">
          <path d={pathD} fill="none" stroke="currentColor" strokeWidth={1.3} />
        </svg>
        <div className="flex justify-between text-[0.6rem] opacity-60 mt-1">
          <span>Epoch {minX}</span>
          <span>Epoch {maxX}</span>
        </div>
      </div>
    </div>
  );
};

const SupplyHistoryPanel: React.FC = () => {
  const [history, setHistory] = React.useState<any[] | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/history?limit=200');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setHistory(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        if (!cancelled) setError(err?.message || 'Failed to load epoch history');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    const id = window.setInterval(load, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const toSeries = (key: 'USDRSupply' | 'GRCSupply') => {
    if (!history) return [];
    return history.map((h: any) => ({
      x: h.epoch ?? h.Epoch,
      y: key === 'USDRSupply' ? (h.usdr_supply ?? h.USDRSupply ?? 0) : (h.grc_supply ?? h.GRCSupply ?? 0),
    }));
  };

  const usdrSeries = toSeries('USDRSupply');
  const grcSeries = toSeries('GRCSupply');

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Supply Over Time</h1>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading epoch history…</div>}
        {error && (
          <div className="rc-error mb-4 text-xs">
            <div className="font-semibold mb-1">Error</div>
            <div className="opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && history && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <HistoryChart label="USDR Supply" yLabel="USDR" series={usdrSeries} />
            <HistoryChart label="GRC Supply" yLabel="GRC" series={grcSeries} />
          </div>
        )}
      </div>
    </div>
  );
};

const EquityHistoryPanel: React.FC = () => {
  const [history, setHistory] = React.useState<any[] | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/history?limit=200');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setHistory(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        if (!cancelled) setError(err?.message || 'Failed to load epoch history');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    const id = window.setInterval(load, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const seriesEquity = (history || []).map((h: any) => ({
    x: h.epoch ?? h.Epoch,
    y: h.equity_usd ?? h.EquityUSD ?? 0,
  }));

  const seriesCoverage = (history || []).map((h: any) => {
    const reserves = h.reserve_assets_usd ?? h.ReserveAssetsUSD ?? 0;
    const usdr = h.usdr_supply ?? h.USDRSupply ?? 0;
    const cov = usdr > 0 ? reserves / usdr : 0;
    return {
      x: h.epoch ?? h.Epoch,
      y: cov,
    };
  });

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Equity & Coverage Over Time</h1>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading epoch history…</div>}
        {error && (
          <div className="rc-error mb-4 text-xs">
            <div className="font-semibold mb-1">Error</div>
            <div className="opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && history && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <HistoryChart label="Equity (USD)" yLabel="USD" series={seriesEquity} />
            <HistoryChart label="USDR Coverage (x)" yLabel="Coverage multiple" series={seriesCoverage} />
          </div>
        )}
      </div>
    </div>
  );
};

const GRCPolicyHistoryPanel: React.FC = () => {
  const [history, setHistory] = React.useState<any[] | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/history?limit=200');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setHistory(Array.isArray(data) ? data : []);
        }
      } catch (err: any) {
        if (!cancelled) setError(err?.message || 'Failed to load epoch history');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    const id = window.setInterval(load, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const seriesDelta = (history || []).map((h: any) => ({
    x: h.epoch ?? h.Epoch,
    y: h.delta_grc ?? h.DeltaGRC ?? 0,
  }));

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">GRC Policy Output</h1>
      </div>
      <div className="rc-ws-panel-body">
        {loading && <div className="rc-loading">Loading epoch history…</div>}
        {error && (
          <div className="rc-error mb-4 text-xs">
            <div className="font-semibold mb-1">Error</div>
            <div className="opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && history && (
          <div className="grid grid-cols-1 gap-4">
            <HistoryChart label="Δ GRC Supply per Epoch" yLabel="Δ GRC" series={seriesDelta} />
            <div className="rc-card text-xs opacity-80">
              <div className="rc-card-header">
                <h2 className="rc-card-title text-sm">Mode Legend</h2>
              </div>
              <div className="rc-card-body space-y-1">
                <p><strong>issue</strong> &mdash; expansionary, positive ΔGRC suggested.</p>
                <p><strong>hold</strong> &mdash; neutral, no supply change suggested.</p>
                <p><strong>constrict</strong> &mdash; contractionary, negative ΔGRC suggested.</p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};


const ReserveCoveragePanel: React.FC = () => {
  const [snap, setSnap] = React.useState<any | null>(null);
  const [error, setError] = React.useState<string | null>(null);
  const [loading, setLoading] = React.useState(true);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/coverage');
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data = await res.json();
        if (!cancelled) {
          setSnap(data);
        }
      } catch (err: any) {
        if (!cancelled) setError(err?.message || 'Failed to load coverage snapshot');
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    const id = window.setInterval(load, 10000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const assets = snap?.assets || [];
  const cov = snap?.coverage_multiple ?? snap?.Coverage ?? 0;
  const rawRes = snap?.raw_reserves_usd ?? snap?.RawReservesUSD ?? 0;
  const effRes = snap?.effective_reserves_usd ?? snap?.EffReservesUSD ?? 0;
  const usdr = snap?.usdr_supply ?? snap?.USDRSupply ?? 0;
  const stableShare = snap?.stable_share ?? snap?.StableShare ?? 0;

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Reserve Coverage</h1>
        <p className="rc-ws-panel-subtitle text-xs opacity-70">
          Crypto-only basket (USDC, ETH, WBTC, STETH) with haircuts applied.
        </p>
      </div>
      <div className="rc-ws-panel-body space-y-4">
        {loading && <div className="rc-loading">Loading coverage snapshot…</div>}
        {error && (
          <div className="rc-error text-xs">
            <div className="font-semibold mb-1">Error</div>
            <div className="opacity-80">{error}</div>
          </div>
        )}
        {!loading && !error && snap && (
          <>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-xs">
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title text-[0.7rem]">USDR Supply</h2>
                </div>
                <div className="rc-card-body text-sm font-mono">
                  {usdr.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                </div>
              </div>
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title text-[0.7rem]">Raw Reserves (USD)</h2>
                </div>
                <div className="rc-card-body text-sm font-mono">
                  {rawRes.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                </div>
              </div>
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title text-[0.7rem]">Effective Reserves (USD)</h2>
                </div>
                <div className="rc-card-body text-sm font-mono">
                  {effRes.toLocaleString(undefined, { maximumFractionDigits: 2 })}
                </div>
              </div>
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title text-[0.7rem]">Coverage Multiple</h2>
                </div>
                <div className="rc-card-body text-sm font-mono">
                  {cov.toFixed(3)}x
                </div>
              </div>
              <div className="rc-card">
                <div className="rc-card-header">
                  <h2 className="rc-card-title text-[0.7rem]">Stable Share</h2>
                </div>
                <div className="rc-card-body text-sm font-mono">
                  {(stableShare * 100).toFixed(1)}%
                </div>
              </div>
            </div>

            <div className="rc-card">
              <div className="rc-card-header">
                <h2 className="rc-card-title text-sm">Per-Asset Breakdown (effective USD)</h2>
              </div>
              <div className="rc-card-body overflow-x-auto">
                <table className="min-w-full text-[0.7rem]">
                  <thead className="opacity-70">
                    <tr>
                      <th className="text-left pr-4">Asset</th>
                      <th className="text-right pr-4">Role</th>
                      <th className="text-right pr-4">USD (eff.)</th>
                      <th className="text-right">Share</th>
                    </tr>
                  </thead>
                  <tbody>
                    {assets.map((a: any, idx: number) => (
                      <tr key={idx} className="border-t border-white/5">
                        <td className="pr-4 py-1 font-mono">{a.kind || a.Kind}</td>
                        <td className="pr-4 py-1 text-right">{a.role || a.Role || '-'}</td>
                        <td className="pr-4 py-1 text-right font-mono">
                          {(a.usd ?? a.USD ?? 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}
                        </td>
                        <td className="py-1 text-right font-mono">
                          {(((a.share ?? a.Share) || 0) * 100).toFixed(2)}%
                        </td>
                      </tr>
                    ))}
                    {!assets.length && (
                      <tr>
                        <td colSpan={4} className="py-2 text-center opacity-60">
                          No reserve assets recorded yet.
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
const MintQueuePanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [snap, setSnap] = React.useState<any | null>(null);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/mints');
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = await res.json();
        if (!cancelled) {
          setSnap(data);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load mint queue');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    const id = window.setInterval(load, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    });
  };

  const rows = snap?.pending ?? snap?.Pending ?? [];

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Mint Queue</h1>
        <span className="rc-ws-panel-subtitle">
          Epoch-based USDR issuance (DevNet simulation)
        </span>
      </div>

      <div className="rc-ws-panel-body space-y-4">
        {loading && <div className="text-xs text-slate-400">Loading mint queue…</div>}
        {error && (
          <div className="text-xs text-rose-400">
            Error loading mint data: {error}
          </div>
        )}

        {!loading && !error && snap && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="rc-pill rc-pill--teal flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Current Epoch
                </span>
                <span className="text-sm font-semibold mt-1">
                  {snap.current_epoch ?? snap.CurrentEpoch ?? '–'}
                </span>
              </div>
              <div className="rc-pill rc-pill--amber flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Total USDR to Mint
                </span>
                <span className="text-sm font-semibold mt-1">
                  {fmt(snap.total_usdr ?? snap.TotalUSDR ?? 0)} USDR
                </span>
              </div>
              <div className="rc-pill rc-pill--purple flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Pending Requests
                </span>
                <span className="text-sm font-semibold mt-1">
                  {rows.length}
                </span>
              </div>
            </div>

            <div className="mt-4 rc-ws-card">
              <div className="rc-ws-card-header">
                <div className="rc-ws-card-title">Pending Mints</div>
                <div className="rc-ws-card-subtitle">Crypto-backed + test mints</div>
              </div>
              <div className="rc-ws-card-body">
                {rows.length === 0 ? (
                  <div className="text-xs text-slate-400">
                    No pending mints in the queue.
                  </div>
                ) : (
                  <div className="rc-ws-table-wrapper">
                    <table className="rc-ws-table">
                      <thead>
                        <tr>
                          <th>Account</th>
                          <th>Type</th>
                          <th>Asset</th>
                          <th>Asset Amt</th>
                          <th>USDR</th>
                          <th>Epoch</th>
                          <th>Status</th>
                          <th>Created</th>
                        </tr>
                      </thead>
                      <tbody>
                        {rows.map((row: any) => (
                          <tr key={row.id ?? row.ID}>
                            <td className="font-mono text-[0.7rem]">
                              {row.account_ref ?? row.AccountRef ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.mint_type ?? row.MintType ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.asset ?? row.Asset ?? '–'}
                            </td>
                            <td className="font-mono text-[0.75rem]">
                              {fmt(row.amount_asset ?? row.AmountAsset ?? 0, 4)}
                            </td>
                            <td className="font-mono text-[0.75rem]">
                              {fmt(row.amount_usdr ?? row.AmountUSDR ?? 0)}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.epoch ?? row.Epoch ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.status ?? row.Status ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.created_at ?? row.CreatedAt ?? '–'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

const RedemptionQueuePanel: React.FC = () => {
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [snap, setSnap] = React.useState<any | null>(null);

  React.useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/econ/redemptions');
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        const data = await res.json();
        if (!cancelled) {
          setSnap(data);
        }
      } catch (err: any) {
        if (!cancelled) {
          setError(err?.message || 'Failed to load redemption queue');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    const id = window.setInterval(load, 5000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    });
  };

  const handleAdvanceEpoch = async () => {
    try {
      await fetch('/econ/advance-epoch', { method: 'POST' });
    } catch (e) {
      console.error('Failed to advance epoch', e);
    }
  };

  };

  const rows = snap?.pending ?? snap?.Pending ?? [];

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <div>
          <h1 className="rc-ws-panel-title">Redemption Queue</h1>
          <span className="rc-ws-panel-subtitle">
            Epoch-based USDR redemptions (DevNet simulation)
          </span>
        </div>
        <button
          type="button"
          onClick={handleAdvanceEpoch}
          className="rc-btn rc-btn-sm rc-btn-ghost"
          title="Force a DevNet epoch advance (operator only)"
        >
          Force Epoch Advance
        </button>
      </div>

      <div className="rc-ws-panel-body space-y-4">
        {loading && <div className="text-xs text-slate-400">Loading redemption queue…</div>}
        {error && (
          <div className="text-xs text-rose-400">
            Error loading redemption data: {error}
          </div>
        )}

        {!loading && !error && snap && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="rc-pill rc-pill--teal flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Current Epoch
                </span>
                <span className="text-sm font-semibold mt-1">
                  {snap.current_epoch ?? snap.CurrentEpoch ?? '–'}
                </span>
              </div>
              <div className="rc-pill rc-pill--amber flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Total Pending USDR
                </span>
                <span className="text-sm font-semibold mt-1">
                  {fmt(snap.total_pending_usdr ?? snap.TotalPendingUSDR ?? 0)} USDR
                </span>
              </div>
              <div className="rc-pill rc-pill--purple flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">
                  Last Epoch Settled
                </span>
                <span className="text-sm font-semibold mt-1">
                  {snap.last_epoch_settled_at ?? snap.LastEpochSettledAt ?? '–'}
                </span>
              </div>
            </div>

            <div className="mt-4 rc-ws-card">
              <div className="rc-ws-card-header">
                <div className="rc-ws-card-title">Pending Requests</div>
                <div className="rc-ws-card-subtitle">Per-account DevNet redemptions</div>
              </div>
              <div className="rc-ws-card-body">
                {rows.length === 0 ? (
                  <div className="text-xs text-slate-400">
                    No pending redemptions in the queue.
                  </div>
                ) : (
                  <div className="rc-ws-table-wrapper">
                    <table className="rc-ws-table">
                      <thead>
                        <tr>
                          <th>Account</th>
                          <th>Tier</th>
                          <th>Amount (USDR)</th>
                          <th>Epoch</th>
                          <th>Status</th>
                          <th>Created</th>
                        </tr>
                      </thead>
                      <tbody>
                        {rows.map((row: any) => (
                          <tr key={row.id ?? row.ID}>
                            <td className="font-mono text-[0.7rem]">
                              {row.account_ref ?? row.AccountRef ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.tier ?? row.Tier ?? '–'}
                            </td>
                            <td className="font-mono text-[0.75rem]">
                              {fmt(row.amount_usdr ?? row.AmountUSDR ?? 0)}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.epoch ?? row.Epoch ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.status ?? row.Status ?? '–'}
                            </td>
                            <td className="text-[0.7rem]">
                              {row.created_at ?? row.CreatedAt ?? '–'}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};


  }, []);

  const reserveCoverage = snap?.reserve_coverage ?? snap?.ReserveCoverage ?? null;
  const grcCoverage = snap?.grc_coverage ?? snap?.GRCCoverage ?? null;
  const reserveAssets = snap?.reserve_assets_usd ?? snap?.ReserveAssetsUSD ?? null;
  const usdrSupply = snap?.usdr_supply ?? snap?.USDRSupply ?? null;
  const grcSupply = snap?.grc_supply ?? snap?.GRCSupply ?? null;
  const pendingRedemptions = snap?.pending_usdr_redemptions ?? snap?.PendingUSDRRedemptions ?? null;

  const fmt = (v: any, decimals: number = 2) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return Number(v).toLocaleString(undefined, {
      minimumFractionDigits: decimals,
      maximumFractionDigits: decimals,
    });
  };

  const pct = (v: any) => {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '–';
    return (Number(v) * 100).toFixed(2) + '%';
  };

  return (
    <div className="rc-ws-panel">
      <div className="rc-ws-panel-header">
        <h1 className="rc-ws-panel-title">Treasury &amp; Reserve Coverage</h1>
        <span className="rc-ws-panel-subtitle">Live snapshot from /econ/treasury</span>
      </div>
      <div className="rc-ws-panel-body space-y-4">
        {loading && <div className="text-xs text-slate-400">Loading treasury snapshot…</div>}
        {error && (
          <div className="text-xs text-rose-400">
            Error loading snapshot: {error}
          </div>
        )}

        {!loading && !error && snap && (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="rc-pill rc-pill--amber flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">Reserve Assets (MTM)</span>
                <span className="text-sm font-semibold mt-1">${fmt(reserveAssets, 2)}</span>
              </div>
              <div className="rc-pill rc-pill--teal flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">USDR Supply</span>
                <span className="text-sm font-semibold mt-1">{fmt(usdrSupply, 2)} USDR</span>
              </div>
              <div className="rc-pill rc-pill--purple flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">GRC Supply</span>
                <span className="text-sm font-semibold mt-1">{fmt(grcSupply, 2)} GRC</span>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-3">
              <div className="rc-pill rc-pill--green flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">USDR Reserve Coverage</span>
                <span className="text-sm font-semibold mt-1">{pct(reserveCoverage)}</span>
                <span className="text-[0.7rem] opacity-70 mt-0.5">
                  Target ≥ 100% — stablecoin fully reserved against crypto basket.
                </span>
              </div>
              <div className="rc-pill rc-pill--gold flex flex-col">
                <span className="text-[0.65rem] uppercase tracking-wide opacity-80">GRC Coverage</span>
                <span className="text-sm font-semibold mt-1">{pct(grcCoverage)}</span>
                <span className="text-[0.7rem] opacity-70 mt-0.5">
                  Hybrid regime — GRC is base money issued on top of the reserve basket.
                </span>
              </div>
            </div>

            <div className="mt-4 rc-ws-card">
              <div className="rc-ws-card-header">
                <div className="rc-ws-card-title">Redemption Queue</div>
                <div className="rc-ws-card-subtitle">Epoch-based settlement (DevNet simulation)</div>
              </div>
              <div className="rc-ws-card-body text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-slate-400">Pending USDR Redemptions</span>
                  <span className="font-mono">{fmt(pendingRedemptions, 2)} USDR</span>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

const AppSidebar: React.FC = () => {{
  const location = useLocation();
  const [financeOpen, setFinanceOpen] = React.useState<FinanceOpenState>(() => loadInitialFinanceState());

  const toggleFinanceGroup = (groupId: string) => {{
    setFinanceOpen(prev => {{
      const next = {{ ...prev, [groupId]: !prev[groupId] }};
      try {{
        window.localStorage.setItem(FINANCE_STATE_KEY, JSON.stringify(next));
      }} catch {{}}
      return next;
    }});
  }};

  return (
    <aside className="rc-ws-sidebar">
      <div className="px-4 py-3 border-b border-slate-800">
        <div className="text-xs font-semibold tracking-wide text-slate-500 uppercase">ReserveChain</div>
        <div className="text-sm text-slate-100">Workstation</div>
      </div>
      <div className="flex-1 overflow-y-auto text-sm">
        {/* Finance */}
        <div className="mt-4">
          <div className="px-4 text-[0.7rem] font-semibold tracking-wide text-slate-500 uppercase">
            Finance
          </div>
          <div className="mt-1">
            {{FINANCE_GROUPS.map(group => (
              <div key={group.id} className="mb-1">
                <button
                  type="button"
                  onClick={() => toggleFinanceGroup(group.id)}
                  className="w-full flex items-center justify-between px-4 py-1.5 text-[0.7rem] uppercase tracking-wide text-slate-500 hover:text-slate-300"
                >
                  <span>{{group.label}}</span>
                  <span className="text-xs">
                    {{financeOpen[group.id] ? '–' : '+'}}
                  </span>
                </button>
                {{financeOpen[group.id] && (
                  <div className="mt-0.5">
                    {{group.panels.map(panel => {{
                      const path = panelIdToPath(panel.id);
                      const active = location.pathname === path;
                      return (
                        <Link
                          key={panel.id}
                          to={path}
                          className={`block pl-8 pr-4 py-1.5 text-xs rounded-sm ${
                            active ? 'bg-slate-800 text-slate-50' : 'text-slate-300 hover:bg-slate-900'
                          }`}
                        >
                          {{panel.label}}
                        </Link>
                      );
                    }})}}
                  </div>
                )}}
              </div>
            ))}}
          </div>
        </div>
        {/* Other top-level sections */}
        {{OTHER_SECTIONS.map(section => (
          <div key={section.id} className="mt-4">
            <div className="px-4 text-[0.7rem] font-semibold tracking-wide text-slate-500 uppercase">
              {{section.label}}
            </div>
            <div className="mt-1">
              {{section.panels.map(panel => {{
                const path = panelIdToPath(panel.id);
                const active = location.pathname === path;
                return (
                  <Link
                    key={panel.id}
                    to={path}
                    className={`block px-4 py-1.5 text-xs rounded-sm ${
                      active ? 'bg-slate-800 text-slate-50' : 'text-slate-300 hover:bg-slate-900'
                    }`}
                  >
                    {{panel.label}}
                  </Link>
                );
              }})}}
            </div>
          </div>
        ))}}
      </div>
      <div className="px-4 py-3 border-t border-slate-800 text-[0.7rem] text-slate-500">
        v2.0.0-pre2 · SPA shell
      </div>
    </aside>
  );
}};

export const App: React.FC = () => {{
  return (
    <div className="rc-ws-shell rc-ws-shell--dim">
      <AppSidebar />
      <div className="rc-ws-main">
        <header className="rc-ws-topbar">
          <div>Workstation SPA</div>
          <div className="flex items-center gap-4 text-[0.7rem] text-slate-400">
            <span>Devnet</span>
            <span>Theme: Extra Dim</span>
          </div>
        </header>
        <main className="flex-1 overflow-y-auto">
          <Routes>
            <Route path="/" element={<Navigate to="/vaults/vault-dashboard" replace />} />
            {{FINANCE_GROUPS.flatMap(group =>
              group.panels.map(panel => {{
                const path = panelIdToPath(panel.id);
                return (
                  <Route
                    key={panel.id}
                    path={path}
                    element={
                        panel.id === 'treasury-overview'
                          ? <TreasuryOverviewPanel />
                          : panel.id === 'treasury-mints'
                            ? <MintQueuePanel />
                            : panel.id === 'treasury-redemptions'
                              ? <RedemptionQueuePanel />
                              : panel.id === 'treasury-grc-policy'
                                ? <GRCPolicyPanel />
                                : <PanelPlaceholder id={panel.id} label={panel.label} />
                      }
                  />
                );
              }})
            )}}
            {{OTHER_SECTIONS.flatMap(section =>
              section.panels.map(panel => {{
                const path = panelIdToPath(panel.id);
                return (
                  <Route
                    key={panel.id}
                    path={path}
                    element={
                        panel.id === 'treasury-overview'
                          ? <TreasuryOverviewPanel />
                          : panel.id === 'network-coverage'
                            ? <ReserveCoveragePanel />
                            : panel.id === 'treasury-mints'
                            ? <MintQueuePanel />
                            : panel.id === 'treasury-redemptions'
                              ? <RedemptionQueuePanel />
                              : panel.id === 'treasury-grc-policy'
                                ? <GRCPolicyPanel />
                                : panel.id === 'network-epochs'
                                  ? <EpochControlPanel />
                                  : panel.id === 'network-history-supply'
                                    ? <SupplyHistoryPanel />
                                    : panel.id === 'network-history-equity'
                                      ? <EquityHistoryPanel />
                                      : panel.id === 'network-history-grc'
                                        ? <GRCPolicyHistoryPanel />
                                        : panel.id === 'network-staking'
                                          ? <StakingRewardsPanel />
                                          : panel.id === 'network-pop'
                                            ? <PoPRewardsPanel />
                                            : <PanelPlaceholder id={panel.id} label={panel.label} />
                      }
                  />
                );
              }})
            )}}
            <Route path="*" element={<Navigate to="/vaults/vault-dashboard" replace />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}};
