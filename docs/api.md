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

### Update User Profile

```http
PUT /api/users/profile
```

Update authenticated user's profile information. Users can only update their own profile and cannot modify their role.

**Request Body:**

```json
{
  "username": "string",
  "email": "string",
  "full_name": "string",
  "bio": "string",
  "avatar_url": "string"
}
```

**Validation Rules:**

- `username`: Optional, minimum 3 characters
- `email`: Optional, valid email format
- `full_name`: Optional, maximum 100 characters
- `bio`: Optional, maximum 500 characters
- `avatar_url`: Optional, valid URL format

**Response:**

```json
{
  "user": {
    "id": "integer",
    "username": "string",
    "email": "string",
    "full_name": "string",
    "bio": "string",
    "avatar_url": "string",
    "role": "string",
    "status": "string",
    "updated_at": "timestamp"
  }
}
```

**Status Codes:**

- `200`: Profile updated successfully
- `400`: Invalid request body
- `401`: Unauthorized - Authentication required
- `409`: Email or username already exists
- `500`: Server error

**Notes:**

- Users cannot modify their role through this endpoint
- Email changes require unique validation
- Username changes require unique validation

### Delete Account

```http
DELETE /api/users/account
```

Soft delete the authenticated user's account. This action is reversible by an administrator.

**Request Body:**

```json
{
  "password": "string"
}
```

**Validation Rules:**

- `password`: Required, user's current password for confirmation

**Response:**

```json
{
  "message": "Account deleted successfully"
}
```

**Status Codes:**

- `200`: Account deleted successfully
- `400`: Invalid request body
- `401`: Unauthorized or invalid password
- `500`: Server error

**Notes:**

- This is a soft delete operation
- Account can be restored by an administrator
- All associated data will be preserved but hidden
- User sessions will be invalidated

### Freeze Account

```http
POST /api/users/freeze
```

Temporarily freeze the authenticated user's account for a specified duration.

**Request Body:**

```json
{
  "duration": "integer", // Number of days to freeze the account
  "reason": "string" // Reason for freezing the account
}
```

**Validation Rules:**

- `duration`: Required, minimum 1 day, maximum 365 days
- `reason`: Required, maximum 500 characters

**Response:**

```json
{
  "message": "Account frozen successfully",
  "freeze_details": {
    "id": "integer",
    "user_id": "integer",
    "reason": "string",
    "duration": "integer",
    "start_date": "timestamp",
    "end_date": "timestamp",
    "is_active": "boolean"
  }
}
```

**Status Codes:**

- `200`: Account frozen successfully
- `400`: Invalid request body
- `401`: Unauthorized
- `500`: Server error

### Get Freeze History

```http
GET /api/users/freeze/history
```

Get the freeze history of the authenticated user's account.

**Response:**

```json
{
  "freeze_history": [
    {
      "id": "integer",
      "reason": "string",
      "duration": "integer",
      "start_date": "timestamp",
      "end_date": "timestamp",
      "is_active": "boolean",
      "unfrozen_at": "timestamp"
    }
  ]
}
```

**Status Codes:**

- `200`: History retrieved successfully
- `401`: Unauthorized
- `500`: Server error

### List Users

```http
GET /api/users
```

Get a paginated list of users.

**Query Parameters:**

```
page: integer (default: 1) - Page number
limit: integer (default: 10, max: 50) - Number of users per page
search: string (optional) - Search by username or email
role: string (optional) - Filter by user role
status: string (optional) - Filter by user status
sort: string (optional) - Sort field (created_at, username, email)
order: string (optional) - Sort order (asc, desc)
```

**Response:**

```json
{
  "users": [
    {
      "id": "integer",
      "username": "string",
      "email": "string",
      "full_name": "string",
      "bio": "string",
      "avatar_url": "string",
      "role": "string",
      "status": "string",
      "created_at": "timestamp"
    }
  ],
  "pagination": {
    "current_page": "integer",
    "total_pages": "integer",
    "total_items": "integer",
    "has_next": "boolean",
    "has_prev": "boolean"
  }
}
```

**Status Codes:**

- `200`: Users retrieved successfully
- `400`: Invalid query parameters
- `401`: Unauthorized
- `500`: Server error

**Notes:**

- Response is paginated
- Soft deleted users are excluded
- Results are cached for performance
- Search is case-insensitive

### Update User Role

```http
PUT /api/admin/users/role
```

Update a user's role. Only ADMIN and SUPER_ADMIN users can access this endpoint.

**Request Body:**

```json
{
  "user_id": "integer",
  "role": "string",
  "reason": "string" // Required for ADMIN users, minimum 15 characters
}
```

**Validation Rules:**

- `user_id`: Required
- `role`: Required, must be one of: "USER", "EDITOR", "ADMIN", "SUPER_ADMIN"
- `reason`: Required for ADMIN users when changing roles, minimum 15 characters

**Response:**

```json
{
  "message": "string",
  "user": {
    "id": "integer",
    "username": "string",
    "role": "string",
    "updated_at": "timestamp"
  }
}
```

**Status Codes:**

- `200`: Role updated successfully
- `400`: Invalid request body
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Insufficient permissions
- `404`: User not found
- `500`: Server error

**Authorization Rules:**

- Only ADMIN and SUPER_ADMIN users can access this endpoint
- First SUPER_ADMIN's role cannot be changed
- Only first SUPER_ADMIN can grant SUPER_ADMIN role to others
- SUPER_ADMIN can assign any role (USER, EDITOR, ADMIN, SUPER_ADMIN)
- ADMIN can only modify between USER and EDITOR roles
- ADMIN cannot modify SUPER_ADMIN or other ADMIN roles
- ADMIN must provide a reason (minimum 15 characters) when changing roles

### Get User Role History

```http
GET /api/admin/users/:user_id/role-history
```

Get the role change history for a specific user.

**Parameters:**

- `user_id`: User ID (path parameter)

**Response:**

```json
{
  "user": {
    "id": "integer",
    "username": "string",
    "role": "string"
  },
  "histories": [
    {
      "id": "integer",
      "user_id": "integer",
      "username": "string",
      "changed_by_id": "integer",
      "changed_by": "string",
      "old_role": "string",
      "new_role": "string",
      "reason": "string",
      "created_at": "timestamp"
    }
  ]
}
```

**Status Codes:**

- `200`: History retrieved successfully
- `400`: Invalid user ID
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Insufficient permissions
- `404`: User not found
- `500`: Server error

### Get All Role Histories

```http
GET /api/admin/role-histories
```

Get all role change histories with pagination.

**Query Parameters:**

- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 50)

**Response:**

```json
{
  "histories": [
    {
      "id": "integer",
      "user_id": "integer",
      "username": "string",
      "changed_by_id": "integer",
      "changed_by": "string",
      "old_role": "string",
      "new_role": "string",
      "reason": "string",
      "created_at": "timestamp"
    }
  ],
  "pagination": {
    "current_page": "integer",
    "total_pages": "integer",
    "total_items": "integer",
    "per_page": "integer",
    "has_next": "boolean",
    "has_prev": "boolean"
  }
}
```

**Status Codes:**

- `200`: Histories retrieved successfully
- `401`: Unauthorized - Authentication required
- `403`: Forbidden - Insufficient permissions
- `500`: Server error

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
