import { useEffect, useState } from "react";
import { getToken, logout, setUnauthorizedHandler } from "./api";
import Login from "./components/Login";
import Dashboard from "./components/Dashboard";

export default function App() {
  const [authed, setAuthed] = useState(() => getToken() !== null);

  useEffect(() => setUnauthorizedHandler(() => setAuthed(false)), []);

  if (!authed) return <Login onLogin={() => setAuthed(true)} />;
  return (
    <Dashboard
      onLogout={() => {
        logout();
        setAuthed(false);
      }}
    />
  );
}
