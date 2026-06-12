import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Box,
  Grid,
  CircularProgress,
  Alert,
  Chip,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Tabs,
  Tab,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  TextField,
  Collapse,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  ToggleButtonGroup,
  ToggleButton,
} from '@mui/material'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import {
  Folder as FolderIcon,
  FolderOpen as FolderOpenIcon,
  InsertDriveFile as FileIcon,
  Add as AddIcon,
  Remove as RemoveIcon,
  Edit as EditIcon,
  ChevronRight as ChevronRightIcon,
  ExpandMore as ExpandMoreIcon,
} from '@mui/icons-material'
import { kuraColors } from '../theme'
import { codeService, RepoTreeEntry, CommitSummary, CommitDetail } from '../services/codeService'
import { projectService } from '../services/projectService'
import { useProject } from '../contexts/ProjectContext'
import ModuleTitle from '../components/ModuleTitle'
import ModuleCard from '../components/ModuleCard'
import CodeBlock from '../components/CodeBlock'
import { ModuleSubtitle, ModuleSecondaryText, ModuleCaption } from '../components/ModuleText'

const LANGUAGE_BY_EXTENSION: Record<string, string> = {
  ts: 'typescript', tsx: 'tsx', js: 'javascript', jsx: 'jsx',
  go: 'go', py: 'python', rb: 'ruby', java: 'java', rs: 'rust',
  tf: 'hcl', tfvars: 'hcl', hcl: 'hcl',
  yaml: 'yaml', yml: 'yaml', json: 'json',
  md: 'markdown', sh: 'bash', bash: 'bash',
  sql: 'sql', html: 'markup', xml: 'markup', css: 'css', scss: 'scss',
  dockerfile: 'docker', toml: 'toml',
}

function languageForPath(path: string): string {
  const fileName = path.split('/').pop() ?? ''
  if (fileName.toLowerCase() === 'dockerfile') return 'docker'
  const ext = fileName.includes('.') ? fileName.split('.').pop()!.toLowerCase() : ''
  return LANGUAGE_BY_EXTENSION[ext] ?? 'plain-text'
}

function isMarkdownPath(path: string): boolean {
  return path.toLowerCase().endsWith('.md') || path.toLowerCase().endsWith('.markdown')
}

