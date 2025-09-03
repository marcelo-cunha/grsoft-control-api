package services

import (
	"fmt"
	"strings"

	"delivery-control/internal/config"
	"delivery-control/internal/models"
)

// PlatformService gerencia a comunicação com plataformas externas
type PlatformService struct {
	anotaAiService     *AnotaAiService
	deliveryVipService *DeliveryVipService
}

// NewPlatformService cria um novo serviço de plataforma
func NewPlatformService(cfg *config.Config) *PlatformService {
	return &PlatformService{
		anotaAiService:     NewAnotaAiService(cfg),
		deliveryVipService: NewDeliveryVipService(cfg),
	}
}

// ActivateStore ativa uma loja na plataforma especificada
func (ps *PlatformService) ActivateStore(plataforma models.Plataforma, idLoja string) (*models.RespostaOperacaoLoja, error) {
	// Valida a plataforma
	if !ps.isValidPlatform(plataforma) {
		return nil, fmt.Errorf("plataforma não suportada: %s", plataforma)
	}

	// Chama o serviço específico baseado na plataforma
	switch plataforma {
	case models.PlataformaAnotaAi:
		if err := ps.anotaAiService.ActivateStore(idLoja); err != nil {
			return nil, fmt.Errorf("erro ao ativar loja no AnotaAI: %w", err)
		}
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusAtivo,
			Mensagem:   "Loja ativada com sucesso",
		}, nil
	case models.PlataformaDeliveryVip:
		if err := ps.deliveryVipService.ActivateStore(idLoja); err != nil {
			return nil, fmt.Errorf("erro ao ativar loja no DeliveryVip: %w", err)
		}
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusAtivo,
			Mensagem:   "Loja ativada com sucesso",
		}, nil
	default:
		return nil, fmt.Errorf("plataforma não implementada: %s", plataforma)
	}
}

// DeactivateStore desativa uma loja na plataforma especificada
func (ps *PlatformService) DeactivateStore(plataforma models.Plataforma, idLoja string) (*models.RespostaOperacaoLoja, error) {
	// Valida a plataforma
	if !ps.isValidPlatform(plataforma) {
		return nil, fmt.Errorf("plataforma não suportada: %s", plataforma)
	}

	// Chama o serviço específico baseado na plataforma
	switch plataforma {
	case models.PlataformaAnotaAi:
		if err := ps.anotaAiService.DeactivateStore(idLoja); err != nil {
			return nil, fmt.Errorf("erro ao desativar loja no AnotaAI: %w", err)
		}
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusBloqueado,
			Mensagem:   "Loja desativada com sucesso",
		}, nil
	case models.PlataformaDeliveryVip:
		if err := ps.deliveryVipService.DeactivateStore(idLoja); err != nil {
			return nil, fmt.Errorf("erro ao desativar loja no DeliveryVip: %w", err)
		}
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusBloqueado,
			Mensagem:   "Loja desativada com sucesso",
		}, nil
	default:
		return nil, fmt.Errorf("plataforma não implementada: %s", plataforma)
	}
}

