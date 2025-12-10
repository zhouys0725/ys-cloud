package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"ys-cloud/internal/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/sirupsen/logrus"
)

type DockerService struct {
	client  *client.Client
	config  *config.DockerConfig
	logger  *logrus.Logger
}

type BuildOptions struct {
	ContextDir    string
	Dockerfile    string
	ImageName     string
	ImageTag      string
	BuildArgs     map[string]*string
	Labels        map[string]string
	NoCache       bool
	Remove        bool
}

type BuildProgress struct {
	Stream string `json:"stream"`
	Error  string `json:"error,omitempty"`
	Status string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
}

func NewDockerService(cfg *config.Config) (*DockerService, error) {
	var cli *client.Client
	var err error

	// Try to connect to Docker daemon
	if cfg.Docker.Host != "" && cfg.Docker.Host != "unix:///var/run/docker.sock" {
		cli, err = client.NewClientWithOpts(client.WithHost(cfg.Docker.Host), client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("failed to create Docker client with custom host: %w", err)
		}
	} else {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("failed to create Docker client from environment: %w", err)
		}
	}

	// Test connection
	_, err = cli.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	return &DockerService{
		client: cli,
		config: &cfg.Docker,
		logger: logrus.New(),
	}, nil
}

func (s *DockerService) BuildImage(opts BuildOptions) error {
	ctx := context.Background()

	// Validate required fields
	if opts.ContextDir == "" {
		return fmt.Errorf("context directory is required")
	}
	if opts.ImageName == "" {
		return fmt.Errorf("image name is required")
	}
	if opts.ImageTag == "" {
		return fmt.Errorf("image tag is required")
	}
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}

	// Create build context
	buildContext, err := archive.TarWithOptions(opts.ContextDir, &archive.TarOptions{
		ExcludePatterns: []string{".git", ".gitignore", "Dockerfile.dockerignore"},
	})
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: opts.Dockerfile,
		Tags:       []string{fmt.Sprintf("%s:%s", opts.ImageName, opts.ImageTag)},
		BuildArgs:  opts.BuildArgs,
		Labels:     opts.Labels,
		NoCache:    opts.NoCache,
		Remove:     opts.Remove,
		ForceRemove: true,
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": opts.ImageName,
		"image_tag":  opts.ImageTag,
		"context":    opts.ContextDir,
		"dockerfile": opts.Dockerfile,
	}).Info("Starting Docker image build")

	resp, err := s.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to start image build: %w", err)
	}
	defer resp.Body.Close()

	// Stream and log build progress
	decoder := json.NewDecoder(resp.Body)
	for {
		var progress BuildProgress
		if err := decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			s.logger.WithError(err).Warn("Failed to decode build progress")
			continue
		}

		if progress.Error != "" {
			return fmt.Errorf("build failed: %s", progress.Error)
		}

		if progress.Stream != "" {
			s.logger.WithField("progress", progress.Stream).Debug("Build progress")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": opts.ImageName,
		"image_tag":  opts.ImageTag,
	}).Info("Docker image build completed successfully")

	return nil
}

func (s *DockerService) BuildImageWithLogs(opts BuildOptions) (string, error) {
	ctx := context.Background()

	// Validate required fields
	if opts.ContextDir == "" {
		return "", fmt.Errorf("context directory is required")
	}
	if opts.ImageName == "" {
		return "", fmt.Errorf("image name is required")
	}
	if opts.ImageTag == "" {
		return "", fmt.Errorf("image tag is required")
	}
	if opts.Dockerfile == "" {
		opts.Dockerfile = "Dockerfile"
	}

	// Create build context
	buildContext, err := archive.TarWithOptions(opts.ContextDir, &archive.TarOptions{
		ExcludePatterns: []string{".git", ".gitignore", "Dockerfile.dockerignore"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: opts.Dockerfile,
		Tags:       []string{fmt.Sprintf("%s:%s", opts.ImageName, opts.ImageTag)},
		BuildArgs:  opts.BuildArgs,
		Labels:     opts.Labels,
		NoCache:    opts.NoCache,
		Remove:     opts.Remove,
		ForceRemove: true,
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": opts.ImageName,
		"image_tag":  opts.ImageTag,
		"context":    opts.ContextDir,
	}).Info("Starting Docker image build with logs")

	resp, err := s.client.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to start image build: %w", err)
	}
	defer resp.Body.Close()

	// Collect build logs
	var logs string
	decoder := json.NewDecoder(resp.Body)
	for {
		var progress BuildProgress
		if err := decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if progress.Error != "" {
			return logs, fmt.Errorf("build failed: %s", progress.Error)
		}

		if progress.Stream != "" {
			logs += progress.Stream
		}
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": opts.ImageName,
		"image_tag":  opts.ImageTag,
	}).Info("Docker image build completed successfully")

	return logs, nil
}

