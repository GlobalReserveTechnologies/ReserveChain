import React from "react";
import { NavLink } from "react-router-dom";

const SideNav: React.FC = () => {
  const navItems = [
    { path: "/", label: "Portal Selector" },
    { path: "/client", label: "Client Portal" },
    { path: "/operators", label: "Operator Portal" },
    { path: "/gov", label: "Governance Portal" },
    { path: "/admin/dashboard", label: "Admin Portal" },
    { path: "/admin/reserve-system", label: "Reserve System" },
    { path: "/explorer", label: "Explorer" },
  ];

  return (
    <nav className="sidenav">
      <div className="sidenav__section-title">PORTALS</div>
      {navItems.map((item) => (
        <NavLink
          key={item.path}
          to={item.path}
          className={({ isActive }) =>
            ["sidenav__item", isActive ? "sidenav__item--active" : ""].join(" ")
          }
        >
          <span>{item.label}</span>
        </NavLink>
      ))}
    </nav>
  );
};

export default SideNav;
