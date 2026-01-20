package service

import (
	"errors"
	"fmt"

	"github.com/modulops/auth-service/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// GetUser récupère un utilisateur par son ID
func (s *AuthService) GetUser(userID string) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	// Ne pas retourner le mot de passe
	user.Password = ""

	return user, nil
}

// UpdateUser met à jour un utilisateur
func (s *AuthService) UpdateUser(userID string, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	// Mettre à jour les champs fournis
	if req.Username != nil {
		// Vérifier si le nom d'utilisateur est déjà utilisé
		existingUser, err := s.repo.GetUserByUsername(*req.Username)
		if err == nil && existingUser.ID != userID {
			return nil, errors.New("ce nom d'utilisateur est déjà utilisé")
		}
		user.Username = *req.Username
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	// Enregistrer les modifications
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise à jour de l'utilisateur: %w", err)
	}

	// Ne pas retourner le mot de passe
	user.Password = ""

	return user, nil
}

// ChangePassword change le mot de passe d'un utilisateur
func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("utilisateur non trouvé")
	}

	// Vérifier le mot de passe actuel
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return errors.New("mot de passe actuel incorrect")
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("erreur lors du hachage du nouveau mot de passe: %w", err)
	}

	// Mettre à jour le mot de passe
	if err := s.repo.UpdateUserPassword(userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du mot de passe: %w", err)
	}

	// Révoquer tous les tokens de l'utilisateur pour forcer une nouvelle connexion
	if err := s.repo.RevokeAllUserTokens(userID); err != nil {
		// Log l'erreur mais ne pas faire échouer le changement de mot de passe
		fmt.Printf("Erreur lors de la révocation des tokens: %v\n", err)
	}

	return nil
}

// ListUsers récupère la liste des utilisateurs
func (s *AuthService) ListUsers(limit, offset int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	users, err := s.repo.ListUsers(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des utilisateurs: %w", err)
	}

	// Ne pas retourner les mots de passe
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

// DeleteUser supprime un utilisateur
func (s *AuthService) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}

// UpdateUserRole met à jour les rôles d'un utilisateur
func (s *AuthService) UpdateUserRole(userID string, roles []string) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	// Valider les rôles
	for _, role := range roles {
		if role != string(models.RoleAdmin) && role != string(models.RoleUser) {
			return nil, fmt.Errorf("rôle invalide: %s", role)
		}
	}

	user.Roles = models.StringArray(roles)

	// Enregistrer les modifications
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, fmt.Errorf("erreur lors de la mise à jour des rôles: %w", err)
	}

	// Révoquer tous les tokens de l'utilisateur pour forcer une nouvelle connexion avec les nouveaux rôles
	if err := s.repo.RevokeAllUserTokens(userID); err != nil {
		// Log l'erreur mais ne pas faire échouer la mise à jour
		fmt.Printf("Erreur lors de la révocation des tokens: %v\n", err)
	}

	// Ne pas retourner le mot de passe
	user.Password = ""

	return user, nil
}
