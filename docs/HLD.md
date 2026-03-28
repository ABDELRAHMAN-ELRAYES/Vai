# High-Level Design (HLD)
## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Date:** June 2025  
**Author:** Lead Software Architect

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Architectural Goals](#architectural-goals)
3. [System Layers](#system-layers)
4. [Core Components](#core-components)
5. [Data Storage Strategy](#data-storage-strategy)
6. [Authentication Architecture](#authentication-architecture)
7. [RAG Pipeline Overview](#rag-pipeline-overview)
8. [Security Architecture](#security-architecture)
9. [Technology Choices & Rationale](#technology-choices--rationale)
10. [Deployment Architecture](#deployment-architecture)

---

## System Overview

Vai is a layered monolith written in Go, designed for self-hosted deployment. It integrates two storage backends — PostgreSQL for relational/transactional data and Qdrant for vector embeddings — with Ollama for local AI inference (both embedding and generation). An external SMTP provider handles transactional email.

All components are containerized and orchestrated via Docker Compose. The API layer is fully stateless; all persistent state lives in PostgreSQL or Qdrant.

---

## Architectural Goals

| Goal | Decision |
|------|----------|
| **Privacy by design** | Zero external AI API calls. All inference local via Ollama. |
| **Developer simplicity** | Single binary Go server. Single `docker compose up` to start everything. |
| **Security** | JWT in HTTP-only cookies. Refresh token rotation. bcrypt passwords. User-scoped Qdrant collections. |
| **Extensibility** | Config-driven model selection. Interface-based service layer for future swappability. |
| **Performance** | Stateless API. Streaming responses via SSE. Connection pooling (pgx). |

---

## System Layers

```
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 1 — CLIENT LAYER                                               │
│  Browser · Mobile App · curl · Third-party Integration (HTTP/HTTPS) │
└─────────────────────────────┬────────────────────────────────────────┘
                              │  REST + SSE
                              ▼
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 2 — API GATEWAY LAYER  (Go HTTP Server :8080)                  │
│  ┌─────────┐ ┌──────────┐ ┌────────┐ ┌────────┐ ┌───────────────┐  │
│  │  /auth  │ │/documents│ │ /chat  │ │/search │ │  /users/me    │  │
│  └─────────┘ └──────────┘ └────────┘ └────────┘ └───────────────┘  │
│  Middleware: JWTAuth · CORS · RateLimit · Logger · RequestID         │
└──────────┬──────────────────┬──────────────────┬─────────────────────┘
           │                  │                  │
           ▼                  ▼                  ▼
┌──────────────┐   ┌──────────────────┐   ┌─────────────────┐
│ TIER 3 — SERVICE LAYER                                     │
│  AuthService │   │   ChatService    │   │   RAGPipeline   │
│  UserService │   │   EmailService   │   │                 │
└──────┬───────┘   └────────┬─────────┘   └────────┬────────┘
       │                    │                       │
       ▼                    ▼                       ▼
┌──────────────────────────────────────────────────────────────────────┐
│  TIER 4 — INTEGRATION LAYER                                          │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │  PostgreSQL       │  │  Qdrant Client   │  │  Ollama Client   │   │
│  │  (pgx/v5)        │  │  (REST)          │  │  (REST)          │   │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘   │
│  ┌──────────────────┐                                                │
│  │  SMTP Client     │                                                │
│  └──────────────────┘                                                │
└──────────────────────────────────────────────────────────────────────┘
           │                    │                       │
           ▼                    ▼                       ▼
┌────────────────┐   ┌─────────────────────┐   ┌───────────────────┐
│  TIER 5 — STORAGE LAYER                                           │
│  PostgreSQL    │   │  Qdrant             │   │  Ollama           │
│  :5432         │   │  :6333              │   │  :11434           │
│  (users, chat, │   │  (vectors per user) │   │  (models cached)  │
│   tokens, docs)│   │                     │   │                   │
└────────────────┘   └─────────────────────┘   └───────────────────┘
```

---

## Core Components

### API Layer

The HTTP server exposes REST endpoints organized into five route groups:

| Route Group | Purpose |
|-------------|---------|
| `/auth` | Registration, login, logout, refresh, OAuth, email verification, password reset |
| `/users` | Profile management, account deletion |
| `/documents` | Upload, list, get metadata, delete |
| `/chat` | Session management, Q&A (sync + streaming), message history |
| `/search` | Debug chunk retrieval without LLM |

The API layer is **stateless** — all state is persisted in PostgreSQL or Qdrant. Middleware runs on every request: JWT validation extracts the user identity and injects it into the request context.

### Authentication Service

Handles all identity operations:
- **Registration**: Validate → bcrypt hash → insert user → send verification email
- **Login**: Lookup by email → bcrypt compare → issue JWT (15min) + refresh token (7d)
- **Token Refresh**: Validate refresh token against DB → rotate (invalidate old, issue new)
- **Google OAuth**: State validation → code exchange → ID token validation → upsert user → issue JWT
- **Email Verification**: HMAC-SHA256 token validation → mark user verified
- **Password Reset**: HMAC token validation → update hash → revoke all refresh tokens

### RAG Pipeline

The core intelligence of Vai. Two distinct workflows:

**Ingestion:**
```
Raw text → Chunker (500-char chunks, 100-char overlap)
         → Ollama /api/embeddings (nomic-embed-text:v1.5)
         → Qdrant Upsert (vector + payload: doc_id, text, index)
         → PostgreSQL INSERT (document metadata)
```

**Query:**
```
Question → Ollama /api/embeddings → query vector
         → Qdrant Search (cosine similarity, top-K)
         → Assemble context prompt (system + chunks + question)
         → Ollama /api/chat (qwen3.5:4b) → stream tokens via SSE
         → Save complete response to PostgreSQL chat_messages
```

### Chat Service

Manages conversation sessions per user. Each session can optionally be scoped to a single document. Messages (user + assistant roles) are persisted in chronological order and retrievable for conversation context.

### Email Service

Wraps SMTP delivery. All emails support HTML + plain text alternatives. Token generation (HMAC-SHA256) happens in AuthService; EmailService receives the pre-generated token and formats the email body.

---

## Data Storage Strategy

| Data Type | Storage | Reason |
|-----------|---------|--------|
| User accounts & credentials | PostgreSQL `users` | ACID, relational queries |
| JWT refresh tokens | PostgreSQL `refresh_tokens` | Needs revocation, multi-device |
| Email verification tokens | PostgreSQL `verification_tokens` | TTL + one-time use semantics |
| Password reset tokens | PostgreSQL `password_reset_tokens` | TTL + one-time use semantics |
| OAuth provider links | PostgreSQL `oauth_accounts` | Relational join with users |
| Document metadata | PostgreSQL `documents` | List/filter by user, FK constraints |
| Chat sessions | PostgreSQL `chat_sessions` | User-scoped, ordered by time |
| Chat messages | PostgreSQL `chat_messages` | Ordered sequence, role + content |
| Document vector embeddings | Qdrant `user_{userID}` collection | Cosine similarity search |
| Uploaded files (temp) | Local filesystem | Deleted after ingestion completes |

---

## Authentication Architecture

### JWT Strategy

```
Login Request
    ↓
AuthService validates credentials
    ↓
Issues: Access Token (JWT, HS256, 15min TTL)
        + Refresh Token (32-byte random, 7-day TTL, stored hashed in DB)
    ↓
Set-Cookie: access_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Max-Age=900
Set-Cookie: refresh_token=<token>; HttpOnly; Secure; SameSite=Strict; Path=/auth/refresh; Max-Age=604800
```

### Token Refresh Flow

```
Client sends request with expired access_token cookie
    ↓
JWT Middleware: token expired → 401
    ↓
Client sends POST /auth/refresh (refresh_token cookie auto-sent by browser)
    ↓
AuthService: lookup token hash in DB → validate not revoked, not expired
    ↓
Rotate: mark old token revoked, generate new refresh token, issue new JWT
    ↓
Return new Set-Cookie headers
```

### OAuth Flow (Google)

```
GET /auth/google
    → generate random state → store in cookie
    → redirect to Google OAuth URL (with state + client_id + scopes)

GET /auth/google/callback?code=...&state=...
    → validate state cookie matches param (CSRF protection)
    → exchange code for Google tokens
    → validate ID token (signature + iss + aud + exp)
    → extract email, name, avatar from ID token claims
    → upsert user in PostgreSQL (create if new, update if exists)
    → issue Vai JWT + refresh token
    → redirect to application
```

---

## RAG Pipeline Overview

### Ingestion Pipeline Detail

```
1. Receive file bytes (multipart/form-data)
2. Validate: size ≤ 10MB, MIME type allowed
3. Decode bytes as UTF-8 text
4. Chunker.Split(text, chunkSize=500, overlap=100) → []Chunk
5. EnsureCollection(userID) → creates Qdrant collection if not exists
6. For each chunk (parallelizable):
   a. Ollama.Embed(chunk.Text, model="nomic-embed-text:v1.5") → []float32 (768-dim)
   b. QdrantClient.Upsert(collection, id=hash(docID+chunkIndex), vector, payload)
7. PostgreSQL.InsertDocument(id, userID, source, chunkCount)
8. Return {document_id, chunks, source}
```

### Query Pipeline Detail

```
1. Receive question string + userID (from JWT)
2. Ollama.Embed(question, model="nomic-embed-text:v1.5") → queryVector
3. QdrantClient.Search(collection=user_{userID}, vector=queryVector, topK, filter=docID?)
   → []SearchResult{text, docID, score}
4. Assemble prompt:
   System: "Answer the question using ONLY the provided context. If the answer is not in the context, say so."
   Context: chunk1.text + "\n---\n" + chunk2.text + ...
   User: question
5. Ollama.StreamChat(prompt, model="qwen3.5:4b") → token channel
6. Write SSE: for each token → "data: {token}\n\n"
7. Send "data: [DONE]\n\n"
8. Assemble full response → PostgreSQL.InsertMessage(sessionID, "assistant", fullText)
```

---

## Security Architecture

| Concern | Mitigation |
|---------|-----------|
| XSS token theft | JWT in HTTP-only cookie — not accessible to JavaScript |
| CSRF attacks | SameSite=Strict cookie + OAuth state parameter validation |
| Password compromise | bcrypt hash with cost factor 12. Plaintext never stored or logged |
| Token replay | Refresh token rotation — old token invalidated on each use |
| OAuth CSRF | State parameter validated before code exchange |
| Brute force | Rate limiting on /auth/login, /auth/register, /auth/forgot-password |
| Data isolation | Qdrant collections namespaced per user. DB queries always filter by user_id |
| Email enumeration | Password reset always returns 202 regardless of email existence |
| Token forging | Email tokens are HMAC-SHA256 signed with server secret |
| Refresh token theft | Tokens stored as SHA-256 hash in DB. Rotating means stolen tokens are quickly invalidated |

---

## Technology Choices & Rationale

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Backend language** | Go 1.22+ | High concurrency, low latency, single binary, excellent stdlib HTTP server |
| **HTTP router** | Chi or standard net/http | Lightweight, idiomatic Go, middleware composition |
| **Relational DB** | PostgreSQL 16 | ACID, mature, excellent pgx/v5 Go driver, JSONB support if needed |
| **Vector DB** | Qdrant | Purpose-built, excellent Go client, cosine/dot product/euclidean, payload filtering |
| **LLM runtime** | Ollama | Unified local runner, REST API, supports CPU + GPU, model management built-in |
| **Auth library** | golang-jwt/jwt | Widely used, well-maintained, HS256 + RS256 support |
| **Password hashing** | golang.org/x/crypto/bcrypt | Standard, configurable cost factor |
| **DB migrations** | golang-migrate/migrate | SQL-first, up/down migrations, CLI + Go API |
| **Containerization** | Docker + Docker Compose | Universal, no Kubernetes complexity for v1.0 |
| **Email** | net/smtp (stdlib) or gomail | Simple, no external dependency for SMTP |

---

## Deployment Architecture

### Development
```
docker compose up
  ├── vai-api        (Go binary, hot-reload via Air)
  ├── postgres       (PostgreSQL 16)
  ├── qdrant         (Qdrant latest)
  └── ollama         (Ollama latest, models volume-cached)
```

### Production (Recommended)
```
                     Internet
                        │
                   [Nginx / Caddy]  ← TLS termination, HTTPS
                        │
                   [vai-api :8080]  ← Go binary
                   /              \
          [PostgreSQL]          [Qdrant]
              :5432               :6333
                                  
          [Ollama :11434]  ← separate machine or GPU server recommended
          
          [SMTP Provider]  ← outbound only (Mailgun, SendGrid, Postfix)
```

### Environment Variables (Key)

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | HS256 signing key (min 32 chars) |
| `OLLAMA_URL` | Ollama server address |
| `QDRANT_URL` | Qdrant HTTP API address |
| `GOOGLE_CLIENT_ID` | OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | OAuth client secret |
| `SMTP_HOST` / `SMTP_PORT` | Email delivery |
| `APP_URL` | Public URL for email links |
