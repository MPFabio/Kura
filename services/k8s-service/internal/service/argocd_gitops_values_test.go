package service

import (
	"encoding/json"
	"testing"
)

// Les values d'une Application engendrée par Kura vivent dans le dépôt GitOps.
// Ces tests couvrent la localisation de ce fichier : s'y tromper ferait écrire
// les values en ligne dans l'Application, et le dépôt ne refléterait plus
// l'état déployé.

func specFromJSON(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &spec); err != nil {
		t.Fatalf("spec de test invalide: %v", err)
	}
	return spec
}

func TestGitOpsValuesPathOnMultiSourceApplication(t *testing.T) {
	spec := specFromJSON(t, `{
		"sources": [
			{"chart":"zot","repoURL":"https://zotregistry.dev/helm-charts/","helm":{"valueFiles":["$values/apps/zot/values.yaml"]}},
			{"repoURL":"https://codeberg.org/MPFabio/Kuro.git","targetRevision":"gitops","ref":"values","path":"."}
		]
	}`)

	if got := gitOpsValuesPath(spec); got != "apps/zot/values.yaml" {
		t.Fatalf("chemin attendu apps/zot/values.yaml, obtenu %q", got)
	}
	if got := gitOpsValuesBranch(spec); got != "gitops" {
		t.Fatalf("branche attendue gitops, obtenue %q", got)
	}
}

// Une Application mono-source n'a pas de fichier de values versionné : la mise
// à jour doit alors retomber sur les values en ligne, et non échouer.
func TestGitOpsValuesPathAbsentOnSingleSource(t *testing.T) {
	spec := specFromJSON(t, `{
		"source": {"chart":"zot","repoURL":"https://example.org/charts","helm":{"values":"persistence: true"}}
	}`)

	if got := gitOpsValuesPath(spec); got != "" {
		t.Fatalf("aucun chemin attendu, obtenu %q", got)
	}
}

// Un fichier de values qui ne passe pas par la source « $values » n'est pas
// géré par Kura : le modifier reviendrait à écrire dans un dépôt tiers.
func TestGitOpsValuesPathIgnoresForeignValueFiles(t *testing.T) {
	spec := specFromJSON(t, `{
		"sources": [
			{"chart":"zot","helm":{"valueFiles":["values-production.yaml"]}}
		]
	}`)

	if got := gitOpsValuesPath(spec); got != "" {
		t.Fatalf("aucun chemin attendu pour un fichier hors $values, obtenu %q", got)
	}
}

func TestGitOpsValuesBranchAbsentWithoutValuesRef(t *testing.T) {
	spec := specFromJSON(t, `{
		"sources": [
			{"chart":"zot","helm":{"valueFiles":["$values/apps/zot/values.yaml"]}},
			{"repoURL":"https://codeberg.org/MPFabio/Kuro.git","targetRevision":"main","path":"."}
		]
	}`)

	if got := gitOpsValuesBranch(spec); got != "" {
		t.Fatalf("aucune branche attendue sans source ref=values, obtenue %q", got)
	}
}

// helmSourceOf doit retenir la source portant le chart, et non celle qui ne
// sert qu'à référencer le dépôt de values.
func TestHelmSourceOfPrefersChartSource(t *testing.T) {
	spec := specFromJSON(t, `{
		"sources": [
			{"repoURL":"https://codeberg.org/MPFabio/Kuro.git","ref":"values","path":"."},
			{"chart":"zot","repoURL":"https://zotregistry.dev/helm-charts/"}
		]
	}`)

	source, err := helmSourceOf(spec)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if chart, _ := source["chart"].(string); chart != "zot" {
		t.Fatalf("source du chart attendue, obtenue %v", source)
	}
}

func TestHelmSourceOfFailsWithoutAnySource(t *testing.T) {
	if _, err := helmSourceOf(map[string]interface{}{}); err == nil {
		t.Fatal("une spec sans source doit produire une erreur explicite")
	}
}
