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
	Success bool   `json:"success"`
	Message string `json:"message"`
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

// startTokenRenewal inicia a rotina que renova o token a cada 6 horas
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

	// Configura renovação a cada 6 horas
	ticker := time.NewTicker(6 * time.Hour)
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
		return fmt.Errorf("ativação falhou: %s", anotaResp.Message)
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
		return fmt.Errorf("desativação falhou: %s", anotaResp.Message)
	}

	return nil
}
