#!/bin/bash

echo "🚀 Setting up TodoApp Monitoring..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Create monitoring directories if they don't exist
mkdir -p monitoring/grafana/provisioning/datasources
mkdir -p monitoring/grafana/provisioning/dashboards
mkdir -p monitoring/grafana/dashboards

echo "📁 Created monitoring directories"

# Start monitoring services
echo "🐳 Starting Prometheus and Grafana..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 10

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo "✅ Monitoring services are running!"
    echo ""
    echo "📊 Access your monitoring tools:"
    echo "   Grafana:     http://localhost:3000 (admin/admin123)"
    echo ""
else
    echo "❌ Failed to start monitoring services"
    echo "Check the logs with: docker-compose logs"
    exit 1
fi
