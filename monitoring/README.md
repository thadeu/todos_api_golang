# TodoApp Monitoring

Este diretório contém a configuração completa de monitoramento para a aplicação TodoApp usando OpenTelemetry, Prometheus e Grafana.

## Arquitetura

```
TodoApp (Go) → OpenTelemetry → Prometheus → Grafana
```

- **TodoApp**: Aplicação Go com instrumentação OpenTelemetry
- **Prometheus**: Coleta e armazena métricas
- **Grafana**: Visualização de métricas através de dashboards

## Métricas Coletadas

### HTTP Requests
- `http_requests_total`: Total de requisições HTTP
- `http_request_duration_seconds`: Duração das requisições HTTP
- `http_active_connections`: Conexões HTTP ativas

### Sistema
- `memory_usage_bytes`: Uso de memória
- `cpu_usage_percent`: Uso de CPU
- `goroutines_total`: Número de goroutines

### Banco de Dados
- `database_operations_total`: Operações de banco de dados
- `todo_operations_total`: Operações de todos
- `user_operations_total`: Operações de usuários

## Como Usar

### 1. Iniciar os Serviços de Monitoramento

```bash
docker-compose up -d
```

### 2. Executar a Aplicação

```bash
go run cmd/api/main.go
```

### 3. Acessar as Interfaces

- **Grafana**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Métricas da App**: http://localhost:9091

## Dashboards Disponíveis

### TodoApp Overview
- Taxa de requisições HTTP
- Percentis de tempo de resposta (P50, P95)
- Distribuição de status HTTP
- Uso de memória
- Contagem de goroutines

### TodoApp Database
- Taxa de operações de banco de dados
- Distribuição de operações por tipo
- Taxa de operações de todos
- Taxa de operações de usuários

### TodoApp System
- Uso de CPU
- Uso de memória detalhado
- Contagem de goroutines
- Conexões HTTP ativas
- Estatísticas de heap

## Configuração

### Portas
- **9090**: Prometheus
- **9091**: Métricas da aplicação
- **3000**: Grafana

### Variáveis de Ambiente
- `METRICS_PORT`: Porta para métricas (padrão: 9091)
- `SERVICE_NAME`: Nome do serviço (padrão: todoapp)
- `SERVICE_VERSION`: Versão do serviço (padrão: 1.0.0)

## Troubleshooting

### Aplicação não está expondo métricas
1. Verifique se a porta 9091 está livre
2. Acesse http://localhost:9091 para verificar se as métricas estão sendo expostas
3. Verifique os logs da aplicação

### Prometheus não está coletando métricas
1. Verifique se a aplicação está rodando na porta 9091
2. Acesse http://localhost:9090/targets para ver o status dos targets
3. Verifique o arquivo `prometheus.yml`

### Grafana não está mostrando dados
1. Verifique se o Prometheus está configurado como datasource
2. Verifique se o Prometheus está coletando métricas
3. Verifique se os dashboards estão sendo carregados corretamente
