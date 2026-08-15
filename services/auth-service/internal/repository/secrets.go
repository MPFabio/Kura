package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Chiffrement au repos des valeurs de service_configs.
//
// Cette table contient les identifiants que les modules utilisent pour joindre
// les systèmes externes : jeton Forgejo/Codeberg, jeton OpenBao, clés de compte
// de service cloud. Ces valeurs y étaient écrites en clair : toute copie de la
// base — sauvegarde, export, accès en lecture seule — livrait donc l'ensemble
// des accès de la plateforme.
//
// Le chiffrement est appliqué aux seules valeurs sensibles (voir isSensitive)
// et il est transparent : les appelants manipulent toujours du clair. Les
// valeurs déjà en base sont relues telles quelles et rechiffrées à la première
// écriture, sans migration préalable.
//
// Pourquoi une clé d'environnement plutôt qu'OpenBao : le jeton d'accès à
// OpenBao est lui-même une entrée de cette table. L'y stocker rendrait la
// lecture de la configuration dépendante d'un secret contenu dans cette même
// configuration, et un OpenBao scellé empêcherait tout démarrage. OpenBao reste
// le coffre des secrets applicatifs des projets ; cette clé ne protège que les
// identifiants de la plateforme elle-même.
const encryptedPrefix = "enc:v1:"

// sensitiveMarkers repère les clés dont la valeur est un secret. Le test porte
// sur le nom de la clé, ce qui couvre les clés futures sans liste à maintenir.
var sensitiveMarkers = []string{"token", "password", "secret", "credential", "key", "passphrase"}

// isSensitive indique si la valeur d'une clé doit être chiffrée.
//
// Les clés purement descriptives sont exclues explicitement : les chiffrer
// n'apporterait rien et empêcherait de diagnostiquer une configuration
// directement en base.
func isSensitive(key string) bool {
	k := strings.ToLower(key)
	switch k {
	case "openbao_addr", "openbao_mount_path", "forgejo_url", "forgejo_repos", "public_key":
		return false
	}
	for _, marker := range sensitiveMarkers {
		if strings.Contains(k, marker) {
			return true
		}
	}
	return false
}

// cipherFromEnv construit le chiffrement AES-GCM à partir de la clé maître.
//
// La clé est dérivée par SHA-256 afin d'accepter une phrase de passe de
// longueur quelconque sans imposer un format à l'exploitant.
func cipherFromEnv() (cipher.AEAD, error) {
	master := os.Getenv("CONFIG_ENCRYPTION_KEY")
	if master == "" {
		return nil, errNoKey
	}
	sum := sha256.Sum256([]byte(master))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

// errNoKey signale l'absence de clé maître.
var errNoKey = errors.New("CONFIG_ENCRYPTION_KEY absente")

// encryptValue chiffre une valeur sensible. Une valeur vide, déjà chiffrée, ou
// non sensible est retournée telle quelle.
func encryptValue(key, value string) (string, error) {
	if value == "" || !isSensitive(key) || strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	aead, err := cipherFromEnv()
	if err != nil {
		if errors.Is(err, errNoKey) {
			// Sans clé configurée, on conserve le comportement antérieur plutôt
			// que de refuser l'écriture : une plateforme déjà déployée ne doit
			// pas cesser de fonctionner à la montée de version.
			return value, nil
		}
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nonce, nonce, []byte(value), nil)
	return encryptedPrefix + base64.StdEncoding.EncodeToString(sealed), nil
}

// decryptValue déchiffre une valeur. Une valeur sans marqueur est retournée
// telle quelle : c'est le cas des entrées écrites avant l'activation du
// chiffrement, et des valeurs non sensibles.
func decryptValue(value string) (string, error) {
	if !strings.HasPrefix(value, encryptedPrefix) {
		return value, nil
	}
	aead, err := cipherFromEnv()
	if err != nil {
		return "", fmt.Errorf("valeur chiffrée illisible: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, encryptedPrefix))
	if err != nil {
		return "", fmt.Errorf("valeur chiffrée corrompue: %w", err)
	}
	if len(raw) < aead.NonceSize() {
		return "", errors.New("valeur chiffrée tronquée")
	}
	nonce, body := raw[:aead.NonceSize()], raw[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, body, nil)
	if err != nil {
		return "", fmt.Errorf("déchiffrement impossible (clé maître différente ?): %w", err)
	}
	return string(plain), nil
}

// EncryptionEnabled indique si une clé maître est configurée. Utilisé au
// démarrage pour avertir l'exploitant plutôt que de chiffrer silencieusement
// rien du tout.
func EncryptionEnabled() bool {
	_, err := cipherFromEnv()
	return err == nil
}

// EncryptExistingConfigs chiffre les valeurs sensibles encore stockées en
// clair. Appelé au démarrage, il rend la migration native : aucun script à
// lancer, et l'opération est idempotente puisqu'une valeur déjà chiffrée porte
// son marqueur et est ignorée.
//
// Retourne le nombre de valeurs migrées.
func (r *Repository) EncryptExistingConfigs() (int, error) {
	if !EncryptionEnabled() {
		return 0, nil
	}

	rows, err := r.db.Query(`SELECT service, key, value FROM service_configs`)
	if err != nil {
		return 0, err
	}

	type entry struct{ service, key, value string }
	var pending []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.service, &e.key, &e.value); err != nil {
			rows.Close()
			return 0, err
		}
		if e.value == "" || !isSensitive(e.key) || strings.HasPrefix(e.value, encryptedPrefix) {
			continue
		}
		pending = append(pending, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}

	migrated := 0
	for _, e := range pending {
		sealed, err := encryptValue(e.key, e.value)
		if err != nil {
			return migrated, err
		}
		if _, err := r.db.Exec(
			`UPDATE service_configs SET value=$1, updated_at=NOW() WHERE service=$2 AND key=$3`,
			sealed, e.service, e.key,
		); err != nil {
			return migrated, err
		}
		migrated++
	}
	return migrated, nil
}
