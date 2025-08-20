<div align="center">
  <h1>🐈‍⬛ Fusioncat</h1>
  <p><strong>
    A communication stack for asynchronous and event-driven backend applications
    and AI systems.
  </strong><br>
  Reduce chaos to zero in your Kafka/AMQP/MQTT/Webhook-based services and architectures</p>
  
  <p>
    <a href="https://github.com/fusioncatltd/fusioncat/actions">
      <img src="https://github.com/fusioncatltd/fusioncat/actions/workflows/docker-publish.yml/badge.svg" alt="CI Status">
    </a>
    <a href="https://github.com/fusioncatltd/fusioncat/releases">
      <img src="https://img.shields.io/github/v/release/fusioncatltd/fusioncat" alt="Release">
    </a>
    <a href="LICENSE">
      <img src="https://img.shields.io/github/license/fusioncatltd/fusioncat" alt="License">
    </a>
    <a href="https://fusioncat.dev">
      <img src="https://img.shields.io/badge/website-fusioncat.dev-blue" alt="Website">
    </a>
  </p>

  <p>
    <a href="#-features">Features</a> •
    <a href="#-quick-start">Quick Start</a> •
    <a href="#-installation">Installation</a> •
    <a href="#-api-documentation">API Docs</a> •
    <a href="#-contributing">Contributing</a>
  </p>
</div>

---

## 📖 Overview

Fusioncat is a solution for managing the complexity of distributed systems and asynchronous messaging.
It helps engineering teams design, develop, deploy, and monitor backend services
that communicate via Kafka, AMQP, MQTT, or Webhooks.

📖 Engineering flow with Fusioncat

1. **Design** — The team defines system architecture using Fusioncat’s visual tools or Fusionlang.
1. **Decompose** — Fusioncat breaks the design into components and schemas.
1. **Generate** — Fusioncat produces boilerplate code for cross-service communications and injects it into the codebase
1. **Deploy** — The team ships services to production.
1. **Monitor** — Fusioncat validates and monitors data flows through the generated code.
1. **Evolve** — The team iterates on the design, and Fusioncat automatically updates the codebase.

### Why Fusioncat?

- **🎯 Reduce Integration Complexity**: Engineering teams spend ~30% of their time designing, building, deploying, and maintaining cross-software asynchronous communications.
Fusioncat dramatically reduces this overhead.
- - **🎯 Documentation first**: Fusionlang enables teams to define their architecture in a human-readable format,
which is then used to automatically generate code and documentation.
- **⚡ 3X Faster Onboarding**: New team members can understand and contribute to complex systems in days, not weeks
- **🛡️ Prevent Production Outages**: Manage data contract compatibility and catch breaking changes before deployment
- **🌐 Protocol Agnostic**: Seamlessly work with Kafka, RabbitMQ, MQTT,
and webhooks from a single platform. Define data contracts using yor preferred schema format (JSON Schema,
Protocol Buffers, Avro).

## ✨ Features

### Core Capabilities

- **🔄 Multi-Protocol Support**
  - Apache Kafka
  - RabbitMQ (AMQP)
  - MQTT
  - Webhooks
  - Database events (coming soon)

- **📝 Built-in Schema Management**
  - JSON Schema (Protocol Buffers, Avro coming soon)
  - Schema versioning,
  - Code generation for multiple languages (currently supports Go, other languages coming soon) 

- **🏗️ Code Generation**
  - Generate boilerplate code for messaging clients
  - Type-safe message contracts
  - Ready-to-use code snippets for common operations.

- **🎨 Visual Design Tools (coming soon)**
  - Map complex system relationships
  - Visualize data flows
  - Collaborate with team members
  - Export architecture diagrams

- **🔐 Enterprise Ready**
  - Self-hosted open source solution.
  - Control the code.
  - Safety: Fusioncat doesn't have any access to your data.

## 🚀 Quick Start

Fusioncat is packaged into a Docker container for easy deployment. 
External PostgreSQL database is required.

### Using Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/fusioncatltd/fusioncat:latest

# Run with environment variables
docker run -d \
  --name fusioncat \
  -p 8080:8080 \
  -e PG_HOST=your-postgres-host \
  -e PG_PORT=5432 \
  -e PG_USER=your-db-user \
  -e PG_PASSWORD=your-db-password \
  -e PG_DB_NAME=fusioncat \
  -e JWT_SECRET=your-secret-key \
  ghcr.io/fusioncatltd/fusioncat:latest
```

### Health Check

```bash
curl http://localhost:8080/health
```

## 📦 Installation

### Prerequisites

- PostgreSQL 13+ database
- Docker (for containerized deployment)
- Go 1.23+ (for local development)

### Option 1: Docker Compose

Create a `docker-compose.yml`:

```yaml
version: '3.8'

