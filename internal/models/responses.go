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
	StatusAtivo   Status = "ativo"
	StatusInativo Status = "inativo"
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
	Message    string     `json:"message"`
}

// RespostaStatusLoja representa a resposta para consulta de status
type RespostaStatusLoja struct {
	Plataforma Plataforma `json:"plataforma"`
	IdLoja     string     `json:"id_loja"`
	Status     Status     `json:"status"`
}

// RespostaErro representa uma resposta de erro
type RespostaErro struct {
	Error   TipoErro `json:"error"`
	Message string   `json:"message"`
}

// RespostaSaude representa a resposta do health check
type RespostaSaude struct {
	Status string `json:"status"`
}
