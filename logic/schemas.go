package logic

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
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

// GenerateCode generates code from this schema version in the specified language
// Returns the generated code and the name of the generated structure/class
// Requires JSON_SCHEMA_CONVERTOR_CMD environment variable to be set with the full path to quicktype
func (schemaVersion *SchemaVersionObject) GenerateCode(language string, schemaType string, schemaName string) (generatedCode string, structName string, err error) {
	if schemaType != "jsonschema" {
		return "", "", fmt.Errorf("unsupported schema type: %s", schemaType)
	}

	// Convert snake_case to CamelCase and ensure it starts with a capital letter
	structName = toCamelCase(schemaName)
	if len(structName) > 0 && structName[0] >= 'a' && structName[0] <= 'z' {
		structName = strings.ToUpper(structName[:1]) + structName[1:]
	}
	
	// Add version suffix to the struct name
	structName = fmt.Sprintf("%sVersion%dFusioncatGeneratedSchema", structName, schemaVersion.dbModel.Version)

	// Get quicktype command from environment (required)
	quicktypeCmd := os.Getenv("JSON_SCHEMA_CONVERTOR_CMD")
	if quicktypeCmd == "" {
		return "", "", fmt.Errorf("JSON_SCHEMA_CONVERTOR_CMD environment variable is not set")
	}

	// Create the quicktype command with appropriate options
	cmd := exec.Command(quicktypeCmd,
		"--lang", language,
		"--just-types",
		"-s", "schema",
		"--top-level", structName)
	
	cmd.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")
	
	// Pipe the schema JSON to quicktype
	cmd.Stdin = strings.NewReader(schemaVersion.dbModel.Schema)
	
	// Capture the output
	var output bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderrBuf

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("code generation failed: %v - stderr: %s", err, stderrBuf.String())
	}

	generatedCode = output.String()
	
	// For Go, ensure the struct name is properly capitalized (extra safety check)
	if language == "go" && len(generatedCode) > 0 {
		// Replace the generated struct name with our capitalized version if needed
		generatedCode = strings.ReplaceAll(generatedCode, "type "+strings.ToLower(structName)+" ", "type "+structName+" ")
		generatedCode = strings.ReplaceAll(generatedCode, "type "+strings.ToLower(structName)+" struct", "type "+structName+" struct")
	}

	return generatedCode, structName, nil
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

func (schema *SchemaObject) GetLatestVersion() int {
	return schema.dbModel.Version
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

// CanNameBeUsed checks if a schema name can be used in a project
func (schemaManager *SchemaObjectsManager) CanNameBeUsed(name string, projectID uuid.UUID) bool {
	return !schemaManager.CheckIfThisSchemaNameAlreadyExists(projectID, name)
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

// SchemaWithVersionExists checks if a schema with the specific ID and version exists and is active
func (schemaManager *SchemaObjectsManager) SchemaWithVersionExists(schemaID uuid.UUID, schemaVersion int) bool {
	var count int64
	connection := db.GetDB()
	
	// Query the database to check if the schema with the specified UUID and version exists
	connection.Model(&db.SchemasDBModel{}).
		Where("id = ? AND version >= ? AND status = 'active'", schemaID, schemaVersion).
		Count(&count)
	
	if count == 0 {
		return false
	}
	
	// Also check if the specific version exists in the versions table
	connection.Model(&db.SchemaVersionsDBModel{}).
		Where("schema_id = ? AND version = ?", schemaID, schemaVersion).
		Count(&count)
	
	return count > 0
}

// capitalizeFirstLetter capitalizes the first letter of a string
func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		result += capitalizeFirstLetter(part)
	}
	return result
}

// GenerateCode generates code from the schema in the specified language
// Returns the generated code and the name of the generated structure/class
// Requires JSON_SCHEMA_CONVERTOR_CMD environment variable to be set with the full path to quicktype
func (schema *SchemaObject) GenerateCode(language string) (generatedCode string, structName string, err error) {
	if schema.dbModel.Type != "jsonschema" {
		return "", "", fmt.Errorf("unsupported schema type: %s", schema.dbModel.Type)
	}

	// Convert snake_case to CamelCase and ensure it starts with a capital letter for all languages
	structName = toCamelCase(schema.dbModel.Name)
	if len(structName) > 0 && structName[0] >= 'a' && structName[0] <= 'z' {
		structName = strings.ToUpper(structName[:1]) + structName[1:]
	}

	// Get quicktype command from environment (required)
	quicktypeCmd := os.Getenv("JSON_SCHEMA_CONVERTOR_CMD")
	if quicktypeCmd == "" {
		return "", "", fmt.Errorf("JSON_SCHEMA_CONVERTOR_CMD environment variable is not set")
	}

	// Create the quicktype command with appropriate options
	// Use --top-level for all languages to ensure proper naming
	cmd := exec.Command(quicktypeCmd,
		"--lang", language,
		"--just-types",
		"-s", "schema",
		"--top-level", structName)
	
	cmd.Env = append(os.Environ(), "NODE_NO_WARNINGS=1")
	
	// Pipe the schema JSON to quicktype
	cmd.Stdin = strings.NewReader(schema.dbModel.Schema)
	
	// Capture the output
	var output bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &stderrBuf

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("code generation failed: %v - stderr: %s", err, stderrBuf.String())
	}

	generatedCode = output.String()
	
	// For Go, ensure the struct name is properly capitalized (extra safety check)
	if language == "go" && len(generatedCode) > 0 {
		// Replace the generated struct name with our capitalized version if needed
		// This handles cases where quicktype might not capitalize properly
		generatedCode = strings.ReplaceAll(generatedCode, "type "+strings.ToLower(structName)+" ", "type "+structName+" ")
		generatedCode = strings.ReplaceAll(generatedCode, "type "+strings.ToLower(structName)+" struct", "type "+structName+" struct")
	}

	return generatedCode, structName, nil
}
