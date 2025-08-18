package logic

import (
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
)

// AppResourceMessageObject represents a relationship between an app, resource, and message
type AppResourceMessageObject struct {
	dbModel db.AppResourceMessagesDBModel
}

// GetAppID returns the app ID
func (arm *AppResourceMessageObject) GetAppID() uuid.UUID {
	return arm.dbModel.AppID
}

// GetResourceID returns the resource ID
func (arm *AppResourceMessageObject) GetResourceID() uuid.UUID {
	return arm.dbModel.ResourceID
}

// GetMessageID returns the message ID
func (arm *AppResourceMessageObject) GetMessageID() uuid.UUID {
	return arm.dbModel.MessageID
}

// GetDirection returns the direction (sends/receives)
func (arm *AppResourceMessageObject) GetDirection() string {
	return arm.dbModel.Direction
}

// GetStatus returns the status
func (arm *AppResourceMessageObject) GetStatus() string {
	return arm.dbModel.Status
}

// AppsResourcesMessagesObjectsManager manages app-resource-message relationships
type AppsResourcesMessagesObjectsManager struct {
}

// GetAllForApp retrieves all app resource messages for a specific app
func (manager *AppsResourcesMessagesObjectsManager) GetAllForApp(appID uuid.UUID) ([]*AppResourceMessageObject, error) {
	var records []db.AppResourceMessagesDBModel
	result := db.GetDB().Where("app_id = ? AND status = ?", appID, "active").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}

	var objects []*AppResourceMessageObject
	for _, record := range records {
		objects = append(objects, &AppResourceMessageObject{dbModel: record})
	}
	return objects, nil
}

// CreateConnection creates a new app-resource-message connection
func (manager *AppsResourcesMessagesObjectsManager) CreateConnection(
	appID uuid.UUID,
	resourceID uuid.UUID,
	messageID uuid.UUID,
	direction string,
	createdByUserID uuid.UUID,
) (*AppResourceMessageObject, error) {
	newConnection := &db.AppResourceMessagesDBModel{
		AppID:           appID,
		ResourceID:      resourceID,
		MessageID:       messageID,
		Direction:       direction,
		CreatedByUserID: createdByUserID,
		Status:          "active",
	}

	if err := db.GetDB().Create(newConnection).Error; err != nil {
		return nil, err
	}

	return &AppResourceMessageObject{dbModel: *newConnection}, nil
}

// ConnectionExists checks if a connection already exists
func (manager *AppsResourcesMessagesObjectsManager) ConnectionExists(
	appID uuid.UUID,
	resourceID uuid.UUID,
	messageID uuid.UUID,
	direction string,
) bool {
	var count int64
	db.GetDB().Model(&db.AppResourceMessagesDBModel{}).
		Where("app_id = ? AND resource_id = ? AND message_id = ? AND direction = ? AND status = ?",
			appID, resourceID, messageID, direction, "active").
		Count(&count)
	return count > 0
}