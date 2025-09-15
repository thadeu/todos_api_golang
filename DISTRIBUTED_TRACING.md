# Distributed Tracing - Melhorias Implementadas

Este documento descreve as melhorias implementadas no distributed tracing da aplicação TodoApp.

## O que é Distributed Tracing?

**Distributed Tracing** é uma técnica de observabilidade que permite rastrear uma requisição através de múltiplos serviços, sistemas e componentes de uma aplicação distribuída.

### Conceitos Básicos

1. **Trace**: Um rastreamento completo de uma operação que pode atravessar múltiplos serviços
2. **Span**: Uma operação individual dentro de um trace (ex: uma chamada HTTP, query no banco)
3. **Context Propagation**: Como o contexto do trace é passado entre serviços

### Exemplo Prático

```
User Request → API Gateway → Auth Service → Todo Service → Database
     ↓              ↓            ↓             ↓           ↓
   Span 1        Span 2       Span 3        Span 4     Span 5
   
Todos os spans fazem parte do mesmo TRACE
```

## Melhorias Implementadas

### 1. **TracerHelper - Helper para Traces**

Criado um helper centralizado para facilitar a criação e gerenciamento de spans:

```go
// Criar span com atributos
ctx, span := CreateChildSpan(ctx, "service.todo.GetTodos", []attribute.KeyValue{
    attribute.Int("user.id", userId),
    attribute.String("operation", "GetTodos"),
})

// Marcar erro no span
AddSpanError(span, err)

// Adicionar atributos de sucesso
span.SetAttributes(attribute.Int("todo.count", len(data)))
```

### 2. **Instrumentação Completa de Camadas**

Agora cada camada tem seu próprio span com contexto propagado:

#### Handler Layer
```go
func (t *TodoHandler) GetAllTodos(c *gin.Context) {
    ctx, span := CreateChildSpan(c.Request.Context(), "handler.todo.GetAllTodos", ...)
    defer span.End()
    
    // Processamento...
    data, err := t.Service.GetTodosWithPagination(ctx, userId, limit, cursor)
}
```

#### Service Layer
```go
func (s *TodoService) GetTodosWithPagination(ctx context.Context, ...) {
    ctx, span := CreateChildSpan(ctx, "service.todo.GetTodosWithPagination", ...)
    defer span.End()
    
    // Processamento...
    rows, hasNext, err := s.repo.GetAllWithCursor(ctx, userId, limit, cursor)
}
```

#### Repository Layer
```go
func (r *TodoRepository) GetAllWithCursor(ctx context.Context, ...) {
    ctx, span := CreateChildSpan(ctx, "db.todo.GetAllWithCursor", ...)
    defer span.End()
    
    // Query no banco...
}
```

### 3. **Atributos Estruturados**

Cada span agora inclui atributos relevantes:

#### Handler Attributes
- `handler.operation`: Nome da operação
- `handler.method`: Método HTTP
- `handler.path`: Path da rota
- `user.id`: ID do usuário
- `http.status_code`: Status da resposta

#### Service Attributes
- `service.name`: Nome do serviço
- `service.operation`: Operação do serviço
- `user.id`: ID do usuário
- `todo.count`: Número de todos retornados
- `todo.has_next`: Se tem próxima página

#### Database Attributes
- `db.table`: Tabela acessada
- `db.operation`: Operação (SELECT, INSERT, etc.)
- `db.query`: Query executada
- `db.rows_returned`: Número de linhas retornadas
- `db.has_next`: Se tem próxima página

### 4. **Error Handling Melhorado**

Erros são agora marcados nos spans com contexto:

```go
if err != nil {
    AddSpanError(span, err)  // Marca span como erro
    // Log com trace_id automaticamente
    t.Logger.Logger.Ctx(ctx).Error("Failed to get todos", zap.Error(err))
    return
}
```

### 5. **Context Propagation Consistente**

O contexto é propagado consistentemente através de todas as camadas:

```
Request → Handler → Service → Repository → Database
   ↓         ↓         ↓          ↓          ↓
Context propagado com trace_id e span_id
```

## Visualização no Grafana/Tempo

### Hierarquia de Spans

**Cache Miss (primeira requisição):**
```
GET /todos
├── cache.response.miss
├── handler.todo.GetAllTodos
│   ├── service.todo.GetTodosWithPagination
│   │   └── db.todo.GetAllWithCursor
│   └── cache.response.store
```

