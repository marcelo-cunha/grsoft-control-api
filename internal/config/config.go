package config

import (
	"os"
	"strconv"
)

// Config contém toda a configuração da aplicação
type Config struct {
	Server    ServerConfig
	Auth      AuthConfig
	Platforms PlatformConfig
}

// ServerConfig contém a configuração do servidor
type ServerConfig struct {
	Port string
}

// AuthConfig contém a configuração de autenticação
type AuthConfig struct {
	BearerToken string
}

// PlatformConfig contém as URLs das plataformas para implementação futura
type PlatformConfig struct {
	AnotaAiURL     string
	DeliveryVipURL string
	AnotaAi        AnotaAiConfig
}

// AnotaAiConfig contém as configurações específicas do AnotaAI
type AnotaAiConfig struct {
	Email    string
	Password string
}

// Load carrega a configuração das variáveis de ambiente
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Auth: AuthConfig{
			BearerToken: getEnv("BEARER_TOKEN", ""),
		},
		Platforms: PlatformConfig{
			AnotaAiURL:     getEnv("ANOTAAI_API_URL", "https://integration-admin.api.anota.ai"),
			DeliveryVipURL: getEnv("DELIVERYVIP_API_URL", "https://api.deliveryvip.com"),
			AnotaAi: AnotaAiConfig{
				Email:    getEnv("ANOTAAI_EMAIL", ""),
				Password: getEnv("ANOTAAI_PASSWORD", ""),
			},
		},
	}
}

// getEnv obtém uma variável de ambiente com um valor padrão
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt obtém uma variável de ambiente como int com um valor padrão
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
