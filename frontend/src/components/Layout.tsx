import { useState } from 'react'
import React from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Box,
  Drawer,
  Toolbar,
  List,
  Typography,
  Divider,
  IconButton,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Avatar,
  Menu,
  MenuItem,
  Select,
  FormControl,
  Chip,
} from '@mui/material'
import {
  AccountCircle,
  Logout,
  Folder as FolderIcon,
} from '@mui/icons-material'
import { useAuth } from '../contexts/AuthContext'
import { useSocket } from '../contexts/SocketContext'
import { useProject } from '../contexts/ProjectContext'
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

const drawerWidth = 240

const menuItems = [
  { text: 'Modules', icon: <ModulesIcon />, path: '/modules', useCustomIcon: true },
  { text: 'Terraform', icon: <TerraformIcon />, path: '/terraform', useCustomIcon: true },
  { text: 'Kubernetes', icon: <KubernetesIcon />, path: '/k8s', useCustomIcon: true },
  { text: 'Ansible', icon: <AnsibleIcon />, path: '/ansible', useCustomIcon: true },
  { text: 'Monitoring', icon: <MonitoringIcon />, path: '/metrics', useCustomIcon: true },
  { text: 'Pipelines', icon: <PipelinesIcon />, path: '/pipelines', useCustomIcon: true },
  { text: 'Alertes', icon: <AlertsIcon />, path: '/alerts', useCustomIcon: true },
  { text: 'Paramètres', icon: <SettingsIcon />, path: '/settings', useCustomIcon: true },
]

