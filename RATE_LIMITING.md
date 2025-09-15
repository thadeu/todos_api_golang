# Rate Limiting, Cache de Resposta e HTTPS Enforcement

Este documento descreve as funcionalidades de rate limiting, cache de resposta e HTTPS enforcement implementadas na aplicação.

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

## Cache de Resposta

### Configuração

O cache de resposta é implementado usando cache local (go-cache) e armazena respostas de rotas GET:

- **Todos**: Cache por 30 segundos
- **Default**: Cache desabilitado por padrão

### Headers de Resposta

O middleware adiciona os seguintes headers informativos:

- `X-Cache`: Status do cache (`HIT` ou `MISS`)
- `X-Cache-Age`: Idade do cache em segundos (apenas para HIT)

### Chave de Cache

A chave do cache é gerada considerando:
- Path do endpoint
- Query parameters
- User ID (se autenticado) ou IP (se não autenticado)

### Métricas

As seguintes métricas são coletadas:

- `cache_hits_total`: Total de cache hits
- `cache_misses_total`: Total de cache misses

Labels:
- `path`: Endpoint que foi acessado

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

### Cache de Resposta

Execute o script de teste:

```bash
./scripts/test-cache.sh
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

As métricas aparecem automaticamente nos dashboards do Grafana:

**Rate Limiting:**
- **Rate Limit Hits**: Gráfico de requests bloqueados
- **Rate Limit Allowed**: Gráfico de requests permitidos
- **Rate Limit by Endpoint**: Distribuição por endpoint
- **Rate Limit by Key Type**: Distribuição por tipo de chave

**Cache:**
- **Cache Hits**: Gráfico de cache hits
- **Cache Misses**: Gráfico de cache misses
- **Cache Hit Ratio**: Taxa de acerto do cache
- **Cache by Endpoint**: Distribuição por endpoint

### Logs

**Rate Limiting** gera logs estruturados:

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

**Cache** gera logs estruturados:

```json
{
  "level": "debug",
  "msg": "Cache hit",
  "path": "/todos",
  "cache_key": "cache:/todos:abc123...",
  "age": "15s",
  "trace_id": "...",
  "span_id": "..."
}
```

## Performance

### Cache Local

**Rate Limiting:**
- **Limpeza Automática**: A cada 5 minutos
- **Expiração**: 10 minutos para entradas não utilizadas
- **Memória**: Baixo uso de memória, apenas contadores
- **Latência**: Impacto mínimo (< 1ms por request)

**Response Cache:**
- **TTL Configurável**: Por endpoint (padrão: 30s para /todos)
- **Limpeza Automática**: A cada 5 minutos
- **Memória**: Armazena respostas completas (JSON)
- **Latência**: Cache hit < 1ms, cache miss = tempo normal da API

### Escalabilidade

Para ambientes com múltiplas instâncias, considere:

1. **Redis**: Substituir go-cache por Redis para rate limiting e cache distribuídos
2. **Load Balancer**: Configurar rate limiting no nível do load balancer
3. **CDN**: Usar CDN com rate limiting e cache (CloudFlare, AWS CloudFront)
4. **Cache Invalidation**: Implementar invalidação de cache em operações de escrita

## Próximos Passos

1. **Redis Integration**: Implementar rate limiting e cache distribuídos
2. **Dynamic Configuration**: Permitir mudança de limites sem restart
3. **Cache Invalidation**: Invalidação automática em operações de escrita
4. **Whitelist/Blacklist**: Listas de IPs permitidos/bloqueados
5. **Burst Allowance**: Permitir picos temporários de tráfego
6. **Rate Limit by User Tier**: Diferentes limites por tipo de usuário
7. **Cache Warming**: Pré-carregar cache com dados frequentes
