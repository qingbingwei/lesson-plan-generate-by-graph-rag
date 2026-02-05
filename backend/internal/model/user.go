package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 用户角色
const (
	RoleAdmin   = "admin"
	RoleTeacher = "teacher"
	RoleStudent = "student"
)

// 用户状态
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusBanned   = "banned"
)

// User 用户模型
type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	FullName     string         `gorm:"size:100" json:"full_name"`
	AvatarURL    string         `gorm:"size:500" json:"avatar_url"`
	Role         string         `gorm:"size:20;default:'teacher'" json:"role"`
	Status       string         `gorm:"size:20;default:'active'" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.Role == "" {
		u.Role = RoleTeacher
	}
	if u.Status == "" {
		u.Status = StatusActive
	}
	return nil
}

// UserProfile 用户资料响应
type UserProfile struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	FullName      string     `json:"full_name"`
	AvatarURL     string     `json:"avatar_url"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	LessonCount   int64      `json:"lesson_count"`
	FavoriteCount int64      `json:"favorite_count"`
}

// ToProfile 转换为用户资料
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FullName:    u.FullName,
		AvatarURL:   u.AvatarURL,
		Role:        u.Role,
		Status:      u.Status,
		CreatedAt:   u.CreatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}

// UserSettings 用户设置
type UserSettings struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	Theme          string    `gorm:"size:20;default:'light'" json:"theme"`
	Language       string    `gorm:"size:10;default:'zh-CN'" json:"language"`
	EmailNotify    bool      `gorm:"default:true" json:"email_notify"`
	DefaultSubject string    `gorm:"size:50" json:"default_subject"`
	DefaultGrade   string    `gorm:"size:20" json:"default_grade"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 表名
func (UserSettings) TableName() string {
	return "user_settings"
}
