package parser

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/modulops/terraform-service/internal/models"
)

// TFStateParser parse les fichiers tfstate.
type TFStateParser struct{}

// NewTFStateParser crée un nouveau parser tfstate.
func NewTFStateParser() *TFStateParser {
	return &TFStateParser{}
}

// ParseState parse un fichier tfstate depuis un reader.
func (p *TFStateParser) ParseState(reader io.Reader) (*models.TerraformState, error) {
	var state models.TerraformState

	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du tfstate: %w", err)
	}

	// Valider les champs essentiels
	if state.Version == 0 {
		return nil, fmt.Errorf("version invalide ou manquante dans le tfstate")
	}

	return &state, nil
}

// ParseStateFromBytes parse un tfstate depuis un tableau de bytes.
func (p *TFStateParser) ParseStateFromBytes(data []byte) (*models.TerraformState, error) {
	var state models.TerraformState

	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du tfstate: %w", err)
	}

	if state.Version == 0 {
		return nil, fmt.Errorf("version invalide ou manquante dans le tfstate")
	}

	return &state, nil
}

// ExtractResources extrait toutes les ressources d'un état Terraform.
func (p *TFStateParser) ExtractResources(state *models.TerraformState) []models.Resource {
	if state == nil {
		return nil
	}
	return state.Resources
}

// ExtractOutputs extrait toutes les sorties d'un état Terraform.
func (p *TFStateParser) ExtractOutputs(state *models.TerraformState) map[string]models.Output {
	if state == nil {
		return nil
	}
	return state.Outputs
}

// GetResourceByAddress retourne une ressource spécifique par son adresse.
func (p *TFStateParser) GetResourceByAddress(state *models.TerraformState, address string) (*models.Resource, error) {
	if state == nil {
		return nil, fmt.Errorf("état Terraform vide")
	}

	for i := range state.Resources {
		resource := &state.Resources[i]
		resourceAddr := p.BuildResourceAddress(resource)
		if resourceAddr == address {
			return resource, nil
		}

		// Vérifier aussi dans les instances
		for j := range resource.Instances {
			instanceAddr := fmt.Sprintf("%s[%d]", resourceAddr, j)
			if instanceAddr == address {
				return resource, nil
			}
		}
	}

	return nil, fmt.Errorf("ressource non trouvée: %s", address)
}

// BuildResourceAddress construit l'adresse complète d'une ressource Terraform.
func (p *TFStateParser) BuildResourceAddress(resource *models.Resource) string {
	var address string

	if resource.Module != "" {
		address = resource.Module + "."
	}

	address += fmt.Sprintf("%s.%s.%s", resource.Type, resource.Name, resource.Provider)

	return address
}

// ValidateState valide qu'un état Terraform est valide.
func (p *TFStateParser) ValidateState(state *models.TerraformState) error {
	if state == nil {
		return fmt.Errorf("état Terraform vide")
	}

	if state.Version == 0 {
		return fmt.Errorf("version invalide")
	}

	if state.Serial < 0 {
		return fmt.Errorf("serial invalide")
	}

	return nil
}
