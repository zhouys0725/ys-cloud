package repository

import (
	"ys-cloud/internal/models"

	"gorm.io/gorm"
)

type PipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) *PipelineRepository {
	return &PipelineRepository{db: db}
}

func (r *PipelineRepository) Create(pipeline *models.Pipeline) error {
	return r.db.Create(pipeline).Error
}

func (r *PipelineRepository) GetByID(id uint) (*models.Pipeline, error) {
	var pipeline models.Pipeline
	err := r.db.Preload("Project").Preload("Triggers").Preload("Builds").First(&pipeline, id).Error
	if err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func (r *PipelineRepository) GetByProjectID(projectID uint) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline
	err := r.db.Preload("Project").Preload("Triggers").Where("project_id = ?", projectID).Find(&pipelines).Error
	return pipelines, err
}

func (r *PipelineRepository) Update(pipeline *models.Pipeline) error {
	return r.db.Save(pipeline).Error
}

func (r *PipelineRepository) Delete(id uint) error {
	return r.db.Delete(&models.Pipeline{}, id).Error
}

func (r *PipelineRepository) List(offset, limit int) ([]*models.Pipeline, error) {
	var pipelines []*models.Pipeline
	err := r.db.Preload("Project").Offset(offset).Limit(limit).Find(&pipelines).Error
	return pipelines, err
}