# API Specification — Vai
**Version:** 1.0  
**Date:** April 2026  

## Vai — REST API Reference

**Version:** 1.0  
**Base URL:** `http://localhost:3000/api/v1`  
**Auth:** HTTP-only cookie (`access_token`) set on login  
**Content-Type:** `application/json` (unless noted)

---

## Table of Contents

1. [Authentication](#authentication)
2. [Error Format](#error-format)
3. [Auth Endpoints](#auth-endpoints)
4. [Document Endpoints](#document-endpoints)
5. [Conversation Endpoints](#conversation-endpoints)
6. [OpenAPI Summary](#openapi-summary)

---

## Authentication

All endpoints except `/auth/register` and `/auth/login` require a valid JWT in the `access_token` HTTP-only cookie.

The cookie is set automatically on login or OAuth callback with a 90-day expiry. There are no refresh tokens — the session remains valid as long as the cookie persists.

```
Cookie: access_token=<JWT>; HttpOnly; SameSite=Lax
```

---

## Error Format

All errors return a consistent JSON envelope:

```json
{
  "error": "Human-readable description of the error"
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
---

## Auth Endpoints

### `POST /auth/register`

Register a new user account. Sends a verification email.

**Request**
```json
{
  "first_name": "Alice",
  "last_name": "Smith",
  "email": "alice@example.com",
  "password": "SecurePass123!"
}
```

| Field | Type | Required | Validation |
|-------|------|----------|-----------|
| `first_name` | string | ✅ | Max 255 chars |
| `last_name` | string | ✅ | Max 255 chars |
| `email` | string | ✅ | Valid email format, max 255 chars |
| `password` | string | ✅ | 3–72 chars |

**Response `201 Created`**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "Alice",
    "last_name": "Smith",
    "email": "alice@example.com",
    "is_active": false,
    "created_at": "2026-04-01T10:00:00Z"
  },
  "token": "... activation token ..."
}
```

---

### `POST /auth/login`

Authenticate with email and password. Sets the `access_token` cookie.

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
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "first_name": "Alice",
    "last_name": "Smith",
    "email": "alice@example.com",
    "is_active": true
  },
  "token": "... jwt ..."
}
```

**Response Header**
```
Set-Cookie: access_token=<jwt>; HttpOnly; SameSite=Lax; Max-Age=7776000; Path=/
```

---

### `POST /auth/logout`

Clear the session cookie.

**Response `200 OK`**

---

### `POST /auth/activate/{token}`

Activate a user's account using the token from the verification email.

**Path Parameters**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | ✅ | Activation token from email |

**Response `204 No Content`**

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
Set-Cookie: access_token=...
```

**Errors:** `400` state mismatch (CSRF) · `502` Google token exchange failed

---

## Document Endpoints

> All require authentication.

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

---

### `GET /documents` (Planned)

List all documents owned by the authenticated user.

---

### `GET /documents/{id}` (Planned)

Get metadata for a specific document.

---

### `DELETE /documents/{id}` (Planned)

Delete a document and all its Qdrant vectors.

---

## Conversation Endpoints

> All require authentication.

### `POST /conversations`

Create a new chat conversation. Optionally scoped to a specific document.

**Request**
```json
{
  "title": "Research on Auth",
  "document_id": "uuid-..."
}
```

**Response `201 Created`**
```json
{
  "id": "uuid-...",
  "owner_id": "uuid-...",
  "title": "Research on Auth",
  "document_id": "uuid-...",
  "created_at": "...",
  "updated_at": "..."
}
```

---

### `GET /conversations`

List all conversations for the authenticated user.

**Response `200 OK`**
```json
[
  {
    "id": "uuid-...",
    "title": "Research on Auth",
    "updated_at": "..."
  }
]
```

---

### `GET /conversations/{id}`

Get a specific conversation and its messages.

**Response `200 OK`**
```json
{
  "id": "uuid-...",
  "title": "Research on Auth",
  "messages": [
    {
      "role": "user",
      "content": "Hello",
      "created_at": "..."
    },
    {
      "role": "assistant",
      "content": "Hi there!",
      "created_at": "..."
    }
  ]
}
```

---

### `POST /conversations/{id}`

Post a message to a conversation. Supports streaming or non-streaming based on `Accept` header.

**Request**
```json
{
  "content": "How does RAG work?",
  "top_k": 5
}
```

**Response (Non-streaming)**
```json
{
  "role": "assistant",
  "content": "RAG works by...",
  "created_at": "..."
}
```

**Response (Streaming - SSE)**
```
data: R
data: AG
data:  works
data: [DONE]
```

---

### `DELETE /conversations/{id}`

Delete a conversation.

**Response `204 No Content`**


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
  - url: http://localhost:3000/api/v1
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
  /auth/register:         { post: { ... } }
  /auth/activate/{token}: { post: { ... } }
  /auth/login:            { post: { ... } }
  /auth/logout:           { post: { ... } }
  /auth/me:               { get: { ... } }
  /auth/google/login:     { get: { ... } }
  /auth/google/callback:  { get: { ... } }
  /documents/upload:      { post: { ... } }
  /documents:             { get: { ... } }
  /documents/{id}:        { get: { ... }, delete: { ... } }
  /conversations:         { post: { ... }, get: { ... } }
  /conversations/{id}:    { get: { ... }, post: { ... }, patch: { ... }, delete: { ... } }
```
