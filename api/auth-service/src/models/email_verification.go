package models

import "time"

type EmailVerification struct {
	ID               uint      `gorm:"primaryKey;column:id"`
	Email            string    `gorm:"column:email;type:varchar(255);not null"`
	VerificationCode string    `gorm:"column:verification_code;type:varchar(6);not null"`
	ExpiresAt        time.Time `gorm:"column:expires_at;not null"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (EmailVerification) TableName() string { return "email_verifications" }
