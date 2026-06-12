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
import ArgoCDPage from './pages/ArgoCDPage'
import RegistryPage from './pages/RegistryPage'
import TerraformPage from './pages/TerraformPage'
import AnsiblePage from './pages/AnsiblePage'
import VaultPage from './pages/VaultPage'
import CodePage from './pages/CodePage'
import PipelinePage from './pages/PipelinePage'
import AlertsPage from './pages/AlertsPage'
import MetricsPage from './pages/MetricsPage'
import SettingsPage from './pages/SettingsPage'
import DocumentationPage from './pages/DocumentationPage'

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
      <div
        style={{
          minHeight: '100vh',
          width: '100vw',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          background: '#1a1d28',
          color: '#ffffff',
          fontFamily: 'system-ui, sans-serif',
          fontSize: '1.25rem',
          gap: '1rem',
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: 99999,
        }}
        role="status"
        aria-live="polite"
      >
        <div style={{ width: 48, height: 48, border: '4px solid #4F8EF7', borderTopColor: '#4F8EF7', borderRadius: '50%', animation: 'spin 0.8s linear infinite', flexShrink: 0 }} />
        <span style={{ color: '#ffffff', fontWeight: 600 }}>Chargement...</span>
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
        <Route path="argocd" element={<ArgoCDPage />} />
        <Route path="registry" element={<RegistryPage />} />
        <Route path="terraform" element={<TerraformPage />} />
        <Route path="code" element={<CodePage />} />
        <Route path="ansible" element={<AnsiblePage />} />
        <Route path="vault" element={<VaultPage />} />
        <Route path="pipelines" element={<PipelinePage />} />
        <Route path="alerts" element={<AlertsPage />} />
        <Route path="metrics" element={<MetricsPage />} />
        <Route path="documentation" element={<DocumentationPage />} />
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
