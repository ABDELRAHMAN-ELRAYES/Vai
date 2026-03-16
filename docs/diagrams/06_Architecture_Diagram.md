# Architecture Diagram
## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Date:** June 2025

---

## System Architecture Overview

```mermaid
graph TB
    subgraph Clients["External Clients"]
        Browser["🌐 Browser"]
        Mobile["📱 Mobile App"]
        CLI["⌨️ curl / CLI"]
        Integrations["🔌 3rd-party Integration"]
    end

    subgraph VaiBackend["Vai Backend  (Go :8080)"]
        subgraph Middleware["Middleware Layer"]
            JWTAuth["JWTAuth"]
            CORS["CORS"]
            RateLimit["RateLimit"]
            Logger["Logger"]
        end

        subgraph Handlers["HTTP Handlers"]
            AuthH["/auth"]
            UsersH["/users"]
            DocsH["/documents"]
            ChatH["/chat"]
            SearchH["/search"]
        end

        subgraph Services["Service Layer"]
            AuthSvc["AuthService"]
            UserSvc["UserService"]
            ChatSvc["ChatService"]
            EmailSvc["EmailService"]
        end

        subgraph RAG["RAG Pipeline"]
            Chunker["Chunker"]
            EmbClient["Embedding Client"]
            Pipeline["RAGPipeline"]
        end
    end

    subgraph Storage["Storage Layer"]
        PG[("PostgreSQL :5432\nusers · sessions\ntokens · chat")]
        QD[("Qdrant :6333\nvector embeddings\nper-user collections")]
    end

    subgraph AIInference["AI Inference (Local)"]
        Ollama["Ollama :11434"]
        EmbModel["nomic-embed-text:v1.5\n(embeddings)"]
        LLMModel["qwen3.5:4b\n(generation)"]
    end

    subgraph External["External Services"]
        Google["Google OAuth 2.0"]
        SMTP["SMTP Server\n(email delivery)"]
    end

    Clients -->|HTTPS REST + SSE| VaiBackend
    Handlers --> Services
    Handlers --> RAG
    Services --> PG
    RAG --> QD
    RAG --> Ollama
    Ollama --- EmbModel
    Ollama --- LLMModel
    AuthSvc --> PG
    AuthSvc --> Google
    EmailSvc --> SMTP
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
    Embed["Ollama Embeddings\nnomic-embed-text:v1.5\n→ 768-dim vector"]
    Qdrant[("Qdrant\nUpsert vectors\n+ payload")]
    PG[("PostgreSQL\nInsert document\nmetadata")]
    Response["✅ Response\n{document_id, chunks}"]

    File --> Chunker
    Chunker -->|[]Chunk| Embed
    Embed -->|[]float32 per chunk| Qdrant
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
    Ollama["Ollama LLM\nqwen3.5:4b\nstreaming"]
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
