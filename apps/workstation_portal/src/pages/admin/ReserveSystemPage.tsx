import React, { useEffect, useMemo, useState } from "react";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
} from "chart.js";
import { Line } from "react-chartjs-2";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend);

type EconTick = {
  epoch: number;
  coverage: number;
  reserves: number;
  liabs: number;
  equity: number;
};

type SimResult = {
  baseline?: {
    epochs: number[];
    coverage: number[];
    equity: number[];
    reserves?: number[];
    liabs?: number[];
    reward_rsx?: number[];
    reward_pop?: number[];
  };
  merged?: Array<{
    mode: "max_impact" | "median" | "weighted";
    epochs: number[];
    coverage: number[];
    equity: number[];
  }>;
};

const pillStyle = (active: boolean) => ({
  fontSize: "0.75rem",
  padding: "4px 10px",
  borderRadius: "999px",
  border: active ? "1px solid rgba(77,163,255,0.8)" : "1px solid rgba(255,255,255,0.12)",
  background: active ? "var(--accent-soft)" : "transparent",
  color: active ? "var(--accent)" : "var(--text-muted)",
  cursor: "pointer",
});

const ReserveSystemPage: React.FC = () => {
  const [latest, setLatest] = useState<EconTick | null>(null);

  const [simOpen, setSimOpen] = useState(true);
  const [simMode, setSimMode] = useState<"current_only" | "current_plus_proposals" | "proposals_only">(
    "current_plus_proposals"
  );
  const [mergePref, setMergePref] = useState<"max_impact" | "median" | "weighted">("weighted");

  const [alphaOverride, setAlphaOverride] = useState<number>(0.55);
  const [treasurySmoothing, setTreasurySmoothing] = useState<number>(0.25);
  const [issuanceHalfLife, setIssuanceHalfLife] = useState<number>(8000);
  const [corridorTarget, setCorridorTarget] = useState<number>(1.05);
  const [corridorCeiling, setCorridorCeiling] = useState<number>(1.25);

  const [simResult, setSimResult] = useState<SimResult | null>(null);
  const [simLoading, setSimLoading] = useState(false);

  useEffect(() => {
    const loc = window.location;
    const wsProto = loc.protocol === "https:" ? "wss" : "ws";
    const ws = new WebSocket(`${wsProto}://${loc.host}/econ/live`);

    ws.onmessage = (msg) => {
      try {
        const data = JSON.parse(msg.data);
        const tick: EconTick = {
          epoch: data.epoch ?? 0,
          coverage: data.coverage ?? 0,
          reserves: data.reserves ?? 0,
          liabs: data.liabs ?? 0,
          equity: data.equity ?? 0,
        };
        setLatest(tick);
      } catch (e) {
        console.error("bad econ tick", e);
      }
    };

    ws.onerror = (e) => console.error("econ ws error", e);
    return () => ws.close();
  }, []);

  async function runSimulation() {
    setSimLoading(true);
    try {
      const body = {
        num_epochs: 300,
        mode: simMode,
        preferred_merge: mergePref,
        overrides: {
          alpha: alphaOverride,
          treasury_smoothing: treasurySmoothing,
          issuance_half_life: issuanceHalfLife,
          corridor_target: corridorTarget,
          corridor_ceiling: corridorCeiling,
        },
      };

      const res = await fetch("/api/sim", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const data = await res.json();
      setSimResult(data);
    } catch (e) {
      console.error("sim failed", e);
    } finally {
      setSimLoading(false);
    }
  }

  const cov = latest?.coverage ?? 0;
  const covDisplay = cov ? cov.toFixed(3) + "×" : "—";
  const equityDisplay =
    latest && !Number.isNaN(latest.equity)
      ? "$" + latest.equity.toLocaleString(undefined, { maximumFractionDigits: 0 })
      : "—";
  const reservesDisplay =
    latest && !Number.isNaN(latest.reserves)
      ? "$" + latest.reserves.toLocaleString(undefined, { maximumFractionDigits: 0 })
      : "—";
  const liabsDisplay =
    latest && !Number.isNaN(latest.liabs)
      ? "$" + latest.liabs.toLocaleString(undefined, { maximumFractionDigits: 0 })
      : "—";

  const chartData = useMemo(() => {
    const epochs = simResult?.baseline?.epochs ?? [];
    const labels = epochs.map((e) => String(e));

    const baselineCov = simResult?.baseline?.coverage ?? [];

    const merged = simResult?.merged ?? [];
    const maxImpact = merged.find((m) => m.mode === "max_impact")?.coverage ?? [];
    const median = merged.find((m) => m.mode === "median")?.coverage ?? [];
    const weighted = merged.find((m) => m.mode === "weighted")?.coverage ?? [];

    return {
      labels,
      datasets: [
        {
          label: "Baseline Coverage",
          data: baselineCov,
          tension: 0.22,
        },
        {
          label: "Merged (Max Impact)",
          data: maxImpact,
          tension: 0.22,
        },
        {
          label: "Merged (Median)",
          data: median,
          tension: 0.22,
        },
        {
          label: "Merged (Weighted)",
          data: weighted,
          tension: 0.22,
        },
      ],
    };
  }, [simResult]);

  const chartOptions = useMemo(
    () => ({
      responsive: true,
      plugins: {
        legend: {
          labels: { color: "#9a9fb0" as any },
        },
      },
      scales: {
        x: { ticks: { color: "#9a9fb0" as any }, grid: { color: "rgba(255,255,255,0.06)" as any } },
        y: { ticks: { color: "#9a9fb0" as any }, grid: { color: "rgba(255,255,255,0.06)" as any } },
      },
    }),
    []
  );

  return (
    <div>
      <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between" }}>
        <div>
          <h1 className="page-title">Reserve System</h1>
          <p className="page-subtitle">
            Live view of coverage, equity, reserves and liabilities for ReserveChain.
          </p>
        </div>

        <button
          style={{
            fontSize: "0.75rem",
            borderRadius: "999px",
            padding: "6px 10px",
            border: "1px solid rgba(255,255,255,0.12)",
            background: "transparent",
            color: "var(--text-muted)",
            cursor: "pointer",
            height: 34,
            marginTop: 10,
          }}
          onClick={() => setSimOpen((v) => !v)}
        >
          {simOpen ? "Hide Parameters" : "Show Parameters"}
        </button>
      </div>

      <div className="reserve-summary-grid">
        <div className="reserve-card">
          <div className="reserve-card__label">Coverage Ratio</div>
          <div className="reserve-card__value">{covDisplay}</div>
          <div className="reserve-card__sub">Reserves / Liabilities</div>
        </div>
        <div className="reserve-card">
          <div className="reserve-card__label">Equity</div>
          <div className="reserve-card__value">{equityDisplay}</div>
          <div className="reserve-card__sub">Reserves minus USDR liabilities</div>
        </div>
        <div className="reserve-card">
          <div className="reserve-card__label">Reserves</div>
          <div className="reserve-card__value">{reservesDisplay}</div>
          <div className="reserve-card__sub">Total effective reserves (USD)</div>
        </div>
        <div className="reserve-card">
          <div className="reserve-card__label">USDR Liabilities</div>
          <div className="reserve-card__value">{liabsDisplay}</div>
          <div className="reserve-card__sub">Circulating USDR obligations</div>
        </div>
      </div>

      {simOpen && (
        <div
          style={{
            marginTop: 16,
            padding: 14,
            borderRadius: 14,
            border: "1px solid rgba(255,255,255,0.06)",
            background: "rgba(15,18,26,0.65)",
          }}
        >
          <div style={{ display: "flex", gap: 10, flexWrap: "wrap", marginBottom: 10 }}>
            <div style={{ minWidth: 280 }}>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 6 }}>
                Simulation Mode
              </div>
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                {[
                  { id: "current_only", label: "Current Policy" },
                  { id: "current_plus_proposals", label: "Current + Proposals" },
                  { id: "proposals_only", label: "Proposals Only" },
                ].map((opt) => (
                  <button key={opt.id} style={pillStyle(simMode === (opt.id as any))} onClick={() => setSimMode(opt.id as any)}>
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>

            <div style={{ minWidth: 280 }}>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 6 }}>
                Merged View Preference
              </div>
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                {[
                  { id: "max_impact", label: "Max Impact" },
                  { id: "median", label: "Median" },
                  { id: "weighted", label: "Weighted" },
                ].map((opt) => (
                  <button key={opt.id} style={pillStyle(mergePref === (opt.id as any))} onClick={() => setMergePref(opt.id as any)}>
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>

            <div style={{ flex: 1, minWidth: 260, display: "flex", justifyContent: "flex-end", alignItems: "flex-end" }}>
              <button
                style={{
                  fontSize: "0.8rem",
                  padding: "8px 14px",
                  borderRadius: "999px",
                  border: "1px solid rgba(77,163,255,0.9)",
                  background: "var(--accent-soft)",
                  color: "var(--accent)",
                  cursor: "pointer",
                }}
                onClick={runSimulation}
                disabled={simLoading}
              >
                {simLoading ? "Running…" : "Run Simulation"}
              </button>
            </div>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))", gap: 10 }}>
            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 4 }}>α (Stake vs PoP)</div>
              <input type="range" min={0} max={1} step={0.01} value={alphaOverride} onChange={(e) => setAlphaOverride(parseFloat(e.target.value))} style={{ width: "100%" }} />
              <div style={{ fontSize: "0.75rem", color: "var(--text-muted)", marginTop: 4 }}>{alphaOverride.toFixed(2)}</div>
            </div>

            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 4 }}>Treasury Smoothing</div>
              <input type="range" min={0} max={1} step={0.01} value={treasurySmoothing} onChange={(e) => setTreasurySmoothing(parseFloat(e.target.value))} style={{ width: "100%" }} />
              <div style={{ fontSize: "0.75rem", color: "var(--text-muted)", marginTop: 4 }}>{treasurySmoothing.toFixed(2)}</div>
            </div>

            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 4 }}>Issuance Half-life</div>
              <input type="range" min={500} max={50000} step={100} value={issuanceHalfLife} onChange={(e) => setIssuanceHalfLife(parseFloat(e.target.value))} style={{ width: "100%" }} />
              <div style={{ fontSize: "0.75rem", color: "var(--text-muted)", marginTop: 4 }}>{Math.round(issuanceHalfLife)}</div>
            </div>

            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 4 }}>Corridor Target</div>
              <input type="range" min={0.7} max={1.5} step={0.01} value={corridorTarget} onChange={(e) => setCorridorTarget(parseFloat(e.target.value))} style={{ width: "100%" }} />
              <div style={{ fontSize: "0.75rem", color: "var(--text-muted)", marginTop: 4 }}>{corridorTarget.toFixed(2)}×</div>
            </div>

            <div>
              <div style={{ fontSize: "0.78rem", color: "var(--text-muted)", marginBottom: 4 }}>Corridor Ceiling</div>
              <input type="range" min={0.9} max={2.5} step={0.01} value={corridorCeiling} onChange={(e) => setCorridorCeiling(parseFloat(e.target.value))} style={{ width: "100%" }} />
              <div style={{ fontSize: "0.75rem", color: "var(--text-muted)", marginTop: 4 }}>{corridorCeiling.toFixed(2)}×</div>
            </div>
          </div>
        </div>
      )}

      <div
        style={{
          marginTop: 16,
          padding: 14,
          borderRadius: 14,
          border: "1px solid rgba(255,255,255,0.06)",
          background: "rgba(15,18,26,0.65)",
        }}
      >
        <div style={{ fontSize: "0.85rem", color: "var(--text-muted)", marginBottom: 10 }}>
          Coverage Simulation (Baseline + Merged Modes)
        </div>
        {!simResult ? (
          <div style={{ color: "var(--text-muted)", fontSize: "0.82rem" }}>
            Run a simulation to populate charts.
          </div>
        ) : (
          <Line data={chartData as any} options={chartOptions as any} />
        )}
      </div>
    </div>
  );
};

export default ReserveSystemPage;
