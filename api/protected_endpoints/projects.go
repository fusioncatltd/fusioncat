package protected_endpoints

import (
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func ProjectsProtectedRoutesV1(router *gin.RouterGroup) {
	router.POST("/projects", CreateNewProjectV1)
	router.GET("/projects", GetAllProjectsV1)
	router.GET("/projects/:id", GetSingleProjectV1)
	router.POST("/projects/:id/imports", ImportProjectArchitectureV1)
	router.POST("/projects/:id/imports/validator", ValidateArchitectureFileV1)
}

// Get information about a single project
// @Summary Get information about a single project
// @Description Get information about a single project
// @Produce json
// @Tags Projects
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} logic.ProjectDBSerializerStruct "Response with project information"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Router /v1/protected/projects/{id} [get]
func GetSingleProjectV1(c *gin.Context) {
	id := c.Param("id")
	parsedID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	projectObject, projectError := projectsManager.GetByID(parsedID)

	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	c.JSON(http.StatusOK, projectObject.Serialize())
}

// Get information about projects I am a member of
// @Summary Get information about projects I am a member of
// @Description Get information about projects I am a member of
// @Produce json
// @Tags Projects
// @Security BearerAuth
// @Success 200 {array} logic.ProjectDBSerializerStruct "Response with projects information"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Router /v1/protected/projects [get]
func GetAllProjectsV1(c *gin.Context) {
	var allProjects []logic.ProjectObject
	response := make([]logic.ProjectDBSerializerStruct, 0)
	projectsManager := logic.ProjectsObjectsManager{}
	userID, _ := c.Get("UserID")

	allProjects, _ = projectsManager.GetAllProjects(userID.(uuid.UUID))

	for _, project := range allProjects {
		serializedProject := project.Serialize()
		response = append(response, *serializedProject)
	}

	c.JSON(http.StatusOK, response)
}

// Create a new project
// @Summary Create a new project
// @Description Create a new project
// @Produce json
// @Tags Projects
// @Security BearerAuth
// @Param project body input_contracts.CreateModifyProjectApiInputContract true "Project create request payload"
// @Success 200 {object} logic.ProjectDBSerializerStruct "New project has been created"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 409 {object} map[string]string "Project with this name already exists"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects [post]
func CreateNewProjectV1(c *gin.Context) {
	var inputWithOwnerData input_contracts.CreateModifyProjectApiInputContract

	if err := c.ShouldBindJSON(&inputWithOwnerData); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	userID, _ := c.Get("UserID")
	manager := logic.ProjectsObjectsManager{}

	if manager.CheckIfProjectWithSpecificNameExists(inputWithOwnerData.Name) {
		c.AbortWithStatusJSON(http.StatusConflict, nil)
		return
	}

	projectObject, _ := manager.CreateANewProject(
		inputWithOwnerData.Name,
		inputWithOwnerData.Description,
		userID.(uuid.UUID),
	)

	c.JSON(http.StatusOK, projectObject.Serialize())
}

// Import a project architecture
// @Summary Import a project architecture
// @Description Import project architecture including servers, resources, schemas, messages, and apps from YAML
// @Produce json
// @Accept json
// @Tags Projects
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param import body input_contracts.ImportFileInputContract true "YAML content to import"
// @Success 200 {object} map[string]string "Import successful"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 409 {object} map[string]interface{} "Import validation errors"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects/{id}/imports [post]
func ImportProjectArchitectureV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.ImportFileInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	// Validate the YAML
	validationErrors := logic.ValidateProjectImportYAML(input.YAML, parsedProjectID)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusConflict, gin.H{"errors": validationErrors})
		return
	}

	// Import the project architecture
	userID, _ := c.Get("UserID")
	importError := logic.ImportProjectFromYAML(input.YAML, parsedProjectID, userID.(uuid.UUID))
	if importError != nil {
		c.JSON(http.StatusConflict, gin.H{"error": importError.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Import completed successfully"})
}

// Validate architecture file
// @Summary Validate architecture file
// @Description Validate YAML file structure for project import
// @Produce json
// @Accept json
// @Tags Projects
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param import body input_contracts.ImportFileInputContract true "YAML content to validate"
// @Success 200 {object} map[string]string "File is valid"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 409 {object} map[string]interface{} "Validation errors"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects/{id}/imports/validator [post]
func ValidateArchitectureFileV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.ImportFileInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	// Validate the YAML
	validationErrors := logic.ValidateProjectImportYAML(input.YAML, parsedProjectID)
	if len(validationErrors) > 0 {
		c.JSON(http.StatusConflict, gin.H{"errors": validationErrors})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "YAML is valid"})
}
