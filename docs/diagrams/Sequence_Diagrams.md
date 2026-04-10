# Sequence Diagrams

## Vai — Component Interaction Over Time

**Version:** 1.0  
**Date:** April 2026

---

## SD-01: User Registration & Email Verification

```mermaid
sequenceDiagram
    participant C as Client
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL
    participant ES as EmailService
    participant SMTP as SMTP Server

    C->>H: POST /api/v1/auth/register {email, password, first_name, last_name}
    H->>H: Validate input
    alt Validation fails
        H-->>C: 400 Bad Request {error: message}
    end
    H->>AS: Register(first_name, last_name, email, password)
    AS->>DB: SELECT * FROM users WHERE email = ?
    DB-->>AS: (no rows)
    AS->>AS: Hash Password (argon2/bcrypt)
    AS->>DB: INSERT INTO users (email, password, first_name, last_name, is_active: false)
    DB-->>AS: user record
    AS->>AS: Generate Activation Token
    AS->>DB: INSERT INTO verification_tokens (user_id, token, expired_at)
    DB-->>AS: ok
    AS->>ES: SendWelcome(email, first_name, token)
    ES->>SMTP: Send activation email
    SMTP-->>ES: accepted
    AS-->>H: User record
    H-->>C: 201 Created {user: {id, email, ...}, token}

    Note over C,SMTP: Later — user activates account

    C->>H: POST /api/v1/auth/activate/<token>
    H->>AS: ActivateUser(token)
    AS->>DB: SELECT * FROM verification_tokens WHERE token = ?
    DB-->>AS: token record
    AS->>AS: Check: not expired
    AS->>DB: UPDATE users SET is_active = TRUE WHERE id = token.user_id
    DB-->>AS: ok
    AS-->>H: nil
    H-->>C: 204 No Content
```

---

## SD-02: Email + Password Login

```mermaid
sequenceDiagram
    participant C as Client
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL

    C->>H: POST /api/v1/auth/login {email, password}
    H->>AS: Authenticate(email, password)
    AS->>DB: SELECT * FROM users WHERE email = ?
    alt User not found
        DB-->>AS: (no rows)
        AS-->>H: ErrInvalidCredentials
        H-->>C: 401 Unauthorized {error: "unauthorized"}
    end
    DB-->>AS: user record
    AS->>AS: Compare Hash
    alt Password mismatch
        AS-->>H: ErrInvalidCredentials
        H-->>C: 401 Unauthorized
    end
    AS->>AS: GenerateJWT(userID, TTL=90d)
    AS-->>H: User + JWT
    H->>H: Set access_token cookie (HttpOnly, SameSite=Lax, Max-Age=90d)
    H-->>C: 200 OK {user: {id, email, ...}, token}
```

---

## SD-04: Google OAuth 2.0 Flow (Planned)

```mermaid
sequenceDiagram
    participant C as Client (Browser)
    participant H as API Handler
    participant AS as AuthService
    participant G as Google OAuth
    participant DB as PostgreSQL
    participant SMTP as EmailService

    C->>H: GET /api/v1/auth/google/login
    H->>AS: GenerateOAuthState()
    AS-->>H: state
    H->>H: Set-Cookie oauth_state
    H-->>C: 302 Redirect -> google.com

    C->>G: (User authenticates)
    G-->>C: 302 Redirect -> /api/v1/auth/google/callback?code=...

    C->>H: GET /api/v1/auth/google/callback?code=...
    H->>AS: OAuthCallback(code)
    AS->>G: Token Exchange
    G-->>AS: {access_token, id_token}
    AS->>AS: Validate ID Token
    AS->>DB: SELECT * FROM users WHERE email = ?
    alt New user
        AS->>DB: INSERT INTO users (email, first_name, last_name, is_active: true)
        AS-)SMTP: Trigger Welcome Email
    end

    AS->>AS: GenerateJWT(userID, 90d)
    AS-->>H: JWT
    H->>H: Set access_token cookie
    H-->>C: 302 Redirect → / (home)
```

---

