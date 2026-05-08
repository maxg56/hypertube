package models

import "time"

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID            uint      `gorm:"primaryKey;column:id" json:"id"`
	Username      string    `gorm:"column:username;type:varchar(50);uniqueIndex;not null" json:"username"`
	Email         string    `gorm:"column:email;type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"column:password_hash;not null" json:"-"`
	FirstName     string    `gorm:"column:first_name" json:"first_name"`
	LastName      string    `gorm:"column:last_name" json:"last_name"`
	AvatarURL     string    `gorm:"column:avatar_url" json:"avatar_url,omitempty"`
	Language      string    `gorm:"column:language;type:varchar(10);default:'fr'" json:"language"`
	Role          UserRole  `gorm:"column:role;type:user_role_enum;default:'user';not null" json:"role"`
	EmailVerified   bool      `gorm:"column:email_verified;default:false" json:"email_verified"`
	IsPublic        bool      `gorm:"column:is_public;default:true" json:"is_public"`
	FavoritesPublic bool      `gorm:"column:favorites_public;default:true" json:"favorites_public"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "users" }
