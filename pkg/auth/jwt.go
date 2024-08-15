package handler

import (
	"auth-service/internal/auth"
	pb "auth-service/proto"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
)

// AuthHandler реализует интерфейс AuthServiceServer и обрабатывает запросы аутентификации.
type AuthHandler struct {
	db         *sql.DB
	jwtManager *auth.JWTManager
	pb.UnimplementedAuthServiceServer
}

// NewAuthHandler создает и возвращает новый экземпляр AuthHandler с внедренными зависимостями.
func NewAuthHandler(db *sql.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jwtManager,
	}
}

// hashPassword хеширует пароль с использованием bcrypt и возвращает хешированный пароль.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// checkPassword сравнивает хешированный пароль с предоставленным паролем и возвращает ошибку, если они не совпадают.
func checkPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return errors.New("invalid credentials")
	}
	return nil
}

// Login обрабатывает запрос на авторизацию пользователя и генерирует JWT токен при успешной проверке учетных данных.
func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var userID, hashedPassword string

	// Извлекаем userID и хешированный пароль из базы данных на основе имени пользователя.
	query := `SELECT id, password FROM users WHERE username=$1`
	if err := h.db.QueryRow(query, req.Username).Scan(&userID, &hashedPassword); err != nil {
		return nil, fmt.Errorf("invalid username or password: %w", err)
	}

	// Проверяем соответствие введенного пароля с хешем в базе данных.
	if err := checkPassword(hashedPassword, req.Password); err != nil {
		return nil, err
	}

	// Генерируем JWT токен.
	token, err := h.jwtManager.Generate(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &pb.LoginResponse{AccessToken: token}, nil
}

// validateToken проверяет JWT токен и возвращает его, если он валидный.
func (h *AuthHandler) validateToken(tokenString string) (*jwt.Token, error) {
	token, err := h.jwtManager.Verify(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}
	return token, nil
}

// GetUserProfile возвращает профиль пользователя, используя JWT токен для аутентификации.
func (h *AuthHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	// Извлекаем токен из метаданных контекста.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing metadata")
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return nil, errors.New("authorization token not provided")
	}

	// Проверяем токен.
	token, err := h.validateToken(tokens[0])
	if err != nil {
		return nil, err
	}

	// Извлекаем user_id из токена.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id claim not found in token")
	}

	// Получаем информацию о пользователе из базы данных на основе user_id.
	var username string
	query := `SELECT username FROM users WHERE id=$1`
	if err := h.db.QueryRow(query, userID).Scan(&username); err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return &pb.GetUserProfileResponse{
		UserId:   userID,
		Username: username,
	}, nil
}
