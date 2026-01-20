package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefreshToken représente un token de rafraîchissement
type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

// CreateRefreshToken crée un nouveau token de rafraîchissement
func (r *Repository) CreateRefreshToken(userID, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6)`

	id := uuid.New().String()
	_, err := r.db.Exec(query, id, userID, token, expiresAt, time.Now(), false)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken récupère un token de rafraîchissement
func (r *Repository) GetRefreshToken(token string) (*RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, created_at, revoked
		FROM refresh_tokens WHERE token = $1`

	rt := &RefreshToken{}
	err := r.db.QueryRow(query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.Revoked,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("refresh token non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du refresh token: %w", err)
	}

	return rt, nil
}

// RevokeRefreshToken révoque un token de rafraîchissement
func (r *Repository) RevokeRefreshToken(token string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token = $1`

	_, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("erreur lors de la révocation du refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens révoque tous les tokens d'un utilisateur
func (r *Repository) RevokeAllUserTokens(userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la révocation des tokens: %w", err)
	}

	return nil
}

// CleanExpiredTokens supprime les tokens expirés
func (r *Repository) CleanExpiredTokens() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < $1 OR revoked = true`

	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("erreur lors du nettoyage des tokens expirés: %w", err)
	}

	return nil
}
