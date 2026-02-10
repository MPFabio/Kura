import { Box, Typography } from '@mui/material'
import { jellyfishColors } from '../theme'

export type CodeBlockProps = {
  language: string
  label?: string
  children: string
  showLineNumbers?: boolean
}

export default function CodeBlock({
  language,
  label,
  children,
  showLineNumbers = true,
}: CodeBlockProps) {
  const text = (children ?? '').trim()
  const lines = text.split('\n')
  const withNumbers = showLineNumbers && language !== 'bash' && language !== 'shell'

  return (
    <Box sx={{ mb: 2.5 }}>
      {label && (
        <Typography
          variant="caption"
          sx={{
            display: 'block',
            color: jellyfishColors.grayLight,
            fontWeight: 600,
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            mb: 0.75,
          }}
        >
          {label}
        </Typography>
      )}
      <Box
        sx={{
          borderRadius: 1,
          overflow: 'auto',
          border: '1px solid rgba(255,255,255,0.12)',
          boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
          background: '#1e1e1e',
          '& pre': {
            margin: 0,
            padding: '1.125rem 1.5rem',
            fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace',
            fontSize: '0.8125rem',
            lineHeight: 1.65,
            color: '#d4d4d4',
          },
          '& code': { fontFamily: 'inherit', background: 'none', padding: 0 },
        }}
      >
        <pre>
          <code>
            {withNumbers
              ? lines.map((line, i) => (
                  <span key={i} style={{ display: 'block' }}>
                    <span
                      style={{
                        display: 'inline-block',
                        minWidth: '2.25em',
                        color: 'rgba(255,255,255,0.3)',
                        userSelect: 'none',
                        marginRight: '1em',
                      }}
                    >
                      {i + 1}
                    </span>
                    {line || '\n'}
                  </span>
                ))
              : text}
          </code>
        </pre>
      </Box>
    </Box>
  )
}
