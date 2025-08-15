package protected_endpoints

import (
	"net/http"

	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ServersProtectedRoutesV1(router *gin.RouterGroup) {
	router.POST("/projects/:id/servers", CreateServerV1)
	router.GET("/projects/:id/servers", GetServersV1)
	router.POST("/servers/:id/resources", CreateResourceV1)
	router.GET("/servers/:id/resources", GetResourcesV1)
	router.POST("/servers/:id/binds", CreateResourceBindV1)
	router.GET("/servers/:id/binds", GetResourceBindsV1)
}

// Create a new server in project
// @Summary Create a new server in project
// @Description Create a new server in project
// @Produce json
// @Tags Servers
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param server body input_contracts.CreateServerApiInputContract true "Server create request payload"
// @Success 200 {object} logic.ServerDBSerializerStruct "Server created"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 409 {object} map[string]string "Server with this name already exists in this project"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/projects/{id}/servers [post]
func CreateServerV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.CreateServerApiInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	serversManager := logic.ServersObjectsManager{}
	if !serversManager.CanNameBeUsed(input.Name, parsedProjectID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Server with this name already exists in this project"})
		return
	}

	userID, _ := c.Get("UserID")
	server, _ := serversManager.CreateANewServer(input.Name, input.Description, input.Protocol, parsedProjectID, userID.(uuid.UUID))

	c.JSON(http.StatusOK, server.Serialize())
}

// Get all servers in a project
// @Summary Get all servers in a project
// @Description Get all servers in a project
// @Produce json
// @Tags Servers
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {array} logic.ServerDBSerializerStruct "List of servers"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Router /v1/protected/projects/{id}/servers [get]
func GetServersV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	serversManager := logic.ServersObjectsManager{}
	servers, err := serversManager.GetAllServersForProject(parsedProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve servers"})
		return
	}

	var serializedServers []*logic.ServerDBSerializerStruct
	for _, server := range servers {
		serializedServers = append(serializedServers, server.Serialize())
	}

	c.JSON(http.StatusOK, serializedServers)
}

