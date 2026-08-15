import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { ThemeProvider, CssBaseline } from '@mui/material'
import App from './App'
import { theme } from './theme'
import { AuthProvider } from './contexts/AuthContext'
import { SocketProvider } from './contexts/SocketContext'
import { ProjectProvider } from './contexts/ProjectContext'
// Polices hébergées localement (évite CORS avec Google Fonts)
import '@fontsource/inter/300.css'
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/600.css'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Kura affiche l'état d'infrastructures qui bougent sans nous : une
      // synchronisation ArgoCD, un pod qui redémarre, un pipeline qui se
      // termine. Les données étaient considérées fraîches pendant 5 minutes et
      // rien ne les rafraîchissait : l'écran restait figé jusqu'à ce que
      // l'utilisateur clique, ce qui obligeait à un bouton sur chaque page.
      refetchOnWindowFocus: true,
      retry: 1,
      staleTime: 10 * 1000,
      // 15 s plutôt que 5 : chaque page déclenche plusieurs requêtes, et un
      // déploiement mono-machine encaisse mal une multiplication par trois du
      // trafic de fond. Les vues qui suivent une opération en cours gardent
      // leur propre intervalle, plus court.
      refetchInterval: 15 * 1000,
      // Onglet en arrière-plan : inutile d'interroger la plateforme pour un
      // écran que personne ne regarde.
      refetchIntervalInBackground: false,
    },
  },
})

ReactDOM.createRoot(document.getElementById('root')!).render(
  <QueryClientProvider client={queryClient}>
    <BrowserRouter>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <AuthProvider>
          <SocketProvider>
            <ProjectProvider>
              <App />
            </ProjectProvider>
          </SocketProvider>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  </QueryClientProvider>,
)
