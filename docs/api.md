# API Documentation

## Authentication Endpoints

### Register User

```http
POST /api/auth/register
```

Register a new user in the system.

**Request Body:**

```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**Validation Rules:**

- `username`: Required, minimum 3 characters
- `email`: Required, valid email format
- `password`: Required, minimum 6 characters

**Response:**

```json
{
  "token": "string",
  "user": {
    "id": "integer",
    "username": "string",
    "email": "string",
    "role": "string",
    "status": "string",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

**Status Codes:**

- `201`: User successfully created
- `400`: Invalid request body
- `409`: Email or username already exists
- `500`: Server error

### Login User

```http
POST /api/auth/login
```

Authenticate an existing user using email or username.

**Request Body:**

```json
{
  "identifier": "string", // email or username
  "password": "string"
}
```

**Validation Rules:**

- `identifier`: Required, minimum 3 characters (can be email or username)
- `password`: Required, minimum 6 characters

**Response:**

```json
{
  "token": "string",
  "user": {
    "id": "integer",
    "username": "string",
    "email": "string",
    "role": "string",
    "status": "string",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

**Status Codes:**

- `200`: Successfully authenticated
- `400`: Invalid request body
- `401`: Invalid credentials
- `403`: Account banned or frozen
- `500`: Server error

### Request Password Reset

```http
POST /api/auth/reset-password
```

Request a password reset link.

**Request Body:**

```json
{
  "email": "string"
}
```

**Validation Rules:**

- `email`: Required, valid email format

**Response:**

```json
{
  "message": "string"
}
```

**Status Codes:**

- `200`: Reset request processed
- `400`: Invalid request body
- `500`: Server error

### Update Password with Reset Token

```http
POST /api/auth/update-password
```

Update password using a reset token.

**Request Body:**

```json
{
  "reset_token": "string",
  "new_password": "string"
}
```

**Validation Rules:**

- `reset_token`: Required
- `new_password`: Required, minimum 6 characters

**Response:**

```json
{
  "message": "string"
}
```

**Status Codes:**

- `200`: Password updated successfully
- `400`: Invalid request body or token
- `404`: User not found
- `500`: Server error

### Change Password (Authenticated)

```http
POST /api/auth/change-password
```

Change password for authenticated user.

**Request Body:**

```json
{
  "old_password": "string",
  "new_password": "string"
}
```

**Validation Rules:**

- `old_password`: Required
- `new_password`: Required, minimum 6 characters

**Response:**

```json
{
  "message": "string"
}
```

**Status Codes:**

- `200`: Password changed successfully
- `400`: Invalid request body
- `401`: Invalid old password
- `500`: Server error

## Authentication

All protected endpoints require a JWT token in the Authorization header:

```http
Authorization: Bearer <token>
```

### User Roles

- `USER`: Basic user privileges
- `EDITOR`: Can edit and moderate content
- `ADMIN`: Administrative privileges
- `SUPER_ADMIN`: Full system access

### User Status

- `active`: Account is active and can be used
- `passive`: Account is temporarily inactive
- `banned`: Account is permanently banned
- `frozen`: Account is temporarily frozen

## Error Responses

All error responses follow this format:

```json
{
  "error": "string"
}
```

Common error status codes:

- `400`: Bad Request - Invalid input
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Insufficient permissions
- `404`: Not Found - Resource doesn't exist
- `409`: Conflict - Resource already exists
- `500`: Internal Server Error - Server-side error
