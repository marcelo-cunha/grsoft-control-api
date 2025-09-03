package handlers

import (
	"errors"
	"net/http"
	"strings"

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

// handlePlatformError trata erros específicos das plataformas
func (sh *StoreHandler) handlePlatformError(c echo.Context, err error) error {
	// Verifica se é um erro específico do DeliveryVip
	var deliveryVipErr *services.DeliveryVipError
	if errors.As(err, &deliveryVipErr) {
		var statusCode int
		switch deliveryVipErr.TipoErro {
		case models.ErroNaoEncontrado:
			statusCode = http.StatusNotFound
		case models.ErroNaoAutorizado:
			statusCode = http.StatusUnauthorized
		case models.ErroRequisicaoInvalida:
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusBadGateway
		}

		return c.JSON(statusCode, models.RespostaErro{
			Error:    deliveryVipErr.TipoErro,
			Mensagem: deliveryVipErr.Mensagem,
		})
	}

	// Verifica se é erro de plataforma não suportada
	if err.Error() == "plataforma não suportada: "+c.Param("plataforma") {
		return c.JSON(http.StatusNotFound, models.RespostaErro{
			Error:    models.ErroNaoEncontrado,
			Mensagem: err.Error(),
		})
	}

	// Erro genérico - bad gateway
	return c.JSON(http.StatusBadGateway, models.RespostaErro{
		Error:    models.ErroBadGateway,
		Mensagem: "Erro ao comunicar com a plataforma: " + err.Error(),
	})
}

// handleBulkOperation gerencia operações em lote (ativar/desativar) com validação comum
func (sh *StoreHandler) handleBulkOperation(c echo.Context, operation func(string, []string) (*models.RespostaOperacaoMultiplasLojas, error)) error {
	plataforma := models.Plataforma(c.Param("plataforma"))

	// Valida parâmetro obrigatório
	if plataforma == "" {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:    models.ErroRequisicaoInvalida,
			Mensagem: "Parâmetro plataforma é obrigatório",
		})
	}

	// Decodifica o body da requisição
	var req models.RequisicaoMultiplasLojas
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:    models.ErroRequisicaoInvalida,
			Mensagem: "Body da requisição inválido: " + err.Error(),
		})
	}

	// Valida se há IDs no body
	if len(req.IdsLojas) == 0 {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:    models.ErroRequisicaoInvalida,
			Mensagem: "Campo 'ids_lojas' é obrigatório e deve conter pelo menos um ID",
		})
	}

	// Executa a operação específica
	response, err := operation(string(plataforma), req.IdsLojas)
	if err != nil {
		return sh.handlePlatformError(c, err)
	}

	return c.JSON(http.StatusOK, response)
}

// ActivateMultiple gerencia PATCH /plataformas/{plataforma}/lojas/ativar
func (sh *StoreHandler) ActivateMultiple(c echo.Context) error {
	return sh.handleBulkOperation(c, sh.platformService.ActivateMultipleStores)
}

// DeactivateMultiple gerencia PATCH /plataformas/{plataforma}/lojas/desativar
func (sh *StoreHandler) DeactivateMultiple(c echo.Context) error {
	return sh.handleBulkOperation(c, sh.platformService.DeactivateMultipleStores)
}

// GetMultipleStatus gerencia GET /plataformas/{plataforma}/lojas/status
// Os IDs das lojas podem ser passados no header "X-Lojas-IDs" separados por vírgula
// Se não informar o header, retorna o status de todas as lojas da plataforma
func (sh *StoreHandler) GetMultipleStatus(c echo.Context) error {
	plataforma := models.Plataforma(c.Param("plataforma"))
	idsParam := c.Request().Header.Get("X-Lojas-IDs")

	// Valida parâmetros obrigatórios
	if plataforma == "" {
		return c.JSON(http.StatusBadRequest, models.RespostaErro{
			Error:    models.ErroRequisicaoInvalida,
			Mensagem: "Parâmetro plataforma é obrigatório",
		})
	}

	// Processa os IDs se fornecidos
	var idsLojas []string
	if idsParam != "" {
		// Separa os IDs por vírgula e remove espaços em branco
		for _, id := range strings.Split(idsParam, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				idsLojas = append(idsLojas, id)
			}
		}

		if len(idsLojas) == 0 {
			return c.JSON(http.StatusBadRequest, models.RespostaErro{
				Error:    models.ErroRequisicaoInvalida,
				Mensagem: "IDs inválidos no header X-Lojas-IDs",
			})
		}
	}
	// Se idsParam estiver vazio, idsLojas será nil e o service retornará todas as lojas

	// Chama o serviço da plataforma
	response, err := sh.platformService.GetMultipleStoreStatus(plataforma, idsLojas)
	if err != nil {
		// Verifica o tipo de erro para determinar o status HTTP apropriado
		if strings.Contains(err.Error(), "plataforma não suportada") {
			return c.JSON(http.StatusNotFound, models.RespostaErro{
				Error:    models.ErroNaoEncontrado,
				Mensagem: err.Error(),
			})
		}

		// Erro genérico - poderia ser bad gateway em implementação real
		return c.JSON(http.StatusBadGateway, models.RespostaErro{
			Error:    models.ErroBadGateway,
			Mensagem: "Erro ao comunicar com a plataforma: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
