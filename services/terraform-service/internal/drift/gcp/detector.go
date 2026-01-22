package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/modulops/terraform-service/internal/models"
	"google.golang.org/api/option"
)

// Detector détecte les dérives pour les ressources GCP.
type Detector struct {
	credentialsJSON string
}

// NewDetector crée un nouveau détecteur de drift pour GCP.
func NewDetector(credentialsJSON string) (*Detector, error) {
	return &Detector{
		credentialsJSON: credentialsJSON,
	}, nil
}

// DetectDrift détecte les dérives pour toutes les ressources d'un état Terraform.
func (d *Detector) DetectDrift(ctx context.Context, stateFile *models.StateFile) ([]*models.DriftResult, error) {
	if stateFile.State == nil {
		return nil, fmt.Errorf("état vide")
	}

	results := make([]*models.DriftResult, 0)

	// Grouper les ressources par type pour optimiser les appels API
	for _, resource := range stateFile.State.Resources {
		result, err := d.detectResourceDrift(ctx, &resource)
		if err != nil {
			// En cas d'erreur, marquer comme unknown mais continuer
			result = &models.DriftResult{
				ResourceAddress: buildResourceAddress(&resource),
				ResourceType:    resource.Type,
				Status:          "unknown",
				Message:         fmt.Sprintf("Erreur lors de la détection: %v", err),
				DetectedAt:      time.Now(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// detectResourceDrift détecte les dérives pour une ressource spécifique.
func (d *Detector) detectResourceDrift(ctx context.Context, resource *models.Resource) (*models.DriftResult, error) {
	result := &models.DriftResult{
		ResourceAddress: buildResourceAddress(resource),
		ResourceType:    resource.Type,
		Status:          "unknown",
		DetectedAt:      time.Now(),
	}

	// Vérifier si la ressource a des instances
	if len(resource.Instances) == 0 {
		result.Status = "missing"
		result.Message = "Aucune instance trouvée dans le tfstate"
		return result, nil
	}

	// Détecter selon le type de ressource
	switch {
	case strings.HasPrefix(resource.Type, "google_compute_instance"):
		return d.detectComputeInstanceDrift(ctx, resource, result)
	case strings.HasPrefix(resource.Type, "google_compute_network"):
		return d.detectComputeNetworkDrift(ctx, resource, result)
	case strings.HasPrefix(resource.Type, "google_compute_firewall"):
		return d.detectComputeFirewallDrift(ctx, resource, result)
	case strings.HasPrefix(resource.Type, "google_compute_address"):
		return d.detectComputeAddressDrift(ctx, resource, result)
	default:
		result.Status = "unknown"
		result.Message = fmt.Sprintf("Type de ressource %s non supporté pour la détection de drift", resource.Type)
		return result, nil
	}
}

// detectComputeInstanceDrift détecte les dérives pour une instance Compute.
func (d *Detector) detectComputeInstanceDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}
	
	instancesClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Compute Instances: %w", err)
	}
	defer instancesClient.Close()

	differences := make([]models.DriftDifference, 0)

	for _, instance := range resource.Instances {
		if instance.Attributes == nil {
			continue
		}

		// Extraire les informations du tfstate
		name, _ := instance.Attributes["name"].(string)
		zone, _ := instance.Attributes["zone"].(string)
		project, _ := instance.Attributes["project"].(string)

		if name == "" || zone == "" || project == "" {
			result.Status = "unknown"
			result.Message = "Informations incomplètes dans le tfstate (name, zone, project requis)"
			return result, nil
		}

		// Extraire le nom de la zone (format: projects/*/zones/us-central1-a -> us-central1-a)
		zoneParts := strings.Split(zone, "/")
		zoneName := zoneParts[len(zoneParts)-1]

		// Vérifier si l'instance existe dans GCP
		req := &computepb.GetInstanceRequest{
			Project:  project,
			Zone:     zoneName,
			Instance: name,
		}

		gcpInstance, err := instancesClient.Get(ctx, req)
		if err != nil {
			result.Status = "missing"
			result.Message = fmt.Sprintf("Instance %s non trouvée dans GCP: %v", name, err)
			return result, nil
		}

		// Comparer les attributs critiques
		expectedStatus := "RUNNING"
		actualStatus := gcpInstance.GetStatus()
		if actualStatus != expectedStatus {
			differences = append(differences, models.DriftDifference{
				Attribute:  "status",
				Expected:   expectedStatus,
				Actual:     actualStatus,
				ChangeType: "modified",
			})
		}

		// Vérifier le machine type
		if expectedMachineType, ok := instance.Attributes["machine_type"].(string); ok {
			actualMachineType := gcpInstance.GetMachineType()
			expectedParts := strings.Split(expectedMachineType, "/")
			expectedMT := expectedParts[len(expectedParts)-1]
			actualParts := strings.Split(actualMachineType, "/")
			actualMT := actualParts[len(actualParts)-1]
			if expectedMT != actualMT {
				differences = append(differences, models.DriftDifference{
					Attribute:  "machine_type",
					Expected:   expectedMT,
					Actual:     actualMT,
					ChangeType: "modified",
				})
			}
		}
	}

	if len(differences) > 0 {
		result.Status = "drifted"
		result.Message = fmt.Sprintf("%d différence(s) détectée(s)", len(differences))
		result.Differences = differences
	} else {
		result.Status = "in_sync"
		result.Message = "Ressource synchronisée avec l'infrastructure GCP"
	}

	return result, nil
}

// detectComputeNetworkDrift détecte les dérives pour un réseau Compute.
func (d *Detector) detectComputeNetworkDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}
	
	networksClient, err := compute.NewNetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Networks: %w", err)
	}
	defer networksClient.Close()

	if len(resource.Instances) == 0 {
		result.Status = "missing"
		result.Message = "Aucune instance trouvée dans le tfstate"
		return result, nil
	}

	instance := resource.Instances[0]
	if instance.Attributes == nil {
		result.Status = "unknown"
		result.Message = "Attributs manquants dans le tfstate"
		return result, nil
	}

	name, _ := instance.Attributes["name"].(string)
	project, _ := instance.Attributes["project"].(string)

	if name == "" || project == "" {
		result.Status = "unknown"
		result.Message = "Informations incomplètes dans le tfstate (name, project requis)"
		return result, nil
	}

	// Vérifier si le réseau existe dans GCP
	req := &computepb.GetNetworkRequest{
		Project: project,
		Network: name,
	}

	_, err = networksClient.Get(ctx, req)
	if err != nil {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Réseau %s non trouvé dans GCP: %v", name, err)
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Réseau synchronisé avec l'infrastructure GCP"
	return result, nil
}

