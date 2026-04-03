# Data Flow Diagram (DFD)
## Vai — How Data Moves Through the System

**Version:** 1.0  
**Date:** June 2025

---

## DFD Level 0 — Context Diagram

The system in its environment. Shows only external actors and the top-level process.

```mermaid
flowchart TD
    %% Architecture Boundaries
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

    %% --- DATA & CONTROL FLOW ---
    
    %% Client Interactions
    User <==>|"Requests: Docs, Queries, Credentials\nResponses: SSE Answers, Tokens, Metadata"| Vai
    
    %% External API Interactions
    Vai <-->|"Req: OAuth Code Exchange\nRes: ID Token & Profile"| Google
    Vai -.->|"Send: Auth & Welcome Emails\nRecv: Delivery Confirmation"| SMTP

    %% --- THE SIGNATURE PALETTE ---
    style Core fill:#ffffff,stroke:#333,stroke-width:2px
    style Clients fill:#e1f5fe,stroke:#01579b,stroke-width:2px,stroke-dasharray: 5 5
    style External fill:#fff4dd,stroke:#d4a017,stroke-width:2px,stroke-dasharray: 5 5
    
    %% Base node styling for clean text contrast
    classDef default fill:#ffffff,stroke:#333,stroke-width:1px
```

---

## DFD Level 1 — Main Processes

```mermaid
flowchart TD
    %% User at the very top
    User([User])

    %% Middle Tier: The Processing Engines (Distributed horizontally)
    subgraph Engines [Logic Layer]
        direction LR
        P1[P1: Auth & Identity]
        P2[P2: Doc Ingestion]
        P3[P3: RAG Query]
        P4[P4: Email Dispatch]
    end

    %% Bottom/Side Tier: Data & AI (Separated to avoid line crossing)
    subgraph Auth_System [External Auth]
        Google([Google OAuth])
    end

    subgraph Storage [Storage Layer]
        DB[(D1: PostgreSQL)]
        FS[(D4: Filesystem)]
        QD[(D2: Qdrant)]
    end

    subgraph Ollama [AI Models]
        OL_EMB[D3a: Nomic-Embed-Text v1.5]
        OL_GEN[D3b: Qwen 3.5 4b]
    end

    subgraph Comms [Outbound]
        SMTP([SMTP])
    end

    %% --- CONNECTIONS (Organized by Process) ---

    %% P1 PATH (Left Side)
    User --->|"1. Auth Req"| P1
    P1 <--->|"2. Verify"| Google
    P1 <--->|"3. Sessions"| DB
    P1 --->|"4. JWT"| User

    %% P2 PATH (Center-Left)
    User --->|"5. Upload"| P2
    P2 --->|"6. Storage"| FS
    FS --->|"7. Text"| P2
    P2 --->|"8. Chunking"| OL_EMB
    OL_EMB --->|"9. Vectors"| P2
    P2 --->|"10. Upsert"| QD
    P2 --->|"11. Meta"| DB

    %% P3 PATH (Center-Right)
    User --->|"12. Query"| P3
    P3 --->|"13. Vectorize"| OL_EMB
    OL_EMB --->|"14. Search"| P3
    P3 <--->|"15. Retrieval"| QD
    P3 --->|"16. Context"| OL_GEN
    OL_GEN --->|"17. Tokens"| P3
    P3 --->|"18. Log"| DB
    P3 -.->|"19. SSE"| User

    %% P4 PATH (Right Side)
    P1 -.-> P4
    P4 ---> SMTP

    %% Visual Layout Fixes
    style Engines fill:none,stroke:none
    style Ollama fill:#f9f9f9,stroke:#333
    style OL_EMB fill:#d1e9ff
    style OL_GEN fill:#fff4dd
```

---

## DFD Level 2 — Document Ingestion (P2 Expanded)