// Create a new resource in server
// @Summary Create a new resource in server
// @Description Create a new resource in server
// @Produce json
// @Tags Server Resources
// @Security BearerAuth
// @Param id path string true "Server ID"
// @Param resource body input_contracts.CreateResourceApiInputContract true "Resource create request payload"
// @Success 200 {object} logic.ResourceDBSerializerStruct "Resource created"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Server not found"
// @Failure 409 {object} map[string]string "Resource with this name already exists in this server"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/servers/{id}/resources [post]
func CreateResourceV1(c *gin.Context) {
	id := c.Param("id")
	parsedServerID, _ := uuid.Parse(id)

	serversManager := logic.ServersObjectsManager{}
	server, serverError := serversManager.GetByID(parsedServerID)
	if serverError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify project access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(server.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.CreateResourceApiInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	resourcesManager := logic.ResourcesObjectsManager{}
	if !resourcesManager.CanNameBeUsed(input.Name, parsedServerID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Resource with this name already exists in this server"})
		return
	}

	userID, _ := c.Get("UserID")
	resource, _ := resourcesManager.CreateANewResource(
		parsedServerID,
		server.GetProjectID(),
		input.Name,
		input.Mode,
		input.ResourceType,
		input.Description,
		userID.(uuid.UUID))

	c.JSON(http.StatusOK, resource.Serialize())
}

// Get all resources in a server
// @Summary Get all resources in a server
// @Description Get all resources in a server
// @Produce json
// @Tags Server Resources
// @Security BearerAuth
// @Param id path string true "Server ID"
// @Success 200 {array} logic.ResourceDBSerializerStruct "List of resources"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Server not found"
// @Router /v1/protected/servers/{id}/resources [get]
func GetResourcesV1(c *gin.Context) {
	id := c.Param("id")
	parsedServerID, _ := uuid.Parse(id)

	serversManager := logic.ServersObjectsManager{}
	server, serverError := serversManager.GetByID(parsedServerID)
	if serverError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify project access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(server.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	resourcesManager := logic.ResourcesObjectsManager{}
	resources, err := resourcesManager.GetAllResourcesForServer(parsedServerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve resources"})
		return
	}

	var serializedResources []*logic.ResourceDBSerializerStruct
	for _, resource := range resources {
		serializedResources = append(serializedResources, resource.Serialize())
	}

	c.JSON(http.StatusOK, serializedResources)
}

// Create a resource binding in server
// @Summary Create a resource binding in server
// @Description Create a resource binding between two resources in the same server
// @Produce json
// @Tags Server Resources
// @Security BearerAuth
// @Param id path string true "Server ID"
// @Param binding body input_contracts.CreateResourceBindApiInputContract true "Resource binding request payload"
// @Success 200 {object} logic.ResourceBindingDBSerializerStruct "Resource binding created"
// @Failure 400 {object} map[string]string "Invalid input or resources not in same server"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Server or resource not found"
// @Failure 409 {object} map[string]string "Resource binding already exists"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Router /v1/protected/servers/{id}/binds [post]
func CreateResourceBindV1(c *gin.Context) {
	id := c.Param("id")
	parsedServerID, _ := uuid.Parse(id)

	serversManager := logic.ServersObjectsManager{}
	server, serverError := serversManager.GetByID(parsedServerID)
	if serverError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify project access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(server.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input input_contracts.CreateResourceBindApiInputContract

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	// Parse resource IDs
	sourceResourceID, err := uuid.Parse(input.SourceResourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source resource ID"})
		return
	}
	targetResourceID, err := uuid.Parse(input.TargetResourceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target resource ID"})
		return
	}

	// Verify both resources exist and belong to the same server
	resourcesManager := logic.ResourcesObjectsManager{}
	sourceResource, err := resourcesManager.GetByID(sourceResourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Source resource not found"})
		return
	}
	targetResource, err := resourcesManager.GetByID(targetResourceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target resource not found"})
		return
	}

	if sourceResource.GetServerID() != parsedServerID || targetResource.GetServerID() != parsedServerID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resources must belong to the same server"})
		return
	}

	// Create the binding
	bindingsManager := logic.ResourceBindingsObjectsManager{}

	// Check if binding already exists
	if bindingsManager.CheckIfBindingExists(sourceResourceID, targetResourceID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Resource binding already exists"})
		return
	}

	binding, err := bindingsManager.CreateABinding(sourceResourceID, targetResourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, binding.Serialize())
}

// Get all resource bindings in a server
// @Summary Get all resource bindings in a server
// @Description Get all resource bindings for resources in a server
// @Produce json
// @Tags Server Resources
// @Security BearerAuth
// @Param id path string true "Server ID"
// @Success 200 {array} logic.ResourceBindingDBSerializerStruct "List of resource bindings"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Server not found"
// @Router /v1/protected/servers/{id}/binds [get]
func GetResourceBindsV1(c *gin.Context) {
	id := c.Param("id")
	parsedServerID, _ := uuid.Parse(id)

	serversManager := logic.ServersObjectsManager{}
	server, serverError := serversManager.GetByID(parsedServerID)
	if serverError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}

	// Verify project access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(server.GetProjectID())
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Get all resources for the server
	resourcesManager := logic.ResourcesObjectsManager{}
	resources, err := resourcesManager.GetAllResourcesForServer(parsedServerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve resources"})
		return
	}

	// Get all bindings for each resource and use a map to ensure uniqueness
	bindingsManager := logic.ResourceBindingsObjectsManager{}
	uniqueBindings := make(map[string]*logic.ResourceBindingDBSerializerStruct)

	for _, resource := range resources {
		bindings, err := bindingsManager.GetAllBindingsForResource(resource.GetID())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bindings"})
			return
		}
		for _, binding := range bindings {
			serialized := binding.Serialize()
			uniqueBindings[serialized.ID] = serialized
		}
	}

	// Convert map values to slice
	var allBindings []*logic.ResourceBindingDBSerializerStruct
	for _, binding := range uniqueBindings {
		allBindings = append(allBindings, binding)
	}

	c.JSON(http.StatusOK, allBindings)
}