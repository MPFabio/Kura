import { useNavigate } from 'react-router-dom'
import { Grid, Typography, Box } from '@mui/material'
import {
  Cloud as CloudIcon,
  Storage as StorageIcon,
  PlayArrow as PlayArrowIcon,
  BarChart as BarChartIcon,
} from '@mui/icons-material'
import ModuleCard from '../components/ModuleCard'

interface Module {
  id: string
  name: string
  icon: React.ReactNode
  path: string
  active: boolean
  inactive?: boolean
  deploying?: boolean
  subtitle?: string
  statusText?: string
  description?: string
}

export default function ModulesPage() {
  const navigate = useNavigate()

  const modules: Module[] = [
    {
      id: 'terraform',
      name: 'Terraform',
      icon: <StorageIcon sx={{ fontSize: 80 }} />,
      path: '/terraform',
      active: true,
      deploying: true,
      statusText: 'Deployment in Progress',
      description: 'Terraform will deploy the given plan in your infrastructure...',
    },
    {
      id: 'kubernetes',
      name: 'Kubernetes',
      icon: <CloudIcon sx={{ fontSize: 80 }} />,
      path: '/k8s',
      active: false,
      inactive: true,
      subtitle: 'Kubernetes',
    },
    {
      id: 'ansible',
      name: 'Ansible',
      icon: <PlayArrowIcon sx={{ fontSize: 80 }} />,
      path: '/ansible',
      active: false,
      inactive: true,
      subtitle: 'Kubernetes',
    },
    {
      id: 'monitoring',
      name: 'Monitoring',
      icon: <BarChartIcon sx={{ fontSize: 80 }} />,
      path: '/metrics',
      active: false,
      inactive: true,
      subtitle: 'Kubernetes',
    },
  ]

  return (
    <Box>
      <Typography
        variant="h4"
        sx={{
          mb: 4,
          fontWeight: 600,
          color: '#FFFFFF',
          fontFamily: '"Inter", sans-serif',
          letterSpacing: '0.02em',
        }}
      >
        Modules
      </Typography>

      <Grid container spacing={4}>
        {modules.map((module) => (
          <Grid item xs={12} sm={6} key={module.id}>
            <ModuleCard
              active={module.active}
              inactive={module.inactive}
              deploying={module.deploying}
              onClick={() => navigate(module.path)}
              sx={{
                minHeight: '400px',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                textAlign: 'center',
                py: 4,
                px: 3,
              }}
            >
              {module.active ? (
                <Box sx={{ position: 'relative', width: '100%', height: '100%', minHeight: '350px' }}>
                  {/* Icône centrale avec filaments lumineux */}
                  <Box
                    sx={{
                      position: 'relative',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      mb: 4,
                    }}
                  >
                    {/* Filaments animés autour de l'icône */}
                    <Box
                      sx={{
                        position: 'absolute',
                        width: 200,
                        height: 200,
                        border: '1px solid transparent',
                        borderTop: '2px solid rgba(0, 255, 255, 0.4)',
                        borderRight: '2px solid rgba(0, 255, 255, 0.3)',
                        borderRadius: '50%',
                        animation: 'constructAnimation 8s linear infinite',
                      }}
                    />
                    <Box
                      sx={{
                        position: 'absolute',
                        width: 180,
                        height: 180,
                        border: '1px solid transparent',
                        borderBottom: '1px solid rgba(0, 255, 255, 0.2)',
                        borderLeft: '1px solid rgba(0, 255, 255, 0.2)',
                        borderRadius: '50%',
                        animation: 'constructAnimation 6s linear infinite reverse',
                      }}
                    />
                    {/* Lignes de circuit */}
                    {[...Array(6)].map((_, i) => (
                      <Box
                        key={i}
                        sx={{
                          position: 'absolute',
                          width: 2,
                          height: 60,
                          background: `linear-gradient(180deg, transparent, rgba(0, 255, 255, ${0.3 - i * 0.05}))`,
                          transform: `rotate(${i * 60}deg)`,
                          transformOrigin: '0 100px',
                          top: '50%',
                          left: '50%',
                          marginTop: '-100px',
                          marginLeft: '-1px',
                          borderRadius: 1,
                        }}
                      />
                    ))}
                    {/* Icône principale */}
                    <Box
                      sx={{
                        position: 'relative',
                        zIndex: 1,
                        color: '#00FFFF',
                        filter: 'drop-shadow(0 0 20px rgba(0, 255, 255, 0.9)) drop-shadow(0 0 40px rgba(0, 255, 255, 0.6))',
                        animation: 'breathingGlow 3s ease-in-out infinite',
                      }}
                    >
                      {module.icon}
                    </Box>
                  </Box>

                  {/* Texte d'état */}
                  <Typography
                    variant="h6"
                    sx={{
                      fontFamily: '"JetBrains Mono", monospace',
                      color: '#FFFFFF',
                      mb: 1,
                      textShadow: '0 0 10px rgba(0, 255, 255, 0.6)',
                      letterSpacing: '0.05em',
                      fontWeight: 500,
                    }}
                  >
                    {module.statusText}
                  </Typography>

                  {/* Description */}
                  {module.description && (
                    <Typography
                      variant="body2"
                      sx={{
                        fontFamily: '"JetBrains Mono", monospace',
                        color: '#A0A0A0',
                        fontSize: '0.875rem',
                        maxWidth: '400px',
                        mx: 'auto',
                        lineHeight: 1.6,
                      }}
                    >
                      {module.description}
                    </Typography>
                  )}

                  {/* Éléments secondaires en bas */}
                  <Box
                    sx={{
                      position: 'absolute',
                      bottom: 16,
                      left: 16,
                      right: 16,
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      zIndex: 2,
                    }}
                  >
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Box
                        sx={{
                          width: 32,
                          height: 32,
                          borderRadius: '50%',
                          background: 'rgba(0, 255, 255, 0.1)',
                          border: '1px solid rgba(0, 255, 255, 0.4)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          color: '#00FFFF',
                          fontFamily: '"JetBrains Mono", monospace',
                          fontWeight: 600,
                          fontSize: '0.875rem',
                          boxShadow: '0 0 10px rgba(0, 255, 255, 0.3)',
                        }}
                      >
                        D
                      </Box>
                      <Typography
                        variant="body2"
                        sx={{
                          fontFamily: '"JetBrains Mono", monospace',
                          color: '#FFFFFF',
                          fontSize: '0.875rem',
                          textShadow: '0 0 5px rgba(0, 255, 255, 0.4)',
                        }}
                      >
                        Dobernaton
                      </Typography>
                    </Box>

                    <Box
                      sx={{
                        width: 32,
                        height: 32,
                        borderRadius: '50%',
                        background: 'rgba(0, 255, 255, 0.1)',
                        border: '1px solid rgba(0, 255, 255, 0.4)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: '#00FFFF',
                        boxShadow: '0 0 10px rgba(0, 255, 255, 0.3)',
                      }}
                    >
                      ↓
                    </Box>
                  </Box>

                  {/* Flèche en haut à droite */}
                  <Box
                    sx={{
                      position: 'absolute',
                      top: 16,
                      right: 16,
                      color: '#00FFFF',
                      fontSize: '1.2rem',
                      filter: 'drop-shadow(0 0 8px rgba(0, 255, 255, 0.8))',
                      zIndex: 2,
                    }}
                  >
                    ↓
                  </Box>
                </Box>
              ) : (
                <>
                  {/* Module inactif - très discret */}
                  <Box
                    sx={{
                      color: 'rgba(160, 160, 160, 0.2)',
                      mb: 2,
                      filter: 'none',
                    }}
                  >
                    {module.icon}
                  </Box>
                  <Typography
                    variant="h6"
                    sx={{
                      fontFamily: '"Inter", sans-serif',
                      color: 'rgba(160, 160, 160, 0.4)',
                      mb: 1,
                      fontWeight: 500,
                    }}
                  >
                    {module.name}
                  </Typography>
                  {module.subtitle && (
                    <Typography
                      variant="body2"
                      sx={{
                        fontFamily: '"JetBrains Mono", monospace',
                        color: 'rgba(96, 96, 96, 0.5)',
                        fontSize: '0.75rem',
                      }}
                    >
                      {module.subtitle}
                    </Typography>
                  )}
                </>
              )}
            </ModuleCard>
          </Grid>
        ))}
      </Grid>
    </Box>
  )
}
