# KenTech Backend Assessment

A robust backend API built with Go following hexagonal architecture principles.

---

## Local Tools & API Client Collections

A `local-tools` folder is provided with useful resources for testing this API:
- **Bruno API Client Collection**: Inside `local-tools/kentech-challenge`, you’ll find a collection for [Bruno](https://www.usebruno.com/). You can use this to quickly test all endpoints.
- **Postman/Insomnia Export**: The Bruno collection is also exported as a JSON file for use with Postman or Insomnia.

## How to Test


1. Start all services:
   ```bash
   docker-compose up -d --build
   ```
   This will start:
   - PostgreSQL database on port 5432
   - The API server on port 8080
   - Mock wallet service on port 9090
   - Jaeger tracing on port 16686

2. Use Bruno, Postman, or Insomnia to import the provided collection and test endpoints.

3. Typical flow:
   - **Register** a new user (`POST /api/auth/register`)
   - **Login** (`POST /api/auth/login`) and copy the JWT token from the response
   - For transaction and informational endpoints, **add the JWT token** to the `Authorization: Bearer <token>` header
   - Test deposit, withdraw, and cancel endpoints
   - Use informational endpoints to view user profile, balance, and transaction history

---

## Features

- **User Authentication**: JWT-based authentication with secure password hashing
- **Player Management**: User profiles and balance tracking
- **Transaction System**: Deposit, withdraw, and cancel operations
- **Wallet Integration**: Mock wallet service integration
- **Database**: PostgreSQL with proper indexing
- **Containerization**: Docker and Docker Compose setup

## Architecture

This project follows hexagonal architecture (ports and adapters):

```
├── cmd/
│   └── main.go                 // Application entry point
├── internal/
│   ├── adapters/
│   │   ├── http/               // HTTP handlers and server
│   │   ├── repository/         // Data persistence adapters
│   │   └── auth/              // JWT authentication
│   └── core/
│       ├── domain/            // Business entities and logic
│       ├── port/              // Interfaces
├── pkg/
│   ├── config/                // Configuration
│   ├── database/              // Database connection
│   ├── logger/                // Logging utility
│   └── security/              // Password hashing
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### Player Management
- `GET /api/player/profile` - Get user profile
- `GET /api/player/balance` - Get current balance
- `GET /api/player/transactions` - Get transaction history

### Transactions
- `POST /api/transactions/deposit` - Make a deposit
- `POST /api/transactions/withdraw` - Make a withdrawal
- `POST /api/transactions/{id}/cancel` - Cancel a transaction

### Health Check
- `GET /health` - Service health check

## Getting Started

### Prerequisites
- Go 1.23+
- Docker and Docker Compose

## Environment Variables

- `PORT` - Server port (default: 8080)
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - JWT signing secret
- `WALLET_URL` - Mock wallet service URL
- `LOG_LEVEL` - Logging level (default: info)
- `WALLET_API_KEY` - Api key for wallet service authentication`

## Testing the API

### Register a user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "password123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

### Make a deposit (requires JWT token)
```bash
curl -X POST http://localhost:8080/api/transactions/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"amount": 100.00}'
```

### Check balance
```bash
curl -X GET http://localhost:8080/api/player/balance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `wallet_user_id` (INT, Unique)
- `username` (VARCHAR, Unique)
- `email` (VARCHAR, Unique)
- `password` (VARCHAR, Hashed)
- `balance` (DECIMAL)
- `created_at`, `updated_at` (TIMESTAMP)

### Transactions Table
- `id` (UUID, Primary Key)
- `user_id` (UUID, Foreign Key)
- `type` (VARCHAR: deposit/withdraw)
- `amount` (DECIMAL)
- `status` (VARCHAR: pending/completed/canceled/failed)
- `reference` (VARCHAR)
- `created_at`, `updated_at` (TIMESTAMP)

## Security Features

- JWT-based authentication
- Bcrypt password hashing
- CORS enabled
- Input validation
- SQL injection prevention
- Authorization middleware

## Design Decisions

1. **Hexagonal Architecture**: Ensures clean separation of concerns and testability
2. **Repository Pattern**: Abstracts data access layer
3. **JWT Authentication**: Stateless authentication suitable for REST APIs
4. **Transaction Safety**: Database transactions for balance updates
5. **Error Handling**: Comprehensive error handling with appropriate HTTP status codes
6. **Validation**: Input validation for all endpoints

## Future Enhancements

- Add unit and integration tests
- Implement rate limiting
- Performance monitoring

## Disclaimer
For personal issues, not everything was implemented as I want, I tried my best to implement the core functionality. The code is structured to allow easy addition of features and improvements in the future. Tracer was not fully implemented due to time constraints, and i didn't stressed all the edge scenarios so the code may have some bugs.
