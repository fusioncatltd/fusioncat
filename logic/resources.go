package logic

import (
	"strings"

	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
)

type ResourceObject struct {
	dbModel db.ResourcesDBModel
}

type ResourceDBSerializerStruct struct {
	ID                string `json:"id"`
	ServerID          string `json:"server_id"`
	Name              string `json:"name"`
	Mode              string `json:"mode"`
	ResourceType      string `json:"resource_type"`
	Status            string `json:"status"`
	Description       string `json:"description"`
	ProjectID         string `json:"project_id"`
	CreatedByUserID   string `json:"created_by_user_id"`
	CreatedByUserName string `json:"created_by_user_name"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func (resource *ResourceObject) Serialize() *ResourceDBSerializerStruct {
	createdByName := ""
	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, resource.dbModel.CreatedByUserID)
	createdByName = userDbRecord.Handle

	return &ResourceDBSerializerStruct{
		ID:                resource.dbModel.ID.String(),
		ServerID:          resource.dbModel.ServerID.String(),
		Name:              resource.dbModel.Name,
		Mode:              resource.dbModel.Mode,
		ResourceType:      resource.dbModel.ResourceType,
		Status:            resource.dbModel.Status,
		Description:       resource.dbModel.Description,
		ProjectID:         resource.dbModel.ProjectID.String(),
		CreatedByUserID:   resource.dbModel.CreatedByUserID.String(),
		CreatedByUserName: createdByName,
		CreatedAt:         resource.dbModel.CreatedAt.String(),
		UpdatedAt:         resource.dbModel.UpdatedAt.String(),
	}
}

func (resource *ResourceObject) GetID() uuid.UUID {
	return resource.dbModel.ID
}

func (resource *ResourceObject) GetServerID() uuid.UUID {
	return resource.dbModel.ServerID
}

type ResourcesObjectsManager struct {
}

func (manager *ResourcesObjectsManager) GetByID(id uuid.UUID) (*ResourceObject, error) {
	resourceDbRecord := db.ResourcesDBModel{}
	dbResult := db.GetDB().Model(db.ResourcesDBModel{}).First(&resourceDbRecord, id)

	if dbResult.Error != nil {
		return nil, common.FusioncatErrRecordNotFound
	}

	resource := &ResourceObject{}
	resource.dbModel = resourceDbRecord
	return resource, nil
}

func (manager *ResourcesObjectsManager) CreateANewResource(
	serverID uuid.UUID,
	projectID uuid.UUID,
	name string,
	mode string,
	resourceType string,
	description string,
	createdByUserID uuid.UUID) (*ResourceObject, error) {
	
	newResource := &db.ResourcesDBModel{
		ServerID:        serverID,
		Name:            strings.TrimSpace(name),
		Mode:            mode,
		ResourceType:    resourceType,
		Description:     description,
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
	}

	if err := db.GetDB().Create(newResource).Error; err != nil {
		return nil, err
	}

	r := &ResourceObject{}
	r.dbModel = *newResource
	return r, nil
}

func (manager *ResourcesObjectsManager) CanNameBeUsed(name string, serverID uuid.UUID) bool {
	var count int64
	_ = db.GetDB().Model(db.ResourcesDBModel{}).
		Where("name = ? AND server_id = ? AND status = 'active'", name, serverID).
		Count(&count).Error
	return count == 0
}

func (manager *ResourcesObjectsManager) GetAllResourcesForServer(
	serverID uuid.UUID) ([]ResourceObject, error) {
	var resources []db.ResourcesDBModel
	var response []ResourceObject

	_ = db.GetDB().Model(db.ResourcesDBModel{}).Where("server_id = ? and status = 'active'", serverID).Find(&resources)

	for _, resource := range resources {
		var resourceObject ResourceObject
		resourceObject = ResourceObject{resource}
		response = append(response, resourceObject)
	}
	return response, nil
}