import React from "react";
import { RouteObject } from "react-router-dom";
import PortalSelector from "./pages/PortalSelector";
import ClientLogin from "./pages/client/ClientLogin";
import ClientDashboard from "./pages/client/ClientDashboard";
import OperatorLogin from "./pages/operator/OperatorLogin";
import OperatorDashboard from "./pages/operator/OperatorDashboard";
import GovernanceLogin from "./pages/governance/GovernanceLogin";
import GovernanceDashboard from "./pages/governance/GovernanceDashboard";
import AdminLogin from "./pages/admin/AdminLogin";
import AdminDashboard from "./pages/admin/AdminDashboard";
import ReserveSystemPage from "./pages/admin/ReserveSystemPage";
import ExplorerHome from "./pages/explorer/ExplorerHome";

export const routes: RouteObject[] = [
  { path: "/", element: <PortalSelector /> },
  {
    path: "/client",
    children: [
      { index: true, element: <ClientLogin /> },
      { path: "dashboard", element: <ClientDashboard /> },
    ],
  },
  {
    path: "/operators",
    children: [
      { index: true, element: <OperatorLogin /> },
      { path: "dashboard", element: <OperatorDashboard /> },
    ],
  },
  {
    path: "/gov",
    children: [
      { index: true, element: <GovernanceLogin /> },
      { path: "dashboard", element: <GovernanceDashboard /> },
    ],
  },
  {
    path: "/admin",
    children: [
      { index: true, element: <AdminLogin /> },
      { path: "dashboard", element: <AdminDashboard /> },
      { path: "reserve-system", element: <ReserveSystemPage /> },
    ],
  },
  {
    path: "/explorer",
    element: <ExplorerHome />,
  },
];
