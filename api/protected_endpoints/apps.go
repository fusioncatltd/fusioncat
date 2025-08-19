package protected_endpoints

import (
	"net/http"
	"strings"

	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AppsProtectedRoutesV1(router *gin.RouterGroup) {
	router.POST("/projects/:id/apps", CreateAppV1)
	router.GET("/projects/:id/apps", GetAppsV1)
	router.GET("/apps/:id/usage", GetAppUsageV1)
	router.GET("/apps/:id/code/:language", GetAppGeneratedCodeV1)
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

// Get app usage information
// @Summary Get app usage information
// @Description Get information about app's connections to resources, servers, and messages
// @Produce json
// @Tags Apps
// @Security BearerAuth
// @Param id path string true "App ID"
// @Success 200 {object} logic.AppUsageMatrixResponse "App usage information"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "App not found"
// @Router /v1/protected/apps/{id}/usage [get]
func GetAppUsageV1(c *gin.Context) {
	id := c.Param("id")
	parsedAppID, _ := uuid.Parse(id)

	// Verify the app exists
	appsManager := logic.AppsObjectsManager{}
	app, err := appsManager.GetByID(parsedAppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	// Get project information
	projectsManager := logic.ProjectsObjectsManager{}
	_, err = projectsManager.GetByID(uuid.MustParse(app.Serialize().ProjectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get app usage matrix
	usage, err := appsManager.GetAppUsageMatrix(parsedAppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve app usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// Get app generated code
// @Summary Get full application code in Go
// @Description Generate complete application code including schemas, messages, resources, servers and app structure. The generated code will be returned as a plain text file with appropriate content type headers.
// @Produce text/plain
// @Tags Apps
// @Security BearerAuth
// @Param id path string true "App ID"
// @Param language path string true "Programming language (currently only 'go' is supported)"
// @Success 200 {string} string "Generated application code as plain text file"
// @Failure 400 {object} map[string]string "Invalid app ID or unsupported language"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "App not found"
// @Failure 500 {object} map[string]string "Internal server error during code generation"
// @Router /v1/protected/apps/{id}/code/{language} [get]
func GetAppGeneratedCodeV1(c *gin.Context) {
	language := c.Param("language")
	if language != "go" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only Go language is supported for full application code generation"})
		return
	}

	id := c.Param("id")
	parsedAppID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid app ID"})
		return
	}

	// Get the app
	appsManager := logic.AppsObjectsManager{}
	app, err := appsManager.GetByID(parsedAppID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App not found"})
		return
	}

	// Get project for authorization check
	projectsManager := logic.ProjectsObjectsManager{}
	_, err = projectsManager.GetByID(uuid.MustParse(app.Serialize().ProjectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get app usage to understand message connections
	appUsageMatrix, err := appsManager.GetAppUsageMatrix(parsedAppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get app usage matrix"})
		return
	}

	// Generate the complete Go code
	generatedCode, err := generateAppCode(app, appUsageMatrix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code: " + err.Error()})
		return
	}

	// Return the generated code
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=generated_app.go")
	c.String(http.StatusOK, generatedCode)
}

// Helper function to capitalize first letter
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Helper function to convert snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		result += capitalizeFirst(part)
	}
	return result
}

