package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 教案状态
const (
	LessonStatusDraft     = "draft"
	LessonStatusPublished = "published"
	LessonStatusArchived  = "archived"
)

// Lesson 教案模型
type Lesson struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	Title         string         `gorm:"size:200;not null" json:"title"`
	Subject       string         `gorm:"size:50;not null;index" json:"subject"`
	Grade         string         `gorm:"size:20;not null;index" json:"grade"`
	Duration      int            `gorm:"default:45" json:"duration"`
	Objectives    string         `gorm:"type:jsonb;default:'{}'" json:"objectives"`
	Content       string         `gorm:"type:jsonb;default:'{}'" json:"content"`
	Activities    string         `gorm:"type:text" json:"activities"`
	Assessment    string         `gorm:"type:text" json:"assessment"`
	Resources     string         `gorm:"type:text" json:"resources"`
	Status        string         `gorm:"size:20;default:'draft';index" json:"status"`
	Tags          string         `gorm:"type:jsonb;default:'[]'" json:"tags"`
	Version       int            `gorm:"default:1" json:"version"`
	ViewCount     int            `gorm:"default:0" json:"view_count"`
	LikeCount     int            `gorm:"default:0" json:"like_count"`
	FavoriteCount int            `gorm:"default:0" json:"favorite_count"`
	CommentCount  int            `gorm:"default:0" json:"comment_count"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comments []Comment `gorm:"foreignKey:LessonID" json:"comments,omitempty"`
}

// TableName 表名
func (Lesson) TableName() string {
	return "lessons"
}

// BeforeCreate 创建前钩子
func (l *Lesson) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	if l.Status == "" {
		l.Status = LessonStatusDraft
	}
	return nil
}

// LessonDetail 教案详情响应
type LessonDetail struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Title         string     `json:"title"`
	Subject       string     `json:"subject"`
	Grade         string     `json:"grade"`
	Duration      int        `json:"duration"`
	Objectives    string     `json:"objectives"`
	Content       string     `json:"content"`
	Activities    string     `json:"activities"`
	Assessment    string     `json:"assessment"`
	Resources     string     `json:"resources"`
	Status        string     `json:"status"`
	Tags          []string   `json:"tags"`
	Version       int        `json:"version"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	FavoriteCount int        `json:"favorite_count"`
	CommentCount  int        `json:"comment_count"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	AuthorName    string     `json:"author_name"`
	AuthorAvatar  string     `json:"author_avatar"`
	IsFavorited   bool       `json:"is_favorited"`
	IsLiked       bool       `json:"is_liked"`
}

// LessonVersion 教案版本历史
type LessonVersion struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID      uuid.UUID      `gorm:"type:uuid;index;not null" json:"lesson_id"`
	VersionNumber int            `gorm:"column:version_number;not null" json:"version"`
	Content       string         `gorm:"type:jsonb;not null" json:"content"`
	ChangeSummary string         `gorm:"column:change_summary;type:text" json:"change_log"`
	CreatedBy     *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 表名
func (LessonVersion) TableName() string {
	return "lesson_versions"
}

// BeforeCreate 创建前钩子
func (v *LessonVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// Comment 评论模型
type Comment struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	LessonID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"lesson_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"user_id"`
	ParentID  *uuid.UUID     `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User    *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// TableName 表名
func (Comment) TableName() string {
	return "comments"
}

// Favorite 收藏模型
type Favorite struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index:idx_favorite_user_lesson,unique;not null" json:"user_id"`
	LessonID  uuid.UUID `gorm:"type:uuid;index:idx_favorite_user_lesson,unique;not null" json:"lesson_id"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Lesson *Lesson `gorm:"foreignKey:LessonID" json:"lesson,omitempty"`
}

// TableName 表名
func (Favorite) TableName() string {
	return "lesson_favorites"
}

// Like 点赞模型
type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index:idx_like_user_lesson,unique;not null" json:"user_id"`
	LessonID  uuid.UUID `gorm:"type:uuid;index:idx_like_user_lesson,unique;not null" json:"lesson_id"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 表名
func (Like) TableName() string {
	return "lesson_likes"
}

// LessonListItem 教案列表项
type LessonListItem struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Subject       string     `json:"subject"`
	Grade         string     `json:"grade"`
	Duration      int        `json:"duration"`
	Status        string     `json:"status"`
	Version       int        `json:"version"`
	ViewCount     int        `json:"view_count"`
	LikeCount     int        `json:"like_count"`
	FavoriteCount int        `json:"favorite_count"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	AuthorName    string     `json:"author_name"`
	AuthorAvatar  string     `json:"author_avatar"`
}
