package protected_endpoints

import (
	"net/http"

	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AppsProtectedRoutesV1(router *gin.RouterGroup) {
	router.POST("/projects/:id/apps", CreateAppV1)
	router.GET("/projects/:id/apps", GetAppsV1)
}

// Create a new application in project
// @Summary Create a new application in project
// @Description Create a new application in project
// @Produce json
// @Tags Apps
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param app body input_contracts.CreateAppApiInputContract true "App create request payload"
// @Success 200 {object} logic.AppDBSerializerStruct "App created"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 409 {object} map[string]string "App with this name already exists in this project"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects/{id}/apps [post]
func CreateAppV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.CreateAppApiInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	appsManager := logic.AppsObjectsManager{}
	if !appsManager.CanNameBeUsed(input.Name, parsedProjectID) {
		c.JSON(http.StatusConflict, gin.H{"error": "App with this name already exists in this project"})
		return
	}

	userID, _ := c.Get("UserID")
	app, _ := appsManager.CreateANewApp(input.Name, input.Description, parsedProjectID, userID.(uuid.UUID))

	c.JSON(http.StatusOK, app.Serialize())
}

// Get all applications in a project
// @Summary Get all applications in a project
// @Description Get all applications in a project
// @Produce json
// @Tags Apps
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {array} logic.AppDBSerializerStruct "List of apps"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Router /v1/protected/projects/{id}/apps [get]
func GetAppsV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	appsManager := logic.AppsObjectsManager{}
	apps, err := appsManager.GetAllAppsForProject(parsedProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve apps"})
		return
	}

	var serializedApps []*logic.AppDBSerializerStruct
	for _, app := range apps {
		serializedApps = append(serializedApps, app.Serialize())
	}

	c.JSON(http.StatusOK, serializedApps)
}