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

type SchemasDBModel struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	ProjectID     uuid.UUID `gorm:"project_id:uuid;column:project_id;uniqueIndex:idx_unique_schema_name,where:status = 'active'"`
	CreatedByType string    `gorm:"column:created_by_type;type:varchar(45);not null; default:'user'"`
	CreatedByID   uuid.UUID `gorm:"type:uuid;column:created_by_id;"`
	Name          string    `gorm:"column:name;type:varchar(45);not null;uniqueIndex:idx_unique_schema_name,where:status = 'active'"`
	Description   string    `gorm:"column:description;type:text;default null"`
	Status        string    `gorm:"column:status;type:varchar(30);not null;default:'active'"`
	Type          string    `gorm:"column:type;type:varchar(30);not null;default:'jsonschema'"`
	Schema        string    `gorm:"column:schema;type:text;not null"`
	Version       int       `gorm:"column:version;type:int;not null;default:1;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (SchemasDBModel) TableName() string {
	return "schemas"
}

func (s SchemasDBModel) CanCreate(projectID uuid.UUID, name string) (bool, error) {
	var count int64
	err := GetDB().Model(SchemasDBModel{}).
		Where("project_id = ? AND name = ? AND status = 'active'", projectID, name).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

type SchemaVersionsDBModel struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	SchemaID  uuid.UUID `gorm:"type:uuid;column:schema_id;"`
	UserID    uuid.UUID `gorm:"type:uuid;column:user_id;"`
	Version   int       `gorm:"column:version;type:int;not null;default:1;"`
	Schema    string    `gorm:"column:schema;type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SchemaVersionsDBModel) TableName() string {
	return "schema_versions"
}

type MessagesDBModel struct {
	gorm.Model
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key;"`
	ProjectID     uuid.UUID `gorm:"type:uuid;column:project_id;uniqueIndex:idx_unique_message_name,where:status = 'active'"`
	Name          string    `gorm:"column:name;type:varchar(45);not null;uniqueIndex:idx_unique_message_name,where:status = 'active'"`
	Description   string    `gorm:"column:description;type:text;default null"`
	Status        string    `gorm:"column:status;type:varchar(30);not null;default:'active'"`
	SchemaID      uuid.UUID `gorm:"type:uuid;column:schema_id;"`
	SchemaVersion int       `gorm:"column:schema_version;type:int;not null;default:1;"`
	CreatedByID   uuid.UUID `gorm:"type:uuid;column:created_by_id;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (MessagesDBModel) TableName() string {
	return "messages"
}
