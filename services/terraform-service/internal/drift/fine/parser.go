package fine

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/modulops/terraform-service/internal/models"
)

// ParsePlan transforme un plan OpenTofu (refresh-only) en résultats de drift.
//
// stateResources liste les ressources managées du tfstate fourni en entrée :
// `tofu plan -refresh-only` n'inclut dans `ResourceDrift` que les ressources
// pour lesquelles au moins un écart (même filtré ensuite) a été détecté lors
// du refresh ; les ressources strictement identiques au cloud n'y apparaissent
// pas du tout. stateResources permet de générer un résultat "in_sync" pour ces
// ressources absentes, afin que toutes les ressources du state soient reflétées
// dans le résultat.
func ParsePlan(plan *tfjson.Plan, stateResources []models.Resource, now time.Time) []*models.DriftResult {
	if plan == nil {
		return nil
	}

	// `tofu plan -refresh-only` rapporte les écarts entre l'état stocké et
	// l'état réel du cloud dans `ResourceDrift`, et non dans `ResourceChanges`
	// (qui ne reflète que les changements liés à la configuration, vide ici
	// puisqu'aucune action de modification n'est planifiée en mode refresh-only).
	changes := plan.ResourceDrift
	if len(changes) == 0 {
		changes = plan.ResourceChanges
	}

	seen := make(map[string]bool, len(changes))
	results := make([]*models.DriftResult, 0, len(changes)+len(stateResources))
	for _, rc := range changes {
		if rc.Mode == tfjson.DataResourceMode {
			continue
		}
		if rc.Change == nil {
			continue
		}

		actions := rc.Change.Actions
		seen[rc.Address] = true

		switch {
		case actions.NoOp() || actions.Read():
			results = append(results, &models.DriftResult{
				ResourceAddress: rc.Address,
				ResourceType:    rc.Type,
				Status:          "in_sync",
				DetectedAt:      now,
				Method:          "fine",
			})

		case actions.Update() || actions.Replace():
			differences := diffValues(rc.Type, rc.Change.Before, rc.Change.After, "")
			status := "drifted"
			if len(differences) == 0 {
				status = "in_sync"
			}
			results = append(results, &models.DriftResult{
				ResourceAddress: rc.Address,
				ResourceType:    rc.Type,
				Status:          status,
				Differences:     differences,
				DetectedAt:      now,
				Method:          "fine",
			})

		case actions.Delete():
			results = append(results, &models.DriftResult{
				ResourceAddress: rc.Address,
				ResourceType:    rc.Type,
				Status:          "missing",
				Message:         "ressource présente dans le state mais absente du code/cloud (refresh-only)",
				DetectedAt:      now,
				Method:          "fine",
			})

		case actions.Create():
			results = append(results, &models.DriftResult{
				ResourceAddress: rc.Address,
				ResourceType:    rc.Type,
				Status:          "drifted",
				Message:         "ressource déclarée dans le code mais absente du state",
				DetectedAt:      now,
				Method:          "fine",
			})
		}
	}

	// Ressources du state pour lesquelles tofu n'a signalé aucun écart : on
	// les considère "in_sync" (refresh-only n'inclut dans ResourceDrift que
	// les ressources ayant au moins une différence détectée).
	for _, res := range stateResources {
		if res.Mode == string(tfjson.DataResourceMode) {
			continue
		}
		address := res.Type + "." + res.Name
		if res.Module != "" {
			address = res.Module + "." + address
		}
		if seen[address] {
			continue
		}
		results = append(results, &models.DriftResult{
			ResourceAddress: address,
			ResourceType:    res.Type,
			Status:          "in_sync",
			DetectedAt:      now,
			Method:          "fine",
		})
	}

	return results
}

// ignoredTopLevelAttributes liste, par type de ressource, les attributs de
// premier niveau à exclure de la comparaison de drift.
//
// `google_container_cluster` expose dans son state les blocs `node_config`
// et `node_pool`, qui reflètent le pool de nœuds par défaut éphémère créé
// puis détruit par GKE lorsque `remove_default_node_pool = true` est utilisé
// (pattern requis pour gérer les node pools séparément via
// `google_container_node_pool`). Ces blocs ne correspondent à aucun attribut
// déclaré dans la configuration `.tf` de `google_container_cluster` : leur
// contenu (ex: `machine_type`) est un artefact figé au moment de la création
// du cluster et ne reflète jamais les node pools gérés séparément. Les
// ignorer ici évite un faux positif de dérive permanent ; la dérive réelle
// des node pools est déjà détectée via `google_container_node_pool`.
var ignoredTopLevelAttributes = map[string]map[string]bool{
	"google_container_cluster": {
		"node_config": true,
		"node_pool":   true,
	},
}

// diffValues compare deux valeurs JSON décodées (map/slice/scalaires) et
// retourne la liste aplatie des différences sous forme de chemins en notation pointée.
func diffValues(resourceType string, before, after interface{}, prefix string) []DriftDifferenceList {
	var diffs []DriftDifferenceList

	beforeMap, beforeIsMap := before.(map[string]interface{})
	afterMap, afterIsMap := after.(map[string]interface{})
	if ignored := ignoredTopLevelAttributes[resourceType]; beforeIsMap && afterIsMap && prefix == "" && ignored != nil {
		filteredBefore := make(map[string]interface{}, len(beforeMap))
		for k, v := range beforeMap {
			if !ignored[k] {
				filteredBefore[k] = v
			}
		}
		filteredAfter := make(map[string]interface{}, len(afterMap))
		for k, v := range afterMap {
			if !ignored[k] {
				filteredAfter[k] = v
			}
		}
		diffValuesInto(filteredBefore, filteredAfter, prefix, &diffs)
	} else {
		diffValuesInto(before, after, prefix, &diffs)
	}

	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Attribute < diffs[j].Attribute })
	return diffs
}

