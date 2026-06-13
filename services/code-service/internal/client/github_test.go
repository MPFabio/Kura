package client

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetTree(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/contents/src" {
			t.Fatalf("chemin inattendu: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"type":"dir","name":"internal","path":"src/internal","size":0},
			{"type":"file","name":"main.go","path":"src/main.go","size":123}
		]`))
	}))
	defer srv.Close()

	c := &GitHubClient{token: "tok", apiBase: srv.URL, httpClient: srv.Client()}
	entries, err := c.GetTree("owner", "repo", "src", "main")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("attendu 2 entrées, obtenu %d", len(entries))
	}
	if entries[0].Type != "dir" || entries[0].Path != "src/internal" {
		t.Errorf("entrée 0 inattendue: %+v", entries[0])
	}
	if entries[1].Type != "file" || entries[1].Name != "main.go" {
		t.Errorf("entrée 1 inattendue: %+v", entries[1])
	}
}

func TestGetFileContent(t *testing.T) {
	content := "package main\n\nfunc main() {}\n"
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer tok" {
			t.Errorf("header Authorization inattendu: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"type":"file","name":"main.go","path":"main.go","size":` +
			itoa(len(content)) + `,"content":"` + encoded + `","encoding":"base64"}`))
	}))
	defer srv.Close()

	c := &GitHubClient{token: "tok", apiBase: srv.URL, httpClient: srv.Client()}
	file, err := c.GetFileContent("owner", "repo", "main.go", "")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if file.Content != content {
		t.Errorf("contenu inattendu: %q", file.Content)
	}
	if file.Truncated {
		t.Errorf("ne devrait pas être tronqué")
	}
}

func TestGetFileContentTooLarge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"type":"file","name":"big.bin","path":"big.bin","size":2097152,"content":"","encoding":"base64"}`))
	}))
	defer srv.Close()

	c := &GitHubClient{apiBase: srv.URL, httpClient: srv.Client()}
	file, err := c.GetFileContent("owner", "repo", "big.bin", "")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if !file.Truncated {
		t.Errorf("attendu Truncated=true pour un fichier trop volumineux")
	}
	if file.Content != "" {
		t.Errorf("contenu attendu vide pour un fichier tronqué")
	}
}

func TestGetCommits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "path=main.go") {
			t.Errorf("query string inattendue: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"sha":"abc123","commit":{"message":"fix: bug","author":{"name":"Alice","date":"2026-01-01T00:00:00Z"}},"html_url":"https://github.com/owner/repo/commit/abc123"}
		]`))
	}))
	defer srv.Close()

	c := &GitHubClient{apiBase: srv.URL, httpClient: srv.Client()}
	commits, err := c.GetCommits("owner", "repo", "main.go", "main", 1)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("attendu 1 commit, obtenu %d", len(commits))
	}
	if commits[0].SHA != "abc123" || commits[0].Author != "Alice" {
		t.Errorf("commit inattendu: %+v", commits[0])
	}
}

func TestGetCommitDiff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/commits/abc123" {
			t.Fatalf("chemin inattendu: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"sha":"abc123",
			"commit":{"message":"fix: bug","author":{"name":"Alice","date":"2026-01-01T00:00:00Z"}},
			"html_url":"https://github.com/owner/repo/commit/abc123",
			"files":[{"filename":"main.go","status":"modified","additions":2,"deletions":1,"patch":"@@ -1,1 +1,2 @@\n+x"}]
		}`))
	}))
	defer srv.Close()

	c := &GitHubClient{apiBase: srv.URL, httpClient: srv.Client()}
	detail, err := c.GetCommitDiff("owner", "repo", "abc123")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if detail.SHA != "abc123" {
		t.Errorf("SHA inattendu: %s", detail.SHA)
	}
	if len(detail.Files) != 1 || detail.Files[0].Filename != "main.go" {
		t.Errorf("fichiers inattendus: %+v", detail.Files)
	}
	if detail.Files[0].Additions != 2 || detail.Files[0].Deletions != 1 {
		t.Errorf("compteurs inattendus: %+v", detail.Files[0])
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
