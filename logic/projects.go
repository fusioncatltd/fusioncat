package logic

import (
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
	"strings"
)

// ProjectObject represents a project in the system which is a container for
// all apps, schemas, and other resources.
// Projects are visible to all users, but can be private or public.
// Private projects are now supported, but not implemented yet.
type ProjectObject struct {
	dbModel db.ProjectsDBModel
}

type ProjectDBSerializerStruct struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	IsPrivate     bool   `json:"is_private"`
	Description   string `json:"description"`
	CreatedByType string `json:"created_by_type"`
	CreatedByID   string `json:"created_by_id"`
	CreatedByName string `json:"created_by_name"`
}

func (project *ProjectObject) Serialize() *ProjectDBSerializerStruct {
	createdByName := ""

	userDbRecord := db.UsersDBModel{}
	_ = db.GetDB().First(&userDbRecord, project.dbModel.CreatedByID)
	createdByName = userDbRecord.Handle

	return &ProjectDBSerializerStruct{
		ID:            project.dbModel.ID.String(),
		Name:          project.dbModel.Name,
		Description:   project.dbModel.Description,
		Status:        project.dbModel.Status,
		IsPrivate:     project.dbModel.IsPrivate,
		CreatedByType: project.dbModel.CreatedByType,
		CreatedByID:   project.dbModel.CreatedByID.String(),
		CreatedByName: createdByName,
	}
}

func (project *ProjectObject) GetID() uuid.UUID {
	return project.dbModel.ID
}

// ProjectsObjectsManager manages project objects in the system. It accumulates functions
// which perform operations over multiple project objects, such as creating new projects,
// retrieving projects by ID or email, etc.
type ProjectsObjectsManager struct {
}

// GetAllProjects retrieves all projects that are active in the system.
// Currently, returns both public and private projects.
func (projectsManager *ProjectsObjectsManager) GetAllProjects(myID uuid.UUID) ([]ProjectObject, error) {
	var projects []db.ProjectsDBModel
	var response []ProjectObject

	_ = db.GetDB().Model(db.ProjectsDBModel{}).Where("status = 'active'").Find(&projects)

	for _, project := range projects {
		var projectObject ProjectObject
		projectObject = ProjectObject{project}
		response = append(response, projectObject)
	}
	return response, nil
}

// CheckIfProjectWithSpecificNameExists checks if a project with the given name exists
// Only one project with a specific name can exist for a user at a time.
func (projectsManager *ProjectsObjectsManager) CheckIfProjectWithSpecificNameExists(name string) bool {
	var count int64
	_ = db.GetDB().Model(db.ProjectsDBModel{}).Where("name = ? AND status = 'active'",
		name).Count(&count).Error
	return count > 0
}

// CreateANewProject creates a new project in the system.
func (projectsManager *ProjectsObjectsManager) CreateANewProject(name string,
	description string, createdById uuid.UUID) (*ProjectObject, error) {
	newProject := &db.ProjectsDBModel{
		Name:          strings.TrimSpace(name),
		Description:   description,
		IsPrivate:     false, // Private projects are not supported yet
		CreatedByType: "user",
		CreatedByID:   createdById,
	}

	if err := db.GetDB().Create(newProject).Error; err != nil {
		return nil, err
	}

	p := &ProjectObject{}
	p.dbModel = *newProject
	return p, nil
}
