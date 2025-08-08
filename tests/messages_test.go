package tests

import (
	"encoding/json"
	"fmt"
	"github.com/fusioncatltd/fusioncat/api/input_contracts"
	"github.com/fusioncatltd/fusioncat/logic"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestMessagesEndpoints(t *testing.T) {
	// Clean database before running test
	CleanDatabase(t)
	
	h := os.Getenv("TESTSERVER_URL")
	e := httpexpect.Default(t, h)

	// Create a user
	userEmail := fmt.Sprintf("test-messages-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	userPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    userEmail,
		Password: "TestPassword123",
	}

	// Sign up and get bearer token
	signUpResponse := e.POST("/v1/public/users").
		WithJSON(userPayload).
		Expect().
		Status(http.StatusOK)

	bearerToken := signUpResponse.Raw().Header.Get("Authorization")
	require.NotEmpty(t, bearerToken, "Bearer token should not be empty")

	// Create a project
	projectName := fmt.Sprintf("MessagesTestProject%d", time.Now().UnixNano())
	projectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        projectName,
		Description: "Project for testing messages",
	}

	projectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", bearerToken).
		WithJSON(projectPayload).
		Expect().
		Status(http.StatusOK)

	var createdProject logic.ProjectDBSerializerStruct
	rawProjectReader := projectResponse.Raw().Body
	defer rawProjectReader.Close()
	rawProjectBytes, _ := io.ReadAll(rawProjectReader)

	require.NoError(t, json.Unmarshal(rawProjectBytes, &createdProject))
	projectID := createdProject.ID

	// Create a schema that messages will reference
	validSchemaContent, err := ReadTestFileString("jsonschemas/validSchema1.json")
	require.NoError(t, err, "Should be able to read valid schema file")

	schemaName := fmt.Sprintf("MessageSchema%d", time.Now().UnixNano())
	schemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        schemaName,
		Description: "Schema for message testing",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	schemaResponse := e.POST("/v1/protected/projects/" + projectID + "/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(schemaPayload).
		Expect().
		Status(http.StatusOK)

	var createdSchema logic.SchemaDBSerializerStruct
	rawSchemaReader := schemaResponse.Raw().Body
	defer rawSchemaReader.Close()
	rawSchemaBytes, _ := io.ReadAll(rawSchemaReader)

	require.NoError(t, json.Unmarshal(rawSchemaBytes, &createdSchema))
	schemaID := createdSchema.ID

	// Test 1: Get messages list - should be empty initially
	messagesListResponse := e.GET("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var initialMessages []logic.MessageDBSerializerStruct
	rawMessagesReader := messagesListResponse.Raw().Body
	defer rawMessagesReader.Close()
	rawMessagesBytes, _ := io.ReadAll(rawMessagesReader)

	require.NoError(t, json.Unmarshal(rawMessagesBytes, &initialMessages))
	require.Empty(t, initialMessages, "Project should have no messages initially")

	// Test 2: Create a message
	messageName := fmt.Sprintf("TestMessage%d", time.Now().UnixNano())
	createMessagePayload := input_contracts.CreateMessageApiInputContract{
		Name:          messageName,
		Description:   "Test message for data exchange",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	createMessageResponse := e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(createMessagePayload).
		Expect().
		Status(http.StatusOK)

	var createdMessage logic.MessageDBSerializerStruct
	rawCreatedMessageReader := createMessageResponse.Raw().Body
	defer rawCreatedMessageReader.Close()
	rawCreatedMessageBytes, _ := io.ReadAll(rawCreatedMessageReader)

	require.NoError(t, json.Unmarshal(rawCreatedMessageBytes, &createdMessage))
	require.Equal(t, messageName, createdMessage.Name)
	require.Equal(t, "Test message for data exchange", createdMessage.Description)
	require.Equal(t, schemaID, createdMessage.SchemaID)
	require.Equal(t, 1, createdMessage.SchemaVersion)
	require.Equal(t, projectID, createdMessage.ProjectID)

	// Test 3: Get messages list - should now have one message
	messagesListAfterCreateResponse := e.GET("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var messagesAfterCreate []logic.MessageDBSerializerStruct
	rawMessagesAfterReader := messagesListAfterCreateResponse.Raw().Body
	defer rawMessagesAfterReader.Close()
	rawMessagesAfterBytes, _ := io.ReadAll(rawMessagesAfterReader)

	require.NoError(t, json.Unmarshal(rawMessagesAfterBytes, &messagesAfterCreate))
	require.Len(t, messagesAfterCreate, 1, "Project should have one message after creation")
	require.Equal(t, messageName, messagesAfterCreate[0].Name)

	// Test 4: Try to create a message with duplicate name (should fail with 409)
	duplicateMessagePayload := input_contracts.CreateMessageApiInputContract{
		Name:          messageName, // Same name as before
		Description:   "Duplicate message attempt",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	_ = e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(duplicateMessagePayload).
		Expect().
		Status(http.StatusConflict)

	// Test 5: Try to create a message in non-existent project
	_ = e.POST("/v1/protected/projects/00000000-0000-0000-0000-000000000000/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(createMessagePayload).
		Expect().
		Status(http.StatusNotFound)

	// Test 6: Try to get messages from non-existent project
	_ = e.GET("/v1/protected/projects/00000000-0000-0000-0000-000000000000/messages").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusNotFound)

	// Test 7: Create a second message
	secondMessageName := fmt.Sprintf("SecondMessage%d", time.Now().UnixNano())
	secondMessagePayload := input_contracts.CreateMessageApiInputContract{
		Name:          secondMessageName,
		Description:   "Second test message",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	secondMessageResponse := e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(secondMessagePayload).
		Expect().
		Status(http.StatusOK)

	var secondCreatedMessage logic.MessageDBSerializerStruct
	rawSecondMessageReader := secondMessageResponse.Raw().Body
	defer rawSecondMessageReader.Close()
	rawSecondMessageBytes, _ := io.ReadAll(rawSecondMessageReader)

	require.NoError(t, json.Unmarshal(rawSecondMessageBytes, &secondCreatedMessage))
	require.Equal(t, secondMessageName, secondCreatedMessage.Name)

	// Test 8: Get messages list - should now have two messages
	finalMessagesListResponse := e.GET("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		Expect().
		Status(http.StatusOK)

	var finalMessages []logic.MessageDBSerializerStruct
	rawFinalMessagesReader := finalMessagesListResponse.Raw().Body
	defer rawFinalMessagesReader.Close()
	rawFinalMessagesBytes, _ := io.ReadAll(rawFinalMessagesReader)

	require.NoError(t, json.Unmarshal(rawFinalMessagesBytes, &finalMessages))
	require.Len(t, finalMessages, 2, "Project should have two messages after creating second one")

	// Verify both messages are in the list
	messageNames := make(map[string]bool)
	for _, message := range finalMessages {
		messageNames[message.Name] = true
	}
	require.True(t, messageNames[messageName], "First message should be in the list")
	require.True(t, messageNames[secondMessageName], "Second message should be in the list")

	// Test 9: Try to create a message with non-existent schema
	invalidSchemaPayload := input_contracts.CreateMessageApiInputContract{
		Name:          "InvalidSchemaMessage",
		Description:   "Message with non-existent schema",
		SchemaID:      "00000000-0000-0000-0000-000000000000",
		SchemaVersion: 1,
	}

	_ = e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidSchemaPayload).
		Expect().
		Status(http.StatusNotFound)

	// Test 10: Modify schema to create version 2
	modifiedSchemaContent := `{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title": "Person",
		"type": "object",
		"properties": {
			"firstName": {
				"type": "string"
			},
			"lastName": {
				"type": "string"
			},
			"age": {
				"type": "integer",
				"minimum": 0
			},
			"email": {
				"type": "string",
				"format": "email"
			}
		},
		"required": ["firstName", "lastName", "email"]
	}`

	modifySchemaPayload := input_contracts.ModifySchemaApiInputContract{
		Schema: modifiedSchemaContent,
	}

	_ = e.PUT("/v1/protected/schemas/" + schemaID).
		WithHeader("Authorization", bearerToken).
		WithJSON(modifySchemaPayload).
		Expect().
		Status(http.StatusOK)

	// Test 11: Create a message with schema version 2
	messageWithV2Name := fmt.Sprintf("MessageWithV2_%d", time.Now().UnixNano())
	messageWithV2Payload := input_contracts.CreateMessageApiInputContract{
		Name:          messageWithV2Name,
		Description:   "Message using schema version 2",
		SchemaID:      schemaID,
		SchemaVersion: 2,
	}

	messageWithV2Response := e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(messageWithV2Payload).
		Expect().
		Status(http.StatusOK)

	var messageWithV2 logic.MessageDBSerializerStruct
	rawMessageWithV2Reader := messageWithV2Response.Raw().Body
	defer rawMessageWithV2Reader.Close()
	rawMessageWithV2Bytes, _ := io.ReadAll(rawMessageWithV2Reader)

	require.NoError(t, json.Unmarshal(rawMessageWithV2Bytes, &messageWithV2))
	require.Equal(t, 2, messageWithV2.SchemaVersion, "Message should use schema version 2")

	// Test 12: Try to create a message with non-existent schema version
	invalidVersionPayload := input_contracts.CreateMessageApiInputContract{
		Name:          "InvalidVersionMessage",
		Description:   "Message with invalid schema version",
		SchemaID:      schemaID,
		SchemaVersion: 99, // Non-existent version
	}

	_ = e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidVersionPayload).
		Expect().
		Status(http.StatusBadRequest)

	// Test 13: Create a schema in a different project and try to use it
	secondProjectName := fmt.Sprintf("SecondProject%d", time.Now().UnixNano())
	secondProjectPayload := input_contracts.CreateModifyProjectApiInputContract{
		Name:        secondProjectName,
		Description: "Second project for testing",
	}

	secondProjectResponse := e.POST("/v1/protected/projects").
		WithHeader("Authorization", bearerToken).
		WithJSON(secondProjectPayload).
		Expect().
		Status(http.StatusOK)

	var secondProject logic.ProjectDBSerializerStruct
	rawSecondProjectReader := secondProjectResponse.Raw().Body
	defer rawSecondProjectReader.Close()
	rawSecondProjectBytes, _ := io.ReadAll(rawSecondProjectReader)

	require.NoError(t, json.Unmarshal(rawSecondProjectBytes, &secondProject))
	secondProjectID := secondProject.ID

	secondSchemaName := fmt.Sprintf("SecondProjectSchema%d", time.Now().UnixNano())
	secondSchemaPayload := input_contracts.CreateSchemaApiInputContract{
		Name:        secondSchemaName,
		Description: "Schema in second project",
		Type:        "jsonschema",
		Schema:      validSchemaContent,
	}

	secondSchemaResponse := e.POST("/v1/protected/projects/" + secondProjectID + "/schemas").
		WithHeader("Authorization", bearerToken).
		WithJSON(secondSchemaPayload).
		Expect().
		Status(http.StatusOK)

	var secondSchema logic.SchemaDBSerializerStruct
	rawSecondSchemaReader := secondSchemaResponse.Raw().Body
	defer rawSecondSchemaReader.Close()
	rawSecondSchemaBytes, _ := io.ReadAll(rawSecondSchemaReader)

	require.NoError(t, json.Unmarshal(rawSecondSchemaBytes, &secondSchema))

	// Test 14: Try to create a message in first project using schema from second project
	crossProjectMessagePayload := input_contracts.CreateMessageApiInputContract{
		Name:          "CrossProjectMessage",
		Description:   "Message trying to use schema from different project",
		SchemaID:      secondSchema.ID,
		SchemaVersion: 1,
	}

	_ = e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(crossProjectMessagePayload).
		Expect().
		Status(http.StatusBadRequest)

	// Test 15: Test with invalid message name (contains special characters)
	invalidNamePayload := input_contracts.CreateMessageApiInputContract{
		Name:          "Invalid-Message-Name!",
		Description:   "Message with invalid name",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	_ = e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(invalidNamePayload).
		Expect().
		Status(http.StatusUnprocessableEntity)

	// Test 16: Test with valid alphanumeric name with underscore
	validNamePayload := input_contracts.CreateMessageApiInputContract{
		Name:          "Valid_Message_123",
		Description:   "Message with valid name containing underscore",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	validNameResponse := e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", bearerToken).
		WithJSON(validNamePayload).
		Expect().
		Status(http.StatusOK)

	var validNameMessage logic.MessageDBSerializerStruct
	rawValidNameReader := validNameResponse.Raw().Body
	defer rawValidNameReader.Close()
	rawValidNameBytes, _ := io.ReadAll(rawValidNameReader)

	require.NoError(t, json.Unmarshal(rawValidNameBytes, &validNameMessage))
	require.Equal(t, "Valid_Message_123", validNameMessage.Name)

	// Test 17: Second user can access messages in shared project
	secondUserEmail := fmt.Sprintf("test-messages-user2-%s@mail.com", strconv.FormatInt(time.Now().UnixNano(), 10))
	secondUserPayload := input_contracts.SignInSignUpApiInputContract{
		Email:    secondUserEmail,
		Password: "SecondUserPass123",
	}

	secondUserSignUpResponse := e.POST("/v1/public/users").
		WithJSON(secondUserPayload).
		Expect().
		Status(http.StatusOK)

	secondUserBearer := secondUserSignUpResponse.Raw().Header.Get("Authorization")

	// Second user should be able to get messages from the project
	secondUserMessagesResponse := e.GET("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", secondUserBearer).
		Expect().
		Status(http.StatusOK)

	var secondUserMessages []logic.MessageDBSerializerStruct
	rawSecondUserMessagesReader := secondUserMessagesResponse.Raw().Body
	defer rawSecondUserMessagesReader.Close()
	rawSecondUserMessagesBytes, _ := io.ReadAll(rawSecondUserMessagesReader)

	require.NoError(t, json.Unmarshal(rawSecondUserMessagesBytes, &secondUserMessages))
	require.GreaterOrEqual(t, len(secondUserMessages), 4, "Second user should see all messages in the project")

	// Second user should be able to create a message too
	secondUserMessageName := fmt.Sprintf("SecondUserMessage%d", time.Now().UnixNano())
	secondUserMessagePayload := input_contracts.CreateMessageApiInputContract{
		Name:          secondUserMessageName,
		Description:   "Message created by second user",
		SchemaID:      schemaID,
		SchemaVersion: 1,
	}

	secondUserCreateResponse := e.POST("/v1/protected/projects/" + projectID + "/messages").
		WithHeader("Authorization", secondUserBearer).
		WithJSON(secondUserMessagePayload).
		Expect().
		Status(http.StatusOK)

	var secondUserCreatedMessage logic.MessageDBSerializerStruct
	rawSecondUserCreatedReader := secondUserCreateResponse.Raw().Body
	defer rawSecondUserCreatedReader.Close()
	rawSecondUserCreatedBytes, _ := io.ReadAll(rawSecondUserCreatedReader)

	require.NoError(t, json.Unmarshal(rawSecondUserCreatedBytes, &secondUserCreatedMessage))
	require.Equal(t, secondUserMessageName, secondUserCreatedMessage.Name)
}