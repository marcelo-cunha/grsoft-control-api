package services

import (
	"fmt"

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
			Status:     models.StatusInativo,
			Mensagem:   "Loja desativada com sucesso",
		}, nil
	case models.PlataformaDeliveryVip:
		if err := ps.deliveryVipService.DeactivateStore(idLoja); err != nil {
			return nil, fmt.Errorf("erro ao desativar loja no DeliveryVip: %w", err)
		}
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusInativo,
			Mensagem:   "Loja desativada com sucesso",
		}, nil
	default:
		return nil, fmt.Errorf("plataforma não implementada: %s", plataforma)
	}
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
				status := models.StatusInativo
				if isActive, exists := statusMap[idLoja]; exists && isActive {
					status = models.StatusAtivo
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja: idLoja,
					Status: status,
				})
			}
		} else {
			lojas = make([]models.StatusLojaDetalhes, 0, len(statusMap))
			for idLoja, isActive := range statusMap {
				status := models.StatusInativo
				if isActive {
					status = models.StatusAtivo
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja: idLoja,
					Status: status,
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
				if result, exists := statusMap[idLoja]; exists {
					if !result.Found {
						status = models.StatusNaoEncontrado
					} else if result.IsActive {
						status = models.StatusAtivo
					} else {
						status = models.StatusInativo
					}
				} else {
					// Fallback case (não deveria acontecer)
					status = models.StatusNaoEncontrado
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja: idLoja,
					Status: status,
				})
			}
		} else {
			lojas = make([]models.StatusLojaDetalhes, 0, len(statusMap))
			for idLoja, result := range statusMap {
				var status models.Status
				if !result.Found {
					status = models.StatusNaoEncontrado
				} else if result.IsActive {
					status = models.StatusAtivo
				} else {
					status = models.StatusInativo
				}

				lojas = append(lojas, models.StatusLojaDetalhes{
					IdLoja: idLoja,
					Status: status,
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
