package logic

import (
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
	"strings"
)

// SchemaObject represents a schema in the system.
// It is heavily coupled with schema versions.
type SchemaObject struct {
	dbModel db.SchemasDBModel
}

// SchemaVersionObject represents a version of a schema.
type SchemaVersionObject struct {
	dbModel db.SchemaVersionsDBModel
}

type SchemaDBSerializerStruct struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	CreatedByType string `json:"created_by_type"`
	CreatedByID   string `json:"created_by_id"`
	ProjectID     string `json:"project_id"`
	CreatedByName string `json:"created_by_name"`
	Schema        string `json:"schema"`
	Version       int    `json:"version"`
}

type SchemaEditShortDBSerializerStruct struct {
	CreatedAt     string `json:"created_at"`
	SchemaID      string `json:"schema_id"`
	UserID        string `json:"user_id"`
	Version       int    `json:"version"`
	CreatedByName string `json:"created_by_name"`
}

type SchemaEditDBSerializerStruct struct {
	CreatedAt     string `json:"created_at"`
	SchemaID      string `json:"schema_id"`
	UserID        string `json:"user_id"`
	Version       int    `json:"version"`
	CreatedByName string `json:"created_by_name"`
	Schema        string `json:"schema"`
}

func (schemaEditObject *SchemaVersionObject) SerializeShort() *SchemaEditShortDBSerializerStruct {
	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, schemaEditObject.dbModel.UserID)
	createdByName := userDbRecord.Handle

	return &SchemaEditShortDBSerializerStruct{
		CreatedAt:     schemaEditObject.dbModel.CreatedAt.String(),
		SchemaID:      schemaEditObject.dbModel.SchemaID.String(),
		UserID:        schemaEditObject.dbModel.UserID.String(),
		Version:       schemaEditObject.dbModel.Version,
		CreatedByName: createdByName,
	}
}

func (schemaEditObject *SchemaVersionObject) SerializeLong() *SchemaEditDBSerializerStruct {
	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, schemaEditObject.dbModel.UserID)
	createdByName := userDbRecord.Handle

	return &SchemaEditDBSerializerStruct{
		CreatedAt:     schemaEditObject.dbModel.CreatedAt.String(),
		SchemaID:      schemaEditObject.dbModel.SchemaID.String(),
		UserID:        schemaEditObject.dbModel.UserID.String(),
		Version:       schemaEditObject.dbModel.Version,
		Schema:        schemaEditObject.dbModel.Schema,
		CreatedByName: createdByName,
	}
}

func (schema *SchemaObject) Serialize() *SchemaDBSerializerStruct {
	createdByName := ""

	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, schema.dbModel.CreatedByID)
	createdByName = userDbRecord.Handle

	return &SchemaDBSerializerStruct{
		ID:            schema.dbModel.ID.String(),
		Name:          schema.dbModel.Name,
		Status:        schema.dbModel.Status,
		Description:   schema.dbModel.Description,
		CreatedByType: schema.dbModel.CreatedByType,
		CreatedByID:   schema.dbModel.CreatedByID.String(),
		CreatedByName: createdByName,
		Schema:        schema.dbModel.Schema,
		Version:       schema.dbModel.Version,
		Type:          schema.dbModel.Type,
		ProjectID:     schema.dbModel.ProjectID.String(),
	}
}

func (schema *SchemaObject) GetProjectID() uuid.UUID {
	return schema.dbModel.ProjectID
}

func (schema *SchemaObject) GetCurrentVersion() int {
	return schema.dbModel.Version
}

func (schema *SchemaObject) GetType() string { return schema.dbModel.Type }

func (schema *SchemaObject) GetName() string { return schema.dbModel.Name }

func (schema *SchemaObject) GetID() uuid.UUID {
	return schema.dbModel.ID
}

func (schemaObject *SchemaObject) GetAllVersions() []*SchemaVersionObject {
	var schemasEdits []db.SchemaVersionsDBModel
	db.GetDB().Where("schema_id = ?", schemaObject.GetID()).Order(
		"version asc").Find(&schemasEdits)

	var schemasEditObjects []*SchemaVersionObject
	for _, schemaEditDBRecord := range schemasEdits {
		schemasEditObjects = append(schemasEditObjects, &SchemaVersionObject{dbModel: schemaEditDBRecord})
	}

	return schemasEditObjects
}

