// Package parser analyse les playbooks Ansible (YAML) : structure des plays,
// tâches, rôles, et statistiques agrégées.
package parser

import (
	"sort"

	"gopkg.in/yaml.v3"
)

// taskReservedKeys sont les clés d'une tâche qui ne désignent pas un module.
var taskReservedKeys = map[string]bool{
	"name": true, "when": true, "loop": true, "until": true, "retries": true,
	"delay": true, "register": true, "changed_when": true, "failed_when": true,
	"ignore_errors": true, "async": true, "poll": true, "delegate_to": true,
	"delegate_facts": true, "become": true, "become_user": true,
	"environment": true, "tags": true, "notify": true, "action": true,
	"local_action": true, "with_": true, "vars": true, "include": true,
	"include_tasks": true, "import_tasks": true, "block": true, "rescue": true,
	"always": true, "role": true, "include_role": true, "import_role": true,
}

// ParsePlaybook parse le contenu YAML d'un playbook. Retourne nil si le
// contenu est vide ou invalide.
func ParsePlaybook(content string) map[string]any {
	var root any
	if err := yaml.Unmarshal([]byte(content), &root); err != nil || root == nil {
		return nil
	}

	plays, ok := root.([]any)
	if !ok {
		plays = []any{root}
	}

	parsedPlays := []map[string]any{}
	for _, play := range plays {
		if parsed := parsePlay(play); parsed != nil {
			parsedPlays = append(parsedPlays, parsed)
		}
	}

	return map[string]any{
		"plays":      parsedPlays,
		"play_count": len(parsedPlays),
	}
}

func parsePlay(raw any) map[string]any {
	play, ok := toStringMap(raw)
	if !ok {
		return nil
	}

	parsed := map[string]any{
		"name":          getOr(play, "name", "Unnamed play"),
		"hosts":         getOr(play, "hosts", "all"),
		"gather_facts":  getOr(play, "gather_facts", true),
		"vars":          getOr(play, "vars", map[string]any{}),
		"vars_files":    getOr(play, "vars_files", []any{}),
		"vars_prompt":   getOr(play, "vars_prompt", []any{}),
		"pre_tasks":     []map[string]any{},
		"tasks":         []map[string]any{},
		"handlers":      []map[string]any{},
		"post_tasks":    []map[string]any{},
		"roles":         []map[string]any{},
		"collections":   getOr(play, "collections", []any{}),
		"become":        getOr(play, "become", false),
		"become_user":   play["become_user"],
		"become_method": play["become_method"],
		"environment":   getOr(play, "environment", map[string]any{}),
	}

	for _, section := range []string{"pre_tasks", "tasks", "handlers", "post_tasks"} {
		if raw, found := play[section]; found {
			parsed[section] = parseTasks(raw)
		}
	}
	if roles, found := play["roles"]; found {
		parsed["roles"] = parseRoles(roles)
	}

	return parsed
}

func parseTasks(raw any) []map[string]any {
	list, ok := raw.([]any)
	if !ok {
		return []map[string]any{}
	}
	parsed := []map[string]any{}
	for _, item := range list {
		switch task := item.(type) {
		case map[string]any:
			if p := parseTask(task); p != nil {
				parsed = append(parsed, p)
			}
		case string:
			parsed = append(parsed, map[string]any{"name": task, "type": "inline"})
		}
	}
	return parsed
}

func parseTask(task map[string]any) map[string]any {
	parsed := map[string]any{
		"name":           getOr(task, "name", "Unnamed task"),
		"module":         nil,
		"args":           map[string]any{},
		"when":           task["when"],
		"loop":           task["loop"],
		"until":          task["until"],
		"retries":        task["retries"],
		"delay":          task["delay"],
		"register":       task["register"],
		"changed_when":   task["changed_when"],
		"failed_when":    task["failed_when"],
		"ignore_errors":  getOr(task, "ignore_errors", false),
		"async_val":      task["async"],
		"poll":           task["poll"],
		"delegate_to":    task["delegate_to"],
		"delegate_facts": getOr(task, "delegate_facts", false),
		"become":         task["become"],
		"become_user":    task["become_user"],
		"environment":    getOr(task, "environment", map[string]any{}),
		"tags":           getOr(task, "tags", []any{}),
		"notify":         getOr(task, "notify", []any{}),
	}

	// Identifier le module : première clé non réservée, sinon "action".
	moduleKeys := []string{}
	for key := range task {
		if !taskReservedKeys[key] {
			moduleKeys = append(moduleKeys, key)
		}
	}
	sort.Strings(moduleKeys)

	if len(moduleKeys) > 0 {
		parsed["module"] = moduleKeys[0]
		parsed["args"] = getOr(task, moduleKeys[0], map[string]any{})
	} else if action, found := task["action"]; found {
		switch a := action.(type) {
		case map[string]any:
			for key, value := range a {
				parsed["module"] = key
				parsed["args"] = value
				break
			}
		case string:
			if fields := splitFirst(a); fields != "" {
				parsed["module"] = fields
			}
		}
	}

	// Includes, imports, rôles et blocs.
	switch {
	case has(task, "include"):
		parsed["type"] = "include"
		parsed["include"] = task["include"]
	case has(task, "include_tasks"):
		parsed["type"] = "include_tasks"
		parsed["include_tasks"] = task["include_tasks"]
	case has(task, "import_tasks"):
		parsed["type"] = "import_tasks"
		parsed["import_tasks"] = task["import_tasks"]
	case has(task, "include_role"):
		parsed["type"] = "include_role"
		parsed["include_role"] = task["include_role"]
	case has(task, "import_role"):
		parsed["type"] = "import_role"
		parsed["import_role"] = task["import_role"]
	case has(task, "role"):
		parsed["type"] = "role"
		parsed["role"] = task["role"]
	case has(task, "block"):
		parsed["type"] = "block"
		parsed["block"] = parseTasks(task["block"])
		parsed["rescue"] = parseTasks(task["rescue"])
		parsed["always"] = parseTasks(task["always"])
	default:
		parsed["type"] = "task"
	}

	return parsed
}

