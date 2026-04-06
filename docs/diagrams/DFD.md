# Data Flow Diagram (DFD)

## Vai — How Data Moves Through the System

**Version:** 1.1  
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

    classDef default fill:#ffffff,stroke:#333,stroke-width:1px
```

---

## DFD Level 1 — Main Processes

```mermaid
flowchart TD
    %% User at the very top
    User([User])

    %% Middle Tier: The Processing Engines
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
        FS_RAW[(D4a: Filesystem\nraw/)]
        FS_TMP[(D4b: Filesystem\nchunks/)]
        QD[(D2: Qdrant)]
    end

    subgraph Ollama [AI Models]
        OL_EMB[D3a: Nomic-Embed-Text v1.5]
        OL_GEN[D3b: Qwen 3.5 4b]
    end

    subgraph Comms [Outbound]
        SMTP([SMTP])
    end

    subgraph Cleanup [Background]
        BG[Cleanup Worker\n24h draft expiry]
    end

    %% P1 PATH
    User --->|"1. Auth Req"| P1
    P1 <--->|"2. Verify"| Google
    P1 <--->|"3. Sessions"| DB
    P1 --->|"4. JWT"| User

    %% P2 PATH — Upload only: validate, decode, chunk, save drafts
    User --->|"5. Upload"| P2
    P2 ==>|"6. Write raw file"| FS_RAW
    P2 ==>|"7. Write chunks"| FS_TMP
    P2 ==>|"8. INSERT (status: draft)"| DB
    P2 --->|"9. 202 Accepted\n(documentID, status: draft)"| User

    %% P3 PATH — Message send: deferred embed + RAG
    User --->|"10. Query + documentID"| P3
    P3 <-->|"11. Load chunks\n(if status=draft)"| FS_TMP
    P3 <-->|"12. Embed chunks\n+ question"| OL_EMB
    OL_EMB --->|"13. []float32 (768)"| P3
    P3 <-->|"14. Upsert vectors\n(if status=draft)"| QD
    P3 <-->|"15. Vector search"| QD
    P3 --->|"16. Context"| OL_GEN
    OL_GEN --->|"17. Tokens"| P3
    P3 ==>|"18. UPDATE status: ready\n+ Log messages"| DB
    P3 -.->|"19. SSE"| User

    %% P4 PATH
    P1 -.-> P4
    P4 ---> SMTP

    %% Cleanup PATH
    BG -.->|"DELETE drafts\nolder than 24h"| FS_RAW
    BG -.->|"DELETE chunks"| FS_TMP
    BG -.->|"DELETE WHERE\nstatus=draft AND age > 24h"| DB

    style Engines fill:none,stroke:none
    style Ollama fill:#f9f9f9,stroke:#333
    style OL_EMB fill:#d1e9ff
    style OL_GEN fill:#fff4dd
    style Cleanup fill:#fff9c4,stroke:#f9a825,stroke-width:2px,stroke-dasharray: 4 4
