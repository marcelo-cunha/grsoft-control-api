package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"delivery-control/internal/config"
	"delivery-control/internal/models"
)

// DeliveryVipError representa um erro específico da API DeliveryVip
type DeliveryVipError struct {
	HTTPStatus int
	TipoErro   models.TipoErro
	Mensagem   string
}

func (e *DeliveryVipError) Error() string {
	return e.Mensagem
}

// NewDeliveryVipError cria um novo erro específico do DeliveryVip baseado no status HTTP
func NewDeliveryVipError(httpStatus int, responseBody string) error {
	switch httpStatus {
	case http.StatusNotFound:
		return &DeliveryVipError{
			HTTPStatus: httpStatus,
			TipoErro:   models.ErroNaoEncontrado,
			Mensagem:   "Loja não encontrada na plataforma",
		}
	case http.StatusUnauthorized:
		return &DeliveryVipError{
			HTTPStatus: httpStatus,
			TipoErro:   models.ErroNaoAutorizado,
			Mensagem:   "Erro de autenticação com a plataforma",
		}
	case http.StatusUnprocessableEntity:
		return &DeliveryVipError{
			HTTPStatus: httpStatus,
			TipoErro:   models.ErroRequisicaoInvalida,
			Mensagem:   "Dados inválidos para a operação",
		}
	default:
		return &DeliveryVipError{
			HTTPStatus: httpStatus,
			TipoErro:   models.ErroBadGateway,
			Mensagem:   fmt.Sprintf("Erro na comunicação com a plataforma - Status: %d, Resposta: %s", httpStatus, responseBody),
		}
	}
}

// DeliveryVipService gerencia a integração com DeliveryVip
type DeliveryVipService struct {
	config      *config.Config
	accessToken string
	tokenMutex  sync.RWMutex
	httpClient  *http.Client
}

// DeliveryVipTokenRequest representa o payload de autenticação OAuth
type DeliveryVipTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// DeliveryVipTokenResponse representa a resposta de autenticação OAuth
type DeliveryVipTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// DeliveryVipMerchant representa um merchant na resposta da API
type DeliveryVipMerchant struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Subscription struct {
		Status  string `json:"status"`
		Blocked bool   `json:"blocked"`
	} `json:"subscription"`
}

// DeliveryVipBlockResponse representa a resposta de block/unblock
type DeliveryVipBlockResponse struct {
	MerchantID string `json:"merchantId"`
}

// NewDeliveryVipService cria um novo serviço DeliveryVip
func NewDeliveryVipService(cfg *config.Config) *DeliveryVipService {
	service := &DeliveryVipService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Inicia a rotina de renovação de token
	go service.startTokenRenewal()

	return service
}

// startTokenRenewal inicia a rotina que renova o token a cada 20 horas (token expira em 24h)
func (s *DeliveryVipService) startTokenRenewal() {
	log.Printf("[DeliveryVip] Iniciando serviço de renovação de token...")
	log.Printf("[DeliveryVip] URL configurada: %s", s.config.Platforms.DeliveryVipURL)

	// Faz o primeiro login imediatamente
	log.Printf("[DeliveryVip] Tentando autenticação inicial...")
	if err := s.renewToken(); err != nil {
		log.Printf("[DeliveryVip] ERRO na autenticação inicial: %v", err)
	} else {
		log.Printf("[DeliveryVip] Autenticação inicial realizada com sucesso!")
	}

	// Configura renovação a cada 20 horas (token expira em 24h)
	ticker := time.NewTicker(20 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("[DeliveryVip] Renovando token automaticamente...")
		if err := s.renewToken(); err != nil {
			log.Printf("[DeliveryVip] ERRO na renovação automática: %v", err)
		} else {
			log.Printf("[DeliveryVip] Token renovado com sucesso!")
		}
	}
}

// renewToken faz o login OAuth e atualiza o token de acesso
func (s *DeliveryVipService) renewToken() error {
	tokenURL := fmt.Sprintf("%s/authentication/v1/oauth/token", s.config.Platforms.DeliveryVipURL)

	// Prepara os dados do formulário
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.config.Platforms.DeliveryVip.ClientID)
	data.Set("client_secret", s.config.Platforms.DeliveryVip.ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de token: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "*/*")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler resposta do token: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro de autenticação OAuth - Status: %d, Resposta: %s", resp.StatusCode, string(body))
	}

	var tokenResp DeliveryVipTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("erro ao decodificar resposta do token: %w", err)
	}

	// Atualiza o token de forma thread-safe
	s.tokenMutex.Lock()
	s.accessToken = tokenResp.AccessToken
	s.tokenMutex.Unlock()

	log.Printf("[DeliveryVip] Token obtido com sucesso! Expira em: %d segundos", tokenResp.ExpiresIn)
	return nil
}