func (s *DockerService) PushImage(imageName, imageTag, username, password string) error {
	ctx := context.Background()

	// Validate required fields
	if imageName == "" {
		return fmt.Errorf("image name is required")
	}
	if imageTag == "" {
		return fmt.Errorf("image tag is required")
	}

	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	authConfig := registry.AuthConfig{
		Username: username,
		Password: password,
	}

	authStr, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return fmt.Errorf("failed to encode auth: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": fullImageName,
		"registry":   s.config.Registry,
	}).Info("Starting Docker image push")

	pushResp, err := s.client.ImagePush(ctx, fullImageName, image.PushOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer pushResp.Close()

	// Stream and log push progress
	decoder := json.NewDecoder(pushResp)
	for {
		var progress map[string]interface{}
		if err := decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		if errorDetail, exists := progress["error"]; exists && errorDetail != nil {
			return fmt.Errorf("push failed: %v", errorDetail)
		}

		if progress["status"] != nil {
			s.logger.WithField("progress", progress["status"]).Debug("Push progress")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"image_name": fullImageName,
		"registry":   s.config.Registry,
	}).Info("Docker image push completed successfully")

	return nil
}

func (s *DockerService) TagImage(sourceImage, targetImage, targetTag string) error {
	ctx := context.Background()

	// Validate required fields
	if sourceImage == "" {
		return fmt.Errorf("source image is required")
	}
	if targetImage == "" {
		return fmt.Errorf("target image is required")
	}
	if targetTag == "" {
		return fmt.Errorf("target tag is required")
	}

	source := fmt.Sprintf("%s:latest", sourceImage)
	target := fmt.Sprintf("%s:%s", targetImage, targetTag)

	err := s.client.ImageTag(ctx, source, target)
	if err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"source": source,
		"target": target,
	}).Info("Docker image tagged successfully")

	return nil
}

func (s *DockerService) RemoveImage(imageName, imageTag string) error {
	ctx := context.Background()

	// Validate required fields
	if imageName == "" {
		return fmt.Errorf("image name is required")
	}
	if imageTag == "" {
		return fmt.Errorf("image tag is required")
	}

	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	_, err := s.client.ImageRemove(ctx, fullImageName, image.RemoveOptions{
		Force:         true,
		PruneChildren: true,
	})
	if err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}

	s.logger.WithField("image_name", fullImageName).Info("Docker image removed successfully")
	return nil
}

func (s *DockerService) ListImages() ([]image.Summary, error) {
	ctx := context.Background()
	images, err := s.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return images, nil
}

func (s *DockerService) GetImageInfo(imageName, imageTag string) (*types.ImageInspect, error) {
	ctx := context.Background()

	// Validate required fields
	if imageName == "" {
		return nil, fmt.Errorf("image name is required")
	}
	if imageTag == "" {
		return nil, fmt.Errorf("image tag is required")
	}

	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	inspect, _, err := s.client.ImageInspectWithRaw(ctx, fullImageName)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}

	return &inspect, nil
}

func (s *DockerService) ImageExists(imageName, imageTag string) (bool, error) {
	ctx := context.Background()
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageTag)

	_, _, err := s.client.ImageInspectWithRaw(ctx, fullImageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	return true, nil
}

func (s *DockerService) PruneImages() (image.PruneReport, error) {
	ctx := context.Background()
	pruneFilters := filters.NewArgs()
	pruneReport, err := s.client.ImagesPrune(ctx, pruneFilters)
	if err != nil {
		return image.PruneReport{}, fmt.Errorf("failed to prune images: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"deleted_images": len(pruneReport.ImagesDeleted),
		"space_reclaimed": pruneReport.SpaceReclaimed,
	}).Info("Image pruning completed")

	return pruneReport, nil
}

// encodeAuthToBase64 encodes the auth configuration to base64
func encodeAuthToBase64(authConfig registry.AuthConfig) (string, error) {
	authBytes, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth config: %w", err)
	}
	return base64.URLEncoding.EncodeToString(authBytes), nil
}

// GetSystemInfo returns Docker system information
func (s *DockerService) GetSystemInfo() (system.Info, error) {
	ctx := context.Background()
	info, err := s.client.Info(ctx)
	if err != nil {
		return system.Info{}, fmt.Errorf("failed to get Docker system info: %w", err)
	}
	return info, nil
}

// TestConnection tests the connection to Docker daemon
func (s *DockerService) TestConnection() error {
	ctx := context.Background()
	_, err := s.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Docker daemon connection failed: %w", err)
	}
	return nil
}