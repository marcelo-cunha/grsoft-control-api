package models

// Plataforma representa as plataformas suportadas
type Plataforma string

const (
	PlataformaAnotaAi     Plataforma = "anotaai"
	PlataformaDeliveryVip Plataforma = "deliveryvip"
)

// Status representa o status de uma loja
type Status string

const (
	StatusAtivo         Status = "ativo"
	StatusInativo       Status = "inativo"
	StatusNaoEncontrado Status = "nao_encontrado"
)

// TipoErro representa os tipos de erro da API
type TipoErro string

const (
	ErroRequisicaoInvalida TipoErro = "invalid_request"
	ErroNaoAutorizado      TipoErro = "unauthorized"
	ErroNaoEncontrado      TipoErro = "not_found"
	ErroBadGateway         TipoErro = "bad_gateway"
	ErroInternoServidor    TipoErro = "internal_server_error"
)

// RespostaOperacaoLoja representa a resposta para operações de ativação/desativação
type RespostaOperacaoLoja struct {
	Plataforma Plataforma `json:"plataforma"`
	IdLoja     string     `json:"id_loja"`
	Status     Status     `json:"status"`
	Mensagem   string     `json:"mensagem"`
}

// RequisicaoMultiplasLojas representa a requisição para operações com múltiplas lojas
type RequisicaoMultiplasLojas struct {
	IdsLojas []string `json:"ids_lojas" validate:"required,min=1"`
}

// RespostaOperacaoMultiplasLojas representa a resposta para operações de ativação/desativação de múltiplas lojas
type RespostaOperacaoMultiplasLojas struct {
	Plataforma Plataforma              `json:"plataforma"`
	Resultados []ResultadoOperacaoLoja `json:"resultados"`
}

// ResultadoOperacaoLoja representa o resultado individual de uma operação
type ResultadoOperacaoLoja struct {
	IdLoja   string    `json:"id_loja"`
	Status   Status    `json:"status"`
	Sucesso  bool      `json:"sucesso"`
	Mensagem string    `json:"mensagem"`
	Erro     *TipoErro `json:"erro,omitempty"`
}

// RespostaStatusMultiplasLojas representa a resposta para consulta de status de múltiplas lojas
type RespostaStatusMultiplasLojas struct {
	Plataforma Plataforma           `json:"plataforma"`
	Lojas      []StatusLojaDetalhes `json:"lojas"`
}

// StatusLojaDetalhes representa os detalhes de status de uma loja específica
type StatusLojaDetalhes struct {
	IdLoja string `json:"id_loja"`
	Status Status `json:"status"`
}

// RespostaErro representa uma resposta de erro
type RespostaErro struct {
	Error    TipoErro `json:"error"`
	Mensagem string   `json:"mensagem"`
}

// RespostaSaude representa a resposta do health check
type RespostaSaude struct {
	Status string `json:"status"`
}
