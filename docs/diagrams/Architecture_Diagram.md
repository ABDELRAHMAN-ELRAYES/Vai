# Architecture

## Vai — Privacy-First AI Document Assistant

**Version:** 1.0
**Date:** April 2026

---

## System Architecture Overview

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
        DB[(PostgreSQL)]
        FS_RAW[(Filesystem — raw/)]
        FS_TMP[(Filesystem — chunks/)]
        QD[(Qdrant)]
    end

    subgraph Ollama [AI Models]
        OL_EMB[nomic-embed-text v1.5]
        OL_GEN[llama3.2:3b]
    end

    subgraph Comms [Outbound]
        SMTP([SMTP])
    end

    subgraph Cleanup [Background]
        BG[Cleanup Worker — 24h draft expiry]
    end

    User --->|"Auth Request"| P1
    P1 <--->|"Verify"| Google
    P1 <--->|"Sessions"| DB
    P1 --->|"JWT Cookie"| User

    User --->|"Upload"| P2
    P2 ==>|"Write raw file"| FS_RAW
    P2 ==>|"Write chunks"| FS_TMP
    P2 ==>|"INSERT (status: draft)"| DB
    P2 --->|"202 Accepted"| User

    User --->|"Query + documentID"| P3
    P3 <-->|"Load chunks (if draft)"| FS_TMP
    P3 <-->|"Embed chunks + question"| OL_EMB
    P3 <-->|"Upsert + search vectors"| QD
    P3 --->|"Context"| OL_GEN
    P3 ==>|"UPDATE status: ready, log messages"| DB
    P3 -.->|"SSE stream"| User

    P1 -.-> P4
    P4 ---> SMTP

    BG -.->|"DELETE drafts older than 24h"| FS_RAW
    BG -.->|"DELETE chunks"| FS_TMP
    BG -.->|"DELETE WHERE status=draft AND age > 24h"| DB

    style Engines fill:none,stroke:none
    style Ollama fill:#f9f9f9,stroke:#333
    style OL_EMB fill:#d1e9ff
    style OL_GEN fill:#fff4dd
    style Cleanup fill:#f9f9f9,stroke:#333,stroke-width:1px
```

---

## Docker Compose Service Topology

```mermaid
graph LR
    subgraph DockerNetwork["Docker Bridge Network: vai_network"]
        WEB["vai-web\n:5173\nVite/React"]
        API["vai-api\n:3000\nGo binary"]
        PG["postgres\n:5432\nPostgreSQL 16"]
        QD["qdrant\n:6334\nVector DB"]
        OL["ollama\n:11434\nLLM Runtime"]
    end

    subgraph Volumes["Persistent Volumes"]
        PGVol[("postgres_data")]
        QDVol[("qdrant_storage")]
        OLVol[("ollama_models")]
    end

    WEB -->|REST/SSE| API
    API -->|pgx/v5| PG
    API -->|HTTP REST| QD
    API -->|HTTP REST| OL
    PG --- PGVol
    QD --- QDVol
    OL --- OLVol

    Host["Host Machine :5173, :3000"] -->|port mapping| WEB
    Host -->|port mapping| API
```

---

## Data Flow — Document Ingestion

Embedding is **deferred** — it does not happen at upload time. The upload endpoint only validates, chunks, and stores the document as `draft`. Embedding runs lazily on the first query.

```mermaid
flowchart LR
    File["📄 Text File\n(multipart upload)"]
    Chunker["Chunker\n512-char chunks\n70-char overlap\nboundary-aware"]
    FS[("Filesystem\nchunks/")]
    PG[("PostgreSQL\nInsert document\nstatus: draft")]
    Response["✅ 202 Accepted\n(document_id, status: draft)"]

    File --> Chunker
    Chunker -->|"[]Chunk"| FS
    FS --> PG
    PG --> Response
```

---

## Data Flow — Deferred Embedding (First Query Only)

Runs automatically inside the RAG engine when a document's status is still `draft`.

```mermaid
flowchart LR
    FS[("Filesystem\nchunks/")]
    Embed["Ollama Embeddings\nnomic-embed-text v1.5\n→ 768-dim vector"]
    Qdrant[("Qdrant\nUpsert vectors\n+ payload")]
    PG[("PostgreSQL\nUPDATE status: ready\nchunk_count")]

    FS -->|"[]Chunk texts"| Embed
    Embed -->|"[]float32 per chunk"| Qdrant
    Qdrant --> PG
