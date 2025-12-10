package service

import (
	"ys-cloud/internal/config"
	"ys-cloud/pkg/k8s"
)

type K8sService struct {
	*k8s.K8sService
}

func NewK8sService(cfg *config.Config) (*K8sService, error) {
	k8sService, err := k8s.NewK8sService(cfg)
	if err != nil {
		return nil, err
	}
	return &K8sService{
		K8sService: k8sService,
	}, nil
}