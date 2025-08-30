package main

import (
	"fmt"
	"log"

	"delivery-control/internal/api/handlers"
	"delivery-control/internal/api/routes"
	"delivery-control/internal/config"
	"delivery-control/internal/services"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	// Carrega o arquivo .env se existir
	_ = godotenv.Load()

	// Carrega a configuração
	cfg := config.Load()

	// Valida a configuração obrigatória
	if cfg.Auth.BearerToken == "" {
		log.Fatal("A variábel de ambiente BEARER_TOKEN é obrigatória")
	}

	// Inicializa os serviços
	platformService := services.NewPlatformService(cfg)

	// Inicializa os handlers
	healthHandler := handlers.NewHealthHandler()
	storeHandler := handlers.NewStoreHandler(platformService)

	// Cria a instância do Echo
	e := echo.New()

	// Configura as rotas
	routes.SetupRoutes(e, cfg, healthHandler, storeHandler)

	// Inicia o servidor
	address := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Iniciando o servidor em %s", address)
	log.Fatal(e.Start(address))
}