// SchemaObjectsManager manages schemas and schema versions together
type SchemaObjectsManager struct {
}

func (schemaManager *SchemaObjectsManager) GetAllSchemasInProject(ProjectID uuid.UUID) []SchemaObject {
	var schemas []db.SchemasDBModel
	db.GetDB().Where("project_id = ? and status = ?", ProjectID, "active").Order(
		"name asc").Find(&schemas)

	var schemaObjects []SchemaObject
	for _, schema := range schemas {
		schemaObjects = append(schemaObjects, SchemaObject{dbModel: schema})
	}

	return schemaObjects
}

// CheckIfThisSchemaNameAlreadyExists checks if a schema with the given name already exists in the project.
// Users can refer to schemas by name, so we need to ensure that the name is unique in order to avoid confusion.
func (schemaManager *SchemaObjectsManager) CheckIfThisSchemaNameAlreadyExists(ProjectID uuid.UUID, name string) bool {
	var count int64
	db.GetDB().Model(db.SchemasDBModel{}).Where(
		"name = ? and project_id = ? and status = ?", name, ProjectID, "active").Count(&count)
	return count > 0
}

// CreateANewSchema creates a new schema in the database.
// Creation of schema also creates a new schema version.
func (schemaManager *SchemaObjectsManager) CreateANewSchema(name string,
	description string,
	schema string,
	schemaType string,
	createdByType string,
	createdById uuid.UUID,
	userID uuid.UUID,
	projectID uuid.UUID,
) (*SchemaObject, error) {

	connection := db.GetDB()
	tx := connection.Begin()

	// Create a new schema
	newSchema := db.SchemasDBModel{
		Name:          strings.TrimSpace(name),
		Description:   description,
		Schema:        schema,
		Type:          schemaType,
		Version:       1,
		Status:        "active",
		CreatedByType: createdByType,
		CreatedByID:   createdById,
		ProjectID:     projectID,
	}

	if err := tx.Create(&newSchema).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create a new schema edit
	// Create a new schema version
	newSchemaEdit := &db.SchemaVersionsDBModel{
		SchemaID: newSchema.ID,
		UserID:   userID,
		Version:  newSchema.Version,
		Schema:   newSchema.Schema,
	}
	if err := tx.Create(&newSchemaEdit).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return &SchemaObject{dbModel: newSchema}, nil
}

// GetByID retrieves a schema by its ID
func (schemaManager *SchemaObjectsManager) GetByID(schemaID uuid.UUID) (*SchemaObject, error) {
	var schema db.SchemasDBModel
	result := db.GetDB().Where("id = ? AND status = ?", schemaID, "active").First(&schema)
	if result.Error != nil {
		return nil, result.Error
	}
	return &SchemaObject{dbModel: schema}, nil
}

// GetSpecificVersionOfSchema retrieves a specific version of a schema
func (schemaManager *SchemaObjectsManager) GetSpecificVersionOfSchema(schemaID uuid.UUID, version int) (*SchemaVersionObject, error) {
	var schemaVersion db.SchemaVersionsDBModel
	result := db.GetDB().Where("schema_id = ? AND version = ?", schemaID, version).First(&schemaVersion)
	if result.Error != nil {
		return nil, result.Error
	}
	return &SchemaVersionObject{dbModel: schemaVersion}, nil
}

// CreateANewVersion creates a new version of an existing schema
func (schema *SchemaObject) CreateANewVersion(newSchemaContent string, userID uuid.UUID) (*SchemaObject, error) {
	connection := db.GetDB()
	tx := connection.Begin()

	// Update the schema version
	schema.dbModel.Version++
	schema.dbModel.Schema = newSchemaContent
	
	if err := tx.Save(&schema.dbModel).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Create a new schema version record
	newSchemaVersion := &db.SchemaVersionsDBModel{
		SchemaID: schema.dbModel.ID,
		UserID:   userID,
		Version:  schema.dbModel.Version,
		Schema:   newSchemaContent,
	}
	
	if err := tx.Create(&newSchemaVersion).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return schema, nil
}
