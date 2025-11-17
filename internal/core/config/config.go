package config

import (
	"os"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Name         string
	Version      string
	Port         string
	Environtment string
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
			Name:         getEnv("APP_NAME", "Study1"),
			Version:      getEnv("APP_VERSION", "1.0.0"),
			Port:         getEnv("SERVER_PORT", "8080"),
			Environtment: getEnv("SERVER_ENV", "development"),
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
