# Task Management API

A robust RESTful API for managing tasks built with Go, Gin, GORM, PostgreSQL, and Redis. This project includes comprehensive features like authentication, rate limiting, and logging.

## Features

### Core Functionality
- Full CRUD operations for tasks
- User authentication with JWT tokens
- Task filtering and pagination
- Input validation and error handling
- Database migrations


## Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Authentication**: JWT tokens
- **Logging**: Logrus (structured logging)
- **Testing**: Testify
- **Containerization**: Docker & Docker Compose
- **Migration**: golang-migrate

## API Endpoints

### Authentication
```
POST /api/v1/auth/register - Register a new user
POST /api/v1/auth/login    - Login user
```

### Tasks (Protected Routes)
```
POST   /api/v1/tasks     - Create a new task
GET    /api/v1/tasks     - Get all tasks (with filtering and pagination)
GET    /api/v1/tasks/:id - Get a specific task
PATCH  /api/v1/tasks/:id - Update a task
DELETE /api/v1/tasks/:id - Delete a task
```

### Utility
```
GET /health - Health check endpoint
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24 (for local development)

### Using Docker (Recommended)

1. **Clone the repository**
```bash
git clone <repository-url>
cd go-corenglish
```

2. **Start the services**
```bash
docker-compose up -d
```