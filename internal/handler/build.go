package handler

import (
	"net/http"
	"strconv"
	"ys-cloud/internal/service"

	"github.com/gin-gonic/gin"
)

type BuildHandler struct {
	buildService     *service.BuildService
	gitService       *service.GitService
	dockerService    *service.DockerService
	k8sService       *service.K8sService
}

func NewBuildHandler(buildService *service.BuildService, gitService *service.GitService, dockerService *service.DockerService, k8sService *service.K8sService) *BuildHandler {
	return &BuildHandler{
		buildService:  buildService,
		gitService:    gitService,
		dockerService: dockerService,
		k8sService:    k8sService,
	}
}

type CreateBuildRequest struct {
	PipelineID  uint   `json:"pipeline_id" binding:"required"`
	CommitHash  string `json:"commit_hash"`
	Branch      string `json:"branch"`
	Tag         string `json:"tag"`
}

func (h *BuildHandler) GetBuilds(c *gin.Context) {
	pipelineIDStr := c.Query("pipelineId")
	if pipelineIDStr != "" {
		pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
			return
		}

		builds, err := h.buildService.GetByPipelineID(uint(pipelineID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get builds"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Builds retrieved successfully",
			"builds":  builds,
		})
		return
	}

	// Get all builds with pagination
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	builds, err := h.buildService.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get builds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Builds retrieved successfully",
		"builds":  builds,
	})
}

func (h *BuildHandler) GetBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid build ID"})
		return
	}

	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Build not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Build retrieved successfully",
		"build":   build,
	})
}

func (h *BuildHandler) GetBuildLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid build ID"})
		return
	}

	build, err := h.buildService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Build not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Build logs retrieved successfully",
		"logs":    build.Logs,
	})
}

func (h *BuildHandler) CancelBuild(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid build ID"})
		return
	}

	if err := h.buildService.CancelBuild(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Build cancelled successfully",
	})
}