package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
)

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
}

// NewAuthService creates an AuthService.
func NewAuthService(db *gorm.DB, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: []byte(jwtSecret)}
}

// RegisterInput is the payload for user registration.
type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"     binding:"omitempty,oneof=admin researcher viewer"`
}

// LoginInput is the payload for user login.
type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register creates a new user after hashing the password.
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*models.User, error) {
	// Check for existing user
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email = ?", in.Username, in.Email).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if count > 0 {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := in.Role
	if role == "" {
		role = "researcher"
	}

	user := models.User{
		Username:     in.Username,
		Email:        in.Email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

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

// GetUserByID fetches a user by primary key.
func (s *AuthService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
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
