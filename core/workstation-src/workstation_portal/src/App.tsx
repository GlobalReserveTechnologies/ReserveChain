import React from "react";
import { useRoutes } from "react-router-dom";
import { routes } from "./router";
import ShellLayout from "./components/layout/ShellLayout";

const App: React.FC = () => {
  const element = useRoutes(routes);
  return <ShellLayout>{element}</ShellLayout>;
};

export default App;
