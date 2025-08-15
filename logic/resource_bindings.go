package logic

import (
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
)

type ResourceBindingObject struct {
	dbModel db.ResourceBindingsDBModel
}

type ResourceBindingDBSerializerStruct struct {
	ID               string `json:"id"`
	SourceResourceID string `json:"source_resource_id"`
	TargetResourceID string `json:"target_resource_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func (binding *ResourceBindingObject) Serialize() *ResourceBindingDBSerializerStruct {
	return &ResourceBindingDBSerializerStruct{
		ID:               binding.dbModel.ID.String(),
		SourceResourceID: binding.dbModel.SourceResourceID.String(),
		TargetResourceID: binding.dbModel.TargetResourceID.String(),
		CreatedAt:        binding.dbModel.CreatedAt.String(),
		UpdatedAt:        binding.dbModel.UpdatedAt.String(),
	}
}

func (binding *ResourceBindingObject) GetID() uuid.UUID {
	return binding.dbModel.ID
}

type ResourceBindingsObjectsManager struct {
}

func (manager *ResourceBindingsObjectsManager) GetByID(id uuid.UUID) (*ResourceBindingObject, error) {
	bindingDbRecord := db.ResourceBindingsDBModel{}
	dbResult := db.GetDB().Model(db.ResourceBindingsDBModel{}).First(&bindingDbRecord, id)

	if dbResult.Error != nil {
		return nil, common.FusioncatErrRecordNotFound
	}

	binding := &ResourceBindingObject{}
	binding.dbModel = bindingDbRecord
	return binding, nil
}

func (manager *ResourceBindingsObjectsManager) CreateABinding(
	sourceResourceID uuid.UUID,
	targetResourceID uuid.UUID) (*ResourceBindingObject, error) {
	
	newBinding := &db.ResourceBindingsDBModel{
		SourceResourceID: sourceResourceID,
		TargetResourceID: targetResourceID,
	}

	if err := db.GetDB().Create(newBinding).Error; err != nil {
		return nil, err
	}

	b := &ResourceBindingObject{}
	b.dbModel = *newBinding
	return b, nil
}

func (manager *ResourceBindingsObjectsManager) CheckIfBindingExists(sourceResourceID, targetResourceID uuid.UUID) bool {
	var count int64
	_ = db.GetDB().Model(db.ResourceBindingsDBModel{}).
		Where("((source_resource_id = ? AND target_resource_id = ?) OR (source_resource_id = ? AND target_resource_id = ?))",
			sourceResourceID, targetResourceID, targetResourceID, sourceResourceID).
		Count(&count).Error
	return count > 0
}

func (manager *ResourceBindingsObjectsManager) GetAllBindingsForResource(
	resourceID uuid.UUID) ([]ResourceBindingObject, error) {
	var bindings []db.ResourceBindingsDBModel
	var response []ResourceBindingObject

	_ = db.GetDB().Model(db.ResourceBindingsDBModel{}).
		Where("source_resource_id = ? OR target_resource_id = ?", resourceID, resourceID).
		Find(&bindings)

	for _, binding := range bindings {
		var bindingObject ResourceBindingObject
		bindingObject = ResourceBindingObject{binding}
		response = append(response, bindingObject)
	}
	return response, nil
}