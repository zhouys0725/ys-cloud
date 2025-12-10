package main

import (
	"fmt"
	"log"
	"ys-cloud/internal/config"
	"ys-cloud/internal/handler"
	"ys-cloud/internal/middleware"
	"ys-cloud/internal/repository"
	"ys-cloud/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() error {
	// Skip .env file loading in Kubernetes to rely on environment variables
	// if err := godotenv.Load(); err != nil {
	//	log.Println("No .env file found, using environment variables")
	// }

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Debug: print JWT config
	log.Printf("JWT Config: %+v", cfg.JWT)

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	pipelineRepo := repository.NewPipelineRepository(db)
	buildRepo := repository.NewBuildRepository(db)
	deploymentRepo := repository.NewDeploymentRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo, userRepo)
	pipelineService := service.NewPipelineService(pipelineRepo, projectRepo)
	buildService := service.NewBuildService(buildRepo, pipelineRepo)
	gitService := service.NewGitService()
	dockerService, err := service.NewDockerService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Docker service: %v", err)
		dockerService = nil
	}
	k8sService, err := service.NewK8sService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize Kubernetes service: %v", err)
		k8sService = nil
	}
	deploymentService := service.NewDeploymentService(deploymentRepo, k8sService)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	pipelineHandler := handler.NewPipelineHandler(pipelineService, gitService)
	buildHandler := handler.NewBuildHandler(buildService, gitService, dockerService, k8sService)
	deploymentHandler := handler.NewDeploymentHandler(deploymentService, k8sService)
	webhookHandler := handler.NewWebhookHandler(projectService, pipelineService, buildService)

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Initialize router
	r := gin.New()

	// Middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Set JWT secret in context for handlers
	r.Use(func(c *gin.Context) {
		c.Set("jwt_secret", cfg.JWT.Secret)
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "YS Cloud API is running",
		})
	})

	// Ready check
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ready",
			"message": "YS Cloud API is ready",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			public.POST("/auth/register", userHandler.Register)
			public.POST("/auth/login", userHandler.Login)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.DELETE("/profile", userHandler.DeleteProfile)
			}

			// Project routes
			projects := protected.Group("/projects")
			{
				projects.POST("/", projectHandler.CreateProject)
				projects.GET("/", projectHandler.GetProjects)
				projects.GET("/:id", projectHandler.GetProject)
				projects.PUT("/:id", projectHandler.UpdateProject)
				projects.DELETE("/:id", projectHandler.DeleteProject)
				projects.POST("/:id/collaborators", projectHandler.AddCollaborator)
				projects.DELETE("/:id/collaborators/:userId", projectHandler.RemoveCollaborator)
			}

			// Pipeline routes
			pipelines := protected.Group("/pipelines")
			{
				pipelines.POST("/", pipelineHandler.CreatePipeline)
				pipelines.GET("/", pipelineHandler.GetPipelines)
				pipelines.GET("/:id", pipelineHandler.GetPipeline)
				pipelines.PUT("/:id", pipelineHandler.UpdatePipeline)
				pipelines.DELETE("/:id", pipelineHandler.DeletePipeline)
				pipelines.POST("/:id/triggers", pipelineHandler.AddTrigger)
				pipelines.PUT("/:id/triggers/:triggerId", pipelineHandler.UpdateTrigger)
				pipelines.DELETE("/:id/triggers/:triggerId", pipelineHandler.RemoveTrigger)
				pipelines.POST("/:id/run", pipelineHandler.RunPipeline)
			}

			// Build routes
			builds := protected.Group("/builds")
			{
				builds.GET("/", buildHandler.GetBuilds)
				builds.GET("/:id", buildHandler.GetBuild)
				builds.GET("/:id/logs", buildHandler.GetBuildLogs)
				builds.POST("/:id/cancel", buildHandler.CancelBuild)
			}

			// Deployment routes
			deployments := protected.Group("/deployments")
			{
				deployments.GET("/", deploymentHandler.GetDeployments)
				deployments.GET("/:id", deploymentHandler.GetDeployment)
				deployments.GET("/:id/logs", deploymentHandler.GetDeploymentLogs)
				deployments.POST("/:id/rollback", deploymentHandler.RollbackDeployment)
			}
		}
	}

	// Webhook routes (public, secured by secret)
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/github/:projectSecret", webhookHandler.HandleGitHub)
		webhooks.POST("/gitlab/:projectSecret", webhookHandler.HandleGitLab)
		webhooks.POST("/gitee/:projectSecret", webhookHandler.HandleGitee)
	}

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger documentation available at http://localhost%s/swagger/index.html", port)

	return r.Run(port)
}

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Failed to start YS Cloud: %v", err)
	}
}