## SD-05: Document Ingestion

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as Auth Middleware
    participant H as Handler
    participant S as DocumentService
    participant FS as Filesystem
    participant CH as Chunker
    participant DB as PostgreSQL

    Note over C,DB: ── Phase 1: Upload (synchronous) ──

    C->>MW: POST /api/v1/documents/upload (multipart + cookie)
    MW->>MW: Validate JWT → extract userID
    MW->>H: Request with user in ctx
    H->>S: CreateDocument(userID, file)

    S->>RP: P2.1 Validate (size ≤ 10MB, MIME)
    S->>DB: P2.2 INSERT INTO documents (owner_id, name, original_name, status: draft)
    DB-->>S: doc record
    S->>CH: P2.3 GenerateChunks(doc)
    CH->>FS: P2.4 Write chunks to /uploads/chunks/
    CH-->>S: ok

    S-->>H: doc{id, status: draft}
    H-->>C: 202 Accepted {id, status: "draft"}

    Note over C,DB: ── Phase 2: First Query (Lazy Embedding) ──

    participant OL as Ollama
    participant QD as Qdrant

    C->>H: POST /api/v1/conversations/{id} {question}
    H->>S: EmbedDocument(docID)
    S->>DB: P2.5 UPDATE status: processing
    S->>FS: Load chunks from filesystem
    FS-->>S: []Chunk

    loop For each chunk
        S->>OL: P2.6 POST /api/embeddings
        OL-->>S: []float32 (768)
    end

    S->>QD: P2.7 Upsert vectors into collection
    QD-->>S: ok
    S->>DB: P2.8 UPDATE status: ready
    DB-->>S: ok

    S-->>H: ready
    H-->>C: (continues to RAG pipeline)
```

---

## SD-06: Chat Query — Streaming Response

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as Auth Middleware
    participant H as Handler
    participant S as ChatService
    participant OL as Ollama
    participant QD as Qdrant
    participant DB as PostgreSQL

    C->>MW: POST /api/v1/conversations/{id} {question} (access_token cookie)
    MW->>MW: Validate JWT → extract user
    MW->>H: Request with user in ctx

    H->>S: SendMessage(conversationID, question)

    S->>DB: P3.1 INSERT INTO messages (conversation_id, role='user', content=question)
    DB-->>S: ok

    S->>OL: P3.2 Embed(question)
    OL-->>S: queryVector (768)

    S->>QD: P3.3 Search(queryVector, topK, filter=documentID)
    QD-->>S: []RetrievedChunks

    S->>OL: P3.4 POST /api/chat {prompt, context, stream: true}

    loop Streaming tokens
        OL-->>H: Token chunks via SSE
        H-->>C: data: <token>
    end

    OL-->>S: [Done]
    S->>DB: P3.5 INSERT INTO messages (conversation_id, role='assistant', content=fullResp)
    DB-->>S: ok
    S->>DB: P3.6 UPDATE conversations SET updated_at = NOW()
```

---

## SD-07: Password Reset Flow (Planned)

```mermaid
sequenceDiagram
    participant C as Client
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL
    participant ES as EmailService

    C->>H: POST /api/v1/auth/forgot-password {email}
    H->>AS: RequestReset(email)
    AS->>DB: INSERT INTO reset_tokens
    AS->>ES: SendResetEmail(email, token)
    H-->>C: 202 Accepted

    Note over C,ES: User submits new password

    C->>H: POST /api/v1/auth/reset-password {token, password}
    H->>AS: ResetPassword(token, password)
    AS->>DB: UPDATE users SET password = ?
    AS-->>H: ok
    H-->>C: 200 OK
```

---

## SD-08: Account Deletion (Planned)

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as Auth Middleware
    participant H as Handler
    participant US as UserService
    participant QD as Qdrant
    participant DB as PostgreSQL

    C->>MW: DELETE /api/v1/auth/me (cookie)
    MW->>MW: Validate JWT
    H->>US: DeleteAccount(userID)
    US->>QD: Delete user vectors
    US->>DB: DELETE FROM users WHERE id = ?
    Note right of DB: CASCADE deletes documents, conversations, etc.
    H-->>C: 204 No Content
```
