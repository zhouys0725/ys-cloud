package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Role      string         `json:"role" gorm:"default:user"`
	Avatar    string         `json:"avatar"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Projects []Project `json:"projects" gorm:"many2many:user_projects;"`
}

type Project struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	GitURL      string         `json:"git_url"`
	GitProvider string         `json:"git_provider"`
	OwnerID     uint           `json:"owner_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Owner     User         `json:"owner" gorm:"foreignKey:OwnerID"`
	Pipelines []Pipeline   `json:"pipelines"`
}

type Pipeline struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	ProjectID   uint           `json:"project_id"`
	Config      string         `json:"config" gorm:"type:text"`
	Status      string         `json:"status" gorm:"default:inactive"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Project  Project        `json:"project" gorm:"foreignKey:ProjectID"`
	Builds   []Build        `json:"builds"`
	Triggers []PipelineTrigger `json:"triggers"`
}

type PipelineTrigger struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PipelineID uint          `json:"pipeline_id"`
	Type      string         `json:"type"` // webhook, schedule, manual
	Branch    string         `json:"branch"`
	Tag       string         `json:"tag"`
	Schedule  string         `json:"schedule"` // cron expression
	Active    bool           `json:"active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Pipeline Pipeline `json:"pipeline" gorm:"foreignKey:PipelineID"`
}

type Build struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	PipelineID  uint           `json:"pipeline_id"`
	CommitHash  string         `json:"commit_hash"`
	Branch      string         `json:"branch"`
	Tag         string         `json:"tag"`
	Status      string         `json:"status"` // pending, running, success, failed, cancelled
	Logs        string         `json:"logs" gorm:"type:text"`
	ImageName   string         `json:"image_name"`
	ImageTag    string         `json:"image_tag"`
	StartedAt   *time.Time     `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Pipeline    Pipeline      `json:"pipeline" gorm:"foreignKey:PipelineID"`
	Deployments []Deployment  `json:"deployments"`
}

type Deployment struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	BuildID      uint           `json:"build_id"`
	Environment  string         `json:"environment"` // dev, staging, prod
	Status       string         `json:"status"`      // pending, running, success, failed, cancelled
	Replicas     int32          `json:"replicas"`
	Namespace    string         `json:"namespace"`
	ServiceName  string         `json:"service_name"`
	IngressHost  string         `json:"ingress_host"`
	StartedAt    *time.Time     `json:"started_at"`
	CompletedAt  *time.Time     `json:"completed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	Build Build `json:"build" gorm:"foreignKey:BuildID"`
}

type EnvironmentVariable struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProjectID uint           `json:"project_id"`
	Key       string         `json:"key" gorm:"not null"`
	Value     string         `json:"value" gorm:"not null"`
	Secret    bool           `json:"secret" gorm:"default:false"`
	Scope     string         `json:"scope"` // build, deploy, all
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Project Project `json:"project" gorm:"foreignKey:ProjectID"`
}

type WebhookLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProjectID uint           `json:"project_id"`
	Event     string         `json:"event"`
	Payload   string         `json:"payload" gorm:"type:text"`
	Processed bool           `json:"processed" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Project Project `json:"project" gorm:"foreignKey:ProjectID"`
}