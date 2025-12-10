package service

import (
	"ys-cloud/internal/config"
	"ys-cloud/pkg/docker"
)

type DockerService struct {
	*docker.DockerService
}

func NewDockerService(cfg *config.Config) (*DockerService, error) {
	dockerService, err := docker.NewDockerService(cfg)
	if err != nil {
		return nil, err
	}
	return &DockerService{
		DockerService: dockerService,
	}, nil
}