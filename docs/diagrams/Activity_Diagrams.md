# Activity Diagrams

## Vai — Process Workflows

**Version:** 1.0  
**Date:** June 2025

---

## AD-01: User Registration

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[User submits registration form\nemail · password · display_name]
    A --> B{Validate input}
    B -->|Invalid: email format\nor password too weak| C[Return 422\nValidation Error\nwith field details]
    C --> End1([🔴 End])

    B -->|Valid| D{Email already\nregistered?}
    D -->|Yes| E[Return 409 Conflict\nEmail already exists]
    E --> End2([🔴 End])

    D -->|No| F[bcrypt.Hash password\ncost factor = 12]
    F --> G[INSERT users table\nis_verified = false]
    G --> H[Generate HMAC-SHA256\nverification token\nexpiry = 24 hours]
    H --> I[INSERT verification_tokens]
    I --> J[Send verification email\nvia SMTP async]
    J --> K[Return 201 Created\nuser object]
    K --> End3([🟢 End])
```

---

## AD-02: Email Verification

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[User clicks verification link\nGET /auth/verify?token=...]
    A --> B[HMAC.Verify token\nusing server secret]
    B --> C{Valid HMAC\nsignature?}
    C -->|No| D[Return 401\nInvalid token]
    D --> End1([🔴 End])

    C -->|Yes| E[Lookup token in DB\nby SHA256 hash]
    E --> F{Token found\nin DB?}
    F -->|No| G[Return 401\nToken not found]
    G --> End2([🔴 End])

    F -->|Yes| H{Token already\nused?}
    H -->|Yes| I[Return 401\nToken already used]
    I --> End3([🔴 End])

    H -->|No| J{Token expired?\nexpiry < NOW}
    J -->|Yes| K[Return 401\nToken expired]
    K --> End4([🔴 End])

    J -->|No| L[UPDATE users\nSET is_verified = TRUE]
    L --> M[UPDATE token\nSET used = TRUE]
    M --> N[Return 200 OK\nEmail verified]
    N --> End5([🟢 End])
```

---

## AD-03: Email + Password Login

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[POST /auth/login\nemail + password]
    A --> B{Input valid?}
    B -->|No| C[Return 422\nValidation Error]
    C --> End1([🔴 End])

    B -->|Yes| D[SELECT user\nWHERE email = ?]
    D --> E{User found?}
    E -->|No| F[Return 401\nInvalid credentials\nnever reveal email existence]
    F --> End2([🔴 End])

    E -->|Yes| G[bcrypt.Compare\npassword vs hash]
    G --> H{Passwords\nmatch?}
    H -->|No| F

    H -->|Yes| I[GenerateJWT\nuserID · email · isVerified\nTTL = 15 min · HS256]
    I --> J[Generate refresh token\n32 bytes crypto/rand]
    J --> K[INSERT refresh_tokens\nstore SHA256 hash\nTTL = 7 days]
    K --> L[Set-Cookie: access_token\nHttpOnly · Secure · SameSite=Strict]
    L --> M[Set-Cookie: refresh_token\nHttpOnly · Secure · Path=/auth/refresh]
    M --> N[Return 200 OK\nuser profile]
    N --> End3([🟢 End])
```

---

## AD-04: Google OAuth Login

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[GET /auth/google]
    A --> B[Generate random state\n32-byte hex string]
    B --> C[Store state in\nHTTP-only cookie\nMax-Age=300s]
    C --> D[302 Redirect\nto Google OAuth URL\nwith state + scopes]
    D --> E[User authenticates\nat Google]
    E --> F[Google redirects\nto /auth/google/callback\nwith code + state]
    F --> G{State matches\ncookie?}
    G -->|No - CSRF attack| H[Return 400\nInvalid state]
    H --> End1([🔴 End])

    G -->|Yes| I[Exchange code\nfor Google tokens\nvia POST /token]
    I --> J{Exchange\nsuccessful?}
    J -->|No| K[Return 502\nGoogle auth failed]
    K --> End2([🔴 End])

    J -->|Yes| L[Validate Google ID token\nsignature · iss · aud · exp]
    L --> M[Extract claims:\nemail · name · picture · sub]
    M --> N{Existing OAuth\naccount?}
    N -->|No - new user| O[INSERT users\nis_verified = TRUE\navatar from Google]
    O --> P[INSERT oauth_accounts\nprovider=google · provider_user_id=sub]

    N -->|Yes| Q[Load existing user\nupdate avatar if changed]

    P --> R[Generate JWT + Refresh Token]
    Q --> R
    R --> S[INSERT refresh_tokens]
    S --> T[Set auth cookies]
    T --> U[302 Redirect → /\napplication home]
    U --> End3([🟢 End])
```

