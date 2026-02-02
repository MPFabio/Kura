import { Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { useAuth } from './contexts/AuthContext'
import { useProject } from './contexts/ProjectContext'
import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import LandingPage from './pages/LandingPage'
import ProjectsPage from './pages/ProjectsPage'
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

function ProjectRoute({ children }: { children: React.ReactNode }) {
  const { currentProject, loading } = useProject()

  if (loading) {
    return <div>Chargement...</div>
  }

  if (!currentProject) {
    return <Navigate to="/projects" replace />
  }

  return <>{children}</>
}

function AppRoot() {
  const { user, loading } = useAuth()
  const location = useLocation()
  const isPublicPath = location.pathname === '/' || location.pathname === '/login' || location.pathname === '/register'

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', background: '#0d0e12', color: '#f0f0f0' }}>
        Chargement...
      </div>
    )
  }

  if (!user && isPublicPath) {
    if (location.pathname === '/') return <LandingPage />
    if (location.pathname === '/login') return <LoginPage />
    if (location.pathname === '/register') return <RegisterPage />
  }

  if (!user) {
    return <Navigate to="/" replace />
  }

  if (user && (location.pathname === '/login' || location.pathname === '/register')) {
    return <Navigate to="/projects" replace />
  }

  return (
    <Routes>
      <Route path="/projects" element={<ProjectsPage />} />
      <Route
        path="/"
        element={
          <ProjectRoute>
            <Layout />
          </ProjectRoute>
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

function App() {
  return (
    <Routes>
      <Route path="/*" element={<AppRoot />} />
    </Routes>
  )
}

export default App
