import React from "react";
import { useNavigate } from "react-router-dom";
import PortalCard from "../components/common/PortalCard";

const PortalSelector: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div>
      <h1 className="page-title">Select Portal</h1>
      <p className="page-subtitle">
        Choose which ReserveChain workstation surface you want to access.
      </p>

      <div className="portal-grid">
        <PortalCard
          title="Client Portal"
          subtitle="Wallet, staking, balances and positions. Crypto-native access."
          tags={["Wallet auth", "Non-custodial", "Staking"]}
          ctaLabel="Enter Client Portal"
          onClick={() => navigate("/client")}
        />
        <PortalCard
          title="Operator Portal"
          subtitle="Validators and PoP nodes. Rewards, uptime, telemetry."
          tags={["Nodes", "Rewards", "PoP / RSX"]}
          ctaLabel="Enter Operator Portal"
          onClick={() => navigate("/operators")}
        />
        <PortalCard
          title="Governance Portal"
          subtitle="Proposals, votes, and economic parameters for ReserveChain."
          tags={["Council", "Voting", "Parameters"]}
          ctaLabel="Enter Governance Portal"
          onClick={() => navigate("/gov")}
        />
        <PortalCard
          title="Admin Portal"
          subtitle="Treasury, risk controls, compliance and control-plane operations."
          tags={["Compliance", "Treasury", "Risk"]}
          ctaLabel="Enter Admin Portal"
          onClick={() => navigate("/admin")}
        />
        <PortalCard
          title="Explorer"
          subtitle="Read-only view of the chain. Blocks, transactions, metrics."
          tags={["Public", "Read-only"]}
          ctaLabel="Open Explorer"
          onClick={() => navigate("/explorer")}
        />
      </div>
    </div>
  );
};

export default PortalSelector;
