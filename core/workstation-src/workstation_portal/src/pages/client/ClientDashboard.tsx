import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

type SessionInfo = {
  ok: boolean;
  role?: string;
  address?: string;
  expires?: string;
};

const ClientDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [sess, setSess] = useState<SessionInfo | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const res = await fetch("/api/session", { credentials: "include" });
        const js = (await res.json()) as SessionInfo;
        if (!js?.ok || js.role !== "client") {
          navigate("/client");
          return;
        }
        setSess(js);
      } catch {
        navigate("/client");
      }
    })();
  }, [navigate]);

  async function logout() {
    try {
      await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
    } finally {
      navigate("/client");
    }
  }

  return (
    <div>
      <h1 className="page-title">Client Dashboard</h1>
      <p className="page-subtitle">
        Authenticated session surface for client actions (staking, balances, transfers next).
      </p>

      <div style={{ maxWidth: 520, background: "rgba(255,255,255,0.02)", border: "1px solid rgba(255,255,255,0.06)", borderRadius: 14, padding: 16 }}>
        <div style={{ fontSize: 12, color: "var(--text-muted)", marginBottom: 10 }}>Session</div>
        <div style={{ fontSize: 13 }}>
          <div><span style={{ color: "var(--text-muted)" }}>Role:</span> {sess?.role || "—"}</div>
          <div><span style={{ color: "var(--text-muted)" }}>Address:</span> {sess?.address || "—"}</div>
          <div><span style={{ color: "var(--text-muted)" }}>Expires:</span> {sess?.expires || "—"}</div>
        </div>

        <button
          onClick={logout}
          style={{
            marginTop: 14,
            padding: "10px 14px",
            borderRadius: "999px",
            border: "1px solid rgba(255,255,255,0.14)",
            background: "rgba(255,255,255,0.03)",
            color: "var(--text-main)",
            cursor: "pointer",
            width: "100%",
          }}
        >
          Logout
        </button>
      </div>
    </div>
  );
};

export default ClientDashboard;