services:
  fusioncat:
    image: ghcr.io/fusioncatltd/fusioncat:latest
    ports:
      - "8080:8080"
    environment:
      PG_HOST: postgres
      PG_PORT: 5432
      PG_USER: fusioncat
      PG_PASSWORD: ${DB_PASSWORD}
      PG_DB_NAME: fusioncat
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: fusioncat
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: fusioncat
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

Run with:
```bash
export DB_PASSWORD=your-secure-password
export JWT_SECRET=your-jwt-secret
docker-compose up -d
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/fusioncatltd/fusioncat.git
cd fusioncat

# Copy environment template
cp .env.template .env
# Edit .env with your configuration

# Run locally
make run

# Or build Docker image
make docker-build
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PG_HOST` | PostgreSQL host | localhost | Yes |
| `PG_PORT` | PostgreSQL port | 5432 | Yes |
| `PG_USER` | Database user | - | Yes |
| `PG_PASSWORD` | Database password | - | Yes |
| `PG_DB_NAME` | Database name | fusioncat | Yes |
| `PG_SSLMODE` | SSL mode for PostgreSQL | require | No |
| `JWT_SECRET` | Secret key for JWT tokens | - | Yes |
| `ADMIN_URL` | Admin panel URL | http://localhost:3000 | No |
| `PATH_TO_STUBS_TEMPLATES_FOLDER` | Code generation templates path | /app/templates | No |
| `JSON_SCHEMA_CONVERTOR_CMD` | Path to quicktype binary | /usr/bin/quicktype | No |

## 📚 API Documentation

### Swagger/OpenAPI

The API documentation is available at:
```
http://localhost:8080/swagger/index.html
```

### Basic Usage Example

```bash
# Create a user
curl -X POST http://localhost:8080/v1/public/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'

# Create a project
curl -X POST http://localhost:8080/v1/protected/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "description": "Project description"
  }'
```

### Key API Endpoints

- **Authentication**
  - `POST /v1/public/users` - Register new user
  - `POST /v1/public/auth/login` - Login

- **Projects**
  - `GET /v1/protected/projects` - List projects
  - `POST /v1/protected/projects` - Create project
  - `POST /v1/protected/projects/:id/imports` - Import AsyncAPI specification

- **Apps & Services**
  - `GET /v1/protected/apps/:id/usage` - Get app usage matrix
  - `GET /v1/protected/apps/:id/code/:language` - Generate code

- **Schemas**
  - `POST /v1/protected/schemas` - Create schema
  - `GET /v1/protected/schemas/:id/code/:language` - Generate schema code

## 🛠️ Development

### Prerequisites

- Go 1.23+
- Node.js 18+ (for quicktype)
- PostgreSQL 13+
- Make

### Setup

```bash
# Install dependencies
go mod download
npm install -g quicktype

# Run tests
make test

# Run with hot reload
go install github.com/cosmtrek/air@latest
air
```

### Project Structure

```
fusioncat/
├── api/                 # HTTP API handlers
│   ├── protected_endpoints/
│   └── public_endpoints/
├── logic/              # Business logic
├── db/                 # Database models and connections
├── templates/          # Code generation templates
├── deploy/             # Deployment configurations
├── migrations/         # Database migrations
└── tests/              # Test suites
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Workflow

```bash
# Run tests
make test

# Check code formatting
go fmt ./...

# Update Swagger docs
swag init

# Test Docker build
make docker-test
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🌟 Support

- **Documentation**: [https://fusioncat.dev/docs](https://fusioncat.dev/docs)
- **Website**: [https://fusioncat.dev](https://fusioncat.dev)
- **Email**: [andrey@fusioncat.dev](mailto:andrey@fusioncat.dev)
- **Issues**: [GitHub Issues](https://github.com/fusioncatltd/fusioncat/issues)

## 🚦 Roadmap

- [x] Schema management with versioning
- [x] Code generation for Golang
- [x] Docker container distribution
- [ ] Multi-protocol support (Kafka, AMQP, MQTT). Currently tested on Kafka.
- [ ] Multi-language code generation. Currently supports Golang.
- [ ] Real-time monitoring dashboard for tracking and analyzing data flows and events.

## 🏆 Acknowledgments

- Built with [Gin Web Framework](https://github.com/gin-gonic/gin)
- Schema conversion powered by [Quicktype](https://quicktype.io)
- AsyncAPI specification support

## 📊 Status

Fusioncat is currently in **closed alpha**. 

---

<div align="center">
  <p>Made with ❤️ by the Fusioncat Team</p>
  <p>
    <a href="https://fusioncat.dev">Website</a> •
    <a href="https://github.com/fusioncatltd/fusioncat">GitHub</a> •
  </p>
</div>