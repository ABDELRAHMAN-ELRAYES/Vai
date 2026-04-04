# Sequence Diagrams
## Vai — Component Interaction Over Time

**Version:** 1.0  
**Date:** June 2025

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

    C->>H: POST /auth/register {email, password, display_name}
    H->>H: Validate input (email format, password strength)
    alt Validation fails
        H-->>C: 422 Unprocessable Entity {field errors}
    end
    H->>AS: Register(email, password, displayName)
    AS->>DB: SELECT * FROM users WHERE email = ?
    DB-->>AS: (no rows)
    AS->>AS: bcrypt.Hash(password, cost=12)
    AS->>DB: INSERT INTO users (id, email, password_hash, display_name)
    DB-->>AS: user record
    AS->>AS: GenerateHMACToken(userID, secret)
    AS->>DB: INSERT INTO verification_tokens (user_id, token_hash, expires_at=+24h)
    DB-->>AS: ok
    AS->>ES: SendVerification(email, displayName, token)
    ES->>SMTP: Send HTML email with verification link
    SMTP-->>ES: accepted
    AS-->>H: User{id, email, display_name, is_verified=false}
    H-->>C: 201 Created {id, email, display_name, is_verified: false}

    Note over C,SMTP: Later — user clicks verification link in email

    C->>H: GET /auth/verify?token=<hmac_token>
    H->>AS: VerifyEmail(token)
    AS->>AS: HMAC.Verify(token, secret) — validate signature
    AS->>DB: SELECT * FROM verification_tokens WHERE token_hash = sha256(token)
    DB-->>AS: token record
    AS->>AS: Check: not used, not expired
    AS->>DB: UPDATE users SET is_verified = TRUE WHERE id = token.user_id
    AS->>DB: UPDATE verification_tokens SET used = TRUE WHERE id = token.id
    DB-->>AS: ok
    AS-->>H: nil
    H-->>C: 200 OK {message: "Email verified successfully"}
```

---

## SD-02: Email + Password Login

```mermaid
sequenceDiagram
    participant C as Client
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL

    C->>H: POST /auth/login {email, password}
    H->>AS: Login(email, password)
    AS->>DB: SELECT * FROM users WHERE email = ?
    alt User not found
        DB-->>AS: (no rows)
        AS-->>H: ErrInvalidCredentials
        H-->>C: 401 Unauthorized {code: INVALID_CREDENTIALS}
    end
    DB-->>AS: user record
    AS->>AS: bcrypt.Compare(password, user.password_hash)
    alt Password mismatch
        AS-->>H: ErrInvalidCredentials
        H-->>C: 401 Unauthorized
    end
    AS->>AS: GenerateJWT(userID, email, isVerified, TTL=15min)
    AS->>AS: GenerateRefreshToken() — crypto/rand 32 bytes
    AS->>DB: INSERT INTO refresh_tokens (user_id, sha256(token), expires_at=+7d)
    DB-->>AS: ok
    AS-->>H: TokenPair{accessToken, refreshToken}
    H->>H: Set access_token cookie (HttpOnly, Secure, Max-Age=900)
    H->>H: Set refresh_token cookie (HttpOnly, Secure, Path=/auth/refresh)
    H-->>C: 200 OK {id, email, display_name}
```

---

## SD-03: Token Refresh

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as JWT Middleware
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL

    C->>MW: GET /some-protected-route (expired access_token cookie)
    MW->>MW: Parse JWT → token expired
    MW-->>C: 401 Unauthorized

    C->>H: POST /auth/refresh (refresh_token cookie auto-sent)
    H->>AS: RefreshTokens(rawRefreshToken)
    AS->>AS: hash = SHA256(rawRefreshToken)
    AS->>DB: SELECT * FROM refresh_tokens WHERE token_hash = hash
    DB-->>AS: token record
    AS->>AS: Check: not revoked, not expired
    AS->>DB: UPDATE refresh_tokens SET revoked = TRUE WHERE id = token.id
    AS->>AS: GenerateJWT(userID, 15min)
    AS->>AS: GenerateNewRefreshToken() — new 32-byte random
    AS->>DB: INSERT INTO refresh_tokens (user_id, sha256(newToken), expires_at=+7d)
    DB-->>AS: ok
    AS-->>H: TokenPair{newAccessToken, newRefreshToken}
    H->>H: Set new cookies (both tokens)
    H-->>C: 200 OK {id, email, display_name}
```

---

## SD-04: Google OAuth 2.0 Flow

