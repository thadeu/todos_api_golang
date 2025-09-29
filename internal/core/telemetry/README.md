# Domain Probe - Telemetry no Core

O **Domain Probe** é um padrão de arquitetura hexagonal que permite ao domínio emitir eventos de telemetria sem acoplamento direto com ferramentas específicas de observabilidade.

## 🎯 Conceito

Em arquitetura hexagonal, o **domínio** (core) não deve conhecer implementações específicas de infraestrutura. O Domain Probe resolve isso criando uma interface no domínio que pode ser implementada por diferentes provedores de telemetria.

## 📋 Benefícios

- ✅ **Separação de responsabilidades**: Domínio foca em regras de negócio
- ✅ **Troca fácil de provedores**: Mude de Prometheus+Loki+Tempo para DataDog sem alterar domínio
- ✅ **Testabilidade**: Use `NoOpProbe` para testes sem telemetria
- ✅ **Observabilidade consistente**: Mesmo padrão em toda aplicação

## 🏗️ Arquitetura

```
┌─────────────────────────────────────────┐
│              Domain                     │
│                                         │
│  ┌────────────────────────────────────┐ │
│  │        TelemetryProbe              │ │  ← Interface no domínio
│  │                                    │ │
│  │  - RecordRepositoryOperation()     │ │
│  │  - RecordServiceOperation()        │ │
│  │  - RecordBusinessEvent()           │ │
│  │  - RecordHTTPOperation()           │ │
│  │  - RecordError()                   │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
                    │
                    │
          ┌─────────┴─────────┐
          │                   │
┌─────────▼─────────┐ ┌───────▼─────────┐
│   OTELProbe       │ │   NoOpProbe     │
│                   │ │                 │ ← Implementações
│  - OpenTelemetry  │ │  - No operation│
│  - Prometheus     │ │  - For testing │
│  - Loki           │ │                 │
│  - Tempo          │ │                 │
└───────────────────┘ └─────────────────┘
```

## 🚀 Como Usar

### 1. Importe a interface

```go
import "todoapp/internal/core/port"

type TodoRepository struct {
    probe port.TelemetryProbe
}
```

### 2. Use nos repositórios

```go
func (r *TodoRepository) Create(ctx context.Context, todo domain.Todo) error {
    // Registra início da operação
    operation := telemetry.StartOperation(r.probe, ctx, "Create", "todo")

    // Executa operação
    err := r.db.Create(ctx, todo)

    // Registra fim da operação
    operation.End(err)

    return err
}
```

### **Padrão Recomendado** (Centralizado no Telemetry Probe):
```go
func (r *TodoRepository) GetAllWithCursor(ctx context.Context, userId int, limit int, cursor string) ([]domain.Todo, bool, error) {
    // 1. Criar span usando o telemetry probe (único ponto de contato)
    ctx, span := r.probe.StartRepositorySpan(ctx, "GetAllWithCursor", "todo", []attribute.KeyValue{
        attribute.String("db.system", "sqlite"),
        attribute.String("db.table", "todos"),
        attribute.Int("user.id", userId),
        attribute.Int("pagination.limit", limit),
        attribute.String("pagination.cursor", cursor),
    })
    defer span.End()

    // 2. Iniciar operação de telemetria (métricas e logs)
    startTime := time.Now()
    operation := telemetry.StartOperation(r.probe, ctx, "GetAllWithCursor", "todo")

    // 3. Garantir registro de duração no span
    defer func() {
        duration := time.Since(startTime)
        span.SetAttributes(attribute.Int64("operation.duration_ns", duration.Nanoseconds()))
    }()

    // ... lógica de negócio ...

    // 4. Executar query
    rows, err := r.db.QueryContext(ctx, sql, args...)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        operation.End(err)
        return nil, false, err
    }

    // 5. Processar resultados
    var todos []domain.Todo
    err = r.scanner.ScanRowsToSlice(rows, &todos)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        span.RecordError(err)
        operation.End(err)
        return nil, false, err
    }

    // 6. Atualizar span com resultados da operação
    span.SetAttributes(
        attribute.Int("db.rows_returned", len(todos)),
        attribute.Bool("db.has_next", hasNext),
    )

    // 7. Sucesso - marcar span como OK
    span.SetStatus(codes.Ok, "")
    operation.End(nil)
    return todos, hasNext, nil
}
```

### 3. Use nos serviços

```go
func (s *TodoService) CreateTodo(ctx context.Context, req CreateTodoRequest) (*Todo, error) {
    start := time.Now()

    todo, err := s.repo.Create(ctx, req.ToDomain())

    duration := time.Since(start).Nanoseconds()
    s.probe.RecordServiceOperation(ctx, "todo", "CreateTodo", req.UserID, duration, err)

    return todo, err
}
```

