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
DELIVERYVIP_API_URL=https://api.deliveryvip.com
```

## Endpoints

### Health Check
- **GET** `/health` - Verificação de saúde (sem autenticação)

### Operações de Loja (requer autenticação)
- **POST** `/plataformas/{plataforma}/lojas/{id_loja}/ativar` - Ativar loja
- **POST** `/plataformas/{plataforma}/lojas/{id_loja}/desativar` - Desativar loja  
- **GET** `/plataformas/{plataforma}/lojas/status?ids=id1,id2,id3` - Consultar status de múltiplas lojas

### Parâmetros
- `plataforma`: `anotaai` ou `deliveryvip`
- `id_loja`: Identificador da loja na plataforma

### Autenticação
Todas as operações de loja requerem autenticação via Bearer token:

```bash
Authorization: Bearer <seu-token>
```

## Respostas da API

Todas as respostas seguem a especificação OpenAPI definida em `docs/openapi.yml`.