#!/bin/bash

# Script para testar cache com distributed tracing
# Execute a aplicação primeiro: go run cmd/api/main.go

echo "Testando Cache com Distributed Tracing..."
echo "========================================"

BASE_URL="http://localhost:8080"

echo "1. Primeiro, vamos fazer login para obter um token"
echo "Fazendo login..."

TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"12345678"}')

echo "Resposta do login: $TOKEN_RESPONSE"

# Extrair token (assumindo que a resposta contém "refresh_token")
TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "Erro: Não foi possível obter o token. Verifique se o usuário existe."
    echo "Criando usuário primeiro..."
    
    curl -s -X POST "$BASE_URL/signup" \
      -H "Content-Type: application/json" \
      -d '{"email":"test@example.com","password":"12345678"}'
    
    echo ""
    echo "Tentando login novamente..."
    TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth" \
      -H "Content-Type: application/json" \
      -d '{"email":"test@example.com","password":"12345678"}')
    
    TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ]; then
    echo "Erro: Ainda não foi possível obter o token. Verifique a aplicação."
    exit 1
fi

echo "Token obtido: ${TOKEN:0:20}..."

echo ""
echo "2. Testando Cache Miss (primeira requisição)"
echo "Fazendo primeira requisição GET para /todos..."

echo "Request 1 (Cache Miss):"
curl -s -w "Status: %{http_code}, Time: %{time_total}s, Cache: %{header_x-cache}\n" \
     -X GET "$BASE_URL/todos" \
     -H "Authorization: Bearer $TOKEN" \
     -o /dev/null

echo ""
echo "3. Testando Cache Hit (requisições subsequentes)"
echo "Fazendo 3 requests GET para /todos (deve usar cache)..."

for i in {2..4}; do
    echo "Request $i (Cache Hit):"
    curl -s -w "Status: %{http_code}, Time: %{time_total}s, Cache: %{header_x-cache}\n" \
         -X GET "$BASE_URL/todos" \
         -H "Authorization: Bearer $TOKEN" \
         -o /dev/null
    echo "---"
    sleep 1
done

echo ""
echo "4. Verificando headers de cache..."
curl -s -I -X GET "$BASE_URL/todos" \
     -H "Authorization: Bearer $TOKEN" \
     | grep -E "(X-Cache|HTTP/)"

echo ""
echo "5. Aguardando 4 segundos para o cache expirar (TTL=3s)..."
sleep 4

echo "Fazendo request após expiração do cache (deve ser Cache Miss):"
curl -s -w "Status: %{http_code}, Time: %{time_total}s, Cache: %{header_x-cache}\n" \
     -X GET "$BASE_URL/todos" \
     -H "Authorization: Bearer $TOKEN" \
     -o /dev/null

echo ""
echo "6. Testando com diferentes parâmetros de query (deve gerar cache separado)"
echo "Request com cursor=test (Cache Miss):"
curl -s -w "Status: %{http_code}, Time: %{time_total}s, Cache: %{header_x-cache}\n" \
     -X GET "$BASE_URL/todos?cursor=test" \
     -H "Authorization: Bearer $TOKEN" \
     -o /dev/null

echo ""
echo "Request com cursor=test novamente (Cache Hit):"
curl -s -w "Status: %{http_code}, Time: %{time_total}s, Cache: %{header_x-cache}\n" \
     -X GET "$BASE_URL/todos?cursor=test" \
     -H "Authorization: Bearer $TOKEN" \
     -o /dev/null

echo ""
echo "========================================"
echo "Teste concluído!"
echo ""
echo "Para visualizar os traces no Grafana:"
echo "1. Acesse: http://localhost:3000"
echo "2. Navegue para: Explore → Tempo"
echo "3. Query: {service_name=\"todoapp\"}"
echo "4. Procure por traces com spans:"
echo "   - cache.response.miss (primeira requisição)"
echo "   - cache.response.hit (requisições subsequentes)"
echo "   - cache.response.store (armazenamento no cache)"
echo ""
echo "Observe que:"
echo "- Primeiro request deve ter X-Cache: MISS e span cache.response.miss"
echo "- Requests subsequentes devem ter X-Cache: HIT e span cache.response.hit"
echo "- Request após expiração deve ter X-Cache: MISS e span cache.response.miss"
echo "- Requests com parâmetros diferentes devem ter cache separado"
echo "- Tempo de resposta deve ser muito menor em cache hits"
