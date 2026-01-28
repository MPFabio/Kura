import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Grid, Box, Chip, Typography } from '@mui/material'
import {
  CheckCircle as CheckCircleIcon,
  Schedule as ScheduleIcon,
} from '@mui/icons-material'
import ModuleCard from '../components/ModuleCard'
import ModuleTitle from '../components/ModuleTitle'
import { ModuleBodyText, ModuleSecondaryText, ModuleSubtitle, ModuleCaption } from '../components/ModuleText'
import TerraformIcon from '../components/icons/TerraformIcon'
import KubernetesIcon from '../components/icons/KubernetesIcon'
import AnsibleIcon from '../components/icons/AnsibleIcon'
import MonitoringIcon from '../components/icons/MonitoringIcon'
import { useProject } from '../contexts/ProjectContext'
import { terraformService } from '../services/terraformService'
import { clusterService } from '../services/clusterService'

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
  stats?: {
    label: string
    value: string | number
  }[]
  features?: string[]
  status?: 'active' | 'inactive' | 'deploying' | 'available'
}

export default function ModulesPage() {
  const navigate = useNavigate()
  const { currentProject } = useProject()

  const { data: terraformStatesData } = useQuery({
    queryKey: ['terraform-states', currentProject?.id],
    queryFn: () => terraformService.getStates(currentProject!.id),
    enabled: !!currentProject?.id,
  })

  const { data: clustersData } = useQuery({
    queryKey: ['clusters', currentProject?.id],
    queryFn: () => clusterService.getClusters(currentProject!.id),
    enabled: !!currentProject?.id,
  })

  const statesCount = terraformStatesData?.items?.length ?? 0
  const clustersCount = clustersData?.items?.length ?? 0

  const modules: Module[] = [
    {
      id: 'terraform',
      name: 'Terraform',
      icon: <TerraformIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={true} />,
      path: '/terraform',
      active: true,
      deploying: false,
      status: 'active',
      statusText: 'Module actif',
      description: 'Gestion complète de vos états Terraform avec synchronisation cloud et détection de drift en temps réel.',
      stats: [
        { label: 'États', value: String(statesCount) },
        { label: 'Sources', value: String(statesCount) },
        { label: 'Drifts', value: '0' },
      ],
      features: [
        'Synchronisation S3, Azure, GCP',
        'Détection de drift automatique',
        'Visualisation des ressources',
      ],
    },
    {
      id: 'kubernetes',
      name: 'Kubernetes',
      icon: <KubernetesIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={true} />,
      path: '/k8s',
      active: true,
      status: 'active',
      statusText: 'Module actif',
      description: 'Gestion complète de vos clusters Kubernetes avec terminal interactif et actions en masse.',
      stats: [
        { label: 'Clusters', value: String(clustersCount) },
        { label: 'Pods', value: '0' },
        { label: 'Services', value: '0' },
      ],
      features: [
        'Gestion multi-clusters',
        'Terminal interactif',
        'Actions en masse',
      ],
    },
    {
      id: 'ansible',
      name: 'Ansible',
      icon: <AnsibleIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={false} />,
      path: '/ansible',
      active: false,
      inactive: true,
      status: 'available',
      subtitle: 'Bientôt disponible',
      description: 'Automatisation de vos déploiements avec Ansible Tower. Configuration et exécution de playbooks.',
      features: [
        'Intégration Ansible Tower',
        'Gestion des inventaires',
        'Exécution de playbooks',
      ],
    },
    {
      id: 'monitoring',
      name: 'Monitoring',
      icon: <MonitoringIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={false} />,
      path: '/metrics',
      active: false,
      inactive: true,
      status: 'available',
      subtitle: 'Bientôt disponible',
      description: 'Surveillance complète de votre infrastructure avec métriques, alertes et dashboards personnalisables.',
      features: [
        'Métriques en temps réel',
        'Alertes configurables',
        'Dashboards Grafana',
      ],
    },
  ]

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 5 }}>
        <ModuleTitle>
          Modules
        </ModuleTitle>
        <Chip
          label={`${modules.filter(m => m.active).length} actif${modules.filter(m => m.active).length > 1 ? 's' : ''} sur ${modules.length}`}
          sx={{
            backgroundColor: 'rgba(0, 229, 255, 0.1)',
            color: '#00E5FF',
            border: '1px solid rgba(0, 229, 255, 0.3)',
            fontFamily: '"JetBrains Mono", monospace',
            fontSize: '0.875rem',
            fontWeight: 500,
            boxShadow: '0 0 8px rgba(0, 229, 255, 0.2)',
          }}
        />
      </Box>

      <Grid container spacing={4}>
        {modules.map((module, idx) => {
          return (
            <Grid item xs={12} sm={6} key={module.id}>
              <ModuleCard
                active={module.active}
                inactive={module.inactive}
                deploying={module.deploying}
                onClick={() => navigate(module.path)}
                sx={{
                minHeight: '520px',
                display: 'flex',
                flexDirection: 'column',
                height: '100%',
                '--card-delay': `${idx * 0.3}s`,
              }}
            >
              {module.active ? (
                <Box sx={{ position: 'relative', width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
                  {/* Header avec icône et statut */}
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', mb: 3 }}>
                    <Box
                      sx={{
                        position: 'relative',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 100,
                        height: 100,
                      }}
                    >
                      {/* Cercles orbitaux */}
                      <Box
                        sx={{
                          position: 'absolute',
                          width: 100,
                          height: 100,
                          border: '1px solid rgba(0, 229, 255, 0.3)',
                          borderRadius: '50%',
                          animation: 'constructAnimation 8s linear infinite',
                        }}
                      />
                      <Box
                        sx={{
                          position: 'absolute',
                          width: 120,
                          height: 120,
                          border: '1px solid transparent',
                          borderTop: '1px solid rgba(0, 229, 255, 0.5)',
                          borderRight: '1px solid rgba(179, 136, 255, 0.5)',
                          borderBottom: '1px solid rgba(179, 136, 255, 0.3)',
                          borderLeft: '1px solid rgba(0, 229, 255, 0.3)',
                          borderRadius: '50%',
                          animation: 'constructAnimation 6s linear infinite reverse',
                        }}
                      />
                      {/* Icône principale */}
                      <Box
                        sx={{
                          position: 'relative',
                          zIndex: 1,
                          color: '#00E5FF',
                          filter: 'drop-shadow(0 0 20px rgba(0, 229, 255, 0.8)) drop-shadow(0 0 40px rgba(179, 136, 255, 0.5))',
                          animation: 'breathingGlow 3s ease-in-out infinite',
                        }}
                      >
                        {module.icon}
                      </Box>
                    </Box>
                    <Chip
                      icon={<CheckCircleIcon />}
                      label={module.statusText}
                      color="success"
                      size="small"
                      sx={{
                        backgroundColor: 'rgba(102, 187, 106, 0.2)',
                        color: '#66BB6A',
                        border: '1px solid rgba(102, 187, 106, 0.4)',
                        fontWeight: 500,
                      }}
                    />
                  </Box>

                  {/* Nom du module */}
                  <ModuleSubtitle
                    sx={{
                      mb: 2.5,
                      background: 'linear-gradient(135deg, #00E5FF, #B388FF)',
                      backgroundClip: 'text',
                      WebkitBackgroundClip: 'text',
                      WebkitTextFillColor: 'transparent',
                      fontSize: '1.5rem',
                    }}
                  >
                    {module.name}
                  </ModuleSubtitle>

                  {/* Description */}
                  {module.description && (
                    <ModuleSecondaryText sx={{ mb: 3 }}>
                      {module.description}
                    </ModuleSecondaryText>
                  )}

                  {/* Statistiques */}
                  {module.stats && (
                    <Box sx={{ display: 'flex', gap: 3, mb: 4 }}>
                      {module.stats.map((stat, idx) => (
                        <Box
                          key={idx}
                          sx={{
                            flex: 1,
                            textAlign: 'center',
                            p: 2.5,
                            borderRadius: 3,
                            background: idx === 0 
                              ? 'linear-gradient(135deg, rgba(0, 229, 255, 0.12), rgba(0, 229, 255, 0.05))'
                              : idx === 1
                              ? 'linear-gradient(135deg, rgba(179, 136, 255, 0.12), rgba(179, 136, 255, 0.05))'
                              : 'linear-gradient(135deg, rgba(0, 229, 255, 0.1), rgba(179, 136, 255, 0.05))',
                            backdropFilter: 'blur(10px)',
                            border: 'none',
                            boxShadow: idx === 0
                              ? '0 4px 16px rgba(0, 229, 255, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.1)'
                              : idx === 1
                              ? '0 4px 16px rgba(179, 136, 255, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.1)'
                              : '0 4px 14px rgba(0, 229, 255, 0.12), inset 0 1px 0 rgba(255, 255, 255, 0.1)',
                            transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                            '&:hover': {
                              boxShadow: idx === 0
                                ? '0 6px 20px rgba(0, 229, 255, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.12)'
                                : idx === 1
                                ? '0 6px 20px rgba(179, 136, 255, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.12)'
                                : '0 6px 18px rgba(0, 229, 255, 0.18), inset 0 1px 0 rgba(255, 255, 255, 0.12)',
                            },
                          }}
                        >
                          <Typography
                            variant="h4"
                            sx={{
                              fontFamily: '"JetBrains Mono", monospace',
                              color: idx === 0 ? '#00E5FF' : idx === 1 ? '#B388FF' : '#00E5FF',
                              fontWeight: 700,
                              mb: 0.75,
                              textShadow: idx === 0
                                ? '0 0 10px rgba(0, 229, 255, 0.5)'
                                : idx === 1
                                ? '0 0 10px rgba(179, 136, 255, 0.5)'
                                : '0 0 8px rgba(0, 229, 255, 0.4)',
                            }}
                          >
                            {stat.value}
                          </Typography>
                          <ModuleCaption
                            sx={{
                              textTransform: 'uppercase',
                              letterSpacing: '0.08em',
                              fontWeight: 500,
                            }}
                          >
                            {stat.label}
                          </ModuleCaption>
                        </Box>
                      ))}
                    </Box>
                  )}

                  {/* Features */}
                  {module.features && (
                    <Box sx={{ mt: 'auto', pt: 4 }}>
                      <ModuleCaption
                        sx={{
                          textTransform: 'uppercase',
                          letterSpacing: '0.12em',
                          mb: 2.5,
                          display: 'block',
                          fontWeight: 600,
                        }}
                      >
                        Fonctionnalités
                      </ModuleCaption>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                        {module.features.map((feature, idx) => (
                          <Box
                            key={idx}
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 2,
                              p: 1.5,
                              borderRadius: 2,
                              background: 'linear-gradient(135deg, rgba(0, 0, 0, 0.2), rgba(0, 0, 0, 0.1))',
                              backdropFilter: 'blur(10px)',
                              border: 'none',
                              transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                              '&:hover': {
                                background: 'linear-gradient(135deg, rgba(0, 229, 255, 0.08), rgba(179, 136, 255, 0.05))',
                                boxShadow: '0 2px 8px rgba(0, 229, 255, 0.08)',
                              },
                            }}
                          >
                            <Box
                              sx={{
                                width: 8,
                                height: 8,
                                borderRadius: '50%',
                                backgroundColor: idx % 2 === 0 ? '#00E5FF' : '#B388FF',
                                boxShadow: idx % 2 === 0 
                                  ? '0 0 10px rgba(0, 229, 255, 0.7)' 
                                  : '0 0 10px rgba(179, 136, 255, 0.7)',
                                flexShrink: 0,
                              }}
                            />
                            <ModuleBodyText>
                              {feature}
                            </ModuleBodyText>
                          </Box>
                        ))}
                      </Box>
                    </Box>
                  )}
                </Box>
              ) : (
                <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', justifyContent: 'center', alignItems: 'center', textAlign: 'center', p: 4 }}>
                  {/* Icône avec effet subtil amélioré */}
                  <Box
                    sx={{
                      mb: 4,
                      position: 'relative',
                      width: 100,
                      height: 100,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      '&::before': {
                        content: '""',
                        position: 'absolute',
                        width: '100%',
                        height: '100%',
                        borderRadius: '50%',
                        background: 'radial-gradient(circle, rgba(179, 136, 255, 0.15) 0%, transparent 70%)',
                        zIndex: 0,
                        animation: 'breathingGlow 4s ease-in-out infinite',
                      },
                      '&::after': {
                        content: '""',
                        position: 'absolute',
                        width: '120%',
                        height: '120%',
                        borderRadius: '50%',
                        border: '1px solid rgba(179, 136, 255, 0.3)',
                        zIndex: 0,
                        animation: 'constructAnimation 10s linear infinite',
                      },
                    }}
                  >
                    <Box
                      sx={{
                        position: 'relative',
                        zIndex: 1,
                        color: 'rgba(179, 136, 255, 0.6)',
                        filter: 'drop-shadow(0 0 12px rgba(179, 136, 255, 0.5))',
                        transition: 'all 0.3s ease',
                        '&:hover': {
                          color: 'rgba(179, 136, 255, 0.8)',
                          filter: 'drop-shadow(0 0 18px rgba(179, 136, 255, 0.7))',
                        },
                      }}
                    >
                      {module.icon}
                    </Box>
                  </Box>

                  {/* Nom du module */}
                  <ModuleSubtitle
                    sx={{
                      mb: 1,
                      color: 'rgba(255, 255, 255, 0.7)',
                      fontSize: '1.25rem',
                      fontWeight: 500,
                    }}
                  >
                    {module.name}
                  </ModuleSubtitle>

                  {/* Subtitle */}
                  {module.subtitle && (
                    <Chip
                      label={module.subtitle}
                      size="small"
                      icon={<ScheduleIcon />}
                      sx={{
                        mb: 2,
                        backgroundColor: 'rgba(179, 136, 255, 0.1)',
                        color: '#B388FF',
                        border: '1px solid rgba(179, 136, 255, 0.3)',
                        fontSize: '0.75rem',
                      }}
                    />
                  )}

                  {/* Description */}
                  {module.description && (
                    <ModuleSecondaryText
                      sx={{
                        mb: 4,
                        color: 'rgba(255, 255, 255, 0.6)',
                        maxWidth: '320px',
                        mx: 'auto',
                      }}
                    >
                      {module.description}
                    </ModuleSecondaryText>
                  )}

                  {/* Features */}
                  {module.features && (
                    <Box sx={{ width: '100%', mt: 'auto' }}>
                      <ModuleCaption
                        sx={{
                          color: 'rgba(255, 255, 255, 0.4)',
                          fontSize: '0.75rem',
                          textTransform: 'uppercase',
                          letterSpacing: '0.1em',
                          mb: 1.5,
                          display: 'block',
                        }}
                      >
                        Fonctionnalités prévues
                      </ModuleCaption>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                        {module.features.map((feature, idx) => (
                          <Box
                            key={idx}
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 1,
                              justifyContent: 'center',
                            }}
                          >
                            <Box
                              sx={{
                                width: 4,
                                height: 4,
                                borderRadius: '50%',
                                backgroundColor: 'rgba(179, 136, 255, 0.4)',
                              }}
                            />
                            <ModuleSecondaryText
                              sx={{
                                color: 'rgba(255, 255, 255, 0.5)',
                                fontSize: '0.8125rem',
                              }}
                            >
                              {feature}
                            </ModuleSecondaryText>
                          </Box>
                        ))}
                      </Box>
                    </Box>
                  )}
                </Box>
              )}
            </ModuleCard>
          </Grid>
          )
        })}
      </Grid>
    </Box>
  )
}
