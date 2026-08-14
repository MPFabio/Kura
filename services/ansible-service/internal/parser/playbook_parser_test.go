package parser

import "testing"

const samplePlaybook = `
- name: Déploiement Kura
  hosts: web
  become: true
  collections:
    - community.general
  vars:
    app_port: 8080
  tasks:
    - name: Installer nginx
      apt:
        name: nginx
        state: present
    - name: Copier la conf
      template:
        src: kura.conf.j2
        dest: /etc/nginx/conf.d/kura.conf
      notify:
        - reload nginx
  handlers:
    - name: reload nginx
      service:
        name: nginx
        state: reloaded
- name: Post-checks
  hosts: monitoring
  tasks:
    - name: Vérifier le endpoint
      uri:
        url: http://localhost:8080/health
`

func TestAnalyzePlaybookStatistics(t *testing.T) {
	result := AnalyzePlaybook(samplePlaybook)
	if _, hasError := result["error"]; hasError {
		t.Fatalf("analyse en erreur: %v", result["error"])
	}

	stats, ok := result["statistics"].(map[string]any)
	if !ok {
		t.Fatal("statistics manquantes")
	}
	if stats["total_plays"] != 2 {
		t.Errorf("total_plays = %v, attendu 2", stats["total_plays"])
	}
	if stats["total_tasks"] != 3 {
		t.Errorf("total_tasks = %v, attendu 3", stats["total_tasks"])
	}
	if stats["total_handlers"] != 1 {
		t.Errorf("total_handlers = %v, attendu 1", stats["total_handlers"])
	}
	if stats["become_used"] != true {
		t.Error("become_used doit être vrai")
	}

	modules, _ := stats["modules_used"].([]string)
	expected := []string{"apt", "service", "template", "uri"}
	if len(modules) != len(expected) {
		t.Fatalf("modules_used = %v, attendu %v", modules, expected)
	}
	for i, m := range expected {
		if modules[i] != m {
			t.Errorf("modules_used[%d] = %s, attendu %s", i, modules[i], m)
		}
	}

	hosts, _ := stats["hosts_targeted"].([]string)
	if len(hosts) != 2 || hosts[0] != "monitoring" || hosts[1] != "web" {
		t.Errorf("hosts_targeted = %v", hosts)
	}
}

func TestAnalyzePlaybookInvalidYAML(t *testing.T) {
	result := AnalyzePlaybook("{{invalide")
	if _, hasError := result["error"]; !hasError {
		t.Error("un YAML invalide doit produire une erreur")
	}
}

func TestParsePlaybookSinglePlay(t *testing.T) {
	parsed := ParsePlaybook("name: solo\nhosts: all\ntasks: []\n")
	if parsed == nil {
		t.Fatal("un play unique (non-liste) doit être accepté")
	}
	if parsed["play_count"] != 1 {
		t.Errorf("play_count = %v, attendu 1", parsed["play_count"])
	}
}

func TestParseTaskBlock(t *testing.T) {
	content := `
- hosts: all
  tasks:
    - name: bloc de secours
      block:
        - name: action risquée
          command: /bin/false
      rescue:
        - name: rattrapage
          debug:
            msg: recovered
`
	parsed := ParsePlaybook(content)
	plays := parsed["plays"].([]map[string]any)
	tasks := plays[0]["tasks"].([]map[string]any)
	if len(tasks) != 1 || tasks[0]["type"] != "block" {
		t.Fatalf("le block doit être typé block: %+v", tasks[0])
	}
	blockTasks := tasks[0]["block"].([]map[string]any)
	rescueTasks := tasks[0]["rescue"].([]map[string]any)
	if len(blockTasks) != 1 || len(rescueTasks) != 1 {
		t.Errorf("block=%d rescue=%d, attendu 1 et 1", len(blockTasks), len(rescueTasks))
	}
}
