# Data Flow Diagram (DFD)

## Vai — How Data Moves Through the System

**Version:** 1.2
**Date:** April 2026

---

## DFD Level 0 — Context Diagram

The system in its environment. Shows only external actors and the top-level process.

```mermaid
flowchart TD
    subgraph Clients [Client Layer]
        User(["User\n(Browser / App)"])
    end

    subgraph Core [Core System]
        Vai["Vai Backend System\n(Local Processing & Orchestration)"]
    end

    subgraph External [External Services]
        Google(["Google OAuth Provider"])
        SMTP(["SMTP Server"])
    end

    User <==>|"Requests: docs, queries, credentials\nResponses: SSE answers, metadata"| Vai
    Vai <-->|"Req: OAuth code exchange\nRes: ID token & profile"| Google
    Vai -.->|"Send: auth & welcome emails"| SMTP

    style Core fill:#ffffff,stroke:#333,stroke-width:2px
    style Clients fill:#e1f5fe,stroke:#01579b,stroke-width:2px,stroke-dasharray: 5 5
    style External fill:#ede9fe,stroke:#6d28d9,stroke-width:2px,stroke-dasharray: 5 5
```

---

## DFD Level 1 — Main Processes

```mermaid
flowchart TD
    User([User])

    subgraph Engines [Logic Layer]
        direction LR
        P1[P1: Auth & Identity]
        P2[P2: Doc Ingestion]
        P3[P3: RAG Query]
        P4[P4: Email Dispatch]
    end

    subgraph Auth_System [External Auth]
        Google([Google OAuth])
    end

    subgraph Storage [Storage Layer]
        DB[(D1: PostgreSQL)]
        FS_RAW[(D4a: Filesystem — raw/)]
        FS_TMP[(D4b: Filesystem — chunks/)]
        QD[(D2: Qdrant)]
    end

    subgraph Ollama [AI Models]
        OL_EMB[D3a: nomic-embed-text v1.5]
        OL_GEN[D3b: llama3.2:3b]
    end

    subgraph Comms [Outbound]
        SMTP([SMTP])
    end

    subgraph Cleanup [Background]
        BG[Cleanup Worker — 24h draft expiry]
    end

    User --->|"1. Auth request"| P1
    P1 <--->|"2. Verify"| Google
    P1 <--->|"3. Sessions"| DB
    P1 --->|"4. JWT cookie"| User

    User --->|"5. Upload"| P2
    P2 ==>|"6. Write raw file"| FS_RAW
    P2 ==>|"7. Write chunks"| FS_TMP
    P2 ==>|"8. INSERT (status: draft)"| DB
    P2 --->|"9. 202 Accepted (documentID)"| User

    User --->|"10. Query + documentID"| P3
    P3 <-->|"11. Load chunks (if draft)"| FS_TMP
    P3 <-->|"12. Embed chunks + question"| OL_EMB
    OL_EMB --->|"13. []float32 (768)"| P3
    P3 <-->|"14. Upsert vectors (if draft)"| QD
    P3 <-->|"15. Vector search"| QD
    P3 --->|"16. Context"| OL_GEN
    OL_GEN --->|"17. Tokens"| P3
    P3 ==>|"18. UPDATE status: ready + log messages"| DB
    P3 -.->|"19. SSE stream"| User

    P1 -.-> P4
    P4 ---> SMTP

    BG -.->|"DELETE drafts older than 24h"| FS_RAW
    BG -.->|"DELETE chunks"| FS_TMP
    BG -.->|"DELETE WHERE status=draft AND age > 24h"| DB

    style Engines fill:none,stroke:none
    style Ollama fill:#f9f9f9,stroke:#333
    style OL_EMB fill:#d1e9ff
    style OL_GEN fill:#ede9fe
    style Cleanup fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,stroke-dasharray: 4 4
```

---

## DFD Level 2 — Document Ingestion (P2 Expanded)

Embedding is **not** performed at upload time. The upload handler only validates, chunks, and stores the document as `draft`. Embedding runs lazily on the first query.

