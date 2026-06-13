package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(server *httptest.Server) *GitHubClient {
	c := NewGitHubClient("test-token")
	c.apiBase = server.URL
	return c
}

func encodeB64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// TestFetchTFFiles_SingleFile teste la récupération d'un unique fichier .tf
// (la réponse Contents API est un objet, pas un tableau).
func TestFetchTFFiles_SingleFile(t *testing.T) {
	const tfContent = `resource "google_compute_instance" "web" {}`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/main.tf", func(w http.ResponseWriter, r *http.Request) {
		entry := contentsEntry{
			Type:     "file",
			Name:     "main.tf",
			Path:     "main.tf",
			Size:     len(tfContent),
			Content:  encodeB64(tfContent),
			Encoding: "base64",
		}
		json.NewEncoder(w).Encode(entry)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "main.tf", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("attendu 1 fichier, obtenu %d", len(files))
	}
	if files[0].Path != "main.tf" || files[0].Content != tfContent {
		t.Errorf("fichier inattendu: %+v", files[0])
	}
}

// TestFetchTFFiles_Directory teste la récupération d'un dossier contenant
// plusieurs fichiers .tf et un fichier non-.tf (ignoré).
func TestFetchTFFiles_Directory(t *testing.T) {
	mainTF := `resource "google_compute_instance" "web" {}`
	varsTF := `variable "region" {}`
	readme := `# README`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		switch path {
		case "":
			entries := []contentsEntry{
				{Type: "file", Name: "main.tf", Path: "main.tf", Size: len(mainTF), Content: encodeB64(mainTF), Encoding: "base64"},
				{Type: "file", Name: "variables.tf", Path: "variables.tf", Size: len(varsTF), Content: encodeB64(varsTF), Encoding: "base64"},
				{Type: "file", Name: "README.md", Path: "README.md", Size: len(readme), Content: encodeB64(readme), Encoding: "base64"},
			}
			json.NewEncoder(w).Encode(entries)
		default:
			http.NotFound(w, r)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("attendu 2 fichiers .tf, obtenu %d: %+v", len(files), files)
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".tf") {
			t.Errorf("fichier non-.tf inclus: %s", f.Path)
		}
	}
}

// TestFetchTFFiles_DirectoryListingWithoutContent simule le comportement réel de
// l'API GitHub Contents : la liste de répertoire ne contient pas le champ
// "content"/"encoding", il faut donc refetcher chaque fichier individuellement
// et utiliser l'encodage de la réponse refetchée (pas celui de l'entrée de liste).
func TestFetchTFFiles_DirectoryListingWithoutContent(t *testing.T) {
	mainTF := `resource "google_compute_network" "demo" {}`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		switch path {
		case "terraform":
			entries := []contentsEntry{
				// Pas de Content/Encoding, comme la vraie API GitHub pour une liste de dossier.
				{Type: "file", Name: "main.tf", Path: "terraform/main.tf", Size: len(mainTF)},
			}
			json.NewEncoder(w).Encode(entries)
		case "terraform/main.tf":
			entry := contentsEntry{
				Type: "file", Name: "main.tf", Path: "terraform/main.tf",
				Size: len(mainTF), Content: encodeB64(mainTF), Encoding: "base64",
			}
			json.NewEncoder(w).Encode(entry)
		default:
			http.NotFound(w, r)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "terraform", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("attendu 1 fichier, obtenu %d: %+v", len(files), files)
	}
	if files[0].Content != mainTF {
		t.Errorf("contenu non décodé correctement: %q", files[0].Content)
	}
	// Le chemin doit être relatif au répertoire racine demandé ("terraform").
	if files[0].Path != "main.tf" {
		t.Errorf("chemin attendu \"main.tf\", obtenu %q", files[0].Path)
	}
}

