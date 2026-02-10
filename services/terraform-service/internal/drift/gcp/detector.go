package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/modulops/terraform-service/internal/models"
	"google.golang.org/api/iterator"
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
	case strings.HasPrefix(resource.Type, "google_compute_subnetwork"):
		return d.detectComputeSubnetworkDrift(ctx, resource, result)
	case strings.HasPrefix(resource.Type, "google_container_cluster"):
		return d.detectContainerClusterDrift(ctx, resource, result)
	case strings.HasPrefix(resource.Type, "google_container_node_pool"):
		return d.detectContainerNodePoolDrift(ctx, resource, result)
	default:
		// Détection générique via l'API Cloud Asset Inventory pour toutes les ressources GCP
		return d.detectGenericDriftViaAssetAPI(ctx, resource, result)
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

// detectComputeSubnetworkDrift détecte les dérives pour un sous-réseau Compute.
func (d *Detector) detectComputeSubnetworkDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}

	subnetworksClient, err := compute.NewSubnetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client Subnetworks: %w", err)
	}
	defer subnetworksClient.Close()

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
	region, _ := instance.Attributes["region"].(string)
	project, _ := instance.Attributes["project"].(string)

	if name == "" || region == "" || project == "" {
		result.Status = "unknown"
		result.Message = "Informations incomplètes dans le tfstate (name, region, project requis)"
		return result, nil
	}

	regionParts := strings.Split(region, "/")
	regionName := regionParts[len(regionParts)-1]

	req := &computepb.GetSubnetworkRequest{
		Project:    project,
		Region:     regionName,
		Subnetwork: name,
	}

	_, err = subnetworksClient.Get(ctx, req)
	if err != nil {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Sous-réseau %s non trouvé dans GCP: %v", name, err)
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Sous-réseau synchronisé avec l'infrastructure GCP"
	return result, nil
}

// detectContainerClusterDrift détecte les dérives pour un cluster GKE.
func (d *Detector) detectContainerClusterDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}

	clusterManagerClient, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client ClusterManager: %w", err)
	}
	defer clusterManagerClient.Close()

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
	location, _ := instance.Attributes["location"].(string)
	zone, _ := instance.Attributes["zone"].(string)
	if location == "" {
		location = zone
	}
	if location != "" {
		locParts := strings.Split(location, "/")
		location = locParts[len(locParts)-1]
	}

	if name == "" || project == "" || location == "" {
		result.Status = "unknown"
		result.Message = "Informations incomplètes dans le tfstate (name, project, location ou zone requis)"
		return result, nil
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
	req := &containerpb.GetClusterRequest{
		Name: parent,
	}

	_, err = clusterManagerClient.GetCluster(ctx, req)
	if err != nil {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Cluster GKE %s non trouvé dans GCP: %v", name, err)
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Cluster GKE synchronisé avec l'infrastructure GCP"
	return result, nil
}

// detectContainerNodePoolDrift détecte les dérives pour un node pool GKE.
func (d *Detector) detectContainerNodePoolDrift(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}

	clusterManagerClient, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création du client ClusterManager: %w", err)
	}
	defer clusterManagerClient.Close()

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
	cluster, _ := instance.Attributes["cluster"].(string)
	location, _ := instance.Attributes["location"].(string)
	zone, _ := instance.Attributes["zone"].(string)
	if location == "" {
		location = zone
	}
	if location != "" {
		locParts := strings.Split(location, "/")
		location = locParts[len(locParts)-1]
	}
	if cluster != "" {
		clusterParts := strings.Split(cluster, "/")
		cluster = clusterParts[len(clusterParts)-1]
	}

	if name == "" || project == "" || location == "" || cluster == "" {
		result.Status = "unknown"
		result.Message = "Informations incomplètes dans le tfstate (name, project, location/zone, cluster requis)"
		return result, nil
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, name)
	req := &containerpb.GetNodePoolRequest{
		Name: parent,
	}

	_, err = clusterManagerClient.GetNodePool(ctx, req)
	if err != nil {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Node pool %s non trouvé dans GCP: %v", name, err)
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Node pool GKE synchronisé avec l'infrastructure GCP"
	return result, nil
}