// detectComputeFirewallDrift détecte les dérives pour une règle firewall.
func (d *Detector) detectComputeFirewallDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}
	
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Firewalls: %w", err)
	}
	defer firewallsClient.Close()

	if len(resource.Instances) == 0 {
		result.Status = "missing"
		result.Message = "Aucune instance trouvée dans le tfstate"
		return result, nil
	}

	instance := resource.Instances[0]
	if instance.Attributes == nil {
		result.Status = "unknown"
		result.Message = "Attributs manquants dans le tfstate"
		return result, nil
	}

	name, _ := instance.Attributes["name"].(string)
	project, _ := instance.Attributes["project"].(string)

	if name == "" || project == "" {
		result.Status = "unknown"
		result.Message = "Informations incomplètes dans le tfstate (name, project requis)"
		return result, nil
	}

	// Vérifier si la règle firewall existe dans GCP
	req := &computepb.GetFirewallRequest{
		Project:  project,
		Firewall: name,
	}

	_, err = firewallsClient.Get(ctx, req)
	if err != nil {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Règle firewall %s non trouvée dans GCP: %v", name, err)
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Règle firewall synchronisée avec l'infrastructure GCP"
	return result, nil
}

// detectComputeAddressDrift détecte les dérives pour une adresse IP Compute.
func (d *Detector) detectComputeAddressDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}
	
	addressesClient, err := compute.NewAddressesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Addresses: %w", err)
	}
	defer addressesClient.Close()

	differences := make([]models.DriftDifference, 0)

	for _, instance := range resource.Instances {
		if instance.Attributes == nil {
			continue
		}

		name, _ := instance.Attributes["name"].(string)
		region, _ := instance.Attributes["region"].(string)
		project, _ := instance.Attributes["project"].(string)

		if name == "" || region == "" || project == "" {
			result.Status = "unknown"
			result.Message = "Informations incomplètes dans le tfstate (name, region, project requis)"
			return result, nil
		}

		// Extraire le nom de la région (format: projects/*/regions/us-central1 -> us-central1)
		regionParts := strings.Split(region, "/")
		regionName := regionParts[len(regionParts)-1]

		// Vérifier si l'adresse existe dans GCP
		req := &computepb.GetAddressRequest{
			Project: project,
			Region:  regionName,
			Address: name,
		}

		gcpAddress, err := addressesClient.Get(ctx, req)
		if err != nil {
			result.Status = "missing"
			result.Message = fmt.Sprintf("Adresse %s non trouvée dans GCP: %v", name, err)
			return result, nil
		}

		// Comparer l'adresse IP
		if expectedAddress, ok := instance.Attributes["address"].(string); ok {
			actualAddress := gcpAddress.GetAddress()
			if expectedAddress != actualAddress {
				differences = append(differences, models.DriftDifference{
					Attribute:  "address",
					Expected:   expectedAddress,
					Actual:     actualAddress,
					ChangeType: "modified",
				})
			}
		}
	}

	if len(differences) > 0 {
		result.Status = "drifted"
		result.Message = fmt.Sprintf("%d différence(s) détectée(s)", len(differences))
		result.Differences = differences
	} else {
		result.Status = "in_sync"
		result.Message = "Adresse IP synchronisée avec l'infrastructure GCP"
	}

	return result, nil
}

// buildResourceAddress construit l'adresse complète d'une ressource.
func buildResourceAddress(resource *models.Resource) string {
	var address string
	if resource.Module != "" {
		address = resource.Module + "."
	}
	address += fmt.Sprintf("%s.%s", resource.Type, resource.Name)
	return address
}
