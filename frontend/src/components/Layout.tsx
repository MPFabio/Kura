import { useState } from 'react'
import React from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Box,
  Drawer,
  List,
  Typography,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Avatar,
  Menu,
  MenuItem,
  Select,
  Divider,
  Tooltip,
} from '@mui/material'
import {
  Logout,
  Folder as FolderIcon,
  KeyboardArrowDown,
  ChevronRight,
} from '@mui/icons-material'
import { useAuth } from '../contexts/AuthContext'
import { useProject } from '../contexts/ProjectContext'
import { useSocket } from '../contexts/SocketContext'
import { kuraColors } from '../theme'
import Logo from './Logo'
import AnimatedBackground from './AnimatedBackground'
import TerraformIcon from './icons/TerraformIcon'
import KubernetesIcon from './icons/KubernetesIcon'
import AnsibleIcon from './icons/AnsibleIcon'
import ModulesIcon from './icons/ModulesIcon'
import MonitoringIcon from './icons/MonitoringIcon'
import PipelinesIcon from './icons/PipelinesIcon'
import AlertsIcon from './icons/AlertsIcon'
import SettingsIcon from './icons/SettingsIcon'
import { MenuBook as MenuBookIcon } from '@mui/icons-material'

const DRAWER_WIDTH = 220

const navItems = [
  { text: 'Modules',       icon: <ModulesIcon />,    path: '/modules',        custom: true },
  { text: 'Terraform',     icon: <TerraformIcon />,  path: '/terraform',      custom: true },
  { text: 'Kubernetes',    icon: <KubernetesIcon />, path: '/k8s',            custom: true },
  { text: 'Ansible',       icon: <AnsibleIcon />,    path: '/ansible',        custom: true },
  { text: 'Monitoring',    icon: <MonitoringIcon />, path: '/metrics',        custom: true },
  { text: 'Pipelines',     icon: <PipelinesIcon />,  path: '/pipelines',      custom: true },
  { text: 'Alertes',       icon: <AlertsIcon />,     path: '/alerts',         custom: true },
]

const bottomItems = [
  { text: 'Documentation', icon: <MenuBookIcon sx={{ fontSize: 18 }} />, path: '/documentation', custom: false },
  { text: 'Paramètres',    icon: <SettingsIcon />,   path: '/settings',       custom: true },
]

