import React from "react";
import { useNavigate } from "react-router-dom";

const OperatorLogin: React.FC = () => {
  const navigate = useNavigate();
  return (
    <div>
      <h1 className="page-title">Operator Portal</h1>
      <p className="page-subtitle">
        Placeholder login surface for the operator portal.
      </p>
      <button
        style={{ padding: "8px 14px", borderRadius: "999px", border: "1px solid rgba(77,163,255,0.8)", background: "var(--accent-soft)", color: "var(--accent)", cursor: "pointer", fontSize: "0.85rem" }}
        onClick={() => navigate("/operators/dashboard")}
      >
        Simulate Login
      </button>
    </div>
  );
};

export default OperatorLogin;