---

## AD-05: Document Upload & Ingestion

```mermaid
flowchart TD
    %% --- Architectural Boundaries ---
    subgraph Middleware [Gatekeeper Layer]
        direction TB
        B{"JWT valid?"}
        D{"Email verified?"}
    end

    subgraph App [Vai Backend Logic]
        direction TB
        Start([Start Request])
        A["POST /documents/upload<br>(multipart file)"]

        %% Validations
        F{"File size<br>≤ 10MB?"}
        H{"Valid MIME<br>type?"}
        J["Decode bytes<br>as UTF-8 text"]

        %% Processing
        M["Chunker.Split text<br>size=500, overlap=100"]
        SAVE["Save chunks to<br>temp storage"]

        %% Returns
        C["Return 401<br>Unauthorized"]
        E["Return 403<br>Email not verified"]
        G["Return 413<br>File too large"]
        I["Return 422<br>Unsupported file type"]
        S["Return 202 Accepted<br>(documentID, status: draft)"]

        End([End Request])
    end

    subgraph Data [Persistence Layer]
        direction TB
        R["INSERT documents<br>(status: draft) → PostgreSQL"]
        CLEANUP["Background Job<br>Delete drafts older than 24h"]
    end

    %% --- CONTROL FLOW ---

    Start --> A
    A --> B

    B -->|No| C
    C --> End
    B -->|Yes| D

    D -->|No| E
    E --> End
    D -->|Yes| F

    F -->|No| G
    G --> End
    F -->|Yes| H

    H -->|No| I
    I --> End
    H -->|Yes| J

    J --> M
    M --> SAVE
    SAVE --> R
    R --> S
    S --> End

    CLEANUP -.->|"Periodic scan<br>DELETE status=draft AND age > 24h"| R

    %% --- STYLING ---
    style Middleware fill:#f9f9f9,stroke:#666,stroke-width:2px,stroke-dasharray: 5 5
    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px

    classDef terminal fill:#333,stroke:#333,stroke-width:2px,color:#fff
    class Start,End terminal

    classDef error fill:#ffebee,stroke:#c62828,stroke-width:1px,color:#b71c1c
    class C,E,G,I error

    classDef success fill:#e8f5e9,stroke:#2e7d32,stroke-width:1px,color:#1b5e20
    class S success

    classDef cleanup fill:#fff9c4,stroke:#f9a825,stroke-width:1px,color:#e65100
    class CLEANUP cleanup
```

---

## AD-06: Chat Query (Streaming)

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[GET /chat/stream\n?question=...&top_k=5]
    A --> B{JWT valid?}
    B -->|No| C[Return 401]
    C --> End1([🔴 End])

    B -->|Yes| D{session_id\nprovided?}
    D -->|No| E[Create new ChatSession\nauto-title from question]
    D -->|Yes| F{Session owned\nby user?}
    F -->|No| G[Return 403 Forbidden]
    G --> End2([🔴 End])

    F -->|Yes| H[Load existing session]
    E --> I
    H --> I[INSERT user message\nrole=user · content=question]
    I --> J[Ollama.Embed question\nnomic-embed-text:v1.5]
    J --> K[Qdrant.Search\nvector · topK · documentID filter?]
    K --> L[Assemble prompt:\nsystem + chunks + question]
    L --> M[Set response headers:\nContent-Type: text/event-stream\nCache-Control: no-cache\nX-Accel-Buffering: no]
    M --> N[Ollama.StreamChat\nllama2.3:3b · stream=true]
    N --> O{More tokens?}
    O -->|Yes| P[Write SSE event:\ndata: token\n\n]
    P --> O
    O -->|No| Q[Write: data: DONE\n\n]
    Q --> R[Assemble full response]
    R --> S[INSERT assistant message\nrole=assistant · content=fullResponse]
    S --> T[UPDATE chat_session.updated_at]
    T --> End3([🟢 End])
