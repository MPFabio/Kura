import {
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
} from '@mui/material'
import { TerraformDriftResult } from '../services/terraformService'
import { kuraColors } from '../theme'

interface DriftDetailDialogProps {
  result: TerraformDriftResult | null
  onClose: () => void
}

/**
 * Met en forme une valeur d'attribut pour la lecture.
 *
 * Les valeurs viennent d'un plan OpenTofu : ce sont aussi bien des chaînes que
 * des objets ou des listes. Les objets sont indentés pour rester lisibles,
 * tandis qu'une chaîne est affichée telle quelle — l'entourer de guillemets
 * JSON n'apporterait rien et allongerait des valeurs déjà longues, comme les
 * clés SSH.
 */
function formatValue(value: unknown): string {
  if (value === null || value === undefined) return '(absent)'
  if (typeof value === 'string') return value
  return JSON.stringify(value, null, 2)
}

/**
 * Détail d'une dérive, présenté comme un diff.
 *
 * Le tableau récapitulatif tronquait les valeurs à trois différences et les
 * affichait sur une ligne : une clé SSH ou une empreinte y devenait illisible,
 * et l'on ne pouvait pas comparer l'attendu au constaté. Ici chaque attribut
 * est présenté sur deux blocs, dans l'ordre attendu puis constaté, selon la
 * convention des outils d'infrastructure : ce qui est déclaré en premier, ce
 * qui existe réellement ensuite.
 */
export default function DriftDetailDialog({ result, onClose }: DriftDetailDialogProps) {
  const mono = { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.75rem' }
  const differences = result?.differences ?? []

  return (
    <Dialog open={!!result} onClose={onClose} maxWidth="lg" fullWidth>
      <DialogTitle sx={{ pb: 1 }}>
        <Typography component="div" sx={{ ...mono, fontSize: '0.95rem', fontWeight: 600 }}>
          {result?.resource_address}
        </Typography>
        <Box sx={{ display: 'flex', gap: 1, mt: 1, alignItems: 'center' }}>
          <Chip
            label={result?.status === 'missing' ? 'Absente de l’infrastructure' : 'Dérive détectée'}
            size="small"
            sx={{ color: kuraColors.warning, borderColor: kuraColors.warning }}
            variant="outlined"
          />
          <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2 }}>
            {result?.resource_type}
          </Typography>
        </Box>
      </DialogTitle>

      <DialogContent dividers>
        {differences.length === 0 ? (
          <Typography sx={{ color: kuraColors.text2 }}>
            {result?.message || 'Aucune différence d’attribut n’a été relevée.'}
          </Typography>
        ) : (
          <>
            <Typography sx={{ fontSize: '0.8125rem', color: kuraColors.text2, mb: 2 }}>
              {differences.length} attribut(s) diffèrent entre l&apos;état enregistré et
              l&apos;infrastructure réelle.
            </Typography>

            {differences.map((diff, index) => (
              <Box key={index} sx={{ mb: 2.5 }}>
                <Typography sx={{ ...mono, color: kuraColors.text0, fontWeight: 600, mb: 0.75 }}>
                  {diff.attribute}
                </Typography>

                <Box
                  sx={{
                    borderRadius: '6px',
                    overflow: 'hidden',
                    border: `1px solid ${kuraColors.border0}`,
                  }}
                >
                  <Box
                    sx={{
                      display: 'flex',
                      gap: 1,
                      p: 1.25,
                      bgcolor: 'rgba(239, 68, 68, 0.08)',
                      borderLeft: `3px solid ${kuraColors.error}`,
                    }}
                  >
                    <Typography sx={{ ...mono, color: kuraColors.error, userSelect: 'none' }}>−</Typography>
                    <Box sx={{ minWidth: 0 }}>
                      <Typography sx={{ fontSize: '0.6875rem', color: kuraColors.text2, mb: 0.25 }}>
                        Attendu (état enregistré)
                      </Typography>
                      <Typography
                        component="pre"
                        sx={{ ...mono, color: kuraColors.text0, m: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}
                      >
                        {formatValue(diff.expected)}
                      </Typography>
                    </Box>
                  </Box>

                  <Box
                    sx={{
                      display: 'flex',
                      gap: 1,
                      p: 1.25,
                      bgcolor: 'rgba(34, 197, 94, 0.08)',
                      borderLeft: `3px solid ${kuraColors.success}`,
                    }}
                  >
                    <Typography sx={{ ...mono, color: kuraColors.success, userSelect: 'none' }}>+</Typography>
                    <Box sx={{ minWidth: 0 }}>
                      <Typography sx={{ fontSize: '0.6875rem', color: kuraColors.text2, mb: 0.25 }}>
                        Constaté (infrastructure réelle)
                      </Typography>
                      <Typography
                        component="pre"
                        sx={{ ...mono, color: kuraColors.text0, m: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}
                      >
                        {formatValue(diff.actual)}
                      </Typography>
                    </Box>
                  </Box>
                </Box>
              </Box>
            ))}

            <Typography sx={{ fontSize: '0.75rem', color: kuraColors.text2, mt: 2 }}>
              Une dérive signale que l&apos;infrastructure a été modifiée hors du code. Corriger
              consiste soit à reporter le changement dans le dépôt, soit à réappliquer le code pour
              rétablir l&apos;état déclaré.
            </Typography>
          </>
        )}
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose}>Fermer</Button>
      </DialogActions>
    </Dialog>
  )
}
