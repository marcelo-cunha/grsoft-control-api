package handlers

import (
	"net/http"

	"delivery-control/internal/models"

	"github.com/labstack/echo/v4"
)

// HealthHandler gerencia requisições de verificação de saúde
type HealthHandler struct{}

// NewHealthHandler cria um novo handler de saúde
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check gerencia GET /health
func (h *HealthHandler) Check(c echo.Context) error {
	response := models.RespostaSaude{
		Status: "ok",
	}

	return c.JSON(http.StatusOK, response)
}