```mermaid
flowchart TD
    subgraph App [Vai Backend Service: Ingestion Pipeline]
        direction TB
        Start(["Raw file\n(multipart bytes)"])

        P2_1{"P2.1: Validate file\n(size & MIME)"}
        ErrSize(["422 / 413 Error"])

        P2_2["P2.2: Decode to UTF-8"]
        P2_3["P2.3: Chunker\n(512 chars, 70-char overlap,\nboundary-aware)"]
        P2_4["P2.4: Save chunks\nto filesystem"]
        P2_5["P2.5: Insert metadata\n(status: draft)"]

        End(["202 Accepted\n(document_id, status: draft)"])
    end

    subgraph Deferred [First-Query Phase — Deferred Embedding]
        direction TB
        D1["Load chunks\nfrom filesystem"]
        D2["Embed chunks\n(nomic-embed-text v1.5)"]
        D3["Ensure Qdrant\ncollection exists"]
        D4["Upsert vectors"]
        D5["UPDATE status → ready"]
        D6(["Run RAG pipeline"])
    end

    subgraph Background [Background Cleanup]
        direction TB
        BG["Periodic scan\nDELETE drafts older than 24h"]
    end

    subgraph AI [Inference: Ollama]
        OL["/api/embeddings\n(nomic-embed-text v1.5)"]
    end

    subgraph Data [Persistence Layer]
        FS_RAW[("Filesystem\n(raw/)")]
        FS_TMP[("Filesystem\n(chunks/)")]
        QD[("Qdrant")]
        DB[("PostgreSQL\n(documents table)")]
    end

    Start --> P2_1
    P2_1 -->|"Invalid"| ErrSize
    P2_1 -->|"Valid"| P2_2
    P2_2 --> P2_3
    P2_3 -->|"[]Chunk{text, index}"| P2_4
    P2_4 --> P2_5
    P2_5 --> End

    D1 --> D2 --> D3 --> D4 --> D5 --> D6

    P2_2 ==>|"Write raw file"| FS_RAW
    P2_4 ==>|"Write chunks"| FS_TMP
    P2_5 ==>|"INSERT (status: draft)"| DB

    D1 <-->|"Read chunks"| FS_TMP
    D2 <-->|"chunk texts → []float32 (768)"| OL
    D4 ==>|"Upsert points (id, vector, payload)"| QD
    D5 ==>|"UPDATE status: ready, chunk_count"| DB

    BG -.->|"DELETE raw + chunks"| FS_RAW
    BG -.->|"DELETE"| FS_TMP
    BG -.->|"DELETE WHERE status=draft AND age > 24h"| DB

    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style Deferred fill:#ede9fe,stroke:#6d28d9,stroke-width:2px,stroke-dasharray: 5 5
    style Background fill:#e0f2fe,stroke:#0284c7,stroke-width:2px,stroke-dasharray: 4 4
    style AI fill:#f9f9f9,stroke:#333,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px

    classDef io fill:#f9f9f9,stroke:#666,stroke-width:1px,stroke-dasharray: 5 5
    class Start,End,ErrSize io
```

---

## DFD Level 2 — RAG Query (P3 Expanded)

```mermaid
flowchart TD
    subgraph App [Vai Backend Service: RAG Engine]
        direction TB
        Start(["User question\n(+ userID, documentID, optional sessionID)"])

        P3_0{"P3.0: Document\nstatus check"}
        P3_0A["P3.0A: Load chunks\nfrom filesystem"]
        P3_0B["P3.0B: Embed chunks\n(deferred phase)"]
        P3_0C["P3.0C: Upsert vectors\nto Qdrant"]
        P3_0D["P3.0D: UPDATE\nstatus → ready"]

        P3_1["P3.1: Get or create\nchat session"]
        P3_2["P3.2: Save user\nmessage"]
        P3_3["P3.3: Embed question"]
        P3_4["P3.4: Vector search\n(similarity match)"]
        P3_5["P3.5: Assemble context\nprompt"]
        P3_6["P3.6: Stream LLM\nresponse"]
        P3_7["P3.7: Save assistant\nmessage"]

        End(["SSE stream output\ndata: token ... data: [DONE]"])
    end

    subgraph AI [Inference: Ollama]
        OL_E["/api/embeddings\n(nomic-embed-text v1.5)"]
        OL_C["/api/chat (stream)\n(llama3.2:3b)"]
    end

    subgraph Data [Persistence Layer]
        DB[("PostgreSQL\n(sessions, messages, documents)")]
        FS_TMP[("Filesystem\n(chunks/)")]
        QD[("Qdrant\n(cosine search)")]
    end

    Start --> P3_0
    P3_0 -->|"status = draft"| P3_0A
    P3_0A --> P3_0B --> P3_0C --> P3_0D --> P3_1
    P3_0 -->|"status = ready"| P3_1

    P3_1 --> P3_2 --> P3_3 --> P3_4 --> P3_5 --> P3_6
    P3_6 -->|"Streams tokens to client"| End
    P3_6 -->|"On completion"| P3_7

    P3_0 <-->|"READ status"| DB
    P3_0A <-->|"Read chunks"| FS_TMP
    P3_0B <-->|"chunk texts → []float32 (768)"| OL_E
    P3_0C ==>|"Upsert points"| QD
    P3_0D ==>|"UPDATE status: ready, chunk_count"| DB

    P3_1 <-->|"Read/Write session"| DB
    P3_2 ==>|"INSERT role=user"| DB
    P3_3 <-->|"[]float32 (768)"| OL_E
    P3_4 <-->|"Vector + filter — Top-K"| QD
    P3_6 <-->|"System prompt + context + question"| OL_C
    P3_7 ==>|"INSERT role=assistant"| DB

    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style AI fill:#f9f9f9,stroke:#333,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px

    classDef io fill:#f9f9f9,stroke:#666,stroke-width:1px,stroke-dasharray: 5 5
    classDef deferred fill:#ede9fe,stroke:#6d28d9,stroke-width:1px
    class Start,End io
    class P3_0A,P3_0B,P3_0C,P3_0D deferred
```

