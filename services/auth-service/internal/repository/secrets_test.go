package repository

import (
	"strings"
	"testing"
)

func TestIsSensitive(t *testing.T) {
	sensitive := []string{"forgejo_token", "openbao_token", "gcp_credentials_json", "admin_password", "SEMAPHORE_ACCESS_KEY_ENCRYPTION"}
	for _, k := range sensitive {
		if !isSensitive(k) {
			t.Errorf("%s devrait être considérée comme sensible", k)
		}
	}
	// Les clés descriptives restent lisibles en base pour permettre le
	// diagnostic d'une configuration sans clé maître.
	for _, k := range []string{"forgejo_url", "openbao_addr", "openbao_mount_path", "forgejo_repos"} {
		if isSensitive(k) {
			t.Errorf("%s ne devrait pas être chiffrée", k)
		}
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "phrase-de-passe-de-test")

	const secret = "a0164493deadbeef"
	stored, err := encryptValue("forgejo_token", secret)
	if err != nil {
		t.Fatalf("chiffrement: %v", err)
	}
	if !strings.HasPrefix(stored, encryptedPrefix) {
		t.Fatalf("valeur non marquée comme chiffrée: %q", stored)
	}
	if strings.Contains(stored, secret) {
		t.Fatal("le secret apparaît en clair dans la valeur stockée")
	}

	plain, err := decryptValue(stored)
	if err != nil {
		t.Fatalf("déchiffrement: %v", err)
	}
	if plain != secret {
		t.Fatalf("aller-retour incorrect: %q != %q", plain, secret)
	}
}

func TestEncryptIsNotDeterministic(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "phrase-de-passe-de-test")

	a, _ := encryptValue("forgejo_token", "meme-valeur")
	b, _ := encryptValue("forgejo_token", "meme-valeur")
	if a == b {
		t.Fatal("deux chiffrements de la même valeur sont identiques : le nonce n'est pas aléatoire")
	}
}

func TestNonSensitiveValueStaysReadable(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "phrase-de-passe-de-test")

	stored, err := encryptValue("forgejo_url", "https://codeberg.org")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if stored != "https://codeberg.org" {
		t.Fatalf("valeur non sensible altérée: %q", stored)
	}
}

// Une valeur écrite avant l'activation du chiffrement doit rester lisible :
// c'est ce qui permet la montée de version sans migration préalable.
func TestPlaintextLegacyValueIsReadBack(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "phrase-de-passe-de-test")

	plain, err := decryptValue("valeur-ecrite-avant-le-chiffrement")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if plain != "valeur-ecrite-avant-le-chiffrement" {
		t.Fatalf("valeur héritée altérée: %q", plain)
	}
}

// Sans clé maître, l'écriture doit rester possible en clair plutôt que
// d'échouer : une plateforme déjà déployée ne doit pas s'arrêter à la montée
// de version.
func TestWriteWithoutKeyFallsBackToPlaintext(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "")

	stored, err := encryptValue("forgejo_token", "secret")
	if err != nil {
		t.Fatalf("erreur inattendue: %v", err)
	}
	if stored != "secret" {
		t.Fatalf("valeur inattendue sans clé: %q", stored)
	}
}

// À l'inverse, une valeur chiffrée ne doit jamais être renvoyée telle quelle
// si la clé manque : mieux vaut une erreur explicite qu'un secret illisible
// propagé jusqu'à l'API distante.
func TestReadEncryptedWithoutKeyFails(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "phrase-de-passe-de-test")
	stored, _ := encryptValue("forgejo_token", "secret")

	t.Setenv("CONFIG_ENCRYPTION_KEY", "")
	if _, err := decryptValue(stored); err == nil {
		t.Fatal("le déchiffrement sans clé devrait échouer")
	}
}

func TestWrongKeyFails(t *testing.T) {
	t.Setenv("CONFIG_ENCRYPTION_KEY", "cle-originale")
	stored, _ := encryptValue("forgejo_token", "secret")

	t.Setenv("CONFIG_ENCRYPTION_KEY", "cle-differente")
	if _, err := decryptValue(stored); err == nil {
		t.Fatal("le déchiffrement avec une clé différente devrait échouer")
	}
}
