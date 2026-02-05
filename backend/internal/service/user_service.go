package service

import (
	"context"
	"errors"

	"lesson-plan/backend/internal/model"
	"lesson-plan/backend/internal/repository"
	"lesson-plan/backend/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrUserExists         = errors.New("用户名或邮箱已存在")
	ErrInvalidPassword    = errors.New("密码格式错误")
	ErrUserInactive       = errors.New("用户已被禁用")
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	FullName string `json:"full_name"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string             `json:"access_token"`
	RefreshToken string             `json:"refresh_token"`
	ExpiresAt    int64              `json:"expires_at"`
	User         *model.UserProfile `json:"user"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
}

// UserService 用户服务接口
type UserService interface {
	GetProfile(ctx context.Context, id uuid.UUID) (*model.UserProfile, error)
	UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*model.User, error)
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// authService 认证服务实现
type authService struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo repository.UserRepository, jwtManager *jwt.Manager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 检查邮箱是否存在
	exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Role:         model.RoleTeacher,
		Status:       model.StatusActive,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status != model.StatusActive {
		return nil, ErrUserInactive
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成令牌
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID.String(), user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		User:         user.ToProfile(),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.Status != model.StatusActive {
		return nil, ErrUserInactive
	}

	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID.String(), user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, _, err := s.jwtManager.GenerateRefreshToken(user.ID.String(), user.Username, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		User:         user.ToProfile(),
	}, nil
}

// userService 用户服务实现
type userService struct {
	userRepo     repository.UserRepository
	lessonRepo   repository.LessonRepository
	favoriteRepo repository.FavoriteRepository
}

// NewUserService 创建用户服务
func NewUserService(
	userRepo repository.UserRepository,
	lessonRepo repository.LessonRepository,
	favoriteRepo repository.FavoriteRepository,
) UserService {
	return &userService{
		userRepo:     userRepo,
		lessonRepo:   lessonRepo,
		favoriteRepo: favoriteRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, id uuid.UUID) (*model.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	profile := user.ToProfile()

	// 获取教案数量
	_, lessonCount, _ := s.lessonRepo.ListByUserID(ctx, id, 1, 1)
	profile.LessonCount = lessonCount

	// 获取收藏数量
	favoriteCount, _ := s.favoriteRepo.CountByUserID(ctx, id)
	profile.FavoriteCount = favoriteCount

	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrUserExists
		}
		user.Email = req.Email
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}