```

---

## DFD Level 2 — Document Ingestion & Message Send (P2 Expanded)

```mermaid
flowchart TD
    subgraph App [Vai Backend Service: Ingestion Pipeline]
        direction TB
        Start(["Raw File\n(multipart bytes)"])

        P2_1{"P2.1: Validate File\n(Size & MIME)"}
        ErrSize(["❌ 422 Error"])

        P2_2["P2.2: Decode to UTF-8"]
        P2_3["P2.3: Chunker\n(Overlapping chunks)"]
        P2_4["P2.4: Save chunks\nto temp storage"]
        P2_5["P2.5: Insert metadata\n(status: draft)"]

        End(["Response\n202 Accepted\n(document_id, status: draft)"])
    end

    subgraph Deferred [Message Send Phase — Deferred]
        direction TB
        D1["Load chunks\nfrom temp storage"]
        D2["Embed chunks\n(nomic-embed-text:v1.5)"]
        D3["Ensure Qdrant\nCollection Exists"]
        D4["Upsert Vectors"]
        D5["UPDATE status → ready"]
        D6(["Run RAG Pipeline"])
    end

    subgraph Background [Background Cleanup Job]
        direction TB
        BG["Periodic scan\nDELETE drafts older than 24h"]
    end

    subgraph AI [Inference: Ollama]
        OL["/api/embeddings\n(nomic-embed-text:v1.5)"]
    end

    subgraph Data [Persistence Layer]
        FS_RAW[("Filesystem\n(raw/)")]
        FS_TMP[("Filesystem\n(chunks/)")]
        QD[("Qdrant\n(user_userID)")]
        DB[("PostgreSQL\n(documents table)")]
    end

    %% Upload phase flow
    Start --> P2_1
    P2_1 -->|"Invalid"| ErrSize
    P2_1 -->|"Valid"| P2_2
    P2_2 --> P2_3
    P2_3 -->|"[]Chunk{text, index}"| P2_4
    P2_4 --> P2_5
    P2_5 --> End

    %% Deferred phase flow
    D1 --> D2
    D2 --> D3
    D3 --> D4
    D4 --> D5
    D5 --> D6

    %% IO interactions — upload phase
    P2_2 ==>|"Write raw file"| FS_RAW
    P2_4 ==>|"Write chunks"| FS_TMP
    P2_5 ==>|"INSERT (status: draft)"| DB

    %% IO interactions — deferred phase
    D1 <-->|"Read chunks"| FS_TMP
    D2 <-->|"chunk texts\n[]float32 (768)"| OL
    D4 ==>|"Upsert Points\n(id, vector, payload)"| QD
    D5 ==>|"UPDATE status: ready"| DB

    %% Cleanup
    BG -.->|"DELETE raw + chunks\nDELETE DB record"| FS_RAW
    BG -.->|"DELETE"| FS_TMP
    BG -.->|"DELETE WHERE\nstatus=draft AND age > 24h"| DB

    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style Deferred fill:#f3e5f5,stroke:#6a1b9a,stroke-width:2px,stroke-dasharray: 5 5
    style Background fill:#fff9c4,stroke:#f9a825,stroke-width:2px,stroke-dasharray: 4 4
    style AI fill:#fff4dd,stroke:#d4a017,stroke-width:2px
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
        Start(["User Question\n(+ userID, documentID, optional sessionID)"])

        P3_0{"P3.0: Document\nStatus Check"}
        P3_0A["P3.0A: Load chunks\nfrom temp storage"]
        P3_0B["P3.0B: Embed chunks\n(deferred phase)"]
        P3_0C["P3.0C: Upsert vectors\nto Qdrant"]
        P3_0D["P3.0D: UPDATE\nstatus → ready"]

        P3_1["P3.1: Get or Create\nChat Session"]
        P3_2["P3.2: Save User\nMessage"]
        P3_3["P3.3: Embed Question\n(via Embedder)"]
        P3_4["P3.4: Vector Search\n(Similarity Match)"]
        P3_5["P3.5: Assemble Context\nPrompt"]
        P3_6["P3.6: Stream LLM\nResponse"]
        P3_7["P3.7: Save Assistant\nMessage"]

        End(["SSE Stream Output\ndata: token ... data: DONE"])
    end

    subgraph AI [Inference: Ollama]
        OL_E["/api/embeddings\n(nomic-embed-text)"]
        OL_C["/api/chat (stream)\n(llama2.3:3b)"]
    end

    subgraph Data [Persistence Layer]
        DB[("PostgreSQL\n(chat_sessions & messages\n& documents)")]
        FS_TMP[("Filesystem\n(chunks/)")]
        QD[("Qdrant\n(Cosine Search)")]
    end

    %% Status gate
    Start --> P3_0
    P3_0 -->|"status = draft\nrun deferred phase"| P3_0A
    P3_0A --> P3_0B
    P3_0B --> P3_0C
    P3_0C --> P3_0D
    P3_0D --> P3_1
    P3_0 -->|"status = ready\nskip to query"| P3_1

    %% Main RAG flow
    P3_1 --> P3_2
    P3_2 --> P3_3
    P3_3 --> P3_4
    P3_4 --> P3_5
    P3_5 --> P3_6
    P3_6 -->|"Streams tokens to client"| End
    P3_6 -->|"On stream completion"| P3_7

    %% IO — deferred phase
    P3_0 <-->|"READ status"| DB
    P3_0A <-->|"Read chunks"| FS_TMP
    P3_0B <-->|"chunk texts\n[]float32 (768)"| OL_E
    P3_0C ==>|"Upsert Points"| QD
    P3_0D ==>|"UPDATE status: ready\nchunk_count"| DB

    %% IO — RAG phase
    P3_1 <-->|"Read/Write Session"| DB
    P3_2 ==>|"Insert Message\n(role=user)"| DB
    P3_3 <-->|"1. Question text\n2. []float32 (768)"| OL_E
    P3_4 <-->|"Vector + Filter\nTop-K {text, score, documentID}"| QD
    P3_6 <-->|"System prompt + context + question\nYields Tokens"| OL_C
    P3_7 ==>|"Insert Message\n(role=assistant)"| DB

    style App fill:#ffffff,stroke:#333,stroke-width:2px
    style AI fill:#fff4dd,stroke:#d4a017,stroke-width:2px
    style Data fill:#e1f5fe,stroke:#01579b,stroke-width:2px

    classDef io fill:#f9f9f9,stroke:#666,stroke-width:1px,stroke-dasharray: 5 5
    classDef deferred fill:#f3e5f5,stroke:#6a1b9a,stroke-width:1px
    class Start,End io
    class P3_0A,P3_0B,P3_0C,P3_0D deferred
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

