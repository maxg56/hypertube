package models

import "time"

type Comment struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	MovieID   int       `gorm:"column:movie_id;not null"`
	UserID    int       `gorm:"column:user_id;not null"`
	Content   string    `gorm:"column:content;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Comment) TableName() string { return "comments" }

type CommentResponse struct {
	ID        uint      `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Movie struct {
	ID     uint   `gorm:"primaryKey;column:id"`
	TMDbID int    `gorm:"column:tmdb_id;uniqueIndex;not null"`
	Title  string `gorm:"column:title;not null"`
}

func (Movie) TableName() string { return "movies" }

type User struct {
	ID        uint   `gorm:"primaryKey;column:id"`
	Username  string `gorm:"column:username"`
	AvatarURL string `gorm:"column:avatar_url"`
}

func (User) TableName() string { return "users" }
