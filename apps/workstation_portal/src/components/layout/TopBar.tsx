import React from "react";

const TopBar: React.FC = () => {
  return (
    <header className="topbar">
      <div className="topbar__title">ReserveChain Workstation</div>
      <div className="topbar__spacer" />
      <div className="topbar__badge">DevNet · Reserve System</div>
    </header>
  );
};

export default TopBar;
