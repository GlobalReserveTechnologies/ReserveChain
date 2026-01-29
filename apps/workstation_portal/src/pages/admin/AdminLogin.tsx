import React from "react";
import { useNavigate } from "react-router-dom";

const AdminLogin: React.FC = () => {
  const navigate = useNavigate();
  return (
    <div>
      <h1 className="page-title">Admin Portal</h1>
      <p className="page-subtitle">
        Placeholder login surface for the admin portal.
      </p>
      <button
        style={{ padding: "8px 14px", borderRadius: "999px", border: "1px solid rgba(77,163,255,0.8)", background: "var(--accent-soft)", color: "var(--accent)", cursor: "pointer", fontSize: "0.85rem" }}
        onClick={() => navigate("/admin/dashboard")}
      >
        Simulate Login
      </button>
    </div>
  );
};

export default AdminLogin;
