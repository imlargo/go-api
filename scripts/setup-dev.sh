#!/bin/bash

# Development setup script for Go API

set -e

echo "🚀 Setting up development environment for Go API..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is required but not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is required but not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed. Please install Go first."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Install development tools
echo "📦 Installing development tools..."

# Install Air for hot reloading
if ! command -v air &> /dev/null; then
    echo "Installing Air for hot reloading..."
    go install github.com/cosmtrek/air@latest
fi

# Install swag for API documentation
if ! command -v swag &> /dev/null; then
    echo "Installing Swag for API documentation..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Install golangci-lint for linting
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint for code linting..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
fi

echo "✅ Development tools installed"

# Download Go dependencies
echo "📥 Downloading Go dependencies..."
go mod tidy
go mod download

# Generate API documentation
echo "📚 Generating API documentation..."
swag init -g cmd/api/main.go --parseDependency --parseInternal --parseVendor -o api/docs

# Setup environment file
if [ ! -f .env ]; then
    echo "⚙️ Creating .env file from template..."
    cp .env.example .env 2>/dev/null || cat > .env << 'EOF'
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Server
API_URL=localhost
PORT=8000

# Rate Limiting
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_TIMEFRAME=60

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ISSUER=go-api
JWT_AUDIENCE=go-api-users

# API Key
API_KEY=your-api-key-change-in-production

# Storage (Cloudflare R2)
STORAGE_BUCKET_NAME=your-bucket
STORAGE_ACCOUNT_ID=your-account-id
STORAGE_ACCESS_KEY_ID=your-access-key
STORAGE_SECRET_ACCESS_KEY=your-secret-key
STORAGE_PUBLIC_DOMAIN=
STORAGE_USE_PUBLIC_URL=false

# Push Notifications
VAPID_PUBLIC_KEY=your-vapid-public-key
VAPID_PRIVATE_KEY=your-vapid-private-key
EOF
    echo "⚠️  Please update the .env file with your actual configuration values"
fi

# Start development services
echo "🐳 Starting development services with Docker Compose..."
docker-compose up -d postgres redis

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Run database migrations
echo "🗃️ Running database migrations..."
go run cmd/migrations/main.go || echo "⚠️ Migrations failed or not configured"

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Available commands:"
echo "  make dev       - Start development server with hot reload"
echo "  make build     - Build the application"
echo "  make test      - Run tests"
echo "  make lint      - Run linter"
echo "  make swag      - Generate API documentation"
echo "  make docker    - Start all services with Docker"
echo ""
echo "The API will be available at: http://localhost:8000"
echo "API documentation: http://localhost:8000/swagger/index.html"
echo ""