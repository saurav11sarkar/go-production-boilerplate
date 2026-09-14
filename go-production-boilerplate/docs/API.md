# API Reference

Base: `http://localhost:8080/api/v1`

## Public auth

| Method | Route | Purpose |
|---|---|---|
| POST | `/auth/register` | User registration |
| POST | `/auth/login` | Login + access token + refresh cookie |
| POST | `/auth/refresh` | Rotate refresh token |
| POST | `/auth/logout` | Revoke current refresh token |
| POST | `/auth/verify-email` | Verify email |
| POST | `/auth/resend-verification` | Send verification again |
| POST | `/auth/forgot-password` | Request reset link |
| POST | `/auth/reset-password` | Reset password |

## Authenticated user

Bearer access token required.

| Method | Route | Purpose |
|---|---|---|
| POST | `/auth/logout-all` | Revoke all refresh sessions |
| GET | `/users/me` | Current profile |
| PATCH | `/users/me` | Update name |
| PATCH | `/users/me/password` | Change password |
| POST | `/users/me/avatar` | Multipart `image` upload |
| DELETE | `/users/me/avatar` | Remove avatar |

## Admin

Bearer token with role `admin` required.

| Method | Route | Purpose |
|---|---|---|
| GET | `/admin/users` | list/search/filter/sort/paginate |
| GET | `/admin/users/{id}` | user details |
| PATCH | `/admin/users/{id}/status` | activate/deactivate |
| PATCH | `/admin/users/{id}/role` | user/admin role |
| DELETE | `/admin/users/{id}` | delete user |

### Admin query example

```text
GET /api/v1/admin/users?page=1&limit=20&search=saurav&role=user&isVerified=true&isActive=true&sortBy=createdAt&sortOrder=desc
```

Allowed sort fields: `createdAt`, `updatedAt`, `name`, `email`, `role`.


## Verify email body

```json
{
  "token": "TOKEN_FROM_EMAIL_LINK"
}
```

## Register body

```json
{
  "name": "Saurav Sarkar",
  "email": "saurav@example.com",
  "password": "StrongPass123!"
}
```

## Login body

```json
{
  "email": "saurav@example.com",
  "password": "StrongPass123!"
}
```

## Standard success

```json
{
  "success": true,
  "message": "Profile retrieved",
  "data": {}
}
```

## Standard validation error

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "email": "A valid email is required"
    }
  }
}
```
