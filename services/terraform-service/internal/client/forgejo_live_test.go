package client

import "testing"

func TestFetchTFFilesLive(t *testing.T) {
	if testing.Short() {
		t.Skip("skip live test")
	}
	c := NewForgejoClient("https://codeberg.org", "")
	files, err := c.FetchTFFiles("MPFabio", "Kuro", "terraform", "main")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	for _, f := range files {
		t.Logf("path=%q size=%d", f.Path, len(f.Content))
	}
}
