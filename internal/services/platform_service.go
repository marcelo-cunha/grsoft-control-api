package services

import (
	"fmt"

	"delivery-control/internal/config"
	"delivery-control/internal/models"
)

// PlatformService gerencia a comunicação com plataformas externas
type PlatformService struct {
	anotaAiService *AnotaAiService
}

// NewPlatformService cria um novo serviço de plataforma
func NewPlatformService(cfg *config.Config) *PlatformService {
	return &PlatformService{
		anotaAiService: NewAnotaAiService(cfg),
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
			Message:    "Loja ativada com sucesso",
		}, nil
	case models.PlataformaDeliveryVip:
		// Implementação mock para DeliveryVip - implementar quando necessário
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusAtivo,
			Message:    "Loja ativada com sucesso (mock)",
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
			Message:    "Loja desativada com sucesso",
		}, nil
	case models.PlataformaDeliveryVip:
		// Implementação mock para DeliveryVip - implementar quando necessário
		return &models.RespostaOperacaoLoja{
			Plataforma: plataforma,
			IdLoja:     idLoja,
			Status:     models.StatusInativo,
			Message:    "Loja desativada com sucesso (mock)",
		}, nil
	default:
		return nil, fmt.Errorf("plataforma não implementada: %s", plataforma)
	}
}

// GetMultipleStoreStatus obtém o status de múltiplas lojas na plataforma especificada
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

		// Converte o mapa para a estrutura de resposta
		lojas := make([]models.StatusLojaDetalhes, 0, len(idsLojas))
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

		return &models.RespostaStatusMultiplasLojas{
			Plataforma: plataforma,
			Lojas:      lojas,
		}, nil
	case models.PlataformaDeliveryVip:
		// Implementação mock para DeliveryVip - implementar quando necessário
		lojas := make([]models.StatusLojaDetalhes, 0, len(idsLojas))
		for _, idLoja := range idsLojas {
			lojas = append(lojas, models.StatusLojaDetalhes{
				IdLoja: idLoja,
				Status: models.StatusAtivo, // Mock sempre ativo
			})
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
