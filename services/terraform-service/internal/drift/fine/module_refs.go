package fine

import (
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/modulops/terraform-service/internal/client"
)

// fileFunctions liste les fonctions Terraform/OpenTofu qui lisent un fichier
// du système de fichiers local à partir d'un chemin.
var fileFunctions = map[string]bool{
	"file":         true,
	"filebase64":   true,
	"filemd5":      true,
	"filesha1":     true,
	"filesha256":   true,
	"filesha512":   true,
	"templatefile": true,
}

// ExtractModuleFileRefs analyse les fichiers .tf fournis (chemins relatifs à
// la racine du dépôt) et retourne, pour chaque appel à file()/templatefile()/...
// dont l'argument est de la forme "${path.module}/<chemin relatif>", le chemin
// du fichier référencé relatif à la racine du dépôt.
//
// dir est le répertoire (relatif à la racine du dépôt) contenant les .tf
// analysés — utilisé pour résoudre "${path.module}/<chemin relatif>".
func ExtractModuleFileRefs(files []client.TFFile, dir string) []string {
	seen := make(map[string]bool)
	var refs []string

	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".tf") {
			continue
		}
		hclFile, diags := hclsyntax.ParseConfig([]byte(f.Content), f.Path, hcl.InitialPos)
		if diags.HasErrors() || hclFile == nil {
			continue
		}
		body, ok := hclFile.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}
		walkBody(body, func(rel string) {
			if rel == "" {
				return
			}
			repoPath := path.Join(dir, rel)
			if !seen[repoPath] {
				seen[repoPath] = true
				refs = append(refs, repoPath)
			}
		})
	}

	return refs
}

// walkBody parcourt récursivement un corps HCL (blocs imbriqués et attributs)
// et appelle onRef pour chaque référence "${path.module}/<chemin relatif>"
// trouvée dans un appel à une fonction de lecture de fichier.
func walkBody(body *hclsyntax.Body, onRef func(rel string)) {
	for _, attr := range body.Attributes {
		walkExpr(attr.Expr, onRef)
	}
	for _, block := range body.Blocks {
		walkBody(block.Body, onRef)
	}
}

func walkExpr(expr hclsyntax.Expression, onRef func(rel string)) {
	switch e := expr.(type) {
	case *hclsyntax.FunctionCallExpr:
		if fileFunctions[e.Name] && len(e.Args) > 0 {
			if rel, ok := pathModuleRelative(e.Args[0]); ok {
				onRef(rel)
			}
		}
		for _, arg := range e.Args {
			walkExpr(arg, onRef)
		}
	case *hclsyntax.TemplateExpr:
		for _, part := range e.Parts {
			walkExpr(part, onRef)
		}
	case *hclsyntax.TemplateWrapExpr:
		walkExpr(e.Wrapped, onRef)
	case *hclsyntax.ConditionalExpr:
		walkExpr(e.Condition, onRef)
		walkExpr(e.TrueResult, onRef)
		walkExpr(e.FalseResult, onRef)
	case *hclsyntax.BinaryOpExpr:
		walkExpr(e.LHS, onRef)
		walkExpr(e.RHS, onRef)
	case *hclsyntax.TupleConsExpr:
		for _, item := range e.Exprs {
			walkExpr(item, onRef)
		}
	case *hclsyntax.ObjectConsExpr:
		for _, item := range e.Items {
			walkExpr(item.KeyExpr, onRef)
			walkExpr(item.ValueExpr, onRef)
		}
	}
}

// pathModuleRelative détecte une expression de la forme
// "${path.module}/<chemin relatif>" (un TemplateExpr dont la première partie
// est la traversal "path.module" et la seconde une chaîne littérale commençant
// par "/"), et retourne <chemin relatif> nettoyé.
func pathModuleRelative(expr hclsyntax.Expression) (string, bool) {
	tmpl, ok := expr.(*hclsyntax.TemplateExpr)
	if !ok || len(tmpl.Parts) != 2 {
		return "", false
	}

	scopeTraversal, ok := tmpl.Parts[0].(*hclsyntax.ScopeTraversalExpr)
	if !ok {
		return "", false
	}
	if !isPathModuleTraversal(scopeTraversal.Traversal) {
		return "", false
	}

	lit, ok := tmpl.Parts[1].(*hclsyntax.LiteralValueExpr)
	if !ok || lit.Val.Type() != cty.String {
		return "", false
	}
	suffix := lit.Val.AsString()
	suffix = strings.TrimPrefix(suffix, "/")
	if suffix == "" {
		return "", false
	}
	return path.Clean(suffix), true
}

// isPathModuleTraversal vérifie qu'une traversal correspond à "path.module"
// (root="path", premier attribut="module").
func isPathModuleTraversal(traversal hcl.Traversal) bool {
	if traversal.RootName() != "path" {
		return false
	}
	if len(traversal) < 2 {
		return false
	}
	attr, ok := traversal[1].(hcl.TraverseAttr)
	return ok && attr.Name == "module"
}
