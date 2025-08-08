package logic

import (
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
)

type MessageObject struct {
	dbModel db.MessagesDBModel
}

type MessagesObjectsManager struct{}

type MessageDBSerializerStruct struct {
	ID            string `json:"id"`
	ProjectID     string `json:"project_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	SchemaID      string `json:"schema_id"`
	SchemaVersion int    `json:"schema_version"`
	CreatedByID   string `json:"created_by_id"`
	CreatedByName string `json:"created_by_name"`
	CreatedAt     string `json:"created_at"`
}

// Serialize converts a MessageObject to its serialized form
func (message *MessageObject) Serialize() *MessageDBSerializerStruct {
	createdByName := ""
	
	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, message.dbModel.CreatedByID)
	createdByName = userDbRecord.Handle
	
	return &MessageDBSerializerStruct{
		ID:            message.dbModel.ID.String(),
		ProjectID:     message.dbModel.ProjectID.String(),
		Name:          message.dbModel.Name,
		Description:   message.dbModel.Description,
		Status:        message.dbModel.Status,
		SchemaID:      message.dbModel.SchemaID.String(),
		SchemaVersion: message.dbModel.SchemaVersion,
		CreatedByID:   message.dbModel.CreatedByID.String(),
		CreatedByName: createdByName,
		CreatedAt:     message.dbModel.CreatedAt.String(),
	}
}

// GetProjectID returns the project ID of the message
func (message *MessageObject) GetProjectID() uuid.UUID {
	return message.dbModel.ProjectID
}

// GetID returns the ID of the message
func (message *MessageObject) GetID() uuid.UUID {
	return message.dbModel.ID
}

// GetAllMessagesInProject retrieves all messages in a project
func (messagesManager *MessagesObjectsManager) GetAllMessagesInProject(projectID uuid.UUID) ([]MessageObject, error) {
	var messages []db.MessagesDBModel
	result := db.GetDB().Where("project_id = ? AND status = ?", projectID, "active").Find(&messages)
	if result.Error != nil {
		return nil, result.Error
	}
	
	var messageObjects []MessageObject
	for _, message := range messages {
		messageObjects = append(messageObjects, MessageObject{dbModel: message})
	}
	
	return messageObjects, nil
}

// GetByID retrieves a message by its ID
func (messagesManager *MessagesObjectsManager) GetByID(messageID uuid.UUID) (*MessageObject, error) {
	var message db.MessagesDBModel
	result := db.GetDB().Where("id = ? AND status = ?", messageID, "active").First(&message)
	if result.Error != nil {
		return nil, result.Error
	}
	return &MessageObject{dbModel: message}, nil
}

// CanNameBeUsed checks if a message name can be used in a project
func (messagesManager *MessagesObjectsManager) CanNameBeUsed(name string, projectID uuid.UUID) bool {
	var count int64
	db.GetDB().Model(&db.MessagesDBModel{}).Where(
		"name = ? AND project_id = ? AND status = ?", name, projectID, "active").Count(&count)
	return count == 0
}

// CreateANewMessage creates a new message in the database
func (messagesManager *MessagesObjectsManager) CreateANewMessage(
	description string,
	userID uuid.UUID,
	projectID uuid.UUID,
	name string,
	schemaID uuid.UUID,
	schemaVersion int,
) (*MessageObject, error) {
	
	connection := db.GetDB()
	tx := connection.Begin()
	
	// Create a new message
	newMessage := db.MessagesDBModel{
		Name:          name,
		Description:   description,
		ProjectID:     projectID,
		SchemaID:      schemaID,
		SchemaVersion: schemaVersion,
		Status:        "active",
		CreatedByID:   userID,
	}
	
	if err := tx.Create(&newMessage).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	tx.Commit()
	
	return &MessageObject{dbModel: newMessage}, nil
}