// DriftDifferenceList est un alias pour models.DriftDifference, utilisé pour
// limiter la portée des imports dans ce fichier.
type DriftDifferenceList = models.DriftDifference

func diffValuesInto(before, after interface{}, prefix string, out *[]DriftDifferenceList) {
	if reflect.DeepEqual(before, after) {
		return
	}

	// Ignore les écarts entre `null` et une valeur "vide" du même type
	// (slice/map/string/bool/number à zéro). Ce motif apparaît typiquement
	// quand le provider ajoute un nouvel attribut calculé/optionnel après
	// l'écriture du tfstate : le state stocke `null` pour cet attribut,
	// alors que le refresh renvoie sa valeur vide par défaut. Ce n'est pas
	// une dérive réelle de l'infrastructure.
	if isNullOrEmpty(before) && isNullOrEmpty(after) {
		return
	}

	beforeMap, beforeIsMap := before.(map[string]interface{})
	afterMap, afterIsMap := after.(map[string]interface{})

	if beforeIsMap && afterIsMap {
		keys := make(map[string]struct{}, len(beforeMap)+len(afterMap))
		for k := range beforeMap {
			keys[k] = struct{}{}
		}
		for k := range afterMap {
			keys[k] = struct{}{}
		}
		for k := range keys {
			path := k
			if prefix != "" {
				path = prefix + "." + k
			}
			bv, bok := beforeMap[k]
			av, aok := afterMap[k]
			switch {
			case bok && !aok:
				if !isNullOrEmpty(bv) {
					*out = append(*out, DriftDifferenceList{Attribute: path, Expected: bv, Actual: nil, ChangeType: "removed"})
				}
			case !bok && aok:
				if !isNullOrEmpty(av) && !isWholeBlockAddition(path, av) {
					*out = append(*out, DriftDifferenceList{Attribute: path, Expected: nil, Actual: av, ChangeType: "added"})
				}
			default:
				diffValuesInto(bv, av, path, out)
			}
		}
		return
	}

	beforeSlice, beforeIsSlice := before.([]interface{})
	afterSlice, afterIsSlice := after.([]interface{})
	if beforeIsSlice && afterIsSlice {
		max := len(beforeSlice)
		if len(afterSlice) > max {
			max = len(afterSlice)
		}
		for i := 0; i < max; i++ {
			path := fmt.Sprintf("%s[%d]", prefix, i)
			var bv, av interface{}
			var bok, aok bool
			if i < len(beforeSlice) {
				bv, bok = beforeSlice[i], true
			}
			if i < len(afterSlice) {
				av, aok = afterSlice[i], true
			}
			switch {
			case bok && !aok:
				if !isNullOrEmpty(bv) {
					*out = append(*out, DriftDifferenceList{Attribute: path, Expected: bv, Actual: nil, ChangeType: "removed"})
				}
			case !bok && aok:
				if !isNullOrEmpty(av) && !isWholeBlockAddition(path, av) {
					*out = append(*out, DriftDifferenceList{Attribute: path, Expected: nil, Actual: av, ChangeType: "added"})
				}
			default:
				diffValuesInto(bv, av, path, out)
			}
		}
		return
	}

	// Ignore l'ajout d'un bloc entier (avant=null, après=map/slice non vide)
	// sur un chemin qui désigne un bloc complet (terminant par "[N]"), motif
	// caractéristique d'un attribut calculé nouvellement exposé par le
	// schéma du provider plutôt qu'une dérive réelle.
	if before == nil && isWholeBlockAddition(prefix, after) {
		return
	}

	// Valeurs scalaires différentes (ou types incompatibles).
	*out = append(*out, DriftDifferenceList{Attribute: prefix, Expected: before, Actual: after, ChangeType: "modified"})
}

// isWholeBlockAddition indique si path désigne un bloc complet (notation
// "...[N]", c'est-à-dire un élément de liste/bloc Terraform plutôt qu'un
// champ scalaire individuel) et si val est un objet (map) ou une liste non
// vide. Combiné à `expected: null`, ce motif correspond typiquement à un
// attribut calculé/optionnel ajouté par une nouvelle version du provider,
// absent du schéma au moment de l'écriture du tfstate — pas une dérive
// réelle de l'infrastructure.
func isWholeBlockAddition(path string, val interface{}) bool {
	if !strings.HasSuffix(path, "]") || !strings.Contains(path, "[") {
		return false
	}
	switch val.(type) {
	case map[string]interface{}, []interface{}:
		return true
	default:
		return false
	}
}

// isNullOrEmpty indique si v est `nil`, une valeur "vide" de son type
// (chaîne vide, false, 0), ou une slice/map dont tous les éléments sont
// eux-mêmes "vides" au sens de cette fonction (récursif). Cela permet
// d'ignorer des blocs entiers (ex: node_config[0]) lorsque tous leurs
// attributs internes sont à leur valeur zéro.
func isNullOrEmpty(v interface{}) bool {
	switch val := v.(type) {
	case nil:
		return true
	case string:
		return val == ""
	case bool:
		return !val
	case float64:
		return val == 0
	case []interface{}:
		for _, e := range val {
			if !isNullOrEmpty(e) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		for _, e := range val {
			if !isNullOrEmpty(e) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
