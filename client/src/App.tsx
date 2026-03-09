import { BrowserRouter, Routes, Route, Link, Navigate } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { UserUpload } from "./components/UserUpload";
import { SolicitorDashboard } from "./components/SolicitorDashboard";
import { Login } from "./components/Login";
import "./App.css";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

function Navigation() {
  const { user, logout, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return null;
  }

  return (
    <nav
      style={{
        padding: "1rem",
        borderBottom: "1px solid #ccc",
        marginBottom: "1rem",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <div>
        <Link to="/user" style={{ marginRight: "1rem" }}>
          User Upload
        </Link>
        <Link to="/solicitor">Solicitor Dashboard</Link>
      </div>
      <div style={{ display: "flex", alignItems: "center", gap: "1rem" }}>
        <span>{user?.email}</span>
        <button onClick={logout} style={{ padding: "0.25rem 0.5rem" }}>
          Logout
        </button>
      </div>
    </nav>
  );
}

function AppRoutes() {
  return (
    <>
      <Navigation />
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route
          path="/user"
          element={
            <ProtectedRoute>
              <UserUpload />
            </ProtectedRoute>
          }
        />
        <Route
          path="/solicitor"
          element={
            <ProtectedRoute>
              <SolicitorDashboard />
            </ProtectedRoute>
          }
        />
        <Route path="/" element={<Navigate to="/login" replace />} />
      </Routes>
    </>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