export default function Layout() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()
  const { connected } = useSocket()
  const { currentProject, projects, setCurrentProject } = useProject()

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen)
  }

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget)
  }

  const handleMenuClose = () => {
    setAnchorEl(null)
  }

  const handleLogout = async () => {
    await logout()
    navigate('/login')
    handleMenuClose()
  }

  const drawer = (
    <div style={{ background: 'transparent' }}>
      <Box
        sx={{
          background: 'rgba(0, 229, 255, 0.05)',
          borderBottom: '2px solid rgba(0, 229, 255, 0.3)',
          px: 2,
          py: 2,
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Box
              sx={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                bgcolor: connected ? '#00FFFF' : '#FF4500',
                boxShadow: connected 
                  ? '0 0 8px rgba(0, 255, 255, 0.8), 0 0 16px rgba(0, 255, 255, 0.4)'
                  : '0 0 8px rgba(255, 69, 0, 0.8), 0 0 16px rgba(255, 69, 0, 0.4)',
                animation: connected ? 'breathingGlow 2s ease-in-out infinite' : 'none',
              }}
            />
            <IconButton
              size="small"
              edge="end"
              aria-label="account menu"
              aria-controls="account-menu"
              aria-haspopup="true"
              onClick={handleMenuClick}
              sx={{ p: 0.5 }}
            >
              <Avatar sx={{ width: 28, height: 28, fontSize: '0.75rem' }}>
                {user?.email?.charAt(0).toUpperCase() || 'U'}
              </Avatar>
            </IconButton>
            <Menu
              id="account-menu"
              anchorEl={anchorEl}
              open={Boolean(anchorEl)}
              onClose={handleMenuClose}
            >
              <MenuItem onClick={() => navigate('/settings')}>
                <AccountCircle sx={{ mr: 1 }} />
                Profil
              </MenuItem>
              <MenuItem onClick={handleLogout}>
                <Logout sx={{ mr: 1 }} />
                Déconnexion
              </MenuItem>
            </Menu>
          </Box>
        </Box>
      </Box>
      <Toolbar
        disableGutters
        sx={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          padding: 0,
          height: 120,
          minHeight: 120,
          overflow: 'hidden',
          background: 'rgba(0, 0, 0, 0.3)',
          borderTop: '1px solid rgba(0, 229, 255, 0.2)',
          borderBottom: '2px solid rgba(0, 229, 255, 0.3)',
          boxSizing: 'border-box',
          '&.MuiToolbar-root': { minHeight: 120, padding: 0 },
        }}
      >
        <Box sx={{ width: '100%', flex: 1, minHeight: 0, display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
          {/* Marge au-dessus du logo (en px) pour l’équilibre vertical ; augmenter pour descendre le bloc */}
          <Logo variant="full" size="small" sx={{ margin: 0 }} />
        </Box>
      </Toolbar>
      <Divider sx={{ borderColor: 'rgba(0, 184, 212, 0.2)' }} />
      <Box sx={{ px: 2, py: 2, background: 'rgba(171, 71, 188, 0.05)', borderBottom: '2px solid rgba(171, 71, 188, 0.3)' }}>
        <FormControl fullWidth size="small">
          <Select
            value={currentProject?.id || ''}
            onChange={(e) => {
              const project = projects.find(p => p.id === e.target.value)
              if (project) {
                setCurrentProject(project)
                if (location.pathname === '/projects') {
                  navigate('/modules')
                }
              }
            }}
            displayEmpty
            MenuProps={{
              PaperProps: {
                sx: {
                  backgroundColor: 'rgba(26, 35, 50, 0.98)',
                  backdropFilter: 'blur(30px) saturate(180%)',
                  border: '1px solid rgba(0, 229, 255, 0.3)',
                  borderRadius: '12px',
                  mt: 1,
                },
              },
            }}
            sx={{
              color: '#f0f0f0',
              backgroundColor: 'rgba(26, 35, 50, 0.8)',
              backdropFilter: 'blur(10px)',
              '& .MuiOutlinedInput-notchedOutline': {
                borderColor: 'rgba(0, 229, 255, 0.3)',
              },
              '&:hover .MuiOutlinedInput-notchedOutline': {
                borderColor: 'rgba(0, 229, 255, 0.5)',
              },
              '&.Mui-focused .MuiOutlinedInput-notchedOutline': {
                borderColor: 'rgba(0, 229, 255, 0.8)',
              },
              '& .MuiSelect-icon': {
                color: '#b8b8b8',
              },
              '& .MuiSelect-select': {
                backgroundColor: 'transparent',
              },
            }}
            renderValue={(selected) => {
              if (!selected) {
                return (
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <FolderIcon sx={{ fontSize: 16, color: '#b8b8b8' }} />
                    <Typography variant="body2" sx={{ color: '#b8b8b8', fontWeight: 500 }}>
                      Sélectionner un projet
                    </Typography>
                  </Box>
                )
              }
              const project = projects.find(p => p.id === selected)
              return (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <FolderIcon sx={{ fontSize: 16, color: '#00E5FF' }} />
                  <Typography variant="body2" sx={{ fontWeight: 600, color: '#f0f0f0' }}>
                    {project?.name || 'Projet'}
                  </Typography>
                </Box>
              )
            }}
          >
            <MenuItem 
              value="" 
              onClick={() => navigate('/projects')}
              sx={{
                backgroundColor: 'transparent',
                '&:hover': {
                  backgroundColor: 'rgba(0, 229, 255, 0.12)',
                },
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}>
                <FolderIcon sx={{ fontSize: 16, mr: 1 }} />
                <Typography variant="body2">Gérer les projets</Typography>
              </Box>
            </MenuItem>
            {projects.map((project) => (
              <MenuItem 
                key={project.id} 
                value={project.id}
                sx={{
                  backgroundColor: currentProject?.id === project.id ? 'rgba(0, 229, 255, 0.15)' : 'transparent',
                  '&:hover': {
                    backgroundColor: 'rgba(0, 229, 255, 0.12)',
                  },
                }}
              >
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}>
                  <FolderIcon sx={{ fontSize: 16, mr: 1, color: '#00E5FF' }} />
                  <Typography variant="body2">{project.name}</Typography>
                  {currentProject?.id === project.id && (
                    <Chip
                      label="Actif"
                      size="small"
                      sx={{
                        ml: 'auto',
                        background: 'transparent',
                        border: '1px solid #66BB6A',
                        color: '#66BB6A',
                        fontSize: '0.65rem',
                        height: '20px',
                        fontWeight: 700,
                      }}
                    />
                  )}
                </Box>
              </MenuItem>
            ))}
          </Select>
        </FormControl>
      </Box>
      <List sx={{ px: 1, py: 2, background: 'rgba(0, 0, 0, 0.2)' }}>
        {menuItems.map((item) => {
          const isSelected = location.pathname === item.path || 
            (item.path === '/k8s' && location.pathname.startsWith('/k8s'))
          return (
            <ListItem key={item.text} disablePadding>
              <ListItemButton
                selected={isSelected}
                onClick={() => {
                  navigate(item.path)
                  setMobileOpen(false)
                }}
              >
                <ListItemIcon>
                  {item.useCustomIcon ? (
                    React.cloneElement(item.icon as React.ReactElement, { active: isSelected })
                  ) : (
                    item.icon
                  )}
                </ListItemIcon>
                <ListItemText 
                  primary={item.text}
                  primaryTypographyProps={{
                    sx: isSelected
                      ? { color: '#f0f0f0', fontWeight: 600 }
                      : {
                          color: '#b8b8b8',
                          fontWeight: 500,
                        },
                  }}
                />
              </ListItemButton>
            </ListItem>
          )
        })}
      </List>
    </div>
  )

  return (
    <Box sx={{ display: 'flex', position: 'relative', minHeight: '100vh' }}>
      <AnimatedBackground />
      <Box
        component="nav"
        sx={{ 
          width: { sm: drawerWidth }, 
          flexShrink: { sm: 0 }, 
          position: 'relative', 
          zIndex: 1,
          background: 'transparent',
        }}
        aria-label="navigation"
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{
            keepMounted: true,
          }}
          sx={{
            display: { xs: 'block', sm: 'none' },
            '& .MuiDrawer-paper': { 
              boxSizing: 'border-box', 
              width: drawerWidth,
              background: 'transparent !important',
              backgroundColor: 'transparent !important',
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { 
              boxSizing: 'border-box', 
              width: drawerWidth,
              background: 'transparent !important',
              backgroundColor: 'transparent !important',
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 4,
          pt: 5,
          width: '100%',
          backgroundColor: 'transparent',
          minHeight: '100vh',
          position: 'relative',
          zIndex: 1,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  )
}
