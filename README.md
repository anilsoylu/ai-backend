# AI Backend

A robust backend service built with Go, using Gin framework and GORM for database operations.

## Technologies Used

- Go 1.21+
- Gin Web Framework
- GORM (with PostgreSQL)
- Air (for live reload during development)
- Better-auth for authentication

## Prerequisites

- Go 1.21 or higher
- PostgreSQL
- Air (optional, for development)

## Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/ai-backend.git
cd ai-backend
```

2. Install dependencies:

```bash
go mod download
```

3. Set up your environment variables:
   Create a `.env` file in the root directory with the following content:

```env
DATABASE_URL="postgresql://username:password@localhost:5432/dbname"
PORT=8080
```

4. Run the application:

For development (with Air):

```bash
air
```

For production:

```bash
go run cmd/api/main.go
```

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go
├── config/
├── internal/
│   ├── database/
│   ├── handlers/
│   ├── middleware/
│   └── models/
├── pkg/
│   └── utils/
└── tests/
```

## API Documentation

API documentation can be found in the `docs/api` directory.

## Testing

To run tests:

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License.
