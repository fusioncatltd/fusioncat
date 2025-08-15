package logic

import (
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
	"strings"
)

type AppObject struct {
	dbModel db.AppsDBModel
}

type AppDBSerializerStruct struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	Description       string `json:"description"`
	ProjectID         string `json:"project_id"`
	CreatedByUserID   string `json:"created_by_user_id"`
	CreatedByUserName string `json:"created_by_user_name"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func (app *AppObject) Serialize() *AppDBSerializerStruct {
	createdByName := ""

	authorDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&authorDbRecord, app.dbModel.CreatedByUserID)
	createdByName = authorDbRecord.Handle

	return &AppDBSerializerStruct{
		ID:                app.dbModel.ID.String(),
		Name:              app.dbModel.Name,
		Status:            app.dbModel.Status,
		Description:       app.dbModel.Description,
		ProjectID:         app.dbModel.ProjectID.String(),
		CreatedByUserID:   app.dbModel.CreatedByUserID.String(),
		CreatedByUserName: createdByName,
		CreatedAt:         app.dbModel.CreatedAt.String(),
		UpdatedAt:         app.dbModel.UpdatedAt.String(),
	}
}

func (app *AppObject) GetID() uuid.UUID {
	return app.dbModel.ID
}

type AppsObjectsManager struct {
}

func (appsManager *AppsObjectsManager) GetByID(id uuid.UUID) (*AppObject, error) {
	appDbRecord := db.AppsDBModel{}
	dbResult := db.GetDB().Model(db.AppsDBModel{}).First(&appDbRecord, id)

	if dbResult.Error != nil {
		return nil, common.FusioncatErrRecordNotFound
	}

	app := &AppObject{}
	app.dbModel = appDbRecord
	return app, nil
}

func (appsManager *AppsObjectsManager) CreateANewApp(
	name string,
	description string,
	projectID uuid.UUID,
	createdByUserID uuid.UUID) (*AppObject, error) {
	newApp := &db.AppsDBModel{
		Name:            strings.TrimSpace(name),
		Description:     description,
		ProjectID:       projectID,
		CreatedByUserID: createdByUserID,
	}

	if err := db.GetDB().Create(newApp).Error; err != nil {
		return nil, err
	}

	a := &AppObject{}
	a.dbModel = *newApp
	return a, nil
}

func (appsManager *AppsObjectsManager) CanNameBeUsed(name string, projectID uuid.UUID) bool {
	result, _ := db.AppsDBModel{}.CanCreate(projectID, name)
	return result
}

func (appsManager *AppsObjectsManager) GetAllAppsForProject(
	projectID uuid.UUID) ([]AppObject, error) {
	var apps []db.AppsDBModel
	var response []AppObject

	_ = db.GetDB().Model(db.AppsDBModel{}).Where("project_id = ? and status = 'active'", projectID).Find(&apps)

	for _, app := range apps {
		var appObject AppObject
		appObject = AppObject{app}
		response = append(response, appObject)
	}
	return response, nil
}