package logic

import (
	"strings"

	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
)

type ServerObject struct {
	dbModel db.ServersDBModel
}

type ServerDBSerializerStruct struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Protocol          string `json:"protocol"`
	Status            string `json:"status"`
	Description       string `json:"description"`
	ProjectID         string `json:"project_id"`
	CreatedByUserID   string `json:"created_by_user_id"`
	CreatedByUserName string `json:"created_by_user_name"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func (server *ServerObject) Serialize() *ServerDBSerializerStruct {
	createdByName := ""
	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, server.dbModel.CreatedByUserID)
	createdByName = userDbRecord.Handle

	return &ServerDBSerializerStruct{
		ID:                server.dbModel.ID.String(),
		Name:              server.dbModel.Name,
		Protocol:          server.dbModel.Protocol,
		Status:            server.dbModel.Status,
		Description:       server.dbModel.Description,
		ProjectID:         server.dbModel.ProjectID.String(),
		CreatedByUserID:   server.dbModel.CreatedByUserID.String(),
		CreatedByUserName: createdByName,
		CreatedAt:         server.dbModel.CreatedAt.String(),
		UpdatedAt:         server.dbModel.UpdatedAt.String(),
	}
}

func (server *ServerObject) GetID() uuid.UUID {
	return server.dbModel.ID
}

func (server *ServerObject) GetProjectID() uuid.UUID {
	return server.dbModel.ProjectID
}

type ServersObjectsManager struct {
}

func (manager *ServersObjectsManager) GetByID(id uuid.UUID) (*ServerObject, error) {
	serverDbRecord := db.ServersDBModel{}
	dbResult := db.GetDB().Model(db.ServersDBModel{}).First(&serverDbRecord, id)

	if dbResult.Error != nil {
		return nil, common.FusioncatErrRecordNotFound
	}

	server := &ServerObject{}
	server.dbModel = serverDbRecord
	return server, nil
}

func (manager *ServersObjectsManager) CreateANewServer(
	name string,
	description string,
	protocol string,
	projectID uuid.UUID,
	createdByUserID uuid.UUID) (*ServerObject, error) {
	
	newServer := &db.ServersDBModel{
		Name:            strings.TrimSpace(name),
		Description:     description,
		Protocol:        protocol,
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
	}

	if err := db.GetDB().Create(newServer).Error; err != nil {
		return nil, err
	}

	s := &ServerObject{}
	s.dbModel = *newServer
	return s, nil
}

func (manager *ServersObjectsManager) CanNameBeUsed(name string, projectID uuid.UUID) bool {
	var count int64
	_ = db.GetDB().Model(db.ServersDBModel{}).
		Where("name = ? AND project_id = ? AND status = 'active'", name, projectID).
		Count(&count).Error
	return count == 0
}

func (manager *ServersObjectsManager) GetAllServersForProject(
	projectID uuid.UUID) ([]ServerObject, error) {
	var servers []db.ServersDBModel
	var response []ServerObject

	_ = db.GetDB().Model(db.ServersDBModel{}).Where("project_id = ? and status = 'active'", projectID).Find(&servers)

	for _, server := range servers {
		var serverObject ServerObject
		serverObject = ServerObject{server}
		response = append(response, serverObject)
	}
	return response, nil
}