```

---

## AD-07: Password Reset

```mermaid
flowchart TD
    StartReq([🟢 Start: Request]) --> A[POST /auth/forgot-password\nemail]
    A --> B[SELECT user by email]
    B --> C{User exists?}
    C -->|Yes| D[Generate HMAC token\nexpiry = 1 hour]
    D --> E[INSERT password_reset_tokens]
    E --> F[Send reset email via SMTP]
    C -->|No| G[No-op silent]
    F --> H[Return 202 Accepted\nsame response always]
    G --> H
    H --> End1([🟢 End])

    StartReset([🟢 Start: Reset]) --> I[POST /auth/reset-password\ntoken + new_password]
    I --> J{Password meets\nstrength requirements?}
    J -->|No| K[Return 422\nPassword too weak]
    K --> End2([🔴 End])

    J -->|Yes| L[Lookup token\nby SHA256 hash]
    L --> M{Token valid?\nnot used · not expired}
    M -->|No| N[Return 401\nToken invalid or expired]
    N --> End3([🔴 End])

    M -->|Yes| O[bcrypt.Hash new password]
    O --> P[UPDATE users\nSET password_hash = ?]
    P --> Q[UPDATE token\nSET used = TRUE]
    Q --> R[UPDATE refresh_tokens\nSET revoked = TRUE\nfor all user sessions]
    R --> S[Return 200 OK\nPassword reset]
    S --> End4([🟢 End])
```

---

## AD-08: Account Deletion

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[DELETE /users/me]
    A --> B{JWT valid?}
    B -->|No| C[Return 401]
    C --> End1([🔴 End])

    B -->|Yes| D["Delete Qdrant collection\nuser_{userID}\nremoves all vectors"]
    D --> E{Qdrant delete\nsuccessful?}
    E -->|No - log error| F[Log error, continue]
    E -->|Yes| G2[DELETE FROM users WHERE id = userID]
    F --> G2
    G2 --> H[PostgreSQL CASCADE deletes documents, sessions, tokens, etc.]
    H --> I[Clear auth cookies (Set-Cookie: Max-Age=0)]

    I --> J[Return 204 No Content]
    J --> End2([🟢 End])
```

---

## AD-09: Token Refresh

```mermaid
flowchart TD
    Start([🟢 Start]) --> A[POST /auth/refresh\nrefresh_token cookie sent automatically]
    A --> B{refresh_token\ncookie present?}
    B -->|No| C[Return 401\nNo refresh token]
    C --> End1([🔴 End])

    B -->|Yes| D[Compute SHA256\nof raw token]
    D --> E[SELECT FROM refresh_tokens\nWHERE token_hash = hash]
    E --> F{Token found?}
    F -->|No| G[Return 401\nToken not found]
    G --> End2([🔴 End])

    F -->|Yes| H{Token revoked?}
    H -->|Yes - possible theft| I[Revoke ALL tokens\nfor this user\nsecurity measure]
    I --> J[Return 401\nToken revoked]
    J --> End3([🔴 End])

    H -->|No| K{Token expired?}
    K -->|Yes| L[Return 401\nToken expired]
    L --> End4([🔴 End])

    K -->|No| M[UPDATE old token\nSET revoked = TRUE]
    M --> N[Generate new JWT\n15 min TTL]
    N --> O[Generate new refresh token\n32-byte random]
    O --> P[INSERT new refresh_token\nTTL = 7 days]
    P --> Q[Set new auth cookies]
    Q --> R[Return 200 OK\nuser profile]
    R --> End5([🟢 End])
```
