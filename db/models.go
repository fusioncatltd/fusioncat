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

type ProjectsDBModel struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	CreatedByType string    `gorm:"column:created_by_type;type:varchar(45);not null; default:'user';uniqueIndex:idx_unique_project_name,where:status = 'active'"`
	CreatedByID   uuid.UUID `gorm:"type:uuid;column:created_by_id;uniqueIndex:idx_unique_project_name,where:status = 'active'"`
	Name          string    `gorm:"column:name;type:varchar(45);not null;uniqueIndex:idx_unique_project_name,where:status = 'active'"`
	Description   string    `gorm:"column:description;type:text;default null"`
	IsPrivate     bool      `gorm:"column:is_private;type:bool;default:false;not null"`
	Status        string    `gorm:"column:status;type:varchar(30);not null;default:'active'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ProjectsDBModel) TableName() string {
	return "projects"
}
