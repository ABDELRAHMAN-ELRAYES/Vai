# Data Flow Diagram (DFD)
## Vai — How Data Moves Through the System

**Version:** 1.0  
**Date:** June 2025

---

## DFD Level 0 — Context Diagram

The system in its environment. Shows only external actors and the top-level process.

```mermaid
flowchart TD
    User(["👤 User\n(Client)"])
    Google(["🔵 Google OAuth\nProvider"])
    SMTP(["📧 SMTP Server"])
    Vai["🖥️ VAI SYSTEM\n(All processing local)"]

    User -->|"Document files\nQuestions\nCredentials\nProfile updates"| Vai
    Vai -->|"Answers (streaming)\nSession history\nAuth tokens (cookie)\nDocument metadata"| User

    Vai -->|"OAuth code exchange\nID token validation"| Google
    Google -->|"ID token\nUser profile (email, name, avatar)"| Vai

    Vai -->|"Verification email\nPassword reset email\nWelcome email"| SMTP
    SMTP -->|"Delivery confirmation"| Vai
```

---

## DFD Level 1 — Main Processes

```mermaid
flowchart TD
    User(["👤 User"])
    Google(["🔵 Google OAuth"])
    SMTP(["📧 SMTP"])

    P1["P1\nAuthentication\n& Identity"]
    P2["P2\nDocument\nIngestion"]
    P3["P3\nRAG Query\n& Response"]
    P4["P4\nEmail\nDispatch"]

    DB[("D1\nPostgreSQL\nusers · sessions\ntokens · messages")]
    QD[("D2\nQdrant\nvectors per user")]
    OL["D3\nOllama\n(inference)"]
    FS[("D4\nFilesystem\ntemp uploads")]

    User -->|"email, password\nOR oauth code"| P1
    P1 -->|"JWT cookie\nuser profile"| User
    P1 <-->|"read/write\nusers, tokens"| DB
    P1 <-->|"code exchange\ntoken validation"| Google
    P1 -->|"token + user info"| P4

    User -->|"document file\n(multipart)"| P2
    P2 -->|"document_id\nchunk count"| User
    P2 -->|"raw file bytes"| FS
    FS -->|"text content"| P2
    P2 -->|"chunk texts"| OL
    OL -->|"embedding vectors"| P2
    P2 -->|"vectors + payload"| QD
    P2 -->|"document metadata"| DB

    User -->|"question\nsession_id?"| P3
    P3 -->|"SSE token stream"| User
    P3 -->|"question text"| OL
    OL -->|"query vector"| P3
    P3 -->|"vector search"| QD
    QD -->|"top-K chunks"| P3
    P3 -->|"prompt + context"| OL
    OL -->|"generated tokens"| P3
    P3 -->|"message records"| DB

    P4 -->|"email content"| SMTP
```

---

## DFD Level 2 — Document Ingestion (P2 Expanded)

```mermaid
flowchart TD
    Start(["📄 Raw File\n(multipart bytes)"])

    P2_1["P2.1\nValidate File\n(size, MIME type)"]
    P2_2["P2.2\nDecode to\nUTF-8 Text"]
    P2_3["P2.3\nChunker\nSplit text into\noverlapping chunks"]
    P2_4["P2.4\nEmbed Each Chunk\nvia Ollama"]
    P2_5["P2.5\nEnsure Qdrant\nCollection Exists"]
    P2_6["P2.6\nUpsert Vectors\nto Qdrant"]
    P2_7["P2.7\nInsert Document\nMetadata to DB"]
    End(["✅ Response\n{document_id, chunks}"])

    FS[("Filesystem\ntemp file")]
    OL["Ollama\n/api/embeddings\nnomic-embed-text:v1.5"]
    QD[("Qdrant\nuser_{userID}")]
    DB[("PostgreSQL\ndocuments table")]

    Start --> P2_1
    P2_1 -->|"valid"| P2_2
    P2_1 -->|"invalid"| ErrSize["422 Error"]
    P2_2 --> FS
    FS --> P2_3
    P2_3 -->|"[]Chunk{text, index}"| P2_4
    P2_4 -->|"chunk text"| OL
    OL -->|"[]float32 (768)"| P2_4
    P2_4 -->|"chunk vectors"| P2_5
    P2_5 --> P2_6
    P2_6 -->|"points: {id, vector, payload}"| QD
    QD --> P2_7
    P2_7 -->|"metadata record"| DB
    DB --> End
```

---

## DFD Level 2 — RAG Query (P3 Expanded)

```mermaid
flowchart TD
    Start(["❓ User Question\n+ userID (from JWT)\n+ optional session_id\n+ optional document_id"])

    P3_1["P3.1\nGet or Create\nChat Session"]
    P3_2["P3.2\nSave User\nMessage"]
    P3_3["P3.3\nEmbed Question\nvia Ollama"]
    P3_4["P3.4\nVector Search\nin Qdrant"]
    P3_5["P3.5\nAssemble\nContext Prompt"]
    P3_6["P3.6\nStream LLM\nResponse"]
    P3_7["P3.7\nSave Assistant\nMessage"]
    End(["📡 SSE Stream\ndata: <token>\ndata: [DONE]"])

    OL_E["Ollama\n/api/embeddings"]
    OL_C["Ollama\n/api/chat (stream)\nqwen3.5:4b"]
    QD[("Qdrant\nCosine Search")]
    DB[("PostgreSQL\nchat_sessions\nchat_messages")]

    Start --> P3_1
    P3_1 <-->|"read/write session"| DB
    P3_1 --> P3_2
    P3_2 -->|"role=user, content=question"| DB
    P3_2 --> P3_3
    P3_3 -->|"question text"| OL_E
    OL_E -->|"[]float32 (768)"| P3_4
    P3_4 -->|"vector + filter"| QD
    QD -->|"top-K {text, score, docID}"| P3_5
    P3_5 -->|"system prompt + context + question"| P3_6
    P3_6 -->|"streaming prompt"| OL_C
    OL_C -->|"tokens"| P3_6
    P3_6 --> End
    P3_6 -->|"assembled response"| P3_7
    P3_7 -->|"role=assistant, content=answer"| DB
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
