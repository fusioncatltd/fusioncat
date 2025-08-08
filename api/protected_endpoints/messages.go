package protected_endpoints

import (
	"github.com/fusioncatltd/fusioncat/api"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func MessagesProtectedRoutesV1(router *gin.RouterGroup) {
	router.GET("/projects/:id/messages", GetAllMessagesInProjectV1)
	router.POST("/projects/:id/messages", NewMessageV1)
}

// Get all messages in project
// @Summary Get messages in project
// @Description Get list of messages in a project
// @Produce json
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {array} logic.MessageDBSerializerStruct "List of messages in project"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/protected/projects/{id}/messages [get]
func GetAllMessagesInProjectV1(c *gin.Context) {
	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)

	// Verify project exists and user has access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Get all messages in the project
	messagesManager := logic.MessagesObjectsManager{}
	allMessages, err := messagesManager.GetAllMessagesInProject(parsedProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Serialize the response
	var response []logic.MessageDBSerializerStruct
	for _, message := range allMessages {
		serializedMessage := message.Serialize()
		response = append(response, *serializedMessage)
	}

	if len(response) == 0 {
		c.JSON(http.StatusOK, make([]logic.MessageDBSerializerStruct, 0))
	} else {
		c.JSON(http.StatusOK, response)
	}
}

// Create new message in project
// @Summary Create message
// @Description Create a new message in a project
// @Produce json
// @Accept json
// @Tags Messages
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param message body input_contracts.CreateMessageApiInputContract true "Message creation payload"
// @Success 200 {object} logic.MessageDBSerializerStruct "Created message"
// @Failure 400 {object} map[string]string "Schema does not belong to this project or schema version does not exist"
// @Failure 401 {object} map[string]string "Access denied: missing or invalid Authorization header"
// @Failure 404 {object} map[string]string "Project or schema not found"
// @Failure 409 {object} map[string]string "Message with this name already exists in this project"
// @Failure 422 {object} api.DataValidationErrorAPIResponse "JSON payload validation errors"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/protected/projects/{id}/messages [post]
func NewMessageV1(c *gin.Context) {
	var input input_contracts.CreateMessageApiInputContract

	// Bind JSON input to the input struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, api.GetValidationErrors(err))
		return
	}

	id := c.Param("id")
	parsedProjectID, _ := uuid.Parse(id)
	userID, _ := c.Get("UserID")

	// Verify project exists and user has access
	projectsManager := logic.ProjectsObjectsManager{}
	_, projectError := projectsManager.GetByID(parsedProjectID)
	if projectError != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}

	// Check if message name is already taken in this project
	messagesManager := logic.MessagesObjectsManager{}
	if !messagesManager.CanNameBeUsed(input.Name, parsedProjectID) {
		c.JSON(http.StatusConflict, gin.H{"error": "Message with this name already exists in this project"})
		return
	}

	// Parse schema ID
	parsedSchemaID, err := uuid.Parse(input.SchemaID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid schema ID format"})
		return
	}

	// Verify the schema exists and belongs to the project
	schemasManager := logic.SchemaObjectsManager{}
	schema, err := schemasManager.GetByID(parsedSchemaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schema not found"})
		return
	}

	if schema.GetProjectID() != parsedProjectID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schema does not belong to this project"})
		return
	}

	// Verify the schema version exists
	if !schemasManager.SchemaWithVersionExists(parsedSchemaID, input.SchemaVersion) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schema version does not exist"})
		return
	}

	// Create the new message
	message, err := messagesManager.CreateANewMessage(
		input.Description,
		userID.(uuid.UUID),
		parsedProjectID,
		input.Name,
		parsedSchemaID,
		input.SchemaVersion,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, message.Serialize())
}