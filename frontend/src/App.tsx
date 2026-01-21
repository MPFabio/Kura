import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import ModulesPage from './pages/ModulesPage'
import K8sPage from './pages/K8sPage'
import TerraformPage from './pages/TerraformPage'
import AnsiblePage from './pages/AnsiblePage'
import PipelinePage from './pages/PipelinePage'
import AlertsPage from './pages/AlertsPage'
import MetricsPage from './pages/MetricsPage'
import SettingsPage from './pages/SettingsPage'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()

  if (loading) {
    return <div>Chargement...</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <Layout />
          </PrivateRoute>
        }
      >
        <Route index element={<ModulesPage />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="modules" element={<ModulesPage />} />
        <Route path="k8s" element={<K8sPage />} />
        <Route path="terraform" element={<TerraformPage />} />
        <Route path="ansible" element={<AnsiblePage />} />
        <Route path="pipelines" element={<PipelinePage />} />
        <Route path="alerts" element={<AlertsPage />} />
        <Route path="metrics" element={<MetricsPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
    </Routes>
  )
}

export default App
