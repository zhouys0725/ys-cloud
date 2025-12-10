package repository

import (
	"ys-cloud/internal/models"

	"gorm.io/gorm"
)

type BuildRepository struct {
	db *gorm.DB
}

func NewBuildRepository(db *gorm.DB) *BuildRepository {
	return &BuildRepository{db: db}
}

func (r *BuildRepository) Create(build *models.Build) error {
	return r.db.Create(build).Error
}

func (r *BuildRepository) GetByID(id uint) (*models.Build, error) {
	var build models.Build
	err := r.db.Preload("Pipeline").Preload("Deployments").First(&build, id).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *BuildRepository) GetByPipelineID(pipelineID uint) ([]*models.Build, error) {
	var builds []*models.Build
	err := r.db.Preload("Pipeline").Where("pipeline_id = ?", pipelineID).Order("created_at DESC").Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) Update(build *models.Build) error {
	return r.db.Save(build).Error
}

func (r *BuildRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Build{}).Where("id = ?", id).Update("status", status).Error
}

func (r *BuildRepository) UpdateLogs(id uint, logs string) error {
	return r.db.Model(&models.Build{}).Where("id = ?", id).Update("logs", logs).Error
}

func (r *BuildRepository) List(offset, limit int) ([]*models.Build, error) {
	var builds []*models.Build
	err := r.db.Preload("Pipeline").Offset(offset).Limit(limit).Order("created_at DESC").Find(&builds).Error
	return builds, err
}

func (r *BuildRepository) GetRunningBuilds() ([]*models.Build, error) {
	var builds []*models.Build
	err := r.db.Where("status IN ?", []string{"pending", "running"}).Find(&builds).Error
	return builds, err
}