// terraformTypeToAssetType convertit un type de ressource Terraform (ex: google_compute_network)
// en type d'asset GCP (ex: compute.googleapis.com/Network).
func terraformTypeToAssetType(terraformType string) string {
	parts := strings.Split(terraformType, "_")
	if len(parts) < 3 {
		return ""
	}
	// google_compute_network -> service=compute, typeName=Network
	// google_container_node_pool -> service=container, typeName=NodePool
	service := parts[1]
	typeParts := parts[2:]
	var b strings.Builder
	for _, p := range typeParts {
		if len(p) == 0 {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	typeName := b.String()
	if typeName == "" {
		return ""
	}
	return service + ".googleapis.com/" + typeName
}

// getResourceNameFromInstance extrait le nom de la ressource côté GCP depuis les attributs tfstate.
func getResourceNameFromInstance(attrs map[string]interface{}) string {
	if name, ok := attrs["name"].(string); ok && name != "" {
		return name
	}
	if id, ok := attrs["id"].(string); ok && id != "" {
		return id
	}
	if selfLink, ok := attrs["self_link"].(string); ok && selfLink != "" {
		parts := strings.Split(selfLink, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	return ""
}

// getProjectFromResource extrait le projet GCP depuis la première instance de la ressource.
func getProjectFromResource(resource *models.Resource) string {
	for _, inst := range resource.Instances {
		if inst.Attributes == nil {
			continue
		}
		if p, ok := inst.Attributes["project"].(string); ok && p != "" {
			return p
		}
	}
	return ""
}

// lastSegment retourne le dernier segment d'un chemin (ex: //compute.../networks/foo -> foo).
func lastSegment(path string) string {
	path = strings.TrimPrefix(path, "//")
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}

// detectGenericDriftViaAssetAPI vérifie l'existence des ressources via l'API Cloud Asset Inventory.
// Permet une détection de drift concrète pour tous les types de ressources GCP sans détecteur dédié.
func (d *Detector) detectGenericDriftViaAssetAPI(ctx context.Context, resource *models.Resource, result *models.DriftResult) (*models.DriftResult, error) {
	assetType := terraformTypeToAssetType(resource.Type)
	if assetType == "" {
		result.Status = "unknown"
		result.Message = "Impossible de mapper le type Terraform vers un type d'asset GCP"
		return result, nil
	}

	project := getProjectFromResource(resource)
	if project == "" {
		result.Status = "unknown"
		result.Message = "Attribut project manquant dans le tfstate"
		return result, nil
	}

	var opts []option.ClientOption
	if d.credentialsJSON != "" {
		opts = append(opts, option.WithCredentialsJSON([]byte(d.credentialsJSON)))
	}

	client, err := asset.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("création du client Asset: %w", err)
	}
	defer client.Close()

	parent := "projects/" + project
	req := &assetpb.ListAssetsRequest{
		Parent:      parent,
		AssetTypes:  []string{assetType},
		ContentType: assetpb.ContentType_RESOURCE,
	}

	it := client.ListAssets(ctx, req)
	assetNames := make(map[string]bool)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			result.Status = "unknown"
			result.Message = fmt.Sprintf("Erreur API Asset Inventory: %v", err)
			return result, nil
		}
		if resp != nil && resp.Name != "" {
			assetNames[lastSegment(resp.Name)] = true
			assetNames[resp.Name] = true
		}
	}

	missing := make([]string, 0)
	for _, instance := range resource.Instances {
		if instance.Attributes == nil {
			continue
		}
		expectedName := getResourceNameFromInstance(instance.Attributes)
		if expectedName == "" {
			missing = append(missing, "(nom non trouvé dans le tfstate)")
			continue
		}
		if !assetNames[expectedName] {
			missing = append(missing, expectedName)
		}
	}

	if len(missing) > 0 {
		result.Status = "missing"
		result.Message = fmt.Sprintf("Ressource(s) absente(s) dans GCP (Asset Inventory): %s", strings.Join(missing, ", "))
		return result, nil
	}

	result.Status = "in_sync"
	result.Message = "Ressource synchronisée avec l'infrastructure GCP (vérification via Asset Inventory)"
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
