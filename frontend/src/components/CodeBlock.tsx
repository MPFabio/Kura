import { Box, Typography } from '@mui/material'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { jellyfishColors } from '../theme'

// Style type VSCode Dark+ : mots-clés, chaînes et variables bien différenciés
export const vscodeLikeStyle = {
  ...oneDark,
  'code[class*="language-"]': {
    ...oneDark['code[class*="language-"]'],
    color: '#d4d4d4',
  },
  'pre[class*="language-"]': {
    ...oneDark['pre[class*="language-"]'],
    background: '#1e1e1e',
  },
  comment: { color: '#6a9955', fontStyle: 'italic' },
  keyword: { color: '#569cd6' },
  string: { color: '#ce9178' },
  number: { color: '#b5cea8' },
  boolean: { color: '#569cd6' },
  function: { color: '#dcdcaa' },
  operator: { color: '#d4d4d4' },
  punctuation: { color: '#d4d4d4' },
  property: { color: '#9cdcfe' },
  tag: { color: '#569cd6' },
  'attr-name': { color: '#9cdcfe' },
  'attr-value': { color: '#ce9178' },
  variable: { color: '#9cdcfe' },
  'class-name': { color: '#4ec9b0' },
  selector: { color: '#d7ba7d' },
  'plain-text': { color: '#d4d4d4' },
}

export type CodeBlockProps = { language: string; label?: string; children: string; showLineNumbers?: boolean }

export default function CodeBlock({ language, label, children, showLineNumbers = true }: CodeBlockProps) {
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
          overflow: 'hidden',
          border: '1px solid rgba(255,255,255,0.12)',
          boxShadow: '0 2px 8px rgba(0,0,0,0.35)',
          '& pre': {
            margin: 0,
            fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace !important',
            fontSize: '0.8125rem !important',
            lineHeight: '1.65 !important',
          },
          '& code': { fontFamily: 'inherit !important', background: 'none !important', padding: 0 },
        }}
      >
        <SyntaxHighlighter
          language={language}
          style={vscodeLikeStyle}
          customStyle={{
            margin: 0,
            padding: '1.125rem 1.5rem',
            background: '#1e1e1e',
            fontSize: '0.8125rem',
            lineHeight: 1.65,
          }}
          codeTagProps={{ style: { fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace' } }}
          showLineNumbers={showLineNumbers && language !== 'bash' && language !== 'shell'}
          lineNumberStyle={{
            minWidth: '2.25em',
            color: 'rgba(255,255,255,0.3)',
            userSelect: 'none',
            paddingRight: '1em',
          }}
        >
          {(children ?? '').trim()}
        </SyntaxHighlighter>
      </Box>
    </Box>
  )
}
