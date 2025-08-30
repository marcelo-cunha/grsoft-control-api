package routes

import (
	"delivery-control/internal/api/handlers"
	"delivery-control/internal/api/middleware"
	"delivery-control/internal/config"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(e *echo.Echo, cfg *config.Config, healthHandler *handlers.HealthHandler, storeHandler *handlers.StoreHandler) {
	// Adiciona middleware comum
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Endpoint de verificação de saúde (sem autenticação e sem logs)
	e.GET("/health", healthHandler.Check)

	// Cria um grupo para rotas protegidas com logger
	protected := e.Group("")
	protected.Use(echomiddleware.Logger())
	protected.Use(middleware.AuthMiddleware(cfg))

	// Rotas de loja com autenticação
	protected.POST("/plataformas/:plataforma/lojas/:store_id/ativar", storeHandler.Activate)
	protected.POST("/plataformas/:plataforma/lojas/:store_id/desativar", storeHandler.Deactivate)
	protected.GET("/plataformas/:plataforma/lojas/:store_id/status", storeHandler.GetStatus)
}
