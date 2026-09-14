# Authentication Flow

## Register

```text
POST /api/v1/auth/register
 -> validate
 -> normalize email
 -> check duplicate
 -> bcrypt password
 -> create user
 -> random email verification token
 -> store SHA-256 token hash
 -> send/log frontend verification URL
```

Raw verification token database-এ রাখা হয় না।

## Login

```text
POST /api/v1/auth/login
 -> find user
 -> bcrypt compare
 -> active check
 -> verified check
 -> create short-lived JWT access token
 -> create random refresh token
 -> store SHA-256 refresh hash + current user token_version
 -> set raw refresh token as HttpOnly cookie
```

Access token response JSON-এ থাকে। Refresh token browser JavaScript থেকে পড়ার দরকার নেই।

## Refresh token rotation

```text
old refresh cookie
 -> hash
 -> SELECT ... FOR UPDATE
 -> lock refresh-token row + user row
 -> expiry/revoked/token_version check
 -> revoke old token
 -> insert new refresh hash bound to the same current token_version
 -> commit transaction
 -> new access token + new refresh cookie
```

একই old refresh token concurrentভাবে দুইবার rotate হতে না দেওয়ার জন্য row lock + transaction আছে।

## Logout

Current refresh hash revoke হয় এবং cookie clear হয়।

## Logout all

User-এর সব active refresh token revoke হয়।

## Forgot password

Response কখনও বলে না email database-এ আছে কি না। এতে account enumeration কমে। Password reset token random, database-এ hash, expiry 20 minutes।

## Change/reset password

Password পরিবর্তনের পরে existing refresh sessions revoke করা হয়।

## Browser security

Production `.env`:

```env
APP_ENV=production
REFRESH_COOKIE_SECURE=true
```

HTTPS ছাড়া production auth deploy করবেন না।


## Why verification uses POST

The email link points to the frontend. The frontend sends the token to `POST /api/v1/auth/verify-email`. This avoids state-changing GET links being consumed by automated email security scanners.
