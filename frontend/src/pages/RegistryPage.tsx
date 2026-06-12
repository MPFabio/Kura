import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Box,
  Alert,
  CircularProgress,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
} from '@mui/material'
import {
  Refresh as RefreshIcon,
  VerifiedUser as VerifiedUserIcon,
  GppMaybe as UnsignedIcon,
} from '@mui/icons-material'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import { ModuleSubtitle, ModuleSecondaryText } from '../components/ModuleText'
import { registryService, RegistryRepository } from '../services/registryService'
import { kuraColors } from '../theme'

function getRegistryErrorMessage(error: unknown): { severity: 'info' | 'error'; message: string } {
  const detail = (error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? ''

  if (detail.includes('aucun pod zot')) {
    return {
      severity: 'info',
      message:
        'Zot n\'est pas encore déployé dans le cluster actif. Déployez-le (Application ArgoCD, namespace "zot", label "app=zot") pour activer ce module.',
    }
  }

  if (detail.includes('aucun cluster actif')) {
    return {
      severity: 'info',
      message: 'Aucun cluster actif. Sélectionnez un cluster pour accéder au registre.',
    }
  }

  return { severity: 'error', message: 'Impossible de contacter le registre.' }
}

function formatSize(bytes: number): string {
  if (!bytes) return '—'
  const units = ['o', 'Ko', 'Mo', 'Go']
  let value = bytes
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(1)} ${units[unitIndex]}`
}

export default function RegistryPage() {
  const [selectedRepo, setSelectedRepo] = useState<RegistryRepository | null>(null)

  const { data: repositories, isLoading, isError, error, refetch, isFetching } = useQuery({
    queryKey: ['registry-repositories'],
    queryFn: () => registryService.listRepositories(),
    retry: false,
  })

  const { data: repoDetail, isLoading: detailLoading } = useQuery({
    queryKey: ['registry-repository', selectedRepo?.name],
    queryFn: () => registryService.getRepository(selectedRepo!.name),
    enabled: !!selectedRepo,
    retry: false,
  })

  return (
    <Box>
      <ModuleTitle>Zot</ModuleTitle>

      <ModuleCard>
        <Box sx={{ p: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Box>
              <ModuleSubtitle>Dépôts (Zot)</ModuleSubtitle>
              <ModuleSecondaryText>
                Registre OCI privé pour vos images et charts Helm, avec signature Cosign.
              </ModuleSecondaryText>
            </Box>
            <Tooltip title="Rafraîchir">
              <IconButton onClick={() => refetch()} color="primary">
                {isFetching ? <CircularProgress size={20} /> : <RefreshIcon />}
              </IconButton>
            </Tooltip>
          </Box>

          {isLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : isError ? (
            (() => {
              const { severity, message } = getRegistryErrorMessage(error)
              return <Alert severity={severity}>{message}</Alert>
            })()
          ) : (repositories?.length ?? 0) === 0 ? (
            <Alert severity="info">Aucun dépôt n&apos;a encore été poussé vers le registre.</Alert>
          ) : (
            <TableContainer component={Paper} variant="outlined" sx={{ bgcolor: 'transparent' }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Dépôt</TableCell>
                    <TableCell align="right">Tags</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {repositories!.map((repo) => (
                    <TableRow
                      key={repo.name}
                      hover
                      sx={{ cursor: 'pointer' }}
                      onClick={() => setSelectedRepo(repo)}
                    >
                      <TableCell sx={{ fontFamily: '"JetBrains Mono", monospace' }}>{repo.name}</TableCell>
                      <TableCell align="right">{repo.tag_count}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      </ModuleCard>

      {/* Détail d'un dépôt : tags + statut Cosign */}
      <Dialog open={!!selectedRepo} onClose={() => setSelectedRepo(null)} maxWidth="md" fullWidth>
        <DialogTitle sx={{ fontFamily: '"JetBrains Mono", monospace' }}>{selectedRepo?.name}</DialogTitle>
        <DialogContent>
          {detailLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
              <CircularProgress />
            </Box>
          ) : (repoDetail?.tags?.length ?? 0) === 0 ? (
            <Alert severity="info">Aucun tag trouvé pour ce dépôt.</Alert>
          ) : (
            <TableContainer component={Paper} variant="outlined" sx={{ bgcolor: 'transparent' }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Tag</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell align="right">Taille</TableCell>
                    <TableCell align="right">Signature</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {repoDetail!.tags.map((tag) => (
                    <TableRow key={tag.name}>
                      <TableCell sx={{ fontFamily: '"JetBrains Mono", monospace' }}>{tag.name}</TableCell>
                      <TableCell>
                        <Chip
                          label={tag.type === 'helm-chart' ? 'Chart Helm' : 'Image'}
                          size="small"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell align="right">{formatSize(tag.size_bytes)}</TableCell>
                      <TableCell align="right">
                        {tag.signed ? (
                          <Chip
                            icon={<VerifiedUserIcon sx={{ fontSize: 16 }} />}
                            label="Signé (Cosign)"
                            size="small"
                            sx={{ color: kuraColors.success, borderColor: kuraColors.success }}
                            variant="outlined"
                          />
                        ) : (
                          <Chip
                            icon={<UnsignedIcon sx={{ fontSize: 16 }} />}
                            label="Non signé"
                            size="small"
                            sx={{ color: kuraColors.text2, borderColor: kuraColors.border1 }}
                            variant="outlined"
                          />
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedRepo(null)}>Fermer</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
