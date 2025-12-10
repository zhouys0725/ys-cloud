package repository

import (
	"ys-cloud/internal/models"

	"gorm.io/gorm"
)

type DeploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *gorm.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) Create(deployment *models.Deployment) error {
	return r.db.Create(deployment).Error
}

func (r *DeploymentRepository) GetByID(id uint) (*models.Deployment, error) {
	var deployment models.Deployment
	err := r.db.Preload("Build").First(&deployment, id).Error
	if err != nil {
		return nil, err
	}
	return &deployment, nil
}

func (r *DeploymentRepository) GetByBuildID(buildID uint) ([]*models.Deployment, error) {
	var deployments []*models.Deployment
	err := r.db.Preload("Build").Where("build_id = ?", buildID).Order("created_at DESC").Find(&deployments).Error
	return deployments, err
}

func (r *DeploymentRepository) GetByEnvironment(environment string) ([]*models.Deployment, error) {
	var deployments []*models.Deployment
	err := r.db.Preload("Build").Where("environment = ?", environment).Order("created_at DESC").Find(&deployments).Error
	return deployments, err
}

func (r *DeploymentRepository) Update(deployment *models.Deployment) error {
	return r.db.Save(deployment).Error
}

func (r *DeploymentRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Deployment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *DeploymentRepository) List(offset, limit int) ([]*models.Deployment, error) {
	var deployments []*models.Deployment
	err := r.db.Preload("Build").Offset(offset).Limit(limit).Order("created_at DESC").Find(&deployments).Error
	return deployments, err
}