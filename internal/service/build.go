package service

import (
	"errors"
	"time"
	"ys-cloud/internal/models"
	"ys-cloud/internal/repository"
)

type BuildService struct {
	buildRepo     *repository.BuildRepository
	pipelineRepo  *repository.PipelineRepository
	gitService    *GitService
	dockerService *DockerService
	k8sService    *K8sService
}

func NewBuildService(buildRepo *repository.BuildRepository, pipelineRepo *repository.PipelineRepository) *BuildService {
	return &BuildService{
		buildRepo:    buildRepo,
		pipelineRepo: pipelineRepo,
	}
}

func (s *BuildService) Create(pipelineID uint, commitHash, branch, tag string) (*models.Build, error) {
	// Check if pipeline exists
	_, err := s.pipelineRepo.GetByID(pipelineID)
	if err != nil {
		return nil, errors.New("pipeline not found")
	}

	build := &models.Build{
		PipelineID: pipelineID,
		CommitHash: commitHash,
		Branch:     branch,
		Tag:        tag,
		Status:     "pending",
		ImageTag:   time.Now().Format("20060102-150405"),
	}

	if err := s.buildRepo.Create(build); err != nil {
		return nil, err
	}

	return s.buildRepo.GetByID(build.ID)
}

func (s *BuildService) GetByID(id uint) (*models.Build, error) {
	return s.buildRepo.GetByID(id)
}

func (s *BuildService) GetByPipelineID(pipelineID uint) ([]*models.Build, error) {
	return s.buildRepo.GetByPipelineID(pipelineID)
}

func (s *BuildService) UpdateStatus(id uint, status string) error {
	return s.buildRepo.UpdateStatus(id, status)
}

func (s *BuildService) UpdateLogs(id uint, logs string) error {
	return s.buildRepo.UpdateLogs(id, logs)
}

func (s *BuildService) List(offset, limit int) ([]*models.Build, error) {
	return s.buildRepo.List(offset, limit)
}

func (s *BuildService) StartBuild(id uint) error {
	build, err := s.buildRepo.GetByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	build.Status = "running"
	build.StartedAt = &now

	if err := s.buildRepo.Update(build); err != nil {
		return err
	}

	// TODO: Start actual build process in background
	return nil
}

func (s *BuildService) CompleteBuild(id uint, status, logs, imageName string) error {
	build, err := s.buildRepo.GetByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	build.Status = status
	build.CompletedAt = &now
	build.Logs = logs
	build.ImageName = imageName

	return s.buildRepo.Update(build)
}

func (s *BuildService) CancelBuild(id uint) error {
	return s.buildRepo.UpdateStatus(id, "cancelled")
}