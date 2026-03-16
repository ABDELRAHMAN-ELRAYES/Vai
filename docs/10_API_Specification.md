# API Specification
## Vai — REST API Reference

**Version:** 1.0  
**Base URL:** `http://localhost:8080`  
**Auth:** HTTP-only cookie (`access_token`) set on login  
**Content-Type:** `application/json` (unless noted)

---

## Table of Contents

1. [Authentication](#authentication)
2. [Error Format](#error-format)
3. [Auth Endpoints](#auth-endpoints)
4. [User Endpoints](#user-endpoints)
5. [Document Endpoints](#document-endpoints)
6. [Chat Endpoints](#chat-endpoints)
7. [Search Endpoint](#search-endpoint)
8. [OpenAPI Summary](#openapi-summary)

---

## Authentication

All endpoints except `/auth/*` require a valid JWT in the `access_token` HTTP-only cookie.

The cookie is set automatically on login or OAuth callback. When the access token expires (15 min), call `POST /auth/refresh` with the `refresh_token` cookie to receive a new token pair.

```
Cookie: access_token=<JWT>; HttpOnly; Secure; SameSite=Strict
Cookie: refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Path=/auth/refresh
```

---

## Error Format

All errors return a consistent JSON envelope:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable description",
    "request_id": "req_01HXYZ..."
  }
}
```

### HTTP Status Codes

| Status | Code | When |
|--------|------|------|
| `400` | `BAD_REQUEST` | Malformed request body or missing required fields |
| `401` | `UNAUTHORIZED` | Missing, expired, or invalid JWT |
| `403` | `FORBIDDEN` | Authenticated but not authorized for this resource |
| `403` | `EMAIL_NOT_VERIFIED` | Endpoint requires verified email |
| `404` | `NOT_FOUND` | Resource not found or not owned by user |
| `409` | `EMAIL_EXISTS` | Email already registered |
| `422` | `VALIDATION_ERROR` | Input validation failed (includes field details) |
| `429` | `RATE_LIMITED` | Too many requests (auth endpoints: 20 req/min per IP) |
| `500` | `INTERNAL_ERROR` | Unexpected server error |

### Validation Error Example

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "email": "must be a valid email address",
      "password": "must be at least 8 characters"
    }
  }
}
```

---

## Auth Endpoints

### `POST /auth/register`

Register a new user account. Sends a verification email.

**Request**
```json
{
  "email": "alice@example.com",
  "password": "SecurePass123!",
  "display_name": "Alice"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|-----------|
| `email` | string | ✅ | Valid email format, max 320 chars |
| `password` | string | ✅ | Min 8 chars, at least 1 uppercase, 1 digit |
| `display_name` | string | ✅ | 2–100 chars |

**Response `201 Created`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "display_name": "Alice",
  "is_verified": false,
  "created_at": "2025-06-01T10:00:00Z"
}
```

**Errors:** `409` email already exists · `422` validation error

---

### `POST /auth/login`

Authenticate with email and password. Sets JWT + refresh token cookies.

**Request**
```json
{
  "email": "alice@example.com",
  "password": "SecurePass123!"
}
```

**Response `200 OK`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "display_name": "Alice",
  "is_verified": true
}
```

**Response Headers**
```
Set-Cookie: access_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Max-Age=900; Path=/
Set-Cookie: refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Max-Age=604800; Path=/auth/refresh
```

**Errors:** `401` invalid credentials

---

### `POST /auth/logout`

Revoke the current refresh token and clear both cookies.

**Response `204 No Content`**

**Response Headers**
```
Set-Cookie: access_token=; Max-Age=0; HttpOnly
Set-Cookie: refresh_token=; Max-Age=0; HttpOnly
```

---

### `POST /auth/refresh`

Issue a new JWT using the refresh token. Rotates the refresh token.

> Reads the `refresh_token` cookie automatically — no request body needed.

**Response `200 OK`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "display_name": "Alice",
  "is_verified": true
}
```

**Response Headers:** New `Set-Cookie` headers with rotated tokens.

**Errors:** `401` refresh token invalid, expired, or revoked

---

### `GET /auth/verify`

Verify a user's email address using the token from the verification email.

**Query Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | ✅ | HMAC-SHA256 token from email |

**Response `200 OK`**
```json
{
  "message": "Email verified successfully"
}
```

**Errors:** `401` token invalid, expired, or already used

---

### `POST /auth/forgot-password`

Request a password reset email. Always returns `202` to prevent email enumeration.

**Request**
```json
{
  "email": "alice@example.com"
}
```

**Response `202 Accepted`**
```json
{
  "message": "If that email is registered, a reset link has been sent"
}
```

---

### `POST /auth/reset-password`

Reset the user's password using the token from the reset email. Revokes all existing sessions.

**Request**
```json
{
  "token": "<reset_token_from_email>",
  "new_password": "NewSecurePass456!"
}
```

**Response `200 OK`**
```json
{
  "message": "Password reset successfully. Please log in again."
}
```

**Errors:** `401` token invalid/expired/used · `422` password too weak

---

### `GET /auth/google`

Initiate Google OAuth 2.0 flow. Redirects to Google consent screen.

**Response `302 Found`**
```
Location: https://accounts.google.com/o/oauth2/auth?client_id=...&redirect_uri=...&state=...&scope=openid+email+profile
```

---

### `GET /auth/google/callback`

OAuth 2.0 callback. Validates state, exchanges code, upserts user, sets cookies.

**Query Parameters:** `code` (string), `state` (string) — set automatically by Google.

**Response `302 Found`**
```
Location: /
Set-Cookie: access_token=...; Set-Cookie: refresh_token=...
```

**Errors:** `400` state mismatch (CSRF) · `502` Google token exchange failed

---

## User Endpoints

> All require authentication (`access_token` cookie).

### `GET /users/me`

Get the authenticated user's profile.

**Response `200 OK`**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "alice@example.com",
  "display_name": "Alice",
  "avatar_url": "https://lh3.googleusercontent.com/...",
  "is_verified": true,
  "created_at": "2025-06-01T10:00:00Z",
  "updated_at": "2025-06-01T10:00:00Z"
}
```

---

### `PATCH /users/me`

Update the authenticated user's profile. All fields optional.

**Request**
```json
{
  "display_name": "Alice Smith",
  "avatar_url": "https://example.com/avatar.png"
}
```

**Response `200 OK`** — Updated user object (same shape as `GET /users/me`).

---

### `DELETE /users/me`

Permanently delete the authenticated user account. Cascades to all documents (Qdrant collection deleted), chat sessions, messages, and tokens.

**Response `204 No Content`** — Cookies cleared.

---

## Document Endpoints

> All require authentication. Document upload additionally requires email verification.

### `POST /documents/upload`

Upload a document and run it through the RAG ingestion pipeline.

**Request** `Content-Type: multipart/form-data`
```
file=@/path/to/document.txt
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `file` | file | ✅ | Plain text (.txt). Max size: 10MB |

**Response `201 Created`**
```json
{
  "document_id": "my-document",
  "chunks": 14,
  "source": "my-document.txt"
}
```

**Errors:** `403` email not verified · `413` file too large · `422` unsupported file type

---

### `GET /documents`

List all documents owned by the authenticated user.

**Response `200 OK`**
```json
{
  "documents": [
    {
      "id": "my-document",
      "source": "my-document.txt",
      "chunk_count": 14,
      "size_bytes": 8192,
      "created_at": "2025-06-01T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### `GET /documents/{id}`

Get metadata for a specific document.

**Path Parameters:** `id` — document ID (slug)

**Response `200 OK`** — Single document object (same shape as list item).

**Errors:** `404` not found or not owned by user

---

### `DELETE /documents/{id}`

Delete a document and all its Qdrant vectors.

**Response `204 No Content`**

**Errors:** `404` not found or not owned by user

---

## Chat Endpoints

> All require authentication.

### `POST /chat/sessions`

Create a new chat session. Optionally scoped to a specific document.

**Request**
```json
{
  "title": "Research on Auth",
  "document_id": "my-document"
}
```

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `title` | string | ✅ | Max 255 chars |
| `document_id` | string | ❌ | Scopes all queries to one document |

**Response `201 Created`**
```json
{
  "id": "abc123-...",
  "user_id": "550e8400-...",
  "title": "Research on Auth",
  "document_id": "my-document",
  "created_at": "2025-06-01T10:00:00Z",
  "updated_at": "2025-06-01T10:00:00Z"
}
```

---

### `GET /chat/sessions`

List all chat sessions for the authenticated user, ordered by most recently updated.

**Response `200 OK`**
```json
{
  "sessions": [
    {
      "id": "abc123-...",
      "title": "Research on Auth",
      "document_id": "my-document",
      "updated_at": "2025-06-01T11:00:00Z"
    }
  ]
}
```

---

### `GET /chat/sessions/{id}/messages`

Get the full message history for a session.

**Response `200 OK`**
```json
{
  "session_id": "abc123-...",
  "messages": [
    {
      "id": "msg-uuid-1",
      "role": "user",
      "content": "How does authentication work?",
      "created_at": "2025-06-01T10:05:00Z"
    },
    {
      "id": "msg-uuid-2",
      "role": "assistant",
      "content": "Based on the documents, authentication works by...",
      "tokens_used": 142,
      "created_at": "2025-06-01T10:05:04Z"
    }
  ]
}
```

**Errors:** `403` session not owned by user · `404` session not found

---

### `DELETE /chat/sessions/{id}`

Delete a chat session and all its messages.

**Response `204 No Content`**

---

### `POST /chat`

Ask a question and receive a complete (non-streaming) answer.

**Request**
```json
{
  "question": "How does authentication work?",
  "top_k": 5,
  "document_id": "my-document",
  "session_id": "abc123-..."
}
```

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `question` | string | ✅ | — | Max 2000 chars |
| `top_k` | int | ❌ | `5` | 1–20. Chunks to retrieve |
| `document_id` | string | ❌ | — | Filter to one document |
| `session_id` | string | ❌ | — | Auto-created if omitted |

**Response `200 OK`**
```json
{
  "answer": "Based on the documents, authentication works by issuing a JWT...",
  "session_id": "abc123-...",
  "message_id": "msg-uuid-2",
  "chunks_used": 5
}
```

---

### `GET /chat/stream`

Ask a question and receive a streaming response via Server-Sent Events.

**Query Parameters**

| Param | Type | Required | Default |
|-------|------|----------|---------|
| `question` | string | ✅ | — |
| `top_k` | int | ❌ | `5` |
| `document_id` | string | ❌ | — |
| `session_id` | string | ❌ | — |

**Example**
```bash
curl -N "http://localhost:8080/chat/stream?question=How+does+auth+work&top_k=5" \
  --cookie "access_token=<jwt>"
```

**Response** `Content-Type: text/event-stream`
```
data: Based

data:  on

data:  the

data:  documents,

data:  authentication

data:  works

data:  by...

data: [DONE]
```

---

## Search Endpoint

### `POST /search`

Debug endpoint. Returns raw retrieved chunks without LLM generation. Useful for tuning retrieval quality.

**Request**
```json
{
  "query": "authentication flow",
  "top_k": 3,
  "document_id": "my-document"
}
```

**Response `200 OK`**
```json
{
  "results": [
    {
      "document_id": "my-document",
      "source": "my-document.txt",
      "chunk_index": 3,
      "text": "...the authentication module validates the JWT by...",
      "score": 0.9432
    },
    {
      "document_id": "my-document",
      "source": "my-document.txt",
      "chunk_index": 7,
      "text": "...refresh tokens are stored hashed in PostgreSQL...",
      "score": 0.8971
    }
  ],
  "query_vector_size": 768
}
```

---

## OpenAPI Summary

```yaml
openapi: 3.0.3
info:
  title: Vai API
  version: 1.0.0
  description: Privacy-first AI document assistant REST API

servers:
  - url: http://localhost:8080
    description: Local development

components:
  securitySchemes:
    cookieAuth:
      type: apiKey
      in: cookie
      name: access_token

security:
  - cookieAuth: []

paths:
  /auth/register:     { post: { ... } }
  /auth/login:        { post: { ... } }
  /auth/logout:       { post: { ... } }
  /auth/refresh:      { post: { ... } }
  /auth/verify:       { get: { ... } }
  /auth/forgot-password: { post: { ... } }
  /auth/reset-password:  { post: { ... } }
  /auth/google:         { get: { ... } }
  /auth/google/callback: { get: { ... } }
  /users/me:          { get: { ... }, patch: { ... }, delete: { ... } }
  /documents/upload:  { post: { ... } }
  /documents:         { get: { ... } }
  /documents/{id}:    { get: { ... }, delete: { ... } }
  /chat/sessions:     { post: { ... }, get: { ... } }
  /chat/sessions/{id}/messages: { get: { ... } }
  /chat/sessions/{id}: { delete: { ... } }
  /chat:              { post: { ... } }
  /chat/stream:       { get: { ... } }
  /search:            { post: { ... } }
```
