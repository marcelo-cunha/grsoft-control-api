package handlers

import (
	"net/http"

	"delivery-control/internal/models"
	"delivery-control/internal/services"

	"github.com/labstack/echo/v4"
)

// StoreHandler gerencia requisições relacionadas às lojas
type StoreHandler struct {
	platformService *services.PlatformService
}

// NewStoreHandler cria um novo handler de loja
func NewStoreHandler(platformService *services.PlatformService) *StoreHandler {
	return &StoreHandler{
		platformService: platformService,
	}
}

// Activate gerencia POST /plataformas/{plataforma}/lojas/{id_loja}/ativar
func (sh *StoreHandler) Activate(c echo.Context) error {
	plataforma := models.Plataforma(c.Param("plataforma"))
	idLoja := c.Param("store_id")

	// Valida parâmetros obrigatórios
	if plataforma == "" || idLoja == "" {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:   models.ErroRequisicaoInvalida,
			Message: "Parâmetros plataforma e id_loja são obrigatórios",
		})
	}

	// Chama o serviço da plataforma
	response, err := sh.platformService.ActivateStore(plataforma, idLoja)
	if err != nil {
		// Verifica o tipo de erro para determinar o status HTTP apropriado
		if err.Error() == "plataforma não suportada: "+string(plataforma) {
			return c.JSON(http.StatusNotFound, models.RespostaErro{
				Error:   models.ErroNaoEncontrado,
				Message: err.Error(),
			})
		}

		// Erro genérico - poderia ser bad gateway em implementação real
		return c.JSON(http.StatusBadGateway, models.RespostaErro{
			Error:   models.ErroBadGateway,
			Message: "Erro ao comunicar com a plataforma: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Deactivate gerencia POST /plataformas/{plataforma}/lojas/{id_loja}/desativar
func (sh *StoreHandler) Deactivate(c echo.Context) error {
	plataforma := models.Plataforma(c.Param("plataforma"))
	idLoja := c.Param("store_id")

	// Valida parâmetros obrigatórios
	if plataforma == "" || idLoja == "" {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:   models.ErroRequisicaoInvalida,
			Message: "Parâmetros plataforma e id_loja são obrigatórios",
		})
	}

	// Chama o serviço da plataforma
	response, err := sh.platformService.DeactivateStore(plataforma, idLoja)
	if err != nil {
		// Verifica o tipo de erro para determinar o status HTTP apropriado
		if err.Error() == "plataforma não suportada: "+string(plataforma) {
			return c.JSON(http.StatusNotFound, models.RespostaErro{
				Error:   models.ErroNaoEncontrado,
				Message: err.Error(),
			})
		}

		// Erro genérico - poderia ser bad gateway em implementação real
		return c.JSON(http.StatusBadGateway, models.RespostaErro{
			Error:   models.ErroBadGateway,
			Message: "Erro ao comunicar com a plataforma: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// GetStatus gerencia GET /plataformas/{plataforma}/lojas/{id_loja}/status
func (sh *StoreHandler) GetStatus(c echo.Context) error {
	plataforma := models.Plataforma(c.Param("plataforma"))
	idLoja := c.Param("store_id")

	// Valida parâmetros obrigatórios
	if plataforma == "" || idLoja == "" {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:   models.ErroRequisicaoInvalida,
			Message: "Parâmetros plataforma e id_loja são obrigatórios",
		})
	}

	// Chama o serviço da plataforma
	response, err := sh.platformService.GetStoreStatus(plataforma, idLoja)
	if err != nil {
		// Verifica o tipo de erro para determinar o status HTTP apropriado
		if err.Error() == "plataforma não suportada: "+string(plataforma) {
			return c.JSON(http.StatusNotFound, models.RespostaErro{
				Error:   models.ErroNaoEncontrado,
				Message: err.Error(),
			})
		}

		// Erro genérico - poderia ser bad gateway em implementação real
		return c.JSON(http.StatusBadGateway, models.RespostaErro{
			Error:   models.ErroBadGateway,
			Message: "Erro ao comunicar com a plataforma: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
