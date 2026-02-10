package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/modulops/auth-service/internal/models"
)

// MigrateExistingDataToDefaultProject migre les données existantes vers un projet par défaut
// Cette fonction crée un projet par défaut pour chaque utilisateur s'il n'en a pas déjà
func (r *Repository) MigrateExistingDataToDefaultProject() error {
	log.Println("🔄 Démarrage de la migration vers un projet par défaut...")

	// Récupérer tous les utilisateurs
	users, err := r.getAllUsers()
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération des utilisateurs: %w", err)
	}

	if len(users) == 0 {
		log.Println("ℹ️  Aucun utilisateur trouvé, migration non nécessaire")
		return nil
	}

	log.Printf("📋 %d utilisateur(s) trouvé(s), vérification des projets...", len(users))

	// Pour chaque utilisateur, vérifier s'il a déjà un projet
	for _, user := range users {
		existingProjects, err := r.GetProjectsByUserID(user.ID)
		if err != nil {
			log.Printf("⚠️  Erreur lors de la récupération des projets pour l'utilisateur %s: %v", user.ID, err)
			continue
		}

		// Ne pas créer de projet automatiquement
		// L'utilisateur devra créer son premier projet via l'interface
		if len(existingProjects) > 0 {
			log.Printf("ℹ️  L'utilisateur %s a déjà %d projet(s)", user.Email, len(existingProjects))
		} else {
			log.Printf("ℹ️  L'utilisateur %s n'a pas encore de projet - il devra en créer un via l'interface", user.Email)
		}
	}

	log.Println("✅ Migration des projets terminée")
	return nil
}

// getAllUsers récupère tous les utilisateurs (fonction utilitaire pour la migration)
func (r *Repository) getAllUsers() ([]*models.User, error) {
	query := `SELECT id, email, username, password, roles, first_name, last_name, active, created_at, updated_at, last_login
		FROM users ORDER BY created_at ASC`

	rows, err := r.db.Query(query)
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
			&user.Password,
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
