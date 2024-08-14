package handler

import (
	"auth-service/internal/auth"
	pb "auth-service/proto"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrMissingMetadata    = errors.New("missing metadata")
	ErrNoAuthToken        = errors.New("authorization token not provided")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
)

type AuthHandler struct {
	db         *sql.DB
	jwtManager *auth.JWTManager
	pb.UnimplementedAuthServiceServer
}

// NewAuthHandler создает новый экземпляр AuthHandler с внедрением зависимостей.
func NewAuthHandler(db *sql.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

// hashPassword хеширует пароль с использованием bcrypt.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// checkPassword сравнивает хешированный пароль с предоставленным паролем.
func checkPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// Login обрабатывает запрос на авторизацию пользователя и генерирует JWT токен.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var userID, hashedPassword string

	// Получаем userID и хешированный пароль из базы данных.
	query := `SELECT id, password FROM users WHERE username=$1`
	err := h.db.QueryRowContext(ctx, query, req.Username).Scan(&userID, &hashedPassword)
	if err != nil {
		log.Printf("Login failed for user %s: %v", req.Username, err)
		return nil, ErrInvalidCredentials
	}

	// Проверка пароля.
	if err := checkPassword(hashedPassword, req.Password); err != nil {
		return nil, err
	}

	// Генерация JWT токена.
	token, err := h.jwtManager.Generate(userID)
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", userID, err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &pb.LoginResponse{AccessToken: token}, nil
}

// validateToken проверяет JWT токен и возвращает его, если он валидный.
func (h *AuthHandler) validateToken(tokenString string) (*jwt.Token, error) {
	token, err := h.jwtManager.Verify(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return token, nil
}

// extractTokenFromMetadata извлекает JWT токен из метаданных контекста.
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMissingMetadata
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return "", ErrNoAuthToken
	}

	// Извлечение токена из строки "Bearer <token>"
	return tokens[0], nil
}

// GetUserProfile возвращает профиль пользователя, используя JWT токен для аутентификации.
func (h *AuthHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	// Извлечение токена из метаданных.
	tokenString, err := extractTokenFromMetadata(ctx)
	if err != nil {
		return nil, err
	}

	// Проверка токена.
	token, err := h.validateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Получение user_id из токена.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenClaims
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidTokenClaims
	}

	// Получение информации о пользователе из базы данных.
	var username string
	query := `SELECT username FROM users WHERE id=$1`
	if err := h.db.QueryRowContext(ctx, query, userID).Scan(&username); err != nil {
		return nil, fmt.Errorf("failed to retrieve user profile: %w", err)
	}

	return &pb.GetUserProfileResponse{UserId: userID, Username: username}, nil
}