func parseRoles(raw any) []map[string]any {
	list, ok := raw.([]any)
	if !ok {
		return []map[string]any{}
	}
	parsed := []map[string]any{}
	for _, item := range list {
		switch role := item.(type) {
		case string:
			parsed = append(parsed, map[string]any{"name": role, "vars": map[string]any{}})
		case map[string]any:
			name := str(role["role"])
			if name == "" {
				name = str(role["name"])
			}
			if name == "" {
				name = "unknown"
			}
			vars := map[string]any{}
			for key, value := range role {
				if key != "role" && key != "name" {
					vars[key] = value
				}
			}
			parsed = append(parsed, map[string]any{"name": name, "vars": vars})
		}
	}
	return parsed
}

// AnalyzePlaybook produit l'analyse complète d'un playbook avec statistiques.
func AnalyzePlaybook(content string) map[string]any {
	parsed := ParsePlaybook(content)
	if parsed == nil {
		return map[string]any{"error": "Impossible de parser le playbook"}
	}

	totalTasks, totalHandlers, totalPreTasks, totalPostTasks, totalRoles := 0, 0, 0, 0, 0
	modulesUsed := map[string]bool{}
	hostsTargeted := map[string]bool{}
	collectionsUsed := map[string]bool{}
	becomeUsed := false

	plays, _ := parsed["plays"].([]map[string]any)
	for _, play := range plays {
		tasks := taskList(play["tasks"])
		handlers := taskList(play["handlers"])
		preTasks := taskList(play["pre_tasks"])
		postTasks := taskList(play["post_tasks"])

		totalTasks += len(tasks)
		totalHandlers += len(handlers)
		totalPreTasks += len(preTasks)
		totalPostTasks += len(postTasks)
		if roles, ok := play["roles"].([]map[string]any); ok {
			totalRoles += len(roles)
		}

		switch hosts := play["hosts"].(type) {
		case string:
			hostsTargeted[hosts] = true
		case []any:
			for _, h := range hosts {
				if hs, ok := h.(string); ok {
					hostsTargeted[hs] = true
				}
			}
		}

		if become, ok := play["become"].(bool); ok && become {
			becomeUsed = true
		}
		if collections, ok := play["collections"].([]any); ok {
			for _, c := range collections {
				if cs, ok := c.(string); ok {
					collectionsUsed[cs] = true
				}
			}
		}

		for _, list := range [][]map[string]any{tasks, handlers, preTasks, postTasks} {
			for _, task := range list {
				if module, ok := task["module"].(string); ok && module != "" {
					modulesUsed[module] = true
				}
			}
		}
	}

	return map[string]any{
		"parsed": parsed,
		"statistics": map[string]any{
			"total_plays":      parsed["play_count"],
			"total_tasks":      totalTasks,
			"total_handlers":   totalHandlers,
			"total_pre_tasks":  totalPreTasks,
			"total_post_tasks": totalPostTasks,
			"total_roles":      totalRoles,
			"modules_used":     sortedKeys(modulesUsed),
			"hosts_targeted":   sortedKeys(hostsTargeted),
			"become_used":      becomeUsed,
			"collections_used": sortedKeys(collectionsUsed),
		},
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// toStringMap convertit les maps YAML (clés any) en map[string]any.
func toStringMap(raw any) (map[string]any, bool) {
	if m, ok := raw.(map[string]any); ok {
		return m, true
	}
	return nil, false
}

func getOr(m map[string]any, key string, fallback any) any {
	if v, found := m[key]; found && v != nil {
		return v
	}
	return fallback
}

func has(m map[string]any, key string) bool {
	_, found := m[key]
	return found
}

func taskList(v any) []map[string]any {
	if list, ok := v.([]map[string]any); ok {
		return list
	}
	return nil
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func splitFirst(s string) string {
	for i, r := range s {
		if r == ' ' {
			return s[:i]
		}
	}
	return s
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
