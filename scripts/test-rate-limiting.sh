#!/bin/bash

# Script para testar rate limiting
# Execute a aplicação primeiro: go run cmd/api/main.go

echo "Testando Rate Limiting..."
echo "=========================="

BASE_URL="http://localhost:8080"

echo "1. Testando rate limiting no endpoint /signup (limite: 5 requests/minuto)"
echo "Fazendo 6 requests rapidamente..."

for i in {1..6}; do
    echo "Request $i:"
    curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -X POST "$BASE_URL/signup" \
         -H "Content-Type: application/json" \
         -d '{"email":"test'$i'@example.com","password":"12345678"}' \
         -o /dev/null
    echo "---"
done

echo ""
echo "2. Testando rate limiting no endpoint /auth (limite: 10 requests/minuto)"
echo "Fazendo 11 requests rapidamente..."

for i in {1..11}; do
    echo "Request $i:"
    curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" \
         -X POST "$BASE_URL/auth" \
         -H "Content-Type: application/json" \
         -d '{"email":"test@example.com","password":"12345678"}' \
         -o /dev/null
    echo "---"
done

echo ""
echo "3. Verificando headers de rate limiting..."
curl -s -I -X POST "$BASE_URL/signup" \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"12345678"}' \
     | grep -E "(X-RateLimit|HTTP/)"

echo ""
echo "Teste concluído!"
echo "Observe que após o limite ser atingido, você deve receber status 429 (Too Many Requests)"
