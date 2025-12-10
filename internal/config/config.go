package config

import (
	"fmt"
	"os"
	"strings"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Docker   DockerConfig   `mapstructure:"docker"`
	K8s      K8sConfig      `mapstructure:"k8s"`
	Git      GitConfig      `mapstructure:"git"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port    string `mapstructure:"port"`
	GinMode string `mapstructure:"gin_mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpiresIn  string `mapstructure:"expires_in"`
}

type DockerConfig struct {
	Host     string `mapstructure:"host"`
	Registry string `mapstructure:"registry"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type K8sConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	Namespace  string `mapstructure:"namespace"`
}

type GitConfig struct {
	GitHubClientID     string `mapstructure:"github_client_id"`
	GitHubClientSecret string `mapstructure:"github_client_secret"`
	GitLabClientID     string `mapstructure:"gitlab_client_id"`
	GitLabClientSecret string `mapstructure:"gitlab_client_secret"`
}

type StorageConfig struct {
	Type     string `mapstructure:"type"`
	Path     string `mapstructure:"path"`
	AWS      AWSConfig `mapstructure:"aws"`
}

type AWSConfig struct {
	AccessKeyID string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Region string `mapstructure:"region"`
	S3Bucket string `mapstructure:"s3_bucket"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./")
	viper.AddConfigPath("../")
	viper.AddConfigPath("../../")

	// Auto read from environment variables first
	viper.AutomaticEnv()

	// Set environment variable prefixes for nested structs
	viper.SetEnvPrefix("") // No prefix for top-level
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values (will be overridden by environment variables)
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.gin_mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.db_name", "ys_cloud")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "ys-cloud-default-jwt-secret-key")
	viper.SetDefault("jwt.expires_in", "168h")
	viper.SetDefault("docker.host", "unix:///var/run/docker.sock")
	viper.SetDefault("docker.registry", "registry.hub.docker.com")
	viper.SetDefault("k8s.namespace", "default")
	viper.SetDefault("storage.type", "local")
	viper.SetDefault("storage.path", "./uploads")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use environment variables and defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Override with direct environment variable reading to ensure values are loaded
	if os.Getenv("DATABASE_HOST") != "" {
		config.Database.Host = os.Getenv("DATABASE_HOST")
	}
	if os.Getenv("DATABASE_PORT") != "" {
		config.Database.Port = os.Getenv("DATABASE_PORT")
	}
	if os.Getenv("DATABASE_USER") != "" {
		config.Database.User = os.Getenv("DATABASE_USER")
	}
	if os.Getenv("DATABASE_PASSWORD") != "" {
		config.Database.Password = os.Getenv("DATABASE_PASSWORD")
	}
	if os.Getenv("DATABASE_DB_NAME") != "" {
		config.Database.DBName = os.Getenv("DATABASE_DB_NAME")
	}
	if os.Getenv("DATABASE_SSL_MODE") != "" {
		config.Database.SSLMode = os.Getenv("DATABASE_SSL_MODE")
	}

	if os.Getenv("REDIS_HOST") != "" {
		config.Redis.Host = os.Getenv("REDIS_HOST")
	}
	if os.Getenv("REDIS_PORT") != "" {
		config.Redis.Port = os.Getenv("REDIS_PORT")
	}
	if os.Getenv("REDIS_PASSWORD") != "" {
		config.Redis.Password = os.Getenv("REDIS_PASSWORD")
	}

	// Debug: print configuration
	fmt.Printf("Database Config: %+v\n", config.Database)
	fmt.Printf("Redis Config: %+v\n", config.Redis)

	return &config, nil
}