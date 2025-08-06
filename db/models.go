package db

import "time"

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UsersDBModel struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	Handle       string    `gorm:"column:handle;type:varchar(30);not null;uniqueIndex:unique_handle"`
	Status       string    `gorm:"column:status;type:varchar(30);not null"`
	Email        string    `gorm:"column:email;uniqueIndex:unique_email;default null"`
	PasswordHash string    `gorm:"column:password;default null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UsersDBModel) TableName() string {
	return "users"
}
