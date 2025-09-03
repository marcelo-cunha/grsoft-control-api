package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"delivery-control/internal/config"
	"delivery-control/internal/models"
	"delivery-control/internal/utils"
)

// AnotaAiService gerencia a integração com AnotaAI
type AnotaAiService struct {
	config      *config.Config
	accessToken string
	tokenMutex  sync.RWMutex
	httpClient  *http.Client
}

// LoginRequest representa o payload de login do AnotaAI
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse representa a resposta de login do AnotaAI
type LoginResponse struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token"`
}

// AnotaAiResponse representa uma resposta padrão do AnotaAI
type AnotaAiResponse struct {
	Success  bool   `json:"success"`
	Mensagem string `json:"mensagem"`
}

// AnotaAiListPagesResponse representa a resposta da API de listagem de páginas
type AnotaAiListPagesResponse struct {
	Success bool `json:"success"`
	Info    struct {
		Docs  []AnotaAiPage `json:"docs"`
		Limit int           `json:"limit"`
		Page  int           `json:"page"`
	} `json:"info"`
}

// CpfCnpjField representa o campo cpf_cnpj que pode ser string ou objeto
type CpfCnpjField struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
	Raw   string // para armazenar valor quando for string
}

// UnmarshalJSON implementa unmarshaling customizado para lidar com string ou objeto
func (c *CpfCnpjField) UnmarshalJSON(data []byte) error {
	// Primeiro tenta como objeto
	var obj struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		c.Type = obj.Type
		c.Value = obj.Value
		c.Raw = obj.Value
		return nil
	}

	// Se falhar, tenta como string
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.Raw = str
		c.Value = str
		return nil
	}

	return fmt.Errorf("cpf_cnpj deve ser string ou objeto")
}

// GetValue retorna o valor do documento independente do formato
func (c *CpfCnpjField) GetValue() string {
	if c.Value != "" {
		return c.Value
	}
	return c.Raw
}

// AnotaAiPage representa uma página/loja na resposta da API
type AnotaAiPage struct {
	ID       string `json:"_id"`
	PageID   string `json:"page_id"`
	PageName string `json:"page_name"`
	Active   bool   `json:"active"`
	Page     struct {
		Establishment struct {
			Sign struct {
				Active  bool         `json:"active"`
				CpfCnpj CpfCnpjField `json:"cpf_cnpj"`
			} `json:"sign"`
		} `json:"establishment"`
	} `json:"page"`
}

// NewAnotaAiService cria um novo serviço AnotaAI
func NewAnotaAiService(cfg *config.Config) *AnotaAiService {
	service := &AnotaAiService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Inicia a rotina de renovação de token
	go service.startTokenRenewal()

	return service
}

// startTokenRenewal inicia a rotina que renova o token a cada 3 horas
func (s *AnotaAiService) startTokenRenewal() {
	log.Printf("[AnotaAI] Iniciando serviço de renovação de token...")
	log.Printf("[AnotaAI] URL configurada: %s", s.config.Platforms.AnotaAiURL)

	// Faz o primeiro login imediatamente
	log.Printf("[AnotaAI] Tentando login inicial...")
	if err := s.renewToken(); err != nil {
		log.Printf("[AnotaAI] ERRO no login inicial: %v", err)
	} else {
		log.Printf("[AnotaAI] Login inicial realizado com sucesso!")
	}

	// Configura renovação a cada 3 horas
	ticker := time.NewTicker(3 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("[AnotaAI] Renovando token automaticamente...")
		if err := s.renewToken(); err != nil {
			log.Printf("[AnotaAI] ERRO ao renovar token: %v", err)
		}
	}
}

// renewToken renova o token de acesso
func (s *AnotaAiService) renewToken() error {
	log.Printf("[AnotaAI] Iniciando processo de renovação de token...")

	// Verifica se as credenciais estão configuradas
	if s.config.Platforms.AnotaAi.Email == "" {
		return fmt.Errorf("email do AnotaAI não configurado")
	}
	if s.config.Platforms.AnotaAi.Password == "" {
		return fmt.Errorf("senha do AnotaAI não configurada")
	}

	loginReq := LoginRequest{
		Email:    s.config.Platforms.AnotaAi.Email,
		Password: s.config.Platforms.AnotaAi.Password,
	}

	payload, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload de login: %w", err)
	}

	url := fmt.Sprintf("%s/noauth/partner/login", s.config.Platforms.AnotaAiURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de login: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Printf("[AnotaAI] Enviando requisição de login...")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição de login: %w", err)
	}
	defer resp.Body.Close()

	// Lê o corpo da resposta para debug
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("erro ao ler corpo da resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[AnotaAI] Corpo da resposta (erro): %s", string(body))
		return fmt.Errorf("erro no login - status: %d, resposta: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("erro ao decodificar resposta de login: %w", err)
	}

	if loginResp.AccessToken != "" {
		log.Printf("[AnotaAI] Token recebido com sucesso")
	}

	if !loginResp.Success {
		return fmt.Errorf("login falhou - success: false")
	}

	// Atualiza o token de forma thread-safe
	s.tokenMutex.Lock()
	s.accessToken = loginResp.AccessToken
	s.tokenMutex.Unlock()

	log.Printf("[AnotaAI] Token AnotaAI renovado com sucesso às %s", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// getAccessToken retorna o token de acesso atual de forma thread-safe
func (s *AnotaAiService) getAccessToken() string {
	s.tokenMutex.RLock()
	defer s.tokenMutex.RUnlock()
	return s.accessToken
}

// ActivateStore ativa uma loja no AnotaAI
func (s *AnotaAiService) ActivateStore(idLoja string) error {
	token := s.getAccessToken()
	if token == "" {
		return fmt.Errorf("token de acesso não disponível")
	}

	url := fmt.Sprintf("%s/partnerauth/partner/active/%s", s.config.Platforms.AnotaAiURL, idLoja)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de ativação: %w", err)
	}

	req.Header.Set("authorization", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição de ativação: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro na ativação - status: %d", resp.StatusCode)
	}

	var anotaResp AnotaAiResponse
	if err := json.NewDecoder(resp.Body).Decode(&anotaResp); err != nil {
		return fmt.Errorf("erro ao decodificar resposta de ativação: %w", err)
	}

	if !anotaResp.Success {
		return fmt.Errorf("ativação falhou: %s", anotaResp.Mensagem)
	}

	return nil
}

