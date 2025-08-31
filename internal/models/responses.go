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
