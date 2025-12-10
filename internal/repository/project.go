package repository

import (
	"ys-cloud/internal/models"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Preload("Owner").Preload("Pipelines").First(&project, id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) GetByOwnerID(ownerID uint) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Owner").Preload("Pipelines").Where("owner_id = ?", ownerID).Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) Update(project *models.Project) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *ProjectRepository) List(offset, limit int) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Preload("Owner").Offset(offset).Limit(limit).Find(&projects).Error
	return projects, err
}

func (r *ProjectRepository) GetByGitURL(gitURL string) (*models.Project, error) {
	var project models.Project
	err := r.db.Where("git_url = ?", gitURL).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) AddCollaborator(projectID, userID uint) error {
	return r.db.Exec("INSERT INTO user_projects (user_id, project_id) VALUES (?, ?)", userID, projectID).Error
}

func (r *ProjectRepository) RemoveCollaborator(projectID, userID uint) error {
	return r.db.Exec("DELETE FROM user_projects WHERE user_id = ? AND project_id = ?", userID, projectID).Error
}