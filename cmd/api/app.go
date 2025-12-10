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
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Run() error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize database: %v. Continuing without database.", err)
		// 不返回错误，继续初始化其他组件
	}

	// Initialize repositories
	var userRepo *repository.UserRepository
	var projectRepo *repository.ProjectRepository
	var pipelineRepo *repository.PipelineRepository
	var buildRepo *repository.BuildRepository
	var deploymentRepo *repository.DeploymentRepository
	
	// 只有在数据库连接成功时才初始化仓库
	if db != nil {
		userRepo = repository.NewUserRepository(db)
		projectRepo = repository.NewProjectRepository(db)
		pipelineRepo = repository.NewPipelineRepository(db)
		buildRepo = repository.NewBuildRepository(db)
		deploymentRepo = repository.NewDeploymentRepository(db)
	}

	// Initialize services
	var userService *service.UserService
	var projectService *service.ProjectService
	var pipelineService *service.PipelineService
	var buildService *service.BuildService
	var deploymentService *service.DeploymentService
	
	// 根据仓库是否初始化来决定服务初始化
	if userRepo != nil {
		userService = service.NewUserService(userRepo)
	}
	
	if projectRepo != nil && userRepo != nil {
		projectService = service.NewProjectService(projectRepo, userRepo)
	}
	
	if pipelineRepo != nil && projectRepo != nil {
		pipelineService = service.NewPipelineService(pipelineRepo, projectRepo)
	}
	
	if buildRepo != nil && pipelineRepo != nil {
		buildService = service.NewBuildService(buildRepo, pipelineRepo)
	}
	
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
	
	if deploymentRepo != nil {
		deploymentService = service.NewDeploymentService(deploymentRepo, k8sService)
	}

	// Initialize handlers
	var userHandler *handler.UserHandler
	var projectHandler *handler.ProjectHandler
	var pipelineHandler *handler.PipelineHandler
	var buildHandler *handler.BuildHandler
	var deploymentHandler *handler.DeploymentHandler
	
	// 根据服务是否初始化来决定处理器初始化
	if userService != nil {
		userHandler = handler.NewUserHandler(userService)
	}
	
	if projectService != nil {
		projectHandler = handler.NewProjectHandler(projectService)
	}
	
	if pipelineService != nil && gitService != nil {
		pipelineHandler = handler.NewPipelineHandler(pipelineService, gitService)
	}
	
	if buildService != nil && gitService != nil && dockerService != nil && k8sService != nil {
		buildHandler = handler.NewBuildHandler(buildService, gitService, dockerService, k8sService)
	}
	
	if deploymentService != nil && k8sService != nil {
		deploymentHandler = handler.NewDeploymentHandler(deploymentService, k8sService)
	}
	
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

	// Health check - 简化健康检查端点，不依赖任何外部服务
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "YS Cloud API is running",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes - 只注册那些所有依赖都已初始化的路由
	v1 := r.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			if userHandler != nil {
				public.POST("/auth/register", userHandler.Register)
				public.POST("/auth/login", userHandler.Login)
			}
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.JWTAuth(cfg.JWT.Secret))
		{
			// User routes
			if userHandler != nil {
				users := protected.Group("/users")
				{
					users.GET("/profile", userHandler.GetProfile)
					users.PUT("/profile", userHandler.UpdateProfile)
					users.DELETE("/profile", userHandler.DeleteProfile)
				}
			}

			// Project routes
			if projectHandler != nil {
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
			}

			// Pipeline routes
			if pipelineHandler != nil {
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
			}

			// Build routes
			if buildHandler != nil {
				builds := protected.Group("/builds")
				{
					builds.GET("/", buildHandler.GetBuilds)
					builds.GET("/:id", buildHandler.GetBuild)
					builds.GET("/:id/logs", buildHandler.GetBuildLogs)
					builds.POST("/:id/cancel", buildHandler.CancelBuild)
				}
			}

			// Deployment routes
			if deploymentHandler != nil {
				deployments := protected.Group("/deployments")
				{
					deployments.GET("/", deploymentHandler.GetDeployments)
					deployments.GET("/:id", deploymentHandler.GetDeployment)
					deployments.GET("/:id/logs", deploymentHandler.GetDeploymentLogs)
					deployments.POST("/:id/rollback", deploymentHandler.RollbackDeployment)
				}
			}
		}
	}

	// Webhook routes (public, secured by secret)
	if webhookHandler != nil {
		webhooks := r.Group("/webhooks")
		{
			webhooks.POST("/github/:projectSecret", webhookHandler.HandleGitHub)
			webhooks.POST("/gitlab/:projectSecret", webhookHandler.HandleGitLab)
			webhooks.POST("/gitee/:projectSecret", webhookHandler.HandleGitee)
		}
	}

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger documentation available at http://localhost%s/swagger/index.html", port)

	return r.Run(port)
}