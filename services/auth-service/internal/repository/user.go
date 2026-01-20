package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/modulops/auth-service/internal/models"
)

// CreateUser crée un nouvel utilisateur
func (r *Repository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (id, email, username, password, roles, first_name, last_name, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.Username,
		user.Password,
		user.Roles,
		user.FirstName,
		user.LastName,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la création de l'utilisateur: %w", err)
	}

	return nil
}

// GetUserByEmail récupère un utilisateur par son email
func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, username, password, roles, first_name, last_name, active, created_at, updated_at, last_login
		FROM users WHERE email = $1`

	user := &models.User{}
	var roles models.StringArray
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&roles,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}

	user.Roles = roles
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetUserByID récupère un utilisateur par son ID
func (r *Repository) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, email, username, password, roles, first_name, last_name, active, created_at, updated_at, last_login
		FROM users WHERE id = $1`

	user := &models.User{}
	var roles models.StringArray
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&roles,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}

	user.Roles = roles
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// GetUserByUsername récupère un utilisateur par son nom d'utilisateur
func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT id, email, username, password, roles, first_name, last_name, active, created_at, updated_at, last_login
		FROM users WHERE username = $1`

	user := &models.User{}
	var roles models.StringArray
	var lastLogin sql.NullTime

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.Password,
		&roles,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLogin,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %w", err)
	}

	user.Roles = roles
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return user, nil
}

// UpdateUser met à jour un utilisateur
func (r *Repository) UpdateUser(user *models.User) error {
	query := `UPDATE users 
		SET email = $2, username = $3, roles = $4, first_name = $5, last_name = $6, active = $7, updated_at = $8
		WHERE id = $1`

	_, err := r.db.Exec(query,
		user.ID,
		user.Email,
		user.Username,
		user.Roles,
		user.FirstName,
		user.LastName,
		user.Active,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de l'utilisateur: %w", err)
	}

	return nil
}

// UpdateUserPassword met à jour le mot de passe d'un utilisateur
func (r *Repository) UpdateUserPassword(userID, hashedPassword string) error {
	query := `UPDATE users SET password = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.Exec(query, userID, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du mot de passe: %w", err)
	}

	return nil
}

// UpdateLastLogin met à jour la date de dernière connexion
func (r *Repository) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = $2 WHERE id = $1`

	_, err := r.db.Exec(query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de la dernière connexion: %w", err)
	}

	return nil
}

// ListUsers récupère la liste des utilisateurs avec pagination
func (r *Repository) ListUsers(limit, offset int) ([]*models.User, error) {
	query := `SELECT id, email, username, roles, first_name, last_name, active, created_at, updated_at, last_login
		FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var roles models.StringArray
		var lastLogin sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&roles,
			&user.FirstName,
			&user.LastName,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLogin,
		)

		if err != nil {
			return nil, fmt.Errorf("erreur lors du scan de l'utilisateur: %w", err)
		}

		user.Roles = roles
		if lastLogin.Valid {
			user.LastLogin = &lastLogin.Time
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des utilisateurs: %w", err)
	}

	return users, nil
}

// DeleteUser supprime un utilisateur
func (r *Repository) DeleteUser(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur lors de la suppression de l'utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification des lignes affectées: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	return nil
}
