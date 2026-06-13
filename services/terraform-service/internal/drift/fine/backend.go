package fine

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// stripBackendBlock retire les blocs `backend "..." { ... }` à l'intérieur des
// blocs `terraform { ... }` d'un fichier .tf, afin que le sandbox utilise
// systématiquement le backend local (le tfstate fourni est écrit directement
// sur disque). Si le fichier ne contient pas de bloc `terraform`, ou si le
// parsing échoue, le contenu est retourné inchangé.
func stripBackendBlock(content []byte, filename string) []byte {
	f, diags := hclwrite.ParseConfig(content, filename, hcl.InitialPos)
	if diags.HasErrors() || f == nil {
		return content
	}

	changed := false
	for _, block := range f.Body().Blocks() {
		if block.Type() != "terraform" {
			continue
		}
		for _, inner := range block.Body().Blocks() {
			if inner.Type() == "backend" {
				block.Body().RemoveBlock(inner)
				changed = true
			}
		}
	}

	if !changed {
		return content
	}
	return f.Bytes()
}