---

## DFD Level 2 — Authentication (P1 Expanded)

Vai uses **Google OAuth 2.0 exclusively**. There is no email/password registration, no refresh token, and no password reset flow. On successful OAuth callback, a signed JWT is issued and stored in an `HttpOnly` cookie. The cookie is read automatically on every subsequent request — no `Authorization` header is needed.

```mermaid
flowchart TD
    OAuth_Start(["GET /auth/google/login"])
    OAuth_CB(["GET /auth/google/callback\n(code + state)"])

    P1_3["P1.1: Generate OAuth\nstate + redirect"]
    P1_4["P1.2: Exchange code\n+ upsert user"]
    P1_5["P1.3: Issue JWT\n(HS256, 90-day TTL)"]

    subgraph Middleware [Auth Middleware — every protected request]
        CK["Read access_token cookie"]
        VF["Validate JWT\n(issuer: vai-server, audience: users)"]
        EV["Check email_verified"]
    end

    DB[("PostgreSQL\n(users)")]
    Google["Google OAuth API"]

    OAuth_Start --> P1_3
    P1_3 -->|"Redirect"| Google
    Google -->|"code + state"| OAuth_CB
    OAuth_CB --> P1_4
    P1_4 <-->|"Token exchange"| Google
    P1_4 <-->|"Upsert user"| DB
    P1_4 --> P1_5
    P1_5 -->|"Set-Cookie: access_token\n(HttpOnly, SameSite=Lax, MaxAge=90d)"| Browser(["Browser"])

    Browser -->|"Cookie on every request"| CK
    CK --> VF --> EV
```

---

## Data Stores Summary

| Store | ID | Read By | Written By | Purpose |
|---|---|---|---|---|
| PostgreSQL | D1 | All services | Auth, Chat, User, RAG pipeline | Relational data: users, documents, sessions, messages |
| Qdrant | D2 | RAG pipeline (P3) | RAG pipeline — deferred phase (P3.0C) | Vector similarity search — written on first query, not on upload |
| Filesystem raw/ | D4a | — | Upload handler (P2) | Original uploaded files |
| Filesystem chunks/ | D4b | RAG pipeline — deferred phase (P3.0A) | Upload handler (P2) | Chunk storage for draft documents — deleted after embedding or after 24h |
| Cookie (client-side) | D5 | All requests | Auth handler (P1) | JWT access token |

---

## Document Status Lifecycle

| Status | Set By | Meaning |
|---|---|---|
| `draft` | Upload handler (P2.5) | File saved, chunks stored, not yet embedded |
| `processing` | RAG engine (P3.0) | Deferred embedding phase in progress |
| `ready` | RAG engine (P3.0D) | Embedded and searchable in Qdrant |
| `failed` | RAG engine (P3.0) | Embedding failed, eligible for retry |

---

## Data Classification

| Data Element | Classification | Storage | Retention |
|---|---|---|---|
| User email | PII | PostgreSQL (plaintext) | Until account deletion |
| OAuth tokens | Sensitive | PostgreSQL | Until expired / revoked |
| Document text (raw) | Confidential | Filesystem raw/ | Until document deleted by user |
| Document chunks | Confidential | Filesystem chunks/ | Deleted after deferred embedding completes, or after 24h draft expiry |
| Vector embeddings | Confidential | Qdrant payloads | Until document deleted |
| Chat messages | Confidential | PostgreSQL | Until session / account deleted |
| JWT (access token) | Internal | HttpOnly cookie (signed, HS256) | 90-day TTL |