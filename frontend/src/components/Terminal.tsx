import { useEffect, useRef, useState } from 'react'
import { Box, Alert, CircularProgress, Typography } from '@mui/material'
import { Terminal as XTerm } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

interface TerminalProps {
  namespace: string
  pod: string
  container?: string
  open: boolean
}

export default function Terminal({ namespace, pod, container, open }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const handleResizeRef = useRef<(() => void) | null>(null)
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open || !namespace || !pod) {
      console.log('[Terminal] Pas de connexion - open:', open, 'namespace:', namespace, 'pod:', pod)
      return
    }
    
    // Réinitialiser l'état
    setConnected(false)
    setError(null)

    // Attendre un peu que le DOM soit prêt
    // Utiliser requestAnimationFrame pour s'assurer que le DOM est complètement rendu
    const initTimeout = setTimeout(() => {
      requestAnimationFrame(() => {
        if (!terminalRef.current) {
          console.error('[Terminal] terminalRef.current est null après timeout et requestAnimationFrame')
          setError('Erreur: le terminal ne peut pas être initialisé. Vérifiez que le composant est correctement monté.')
          return
        }

        console.log('[Terminal] Initialisation du terminal...')

        // Initialiser xterm.js
        const xterm = new XTerm({
          cursorBlink: true,
          fontSize: 14,
          fontFamily: 'Consolas, "Courier New", monospace',
          theme: {
            background: '#1e1e1e',
            foreground: '#d4d4d4',
            cursor: '#d4d4d4',
          },
        })

        const fitAddon = new FitAddon()
        xterm.loadAddon(fitAddon)
        
        try {
          xterm.open(terminalRef.current)
          fitAddon.fit()
          xtermRef.current = xterm
          fitAddonRef.current = fitAddon
          console.log('[Terminal] xterm.js initialisé')
        } catch (err) {
          console.error('[Terminal] Erreur lors de l\'initialisation de xterm:', err)
          setError('Erreur lors de l\'initialisation du terminal')
          return
        }

        // Construire l'URL WebSocket
        // En dev, se connecter directement au k8s-service sur le port 8081
        // En prod, passer par Kong sur le port 8000
        const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
        const wsHost = import.meta.env.DEV 
          ? 'localhost:8081'  // En dev, se connecter directement au k8s-service
          : window.location.host
        const wsUrl = `${wsProtocol}//${wsHost}/api/v1/k8s/namespaces/${encodeURIComponent(namespace)}/pods/${encodeURIComponent(pod)}/terminal`
        
        console.log('[Terminal] Connexion WebSocket à:', wsUrl)
        console.log('[Terminal] Namespace:', namespace, 'Pod:', pod)

        // Créer la connexion WebSocket
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('[Terminal] WebSocket connecté')
          setConnected(true)
          setError(null)
          
          // Envoyer le message d'initialisation
          const size = fitAddon.proposeDimensions()
          const initMessage = {
            type: 'init',
            namespace,
            pod,
            container: container || '',
            width: size?.cols || 80,
            height: size?.rows || 24,
          }
          console.log('[Terminal] Envoi du message init:', initMessage)
          ws.send(JSON.stringify(initMessage))
        }

        ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            
            if (message.type === 'output' || message.type === 'error') {
              // Écrire dans le terminal xterm
              if (xtermRef.current) {
                xtermRef.current.write(message.data)
              }
            }
          } catch (err) {
            console.error('[Terminal] Erreur lors du parsing du message:', err, 'Data:', event.data)
          }
        }

        ws.onerror = (err) => {
          console.error('[Terminal] Erreur WebSocket:', err)
          console.error('[Terminal] URL tentée:', wsUrl)
          console.error('[Terminal] État WebSocket:', ws.readyState)
          setError(`Erreur de connexion au terminal. Vérifiez que le k8s-service est démarré sur le port 8081 et que le pod existe.`)
          setConnected(false)
        }

        ws.onclose = (event) => {
          setConnected(false)
          console.log('[Terminal] WebSocket fermé:', event.code, event.reason)
          if (event.code !== 1000) { // 1000 = fermeture normale
            const errorMsg = `Connexion fermée (code ${event.code}${event.reason ? ': ' + event.reason : ''})`
            if (xtermRef.current) {
              xtermRef.current.write(`\r\n\x1b[31m[${errorMsg}]\x1b[0m\r\n`)
            }
            if (event.code === 1006) {
              setError('Connexion refusée. Vérifiez que le k8s-service est démarré sur le port 8081.')
            } else if (event.code === 1002) {
              setError('Erreur de protocole. Vérifiez la configuration WebSocket.')
            } else if (event.code === 1003) {
              setError('Type de données non supporté.')
            } else {
              setError(`Erreur de connexion (code ${event.code})`)
            }
          } else {
            if (xtermRef.current) {
              xtermRef.current.write('\r\n\x1b[31m[Connexion fermée]\x1b[0m\r\n')
            }
          }
        }

        // Gérer les entrées clavier via xterm
        xterm.onData((data) => {
          if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({
              type: 'input',
              data: data,
            }))
          }
        })

        // Gérer le resize du terminal
        const handleResize = () => {
          if (fitAddonRef.current) {
            fitAddonRef.current.fit()
            const size = fitAddonRef.current.proposeDimensions()
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && size) {
              wsRef.current.send(JSON.stringify({
                type: 'resize',
                width: size.cols,
                height: size.rows,
              }))
            }
          }
        }

        handleResizeRef.current = handleResize
        window.addEventListener('resize', handleResize)
      })
    }, 200) // Délai de 200ms pour laisser le DOM se stabiliser

    return () => {
      clearTimeout(initTimeout)
      if (handleResizeRef.current) {
        window.removeEventListener('resize', handleResizeRef.current)
        handleResizeRef.current = null
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
      if (xtermRef.current) {
        xtermRef.current.dispose()
        xtermRef.current = null
      }
      if (fitAddonRef.current) {
        fitAddonRef.current = null
      }
    }
  }, [open, namespace, pod, container])

  return (
    <Box sx={{ position: 'relative', width: '100%', height: '500px' }}>
      {/* Le terminal doit toujours être rendu pour que le ref soit attaché */}
      <Box
        ref={terminalRef}
        sx={{
          width: '100%',
          height: '100%',
          bgcolor: '#1e1e1e',
          padding: 1,
          border: '1px solid #333',
          borderRadius: 1,
          display: error ? 'none' : 'block',
        }}
      />
      
      {/* Overlay d'erreur */}
      {error && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            bgcolor: 'rgba(0, 0, 0, 0.8)',
            borderRadius: 1,
          }}
        >
          <Alert severity="error" sx={{ maxWidth: '80%' }}>
            {error}
          </Alert>
        </Box>
      )}
      
      {/* Overlay de chargement */}
      {!connected && open && !error && (
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            display: 'flex',
            flexDirection: 'column',
            justifyContent: 'center',
            alignItems: 'center',
            bgcolor: 'rgba(0, 0, 0, 0.8)',
            borderRadius: 1,
            gap: 2,
          }}
        >
          <CircularProgress />
          <Typography variant="body2" color="text.secondary">
            Connexion au terminal...
          </Typography>
        </Box>
      )}
    </Box>
  )
}