// ActivateMultipleStores ativa múltiplas lojas em uma plataforma específica
func (ps *PlatformService) ActivateMultipleStores(plataforma string, idsLojas []string) (*models.RespostaOperacaoMultiplasLojas, error) {
	finalResponse := &models.RespostaOperacaoMultiplasLojas{
		Plataforma: models.Plataforma(plataforma),
		Resultados: make([]models.ResultadoOperacaoLoja, 0, len(idsLojas)),
	}

	// Processa cada loja individualmente
	for _, idLoja := range idsLojas {
		var err error

		switch plataforma {
		case "anotaai":
			err = ps.anotaAiService.ActivateStore(idLoja)
		case "deliveryvip":
			err = ps.deliveryVipService.ActivateStore(idLoja)
		default:
			return nil, fmt.Errorf("plataforma não suportada: %s", plataforma)
		}

		resultado := models.ResultadoOperacaoLoja{
			IdLoja: idLoja,
		}

		if err != nil {
			// Verifica se é um erro específico do DeliveryVip
			if deliveryVipErr, ok := err.(*DeliveryVipError); ok {
				switch deliveryVipErr.TipoErro {
				case models.ErroNaoEncontrado:
					resultado.Status = models.StatusNaoEncontrado
				default:
					resultado.Status = models.StatusNaoEncontrado
				}
				resultado.Sucesso = false
				resultado.Mensagem = fmt.Sprintf("Erro ao ativar loja: %s", deliveryVipErr.Mensagem)
				resultado.Erro = &deliveryVipErr.TipoErro
			} else if strings.Contains(err.Error(), "loja não encontrada") || strings.Contains(err.Error(), "store not found") {
				resultado.Status = models.StatusNaoEncontrado
				resultado.Sucesso = false
				resultado.Mensagem = "Loja não encontrada na plataforma"
				errType := models.ErroNaoEncontrado
				resultado.Erro = &errType
			} else {
				resultado.Status = models.StatusNaoEncontrado
				resultado.Sucesso = false
				resultado.Mensagem = fmt.Sprintf("Erro ao ativar loja: %s", err.Error())
				errType := models.ErroBadGateway
				resultado.Erro = &errType
			}
		} else {
			resultado.Status = models.StatusAtivo
			resultado.Sucesso = true
			resultado.Mensagem = "Loja ativada com sucesso"
		}

		finalResponse.Resultados = append(finalResponse.Resultados, resultado)
	}

	return finalResponse, nil
} // DeactivateMultipleStores desativa múltiplas lojas em uma plataforma específica
func (ps *PlatformService) DeactivateMultipleStores(plataforma string, idsLojas []string) (*models.RespostaOperacaoMultiplasLojas, error) {
	finalResponse := &models.RespostaOperacaoMultiplasLojas{
		Plataforma: models.Plataforma(plataforma),
		Resultados: make([]models.ResultadoOperacaoLoja, 0, len(idsLojas)),
	}

	// Processa cada loja individualmente
	for _, idLoja := range idsLojas {
		var err error

		switch plataforma {
		case "anotaai":
			err = ps.anotaAiService.DeactivateStore(idLoja)
		case "deliveryvip":
			err = ps.deliveryVipService.DeactivateStore(idLoja)
		default:
			return nil, fmt.Errorf("plataforma não suportada: %s", plataforma)
		}

		resultado := models.ResultadoOperacaoLoja{
			IdLoja: idLoja,
		}

		if err != nil {
			// Verifica se é um erro específico do DeliveryVip
			if deliveryVipErr, ok := err.(*DeliveryVipError); ok {
				switch deliveryVipErr.TipoErro {
				case models.ErroNaoEncontrado:
					resultado.Status = models.StatusNaoEncontrado
				default:
					resultado.Status = models.StatusNaoEncontrado
				}
				resultado.Sucesso = false
				resultado.Mensagem = fmt.Sprintf("Erro ao desativar loja: %s", deliveryVipErr.Mensagem)
				resultado.Erro = &deliveryVipErr.TipoErro
			} else if strings.Contains(err.Error(), "loja não encontrada") || strings.Contains(err.Error(), "store not found") {
				resultado.Status = models.StatusNaoEncontrado
				resultado.Sucesso = false
				resultado.Mensagem = "Loja não encontrada na plataforma"
				errType := models.ErroNaoEncontrado
				resultado.Erro = &errType
			} else {
				resultado.Status = models.StatusNaoEncontrado
				resultado.Sucesso = false
				resultado.Mensagem = fmt.Sprintf("Erro ao desativar loja: %s", err.Error())
				errType := models.ErroBadGateway
				resultado.Erro = &errType
			}
		} else {
			resultado.Status = models.StatusBloqueado
			resultado.Sucesso = true
			resultado.Mensagem = "Loja desativada com sucesso"
		}

		finalResponse.Resultados = append(finalResponse.Resultados, resultado)
	}

	return finalResponse, nil
}

