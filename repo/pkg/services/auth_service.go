package services

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mindflow/agri-platform/pkg/crypto"
	"github.com/mindflow/agri-platform/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists       = errors.New("username or email already exists")
	ErrInvalidCreds     = errors.New("invalid username or password")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrForbidden        = errors.New("insufficient permissions")
	ErrWeakPassword     = errors.New("password must be at least 8 characters and contain both letters and numbers")
	ErrRoleNotAllowed   = errors.New("requested role is not allowed for public registration")
)

// publicRegistrationRoles lists roles that may be self-assigned at the public
// registration endpoint. Privileged roles (admin, reviewer, customer_service)
// must be provisioned through an admin-managed flow.
var publicRegistrationRoles = map[string]struct{}{
	"researcher": {},
	"viewer":     {},
}

// IsPublicRegistrationRole reports whether a role is valid for public registration.
func IsPublicRegistrationRole(role string) bool {
	_, ok := publicRegistrationRoles[role]
	return ok
}

// Claims embeds standard JWT claims plus our domain fields.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService handles registration, login, and token operations.
type AuthService struct {
	db        *gorm.DB
	jwtSecret []byte
	encryptor *crypto.FieldEncryptor
}

// NewAuthService creates an AuthService.
func NewAuthService(db *gorm.DB, jwtSecret string, encryptionKey string) *AuthService {
	enc, err := crypto.NewFieldEncryptor(encryptionKey)
	if err != nil {
		panic(fmt.Sprintf("invalid encryption key: %v", err))
	}
	return &AuthService{db: db, jwtSecret: []byte(jwtSecret), encryptor: enc}
}

// RegisterInput is the payload for user registration. Only non-privileged
// roles (researcher, viewer) may be self-assigned via public registration.
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"     binding:"omitempty,oneof=researcher viewer"`
}

// AdminCreateUserInput is the payload for admin-managed user creation,
// which permits elevated role assignment.
type AdminCreateUserInput struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"     binding:"required,oneof=admin researcher reviewer customer_service viewer"`
}

// LoginInput is the payload for user login.
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// validatePasswordComplexity ensures the password contains both letters and numbers.
func validatePasswordComplexity(password string) error {
	var hasLetter, hasDigit bool
	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
		if hasLetter && hasDigit {
			return nil
		}
	}
	return ErrWeakPassword
}

// Register creates a new user via the public registration endpoint. The role
// must be one of the non-privileged roles (researcher, viewer); any other
// requested role is rejected to prevent self-assigned privilege escalation.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, error) {
	// Enforce password complexity: must contain both letters and numbers
	if err := validatePasswordComplexity(in.Password); err != nil {
		return nil, err
	}

	role := in.Role
	if role == "" {
		role = "researcher"
	}
	if !IsPublicRegistrationRole(role) {
		return nil, ErrRoleNotAllowed
	}

	return s.createUser(ctx, in.Username, in.Email, in.Password, role)
}

// CreateUserByAdmin allows an admin caller to provision a user with any role.
// This path must only be reachable through an admin-guarded route.
func (s *AuthService) CreateUserByAdmin(ctx context.Context, in AdminCreateUserInput) (*models.User, error) {
	if err := validatePasswordComplexity(in.Password); err != nil {
		return nil, err
	}
	return s.createUser(ctx, in.Username, in.Email, in.Password, in.Role)
}

func (s *AuthService) createUser(ctx context.Context, username, email, password, role string) (*models.User, error) {
	// Check for existing user by username or email hash
	emailHash := models.HashEmail(email)
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email_hash = ?", username, emailHash).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Encrypt email for at-rest protection
	encryptedEmail, err := s.encryptor.Encrypt(email)
	if err != nil {
		return nil, fmt.Errorf("encrypt email: %w", err)
	}

	user := models.User{
		Username:     username,
		Email:        encryptedEmail,
		EmailHash:    emailHash,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	user.MaskedEmail = crypto.MaskEmail(email)
	return &user, nil
}

// Login validates credentials and returns a JWT token string.
func (s *AuthService) Login(ctx context.Context, in LoginInput) (string, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("username = ?", in.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrInvalidCreds
		}
		return "", fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return "", ErrInvalidCreds
	}

	return s.generateToken(&user)
}

// ValidateToken parses and validates a JWT, returning the claims.
func (s *AuthService) ValidateToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GetUserByID fetches a user by primary key and populates the masked email.
func (s *AuthService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	s.populateMaskedEmail(&user)
	return &user, nil
}

// DecryptUserEmail returns the full decrypted email for the given user (admin/self use only).
func (s *AuthService) DecryptUserEmail(user *models.User) (string, error) {
	return s.encryptor.Decrypt(user.Email)
}

// populateMaskedEmail decrypts the email and sets the masked version for display.
func (s *AuthService) populateMaskedEmail(user *models.User) {
	decrypted, err := s.encryptor.Decrypt(user.Email)
	if err != nil {
		user.MaskedEmail = "***"
		return
	}
	user.MaskedEmail = crypto.MaskEmail(decrypted)
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
