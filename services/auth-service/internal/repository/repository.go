package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/modulops/auth-service/internal/config"

	_ "github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func New(cfg *config.Config) (*Repository, error) {
	dsn := cfg.GetDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la connexion: %w", err)
	}

	// Tester la connexion
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erreur lors du ping de la base de données: %w", err)
	}

	// Configurer le pool de connexions
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	repo := &Repository{db: db}

	// Créer les tables si elles n'existent pas
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("erreur lors de la migration: %w", err)
	}

	// Migrer les données existantes vers un projet par défaut
	// Cette migration est idempotente et peut être exécutée plusieurs fois
	if err := repo.MigrateExistingDataToDefaultProject(); err != nil {
		// Log l'erreur mais ne pas faire échouer le démarrage
		log.Printf("⚠️  Erreur lors de la migration vers un projet par défaut (non bloquant): %v", err)
	}

	return repo, nil
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			username VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			roles JSONB DEFAULT '["user"]'::jsonb,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_active ON users(active)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(500) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			revoked BOOLEAN DEFAULT false
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			owner_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name)`,
		`CREATE TABLE IF NOT EXISTS project_members (
			id VARCHAR(36) PRIMARY KEY,
			project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(20) NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, user_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_members_role ON project_members(role)`,
		`CREATE TABLE IF NOT EXISTS project_mappings (
			id VARCHAR(36) PRIMARY KEY,
			project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			github_repository VARCHAR(255),
			terraform_state_id VARCHAR(255),
			terraform_source_id VARCHAR(255),
			cluster_id VARCHAR(255),
			cluster_namespace VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_project_mappings_github ON project_mappings(project_id, github_repository) WHERE github_repository IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_project_mappings_terraform ON project_mappings(project_id, terraform_state_id) WHERE terraform_state_id IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_project_mappings_cluster ON project_mappings(project_id, cluster_id, cluster_namespace) WHERE cluster_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_project_mappings_project_id ON project_mappings(project_id)`,
		`CREATE TABLE IF NOT EXISTS project_permissions (
			id VARCHAR(36) PRIMARY KEY,
			project_id VARCHAR(36) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			module VARCHAR(50) NOT NULL,
			scope VARCHAR(20) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, user_id, module)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_permissions_project_id ON project_permissions(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_permissions_user_id ON project_permissions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_permissions_module ON project_permissions(module)`,
		`CREATE INDEX IF NOT EXISTS idx_project_mappings_github_repo ON project_mappings(github_repository)`,
		`CREATE INDEX IF NOT EXISTS idx_project_mappings_terraform_state ON project_mappings(terraform_state_id)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			return fmt.Errorf("erreur lors de l'exécution de la migration: %w", err)
		}
	}

	return nil
}
