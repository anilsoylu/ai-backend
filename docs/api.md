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

### Login

```http
POST /api/auth/login
```

Authenticate a user and receive a JWT token.

**Request Body:**

```json
{
  "identifier": "string", // email or username
  "password": "string"
}
```

**Response:**

```json
{
  "token": "string",
  "user": {
    "id": "integer",
    "username": "string",
    "email": "string",
    "role": "string",
    "status": "string"
  }
}
```

**Status Codes:**

- `200`: Login successful
- `400`: Invalid request body
- `401`: Invalid credentials
- `403`: Account is banned, frozen, or passive
- `500`: Server error

**Notes:**

- Users with banned, frozen, or passive status cannot log in
- Banned users will receive a "Account is banned" message
- Frozen users will receive a "Account is frozen" message
- Passive users will receive a "Account is passive. Please contact support to reactivate your account." message

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

## User Management Endpoints

### Update User Status

```http
PUT /api/users/status
```

Update a user's status. Regular users can update their own status to active, passive, or frozen. Only ADMIN and SUPER_ADMIN can set banned status.

**Request Body:**

```json
{
  "user_id": "integer",
  "status": "string"
}
```

**Validation Rules:**

- `user_id`: Required
- `status`: Required, must be one of:
  - For regular users: "active", "passive", "frozen" (can only update their own status)
  - For admins: "active", "passive", "frozen", "banned" (can update any user's status)

**Response:**

```json
{
  "message": "string",
  "user": {
    "id": "integer",
    "status": "string"
  }
}
```

**Status Codes:**

- `200`: Status updated successfully
- `400`: Invalid request body or status
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Cannot update other users' status or use banned status
- `404`: User not found
- `500`: Server error

**Authorization:**

- Requires valid JWT token
- Regular users can only update their own status
- Regular users cannot set banned status
- ADMIN and SUPER_ADMIN can update any user's status
- ADMIN cannot modify SUPER_ADMIN users

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
