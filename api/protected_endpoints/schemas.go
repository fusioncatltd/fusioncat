package protected_endpoints

import (
	"fmt"
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func SchemasProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/projects/:id/schemas", GetAllSchemasInProjectV1)
	router.POST("/projects/:id/schemas", NewSchemaInProjectV1)
	//	router.GET("/schemas/:schemaID", GetSingleSchemaV1)
	//	router.GET("/schemas/:schemaID/versions", GetSchemaVersionsV1)
	//	router.GET("/schemas/:schemaID/versions/:versionID", GetSingleSchemaVersionsV1)
	//	router.PUT("/schemas/:schemaID", ModifySchemaV1)
	//	router.GET("/schemas/:schemaID/usage", GetASingleSchemaUsageV1)
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