// getAccessToken retorna o token de acesso atual de forma thread-safe
func (s *DeliveryVipService) getAccessToken() string {
	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()
	return s.accessToken
}

// ActivateStore desbloqueia uma loja no DeliveryVip
func (s *DeliveryVipService) ActivateStore(merchantID string) error {
	token := s.getAccessToken()
	if token == "" {
		return fmt.Errorf("token de acesso não disponível")
	}

	unblockURL := fmt.Sprintf("%s/partner/v2/merchants/%s/unblock", s.config.Platforms.DeliveryVipURL, merchantID)

	req, err := http.NewRequest("POST", unblockURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de desbloqueio: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "*/*")

	log.Printf("[DeliveryVip] Desbloqueando loja: %s", merchantID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de desbloqueio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return NewDeliveryVipError(resp.StatusCode, string(body))
	}

	log.Printf("[DeliveryVip] Loja %s desbloqueada com sucesso", merchantID)
	return nil
}

// DeactivateStore bloqueia uma loja no DeliveryVip
func (s *DeliveryVipService) DeactivateStore(merchantID string) error {
	token := s.getAccessToken()
	if token == "" {
		return fmt.Errorf("token de acesso não disponível")
	}

	blockURL := fmt.Sprintf("%s/partner/v2/merchants/%s/block", s.config.Platforms.DeliveryVipURL, merchantID)

	req, err := http.NewRequest("POST", blockURL, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de bloqueio: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "*/*")

	log.Printf("[DeliveryVip] Bloqueando loja: %s", merchantID)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de bloqueio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return NewDeliveryVipError(resp.StatusCode, string(body))
	}

	log.Printf("[DeliveryVip] Loja %s bloqueada com sucesso", merchantID)
	return nil
}

// StoreStatusResult representa o resultado do status de uma loja
type StoreStatusResult struct {
	Found    bool
	IsActive bool
}

// GetMultipleStoreStatus consulta o status de múltiplas lojas no DeliveryVip
func (s *DeliveryVipService) GetMultipleStoreStatus(merchantIDs []string) (map[string]StoreStatusResult, error) {
	token := s.getAccessToken()
	if token == "" {
		return nil, fmt.Errorf("token de acesso não disponível")
	}

	merchantsURL := fmt.Sprintf("%s/partner/v2/merchants", s.config.Platforms.DeliveryVipURL)

	req, err := http.NewRequest("GET", merchantsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de consulta: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	log.Printf("[DeliveryVip] Consultando todas as lojas para filtrar %d IDs solicitados", len(merchantIDs))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição de consulta: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta de consulta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao consultar merchants - Status: %d, Resposta: %s", resp.StatusCode, string(body))
	}

	var merchants []DeliveryVipMerchant
	if err := json.Unmarshal(body, &merchants); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta de merchants: %w", err)
	}

	// Cria um set dos IDs solicitados para busca rápida
	requestedIDs := make(map[string]bool)
	for _, id := range merchantIDs {
		requestedIDs[id] = true
	}

	// Filtra apenas as lojas solicitadas e constrói mapa de status
	statusMap := make(map[string]StoreStatusResult)

	for _, merchant := range merchants {
		// Só processa se o ID foi solicitado
		if requestedIDs[merchant.ID] {
			// Status ativo = subscription.status == "ACTIVATED" && !subscription.blocked
			isActive := merchant.Subscription.Status == "ACTIVATED" && !merchant.Subscription.Blocked
			statusMap[merchant.ID] = StoreStatusResult{
				Found:    true,
				IsActive: isActive,
			}
		}
	}

	// Adiciona lojas não encontradas
	for _, id := range merchantIDs {
		if _, exists := statusMap[id]; !exists {
			statusMap[id] = StoreStatusResult{
				Found:    false,
				IsActive: false,
			}
		}
	}

	log.Printf("[DeliveryVip] Status consultado: %d/%d lojas encontradas", len(statusMap)-countNotFound(statusMap), len(merchantIDs))
	return statusMap, nil
}

// countNotFound conta quantas lojas não foram encontradas
func countNotFound(statusMap map[string]StoreStatusResult) int {
	count := 0
	for _, result := range statusMap {
		if !result.Found {
			count++
		}
	}
	return count
}
