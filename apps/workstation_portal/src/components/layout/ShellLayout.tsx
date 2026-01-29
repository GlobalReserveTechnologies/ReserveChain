import React from "react";
import TopBar from "./TopBar";
import SideNav from "./SideNav";

interface Props {
  children: React.ReactNode;
}

const ShellLayout: React.FC<Props> = ({ children }) => {
  return (
    <div className="app-shell">
      <aside className="app-shell__sidebar">
        <SideNav />
      </aside>
      <div className="app-shell__main">
        <TopBar />
        <main className="app-main-content">{children}</main>
      </div>
    </div>
  );
};

export default ShellLayout;