// TestFetchTFFiles_Recursion teste la traversée récursive des sous-dossiers.
func TestFetchTFFiles_Recursion(t *testing.T) {
	rootTF := `resource "google_compute_instance" "web" {}`
	moduleTF := `resource "google_storage_bucket" "data" {}`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		switch path {
		case "":
			entries := []contentsEntry{
				{Type: "file", Name: "main.tf", Path: "main.tf", Size: len(rootTF), Content: encodeB64(rootTF), Encoding: "base64"},
				{Type: "dir", Name: "modules", Path: "modules"},
			}
			json.NewEncoder(w).Encode(entries)
		case "modules":
			entries := []contentsEntry{
				{Type: "dir", Name: "storage", Path: "modules/storage"},
			}
			json.NewEncoder(w).Encode(entries)
		case "modules/storage":
			entries := []contentsEntry{
				{Type: "file", Name: "storage.tf", Path: "modules/storage/storage.tf", Size: len(moduleTF), Content: encodeB64(moduleTF), Encoding: "base64"},
			}
			json.NewEncoder(w).Encode(entries)
		default:
			http.NotFound(w, r)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("attendu 2 fichiers .tf, obtenu %d: %+v", len(files), files)
	}

	paths := map[string]bool{}
	for _, f := range files {
		paths[f.Path] = true
	}
	if !paths["main.tf"] || !paths["modules/storage/storage.tf"] {
		t.Errorf("chemins inattendus: %+v", files)
	}
}

// TestFetchTFFiles_MaxFileSize vérifie que l'erreur de taille max est respectée.
func TestFetchTFFiles_MaxFileSize(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		if path == "" {
			entries := []contentsEntry{
				{Type: "file", Name: "huge.tf", Path: "huge.tf", Size: maxFileSize + 1, Content: "", Encoding: "base64"},
			}
			json.NewEncoder(w).Encode(entries)
			return
		}
		http.NotFound(w, r)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	_, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err == nil {
		t.Fatal("attendu une erreur pour fichier trop volumineux")
	}
}

// TestFetchTFFiles_NoTFFiles vérifie l'erreur quand aucun fichier .tf n'est trouvé.
func TestFetchTFFiles_NoTFFiles(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		if path == "" {
			entries := []contentsEntry{
				{Type: "file", Name: "README.md", Path: "README.md", Size: 3, Content: encodeB64("foo"), Encoding: "base64"},
			}
			json.NewEncoder(w).Encode(entries)
			return
		}
		http.NotFound(w, r)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	_, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err == nil {
		t.Fatal("attendu une erreur quand aucun fichier .tf trouvé")
	}
}

// TestFetchTFFiles_MaxDepth vérifie que la traversée s'arrête à la profondeur max
// sans erreur (les fichiers au-delà ne sont simplement pas inclus).
func TestFetchTFFiles_MaxDepth(t *testing.T) {
	rootTF := `resource "google_compute_instance" "web" {}`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		depth := 0
		if path != "" {
			depth = strings.Count(path, "/") + 1
		}

		// À chaque niveau, un fichier .tf et un sous-dossier "nested".
		nestedPath := "nested"
		if path != "" {
			nestedPath = path + "/nested"
		}
		filePath := "level.tf"
		if path != "" {
			filePath = path + "/level.tf"
		}

		entries := []contentsEntry{
			{Type: "file", Name: "level.tf", Path: filePath, Size: len(rootTF), Content: encodeB64(rootTF), Encoding: "base64"},
			{Type: "dir", Name: "nested", Path: nestedPath},
		}
		_ = depth
		json.NewEncoder(w).Encode(entries)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}

	// La profondeur est bornée par maxTFDepth ; on ne doit pas boucler indéfiniment
	// et le nombre de fichiers doit rester raisonnable (<= maxTFDepth + 2).
	if len(files) == 0 {
		t.Fatal("attendu au moins quelques fichiers")
	}
	if len(files) > maxTFDepth+2 {
		t.Errorf("trop de fichiers récupérés (limite de profondeur non respectée): %d", len(files))
	}
}

// TestFetchTFFiles_MaxFiles vérifie que le nombre max de fichiers est respecté.
func TestFetchTFFiles_MaxFiles(t *testing.T) {
	tfContent := `resource "google_compute_instance" "web" {}`

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/infra/contents/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/repos/acme/infra/contents/")
		if path != "" {
			http.NotFound(w, r)
			return
		}
		entries := make([]contentsEntry, 0, maxTFFiles+10)
		for i := 0; i < maxTFFiles+10; i++ {
			name := fmt.Sprintf("file%d.tf", i)
			entries = append(entries, contentsEntry{
				Type: "file", Name: name, Path: name, Size: len(tfContent),
				Content: encodeB64(tfContent), Encoding: "base64",
			})
		}
		json.NewEncoder(w).Encode(entries)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := newTestClient(server)
	files, err := c.FetchTFFiles("acme", "infra", "", "main")
	if err != nil {
		t.Fatalf("FetchTFFiles a échoué: %v", err)
	}
	if len(files) > maxTFFiles {
		t.Errorf("attendu au plus %d fichiers, obtenu %d", maxTFFiles, len(files))
	}
}
