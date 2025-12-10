package handler

import (
	"net/http"
	"strconv"
	"ys-cloud/internal/service"

	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	pipelineService *service.PipelineService
	gitService      *service.GitService
}

func NewPipelineHandler(pipelineService *service.PipelineService, gitService *service.GitService) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
		gitService:      gitService,
	}
}

type CreatePipelineRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Config      string `json:"config" binding:"required"`
}

type UpdatePipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      string `json:"config"`
}

func (h *PipelineHandler) CreatePipeline(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		projectID, err = strconv.ParseUint(c.Query("projectId"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pipeline, err := h.pipelineService.Create(req.Name, req.Description, req.Config, uint(projectID))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Pipeline created successfully",
		"pipeline": pipeline,
	})
}

func (h *PipelineHandler) GetPipelines(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	if projectIDStr != "" {
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		pipelines, err := h.pipelineService.GetByProjectID(uint(projectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pipelines"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "Pipelines retrieved successfully",
			"pipelines": pipelines,
		})
		return
	}

	// Get all pipelines
	c.JSON(http.StatusOK, gin.H{
		"message":   "Pipelines retrieved successfully",
		"pipelines": []interface{}{},
	})
}

func (h *PipelineHandler) GetPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	pipeline, err := h.pipelineService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pipeline not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Pipeline retrieved successfully",
		"pipeline": pipeline,
	})
}

func (h *PipelineHandler) UpdatePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	var req UpdatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get project ID from pipeline
	projectID := uint(1) // Placeholder

	pipeline, err := h.pipelineService.Update(uint(id), projectID, req.Name, req.Description, req.Config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Pipeline updated successfully",
		"pipeline": pipeline,
	})
}

func (h *PipelineHandler) DeletePipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	// TODO: Get project ID from pipeline
	projectID := uint(1) // Placeholder

	if err := h.pipelineService.Delete(uint(id), projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pipeline deleted successfully",
	})
}

func (h *PipelineHandler) AddTrigger(c *gin.Context) {
	// TODO: Implement trigger addition
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *PipelineHandler) UpdateTrigger(c *gin.Context) {
	// TODO: Implement trigger update
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *PipelineHandler) RemoveTrigger(c *gin.Context) {
	// TODO: Implement trigger removal
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (h *PipelineHandler) RunPipeline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pipeline ID"})
		return
	}

	// TODO: Implement pipeline execution
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Pipeline execution not implemented yet",
		"id":      id,
	})
}