// DeactivateStore desativa uma loja no AnotaAI
func (s *AnotaAiService) DeactivateStore(idLoja string) error {
	token := s.getAccessToken()
	if token == "" {
		return fmt.Errorf("token de acesso não disponível")
	}

	url := fmt.Sprintf("%s/partnerauth/partner/block/%s", s.config.Platforms.AnotaAiURL, idLoja)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição de desativação: %w", err)
	}

	req.Header.Set("authorization", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro na requisição de desativação: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro na desativação - status: %d", resp.StatusCode)
	}

	var anotaResp AnotaAiResponse
	if err := json.NewDecoder(resp.Body).Decode(&anotaResp); err != nil {
		return fmt.Errorf("erro ao decodificar resposta de desativação: %w", err)
	}

	if !anotaResp.Success {
		return fmt.Errorf("desativação falhou: %s", anotaResp.Mensagem)
	}

	return nil
}

// GetMultipleStoreStatus obtém o status de múltiplas lojas no AnotaAI
// Se idsLojas for nil ou vazio, retorna o status de todas as lojas
func (s *AnotaAiService) GetMultipleStoreStatus(idsLojas []string) (map[string]models.StoreInfo, error) {
	token := s.getAccessToken()
	if token == "" {
		return nil, fmt.Errorf("token de acesso não disponível")
	}

	url := fmt.Sprintf("%s/partnerauth/partner/listpages/v2?limit=2000&page=1", s.config.Platforms.AnotaAiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição de status: %w", err)
	}

	req.Header.Set("authorization", token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição de status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro na consulta de status - status: %d", resp.StatusCode)
	}

	var listResp AnotaAiListPagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta de status: %w", err)
	}

	if !listResp.Success {
		return nil, fmt.Errorf("consulta de status falhou - success: false")
	}

	// Mapa para armazenar as informações das lojas
	storeMap := make(map[string]models.StoreInfo)

	// Se nenhum ID específico foi solicitado, retorna todas as lojas
	if len(idsLojas) == 0 {
		for _, page := range listResp.Info.Docs {
			// No AnotaAI só existe ativo e bloqueado
			var status models.Status
			if page.Page.Establishment.Sign.Active {
				status = models.StatusAtivo
			} else {
				status = models.StatusBloqueado
			}

			storeMap[page.PageID] = models.StoreInfo{
				Found:        true,
				IsActive:     page.Page.Establishment.Sign.Active,
				Status:       status,
				Documento:    utils.CleanDocument(page.Page.Establishment.Sign.CpfCnpj.GetValue()),
				NomeFantasia: page.PageName,
			}
		}
		return storeMap, nil
	}

	// Inicializa todas as lojas solicitadas como não encontradas
	for _, idLoja := range idsLojas {
		storeMap[idLoja] = models.StoreInfo{
			Found:        false,
			IsActive:     false,
			Status:       models.StatusNaoEncontrado,
			Documento:    "",
			NomeFantasia: "",
		}
	}

	// Procura cada loja solicitada na resposta da API
	for _, idLoja := range idsLojas {
		for _, page := range listResp.Info.Docs {
			if page.PageID == idLoja {
				// No AnotaAI só existe ativo e bloqueado
				var status models.Status
				if page.Page.Establishment.Sign.Active {
					status = models.StatusAtivo
				} else {
					status = models.StatusBloqueado
				}

				storeMap[idLoja] = models.StoreInfo{
					Found:        true,
					IsActive:     page.Page.Establishment.Sign.Active,
					Status:       status,
					Documento:    utils.CleanDocument(page.Page.Establishment.Sign.CpfCnpj.GetValue()),
					NomeFantasia: page.PageName,
				}
				break
			}
		}
	}

	return storeMap, nil
}