### 4. Registre eventos de negócio

```go
func (s *TodoService) CompleteTodo(ctx context.Context, todoID string, userID int) error {
    // ... lógica de negócio ...

    // Registra evento de negócio
    s.probe.RecordBusinessEvent(ctx, "completed", "todo", todoID, userID, map[string]interface{}{
        "completed_at": time.Now(),
        "previous_status": "in_progress",
    })

    return nil
}
```

## 📊 Tipos de Telemetria

### Repository Operations
```go
probe.RecordRepositoryOperation(ctx, "FindByID", "user", duration, err)
```

### Service Operations
```go
probe.RecordServiceOperation(ctx, "auth", "Login", userID, duration, err)
```

### Business Events
```go
probe.RecordBusinessEvent(ctx, "created", "todo", todoID, userID, metadata)
```

### HTTP Operations
```go
probe.RecordHTTPOperation(ctx, "POST", "/api/todos", 201, duration)
```

### Errors
```go
probe.RecordError(ctx, "CreateTodo", err, map[string]interface{}{
    "user_id": 123,
    "validation_errors": []string{"title required"},
})
```

## 🧪 Testes

### Usando NoOpProbe (sem telemetria)

```go
func TestTodoService(t *testing.T) {
    probe := telemetry.NewNoOpProbe() // Sem efeitos colaterais
    service := NewTodoService(repo, probe)

    // Testes normais...
}
```

### Usando OTELProbe (com telemetria)

```go
func TestTodoServiceWithTelemetry(t *testing.T) {
    probe := telemetry.NewOTELProbe(slog.Default())
    service := NewTodoService(repo, probe)

    // Testes com telemetria real...
}
```

## 🔧 Implementações Disponíveis

### OTELProbe
- ✅ Tracing com OpenTelemetry
- ✅ Métricas com Prometheus
- ✅ Logging com Loki
- ✅ Correlação trace_id/span_id

### NoOpProbe
- ✅ Não faz nada (útil para testes)
- ✅ Zero overhead
- ✅ Interface completa implementada

### Future Implementações
- DataDogProbe
- NewRelicProbe
- CustomProbe

## 📈 Métricas Geradas

### Counters
- `todo_operations_total{operation="CreateTodo"}`
- `user_operations_total{operation="Login"}`
- `database_operations_total{operation="SELECT", table="todos"}`

### Histograms
- `http_request_duration_seconds{method="GET", path="/todos"}`

### Gauges
- `memory_usage_bytes`
- `goroutines_total`

## 🔍 Tracing

Cada operação cria spans com:
- **Service**: `repository.todo.Create`
- **Operation**: nome da operação
- **Attributes**: user_id, entity, duration, etc.
- **Error handling**: spans marcados como erro quando há falhas

## 🎯 Vantagens da Implementação

### **🏗️ Arquitetura Centralizada** (Telemetry como Único Ponto)

| Aspecto | Múltiplos Pontos | **Telemetry Centralizado** |
|---------|------------------|---------------------------|
| Tracing Distribuído | ✅ Vários otel.Tracer() | ✅ **probe.StartRepositorySpan()** |
| Métricas & Logs | ❌ Código espalhado | ✅ **probe.Record*() unificado** |
| Acoplamento | ❌ Import tracing everywhere | ✅ **Apenas port.TelemetryProbe** |
| Testabilidade | ❌ Mocks complexos | ✅ **NoOpProbe ou custom probe** |
| Manutenibilidade | ❌ Mudanças em múltiplos lugares | ✅ **Mudança apenas na implementação** |
| Consistência | ❌ Padrões diferentes | ✅ **Padrão único e consistente** |
| Business Events | ❌ Implementação ad-hoc | ✅ **probe.RecordBusinessEvent()** |

### **🎨 Padrões Implementados**

#### **1. Domain Probe Pattern**
- ✅ Interface no domínio (`TelemetryProbe`)
- ✅ Implementações plugáveis (`OTELProbe`, `NoOpProbe`)
- ✅ Separação de responsabilidades

#### **2. OpenTelemetry Spans**
- ✅ Distributed tracing nativo
- ✅ Atributos ricos e contextuais
- ✅ Integração com Jaeger/Tempo

#### **3. Business Events**
- ✅ Eventos de domínio estruturados
- ✅ Metadata contextual
- ✅ Correlação com operações

#### **4. Operation Timing**
- ✅ Medição precisa de duração
- ✅ Tratamento de erros consistente
- ✅ Métricas automáticas

## 🚀 Próximos Passos

1. **Implementar em todos os repositories**
2. **Adicionar probes em todos os services**
3. **Criar probes para outros providers** (DataDog, etc.)
4. **Adicionar métricas customizadas** por domínio
5. **Implementar sampling** baseado em configuração
