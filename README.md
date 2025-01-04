# Messenger Backend API

![Go](https://img.shields.io/badge/Go-1.23-blue)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)
![Redis](https://img.shields.io/badge/Redis-7-darkred)
![License](https://img.shields.io/badge/License-MIT-green)
![Build](https://img.shields.io/badge/Build-Passing-brightgreen)

## 📖 Overview

The **Messenger Backend API** is a robust and scalable backend service built in Go, designed for managing users, messages, and authentication for a messaging application. It uses PostgreSQL as the database and provides a RESTful API for seamless integration with frontend applications.

---

## 🚀 Features

- **User Management**: Register, login, and manage user accounts.
- **Message Handling**: Send, receive, and store messages.
- **Authentication**: Secure authentication using JWT.
- **Localization**: Support for multiple languages (English, Ukrainian, Polish).
- **Validation**: Field-specific error handling with detailed feedback.

---

## 🛠️ Tech Stack

- **Language**: Go
- **Database**: PostgreSQL (16 or higher), Redis (7 or higher)
- **Framework**: Gorilla Mux
- **Tools**: Makefile, `go-playground/validator`, `go-i18n`, `golang-migrate`
- **Testing**: `testing` package (built-in)

---

## 📦 Installation

### Prerequisites

1. [Go](https://go.dev/dl/) 1.23 or higher installed.
2. [PostgreSQL](https://www.postgresql.org/) database installed.
3. [Redis](https://redis.io/) database installed.
4. [golang-migrate](https://github.com/golang-migrate/migrate) installed globally for managing database migrations.

### Clone the Repository

```bash
  git clone https://github.com/drTragger/messenger-backend.git
  cd messenger-backend
```

### Setup Environment Variables

1. Copy the example environment file:

    ```bash
    cp .env.example .env
    ```

2. Update the .env file with your PostgreSQL credentials and JWT secret.

3. Generate JWT Secret Automatically:

    Use the provided script to generate a secure JWT secret:

    * Using Makefile:
   
    ```bash
    make generate-jwt
    ```
   
    * Without Makefile:

    ```bash
    ./cmd/generate-secret
    ```

### Install Dependencies

Install Go modules required for the project:

```bash
  go mod download
```

### Run Database Migrations

Apply the database migrations using the Makefile:

```bash
  make migrate-up
```

Alternatively, run migrations directly:

```bash
  migrate -path db/migrations -database "postgres://your_db_user:your_db_password@localhost:5432/messenger?sslmode=disable" up
```

### Start the Application

Using Makefile:

```bash
  make run
```

Without Makefile:

```bash
  go run cmd/main.go
```

The server will be available at http://localhost:8080.

## 🗄️ Project Structure

```plaintext
messenger-backend/
│
├── cmd/                  # Entry points for the application
│   └── main.go           # Main entry point for the API
│
├── config/               # Configuration management
│   └── config.go         # App configuration logic
│
├── db/                   # Database-related files
│   └── migrations/       # Migration files
│
├── internal/             # Core application logic
│   ├── handlers/         # Request handlers
│   ├── middleware/       # Middleware functions
│   ├── models/           # Data models
│   ├── routes/           # Route definitions
│   └── utils/            # Utility functions
│
├── locales/              # Localization files
│
├── Makefile              # Automation tasks
├── Dockerfile            # Docker configuration
├── docker-compose.yml    # Docker Compose configuration
├── go.mod                # Go module file
├── go.sum                # Go module dependencies
├── .env.example          # Example environment variables
└── README.md             # Documentation
```

## 📧 Contact

If you have any questions or issues, feel free to contact me at [my email](mailto:mishaponomarenko11082001@gmail.com).