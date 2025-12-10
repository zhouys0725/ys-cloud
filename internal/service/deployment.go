package service

import (
	"errors"
	"time"
	"ys-cloud/internal/models"
	"ys-cloud/internal/repository"
)

type DeploymentService struct {
	deploymentRepo *repository.DeploymentRepository
	buildRepo      *repository.BuildRepository
	k8sService     *K8sService
}

func NewDeploymentService(deploymentRepo *repository.DeploymentRepository, k8sService *K8sService) *DeploymentService {
	return &DeploymentService{
		deploymentRepo: deploymentRepo,
		k8sService:     k8sService,
	}
}

func (s *DeploymentService) Create(buildID uint, environment string, replicas int32, namespace, serviceName, ingressHost string) (*models.Deployment, error) {
	// Check if build exists
	build, err := s.buildRepo.GetByID(buildID)
	if err != nil {
		return nil, errors.New("build not found")
	}

	// Check if build was successful
	if build.Status != "success" {
		return nil, errors.New("build was not successful")
	}

	deployment := &models.Deployment{
		BuildID:     buildID,
		Environment: environment,
		Status:      "pending",
		Replicas:    replicas,
		Namespace:   namespace,
		ServiceName: serviceName,
		IngressHost: ingressHost,
	}

	if err := s.deploymentRepo.Create(deployment); err != nil {
		return nil, err
	}

	return s.deploymentRepo.GetByID(deployment.ID)
}

func (s *DeploymentService) GetByID(id uint) (*models.Deployment, error) {
	return s.deploymentRepo.GetByID(id)
}

func (s *DeploymentService) GetByBuildID(buildID uint) ([]*models.Deployment, error) {
	return s.deploymentRepo.GetByBuildID(buildID)
}

func (s *DeploymentService) GetByEnvironment(environment string) ([]*models.Deployment, error) {
	return s.deploymentRepo.GetByEnvironment(environment)
}

func (s *DeploymentService) UpdateStatus(id uint, status string) error {
	return s.deploymentRepo.UpdateStatus(id, status)
}

func (s *DeploymentService) List(offset, limit int) ([]*models.Deployment, error) {
	return s.deploymentRepo.List(offset, limit)
}

func (s *DeploymentService) StartDeployment(id uint) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	deployment.Status = "running"
	deployment.StartedAt = &now

	if err := s.deploymentRepo.Update(deployment); err != nil {
		return err
	}

	// TODO: Start actual deployment process
	return nil
}

func (s *DeploymentService) CompleteDeployment(id uint, status string) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return err
	}

	now := time.Now()
	deployment.Status = status
	deployment.CompletedAt = &now

	return s.deploymentRepo.Update(deployment)
}

func (s *DeploymentService) CancelDeployment(id uint) error {
	return s.deploymentRepo.UpdateStatus(id, "cancelled")
}

func (s *DeploymentService) Rollback(id uint) error {
	deployment, err := s.deploymentRepo.GetByID(id)
	if err != nil {
		return err
	}

	// TODO: Implement rollback logic
	return s.k8sService.RollbackDeployment(deployment.Namespace, deployment.ServiceName)
}