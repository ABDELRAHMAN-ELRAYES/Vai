# Low-Level Design (LLD)
## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Date:** June 2025  
**Author:** Lead Software Architect

---

## Table of Contents

1. [Package Structure](#package-structure)
2. [Data Models](#data-models)
3. [Service Interfaces](#service-interfaces)
4. [Key Algorithms](#key-algorithms)
5. [Middleware Design](#middleware-design)
6. [Error Handling Strategy](#error-handling-strategy)
7. [Database Access Layer](#database-access-layer)
8. [Configuration](#configuration)

---

## Package Structure

```
vai/
├── main.go                        # Entry point: wire all layers, start HTTP server
├── go.mod
├── go.sum
│
├── config/
│   └── config.go                  # Load env vars → typed Config struct; validate required fields
│
├── db/
│   ├── db.go                      # pgx/v5 connection pool setup
│   └── migrations/                # SQL migration files (golang-migrate)
│       ├── 001_create_users.up.sql
│       ├── 001_create_users.down.sql
│       ├── 002_create_oauth_accounts.up.sql
│       ├── 003_create_refresh_tokens.up.sql
│       ├── 004_create_verification_tokens.up.sql
│       ├── 005_create_password_reset_tokens.up.sql
│       ├── 006_create_documents.up.sql
│       ├── 007_create_chat_sessions.up.sql
│       └── 008_create_chat_messages.up.sql
│
├── models/
│   ├── user.go                    # User, OAuthAccount structs
│   ├── token.go                   # RefreshToken, VerificationToken, PasswordResetToken
│   ├── document.go                # Document struct
│   └── chat.go                    # ChatSession, ChatMessage structs
│
├── handlers/
│   ├── auth.go                    # Register, Login, Logout, Refresh, VerifyEmail, ForgotPassword, ResetPassword, GoogleOAuth, GoogleCallback
│   ├── users.go                   # GetMe, UpdateMe, DeleteMe
│   ├── documents.go               # Upload, List, Get, Delete
│   ├── chat.go                    # CreateSession, ListSessions, GetMessages, Chat, StreamChat, DeleteSession
│   └── search.go                  # Search (debug)
│
├── services/
│   ├── auth/
│   │   ├── auth.go                # AuthService implementation
│   │   ├── jwt.go                 # JWT generation, validation, claims
│   │   └── oauth.go               # Google OAuth client, token exchange
│   ├── user/
│   │   └── user.go                # UserService implementation
│   ├── chat/
│   │   └── chat.go                # ChatService implementation
│   └── email/
│       ├── email.go               # EmailService interface + SMTP implementation
│       └── templates/             # HTML email templates
│           ├── verification.html
│           ├── password_reset.html
│           └── welcome.html
│
├── rag/
│   └── pipeline.go                # RAGPipeline: IngestDocument, Search, Answer, StreamAnswer
│
├── chunker/
│   └── chunker.go                 # Split text into overlapping chunks
│
├── embeddings/
│   └── embeddings.go              # OllamaEmbeddingClient: Embed(text) → []float32
│
├── vectorstore/
│   └── qdrant.go                  # QdrantClient: Upsert, Search, Delete, EnsureCollection
│
└── middleware/
    ├── auth.go                    # JWTAuth middleware
    ├── cors.go                    # CORS headers
    ├── ratelimit.go               # Token bucket rate limiter per IP
    ├── logger.go                  # Structured request logging
    └── requestid.go               # Inject X-Request-ID header
```

---

## Data Models

### User

```go
// models/user.go

type User struct {
    ID           uuid.UUID `db:"id"`
    Email        string    `db:"email"`
    PasswordHash *string   `db:"password_hash"` // NULL for OAuth-only users
    DisplayName  string    `db:"display_name"`
    AvatarURL    *string   `db:"avatar_url"`
    IsVerified   bool      `db:"is_verified"`
    CreatedAt    time.Time `db:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"`
}

type OAuthAccount struct {
    ID             uuid.UUID  `db:"id"`
    UserID         uuid.UUID  `db:"user_id"`
    Provider       string     `db:"provider"`        // "google"
    ProviderUserID string     `db:"provider_user_id"`
    AccessToken    *string    `db:"access_token"`
    RefreshToken   *string    `db:"refresh_token"`
    ExpiresAt      *time.Time `db:"expires_at"`
    CreatedAt      time.Time  `db:"created_at"`
}
```

### Tokens

```go
// models/token.go

type RefreshToken struct {
    ID        uuid.UUID `db:"id"`
    UserID    uuid.UUID `db:"user_id"`
    TokenHash string    `db:"token_hash"` // SHA-256 hex of the random token
    ExpiresAt time.Time `db:"expires_at"`
    Revoked   bool      `db:"revoked"`
    CreatedAt time.Time `db:"created_at"`
}

type VerificationToken struct {
    ID        uuid.UUID `db:"id"`
    UserID    uuid.UUID `db:"user_id"`
    TokenHash string    `db:"token_hash"` // HMAC-SHA256 hex
    ExpiresAt time.Time `db:"expires_at"` // 24 hours
    Used      bool      `db:"used"`
    CreatedAt time.Time `db:"created_at"`
}

type PasswordResetToken struct {
    ID        uuid.UUID `db:"id"`
    UserID    uuid.UUID `db:"user_id"`
    TokenHash string    `db:"token_hash"`
    ExpiresAt time.Time `db:"expires_at"` // 1 hour
    Used      bool      `db:"used"`
    CreatedAt time.Time `db:"created_at"`
}
```

### Document

```go
// models/document.go

type Document struct {
    ID             string    `db:"id"`              // slug: "my-document"
    UserID         uuid.UUID `db:"user_id"`
    Source         string    `db:"source"`          // original filename
    ChunkCount     int       `db:"chunk_count"`
    SizeBytes      int64     `db:"size_bytes"`
    CollectionName string    `db:"collection_name"` // "user_<userID>"
    CreatedAt      time.Time `db:"created_at"`
}
```

### Chat

```go
// models/chat.go

type ChatSession struct {
    ID         uuid.UUID  `db:"id"`
    UserID     uuid.UUID  `db:"user_id"`
    Title      string     `db:"title"`
    DocumentID *string    `db:"document_id"` // optional document scope
    CreatedAt  time.Time  `db:"created_at"`
    UpdatedAt  time.Time  `db:"updated_at"`
}

type ChatMessage struct {
    ID         uuid.UUID `db:"id"`
    SessionID  uuid.UUID `db:"session_id"`
    Role       string    `db:"role"`       // "user" | "assistant"
    Content    string    `db:"content"`
    TokensUsed *int      `db:"tokens_used"`
    CreatedAt  time.Time `db:"created_at"`
}
```

### JWT Claims

```go
// services/auth/jwt.go

type JWTClaims struct {
    UserID     string `json:"sub"`
    Email      string `json:"email"`
    IsVerified bool   `json:"verified"`
    jwt.RegisteredClaims
}

// Access token: HS256, 15-minute TTL
// Signed with env JWT_SECRET (min 32 chars)
```

---

## Service Interfaces

### AuthService

```go
// services/auth/auth.go

type AuthService interface {
    Register(ctx context.Context, email, password, displayName string) (*models.User, error)
    Login(ctx context.Context, email, password string) (*TokenPair, error)
    Logout(ctx context.Context, refreshToken string) error
    RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
    VerifyEmail(ctx context.Context, token string) error
    RequestPasswordReset(ctx context.Context, email string) error
    ResetPassword(ctx context.Context, token, newPassword string) error
    OAuthCallback(ctx context.Context, provider, code, state string) (*TokenPair, error)
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    AccessExpiry  time.Time
    RefreshExpiry time.Time
}
```

### UserService

```go
// services/user/user.go

type UserService interface {
    GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
    Update(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*models.User, error)
    Delete(ctx context.Context, userID uuid.UUID) error
}

type UpdateUserRequest struct {
    DisplayName *string `json:"display_name"`
    AvatarURL   *string `json:"avatar_url"`
}
```

### ChatService

```go
// services/chat/chat.go

type ChatService interface {
    CreateSession(ctx context.Context, userID uuid.UUID, title string, docID *string) (*models.ChatSession, error)
    ListSessions(ctx context.Context, userID uuid.UUID) ([]models.ChatSession, error)
    GetSession(ctx context.Context, userID, sessionID uuid.UUID) (*models.ChatSession, error)
    DeleteSession(ctx context.Context, userID, sessionID uuid.UUID) error
    AddMessage(ctx context.Context, sessionID uuid.UUID, role, content string, tokens *int) (*models.ChatMessage, error)
    GetMessages(ctx context.Context, userID, sessionID uuid.UUID) ([]models.ChatMessage, error)
}
```

### EmailService

```go
// services/email/email.go

type EmailService interface {
    SendVerification(ctx context.Context, to, displayName, token string) error
    SendPasswordReset(ctx context.Context, to, displayName, token string) error
    SendWelcome(ctx context.Context, to, displayName string) error
}
```

### RAGPipeline

```go
// rag/pipeline.go

type RAGPipeline interface {
    IngestDocument(ctx context.Context, userID uuid.UUID, docID, source, text string) (*IngestResult, error)
    Search(ctx context.Context, userID uuid.UUID, query string, topK int, docID *string) ([]SearchResult, error)
    Answer(ctx context.Context, userID uuid.UUID, question string, topK int, docID *string) (string, error)
    StreamAnswer(ctx context.Context, userID uuid.UUID, question string, topK int, docID *string, w io.Writer) error
    DeleteDocument(ctx context.Context, userID uuid.UUID, docID string) error
}

type IngestResult struct {
    DocumentID string `json:"document_id"`
    ChunkCount int    `json:"chunks"`
    Source     string `json:"source"`
}

type SearchResult struct {
    DocumentID string  `json:"document_id"`
    ChunkText  string  `json:"text"`
    Score      float64 `json:"score"`
    ChunkIndex int     `json:"chunk_index"`
}
```

### Chunker

```go
// chunker/chunker.go

type Chunk struct {
    Text       string
    Index      int
    StartChar  int
    EndChar    int
}

type Chunker struct {
    ChunkSize    int // default: 500
    ChunkOverlap int // default: 100
}

func (c *Chunker) Split(text string) []Chunk
```

### Embedding Client

```go
// embeddings/embeddings.go

type EmbeddingClient interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

type OllamaEmbeddingClient struct {
    BaseURL string
    Model   string // "nomic-embed-text:v1.5"
    Client  *http.Client
}
```

### Qdrant Client

```go
// vectorstore/qdrant.go

type VectorStore interface {
    EnsureCollection(ctx context.Context, collection string, vectorSize int) error
    Upsert(ctx context.Context, collection string, points []Point) error
    Search(ctx context.Context, collection string, vector []float32, topK int, filter *Filter) ([]SearchResult, error)
    DeleteByFilter(ctx context.Context, collection string, filter Filter) error
    DeleteCollection(ctx context.Context, collection string) error
}

type Point struct {
    ID      string
    Vector  []float32
    Payload map[string]interface{}
}

type Filter struct {
    Must []Condition
}

type Condition struct {
    Key   string
    Value interface{}
}
```

---

## Key Algorithms

### Chunker Algorithm

```
function Split(text, chunkSize, overlap):
    chunks = []
    start = 0
    index = 0
    while start < len(text):
        end = min(start + chunkSize, len(text))
        chunk = text[start:end]
        chunks.append(Chunk{Text: chunk, Index: index, Start: start, End: end})
        start = start + chunkSize - overlap  // move forward by (size - overlap)
        index++
    return chunks
```

**Example:** text of 1200 chars, size=500, overlap=100:
- Chunk 0: chars 0–500
- Chunk 1: chars 400–900
- Chunk 2: chars 800–1200

### Qdrant Point ID Generation

```go
// Deterministic UUID from document ID + chunk index
// Ensures re-ingestion overwrites existing vectors (upsert semantics)
func pointID(docID string, chunkIndex int) string {
    h := sha256.New()
    h.Write([]byte(fmt.Sprintf("%s:%d", docID, chunkIndex)))
    sum := h.Sum(nil)
    id, _ := uuid.FromBytes(sum[:16])
    return id.String()
}
```

### Refresh Token Rotation

```go
func (s *authService) RefreshTokens(ctx, rawToken string) (*TokenPair, error) {
    hash := sha256hex(rawToken)
    
    token, err := s.db.GetRefreshTokenByHash(ctx, hash)
    if err != nil || token.Revoked || token.ExpiresAt.Before(time.Now()) {
        return nil, ErrInvalidToken
    }
    
    // Rotate: revoke old
    s.db.RevokeRefreshToken(ctx, token.ID)
    
    // Issue new pair
    newRefresh := generateSecureToken(32)
    s.db.InsertRefreshToken(ctx, token.UserID, sha256hex(newRefresh), 7*24*time.Hour)
    newJWT := generateJWT(token.UserID, 15*time.Minute)
    
    return &TokenPair{AccessToken: newJWT, RefreshToken: newRefresh}, nil
}
```

---

## Middleware Design

### JWTAuth Middleware

```go
func JWTAuth(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("access_token")
            if err != nil {
                writeError(w, 401, "UNAUTHORIZED", "Missing access token")
                return
            }
            
            claims, err := validateJWT(cookie.Value, secret)
            if err != nil {
                writeError(w, 401, "UNAUTHORIZED", "Invalid or expired token")
                return
            }
            
            // Inject user ID into context
            ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Rate Limiter

Token bucket per IP address. Auth endpoints: 20 requests/minute. Implemented using `golang.org/x/time/rate` with a per-IP limiter map (with TTL cleanup goroutine).

---

## Error Handling Strategy

### Custom Error Types

```go
// Sentinel errors for domain conditions
var (
    ErrEmailAlreadyExists   = errors.New("email already registered")
    ErrInvalidCredentials   = errors.New("invalid email or password")
    ErrEmailNotVerified     = errors.New("email address not verified")
    ErrInvalidToken         = errors.New("token is invalid, expired, or already used")
    ErrDocumentNotFound     = errors.New("document not found")
    ErrSessionNotFound      = errors.New("chat session not found")
    ErrUnauthorized         = errors.New("not authorized to access this resource")
)
```

### HTTP Error Mapping

```go
func writeServiceError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, ErrEmailAlreadyExists):
        writeError(w, 409, "EMAIL_EXISTS", "Email is already registered")
    case errors.Is(err, ErrInvalidCredentials):
        writeError(w, 401, "INVALID_CREDENTIALS", "Email or password is incorrect")
    case errors.Is(err, ErrEmailNotVerified):
        writeError(w, 403, "EMAIL_NOT_VERIFIED", "Please verify your email before continuing")
    case errors.Is(err, ErrInvalidToken):
        writeError(w, 401, "INVALID_TOKEN", "Token is invalid or has expired")
    case errors.Is(err, ErrDocumentNotFound), errors.Is(err, ErrSessionNotFound):
        writeError(w, 404, "NOT_FOUND", "Resource not found")
    case errors.Is(err, ErrUnauthorized):
        writeError(w, 403, "FORBIDDEN", "Access denied")
    default:
        log.Error("internal error", "err", err)
        writeError(w, 500, "INTERNAL_ERROR", "An unexpected error occurred")
    }
}
```

### Error Response Envelope

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Email or password is incorrect",
    "request_id": "req_01HXYZ..."
  }
}
```

---

## Database Access Layer

All DB operations use `pgx/v5` directly (no ORM). Queries are written as plain SQL. Functions follow the pattern:

```go
// Example: Get user by email
func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    row := q.db.QueryRow(ctx, `
        SELECT id, email, password_hash, display_name, avatar_url, is_verified, created_at, updated_at
        FROM users
        WHERE email = $1
    `, email)
    
    var u models.User
    err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.AvatarURL, &u.IsVerified, &u.CreatedAt, &u.UpdatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrUserNotFound
    }
    return &u, err
}
```

**Conventions:**
- All queries parameterized (`$1`, `$2`, ...) — no string interpolation
- Transactions used for multi-step operations (e.g., rotate refresh token, insert user + verification token)
- Connection pool configured: max 25 connections, 5-minute idle timeout

---

## Configuration

```go
// config/config.go

type Config struct {
    // Server
    Port string `env:"PORT" default:"8080"`

    // Database
    DatabaseURL string `env:"DATABASE_URL" required:"true"`

    // JWT
    JWTSecret           string        `env:"JWT_SECRET" required:"true"`
    JWTAccessTokenTTL   time.Duration `env:"JWT_ACCESS_TTL" default:"15m"`
    JWTRefreshTokenTTL  time.Duration `env:"JWT_REFRESH_TTL" default:"168h"` // 7 days

    // Ollama
    OllamaURL       string `env:"OLLAMA_URL" default:"http://localhost:11434"`
    EmbeddingModel  string `env:"EMBEDDING_MODEL" default:"nomic-embed-text:v1.5"`
    ChatModel       string `env:"CHAT_MODEL" default:"qwen3.5:4b"`

    // Qdrant
    QdrantURL   string `env:"QDRANT_URL" default:"http://localhost:6333"`
    VectorSize  int    `env:"VECTOR_SIZE" default:"768"`

    // Chunker
    ChunkSize    int `env:"CHUNK_SIZE" default:"500"`
    ChunkOverlap int `env:"CHUNK_OVERLAP" default:"100"`

    // OAuth
    GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
    GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
    GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL"`

    // Email
    SMTPHost     string `env:"SMTP_HOST"`
    SMTPPort     int    `env:"SMTP_PORT" default:"587"`
    SMTPUsername string `env:"SMTP_USERNAME"`
    SMTPPassword string `env:"SMTP_PASSWORD"`
    SMTPFrom     string `env:"SMTP_FROM"`

    // App
    AppURL      string `env:"APP_URL" default:"http://localhost:8080"`
    RateLimitRPM int   `env:"RATE_LIMIT_RPM" default:"20"`
}
```
