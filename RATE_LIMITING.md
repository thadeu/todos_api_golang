# Rate Limiting e HTTPS Enforcement

Este documento descreve as funcionalidades de rate limiting e HTTPS enforcement implementadas na aplicação.

## Rate Limiting

### Configuração

O rate limiting é implementado usando cache local (go-cache) e pode ser configurado por endpoint:

- **Signup**: 5 requests por minuto por IP
- **Auth**: 10 requests por minuto por IP  
- **Todos**: 100 requests por minuto por usuário autenticado
- **Default**: 60 requests por minuto por IP para outros endpoints

### Headers de Resposta

O middleware adiciona os seguintes headers informativos:

- `X-RateLimit-Limit`: Número máximo de requests permitidos
- `X-RateLimit-Remaining`: Número de requests restantes na janela atual
- `X-RateLimit-Reset`: Timestamp Unix de quando a janela será resetada

### Resposta de Rate Limit Excedido

Quando o limite é excedido, a aplicação retorna:

```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests. Limit: 5 per 1m0s",
  "retry_after": 45
}
```

Status HTTP: `429 Too Many Requests`

### Métricas

As seguintes métricas são coletadas:

- `rate_limit_hits_total`: Total de requests bloqueados por rate limiting
- `rate_limit_allowed_total`: Total de requests permitidos pelo rate limiter

Labels:
- `path`: Endpoint que foi acessado
- `key_type`: Tipo de chave usada (`ip` ou `user`)

## HTTPS Enforcement

### Configuração Automática

O HTTPS enforcement é habilitado automaticamente quando:

- `GIN_MODE=release` (modo de produção)
- `ENFORCE_HTTPS=true` (forçar via variável de ambiente)

### Comportamento

- **Desenvolvimento**: HTTPS enforcement desabilitado
- **Produção**: Redirecionamento automático para HTTPS
- **Localhost**: Sempre permitido (desenvolvimento local)
- **Headers de Proxy**: Suporte para `X-Forwarded-Proto: https`

### Redirecionamento

Quando uma requisição HTTP é feita em produção, a aplicação:

1. Verifica se já é HTTPS
2. Verifica headers de proxy
3. Verifica se é localhost
4. Redireciona para HTTPS com status `301 Moved Permanently`

## Configuração

### Variáveis de Ambiente

```bash
# Habilitar HTTPS enforcement
ENFORCE_HTTPS=true

# Modo de produção (habilita HTTPS automaticamente)
GIN_MODE=release

# Desabilitar rate limiting
RATE_LIMIT_ENABLED=false
```

### Configuração Programática

```go
config := &AppConfig{
    RateLimitEnabled: true,
    EnforceHTTPS:     true,
    Environment:      "production",
    RateLimitConfigs: map[string]RateLimitConfig{
        "/custom": {
            Requests: 20,
            Window:   time.Minute,
        },
    },
}
```

## Testando

### Rate Limiting

Execute o script de teste:

```bash
./scripts/test-rate-limiting.sh
```

### HTTPS Enforcement

Para testar em produção:

```bash
# Definir modo de produção
export GIN_MODE=release

# Executar aplicação
go run cmd/api/main.go

# Testar redirecionamento
curl -v http://localhost:8080/signup
# Deve retornar 301 redirect para https://
```

## Monitoramento

### Grafana Dashboards

As métricas de rate limiting aparecem automaticamente nos dashboards do Grafana:

- **Rate Limit Hits**: Gráfico de requests bloqueados
- **Rate Limit Allowed**: Gráfico de requests permitidos
- **Rate Limit by Endpoint**: Distribuição por endpoint
- **Rate Limit by Key Type**: Distribuição por tipo de chave

### Logs

Rate limiting gera logs estruturados:

```json
{
  "level": "warn",
  "msg": "Rate limit exceeded",
  "key": "rate_limit:/signup:192.168.1.1",
  "path": "/signup",
  "limit": 5,
  "window": "1m0s",
  "trace_id": "...",
  "span_id": "..."
}
```

## Performance

### Cache Local

- **Limpeza Automática**: A cada 5 minutos
- **Expiração**: 10 minutos para entradas não utilizadas
- **Memória**: Baixo uso de memória, apenas contadores
- **Latência**: Impacto mínimo (< 1ms por request)

### Escalabilidade

Para ambientes com múltiplas instâncias, considere:

1. **Redis**: Substituir go-cache por Redis para rate limiting distribuído
2. **Load Balancer**: Configurar rate limiting no nível do load balancer
3. **CDN**: Usar CDN com rate limiting (CloudFlare, AWS CloudFront)

## Próximos Passos

1. **Redis Integration**: Implementar rate limiting distribuído
2. **Dynamic Configuration**: Permitir mudança de limites sem restart
3. **Whitelist/Blacklist**: Listas de IPs permitidos/bloqueados
4. **Burst Allowance**: Permitir picos temporários de tráfego
5. **Rate Limit by User Tier**: Diferentes limites por tipo de usuário
