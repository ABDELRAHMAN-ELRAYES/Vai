# Architecture Diagram

## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Date:** June 2025

---

## System Architecture Overview

```mermaid
graph LR
    %% Column 1: Client
    User([User / Browser])

    %% Column 2: Gateway
    subgraph Middleware [Gatekeeper Layer]
        direction TB
        CORS[CORS Policy]
        Auth[JWT Validator]
        Limit[Rate Limiter]
    end

    %% Column 3: Go Backend logic
    subgraph App [Vai Backend Service]
        direction TB
        subgraph Handlers [API Endpoints]
            H_Auth["/auth"]
            H_Chat["/chat (SSE)"]
            H_Docs["/documents"]
        end

        subgraph Logic [Business Logic Services]
            S_Auth[AuthService]
            S_Chat[ChatService]
            S_Email[EmailService]
        end

        subgraph P2_Ingestion [P2: Ingestion Pipeline]
            P2_1[P2.1: Validate]
            P2_2[P2.2: Decode]
            P2_3[P2.3: Chunk]
            P2_4[P2.4: Embedder]
        end

        subgraph P3_RAG [P3: RAG Engine]
            Pipe[Query Pipeline]
        end
    end

    %% Column 4: Infrastructure & AI
    subgraph AI [Inference: Ollama]
        direction TB
        M_Emb["D3a: Nomic-Embed-Text v1.5"]
        M_Gen["D3b: Qwen 3.5 4b"]
    end

    subgraph Data [Persistence Layer]
        direction TB
        PG[("D1: PostgreSQL")]
        QD[("D2: Qdrant")]
        FS[("D4: Filesystem")]
    end

    subgraph External [External Services]
        direction TB
        Google([Google OAuth 2.0])
        SMTP([SMTP Server])
    end

    %% --- THE UPDATED FLOW ---

    %% Entry
    User --> Middleware
    Middleware --> Handlers

    %% Auth Flow
    H_Auth --> S_Auth
    S_Auth --> PG
    S_Auth -.-> Google

    %% Document Ingestion Flow (P2 Logic)
    H_Docs --> P2_1
    P2_1 --> P2_2
    P2_2 <--> FS
    P2_2 --> P2_3
    P2_3 --> P2_4
    P2_4 <--> M_Emb
    P2_4 --> QD
    P2_4 --> PG

    %% Chat/RAG Flow (P3 Logic)
    H_Chat --> S_Chat
    S_Chat --> Pipe
    Pipe <--> M_Emb
    Pipe <--> QD
    Pipe <--> M_Gen
    Pipe --> PG

    %% Email Trigger
    S_Chat --> S_Email
    S_Email -.-> SMTP

    %% Final Styling applied from the second diagram
    style Middleware fill:#f9f9f9,stroke-dasharray: 5 5
    style App fill:#ffffff,stroke:#333
    style AI fill:#fff4dd,stroke:#d4a017
    style Data fill:#e1f5fe,stroke:#01579b
    style External fill:#ffffff,stroke:#333
```

---

## Docker Compose Service Topology

```mermaid
graph LR
    subgraph DockerNetwork["Docker Bridge Network: vai_network"]
        API["vai-api\n:8080\nGo binary"]
        PG["postgres\n:5432\nPostgreSQL 16"]
        QD["qdrant\n:6333\nVector DB"]
        OL["ollama\n:11434\nLLM Runtime"]
    end

    subgraph Volumes["Persistent Volumes"]
        PGVol[("postgres_data")]
        QDVol[("qdrant_storage")]
        OLVol[("ollama_models")]
    end

    API -->|pgx/v5| PG
    API -->|HTTP REST| QD
    API -->|HTTP REST| OL
    PG --- PGVol
    QD --- QDVol
    OL --- OLVol

    Host["Host Machine :8080"] -->|port mapping| API
```

---

## Data Flow — Document Ingestion

```mermaid
flowchart LR
    File["📄 Text File\n(multipart upload)"]
    Chunker["Chunker\n500-char chunks\n100-char overlap"]
    Embed["Ollama Embeddings\nnomic-embed-text:v1.5\n-> 768-dim vector"]
    Qdrant[("Qdrant\nUpsert vectors\n+ payload")]
    PG[("PostgreSQL\nInsert document\nmetadata")]
    Response["✅ Response\n(document_id, chunks)"]

    File --> Chunker
    Chunker -->|"[]Chunk"| Embed
    Embed -->|"[]float32 per chunk"| Qdrant
    Qdrant --> PG
    PG --> Response
```

---

## Data Flow — Chat Query (Streaming)

```mermaid
flowchart LR
    Q["❓ User Question"]
    Embed["Ollama Embeddings\n→ query vector"]
    Qdrant[("Qdrant\nCosine Search\ntop-K chunks")]
    Prompt["Prompt Assembly\nsystem + context + question"]
    Ollama["Ollama LLM\nllama2.3:3b\nstreaming"]
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

```mermaid
flowchart TD
    subgraph EmailAuth["Email / Password Auth"]
        Reg["POST /auth/register"]
        Login["POST /auth/login"]
        Logout["POST /auth/logout"]
        Refresh["POST /auth/refresh"]
    end

    subgraph OAuth["Google OAuth 2.0"]
        Init["GET /auth/google\nGenerate state → cookie\nRedirect to Google"]
        CB["GET /auth/google/callback\nValidate state\nExchange code\nValidate ID token\nUpsert user"]
    end

    subgraph TokenMgmt["Token Management"]
        JWT["Access Token\nJWT HS256\n15 min TTL\nHTTP-only cookie"]
        RT["Refresh Token\n32-byte random\n7-day TTL\nStored hashed in DB\nHTTP-only cookie"]
    end

    Reg -->|bcrypt hash, verify email| PG[("PostgreSQL")]
    Login -->|bcrypt compare| JWT
    Login --> RT
    CB --> JWT
    CB --> RT
    Refresh -->|rotate| RT
    Refresh --> JWT
```

---

## Network & Port Map

```mermaid
graph TD
    Internet["🌍 Internet"]
    Nginx["Nginx / Caddy\nTLS Termination\n:443 → :8080"]
    API["vai-api\n:8080"]
    PG["PostgreSQL\n:5432\n⛔ internal only"]
    QD["Qdrant\n:6333\n⛔ internal only"]
    OL["Ollama\n:11434\n⛔ internal only"]
    SMTP["SMTP\n:587 / :465\n🌐 outbound only"]
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
        DevCompose["docker compose up\nAll services local\nHot-reload via Air"]
    end

    subgraph Staging["Staging"]
        StagingVM["Remote VM\ndocker compose up\nMirrors production\nSMTP sandbox"]
    end

    subgraph Prod["Production"]
        ProdNginx["Nginx + Let's Encrypt\nTLS :443"]
        ProdAPI["vai-api\n:8080"]
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
