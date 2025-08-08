package protected_endpoints

import (
	"fmt"
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

func SchemasProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/projects/:id/schemas", GetAllSchemasInProjectV1)
	router.POST("/projects/:id/schemas", NewSchemaInProjectV1)
	router.GET("/schemas/:schemaID", GetSingleSchemaV1)
	router.PUT("/schemas/:schemaID", ModifySchemaV1)
	router.GET("/schemas/:schemaID/versions", GetSchemaVersionsV1)
	router.GET("/schemas/:schemaID/versions/:versionID", GetSingleSchemaVersionsV1)
}

// Get all schemas in project
// @Summary Get all schemas in project
// @Description Get all schemas in project
// @Produce json
// @Tags Schemas
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {array} logic.SchemaDBSerializerStruct "List of schemas in project"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Router /v1/protected/projects/{id}/schemas [get]
func GetAllSchemasInProjectV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	schemasManager := logic.SchemaObjectsManager{}
	schemas := schemasManager.GetAllSchemasInProject(parsedProjectID)

	var schemasSerialized []logic.SchemaDBSerializerStruct
	for _, schema := range schemas {
		schemasSerialized = append(schemasSerialized, *schema.Serialize())
	}

	if len(schemasSerialized) == 0 {
		c.JSON(http.StatusOK, make([]logic.SchemaDBSerializerStruct, 0))
	} else {
		c.JSON(http.StatusOK, schemasSerialized)
	}
}

// Create new schema in project
// @Summary Create new schema in project
// @Description Create new schema in project
// @Produce json
// @Accept json
// @Tags Schemas
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param project body input_contracts.CreateSchemaApiInputContract true "New schema request payload"
// @Success 200 {object} logic.SchemaDBSerializerStruct "Nodified schema"
// @Success 401 "Access denied: missing or invalid Authorization header"
// @Success 404 "Project not found"
// @Success 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects/{id}/schemas [post]
func NewSchemaInProjectV1(c *gin.Context) {

	var input input_contracts.CreateSchemaApiInputContract

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	userID, _ := c.Get("UserID")
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	schemasManager := logic.SchemaObjectsManager{}

	// Making sure the schema name is unique for the project
	if schemasManager.CheckIfThisSchemaNameAlreadyExists(parsedProjectID, input.Name) {
		c.AbortWithStatusJSON(http.StatusConflict, nil)
		return
	}

	schema, _ := schemasManager.CreateANewSchema(
		input.Name,
		input.Description,
		input.Schema,
		input.Type,
		"user", userID.(uuid.UUID),
		userID.(uuid.UUID),
		parsedProjectID,
	)
	c.JSON(http.StatusOK, schema.Serialize())
}

// Get schema
// @Summary Get schema
// @Description Get schema
// @Produce json
// @Tags Schemas
// @Security BearerAuth
// @Param schemaID path string true "Schema ID"
// @Success 200 {object} logic.SchemaDBSerializerStruct "Schema information"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Schema not found"
// @Router /v1/protected/schemas/{schemaID} [get]
func GetSingleSchemaV1(c *gin.Context) {
	schemaID := c.Param("schemaID")
	parsedSchemaID, _ := uuid.Parse(schemaID)

	schemasManager := logic.SchemaObjectsManager{}
	schema, err := schemasManager.GetByID(parsedSchemaID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Schema is found, verify user has access to its project
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(schema.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, schema.Serialize())
}

// Modify schema
// @Summary Modify schema
// @Description Modify schema by creating a new version
// @Produce json
// @Accept json
// @Tags Schemas
// @Security BearerAuth
// @Param schemaID path string true "Schema ID"
// @Param schema body input_contracts.ModifySchemaApiInputContract true "Schema modification payload"
// @Success 200 {object} logic.SchemaDBSerializerStruct "Modified schema"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Schema not found"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/schemas/{schemaID} [put]
func ModifySchemaV1(c *gin.Context) {
	var input input_contracts.ModifySchemaApiInputContract

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	schemaID := c.Param("schemaID")
	parsedSchemaID, _ := uuid.Parse(schemaID)

	userID, _ := c.Get("UserID")

	schemasManager := logic.SchemaObjectsManager{}
	schema, err := schemasManager.GetByID(parsedSchemaID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Schema is found, verify user has access to its project
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(schema.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Create a new version of the schema
	schema, err = schema.CreateANewVersion(input.Schema, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	c.JSON(http.StatusOK, schema.Serialize())
}

// Get list of schema versions
// @Summary Get list of schema versions
// @Description Get list of schema versions
// @Produce json
// @Tags Schemas
// @Security BearerAuth
// @Param schemaID path string true "Schema ID"
// @Success 200 {array} logic.SchemaEditDBSerializerStruct "List of schema versions"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Schema not found"
// @Router /v1/protected/schemas/{schemaID}/versions [get]
func GetSchemaVersionsV1(c *gin.Context) {
	schemaID := c.Param("schemaID")
	parsedSchemaID, _ := uuid.Parse(schemaID)

	schemasManager := logic.SchemaObjectsManager{}
	schema, err := schemasManager.GetByID(parsedSchemaID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Schema is found, verify user has access to its project
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(schema.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Get all versions of the schema
	versionsObjects := schema.GetAllVersions()
	var response []logic.SchemaEditDBSerializerStruct

	for _, version := range versionsObjects {
		response = append(response, *version.SerializeLong())
	}

	if len(response) == 0 {
		c.JSON(http.StatusOK, make([]logic.SchemaEditDBSerializerStruct, 0))
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// Get a single schema version
// @Summary Get a single schema version
// @Description Get a single schema version
// @Produce json
// @Tags Schemas
// @Security BearerAuth
// @Param schemaID path string true "Schema ID"
// @Param versionID path string true "Version ID (integer number)"
// @Success 200 {object} logic.SchemaEditDBSerializerStruct "Schema version information"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Schema version not found or can't be accessed by user"
// @Router /v1/protected/schemas/{schemaID}/versions/{versionID} [get]
func GetSingleSchemaVersionsV1(c *gin.Context) {
	schemaID := c.Param("schemaID")
	parsedSchemaID, _ := uuid.Parse(schemaID)

	versionID := c.Param("versionID")
	parsedVersionID, err := strconv.ParseInt(versionID, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	schemasManager := logic.SchemaObjectsManager{}
	schemaVersion, err := schemasManager.GetSpecificVersionOfSchema(parsedSchemaID, int(parsedVersionID))

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Verify the schema exists and user has access to its project
	schema, err := schemasManager.GetByID(parsedSchemaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(schema.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, schemaVersion.SerializeLong())
}
