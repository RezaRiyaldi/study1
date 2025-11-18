package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Environtment string
	Name         string
	Version      string
	Protocol     string
	Host         string
	Port         string
	BasePath     string
	URL          string
}

type DatabaseConfig struct {
	Host     string
	Name     string
	Port     string
	User     string
	Password string
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Environtment: getEnv("APP_ENVIRONMENT", "development"),
			Name:         getEnv("APP_NAME", "Study1"),
			Version:      getEnv("APP_VERSION", "1.0.0"),
			Protocol:     getEnv("APP_PROTOCOL", "http"),
			Host:         getEnv("APP_HOST", "localhost"),
			Port:         getEnv("APP_PORT", "8080"),
			BasePath:     getEnv("APP_BASE_PATH", "/api/v1/"),
			URL:          getEnv("APP_URL", "http://localhost:8080/api/v1/"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Name:     getEnv("DB_NAME", "study1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultVal
}

func (dbCfg DatabaseConfig) GetDSNMySQL() string {
	return dbCfg.User + ":" + dbCfg.Password + "@tcp(" + dbCfg.Host + ":" + dbCfg.Port + ")/" + dbCfg.Name + "?charset=utf8mb4&parseTime=True&loc=Asia%2FJakarta"
}
