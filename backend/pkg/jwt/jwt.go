package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair Token对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Manager JWT管理器
type Manager struct {
	secretKey     []byte
	expiry        time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewManager 创建JWT管理器
func NewManager(secret string, expiry, refreshExpiry time.Duration, issuer string) *Manager {
	return &Manager{
		secretKey:     []byte(secret),
		expiry:        expiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

// GenerateTokenPair 生成Token对
func (m *Manager) GenerateTokenPair(userID, username, email, role string) (*TokenPair, error) {
	// 生成Access Token
	accessToken, expiresAt, err := m.generateToken(userID, username, email, role, m.expiry)
	if err != nil {
		return nil, err
	}

	// 生成Refresh Token
	refreshToken, _, err := m.generateToken(userID, username, email, role, m.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// GenerateAccessToken 生成Access Token
func (m *Manager) GenerateAccessToken(userID, username, email, role string) (string, int64, error) {
	return m.generateToken(userID, username, email, role, m.expiry)
}

// GenerateRefreshToken 生成Refresh Token
func (m *Manager) GenerateRefreshToken(userID, username, email, role string) (string, int64, error) {
	return m.generateToken(userID, username, email, role, m.refreshExpiry)
}

// generateToken 生成Token
func (m *Manager) generateToken(userID, username, email, role string, expiry time.Duration) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken 验证Token
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// RefreshToken 刷新Token
func (m *Manager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return m.GenerateTokenPair(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// ExtractUserID 从Token中提取用户ID
func (m *Manager) ExtractUserID(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// IsTokenExpired 检查Token是否过期
func (m *Manager) IsTokenExpired(tokenString string) bool {
	_, err := m.ValidateToken(tokenString)
	return errors.Is(err, ErrExpiredToken)
}
