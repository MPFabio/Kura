package fine

import (
	"os"
	"path/filepath"
	"testing"
)

// Le cache de plugins est recopié pour chaque exécution afin d'éviter les
// « text file busy » : ces tests couvrent la copie, puisqu'une erreur silencieuse
// ferait retélécharger le fournisseur à chaque détection, ou pire, produirait un
// binaire non exécutable.

func TestCopyTreeAbsentSourceIsNotAnError(t *testing.T) {
	dst := t.TempDir()

	seeded, err := copyTree(filepath.Join(t.TempDir(), "inexistant"), dst)
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if seeded {
		t.Fatal("une source absente ne doit pas être signalée comme amorcée")
	}
}

func TestCopyTreeEmptySourceIsNotSeeded(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "vide"), 0o700); err != nil {
		t.Fatal(err)
	}

	seeded, err := copyTree(src, t.TempDir())
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if seeded {
		t.Fatal("un cache sans fichier ne doit pas être considéré comme amorcé")
	}
}

func TestCopyTreePreservesContentAndExecutableBit(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	nested := filepath.Join(src, "registry.opentofu.org", "hashicorp", "google", "7.44.0", "linux_amd64")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(nested, "terraform-provider-google")
	if err := os.WriteFile(binary, []byte("faux binaire"), 0o755); err != nil {
		t.Fatal(err)
	}

	seeded, err := copyTree(src, dst)
	if err != nil {
		t.Fatalf("copie: %v", err)
	}
	if !seeded {
		t.Fatal("la copie d'un cache peuplé doit être signalée comme amorcée")
	}

	copied := filepath.Join(dst, "registry.opentofu.org", "hashicorp", "google", "7.44.0", "linux_amd64", "terraform-provider-google")
	content, err := os.ReadFile(copied)
	if err != nil {
		t.Fatalf("fichier non recopié: %v", err)
	}
	if string(content) != "faux binaire" {
		t.Fatalf("contenu altéré: %q", content)
	}

	// On compare au mode de la source plutôt qu'à une valeur en dur : Windows
	// n'ayant pas de bit exécutable, exiger 0755 ferait échouer le test sur le
	// poste de développement alors que le service tourne sous Linux. L'invariant
	// utile est la conservation du mode — sans lui, tofu retélécharge le
	// fournisseur, ou échoue à le lancer.
	srcInfo, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	dstInfo, err := os.Stat(copied)
	if err != nil {
		t.Fatal(err)
	}
	if srcInfo.Mode().Perm() != dstInfo.Mode().Perm() {
		t.Fatalf("permissions non conservées: source %v, copie %v", srcInfo.Mode().Perm(), dstInfo.Mode().Perm())
	}
}

func TestCopyTreeRejectsFileAsSource(t *testing.T) {
	file := filepath.Join(t.TempDir(), "fichier")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := copyTree(file, t.TempDir()); err == nil {
		t.Fatal("une source qui n'est pas un répertoire doit être refusée")
	}
}
