import { useQuery } from '@tanstack/react-query'
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'
import { registryService } from '../services/registryService'
import { ModuleSubtitle } from './ModuleText'
import { kuraColors } from '../theme'

interface ImageDetailDialogProps {
  repository: string
  tag: string | null
  onClose: () => void
}

function formatSize(bytes: number): string {
  if (!bytes) return '—'
  const mo = bytes / 1024 / 1024
  return mo >= 1 ? `${mo.toFixed(1)} Mo` : `${(bytes / 1024).toFixed(0)} Ko`
}

/**
 * Contenu d'une image du registre.
 *
 * Un registre OCI ne conserve pas les sources : le Dockerfile n'y figure pas et
 * ne peut pas en être extrait. Ce que l'on peut restituer, et qui sert
 * précisément à auditer une image, c'est sa configuration d'exécution et
 * l'historique des instructions ayant produit chaque couche.
 */
export default function ImageDetailDialog({ repository, tag, onClose }: ImageDetailDialogProps) {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['registry-image', repository, tag],
    queryFn: () => registryService.getImage(repository, tag as string),
    enabled: !!tag,
    retry: false,
    // Une image est immuable à digest donné : rien ne justifie de la
    // réinterroger périodiquement, d'autant que chaque appel ouvre un
    // port-forward vers le registre.
    refetchInterval: false,
  })

  const mono = { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }

  return (
    <Dialog open={!!tag} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle sx={{ ...mono, fontSize: '0.95rem' }}>
        {repository}:{tag}
      </DialogTitle>
      <DialogContent dividers>
        {isLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
            <CircularProgress />
          </Box>
        ) : isError ? (
          <Alert severity="error">
            Lecture de l&apos;image impossible : {(error as any)?.response?.data?.error || (error as any)?.message}
          </Alert>
        ) : data ? (
          <>
            <Alert severity="info" sx={{ mb: 2 }}>
              Un registre ne stocke pas les sources : le Dockerfile n&apos;y figure pas. Sont
              présentées ci-dessous la configuration d&apos;exécution de l&apos;image et les
              instructions ayant produit chacune de ses couches.
            </Alert>

            <ModuleSubtitle>Identité</ModuleSubtitle>
            <Table size="small" sx={{ mb: 3 }}>
              <TableBody>
                <TableRow>
                  <TableCell sx={{ width: 180, color: kuraColors.text2 }}>Empreinte</TableCell>
                  <TableCell sx={mono}>{data.digest}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ color: kuraColors.text2 }}>Plateforme</TableCell>
                  <TableCell sx={mono}>
                    {data.os}/{data.architecture}
                  </TableCell>
                </TableRow>
                {data.created_at && (
                  <TableRow>
                    <TableCell sx={{ color: kuraColors.text2 }}>Construite le</TableCell>
                    <TableCell sx={mono}>{new Date(data.created_at).toLocaleString('fr-FR')}</TableCell>
                  </TableRow>
                )}
                <TableRow>
                  <TableCell sx={{ color: kuraColors.text2 }}>Taille des couches</TableCell>
                  <TableCell sx={mono}>{formatSize(data.total_bytes)}</TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <ModuleSubtitle>Exécution</ModuleSubtitle>
            <Table size="small" sx={{ mb: 3 }}>
              <TableBody>
                {data.entrypoint?.length ? (
                  <TableRow>
                    <TableCell sx={{ width: 180, color: kuraColors.text2 }}>Point d&apos;entrée</TableCell>
                    <TableCell sx={mono}>{data.entrypoint.join(' ')}</TableCell>
                  </TableRow>
                ) : null}
                {data.cmd?.length ? (
                  <TableRow>
                    <TableCell sx={{ color: kuraColors.text2 }}>Commande</TableCell>
                    <TableCell sx={mono}>{data.cmd.join(' ')}</TableCell>
                  </TableRow>
                ) : null}
                {data.working_dir ? (
                  <TableRow>
                    <TableCell sx={{ color: kuraColors.text2 }}>Répertoire</TableCell>
                    <TableCell sx={mono}>{data.working_dir}</TableCell>
                  </TableRow>
                ) : null}
                {data.exposed_ports?.length ? (
                  <TableRow>
                    <TableCell sx={{ color: kuraColors.text2 }}>Ports exposés</TableCell>
                    <TableCell>
                      {data.exposed_ports.map((p) => (
                        <Chip key={p} label={p} size="small" variant="outlined" sx={{ mr: 0.5, ...mono }} />
                      ))}
                    </TableCell>
                  </TableRow>
                ) : null}
              </TableBody>
            </Table>

            {data.labels && Object.keys(data.labels).length > 0 && (
              <>
                <ModuleSubtitle>Étiquettes</ModuleSubtitle>
                <Table size="small" sx={{ mb: 3 }}>
                  <TableBody>
                    {Object.entries(data.labels).map(([key, value]) => (
                      <TableRow key={key}>
                        <TableCell sx={{ width: 280, ...mono, color: kuraColors.text2 }}>{key}</TableCell>
                        <TableCell sx={mono}>{value}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </>
            )}

            <ModuleSubtitle>Construction</ModuleSubtitle>
            <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, mb: 1 }}>
              Les étapes sans taille (ENV, LABEL, EXPOSE…) ne produisent aucune couche.
            </Typography>
            <Box sx={{ overflowX: 'auto' }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell sx={{ width: 40 }}>#</TableCell>
                    <TableCell>Instruction</TableCell>
                    <TableCell align="right" sx={{ width: 110 }}>
                      Taille
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {data.layers?.map((layer, index) => (
                    <TableRow key={index}>
                      <TableCell sx={{ color: kuraColors.text2 }}>{index + 1}</TableCell>
                      <TableCell
                        sx={{
                          ...mono,
                          color: layer.empty ? kuraColors.text2 : kuraColors.text0,
                          whiteSpace: 'pre-wrap',
                          wordBreak: 'break-word',
                        }}
                      >
                        {layer.instruction || '—'}
                      </TableCell>
                      <TableCell align="right" sx={mono}>
                        {layer.empty ? '—' : formatSize(layer.size_bytes)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Box>
          </>
        ) : null}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Fermer</Button>
      </DialogActions>
    </Dialog>
  )
}
