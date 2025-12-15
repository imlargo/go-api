# Go API - Medusa Framework

Un template moderno y completo de API en Go construido sobre el framework personalizado **Medusa**. Este proyecto está diseñado para ser una base sólida y escalable para aplicaciones backend que requieren funcionalidades avanzadas como autenticación, caché, almacenamiento en la nube, mensajería en tiempo real y más.

## 🎯 Objetivo del Proyecto

El objetivo es construir un framework de Go robusto y modular que simplifique el desarrollo de APIs empresariales modernas. Medusa proporciona componentes reutilizables y abstracciones bien definidas para servicios comunes, permitiendo un desarrollo rápido sin sacrificar la calidad o la escalabilidad.

## ✨ Características Principales

### Servicios Integrados

- **Server-Sent Events (SSE)**: Comunicación en tiempo real del servidor al cliente
  - Sistema de publicación/suscripción
  - Gestión de eventos y clientes
  - Dos implementaciones: SSE v1 y SSE v2
  
- **Caché**: Sistema de caché distribuido basado en Redis
  - Caché en memoria con respaldo persistente
  - Operaciones optimizadas de lectura/escritura

- **Almacenamiento**: Gestión de archivos en la nube
  - Soporte para múltiples proveedores (AWS S3, Cloudflare R2)
  - Integración con AWS SDK v2

- **PubSub**: Sistema de mensajería asíncrona
  - Integración con RabbitMQ
  - Patrón publicador-suscriptor para desacoplamiento de componentes

- **Email**: Servicio de correo electrónico transaccional
  - Integración con Resend
  - Plantillas y envío masivo

- **Push Notifications**: Notificaciones push web
  - Soporte para Web Push API
  - Gestión de suscripciones

### Middleware

- **Autenticación JWT**: Validación de tokens y gestión de sesiones
- **API Keys**: Autenticación por clave de API (header y bearer)
- **CORS**: Configuración flexible de políticas de origen cruzado
- **Rate Limiting**: Limitación de tasa basada en token bucket
- **Métricas**: Integración con Prometheus para monitoreo

### Core Components

- **Logger**: Sistema de logging estructurado con Zap
- **HTTP Server**: Servidor HTTP basado en Gin
- **Repository Pattern**: Abstracción de capa de datos con GORM
- **Responses**: Utilidades para respuestas HTTP estandarizadas
- **Configuration**: Gestión de configuración con variables de entorno

## 🏗️ Arquitectura

El proyecto sigue una arquitectura limpia y modular:

```
.
├── cmd/                    # Puntos de entrada de la aplicación
│   ├── api/               # API principal
│   ├── sse/               # Servidor SSE dedicado
│   └── cli/               # Herramientas CLI
├── internal/              # Código interno de la aplicación
│   ├── config/           # Configuración
│   ├── database/         # Conexiones a BD
│   ├── handlers/         # Manejadores HTTP
│   ├── models/           # Modelos de dominio
│   ├── repository/       # Capa de repositorio
│   ├── service/          # Lógica de negocio
│   └── store/            # Almacenamiento de datos
└── pkg/                   # Código reutilizable
    └── medusa/           # Framework Medusa
        ├── core/         # Componentes centrales
        ├── middleware/   # Middleware HTTP
        ├── services/     # Servicios externos
        └── tools/        # Utilidades
```

## 🚀 Instalación

### Prerrequisitos

- Go 1.25.4 o superior
- PostgreSQL
- Redis
- RabbitMQ (opcional, para PubSub)
- Cuentas en servicios externos (AWS S3/R2, Resend, etc.)

### Configuración

1. Clona el repositorio:
```bash
git clone https://github.com/imlargo/go-api.git
cd go-api
```

2. Instala las dependencias:
```bash
go mod download
```

3. Configura las variables de entorno (crea un archivo `.env`):
```bash
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database
DATABASE_URL=postgres://user:password@localhost:5432/dbname

# Redis
REDIS_URL=redis://localhost:6379

# Rate Limiter
RATE_LIMITER_ENABLED=true
RATE_LIMITER_REQUESTS_PER_TIME_FRAME=100
RATE_LIMITER_TIME_FRAME=60s

# AWS S3 / Cloudflare R2
STORAGE_PROVIDER=r2
STORAGE_ACCOUNT_ID=your_account_id
STORAGE_ACCESS_KEY_ID=your_access_key
STORAGE_SECRET_ACCESS_KEY=your_secret_key
STORAGE_BUCKET_NAME=your_bucket

# Otros servicios...
```

4. Ejecuta las migraciones de base de datos:
```bash
go run cmd/api/main.go
```

## 📖 Uso

### Ejecutar el servidor API principal

```bash
go run cmd/api/main.go
```

El servidor estará disponible en `http://localhost:8080`

### Ejecutar el servidor SSE

```bash
go run cmd/sse/main.go
```

### Endpoints básicos

- `GET /ping` - Health check
- `GET /sse/listen` - Conectarse al stream SSE
- `POST /sse/publish` - Publicar eventos SSE

## 🛠️ Comandos Disponibles

```bash
# Formatear código
make format

# Generar documentación Swagger
make swag
```

## 🔧 Desarrollo

El proyecto utiliza:

- **Gin** como framework HTTP
- **GORM** para ORM y migraciones
- **Zap** para logging estructurado
- **Prometheus** para métricas
- **Air** para hot reload en desarrollo

### Hot Reload

Para desarrollo con recarga automática, usa Air:

```bash
air
```

La configuración está en `.air.toml`

## 🤝 Contribuciones

Este es un proyecto personal en desarrollo activo. Las sugerencias y contribuciones son bienvenidas.

## 📝 Licencia

Este proyecto está en desarrollo y no tiene una licencia definida aún.

## 🎓 Aprendizajes y Objetivos

Este proyecto es parte de mi proceso de aprendizaje en Go y arquitecturas backend modernas. Los objetivos incluyen:

- Dominar patrones de diseño en Go
- Construir un framework modular y reutilizable
- Implementar servicios distribuidos y escalables
- Aplicar mejores prácticas de ingeniería de software
- Crear una base de código mantenible y bien documentada

---

**Nota**: Este proyecto está en desarrollo activo y la API puede cambiar sin previo aviso.
