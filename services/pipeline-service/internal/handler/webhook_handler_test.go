package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/modulops/pipeline-service/internal/config"
)

// buildWebhookHandler crée un handler minimal pour tester la validation HMAC.
func buildWebhookHandler(secret string) *WebhookHandler {
	return &WebhookHandler{
		svc: nil, // non utilisé dans les tests HMAC
		cfg: &config.Config{
			GitHubWebhookSecret: secret,
			GitLabWebhookSecret: secret,
		},
	}
}

// computeGitHubSignature calcule la signature HMAC-SHA256 attendue par GitHub.
func computeGitHubSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyGitHubSignature_ValidSignature(t *testing.T) {
	h := buildWebhookHandler("webhook-secret")
	body := []byte(`{"action":"completed","workflow_run":{"id":42}}`)
	sig := computeGitHubSignature("webhook-secret", body)

	if !h.verifyGitHubSignature(body, sig) {
		t.Fatal("une signature HMAC valide devrait être acceptée")
	}
}

func TestVerifyGitHubSignature_WrongSecret(t *testing.T) {
	h := buildWebhookHandler("correct-secret")
	body := []byte(`{"action":"completed"}`)
	sig := computeGitHubSignature("wrong-secret", body)

	if h.verifyGitHubSignature(body, sig) {
		t.Fatal("une signature calculée avec un mauvais secret devrait être rejetée")
	}
}

func TestVerifyGitHubSignature_MissingPrefix(t *testing.T) {
	h := buildWebhookHandler("webhook-secret")
	body := []byte(`{"action":"completed"}`)

	// Signature sans le préfixe "sha256="
	mac := hmac.New(sha256.New, []byte("webhook-secret"))
	mac.Write(body)
	sigWithoutPrefix := hex.EncodeToString(mac.Sum(nil))

	if h.verifyGitHubSignature(body, sigWithoutPrefix) {
		t.Fatal("une signature sans le préfixe 'sha256=' devrait être rejetée")
	}
}

func TestVerifyGitHubSignature_EmptySignature(t *testing.T) {
	h := buildWebhookHandler("webhook-secret")
	body := []byte(`{"action":"completed"}`)

	if h.verifyGitHubSignature(body, "") {
		t.Fatal("une signature vide devrait être rejetée")
	}
}

func TestVerifyGitHubSignature_TamperedBody(t *testing.T) {
	h := buildWebhookHandler("webhook-secret")
	body := []byte(`{"action":"completed"}`)
	sig := computeGitHubSignature("webhook-secret", body)

	// Corps altéré après signature
	tamperedBody := []byte(`{"action":"modified_by_attacker"}`)

	if h.verifyGitHubSignature(tamperedBody, sig) {
		t.Fatal("un corps altéré ne devrait pas correspondre à la signature originale")
	}
}

func TestVerifyGitHubSignature_NoSecretConfigured(t *testing.T) {
	h := buildWebhookHandler("") // secret non configuré = mode dev
	body := []byte(`{"action":"completed"}`)

	// Sans secret configuré, tout webhook est accepté (développement local)
	if !h.verifyGitHubSignature(body, "n'importe-quoi") {
		t.Fatal("sans secret configuré, le webhook devrait être accepté (mode dev)")
	}
}

func TestVerifyGitLabToken_ValidToken(t *testing.T) {
	h := buildWebhookHandler("gitlab-secret")

	if !h.verifyGitLabToken("gitlab-secret") {
		t.Fatal("un token GitLab valide devrait être accepté")
	}
}

func TestVerifyGitLabToken_WrongToken(t *testing.T) {
	h := buildWebhookHandler("correct-token")

	if h.verifyGitLabToken("wrong-token") {
		t.Fatal("un token GitLab incorrect devrait être rejeté")
	}
}

func TestVerifyGitLabToken_EmptyToken(t *testing.T) {
	h := buildWebhookHandler("correct-token")

	if h.verifyGitLabToken("") {
		t.Fatal("un token vide devrait être rejeté lorsque le secret est configuré")
	}
}

func TestVerifyGitLabToken_NoSecretConfigured(t *testing.T) {
	h := buildWebhookHandler("") // mode dev

	if !h.verifyGitLabToken("n'importe-quoi") {
		t.Fatal("sans secret configuré, tout token GitLab devrait être accepté (mode dev)")
	}
}
