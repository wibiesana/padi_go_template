<div align="center">

# 🌾 PADI REST API FRAMEWORK (GO)

**The Industrial-Grade, Zero-Bloat REST API Engine for Go**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![Architecture](https://img.shields.io/badge/Architecture-Code_Generation-success?style=flat-square)](#)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](#)

> **"Stop wrestling with massive frameworks. Start shipping APIs in seconds."**

</div>

Welcome to the **Padi REST API Go Framework** — a high-performance Go port of the Padi REST API framework designed for developers who crave raw speed, absolute clarity, and an uncompromising Developer Experience (DX).

---

## ✨ Features

- ⚡ **Zero-Bloat & Raw Speed**: Minimal latency HTTP pipeline powered by standard `database/sql`, `chi` router, and pure Go architecture (no external ORM dependencies).
- 🌾 **Pure Native ActiveRecord & Fluent Query Builder**: Built-in `model.Save()`, `model.Delete()`, `(User{}).Find(id)`, and `query.New("table").Where(...).Paginate(...)`.
- 🛠️ **Signature Auto CRUD Generator**: Generate Base Models, Custom Models, Resources, Controllers, Routes, and Postman Collections in seconds from DB schema with `padi g <table_name>` or batch all tables with `padi ga`.
- 🔐 **Security Built-In**: JWT Authentication, Bcrypt hashing, Request Rate Limiting, CORS management, and MIME content sniffing.
- 📜 **Daily Rotating Structured Logger**: Dual output (stdout + `storage/logs/app-YYYY-MM-DD.log` & `error-YYYY-MM-DD.log`) with 14-day auto-retention and structured JSON context.
- 🗄️ **Database Agnostic**: Native support for **SQLite (pure Go)**, **MySQL**, and **PostgreSQL** via standard `database/sql` connection pooling.
- 🚀 **Built-in Migrations**: Automatic migration registry and rollback system (`padi migrate`, `padi migrate:rollback`).
- 📦 **L1 & L2 Caching**: In-Memory + Redis / File caching with atomic writes, atomic increment/decrement, `cache.Has()`, `cache.DeleteMany()`, `cache.Cleanup()`, and `cache.Remember()`.
- 📁 **File & Media Storage**: Secure file uploads with extension filtering, MIME content sniffing, automatic URL resolution, and `storage.URLOrNull()` / `file.URLOrNull()` for nullable DB columns.
- 📧 **Email Engine**: SMTP mailer with TLS/SSL support (`email.Send()`).
- 👷 **Async Background Queue & Workers**: Database-backed asynchronous queue and worker runner (`queue.Push()`, `padi queue:work`).
- 📡 **Real-time Pub/Sub**: Native Server-Sent Events (SSE) broadcaster (`realtime.Publish()`, `realtime.SubscribeSSE()`).
- ✅ **Type-Safe Validation**: Field-level request validation with clean JSON error responses.

---

## 🚀 Installation & Quick Start

### 1. Clone Starter Template
```bash
git clone https://github.com/wibiesana/padi_go.git my-api-project
cd my-api-project
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Setup Environment
Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```
*Or run the interactive setup wizard:*
```bash
go run ./cmd/padi init
```

### 4. Run Migrations & Start Server
```bash
# Run database migrations
go run ./cmd/padi migrate

# Start the development server
go run main.go
# or using the CLI:
go run ./cmd/padi serve
```
The API server will be available at **`http://localhost:8080`**.

---

## 📁 Project Structure

```text
padi_go_template/
├── api_collection/             # Generated Postman/Thunder Client collections
├── app/
│   ├── Controllers/            # HTTP Handlers (AuthController, UserController, etc.)
│   ├── Models/                 # Custom domain models (business logic, custom methods)
│   │   └── Base/               # Auto-generated base models (DO NOT EDIT directly)
│   ├── Resources/              # API Resource transformers (JSON response shaping)
│   │   └── Base/               # Auto-generated base resources
│   └── Routes/
│       └── api.go              # API route registry with versioning
├── cmd/
│   └── padi/                   # Padi CLI tool (generate, migrate, queue, wizard)
│       └── main.go
├── database/
│   ├── migrations/             # Database migrations
│   └── database.sqlite         # SQLite default database file
├── storage/                    # Uploaded assets, file cache & daily rotating logs
│   ├── cache/                  # File-based cache storage
│   ├── logs/                   # Daily rotating logs (app-*.log, error-*.log)
│   └── uploads/                # User uploaded files & media
├── .env.example                # Environment variables template
├── .gitignore
├── go.mod                      # Module definition (imports github.com/wibiesana/padi_go_core)
├── go.sum
├── main.go                     # HTTP API Server application entrypoint
└── README.md
```

> 💡 **Framework Core Packages**:  
> All core packages reside in the independent module `github.com/wibiesana/padi_go_core` which includes: `activerecord`, `auth`, `cache`, `config`, `database`, `email`, `file`, `generator`, `logger`, `middleware`, `migrator`, `model`, `query`, `queue`, `realtime`, `response`, `router`, `storage`, `validator`, and `wizard`.

---

## 🛠️ Padi CLI Commands

| Command | Alias | Description |
| :--- | :--- | :--- |
| `go run ./cmd/padi init` | | Run the Interactive Setup Wizard (configures .env & database) |
| `go run ./cmd/padi serve` | | Start the HTTP API server |
| `go run ./cmd/padi migrate` | | Run pending database migrations |
| `go run ./cmd/padi migrate:rollback` | | Rollback last batch of migrations |
| `go run ./cmd/padi generate:crud <table_name>` | `padi g <table_name>` | Auto-generate Base Model, Custom Model, Resource, Controller, and Route snippet |
| `go run ./cmd/padi generate:crud-all` | `padi ga` | Auto-generate CRUD for **ALL** database tables |
| `go run ./cmd/padi queue:work [queue_name]` | `padi queue` | Run the background queue worker |
| `go run ./cmd/padi make:model <name>` | `padi create:model` | Create a new model file |
| `go run ./cmd/padi make:controller <name>` | `padi create:controller` | Create a new controller file |
| `go run ./cmd/padi make:migration <name>` | `padi create:migration` | Create a new migration file |
| `go run ./cmd/padi update` | | Update Padi Core dependencies and modules to latest versions |

---

## 📡 Standard API Response Format

```json
{
  "status": 200,
  "success": true,
  "message": "Operation successful",
  "data": { ... },
  "meta": {
    "total": 100,
    "per_page": 15,
    "current_page": 1,
    "last_page": 7,
    "from": 1,
    "to": 15
  }
}
```
