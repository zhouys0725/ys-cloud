package service

import (
	"errors"
	"ys-cloud/internal/models"
	"ys-cloud/internal/repository"
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
	userRepo    *repository.UserRepository
}

func NewProjectService(projectRepo *repository.ProjectRepository, userRepo *repository.UserRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *ProjectService) Create(name, description, gitURL, gitProvider string, ownerID uint) (*models.Project, error) {
	// Check if user exists
	if _, err := s.userRepo.GetByID(ownerID); err != nil {
		return nil, errors.New("user not found")
	}

	// Check if Git URL already exists
	if _, err := s.projectRepo.GetByGitURL(gitURL); err == nil {
		return nil, errors.New("Git URL already exists")
	}

	project := &models.Project{
		Name:        name,
		Description: description,
		GitURL:      gitURL,
		GitProvider: gitProvider,
		OwnerID:     ownerID,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	return s.projectRepo.GetByID(project.ID)
}

func (s *ProjectService) GetByID(id uint) (*models.Project, error) {
	return s.projectRepo.GetByID(id)
}

func (s *ProjectService) GetByOwnerID(ownerID uint) ([]*models.Project, error) {
	return s.projectRepo.GetByOwnerID(ownerID)
}

func (s *ProjectService) Update(id uint, name, description string, ownerID uint) (*models.Project, error) {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if project.OwnerID != ownerID {
		return nil, errors.New("access denied")
	}

	if name != "" {
		project.Name = name
	}
	if description != "" {
		project.Description = description
	}

	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	return s.projectRepo.GetByID(id)
}

func (s *ProjectService) Delete(id uint, ownerID uint) error {
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Check ownership
	if project.OwnerID != ownerID {
		return errors.New("access denied")
	}

	return s.projectRepo.Delete(id)
}

func (s *ProjectService) AddCollaborator(projectID, userID, ownerID uint) error {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}

	// Check ownership
	if project.OwnerID != ownerID {
		return errors.New("access denied")
	}

	return s.projectRepo.AddCollaborator(projectID, userID)
}

func (s *ProjectService) RemoveCollaborator(projectID, userID, ownerID uint) error {
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		return err
	}

	// Check ownership
	if project.OwnerID != ownerID {
		return errors.New("access denied")
	}

	return s.projectRepo.RemoveCollaborator(projectID, userID)
}