| Store                | ID  | Read By                            | Written By                                         | Purpose                                                                            |
| -------------------- | --- | ---------------------------------- | -------------------------------------------------- | ---------------------------------------------------------------------------------- |
| PostgreSQL           | D1  | All services                       | AuthService, ChatService, UserService, RAGPipeline | Relational/transactional data including document status lifecycle                  |
| Qdrant               | D2  | RAGPipeline (P3)                   | RAGPipeline deferred phase (P3.0C)                 | Vector similarity search — only written on first message send, not on upload       |
| Filesystem raw/      | D4a | RAGPipeline worker                 | Upload handler (P2)                                | Permanent storage of original uploaded files                                       |
| Filesystem chunks/   | D4b | RAGPipeline deferred phase (P3.0A) | Upload handler (P2)                                | Temporary chunk storage for draft documents — deleted after embedding or after 24h |
| Cookie (client-side) | D5  | All requests                       | Auth handlers                                      | JWT access + refresh tokens                                                        |

---

## Data Stores — Document Status Lifecycle

| Status       | Set By                | Meaning                                     |
| ------------ | --------------------- | ------------------------------------------- |
| `draft`      | Upload handler (P2.5) | File saved, chunks stored, not yet embedded |
| `processing` | RAG engine (P3.0)     | Deferred embedding phase in progress        |
| `ready`      | RAG engine (P3.0D)    | Embedded and searchable in Qdrant           |
| `failed`     | RAG engine (P3.0)     | Embedding failed, eligible for retry        |

---

## Data Classification

| Data Element        | Classification | Storage                          | Retention                                                                   |
| ------------------- | -------------- | -------------------------------- | --------------------------------------------------------------------------- |
| User email          | PII            | PostgreSQL (plaintext)           | Until account deletion                                                      |
| Password hash       | Sensitive      | PostgreSQL                       | Until account deletion                                                      |
| OAuth tokens        | Sensitive      | PostgreSQL (encrypt recommended) | Until expired/revoked                                                       |
| Document text (raw) | Confidential   | Filesystem raw/                  | Until document deleted by user                                              |
| Document chunks     | Confidential   | Filesystem chunks/               | Until first message send (then deleted after embedding) or 24h draft expiry |
| Vector embeddings   | Confidential   | Qdrant payloads                  | Until document deleted                                                      |
| Chat messages       | Confidential   | PostgreSQL                       | Until session/account deleted                                               |
| JWT claims          | Internal       | HTTP cookie (signed)             | 15-minute TTL                                                               |
| Refresh tokens      | Sensitive      | PostgreSQL (hashed)              | 7-day TTL or revocation                                                     |
