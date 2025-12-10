package service

import (
	"errors"
	"ys-cloud/internal/models"
	"ys-cloud/internal/repository"
)

type PipelineService struct {
	pipelineRepo *repository.PipelineRepository
	projectRepo  *repository.ProjectRepository
}

func NewPipelineService(pipelineRepo *repository.PipelineRepository, projectRepo *repository.ProjectRepository) *PipelineService {
	return &PipelineService{
		pipelineRepo: pipelineRepo,
		projectRepo:  projectRepo,
	}
}

func (s *PipelineService) Create(name, description, config string, projectID uint) (*models.Pipeline, error) {
	// Check if project exists
	_, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}

	pipeline := &models.Pipeline{
		Name:        name,
		Description: description,
		Config:      config,
		ProjectID:   projectID,
		Status:      "inactive",
	}

	if err := s.pipelineRepo.Create(pipeline); err != nil {
		return nil, err
	}

	return s.pipelineRepo.GetByID(pipeline.ID)
}

func (s *PipelineService) GetByID(id uint) (*models.Pipeline, error) {
	return s.pipelineRepo.GetByID(id)
}

func (s *PipelineService) GetByProjectID(projectID uint) ([]*models.Pipeline, error) {
	return s.pipelineRepo.GetByProjectID(projectID)
}

func (s *PipelineService) Update(id, projectID uint, name, description, config string) (*models.Pipeline, error) {
	pipeline, err := s.pipelineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check project ownership
	if pipeline.ProjectID != projectID {
		return nil, errors.New("access denied")
	}

	if name != "" {
		pipeline.Name = name
	}
	if description != "" {
		pipeline.Description = description
	}
	if config != "" {
		pipeline.Config = config
	}

	if err := s.pipelineRepo.Update(pipeline); err != nil {
		return nil, err
	}

	return s.pipelineRepo.GetByID(id)
}

func (s *PipelineService) Delete(id, projectID uint) error {
	pipeline, err := s.pipelineRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Check project ownership
	if pipeline.ProjectID != projectID {
		return errors.New("access denied")
	}

	return s.pipelineRepo.Delete(id)
}