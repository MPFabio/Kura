import { render, screen } from '@testing-library/react'
import CodeBlock from './CodeBlock'

describe('CodeBlock', () => {
  it('affiche le label quand il est fourni', () => {
    render(
      <CodeBlock language="bash" label="Commande" showLineNumbers={false}>
        echo "hello"
      </CodeBlock>,
    )

    expect(screen.getByText('Commande')).toBeInTheDocument()
  })

  it('rend le contenu du code', () => {
    const code = 'console.log("test")'

    render(
      <CodeBlock language="javascript" showLineNumbers={false}>
        {code}
      </CodeBlock>,
    )

    // Le code est fragmenté en plusieurs spans par Prism, on vérifie donc des morceaux
    expect(screen.getByText('console')).toBeInTheDocument()
    expect(screen.getByText('log')).toBeInTheDocument()
  })
})

