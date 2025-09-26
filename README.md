# Go API Template

Un template/boilerplate completo para APIs en Go usando Gin, GORM, Redis, y otras mejores prácticas.

## 🚀 Características

- **Framework Web**: Gin HTTP web framework
- **Base de Datos**: PostgreSQL con GORM ORM
- **Cache**: Redis para caching y rate limiting  
- **Autenticación**: JWT tokens
- **Documentación**: Swagger/OpenAPI automático
- **Monitoring**: Métricas Prometheus
- **Storage**: Soporte para Cloudflare R2 (S3 compatible)
- **Notifications**: Server-Sent Events (SSE) y Web Push
- **Containerización**: Docker y Docker Compose
- **CI/CD**: GitHub Actions pipeline
- **Testing**: Framework de testing con testify
- **Health Checks**: Endpoints de salud para dependencias

## 📋 Prerrequisitos

- Go 1.24.4 o superior
- Docker y Docker Compose
- PostgreSQL (para desarrollo local sin Docker)
- Redis (para desarrollo local sin Docker)

## 🛠️ Setup Rápido

### Opción 1: Setup Automático (Recomendado)

```bash
# Clonar el repositorio
git clone https://github.com/imlargo/go-api.git
cd go-api

# Ejecutar setup automático
./scripts/setup-dev.sh
```

### Opción 2: Setup Manual

1. **Instalar herramientas de desarrollo:**

```bash
# Air para hot reloading
go install github.com/cosmtrek/air@latest

# Swag para documentación API
go install github.com/swaggo/swag/cmd/swag@latest

# golangci-lint para linting
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

2. **Configurar variables de entorno:**

```bash
cp .env.example .env
# Editar .env con tus valores
```

3. **Iniciar servicios con Docker:**

```bash
make docker
```

4. **Ejecutar migraciones:**

```bash
make migrations
```

5. **Generar documentación:**

```bash
make swag
```

## 🐳 Docker

### Desarrollo con Docker Compose

```bash
# Iniciar todos los servicios
make docker

# Ver logs
make docker-logs

# Parar servicios
make docker-down

# Incluir herramientas de administración (pgAdmin, Redis Commander)
docker-compose --profile tools up -d
```

### Build de imagen Docker

```bash
# Build imagen
make docker-build

# Run contenedor
docker run -p 8000:8000 go-api:latest
```

## 🔧 Comandos de Desarrollo

```bash
# Ayuda con comandos disponibles
make help

# Desarrollo con hot reload
make dev

# Build de la aplicación
make build

# Ejecutar tests
make test

# Coverage de tests
make test-coverage

# Linting
make lint

# Formatear código
make fmt

# Pipeline CI completo
make ci
```

## 📊 Endpoints Importantes

- **API**: `http://localhost:8000`
- **Health Check**: `http://localhost:8000/health`
- **Readiness**: `http://localhost:8000/ready`
- **Liveness**: `http://localhost:8000/live`
- **Documentación**: `http://localhost:8000/internal/docs/`
- **Métricas**: `http://localhost:8000/internal/metrics`

### Herramientas de Administración (con --profile tools)

- **pgAdmin**: `http://localhost:8080` (admin@admin.com / admin)
- **Redis Commander**: `http://localhost:8081`

## 🏗️ Estructura del Proyecto

```
.
├── api/docs/           # Documentación Swagger generada
├── cmd/
│   ├── api/           # Aplicación principal
│   └── migrations/    # Migraciones de base de datos
├── internal/          # Código interno de la aplicación
│   ├── cache/        # Configuración de cache
│   ├── config/       # Configuración de la app
│   ├── handlers/     # HTTP handlers
│   ├── middleware/   # HTTP middleware
│   ├── models/       # Modelos de datos
│   ├── repositories/ # Capa de datos
│   ├── services/     # Lógica de negocio
│   └── ...
├── pkg/              # Librerías reutilizables
├── scripts/          # Scripts de desarrollo
├── .github/workflows/ # CI/CD pipelines
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## ⚙️ Variables de Entorno

```bash
# Base de datos
DATABASE_URL=postgres://user:pass@localhost:5432/dbname

# Redis
REDIS_URL=redis://localhost:6379

# Servidor
API_URL=localhost
PORT=8000

# Rate Limiting
RATE_LIMIT_MAX_REQUESTS=100
RATE_LIMIT_TIMEFRAME=60

# JWT
JWT_SECRET=your-jwt-secret
JWT_ISSUER=your-app
JWT_AUDIENCE=your-audience

# API Key
API_KEY=your-api-key

# Storage (Cloudflare R2)
STORAGE_BUCKET_NAME=your-bucket
STORAGE_ACCOUNT_ID=your-account-id
STORAGE_ACCESS_KEY_ID=your-access-key
STORAGE_SECRET_ACCESS_KEY=your-secret-key
```

## 🧪 Testing

```bash
# Ejecutar todos los tests
make test

# Tests con coverage
make test-coverage

# Tests específicos
go test ./internal/handlers/...

# Benchmarks
go test -bench=. ./...
```

## 📈 Monitoring y Observabilidad

### Health Checks

- `GET /health` - Estado general con checks de dependencias
- `GET /ready` - Readiness probe para Kubernetes
- `GET /live` - Liveness probe para Kubernetes

### Métricas Prometheus

- `GET /internal/metrics` - Métricas en formato Prometheus
- Métricas incluidas:
  - HTTP requests (duración, código de estado)
  - Rate limiting
  - Database connections
  - Custom business metrics

## 🔐 Seguridad

- JWT authentication
- Rate limiting configurable
- API key authentication para endpoints internos
- CORS configuration
- Input validation
- SQL injection protection (GORM)

## 🚀 Despliegue

### CI/CD con GitHub Actions

El proyecto incluye un pipeline completo que:

1. Ejecuta tests y linting
2. Escanea vulnerabilidades de seguridad
3. Build y push de imagen Docker
4. Deploy automático (configurable)

### Kubernetes

```yaml
# Ejemplo de deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-api
spec:
  template:
    spec:
      containers:
      - name: go-api
        image: ghcr.io/imlargo/go-api:latest
        ports:
        - containerPort: 8000
        livenessProbe:
          httpGet:
            path: /live
            port: 8000
        readinessProbe:
          httpGet:
            path: /ready
            port: 8000
```

## 🤝 Contribuir

1. Fork el proyecto
2. Crear branch feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push al branch (`git push origin feature/AmazingFeature`)
5. Abrir Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para detalles.

## 📚 Recursos Adicionales

- [Documentación Gin](https://gin-gonic.com/)
- [Documentación GORM](https://gorm.io/)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Métricas Prometheus](https://prometheus.io/docs/guides/go-application/)

---

**¿Preguntas o problemas?** Abre un [issue](https://github.com/imlargo/go-api/issues) en GitHub.
