package handler

import (
	"net/http"
	"strconv"
	"ys-cloud/internal/service"

	"github.com/gin-gonic/gin"
)

type DeploymentHandler struct {
	deploymentService *service.DeploymentService
	k8sService       *service.K8sService
}

func NewDeploymentHandler(deploymentService *service.DeploymentService, k8sService *service.K8sService) *DeploymentHandler {
	return &DeploymentHandler{
		deploymentService: deploymentService,
		k8sService:       k8sService,
	}
}

type CreateDeploymentRequest struct {
	BuildID      uint   `json:"build_id" binding:"required"`
	Environment  string `json:"environment" binding:"required"`
	Replicas     int32  `json:"replicas" binding:"required"`
	Namespace    string `json:"namespace"`
	ServiceName  string `json:"service_name"`
	IngressHost  string `json:"ingress_host"`
}

func (h *DeploymentHandler) GetDeployments(c *gin.Context) {
	buildIDStr := c.Query("buildId")
	if buildIDStr != "" {
		buildID, err := strconv.ParseUint(buildIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid build ID"})
			return
		}

		deployments, err := h.deploymentService.GetByBuildID(uint(buildID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deployments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Deployments retrieved successfully",
			"deployments": deployments,
		})
		return
	}

	environment := c.Query("environment")
	if environment != "" {
		deployments, err := h.deploymentService.GetByEnvironment(environment)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deployments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Deployments retrieved successfully",
			"deployments": deployments,
		})
		return
	}

	// Get all deployments with pagination
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	deployments, err := h.deploymentService.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get deployments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Deployments retrieved successfully",
		"deployments": deployments,
	})
}

func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	deployment, err := h.deploymentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Deployment retrieved successfully",
		"deployment": deployment,
	})
}

func (h *DeploymentHandler) GetDeploymentLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	deployment, err := h.deploymentService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	// TODO: Get actual deployment logs from Kubernetes
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment logs retrieved successfully",
		"logs":    "Deployment logs placeholder",
		"deployment": deployment,
	})
}

func (h *DeploymentHandler) RollbackDeployment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	if err := h.deploymentService.Rollback(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment rollback initiated successfully",
	})
}