// GetMultipleStoreStatus obtém o status de múltiplas lojas na plataforma especificada
// Se idsLojas for nil ou vazio, retorna o status de todas as lojas da plataforma
func (ps *PlatformService) GetMultipleStoreStatus(plataforma models.Plataforma, idsLojas []string) (*models.RespostaStatusMultiplasLojas, error) {
	// Valida a plataforma
	if !ps.isValidPlatform(plataforma) {
		return nil, fmt.Errorf("plataforma não suportada: %s", plataforma)
	}

	// Chama o serviço específico baseado na plataforma
	switch plataforma {
	case models.PlataformaAnotaAi:
		statusMap, err := ps.anotaAiService.GetMultipleStoreStatus(idsLojas)
		if err != nil {
			return nil, fmt.Errorf("erro ao consultar status no AnotaAI: %w", err)
		}

		// Se IDs específicos foram solicitados, itera sobre eles
		// Caso contrário, itera sobre todas as chaves do mapa
		var lojas []models.StatusLojaDetalhes
		if len(idsLojas) > 0 {
			lojas = make([]models.StatusLojaDetalhes, 0, len(idsLojas))
			for _, idLoja := range idsLojas {
				var status models.Status
				var documento, nomeFantasia string

				if storeInfo, exists := statusMap[idLoja]; exists {
					if !storeInfo.Found {
						status = models.StatusNaoEncontrado
					} else {
						// Usa o status específico retornado pelo serviço
						status = storeInfo.Status
					}
					documento = storeInfo.Documento
					nomeFantasia = storeInfo.NomeFantasia
				} else {
					status = models.StatusNaoEncontrado
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja:       idLoja,
					Status:       status,
					Documento:    documento,
					NomeFantasia: nomeFantasia,
				})
			}
		} else {
			lojas = make([]models.StatusLojaDetalhes, 0, len(statusMap))
			for idLoja, storeInfo := range statusMap {
				var status models.Status
				if !storeInfo.Found {
					status = models.StatusNaoEncontrado
				} else {
					// Usa o status específico retornado pelo serviço
					status = storeInfo.Status
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja:       idLoja,
					Status:       status,
					Documento:    storeInfo.Documento,
					NomeFantasia: storeInfo.NomeFantasia,
				})
			}
		}

		return &models.RespostaStatusMultiplasLojas{
			Plataforma: plataforma,
			Lojas:      lojas,
		}, nil
	case models.PlataformaDeliveryVip:
		statusMap, err := ps.deliveryVipService.GetMultipleStoreStatus(idsLojas)
		if err != nil {
			return nil, fmt.Errorf("erro ao consultar status das lojas no DeliveryVip: %w", err)
		}

		// Se IDs específicos foram solicitados, itera sobre eles
		// Caso contrário, itera sobre todas as chaves do mapa
		var lojas []models.StatusLojaDetalhes
		if len(idsLojas) > 0 {
			lojas = make([]models.StatusLojaDetalhes, 0, len(idsLojas))
			for _, idLoja := range idsLojas {
				var status models.Status
				var documento, nomeFantasia string

				if storeInfo, exists := statusMap[idLoja]; exists {
					if !storeInfo.Found {
						status = models.StatusNaoEncontrado
					} else {
						// Usa o status específico retornado pelo serviço
						status = storeInfo.Status
					}
					documento = storeInfo.Documento
					nomeFantasia = storeInfo.NomeFantasia
				} else {
					// Fallback case (não deveria acontecer)
					status = models.StatusNaoEncontrado
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja:       idLoja,
					Status:       status,
					Documento:    documento,
					NomeFantasia: nomeFantasia,
				})
			}
		} else {
			lojas = make([]models.StatusLojaDetalhes, 0, len(statusMap))
			for idLoja, storeInfo := range statusMap {
				var status models.Status
				if !storeInfo.Found {
					status = models.StatusNaoEncontrado
				} else {
					status = storeInfo.Status
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja:       idLoja,
					Status:       status,
					Documento:    storeInfo.Documento,
					NomeFantasia: storeInfo.NomeFantasia,
				})
			}
		}

		return &models.RespostaStatusMultiplasLojas{
			Plataforma: plataforma,
			Lojas:      lojas,
		}, nil
	default:
		return nil, fmt.Errorf("plataforma não implementada: %s", plataforma)
	}
}

// isValidPlatform verifica se a plataforma é suportada
func (ps *PlatformService) isValidPlatform(plataforma models.Plataforma) bool {
	return plataforma == models.PlataformaAnotaAi || plataforma == models.PlataformaDeliveryVip
}