```mermaid
sequenceDiagram
    participant C as Client (Browser)
    participant H as API Handler
    participant AS as AuthService
    participant G as Google OAuth
    participant DB as PostgreSQL
    participant SMTP as EmailService

    C->>H: GET /auth/google
    H->>AS: GenerateOAuthState()
    AS-->>H: state (random string)
    H->>H: Set-Cookie oauth_state=<state> (HttpOnly, Max-Age=300)
    H-->>C: 302 Redirect -> accounts.google.com/o/oauth2/auth?...

    C->>G: (User authenticates at Google, grants consent)
    G-->>C: 302 Redirect -> /auth/google/callback?code=<auth_code>&state=<state>

    C->>H: GET /auth/google/callback?code=...&state=...
    H->>H: Read oauth_state cookie, compare to state param
    
    alt State mismatch (CSRF)
        H-->>C: 400 Bad Request {code: INVALID_STATE}
    end
    
    H->>AS: OAuthCallback("google", code, state)
    AS->>G: POST /token {code, client_id, client_secret, redirect_uri}
    G-->>AS: {access_token, id_token, refresh_token}
    
    AS->>AS: ValidateIDToken(id_token) — verify signature, iss, aud, exp
    AS->>AS: Extract claims: email, name, picture, sub (Google user ID)
    
    AS->>DB: SELECT * FROM oauth_accounts WHERE provider='google' AND provider_user_id=sub
    
    alt New user
        DB-->>AS: (no rows)
        AS->>DB: INSERT INTO users (email, display_name, avatar_url, is_verified=TRUE)
        AS->>DB: INSERT INTO oauth_accounts (user_id, provider, provider_user_id)
        AS-)SMTP: Trigger Welcome Email (Async)
    else Existing user
        DB-->>AS: oauth_account record
        AS->>DB: UPDATE users SET avatar_url = ? (refresh from Google)
    end
    
    AS->>AS: GenerateJWT(userID, 15min)
    AS->>AS: GenerateRefreshToken()
    AS->>DB: INSERT INTO refresh_tokens
    
    AS-->>H: TokenPair
    H->>H: Set access_token + refresh_token cookies
    H-->>C: 302 Redirect → / (application home)
```

---

## SD-05: Document Ingestion

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as JWT Middleware
    participant H as Handler
    participant RP as RAGPipeline
    participant FS as Filesystem
    participant CH as Chunker
    participant DB as PostgreSQL

    Note over C,DB: ── Phase 1: Upload (synchronous, client waits) ──

    C->>MW: POST /documents/upload (multipart file + access_token cookie)
    MW->>MW: Validate JWT → extract userID
    MW->>H: Request with userID in context
    H->>RP: IngestDocument(userID, docID, source, file)

    RP->>RP: P2.1 Validate File (size ≤ 10MB, MIME type)
    RP->>FS: P2.2 Write raw file to raw/
    RP->>CH: P2.3 Split(text, size=500, overlap=100)
    CH-->>RP: []Chunk (N chunks)
    RP->>FS: P2.4 Write chunks to chunks/
    RP->>DB: P2.5 INSERT INTO documents (id, user_id, source, size_bytes, status: draft)
    DB-->>RP: ok

    RP-->>H: IngestResult{documentID, status: draft}
    H-->>C: 202 Accepted {document_id, status: "draft"}

    Note over C,DB: ── Phase 2: Message Send (triggered when user sends a message) ──

    participant EC as EmbeddingClient
    participant OL as Ollama
    participant QD as Qdrant

    C->>H: POST /chat {message, document_id}
    H->>RP: PrepareDocument(documentID)
    RP->>FS: Load chunks from chunks/
    FS-->>RP: []Chunk

    loop For each chunk
        RP->>EC: Embed(chunk.Text)
        EC->>OL: P2.6 POST /api/embeddings {model: "nomic-embed-text:v1.5", prompt: chunk}
        OL-->>EC: {embedding: [f32 × 768]}
        EC-->>RP: []float32
    end

    RP->>QD: P2.7 EnsureCollection("user_<userID>", vectorSize=768)
    QD-->>RP: ok (created or already exists)
    RP->>QD: P2.8 Upsert(collection, Point{id, vector, payload{docID, text, index}})
    QD-->>RP: ok
    RP->>DB: P2.9 UPDATE documents SET status=ready, chunk_count=N WHERE id=docID
    DB-->>RP: ok

    RP-->>H: ready
    H->>RP: RunRAG(userID, documentID, message)
    RP-->>H: streamed answer
    H-->>C: 200 OK (streamed response)

    Note over C,DB: ── Background: Cleanup Job (runs on schedule) ──

    participant BG as Cleanup Worker

    BG->>DB: SELECT id, local_path FROM documents WHERE status=draft AND created_at < NOW()-24h
    DB-->>BG: []staleDrafts
    BG->>FS: Delete raw file + chunks for each
    BG->>DB: DELETE FROM documents WHERE id IN (staleDraftIDs)
