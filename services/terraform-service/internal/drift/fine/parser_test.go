package fine

import (
	"testing"
	"time"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestParsePlan(t *testing.T) {
	now := time.Now()

	plan := &tfjson.Plan{
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "google_compute_instance.web",
				Type:    "google_compute_instance",
				Mode:    tfjson.ManagedResourceMode,
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionNoop},
					Before:  map[string]interface{}{"machine_type": "e2-medium"},
					After:   map[string]interface{}{"machine_type": "e2-medium"},
				},
			},
			{
				Address: "google_compute_instance.api",
				Type:    "google_compute_instance",
				Mode:    tfjson.ManagedResourceMode,
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionUpdate},
					Before: map[string]interface{}{
						"machine_type": "e2-medium",
						"labels":       map[string]interface{}{"env": "prod"},
					},
					After: map[string]interface{}{
						"machine_type": "e2-large",
						"labels":       map[string]interface{}{"env": "prod"},
					},
				},
			},
			{
				Address: "google_compute_address.removed",
				Type:    "google_compute_address",
				Mode:    tfjson.ManagedResourceMode,
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionDelete},
					Before:  map[string]interface{}{"name": "old-ip"},
					After:   nil,
				},
			},
			{
				Address: "google_compute_firewall.new",
				Type:    "google_compute_firewall",
				Mode:    tfjson.ManagedResourceMode,
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionCreate},
					Before:  nil,
					After:   map[string]interface{}{"name": "allow-https"},
				},
			},
			{
				Address: "data.google_project.current",
				Type:    "google_project",
				Mode:    tfjson.DataResourceMode,
				Change: &tfjson.Change{
					Actions: tfjson.Actions{tfjson.ActionRead},
				},
			},
		},
	}

	results := ParsePlan(plan, nil, now)

	if len(results) != 4 {
		t.Fatalf("attendu 4 résultats (data source ignorée), obtenu %d", len(results))
	}

	byAddr := make(map[string]int)
	for i, r := range results {
		byAddr[r.ResourceAddress] = i
		if r.Method != "fine" {
			t.Errorf("%s: Method = %q, attendu \"fine\"", r.ResourceAddress, r.Method)
		}
		if r.DetectedAt.IsZero() {
			t.Errorf("%s: DetectedAt non renseigné", r.ResourceAddress)
		}
	}

	noop := results[byAddr["google_compute_instance.web"]]
	if noop.Status != "in_sync" {
		t.Errorf("no-op: Status = %q, attendu in_sync", noop.Status)
	}

	updated := results[byAddr["google_compute_instance.api"]]
	if updated.Status != "drifted" {
		t.Errorf("update: Status = %q, attendu drifted", updated.Status)
	}
	if len(updated.Differences) != 1 || updated.Differences[0].Attribute != "machine_type" {
		t.Errorf("update: différences inattendues: %+v", updated.Differences)
	}
	if updated.Differences[0].Expected != "e2-medium" || updated.Differences[0].Actual != "e2-large" {
		t.Errorf("update: valeurs avant/après inattendues: %+v", updated.Differences[0])
	}

	deleted := results[byAddr["google_compute_address.removed"]]
	if deleted.Status != "missing" {
		t.Errorf("delete: Status = %q, attendu missing", deleted.Status)
	}

	created := results[byAddr["google_compute_firewall.new"]]
	if created.Status != "drifted" {
		t.Errorf("create: Status = %q, attendu drifted", created.Status)
	}
	if created.Message == "" {
		t.Errorf("create: Message attendu non vide")
	}
}

func TestDiffValuesNested(t *testing.T) {
	before := map[string]interface{}{
		"tags": []interface{}{"a", "b"},
		"meta": map[string]interface{}{
			"size": float64(10),
		},
	}
	after := map[string]interface{}{
		"tags": []interface{}{"a", "c"},
		"meta": map[string]interface{}{
			"size": float64(20),
		},
	}

	diffs := diffValues("google_compute_instance", before, after, "")

	if len(diffs) != 2 {
		t.Fatalf("attendu 2 différences, obtenu %d: %+v", len(diffs), diffs)
	}

	attrs := map[string]bool{}
	for _, d := range diffs {
		attrs[d.Attribute] = true
		if d.ChangeType != "modified" {
			t.Errorf("%s: ChangeType = %q, attendu modified", d.Attribute, d.ChangeType)
		}
	}
	if !attrs["meta.size"] {
		t.Errorf("différence meta.size manquante: %+v", diffs)
	}
	if !attrs["tags[1]"] {
		t.Errorf("différence tags[1] manquante: %+v", diffs)
	}
}

// TestDiffValuesIgnoresContainerClusterNodeConfig vérifie que les blocs
// `node_config`/`node_pool` de premier niveau sur `google_container_cluster`
// sont ignorés, y compris quand leur contenu diffère réellement entre le
// tfstate et le refresh (ex: machine_type) : ils reflètent le pool par défaut
// éphémère créé par remove_default_node_pool=true, pas la configuration
// déclarée du cluster.
func TestDiffValuesIgnoresContainerClusterNodeConfig(t *testing.T) {
	before := map[string]interface{}{
		"name": "demo-kura",
		"node_config": []interface{}{
			map[string]interface{}{"machine_type": "e2-small"},
		},
		"node_pool": []interface{}{
			map[string]interface{}{
				"node_config": []interface{}{
					map[string]interface{}{"machine_type": "e2-small"},
				},
			},
		},
	}
	after := map[string]interface{}{
		"name": "demo-kura",
		"node_config": []interface{}{
			map[string]interface{}{"machine_type": "e2-medium"},
		},
		"node_pool": []interface{}{
			map[string]interface{}{
				"node_config": []interface{}{
					map[string]interface{}{"machine_type": "e2-medium"},
				},
			},
		},
	}

	diffs := diffValues("google_container_cluster", before, after, "")

	if len(diffs) != 0 {
		t.Fatalf("attendu 0 différence (node_config/node_pool ignorés sur google_container_cluster), obtenu %d: %+v", len(diffs), diffs)
	}
}

// TestDiffValuesDoesNotIgnoreNodeConfigForOtherResourceTypes vérifie que le
// filtre node_config/node_pool ne s'applique qu'à google_container_cluster :
// pour les autres types de ressources, une différence sur ces attributs doit
// toujours être signalée.
func TestDiffValuesDoesNotIgnoreNodeConfigForOtherResourceTypes(t *testing.T) {
	before := map[string]interface{}{
		"node_config": []interface{}{
			map[string]interface{}{"machine_type": "e2-small"},
		},
	}
	after := map[string]interface{}{
		"node_config": []interface{}{
			map[string]interface{}{"machine_type": "e2-medium"},
		},
	}

	diffs := diffValues("google_container_node_pool", before, after, "")

	if len(diffs) != 1 {
		t.Fatalf("attendu 1 différence, obtenu %d: %+v", len(diffs), diffs)
	}
}
