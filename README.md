# Control API

API wrapper para ativação e desativação de lojas em plataformas de terceiros.

## Configuração

Configure as variáveis de ambiente no arquivo `.env`:

```env
BEARER_TOKEN=meu-token-secreto-123
PORT=8080
ANOTAAI_API_URL=https://integration-admin.api.anota.ai
ANOTAAI_EMAIL=seu-email@exemplo.com.br
ANOTAAI_PASSWORD=sua-senha
DELIVERYVIP_API_URL=https://api.deliveryvip.com.br
DELIVERYVIP_CLIENT_ID=1a2b3c4d-2dcb-3c4d-1a2b-3c9f6e5e8a1b
DELIVERYVIP_CLIENT_SECRET=example
```

## Endpoints

### Health Check
- **GET** `/health` - Verificação de saúde (sem autenticação)

### Operações de Loja (requer autenticação)
- **PATCH** `/plataformas/{plataforma}/lojas/ativar` - Ativar múltiplas lojas
- **PATCH** `/plataformas/{plataforma}/lojas/desativar` - Desativar múltiplas lojas  
- **GET** `/plataformas/{plataforma}/lojas/status` - Consultar status de múltiplas lojas (IDs no header X-Lojas-IDs)
  - Retorna: id_loja, status, documento (CPF/CNPJ) e nome_fantasia para cada loja

### Parâmetros
- `plataforma`: `anotaai` ou `deliveryvip`
- Para ativar/desativar: IDs das lojas no body da requisição no formato `{"ids_lojas": ["id1", "id2", "id3"]}`
- Para consultar status: IDs das lojas no header `X-Lojas-IDs` (opcional - se vazio, retorna todas as lojas)

### Autenticação
Todas as operações de loja requerem autenticação via Bearer token:

```bash
Authorization: Bearer <seu-token>
```

## Respostas da API

Todas as respostas seguem a especificação OpenAPI definida em `docs/openapi.yml`.