const markdownComponentStyles = {
  h1: { fontSize: '1.5rem', fontWeight: 700, color: kuraColors.text0, mt: 2, mb: 1.5, pb: 1, borderBottom: `1px solid ${kuraColors.border1}` },
  h2: { fontSize: '1.25rem', fontWeight: 700, color: kuraColors.text0, mt: 2, mb: 1, pb: 0.5, borderBottom: `1px solid ${kuraColors.border1}` },
  h3: { fontSize: '1.0625rem', fontWeight: 600, color: kuraColors.text0, mt: 1.5, mb: 1 },
  p: { fontSize: '0.875rem', color: kuraColors.text1, lineHeight: 1.7, mb: 1 },
  li: { fontSize: '0.875rem', color: kuraColors.text1, lineHeight: 1.7 },
  a: { color: kuraColors.accent },
  code: { fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8125rem', background: kuraColors.bg3, padding: '0.1em 0.35em', borderRadius: 4, color: kuraColors.text0 },
  table: { borderCollapse: 'collapse' as const, width: '100%', fontSize: '0.8125rem', color: kuraColors.text1, marginBottom: '1rem' },
  th: { border: `1px solid ${kuraColors.border1}`, padding: '0.4rem 0.6rem', textAlign: 'left' as const, color: kuraColors.text0 },
  td: { border: `1px solid ${kuraColors.border1}`, padding: '0.4rem 0.6rem' },
  blockquote: { borderLeft: `3px solid ${kuraColors.border2}`, margin: 0, paddingLeft: '1rem', color: kuraColors.text2 },
}

function MarkdownView({ content }: { content: string }) {
  return (
    <Box sx={{ '& > *:first-of-type': { mt: 0 } }}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={{
          h1: ({ children }) => <Box component="h1" sx={markdownComponentStyles.h1}>{children}</Box>,
          h2: ({ children }) => <Box component="h2" sx={markdownComponentStyles.h2}>{children}</Box>,
          h3: ({ children }) => <Box component="h3" sx={markdownComponentStyles.h3}>{children}</Box>,
          p: ({ children }) => <Box component="p" sx={markdownComponentStyles.p}>{children}</Box>,
          li: ({ children }) => <Box component="li" sx={markdownComponentStyles.li}>{children}</Box>,
          a: ({ children, href }) => <Box component="a" href={href} target="_blank" rel="noopener noreferrer" sx={markdownComponentStyles.a}>{children}</Box>,
          table: ({ children }) => <Box component="table" sx={markdownComponentStyles.table}>{children}</Box>,
          th: ({ children }) => <Box component="th" sx={markdownComponentStyles.th}>{children}</Box>,
          td: ({ children }) => <Box component="td" sx={markdownComponentStyles.td}>{children}</Box>,
          blockquote: ({ children }) => <Box component="blockquote" sx={markdownComponentStyles.blockquote}>{children}</Box>,
          pre: ({ children }) => <Box sx={{ mb: 1.5 }}>{children}</Box>,
          code: ({ className, children }) => {
            const match = /language-(\w+)/.exec(className || '')
            if (match) {
              return (
                <CodeBlock language={match[1]} showLineNumbers={false}>
                  {String(children).replace(/\n$/, '')}
                </CodeBlock>
              )
            }
            return <Box component="code" sx={markdownComponentStyles.code}>{children}</Box>
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </Box>
  )
}

function sortEntries(entries: RepoTreeEntry[]): RepoTreeEntry[] {
  return entries
    .slice()
    .sort((a, b) => (a.type === b.type ? a.name.localeCompare(b.name) : a.type === 'dir' ? -1 : 1))
}

interface TreeNodeProps {
  entry: RepoTreeEntry
  depth: number
  repo: string
  gitRef: string
  selectedFile: string | null
  onFileClick: (entry: RepoTreeEntry) => void
}

function TreeNode({ entry, depth, repo, gitRef, selectedFile, onFileClick }: TreeNodeProps) {
  const [open, setOpen] = useState(false)

  const { data: children, isLoading, error } = useQuery({
    queryKey: ['code-tree', repo, entry.path, gitRef],
    queryFn: () => codeService.getTree(repo, entry.path, gitRef),
    enabled: entry.type === 'dir' && open,
  })

  const handleClick = () => {
    if (entry.type === 'dir') {
      setOpen((prev) => !prev)
    } else {
      onFileClick(entry)
    }
  }

  return (
    <>
      <ListItemButton
        selected={selectedFile === entry.path}
        onClick={handleClick}
        sx={{ borderRadius: 1, pl: 1 + depth * 2 }}
      >
        <ListItemIcon sx={{ minWidth: 24 }}>
          {entry.type === 'dir' ? (
            open ? <ExpandMoreIcon fontSize="small" sx={{ color: kuraColors.text2 }} /> : <ChevronRightIcon fontSize="small" sx={{ color: kuraColors.text2 }} />
          ) : null}
        </ListItemIcon>
        <ListItemIcon sx={{ minWidth: 28 }}>
          {entry.type === 'dir'
            ? (open ? <FolderOpenIcon fontSize="small" sx={{ color: kuraColors.warning }} /> : <FolderIcon fontSize="small" sx={{ color: kuraColors.warning }} />)
            : <FileIcon fontSize="small" sx={{ color: kuraColors.text2 }} />}
        </ListItemIcon>
        <ListItemText
          primary={entry.name}
          primaryTypographyProps={{ sx: { fontSize: '0.8125rem', fontFamily: '"JetBrains Mono", monospace' } }}
        />
      </ListItemButton>
      {entry.type === 'dir' && (
        <Collapse in={open} timeout="auto" unmountOnExit>
          {isLoading ? (
            <Box sx={{ pl: 1 + (depth + 1) * 2, py: 0.5 }}>
              <CircularProgress size={16} />
            </Box>
          ) : error ? (
            <Box sx={{ pl: 1 + (depth + 1) * 2, py: 0.5 }}>
              <ModuleCaption sx={{ color: kuraColors.error }}>Erreur de chargement</ModuleCaption>
            </Box>
          ) : (
            <List dense disablePadding>
              {sortEntries(children ?? []).map((child) => (
                <TreeNode
                  key={child.path}
                  entry={child}
                  depth={depth + 1}
                  repo={repo}
                  gitRef={gitRef}
                  selectedFile={selectedFile}
                  onFileClick={onFileClick}
                />
              ))}
            </List>
          )}
        </Collapse>
      )}
    </>
  )
}

function statusColor(status: string): string {
  switch (status) {
    case 'added': return kuraColors.success
    case 'removed': return kuraColors.error
    case 'modified':
    case 'changed': return kuraColors.warning
    default: return kuraColors.text2
  }
}

export default function CodePage() {
  const { currentProject } = useProject()
  const [activeTab, setActiveTab] = useState(0)
  const [selectedRepo, setSelectedRepo] = useState<string>('')
  const [ref, setRef] = useState('main')
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [selectedCommit, setSelectedCommit] = useState<CommitSummary | null>(null)
  const [viewMode, setViewMode] = useState<'preview' | 'source'>('preview')

  const { data: mappingsData, isLoading: mappingsLoading } = useQuery({
    queryKey: ['project-mappings', currentProject?.id],
    queryFn: () => projectService.listMappings(currentProject!.id),
    enabled: !!currentProject,
  })

  const repos = (mappingsData?.items ?? []).filter((m) => !!m.github_repository)

  // Sélectionne automatiquement le premier dépôt disponible
  if (!selectedRepo && repos.length > 0) {
    setSelectedRepo(repos[0].github_repository!)
  }

  const { data: treeEntries, isLoading: treeLoading, error: treeError } = useQuery({
    queryKey: ['code-tree', selectedRepo, '', ref],
    queryFn: () => codeService.getTree(selectedRepo, '', ref),
    enabled: !!selectedRepo && activeTab === 0,
  })

  const { data: fileContent, isLoading: fileLoading, error: fileError } = useQuery({
    queryKey: ['code-file', selectedRepo, selectedFile, ref],
    queryFn: () => codeService.getFile(selectedRepo, selectedFile!, ref),
    enabled: !!selectedRepo && !!selectedFile && activeTab === 0,
  })

  const { data: commits, isLoading: commitsLoading, error: commitsError } = useQuery({
    queryKey: ['code-commits', selectedRepo, ref],
    queryFn: () => codeService.getCommits(selectedRepo, '', ref),
    enabled: !!selectedRepo && activeTab === 1,
  })

  const { data: commitDetail, isLoading: commitDetailLoading } = useQuery({
    queryKey: ['code-commit-detail', selectedRepo, selectedCommit?.sha],
    queryFn: () => codeService.getCommitDetail(selectedRepo, selectedCommit!.sha),
    enabled: !!selectedRepo && !!selectedCommit,
  })

  const handleRepoChange = (repo: string) => {
    setSelectedRepo(repo)
    setSelectedFile(null)
  }

  const handleEntryClick = (entry: RepoTreeEntry) => {
    setSelectedFile(entry.path)
    setViewMode('preview')
  }

  return (
    <Box sx={{ p: 3 }}>
      <ModuleTitle>Repository</ModuleTitle>

      {mappingsLoading ? (
        <CircularProgress size={24} />
      ) : repos.length === 0 ? (
        <Alert severity="info">
          Aucun dépôt GitHub n'est lié à ce projet. Ajoutez un mapping de dépôt GitHub depuis la page Projets pour pouvoir parcourir son code ici.
        </Alert>
      ) : (
        <>
          <Box sx={{ display: 'flex', gap: 2, mb: 3, flexWrap: 'wrap', alignItems: 'center' }}>
            <FormControl size="small" sx={{ minWidth: 240 }}>
              <InputLabel id="repo-select-label">Dépôt</InputLabel>
              <Select
                labelId="repo-select-label"
                label="Dépôt"
                value={selectedRepo}
                onChange={(e) => handleRepoChange(e.target.value)}
              >
                {repos.map((m) => (
                  <MenuItem key={m.id} value={m.github_repository}>
                    {m.github_repository}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              size="small"
              label="Branche / ref"
              value={ref}
              onChange={(e) => setRef(e.target.value)}
              sx={{ width: 180 }}
            />
          </Box>

          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 2 }}>
            <Tab label="Fichiers" />
            <Tab label="Historique" />
          </Tabs>

          {activeTab === 0 && (
            <Grid container spacing={2}>
              <Grid item xs={12} md={4}>
                <ModuleCard sx={{ p: 2, minHeight: 400 }}>
                  <ModuleSecondaryText sx={{ mb: 1.5, fontFamily: '"JetBrains Mono", monospace' }}>
                    {selectedRepo}
                  </ModuleSecondaryText>

                  {treeLoading ? (
                    <CircularProgress size={20} />
                  ) : treeError ? (
                    <Alert severity="error">Impossible de charger l'arborescence : {(treeError as Error).message}</Alert>
                  ) : (
                    <List dense disablePadding>
                      {sortEntries(treeEntries ?? []).map((entry) => (
                        <TreeNode
                          key={entry.path}
                          entry={entry}
                          depth={0}
                          repo={selectedRepo}
                          gitRef={ref}
                          selectedFile={selectedFile}
                          onFileClick={handleEntryClick}
                        />
                      ))}
                    </List>
                  )}
                </ModuleCard>
              </Grid>

              <Grid item xs={12} md={8}>
                <ModuleCard sx={{ p: 2, minHeight: 400 }}>
                  {!selectedFile ? (
                    <ModuleSecondaryText sx={{ color: kuraColors.text2 }}>
                      Sélectionnez un fichier dans l'arborescence pour afficher son contenu.
                    </ModuleSecondaryText>
                  ) : fileLoading ? (
                    <CircularProgress size={20} />
                  ) : fileError ? (
                    <Alert severity="error">Impossible de charger le fichier : {(fileError as Error).message}</Alert>
                  ) : fileContent?.truncated ? (
                    <Alert severity="warning">
                      Fichier trop volumineux pour être affiché ({fileContent.size} octets).
                    </Alert>
                  ) : isMarkdownPath(fileContent!.path) ? (
                    <Box>
                      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 1 }}>
                        <ToggleButtonGroup
                          size="small"
                          value={viewMode}
                          exclusive
                          onChange={(_, v) => { if (v) setViewMode(v) }}
                        >
                          <ToggleButton value="preview" sx={{ fontSize: '0.75rem', px: 1.5 }}>Aperçu</ToggleButton>
                          <ToggleButton value="source" sx={{ fontSize: '0.75rem', px: 1.5 }}>Source</ToggleButton>
                        </ToggleButtonGroup>
                      </Box>
                      {viewMode === 'preview' ? (
                        <MarkdownView content={fileContent!.content} />
                      ) : (
                        <CodeBlock language="markdown" label={fileContent!.path}>
                          {fileContent!.content}
                        </CodeBlock>
                      )}
                    </Box>
                  ) : (
                    <CodeBlock language={languageForPath(fileContent!.path)} label={fileContent!.path}>
                      {fileContent!.content}
                    </CodeBlock>
                  )}
                </ModuleCard>
              </Grid>
            </Grid>
          )}

          {activeTab === 1 && (
            <ModuleCard sx={{ p: 2 }}>
              {commitsLoading ? (
                <CircularProgress size={20} />
              ) : commitsError ? (
                <Alert severity="error">Impossible de charger l'historique : {(commitsError as Error).message}</Alert>
              ) : (
                <List dense disablePadding>
                  {(commits ?? []).map((commit) => (
                    <ListItemButton
                      key={commit.sha}
                      onClick={() => setSelectedCommit(commit)}
                      sx={{ borderRadius: 1, alignItems: 'flex-start', py: 1 }}
                    >
                      <ListItemIcon sx={{ minWidth: 32, mt: 0.5 }}>
                        <EditIcon fontSize="small" sx={{ color: kuraColors.text2 }} />
                      </ListItemIcon>
                      <ListItemText
                        primary={commit.message.split('\n')[0]}
                        secondary={
                          <>
                            <Chip
                              label={commit.sha.slice(0, 7)}
                              size="small"
                              sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.6875rem', mr: 1, height: 20 }}
                            />
                            {commit.author} · {new Date(commit.date).toLocaleString('fr-FR')}
                          </>
                        }
                        primaryTypographyProps={{ sx: { fontSize: '0.875rem' } }}
                        secondaryTypographyProps={{ component: 'div', sx: { mt: 0.5 } }}
                      />
                    </ListItemButton>
                  ))}
                </List>
              )}
            </ModuleCard>
          )}
        </>
      )}

      <Dialog open={!!selectedCommit} onClose={() => setSelectedCommit(null)} maxWidth="md" fullWidth>
        <DialogTitle>{selectedCommit?.message.split('\n')[0]}</DialogTitle>
        <DialogContent>
          {commitDetailLoading ? (
            <CircularProgress size={20} />
          ) : (
            <CommitDetailView detail={commitDetail} />
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedCommit(null)}>Fermer</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}

function CommitDetailView({ detail }: { detail?: CommitDetail }) {
  if (!detail) return null
  return (
    <Box>
      <ModuleSubtitle sx={{ mb: 1 }}>{detail.files.length} fichier(s) modifié(s)</ModuleSubtitle>
      {detail.files.map((file) => (
        <Box key={file.filename} sx={{ mb: 2 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 0.5 }}>
            {file.status === 'added' && <AddIcon fontSize="small" sx={{ color: statusColor(file.status) }} />}
            {file.status === 'removed' && <RemoveIcon fontSize="small" sx={{ color: statusColor(file.status) }} />}
            {file.status !== 'added' && file.status !== 'removed' && <EditIcon fontSize="small" sx={{ color: statusColor(file.status) }} />}
            <ModuleSecondaryText sx={{ fontFamily: '"JetBrains Mono", monospace', fontSize: '0.8125rem' }}>
              {file.filename}
            </ModuleSecondaryText>
            <ModuleCaption sx={{ color: kuraColors.success }}>+{file.additions}</ModuleCaption>
            <ModuleCaption sx={{ color: kuraColors.error }}>-{file.deletions}</ModuleCaption>
          </Box>
          {file.patch ? (
            <CodeBlock language="diff" showLineNumbers={false}>{file.patch}</CodeBlock>
          ) : (
            <ModuleCaption>Aperçu du diff non disponible (fichier binaire ou trop volumineux).</ModuleCaption>
          )}
        </Box>
      ))}
    </Box>
  )
}
