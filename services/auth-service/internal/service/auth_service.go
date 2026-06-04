package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/modulops/auth-service/internal/config"
	"github.com/modulops/auth-service/internal/models"
	"github.com/modulops/auth-service/internal/repository"

	"golang.org/x/crypto/bcrypt"
)


type AuthService struct {
	repo *repository.Repository
	cfg  *config.Config
}

func NewAuthService(repo *repository.Repository, cfg *config.Config) *AuthService {
	return &AuthService{
		repo: repo,
		cfg:  cfg,
	}
}

// Register enregistre un nouvel utilisateur
func (s *AuthService) Register(req *models.RegisterRequest) (*models.User, error) {

    //Vérifier champs obligatoires
    if req.Email == "" || req.Password == "" || req.Username == "" {
        return nil, errors.New("champs obligatoires manquants")
    }

    //Vérifier email
    userByEmail, err := s.repo.GetUserByEmail(req.Email)
    if err == nil && userByEmail != nil {
        return nil, errors.New("cet email est déjà utilisé")
    }
    if err != nil && !errors.Is(err, repository.ErrNotFound) {
        return nil, fmt.Errorf("erreur DB email: %w", err)
    }

    //Vérifier username
    userByUsername, err := s.repo.GetUserByUsername(req.Username)
    if err == nil && userByUsername != nil {
        return nil, errors.New("ce nom d'utilisateur est déjà utilisé")
    }
    if err != nil && !errors.Is(err, repository.ErrNotFound) {
        return nil, fmt.Errorf("erreur DB username: %w", err)
    }

    //Hasher le mot de passe
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("erreur hash password: %w", err)
    }

    //Créer utilisateur
    user := models.NewUser(req.Email, req.Username, string(hashedPassword))

    //Sécuriser ID
    if user.ID == "" {
        user.ID = uuid.NewString()
    }

    user.FirstName = req.FirstName
    user.LastName = req.LastName

    //Insert DB + debug
    if err := s.repo.CreateUser(user); err != nil {
        fmt.Printf("CreateUser ERROR: %v\n", err)
        return nil, fmt.Errorf("erreur lors de la création de l'utilisateur: %w", err)
    }

    //Nettoyer réponse
    user.Password = ""

    return user, nil
}

// Login authentifie un utilisateur
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// Récupérer l'utilisateur
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("email ou mot de passe incorrect")
	}

	// Vérifier si l'utilisateur est actif
	if !user.Active {
		return nil, errors.New("compte utilisateur désactivé")
	}

	// Vérifier le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("email ou mot de passe incorrect")
	}

	// Générer le token JWT
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération du token: %w", err)
	}

	// Générer le refresh token
	refreshToken, refreshExpiresAt, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération du refresh token: %w", err)
	}

	// Enregistrer le refresh token
	if err := s.repo.CreateRefreshToken(user.ID, refreshToken, refreshExpiresAt); err != nil {
		return nil, fmt.Errorf("erreur lors de l'enregistrement du refresh token: %w", err)
	}

	// Mettre à jour la dernière connexion
	if err := s.repo.UpdateLastLogin(user.ID); err != nil {
		// Log l'erreur mais ne pas faire échouer la connexion
		fmt.Printf("Erreur lors de la mise à jour de la dernière connexion: %v\n", err)
	}

	// Ne pas retourner le mot de passe
	user.Password = ""

	return &models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
		ExpiresAt:    expiresAt,
	}, nil
}


// RefreshToken génère un nouveau token à partir d'un refresh token
func (s *AuthService) RefreshToken(refreshToken string) (*models.LoginResponse, error) {
	// Récupérer le refresh token
	rt, err := s.repo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("refresh token invalide")
	}

	// Vérifier si le token est révoqué
	if rt.Revoked {
		return nil, errors.New("refresh token révoqué")
	}

	// Vérifier si le token est expiré
	if time.Now().After(rt.ExpiresAt) {
		return nil, errors.New("refresh token expiré")
	}

	// Récupérer l'utilisateur
	user, err := s.repo.GetUserByID(rt.UserID)
	if err != nil {
		return nil, errors.New("utilisateur non trouvé")
	}

	// Vérifier si l'utilisateur est actif
	if !user.Active {
		return nil, errors.New("compte utilisateur désactivé")
	}

	// Générer un nouveau token
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération du token: %w", err)
	}

	// Révoquer l'ancien refresh token
	if err := s.repo.RevokeRefreshToken(refreshToken); err != nil {
		// Log l'erreur mais continuer
		fmt.Printf("Erreur lors de la révocation du refresh token: %v\n", err)
	}

	// Générer un nouveau refresh token
	newRefreshToken, newRefreshExpiresAt, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la génération du nouveau refresh token: %w", err)
	}

	// Enregistrer le nouveau refresh token
	if err := s.repo.CreateRefreshToken(user.ID, newRefreshToken, newRefreshExpiresAt); err != nil {
		return nil, fmt.Errorf("erreur lors de l'enregistrement du nouveau refresh token: %w", err)
	}

	// Ne pas retourner le mot de passe
	user.Password = ""

	return &models.LoginResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user,
		ExpiresAt:    expiresAt,
	}, nil
}

// Logout révoque le refresh token
func (s *AuthService) Logout(refreshToken string) error {
	if refreshToken != "" {
		return s.repo.RevokeRefreshToken(refreshToken)
	}
	return nil
}

// ValidateToken valide un token JWT et retourne les claims
func (s *AuthService) ValidateToken(tokenString string) (*config.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &config.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Vérifier la méthode de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return s.cfg.GetJWTKey(), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erreur lors de la validation du token: %w", err)
	}

	if claims, ok := token.Claims.(*config.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token invalide")
}

// generateToken génère un token JWT pour un utilisateur
func (s *AuthService) generateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.cfg.JWTExpiration)

	claims := &config.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Roles:  []string(user.Roles),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.JWTIssuer,
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.cfg.GetJWTKey())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// generateRefreshToken génère un refresh token
func (s *AuthService) generateRefreshToken(userID string) (string, time.Time, error) {
	// Refresh token expire après 7 jours
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	claims := &config.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.JWTIssuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.cfg.GetJWTKey())
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
