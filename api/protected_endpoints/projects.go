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
