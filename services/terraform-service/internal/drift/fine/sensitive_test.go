package fine

import "testing"

// Un état Terraform contient couramment des secrets en clair. Ces tests
// couvrent leur masquage : une régression ici exposerait des mots de passe ou
// des clés privées dans l'interface, à quiconque peut consulter une détection
// de dérive.

func TestMaskingByAttributeName(t *testing.T) {
	diffs := []DriftDifferenceList{
		{Attribute: "admin_password", Expected: "ancien", Actual: "nouveau"},
		{Attribute: "service_account.private_key", Expected: "-----BEGIN", Actual: "-----BEGIN"},
		{Attribute: "api_token", Expected: "tok-1", Actual: "tok-2"},
		{Attribute: "machine_type", Expected: "e2-small", Actual: "e2-medium"},
	}

	maskSensitiveDifferences(diffs, nil, nil)

	for _, d := range diffs[:3] {
		if d.Expected != maskedValue || d.Actual != maskedValue {
			t.Errorf("%s devrait être masqué, obtenu %v / %v", d.Attribute, d.Expected, d.Actual)
		}
	}
	// Un attribut ordinaire doit rester lisible : masquer trop large rendrait
	// la détection inutilisable.
	if diffs[3].Expected != "e2-small" || diffs[3].Actual != "e2-medium" {
		t.Errorf("machine_type ne doit pas être masqué, obtenu %v / %v", diffs[3].Expected, diffs[3].Actual)
	}
}

func TestMaskingFromPlanMarkers(t *testing.T) {
	// Le fournisseur marque « result » comme sensible côté état réel seulement.
	afterSensitive := map[string]interface{}{"result": true}

	diffs := []DriftDifferenceList{
		{Attribute: "result", Expected: "visible", Actual: "à masquer"},
	}
	maskSensitiveDifferences(diffs, nil, afterSensitive)

	if diffs[0].Actual != maskedValue {
		t.Errorf("la valeur constatée devait être masquée, obtenu %v", diffs[0].Actual)
	}
	if diffs[0].Expected != "visible" {
		t.Errorf("la valeur attendue n'était pas marquée sensible, obtenu %v", diffs[0].Expected)
	}
}

func TestMaskingOnNestedAndIndexedPaths(t *testing.T) {
	sensitive := map[string]interface{}{
		"network_interface": []interface{}{
			map[string]interface{}{"access_config": []interface{}{
				map[string]interface{}{"nat_ip": true},
			}},
		},
	}

	diffs := []DriftDifferenceList{
		{Attribute: "network_interface[0].access_config[0].nat_ip", Expected: "1.2.3.4", Actual: "5.6.7.8"},
	}
	maskSensitiveDifferences(diffs, sensitive, sensitive)

	if diffs[0].Expected != maskedValue || diffs[0].Actual != maskedValue {
		t.Errorf("chemin imbriqué non masqué: %v / %v", diffs[0].Expected, diffs[0].Actual)
	}
}

// Un marquage posé sur un niveau intermédiaire vaut pour tout ce qu'il
// contient : c'est ainsi que Terraform signale un bloc entièrement sensible.
func TestMaskingInheritedFromParentNode(t *testing.T) {
	sensitive := map[string]interface{}{"metadata": true}

	diffs := []DriftDifferenceList{
		{Attribute: "metadata.startup_script", Expected: "a", Actual: "b"},
	}
	maskSensitiveDifferences(diffs, sensitive, sensitive)

	if diffs[0].Expected != maskedValue {
		t.Errorf("le marquage du parent devait s'appliquer, obtenu %v", diffs[0].Expected)
	}
}

func TestUnmarkedAttributeStaysVisible(t *testing.T) {
	sensitive := map[string]interface{}{"autre": true}

	diffs := []DriftDifferenceList{
		{Attribute: "tags", Expected: "prod", Actual: "staging"},
	}
	maskSensitiveDifferences(diffs, sensitive, sensitive)

	if diffs[0].Expected != "prod" || diffs[0].Actual != "staging" {
		t.Errorf("attribut non marqué masqué à tort: %v / %v", diffs[0].Expected, diffs[0].Actual)
	}
}

func TestSplitAttributePath(t *testing.T) {
	segments := splitAttributePath("network_interface[0].access_config[1].nat_ip")

	if len(segments) != 5 {
		t.Fatalf("5 segments attendus, obtenu %d: %+v", len(segments), segments)
	}
	if segments[0].key != "network_interface" {
		t.Errorf("premier segment: %q", segments[0].key)
	}
	if segments[1].index != 0 {
		t.Errorf("indice attendu 0, obtenu %d", segments[1].index)
	}
	if segments[3].index != 1 {
		t.Errorf("indice attendu 1, obtenu %d", segments[3].index)
	}
	if segments[4].key != "nat_ip" {
		t.Errorf("dernier segment: %q", segments[4].key)
	}
}
