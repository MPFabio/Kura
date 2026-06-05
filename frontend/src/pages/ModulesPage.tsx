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
import PipelinesIcon from '../components/icons/PipelinesIcon'
import MonitoringIcon from '../components/icons/MonitoringIcon'
import { useProject } from '../contexts/ProjectContext'
import { terraformService } from '../services/terraformService'
import { clusterService } from '../services/clusterService'
import { ansibleService } from '../services/ansibleService'
import { pipelineService } from '../services/pipelineService'
import { k8sService } from '../services/k8sService'
import { kuraColors } from '../theme'

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

  const { data: ansibleJobsData } = useQuery({
    queryKey: ['ansible-jobs-summary'],
    queryFn: () => ansibleService.getJobs(),
    retry: false,
  })

  const { data: ansibleInventoriesData } = useQuery({
    queryKey: ['ansible-inventories-summary'],
    queryFn: () => ansibleService.getInventories(),
    retry: false,
  })

  const { data: ansibleTemplatesData } = useQuery({
    queryKey: ['ansible-templates-summary'],
    queryFn: () => ansibleService.getJobTemplates(),
    retry: false,
  })

  // Namespaces pour agréger pods et services
  const SYSTEM_NAMESPACES = ['kube-system', 'kube-public', 'kube-node-lease', 'gke-system', 'gmp-system', 'gmp-public']

  const { data: namespacesData } = useQuery({
    queryKey: ['k8s-namespaces-modules'],
    queryFn: () => k8sService.getNamespaces(),
    retry: false,
  })

  const userNamespaces = (namespacesData?.items ?? [])
    .map((ns: any) => ns.name ?? ns.metadata?.name ?? ns)
    .filter((name: string) => !SYSTEM_NAMESPACES.includes(name))

  const { data: podsData } = useQuery({
    queryKey: ['k8s-pods-all-modules', userNamespaces],
    queryFn: async () => {
      const results = await Promise.all(
        userNamespaces.map((ns: string) => k8sService.getPods(ns).catch(() => ({ items: [] })))
      )
      return results.flatMap(r => r.items ?? [])
    },
    enabled: userNamespaces.length > 0,
    retry: false,
  })

  const { data: servicesData } = useQuery({
    queryKey: ['k8s-services-all-modules', userNamespaces],
    queryFn: async () => {
      const results = await Promise.all(
        userNamespaces.map((ns: string) => k8sService.getServices(ns).catch(() => ({ items: [] })))
      )
      return results.flatMap(r => r.items ?? [])
    },
    enabled: userNamespaces.length > 0,
    retry: false,
  })

  const podsCount = podsData?.length ?? 0
  const servicesCount = servicesData?.length ?? 0

  const hasProject = !!currentProject?.id
  const statesCount = hasProject ? (terraformStatesData?.items?.length ?? 0) : null
  const clustersCount = hasProject ? (clustersData?.items?.length ?? 0) : null
  const ansibleJobsCount = ansibleJobsData?.items?.length ?? 0
  const ansibleInventoriesCount = ansibleInventoriesData?.items?.length ?? 0
  const ansibleTemplatesCount = ansibleTemplatesData?.items?.length ?? 0
  const formatStat = (n: number | null) => (n === null ? '—' : String(n))
  
  const { data: pipelineData } = useQuery({
    queryKey: ['pipeline-runs'],
    queryFn: () => pipelineService.getRuns({ limit: 50 }),
  })


    const runs = pipelineData?.runs ?? []

    const runsCount = runs.length

    const pipelinesCount = new Set(
      runs.map(r => `${r.repository}-${r.workflow_name}`)
    ).size

    const lastRun = runs[0]

    const formatDate = (s?: string) => {
      if (!s) return '—'
      return new Date(s).toLocaleString('fr-FR')
    }


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
        { label: 'États', value: formatStat(statesCount) },
        { label: 'Sources', value: formatStat(statesCount) },
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
        { label: 'Clusters', value: formatStat(clustersCount) },
        { label: 'Pods', value: String(podsCount) },
        { label: 'Services', value: String(servicesCount) },
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
      icon: <AnsibleIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={true} />,
      path: '/ansible',
      active: true,
      status: 'active',
      statusText: 'Module actif',
      description: 'Automatisation de vos déploiements via Ansible Semaphore. Gestion des jobs, inventaires, templates et exécution de playbooks.',
      stats: [
        { label: 'Jobs', value: formatStat(ansibleJobsCount) },
        { label: 'Inventaires', value: formatStat(ansibleInventoriesCount) },
        { label: 'Templates', value: formatStat(ansibleTemplatesCount) },
      ],
      features: [
        'Intégration Ansible Semaphore',
        'Gestion des inventaires et hôtes',
        'Exécution de playbooks et templates',
      ],
    },
    {
      id: 'pipelines',
      name: 'Pipelines',
      icon: <PipelinesIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={true} />,
      path: '/pipelines',
      active: true,
      status: 'active',
      statusText: 'Module actif',
      description: 'CI/CD et pipelines de déploiement. Déclenchement, suivi des runs et intégration avec vos dépôts.',
      stats: [
        { label: 'Pipelines', value: pipelinesCount },
        { label: 'Runs', value: runsCount },
        {
          label: 'Dernier run',
          value: lastRun ? formatDate(lastRun.created_at) : '—',
        },
      ],
      features: [
        'Workflows GitHub Actions / GitLab CI',
        'Déploiement Terraform et K8s',
        'Historique et statut des runs',
      ],
    },
    {
      id: 'monitoring',
      name: 'Monitoring',
      icon: <MonitoringIcon sx={{ fontSize: 80, width: 80, height: 80 }} active={true} />,
      path: '/metrics',
      active: true,
      inactive: false,
      status: 'active',
      subtitle: 'Prometheus · Grafana',
      description: 'Surveillance complète de votre infrastructure avec métriques en temps réel, health checks et dashboards Grafana.',
      stats: [
        { label: 'Services', value: '6' },
        { label: 'Alertes', value: '0' },
        { label: 'Dashboard', value: 'Grafana' },
      ],
      features: [
        'Health checks en temps réel',
        'Métriques Prometheus',
        'Dashboard Grafana intégré',
      ],
    },
  ]

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 6, pb: 3, borderBottom: '2px solid rgba(0, 229, 255, 0.15)' }}>
        <ModuleTitle>
          Modules
        </ModuleTitle>
        <Box
          sx={{
            px: 2,
            py: 0.75,
            backgroundColor: '#2c2f3f',
            color: '#4F8EF7',
            border: '2px solid #4F8EF7',
            fontFamily: '"JetBrains Mono", monospace',
            fontSize: '0.8125rem',
            fontWeight: 700,
            letterSpacing: '0.05em',
          }}
        >
          {modules.filter(m => m.active).length} / {modules.length}
        </Box>
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
                <Box sx={{ position: 'relative', width: '100%', height: '100%', display: 'flex', flexDirection: 'column', p: 3 }}>
                  {/* Header avec icône et statut */}
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', mb: 4 }}>
                    <Box sx={{ color: '#4F8EF7' }}>
                      {module.icon}
                    </Box>
                    <Box
                      sx={{
                        px: 1.5,
                        py: 0.5,
                        backgroundColor: 'transparent',
                        color: '#81C784',
                        border: '1px solid #81C784',
                        fontFamily: '"JetBrains Mono", monospace',
                        fontSize: '0.6875rem',
                        fontWeight: 700,
                        letterSpacing: '0.05em',
                        textTransform: 'uppercase',
                      }}
                    >
                      ACTIF
                    </Box>
                  </Box>

                  {/* Nom du module */}
                  <Typography
                    variant="h4"
                    sx={{
                      mb: 3,
                      color: '#f0f0f0',
                      fontSize: '2rem',
                      fontWeight: 700,
                      letterSpacing: '-0.02em',
                    }}
                  >
                    {module.name}
                  </Typography>

                  {/* Description */}
                  {module.description && (
                    <Typography sx={{ mb: 4, color: '#a0a0a0', fontSize: '0.9375rem', lineHeight: 1.6 }}>
                      {module.description}
                    </Typography>
                  )}

                  {module.stats && (
                    <Box sx={{ display: 'flex', gap: 1.5, mb: 3 }}>
                      {module.stats.map((stat, idx) => (
                        <Box
                          key={idx}
                          sx={{
                            flex: 1,
                            textAlign: 'center',
                            py: 2,
                            px: 1,
                            borderRadius: '6px',
                            bgcolor: 'rgba(255,255,255,0.04)',
                            border: '1px solid rgba(255,255,255,0.07)',
                          }}
                        >
                          <Typography sx={{ fontFamily: '"JetBrains Mono", monospace', color: kuraColors.text0, fontWeight: 600, mb: 0.25, fontSize: '1.5rem', lineHeight: 1 }}>
                            {stat.value}
                          </Typography>
                          <Typography sx={{ textTransform: 'uppercase', letterSpacing: '0.08em', fontWeight: 500, color: kuraColors.text2, fontSize: '0.6875rem' }}>
                            {stat.label}
                          </Typography>
                        </Box>
                      ))}
                    </Box>
                  )}

                  {/* Features */}
                  {module.features && (
                    <Box sx={{ mt: 'auto', pt: 4, borderTop: '1px solid rgba(255, 255, 255, 0.08)' }}>
                      <Typography
                        sx={{
                          textTransform: 'uppercase',
                          letterSpacing: '0.15em',
                          mb: 3,
                          fontWeight: 700,
                          color: '#808080',
                          fontSize: '0.75rem',
                        }}
                      >
                        Fonctionnalités
                      </Typography>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                        {module.features.map((feature, idx) => (
                          <Box
                            key={idx}
                            sx={{
                              display: 'flex',
                              alignItems: 'center',
                              gap: 2,
                              py: 1,
                              px: 0,
                            }}
                          >
                            <Box sx={{ width: 3, height: 3, borderRadius: '50%', flexShrink: 0, bgcolor: kuraColors.border2 }} />
                            <Typography sx={{ color: '#f0f0f0', fontSize: '0.9375rem', fontWeight: 400 }}>
                              {feature}
                            </Typography>
                          </Box>
                        ))}
                      </Box>
                    </Box>
                  )}
                </Box>
              ) : (
                <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', justifyContent: 'center', alignItems: 'center', textAlign: 'center', p: 4 }}>
                  {/* Icône simple */}
                  <Box sx={{ mb: 4, color: '#A78BFA', opacity: 0.6 }}>
                    {module.icon}
                  </Box>

                  {/* Nom du module */}
                  <Typography
                    sx={{
                      mb: 2,
                      fontSize: '1.5rem',
                      fontWeight: 700,
                      color: '#f0f0f0',
                    }}
                  >
                    {module.name}
                  </Typography>

                  {/* Subtitle */}
                  {module.subtitle && (
                    <Box
                      sx={{
                        mb: 3,
                        px: 2,
                        py: 0.5,
                        border: '1px solid #A78BFA',
                        color: '#A78BFA',
                        fontSize: '0.75rem',
                        fontWeight: 700,
                        letterSpacing: '0.05em',
                        textTransform: 'uppercase',
                      }}
                    >
                      {module.subtitle}
                    </Box>
                  )}

                  {/* Description */}
                  {module.description && (
                    <Typography
                      sx={{
                        mb: 4,
                        maxWidth: '320px',
                        mx: 'auto',
                        color: '#a0a0a0',
                        fontSize: '0.875rem',
                        lineHeight: 1.6,
                      }}
                    >
                      {module.description}
                    </Typography>
                  )}

                  {/* Features */}
                  {module.features && (
                    <Box sx={{ width: '100%', mt: 'auto', pt: 3, borderTop: '1px solid rgba(255, 255, 255, 0.05)' }}>
                      <Typography
                        sx={{
                          textTransform: 'uppercase',
                          letterSpacing: '0.15em',
                          mb: 2,
                          fontWeight: 700,
                          color: '#606060',
                          fontSize: '0.6875rem',
                        }}
                      >
                        À venir
                      </Typography>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1.5 }}>
                        {module.features.map((feature, idx) => (
                          <Typography
                            key={idx}
                            sx={{
                              color: '#909090',
                              fontSize: '0.875rem',
                              fontWeight: 400,
                              textAlign: 'center',
                            }}
                          >
                            {feature}
                          </Typography>
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
