import { useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import {
  Box,
  Drawer,
  AppBar,
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
} from '@mui/material'
import {
  Menu as MenuIcon,
  Dashboard as DashboardIcon,
  Cloud as CloudIcon,
  Settings as SettingsIcon,
  AccountCircle,
  Logout,
  Storage as StorageIcon,
  PlayArrow as PlayArrowIcon,
  Timeline as TimelineIcon,
  Notifications as NotificationsIcon,
  BarChart as BarChartIcon,
} from '@mui/icons-material'
import { useAuth } from '../contexts/AuthContext'
import { useSocket } from '../contexts/SocketContext'
import Logo from './Logo'

const drawerWidth = 240

const menuItems = [
  { text: 'Modules', icon: <DashboardIcon />, path: '/modules' },
  { text: 'Terraform', icon: <StorageIcon />, path: '/terraform' },
  { text: 'Kubernetes', icon: <CloudIcon />, path: '/k8s' },
  { text: 'Ansible', icon: <PlayArrowIcon />, path: '/ansible' },
  { text: 'Monitoring', icon: <BarChartIcon />, path: '/metrics' },
  { text: 'Pipelines', icon: <TimelineIcon />, path: '/pipelines' },
  { text: 'Alertes', icon: <NotificationsIcon />, path: '/alerts' },
  { text: 'Paramètres', icon: <SettingsIcon />, path: '/settings' },
]

export default function Layout() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()
  const { connected } = useSocket()

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
    <div>
      <Toolbar
        sx={{
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          py: 3,
          px: 2,
          minHeight: 'auto',
        }}
      >
        <Logo variant="full" size="small" />
      </Toolbar>
      <Divider sx={{ borderColor: 'rgba(160, 160, 160, 0.1)', mb: 1 }} />
      <List>
        {menuItems.map((item) => (
          <ListItem key={item.text} disablePadding>
            <ListItemButton
              selected={location.pathname === item.path}
              onClick={() => {
                navigate(item.path)
                setMobileOpen(false)
              }}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </div>
  )

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar
        position="fixed"
        sx={{
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          ml: { sm: `${drawerWidth}px` },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography 
            variant="h6" 
            noWrap 
            component="div" 
            sx={{ 
              flexGrow: 1,
              fontFamily: '"Inter", sans-serif',
              fontWeight: 700,
              fontSize: '1.25rem',
              letterSpacing: '0.15em',
              background: 'linear-gradient(135deg, #00FFFF, #BF00FF)',
              backgroundClip: 'text',
              WebkitBackgroundClip: 'text',
              WebkitTextFillColor: 'transparent',
              textShadow: 'none',
              position: 'relative',
              '&::before': {
                content: '""',
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                background: 'linear-gradient(135deg, rgba(0, 255, 255, 0.5), rgba(191, 0, 255, 0.5))',
                filter: 'blur(10px)',
                zIndex: -1,
                borderRadius: '4px',
              },
            }}
          >
            Plateforme DevOps KURA
          </Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Box
              sx={{
                width: 10,
                height: 10,
                borderRadius: '50%',
                bgcolor: connected ? '#00FFFF' : '#FF4500',
                mr: 1,
                boxShadow: connected 
                  ? '0 0 10px rgba(0, 255, 255, 0.8), 0 0 20px rgba(0, 255, 255, 0.4)'
                  : '0 0 10px rgba(255, 69, 0, 0.8), 0 0 20px rgba(255, 69, 0, 0.4)',
                animation: connected ? 'breathingGlow 2s ease-in-out infinite' : 'none',
              }}
            />
            <IconButton
              size="large"
              edge="end"
              aria-label="account menu"
              aria-controls="account-menu"
              aria-haspopup="true"
              onClick={handleMenuClick}
              color="inherit"
            >
              <Avatar sx={{ width: 32, height: 32 }}>
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
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
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
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
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
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          mt: 8,
        }}
      >
        <Outlet />
      </Box>
    </Box>
  )
}