**Cache Hit (requisições subsequentes):**
```
GET /todos
├── cache.response.hit
└── (não executa handler/service/database)
```

### Atributos por Span

**Root Span (Gin Middleware):**
- `http.method`: GET
- `http.url`: /todos
- `http.status_code`: 200

**Handler Span:**
- `handler.operation`: GetAllTodos
- `user.id`: 123
- `todo.limit`: 10
- `todo.cursor`: "abc123"

**Service Span:**
- `service.operation`: GetTodosWithPagination
- `todo.count`: 5
- `todo.has_next`: true

**Database Span:**
- `db.table`: todos
- `db.operation`: SELECT
- `db.rows_returned`: 5
- `db.has_next`: true

**Cache Hit Span:**
- `cache.key`: cache:/todos:abc123...
- `cache.path`: /todos
- `cache.age`: 15s
- `cache.source`: memory
- `cache.status_code`: 200
- `cache.body_size`: 1024
- `cache.ttl`: 3s

**Cache Miss Span:**
- `cache.key`: cache:/todos:abc123...
- `cache.path`: /todos
- `cache.source`: memory

**Cache Store Span:**
- `cache.key`: cache:/todos:abc123...
- `cache.path`: /todos
- `cache.source`: memory
- `cache.status_code`: 200
- `cache.body_size`: 1024
- `cache.ttl`: 3s

## Benefícios das Melhorias

### 1. **Visibilidade Completa**
- Rastreamento de requisições através de todas as camadas
- Identificação de gargalos por camada
- Correlação de logs com traces

### 2. **Debugging Melhorado**
- Erros marcados nos spans com contexto
- Logs automaticamente correlacionados com trace_id
- Atributos de negócio para análise

### 3. **Performance Monitoring**
- Tempo gasto em cada camada
- Identificação de queries lentas
- Análise de padrões de uso

### 4. **Operational Insights**
- Distribuição de operações por usuário
- Padrões de acesso aos dados
- Métricas de negócio nos traces

## Como Usar

### 1. **Visualizar Traces**

Acesse o Grafana em `http://localhost:3000` e navegue para:
- **Explore** → **Tempo** → Query: `{service_name="todoapp"}`

### 2. **Filtrar por Operação**

```
{service_name="todoapp" && name="handler.todo.GetAllTodos"}
{service_name="todoapp" && name="db.todo.GetAllWithCursor"}
```

### 3. **Analisar Performance**

- Clique em um trace para ver a hierarquia de spans
- Analise o tempo gasto em cada camada
- Identifique spans com erro (vermelho)

### 4. **Correlacionar com Logs**

- Use o `trace_id` dos spans para encontrar logs relacionados
- Logs automaticamente incluem `trace_id` e `span_id`

## Exemplo de Trace Completo

**Cache Miss (primeira requisição):**
```
Trace ID: 1234567890abcdef
├── Span 1: GET /todos (200ms)
│   ├── Span 2: cache.response.miss (1ms)
│   ├── Span 3: handler.todo.GetAllTodos (180ms)
│   │   ├── Span 4: service.todo.GetTodosWithPagination (150ms)
│   │   │   └── Span 5: db.todo.GetAllWithCursor (120ms)
│   │   └── Span 6: response formatting (30ms)
│   ├── Span 7: cache.response.store (1ms)
│   └── Span 8: middleware processing (20ms)
```

**Cache Hit (requisições subsequentes):**
```
Trace ID: 1234567890abcdef
├── Span 1: GET /todos (5ms)
│   ├── Span 2: cache.response.hit (1ms)
│   └── Span 3: middleware processing (4ms)
```

## Configuração

### Variáveis de Ambiente

```bash
# Endpoint do Tempo (OpenTelemetry)
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

# Nome do serviço
OTEL_SERVICE_NAME=todoapp

# Modo de sampling
OTEL_TRACES_SAMPLER=always_on
```

### Configuração no Código

```go
// Telemetry config
telemetry, err := InitTelemetry(TelemetryConfig{
    ServiceName:    "todoapp",
    ServiceVersion: "1.0.0",
    MetricsPort:    "9091",
    OTLPEndpoint:   "localhost:4317",
})
```