```mermaid
flowchart TD
    %% Architecture Boundaries
    subgraph App [Vai Backend Service: Ingestion Pipeline]
        direction TB
        Start(["Raw File\n(multipart bytes)"])
        
        P2_1{"P2.1: Validate File\n(Size & MIME)"}
        ErrSize(["❌ 422 Error"])
        
        P2_2["P2.2: Decode to UTF-8"]
        P2_3["P2.3: Chunker\n(Overlapping chunks)"]
        P2_4["P2.4: Embed Chunks"]
        P2_5["P2.5: Ensure Qdrant\nCollection Exists"]
        P2_6["P2.6: Upsert Vectors"]
        P2_7["P2.7: Insert Metadata"]
        
        End(["Response\n(doc_id, chunks)"])
    end

    subgraph AI [Inference: Ollama]
        OL["/api/embeddings\n(nomic-embed-text:v1.5)"]
    end

    subgraph Data [Persistence Layer]
        FS[("Filesystem\n(Temp)")]
        QD[("Qdrant\n(user_userID)")]
        DB[("PostgreSQL\n(documents table)")]
    end

    %% --- CONTROL FLOW (Go Execution Sequence) ---
    Start --> P2_1
    P2_1 -->|"Invalid"| ErrSize
    P2_1 -->|"Valid"| P2_2
    P2_2 --> P2_3
    P2_3 -->|"[]Chunk{text, index}"| P2_4
    P2_4 --> P2_5
    P2_5 --> P2_6
    P2_6 --> P2_7
    P2_7 --> End

    %% --- DATA & IO FLOW (Interactions with external systems) ---
    P2_2 <-->|"Write/Read\nTemp File"| FS
    P2_4 <-->|"1. chunk texts\n2. [ ] float32 (768)"| OL
    P2_6 ==>|"Upsert Points\n(id, vector, payload)"| QD
    P2_7 ==>|"Insert Record\n(metadata)"| DB

    %% --- STYLING ---
    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style AI fill:#fff4dd,stroke:#d4a017,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    
    %% Subtle styling for entry/exit nodes
    classDef io fill:#f9f9f9,stroke:#666,stroke-width:1px,stroke-dasharray: 5 5
    class Start,End,ErrSize io
```

---

## DFD Level 2 — RAG Query (P3 Expanded)

```mermaid
flowchart TD
    %% Architecture Boundaries
    subgraph App [Vai Backend Service: RAG Engine]
        direction TB
        Start(["User Question\n(+ userID, optional session/doc IDs)"])
        
        P3_1["P3.1: Get or Create\nChat Session"]
        P3_2["P3.2: Save User\nMessage"]
        P3_3["P3.3: Embed Question\n(via Embedder)"]
        P3_4["P3.4: Vector Search\n(Similarity Match)"]
        P3_5["P3.5: Assemble Context\nPrompt"]
        P3_6["P3.6: Stream LLM\nResponse"]
        P3_7["P3.7: Save Assistant\nMessage"]
        
        End(["SSE Stream Output\ndata: <token> ... data: [DONE]"])
    end

    subgraph AI [Inference: Ollama]
        OL_E["/api/embeddings\n(nomic-embed-text)"]
        OL_C["/api/chat (stream)\n(qwen3.5:4b)"]
    end

    subgraph Data [Persistence Layer]
        DB[("PostgreSQL\n(chat_sessions & messages)")]
        QD[("Qdrant\n(Cosine Search)")]
    end

    %% --- CONTROL FLOW (Go Execution Sequence) ---
    Start --> P3_1
    P3_1 --> P3_2
    P3_2 --> P3_3
    P3_3 --> P3_4
    P3_4 --> P3_5
    P3_5 --> P3_6
    P3_6 -->|"Streams tokens to client"| End
    P3_6 -->|"On stream completion"| P3_7

    %% --- DATA & IO FLOW (Interactions with external systems) ---
    P3_1 <-->|"Read/Write Session"| DB
    P3_2 ==>|"Insert Message\n(role=user)"| DB
    
    P3_3 <-->|"1. Question text\n2. []float32 (768)"| OL_E
    P3_4 <-->|"Vector + Filter\nReturns Top-K {text, score, docID}"| QD
    
    P3_6 <-->|"1. System prompt + context + question\n2. Yields Tokens"| OL_C
    
    P3_7 ==>|"Insert Message\n(role=assistant)"| DB

    %% --- STYLING ---
    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style AI fill:#fff4dd,stroke:#d4a017,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    
    %% Subtle styling for entry/exit nodes
    classDef io fill:#f9f9f9,stroke:#666,stroke-width:1px,stroke-dasharray: 5 5
    class Start,End io
```

