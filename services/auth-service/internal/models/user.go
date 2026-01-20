package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role représente un rôle utilisateur
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// StringArray est un type personnalisé pour stocker un tableau de strings en PostgreSQL
type StringArray []string

// Value implémente driver.Valuer pour StringArray
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(a)
}

// Scan implémente sql.Scanner pour StringArray
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("type non supporté pour StringArray")
	}

	return json.Unmarshal(bytes, a)
}

// User représente un utilisateur dans le système
type User struct {
	ID        string     `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Username  string     `json:"username" db:"username"`
	Password  string     `json:"-" db:"password"` // Ne jamais exposer le mot de passe
	Roles     StringArray `json:"roles" db:"roles"`
	FirstName string     `json:"first_name" db:"first_name"`
	LastName  string     `json:"last_name" db:"last_name"`
	Active    bool       `json:"active" db:"active"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty" db:"last_login"`
}

// HasRole vérifie si l'utilisateur a un rôle spécifique
func (u *User) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if Role(r) == role {
			return true
		}
	}
	return false
}

// IsAdmin vérifie si l'utilisateur est administrateur
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// NewUser crée un nouvel utilisateur avec un ID généré
func NewUser(email, username, password string) *User {
	return &User{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  username,
		Password:  password,
		Roles:     StringArray{string(RoleUser)},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// RegisterRequest représente une requête d'inscription
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest représente une requête de connexion
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse représente la réponse de connexion
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	User         *User     `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshTokenRequest représente une requête de rafraîchissement de token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateUserRequest représente une requête de mise à jour d'utilisateur
type UpdateUserRequest struct {
	Username  *string `json:"username"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Active    *bool   `json:"active"`
}

// ChangePasswordRequest représente une requête de changement de mot de passe
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UpdateRoleRequest représente une requête de mise à jour de rôle
type UpdateRoleRequest struct {
	Roles []string `json:"roles" binding:"required"`
}