```

---

## Data Flow — Chat Query (Streaming)

```mermaid
flowchart LR
    Q["❓ User Question"]
    Embed["Ollama Embeddings\n→ query vector"]
    Qdrant[("Qdrant\nCosine Search\ntop-K chunks")]
    Prompt["Prompt Assembly\nsystem + context + question"]
    Ollama["Ollama LLM\nllama3.2:3b\nstreaming"]
    SSE["📡 SSE Stream\ntoken by token"]
    PG[("PostgreSQL\nSave assistant\nmessage")]

    Q --> Embed
    Embed -->|queryVector| Qdrant
    Qdrant -->|top-K chunks| Prompt
    Prompt --> Ollama
    Ollama -->|tokens| SSE
    SSE --> PG
```

---

## Authentication Architecture

Vai uses **Google OAuth 2.0** exclusively. There is no email/password registration. On successful OAuth callback, a signed JWT is issued and stored in an `HttpOnly` cookie. No refresh token — the JWT has a 90-day TTL and is read automatically from the cookie on every request.

```mermaid
flowchart TD
    subgraph OAuth["Google OAuth 2.0"]
        Init["GET /auth/google/login\nGenerate state → cookie\nRedirect to Google"]
        CB["GET /auth/google/callback\nValidate state\nExchange code\nValidate ID token\nUpsert user"]
    end

    subgraph TokenMgmt["Token Management"]
        JWT["Access Token\nJWT HS256\n90-day TTL\nHttpOnly cookie (SameSite=Lax)\nKey: access_token"]
    end

    subgraph Middleware["Auth Middleware"]
        CK["Read access_token cookie"]
        VF["Validate JWT\n(issuer: VAI-API, audience: USERS)"]
        EV["Check email_verified"]
    end

    Init -->|"OAuth redirect"| Google([Google])
    Google -->|"code + state"| CB
    CB -->|"Upsert user"| PG[("PostgreSQL")]
    CB --> JWT

    JWT -->|"Set-Cookie on response"| Browser(["Browser"])
    Browser -->|"Cookie on every request"| CK
    CK --> VF --> EV
```

---

## Network & Port Map

```mermaid
graph TD
    Internet["🌍 Internet"]
    Nginx["Nginx / Caddy\nTLS Termination\n:443 → :3000"]
    API["vai-api\n:3000"]
    PG["PostgreSQL\n:5432\n⛔ internal only"]
    QD["Qdrant\n:6334\n⛔ internal only"]
    OL["Ollama\n:11434\n⛔ internal only"]
    SMTP["SMTP\n:587\n🌐 outbound only"]
    Google["Google OAuth\n:443\n🌐 outbound only"]

    Internet -->|HTTPS| Nginx
    Nginx --> API
    API --> PG
    API --> QD
    API --> OL
    API -->|outbound| SMTP
    API -->|outbound| Google
```

---

## Deployment Environments

```mermaid
graph LR
    subgraph Dev["Development"]
        DevCompose["docker compose up\nAll services local\nHot-reload via Air\nFrontend: :5173\nAPI: :3000"]
    end

    subgraph Staging["Staging"]
        StagingVM["Remote VM\ndocker compose up\nMirrors production\nSMTP sandbox"]
    end

    subgraph Prod["Production"]
        ProdNginx["Nginx + Let's Encrypt\nTLS :443"]
        ProdAPI["vai-api\n:3000"]
        ProdPG["PostgreSQL\nVolume-backed"]
        ProdQD["Qdrant\nVolume-backed"]
        ProdOL["Ollama\nGPU if available"]
        ProdNginx --> ProdAPI
        ProdAPI --> ProdPG
        ProdAPI --> ProdQD
        ProdAPI --> ProdOL
    end

    subgraph Future["Future (v1.3+)"]
        K8s["Kubernetes\nHelm Chart\nHPA + managed DB"]
    end
```