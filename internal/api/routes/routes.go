package routes

import (
	"delivery-control/internal/api/handlers"
	"delivery-control/internal/api/middleware"
	"delivery-control/internal/config"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// SetupRoutes configura todas as rotas da aplicação
func SetupRoutes(e *echo.Echo, cfg *config.Config, healthHandler *handlers.HealthHandler, storeHandler *handlers.StoreHandler, docsHandler *handlers.DocsHandler) {
	// Adiciona middleware comum
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())

	// Cria um grupo para rotas públicas (sem autenticação)
	public := e.Group("")

	// Health check
	public.GET("/health", healthHandler.Check)

	// Documentação da API
	public.GET("/docs", docsHandler.ServeHTML)
	public.GET("/docs/openapi.yml", docsHandler.ServeOpenAPI)

	// Cria um grupo para rotas protegidas com logger
	protected := e.Group("")
	protected.Use(echomiddleware.Logger())
	protected.Use(middleware.AuthMiddleware(cfg))

	// Operações de loja
	protected.POST("/plataformas/:plataforma/lojas/ativar", storeHandler.ActivateMultiple)
	protected.POST("/plataformas/:plataforma/lojas/desativar", storeHandler.DeactivateMultiple)
	protected.GET("/plataformas/:plataforma/lojas/status", storeHandler.GetMultipleStatus)
}
