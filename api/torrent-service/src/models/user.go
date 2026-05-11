package models

type User struct {
	ID int `gorm:"primaryKey;column:id"`
}

func (User) TableName() string { return "users" }
