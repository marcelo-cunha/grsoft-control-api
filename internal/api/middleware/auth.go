package middleware

import (
	"net/http"
	"strings"

	"delivery-control/internal/config"
	"delivery-control/internal/models"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware cria um novo middleware de autenticação
func AuthMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extrai o header de Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, models.RespostaErro{
					Error:    models.ErroNaoAutorizado,
					Mensagem: "Token de autorização é obrigatório",
				})
			}

			// Verifica se começa com "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, models.RespostaErro{
					Error:    models.ErroNaoAutorizado,
					Mensagem: "Formato do token inválido. Use 'Bearer <token>'",
				})
			}

			// Extrai o token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, models.RespostaErro{
					Error:    models.ErroNaoAutorizado,
					Mensagem: "Token não fornecido",
				})
			}

			// Valida o token contra o token configurado
			if token != cfg.Auth.BearerToken {
				return c.JSON(http.StatusUnauthorized, models.RespostaErro{
					Error:    models.ErroNaoAutorizado,
					Mensagem: "Token inválido",
				})
			}

			// Token é válido, prossegue para o próximo handler
			return next(c)
		}
	}
}