export default function Layout() {
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()
  const { currentProject, projects, setCurrentProject } = useProject()
  const { connected } = useSocket()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
    setAnchorEl(null)
  }

  const isActive = (path: string) =>
    location.pathname === path ||
    (path === '/k8s' && location.pathname.startsWith('/k8s')) ||
    (path === '/documentation' && location.pathname.startsWith('/documentation'))

  const NavItem = ({ item }: { item: typeof navItems[0] }) => {
    const active = isActive(item.path)
    return (
      <ListItem disablePadding sx={{ mb: 0.25 }}>
        <ListItemButton
          selected={active}
          onClick={() => navigate(item.path)}
          sx={{
            borderRadius: '6px',
            mx: 1,
            px: 1.5,
            py: 0.875,
            minHeight: 36,
            '&.Mui-selected': {
              bgcolor: kuraColors.accentMuted,
              '&:hover': { bgcolor: kuraColors.accentMuted },
            },
          }}
        >
          <ListItemIcon sx={{ minWidth: 32 }}>
            {item.custom
              ? React.cloneElement(item.icon as React.ReactElement, { active })
              : React.cloneElement(item.icon as React.ReactElement, {
                  sx: { fontSize: 18, color: active ? kuraColors.accent : kuraColors.text2 },
                })}
          </ListItemIcon>
          <ListItemText
            primary={item.text}
            primaryTypographyProps={{
              sx: {
                fontSize: '0.875rem',
                fontWeight: active ? 500 : 400,
                color: active ? kuraColors.text0 : kuraColors.text1,
              },
            }}
          />
          {active && <ChevronRight sx={{ fontSize: 14, color: kuraColors.accent, opacity: 0.6 }} />}
        </ListItemButton>
      </ListItem>
    )
  }

  const sidebar = (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100%', bgcolor: kuraColors.bg1 }}>
      {/* Logo */}
      <Box sx={{
        height: 72,
        borderBottom: `1px solid ${kuraColors.border1}`,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'flex-start',
        pl: 3,
        bgcolor: kuraColors.bg1,
      }}>
        <Logo variant="icon" size="small" />
      </Box>

      {/* Project selector */}
      <Box sx={{ px: 2, py: 1.5, borderBottom: `1px solid ${kuraColors.border0}` }}>
        <Typography sx={{ fontSize: '0.65rem', fontWeight: 600, color: kuraColors.text2, textTransform: 'uppercase', letterSpacing: '0.08em', mb: 0.75 }}>
          Projet
        </Typography>
        <Select
          value={currentProject?.id || ''}
          onChange={(e) => {
            const project = projects.find(p => p.id === e.target.value)
            if (project) {
              setCurrentProject(project)
              if (location.pathname === '/projects') navigate('/modules')
            }
          }}
          displayEmpty
          size="small"
          fullWidth
          IconComponent={KeyboardArrowDown}
          MenuProps={{
            PaperProps: {
              sx: {
                bgcolor: kuraColors.bg2,
                border: `1px solid ${kuraColors.border2}`,
                borderRadius: '8px',
                boxShadow: '0 8px 24px rgba(0,0,0,0.4)',
                mt: 0.5,
              },
            },
          }}
          sx={{
            fontSize: '0.875rem',
            fontWeight: 500,
            color: kuraColors.text0,
            bgcolor: kuraColors.bg2,
            borderRadius: '6px',
            '& .MuiOutlinedInput-notchedOutline': { borderColor: kuraColors.border1 },
            '&:hover .MuiOutlinedInput-notchedOutline': { borderColor: kuraColors.border2 },
            '&.Mui-focused .MuiOutlinedInput-notchedOutline': { borderColor: kuraColors.accent, borderWidth: 1 },
            '& .MuiSelect-icon': { color: kuraColors.text2, fontSize: 18 },
            '& .MuiSelect-select': { py: 1, px: 1.5 },
          }}
          renderValue={(selected) => {
            if (!selected) return <Typography sx={{ fontSize: '0.875rem', color: kuraColors.text2 }}>Sélectionner...</Typography>
            const project = projects.find(p => p.id === selected)
            return (
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <FolderIcon sx={{ fontSize: 14, color: kuraColors.accent }} />
                <Typography sx={{ fontSize: '0.875rem', fontWeight: 500, color: kuraColors.text0 }}>{project?.name}</Typography>
              </Box>
            )
          }}
        >
          <MenuItem value="" onClick={() => navigate('/projects')} sx={{ fontSize: '0.875rem', color: kuraColors.text1 }}>
            <FolderIcon sx={{ fontSize: 14, mr: 1.5 }} /> Gérer les projets
          </MenuItem>
          <Divider sx={{ my: 0.5, borderColor: kuraColors.border0 }} />
          {projects.map((p) => (
            <MenuItem key={p.id} value={p.id} sx={{ fontSize: '0.875rem', justifyContent: 'space-between' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <FolderIcon sx={{ fontSize: 14, color: currentProject?.id === p.id ? kuraColors.accent : kuraColors.text2 }} />
                {p.name}
              </Box>
              {currentProject?.id === p.id && (
                <Box sx={{ width: 6, height: 6, borderRadius: '50%', bgcolor: kuraColors.success, ml: 1 }} />
              )}
            </MenuItem>
          ))}
        </Select>
      </Box>

      {/* Navigation principale */}
      <List sx={{ flex: 1, py: 1, overflowY: 'auto' }}>
        {navItems.map((item) => <NavItem key={item.path} item={item} />)}
      </List>

      {/* Navigation secondaire + user */}
      <Box sx={{ borderTop: `1px solid ${kuraColors.border0}`, pb: 1 }}>
        <List sx={{ py: 1 }}>
          {bottomItems.map((item) => <NavItem key={item.path} item={item} />)}
        </List>

        {/* User */}
        <Box
          onClick={(e) => setAnchorEl(e.currentTarget)}
          sx={{
            mx: 1,
            px: 1.5,
            py: 1,
            borderRadius: '6px',
            display: 'flex',
            alignItems: 'center',
            gap: 1.5,
            cursor: 'pointer',
            transition: 'background 0.12s',
            '&:hover': { bgcolor: kuraColors.bg3 },
          }}
        >
          <Box sx={{ position: 'relative' }}>
            <Avatar sx={{ width: 28, height: 28, fontSize: '0.75rem', bgcolor: kuraColors.accent, color: '#fff' }}>
              {user?.email?.charAt(0).toUpperCase() || 'U'}
            </Avatar>
            <Tooltip title={connected ? 'Connecté' : 'Déconnecté'}>
              <Box sx={{
                position: 'absolute',
                bottom: 0,
                right: 0,
                width: 8,
                height: 8,
                borderRadius: '50%',
                bgcolor: connected ? kuraColors.success : kuraColors.error,
                border: `2px solid ${kuraColors.bg1}`,
              }} />
            </Tooltip>
          </Box>
          <Box sx={{ flex: 1, minWidth: 0 }}>
            <Typography sx={{ fontSize: '0.8rem', fontWeight: 500, color: kuraColors.text0, lineHeight: 1.3, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {user?.username || user?.email?.split('@')[0]}
            </Typography>
            <Typography sx={{ fontSize: '0.7rem', color: kuraColors.text2, lineHeight: 1.3, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {user?.email}
            </Typography>
          </Box>
          <KeyboardArrowDown sx={{ fontSize: 14, color: kuraColors.text2 }} />
        </Box>

        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={() => setAnchorEl(null)}
          anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
          transformOrigin={{ vertical: 'bottom', horizontal: 'right' }}
        >
          <MenuItem onClick={handleLogout} sx={{ color: kuraColors.error, gap: 1 }}>
            <Logout sx={{ fontSize: 16 }} /> Déconnexion
          </MenuItem>
        </Menu>
      </Box>
    </Box>
  )

  return (
    <Box sx={{ display: 'flex', minHeight: '100vh', position: 'relative' }}>
      <AnimatedBackground />
      <Box component="nav" sx={{ width: { sm: DRAWER_WIDTH }, flexShrink: { sm: 0 }, zIndex: 1 }}>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': {
              width: DRAWER_WIDTH,
              bgcolor: kuraColors.bg1,
              borderRight: `1px solid ${kuraColors.border1}`,
              boxSizing: 'border-box',
            },
          }}
          open
        >
          {sidebar}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 4,
          width: { sm: `calc(100% - ${DRAWER_WIDTH}px)` },
          minHeight: '100vh',
          position: 'relative',
          zIndex: 1,
          bgcolor: 'transparent',
        }}
      >
        <Outlet />
      </Box>
    </Box>
  )
}