```

---

## SD-06: Chat Query — Streaming Response

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as JWT Middleware
    participant H as Handler
    participant CS as ChatService
    participant RP as RAGPipeline
    participant EC as EmbeddingClient
    participant QD as Qdrant
    participant OL as Ollama
    participant DB as PostgreSQL

    C->>MW: GET /chat/stream?question=...&top_k=5 (access_token cookie)
    MW->>MW: Validate JWT → extract userID
    MW->>H: Request with userID in context
    
    H->>CS: P3.1 GetOrCreateSession(userID, sessionID?, docID?)
    CS->>DB: SELECT / INSERT chat_sessions
    DB-->>CS: ChatSession
    
    CS->>DB: P3.2 INSERT INTO chat_messages (session_id, role='user', content=question)
    DB-->>CS: ok
    
    H->>RP: StreamAnswer(userID, question, topK, docID?, responseWriter)
    
    RP->>EC: P3.3 Embed(question)
    EC->>OL: POST /api/embeddings {model: nomic-embed-text:v1.5, prompt: question}
    OL-->>EC: {embedding: [f32 × 768]}
    EC-->>RP: queryVector
    
    RP->>QD: P3.4 Search(collection, queryVector, topK, filter=docID?)
    QD-->>RP: []SearchResult{text, docID, score}
    
    RP->>RP: P3.5 AssemblePrompt(systemInstruction, chunks, question)
    
    H->>H: Set headers: Content-Type: text/event-stream, Cache-Control: no-cache
    
    RP->>OL: P3.6 POST /api/chat {model: qwen3.5:4b, messages, stream: true}
    
    loop Streaming tokens
        OL-->>RP: {message: {content: "<token>"}, done: false}
        RP-->>C: data: <token>\n\n
    end
    
    OL-->>RP: {done: true}
    RP-->>C: data: [DONE]\n\n
    
    RP->>DB: P3.7 INSERT INTO chat_messages (session_id, role='assistant', content=fullResponse)
    DB-->>RP: ok
    RP->>DB: UPDATE chat_sessions SET updated_at = NOW()
```

---

## SD-07: Password Reset Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant H as API Handler
    participant AS as AuthService
    participant DB as PostgreSQL
    participant ES as EmailService
    participant SMTP as SMTP Server

    Note over C,SMTP: Step 1 — Request reset

    C->>H: POST /auth/forgot-password {email}
    H->>AS: RequestPasswordReset(email)
    AS->>DB: SELECT * FROM users WHERE email = ?
    Note right of AS: Always returns 202 regardless of result (prevents email enumeration)
    alt User exists
        DB-->>AS: user record
        AS->>AS: GenerateHMACToken(userID, secret, 1h)
        AS->>DB: INSERT INTO password_reset_tokens (user_id, token_hash, expires_at=+1h)
        DB-->>AS: ok
        AS->>ES: SendPasswordReset(email, displayName, token)
        ES->>SMTP: Send email with reset link
    else User not found
        DB-->>AS: (no rows)
        AS->>AS: (no-op)
    end
    AS-->>H: nil (always)
    H-->>C: 202 Accepted {message: "If that email is registered, a reset link has been sent"}

    Note over C,SMTP: Step 2 — Submit new password

    C->>H: POST /auth/reset-password {token, new_password}
    H->>AS: ResetPassword(token, newPassword)
    AS->>AS: Validate password strength
    AS->>AS: hash = SHA256(token)
    AS->>DB: SELECT * FROM password_reset_tokens WHERE token_hash = hash
    DB-->>AS: token record
    AS->>AS: Check: not used, not expired
    AS->>AS: bcrypt.Hash(newPassword, cost=12)
    AS->>DB: UPDATE users SET password_hash = ? WHERE id = token.user_id
    AS->>DB: UPDATE password_reset_tokens SET used = TRUE
    AS->>DB: UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = ? (revoke all sessions)
    DB-->>AS: ok
    AS-->>H: nil
    H-->>C: 200 OK {message: "Password reset successfully. Please log in again."}
```

---

## SD-08: Account Deletion

```mermaid
sequenceDiagram
    participant C as Client
    participant MW as JWT Middleware
    participant H as Handler
    participant US as UserService
    participant QD as Qdrant
    participant DB as PostgreSQL

    C->>MW: DELETE /users/me (access_token cookie)
    MW->>MW: Validate JWT → extract userID
    MW->>H: Request with userID in context
    H->>US: Delete(userID)
    US->>QD: DeleteCollection("user_<userID>")
    QD-->>US: ok (all vectors removed)
    US->>DB: DELETE FROM users WHERE id = userID
    Note right of DB: CASCADE deletes: oauth_accounts, refresh_tokens,\nverification_tokens, password_reset_tokens,\ndocuments, chat_sessions → chat_messages
    DB-->>US: ok
    US-->>H: nil
    H->>H: Clear access_token and refresh_token cookies
    H-->>C: 204 No Content
```
