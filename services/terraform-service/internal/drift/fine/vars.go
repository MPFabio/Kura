package fine

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/modulops/terraform-service/internal/client"
)

// declaredVariable décrit une variable Terraform/OpenTofu déclarée dans le code.
type declaredVariable struct {
	Name       string
	HasDefault bool
	Type       string // "string", "number", "bool", "" si inconnu/complexe
}

// declaredVariables extrait les variables déclarées (blocs `variable "nom" { ... }`)
// dans les fichiers .tf fournis, avec leur type simple et la présence d'un défaut.
func declaredVariables(files []client.TFFile) []declaredVariable {
	parser := hclparse.NewParser()
	var result []declaredVariable
	for _, f := range files {
		hclFile, diags := parser.ParseHCL([]byte(f.Content), f.Path)
		if diags.HasErrors() || hclFile == nil {
			continue
		}
		content, _, _ := hclFile.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "variable", LabelNames: []string{"name"}},
			},
		})
		for _, block := range content.Blocks {
			if block.Type != "variable" || len(block.Labels) != 1 {
				continue
			}
			v := declaredVariable{Name: block.Labels[0]}
			attrs, _ := block.Body.JustAttributes()
			if _, ok := attrs["default"]; ok {
				v.HasDefault = true
			}
			if typeAttr, ok := attrs["type"]; ok {
				// Le type est une expression de type (ex: string, number, bool,
				// list(string)...). On ne s'intéresse qu'aux types primitifs,
				// dont l'expression source est un simple identifiant.
				if traversal, diags := hcl.AbsTraversalForExpr(typeAttr.Expr); !diags.HasErrors() && len(traversal) == 1 {
					v.Type = traversal.RootName()
				}
			}
			result = append(result, v)
		}
	}
	return result
}

// buildTFVarsJSON génère le contenu d'un fichier terraform.tfvars.json affectant
// une valeur à chaque variable déclarée sans valeur par défaut, afin que
// `tofu plan -refresh-only` puisse s'exécuter sans -var/-var-file fournis par
// l'utilisateur. Les variables sans défaut sont résolues, par ordre de
// priorité :
//  1. correspondance exacte de nom avec une sortie (output) du tfstate ;
//  2. pour les variables "zone"/"region", une sortie dont le nom contient
//     ce mot-clé, ou (pour "region") une région dérivée d'une sortie "zone" ;
//  3. valeur zéro adaptée au type déclaré (0, false, "", ou "" par défaut).
//
// Les variables ayant un défaut dans le code sont omises pour laisser
// `tofu` utiliser ce défaut.
func buildTFVarsJSON(vars []declaredVariable, outputs map[string]interface{}) ([]byte, error) {
	values := make(map[string]interface{})
	for _, v := range vars {
		if v.HasDefault {
			continue
		}
		if out, ok := outputs[v.Name]; ok {
			values[v.Name] = out
			continue
		}
		if alt, ok := matchByHeuristic(v.Name, outputs); ok {
			values[v.Name] = alt
			continue
		}
		values[v.Name] = zeroValueForType(v.Type)
	}
	if len(values) == 0 {
		return nil, nil
	}
	return json.Marshal(values)
}

// matchByHeuristic tente de trouver une valeur pertinente pour une variable
// "region"/"zone" dont le nom ne correspond pas exactement à une sortie.
//
//   - Si la variable contient "zone", on cherche une sortie dont le nom
//     contient "zone".
//   - Si la variable contient "region", on cherche d'abord une sortie dont le
//     nom contient "region" ; à défaut, on dérive la région à partir d'une
//     sortie "zone" (ex: "europe-west1-b" -> "europe-west1"), une zone GCP
//     étant toujours "<région>-<lettre>".
func matchByHeuristic(varName string, outputs map[string]interface{}) (interface{}, bool) {
	lowerName := strings.ToLower(varName)

	if strings.Contains(lowerName, "zone") {
		if v, ok := findOutputContaining(outputs, "zone"); ok {
			return v, true
		}
	}

	if strings.Contains(lowerName, "region") {
		if v, ok := findOutputContaining(outputs, "region"); ok {
			return v, true
		}
		if zone, ok := findOutputContaining(outputs, "zone"); ok {
			if zoneStr, ok := zone.(string); ok {
				if region := regionFromZone(zoneStr); region != "" {
					return region, true
				}
			}
		}
	}

	return nil, false
}

func findOutputContaining(outputs map[string]interface{}, keyword string) (interface{}, bool) {
	for outName, outVal := range outputs {
		if strings.Contains(strings.ToLower(outName), keyword) {
			return outVal, true
		}
	}
	return nil, false
}

// regionFromZone dérive une région GCP à partir d'une zone (ex: "europe-west1-b" -> "europe-west1").
func regionFromZone(zone string) string {
	idx := strings.LastIndex(zone, "-")
	if idx <= 0 {
		return ""
	}
	return zone[:idx]
}

func zeroValueForType(t string) interface{} {
	switch t {
	case "number":
		return 0
	case "bool":
		return false
	default:
		return ""
	}
}