---

## DFD Level 2 — Authentication (P1 Expanded)

```mermaid
flowchart TD
    Email_Login(["📧 Email + Password"])
    OAuth_Start(["🔵 Google OAuth Init"])
    OAuth_CB(["🔄 OAuth Callback\n(code + state)"])
    Verify_Token(["🔗 Email Verification\nLink Click"])
    Reset_Request(["🔑 Password Reset\nRequest"])
    Reset_Submit(["🔒 New Password\nSubmission"])

    P1_1["P1.1\nValidate\nCredentials"]
    P1_2["P1.2\nIssue JWT +\nRefresh Token"]
    P1_3["P1.3\nGenerate OAuth\nState + Redirect"]
    P1_4["P1.4\nExchange Code\n+ Upsert User"]
    P1_5["P1.5\nVerify Email\nToken"]
    P1_6["P1.6\nGenerate Reset\nToken + Email"]
    P1_7["P1.7\nValidate Reset\nToken + Update Hash"]

    DB[("PostgreSQL\nusers · tokens")]
    Google["Google OAuth API"]
    ES["Email Service\n→ SMTP"]

    Email_Login --> P1_1
    P1_1 <-->|"lookup user\nbcrypt compare"| DB
    P1_1 --> P1_2
    P1_2 -->|"store refresh token hash"| DB
    P1_2 -->|"Set-Cookie: access_token\nSet-Cookie: refresh_token"| Out1(["🍪 JWT Cookies"])

    OAuth_Start --> P1_3
    P1_3 -->|"redirect"| Google
    OAuth_CB --> P1_4
    P1_4 <-->|"token exchange"| Google
    P1_4 <-->|"upsert user\noauth_account"| DB
    P1_4 --> P1_2

    Verify_Token --> P1_5
    P1_5 <-->|"validate token\nmark used"| DB
    P1_5 -->|"UPDATE is_verified=true"| DB

    Reset_Request --> P1_6
    P1_6 <-->|"lookup user\ninsert token"| DB
    P1_6 --> ES

    Reset_Submit --> P1_7
    P1_7 <-->|"validate token\nupdate password\nrevoke refresh tokens"| DB
```

---

## Data Stores Summary

| Store | ID | Read By | Written By | Purpose |
|-------|----|---------|-----------|---------|
| PostgreSQL | D1 | All services | AuthService, ChatService, UserService, RAGPipeline | Relational/transactional data |
| Qdrant | D2 | RAGPipeline | RAGPipeline | Vector similarity search |
| Filesystem (temp) | D4 | RAGPipeline | Upload handler | Temporary file storage during ingestion |
| Cookie (client-side) | D5 | All requests | Auth handlers | JWT access + refresh tokens |

## Data Classification

| Data Element | Classification | Storage | Retention |
|-------------|---------------|---------|-----------|
| User email | PII | PostgreSQL (plaintext) | Until account deletion |
| Password hash | Sensitive | PostgreSQL | Until account deletion |
| OAuth tokens | Sensitive | PostgreSQL (encrypt recommended) | Until expired/revoked |
| Document text | Confidential | Qdrant payloads + filesystem (temp) | Until document deleted |
| Chat messages | Confidential | PostgreSQL | Until session/account deleted |
| JWT claims | Internal | HTTP cookie (signed) | 15-minute TTL |
| Refresh tokens | Sensitive | PostgreSQL (hashed) | 7-day TTL or revocation |
