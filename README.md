# 📘 Bukuo - Financial Accounting Platform

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![API Docs](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger)](http://localhost:8080/swagger/index.html)

A modern, enterprise-grade financial accounting platform built with **Clean Architecture** principles. Designed for Indonesian PSAK compliance.

## ✨ Features

- 🔐 **Authentication** - JWT-based auth with role-based access control
- 📊 **Chart of Accounts** - Hierarchical account structure (PSAK compliant)
- 📖 **Journal Entries** - Double-entry bookkeeping with validation
- 📈 **Financial Reports** - Trial Balance, Ledger, Income Statement, Balance Sheet, Cash Flow
- 🏢 **Multi-Tenant** - Company-based data isolation
- 🚀 **Production Ready** - Rate limiting, error handling, Swagger docs

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Delivery Layer                          │
│                 (HTTP Handlers, Middleware)                 │
├─────────────────────────────────────────────────────────────┤
│                     Usecase Layer                           │
│              (Business Logic, Orchestration)                │
├─────────────────────────────────────────────────────────────┤
│                     Domain Layer                            │
│           (Entities, Repository Interfaces)                 │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure Layer                      │
│              (PostgreSQL, External Services)                │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
bukuo/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── docs/                        # Swagger documentation
├── internal/
│   ├── config/                  # Configuration management
│   ├── database/                # Database connection & migrations
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/         # HTTP handlers
│   │       ├── middleware/      # Auth, rate limiting, error handling
│   │       └── router.go        # Route definitions
│   ├── domain/
│   │   ├── entity/              # Business entities
│   │   └── repository/          # Repository interfaces
│   ├── infrastructure/
│   │   └── persistence/
│   │       └── postgres/        # PostgreSQL implementations
│   └── usecase/                 # Business logic
│       ├── account/
│       ├── auth/
│       ├── journal/
│       ├── period/
│       └── report/
└── go.mod
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- [golang-migrate](https://github.com/golang-migrate/migrate) (for migrations)

### Installation

```bash
# Clone repository
git clone https://github.com/herman-xphp/bukuo.git
cd bukuo

# Install dependencies
go mod download

# Copy environment file
cp .env.example .env
# Edit .env with your database credentials
```

### Database Setup

```bash
# Create database
createdb bukuo_db

# Run migrations
migrate -path internal/database/migrations \
  -database "postgres://postgres:password@localhost:5432/bukuo_db?sslmode=disable" up
```

### Run Application

```bash
# Development
go run cmd/api/main.go

# Production build
go build -o bukuo cmd/api/main.go
./bukuo
```

Server starts at: `http://localhost:8080`

## 📚 API Documentation

### Swagger UI

After starting the server, visit: **http://localhost:8080/swagger/index.html**

### Endpoints Overview

| Method | Endpoint                        | Description             | Auth |
| ------ | ------------------------------- | ----------------------- | ---- |
| POST   | `/auth/register`                | Register company + user | ❌   |
| POST   | `/auth/login`                   | Login & get JWT         | ❌   |
| GET    | `/api/me`                       | Get current user        | ✅   |
| GET    | `/api/accounts`                 | List accounts           | ✅   |
| POST   | `/api/accounts`                 | Create account          | ✅   |
| GET    | `/api/periods`                  | List periods            | ✅   |
| POST   | `/api/periods`                  | Create period           | ✅   |
| POST   | `/api/journals`                 | Create journal          | ✅   |
| POST   | `/api/journals/:id/post`        | Post journal            | ✅   |
| POST   | `/api/journals/:id/reverse`     | Reverse journal         | ✅   |
| GET    | `/api/reports/trial-balance`    | Trial Balance           | ✅   |
| GET    | `/api/reports/income-statement` | Income Statement        | ✅   |
| GET    | `/api/reports/balance-sheet`    | Balance Sheet           | ✅   |
| GET    | `/api/reports/cash-flow`        | Cash Flow               | ✅   |

### Authentication

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "PT Example",
    "email": "admin@example.com",
    "password": "password123",
    "name": "Admin User"
  }'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'

# Use token for protected endpoints
curl http://localhost:8080/api/accounts \
  -H "Authorization: Bearer <your-token>"
```

## ⚙️ Configuration

| Variable      | Description                          | Default     |
| ------------- | ------------------------------------ | ----------- |
| `PORT`        | Server port                          | 8080        |
| `ENV`         | Environment (development/production) | development |
| `DB_HOST`     | Database host                        | localhost   |
| `DB_PORT`     | Database port                        | 5432        |
| `DB_USER`     | Database user                        | postgres    |
| `DB_PASSWORD` | Database password                    | -           |
| `DB_NAME`     | Database name                        | bukuo_db    |
| `JWT_SECRET`  | JWT signing secret                   | -           |

## 🧪 Testing

We respect your production data! Testing uses a separate database (`bukuo_test`).

### 1. Setup Test Database

One-time setup to create database and run migrations:

```bash
make setup-test-db
```

### 2. Run Tests

This command automatically sets `DB_NAME=bukuo_test` to prevent data pollution:

```bash
make test
```

### Manual Testing

If you need to run tests manually without Make, ensure you set the environment variable:

```bash
export DB_NAME=bukuo_test && go test ./...
```

### Test Coverage (Latest Sprint)

- **Domain Layer**: ~85%
- **Usecase Layer**: ~80% (Closing, Journal, Opening, Auth)
- **Middleware**: ~55%

## 🛡️ Security Features

- **JWT Authentication** - Token-based stateless auth
- **Password Hashing** - bcrypt with cost factor
- **Rate Limiting** - 100 requests/minute per IP
- **CORS** - Configurable origins
- **Input Validation** - Request validation with gin-binding

## 📦 Dependencies

| Package                                                     | Purpose                    |
| ----------------------------------------------------------- | -------------------------- |
| [gin-gonic/gin](https://github.com/gin-gonic/gin)           | HTTP framework             |
| [jackc/pgx](https://github.com/jackc/pgx)                   | PostgreSQL driver          |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt)         | JWT handling               |
| [shopspring/decimal](https://github.com/shopspring/decimal) | Precise decimal arithmetic |
| [google/uuid](https://github.com/google/uuid)               | UUID generation            |
| [swaggo/swag](https://github.com/swaggo/swag)               | Swagger documentation      |

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `test:` - Tests
- `refactor:` - Code refactoring

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **Herman** - _Initial work_ - [@herman-xphp](https://github.com/herman-xphp)

---

<p align="center">Made with ❤️ in Indonesia</p>
