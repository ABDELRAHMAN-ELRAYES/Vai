# Product Requirements Document (PRD)
## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Status:** Draft  
**Date:** June 2025  
**Owner:** Product Team

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Product Vision](#product-vision)
4. [Target Users](#target-users)
5. [Functional Requirements](#functional-requirements)
6. [Non-Functional Requirements](#non-functional-requirements)
7. [User Stories](#user-stories)
8. [Acceptance Criteria](#acceptance-criteria)

---

## Executive Summary

Vai is a self-hosted, privacy-first AI document assistant that enables users to upload their documents and receive accurate, context-grounded answers through a retrieval-augmented generation (RAG) pipeline. Unlike cloud-based AI services, Vai processes all data locally — no document content ever leaves the user's infrastructure.

The system is built on Go, Ollama (qwen3.5:4b + nomic-embed-text:v1.5), Qdrant, and PostgreSQL — all open-source, all self-hostable.

---

## Problem Statement

Organizations and developers require intelligent document Q&A capabilities but face critical barriers with existing solutions:

| Problem | Impact |
|---------|--------|
| **Privacy risk** | Existing tools (ChatGPT, Claude, Gemini) transmit user documents to third-party servers |
| **Data sovereignty** | Regulated industries (healthcare, finance, legal) cannot use cloud AI for sensitive documents |
| **Hallucination** | Generic AI answers without grounding in specific documents are unreliable |
| **Vendor lock-in** | Dependency on proprietary APIs creates cost and availability risks |
| **Cost at scale** | Cloud LLM API costs grow proportionally with usage |

---

## Product Vision

> *"Vai gives every team the power of AI-driven document intelligence, entirely on their own terms — no cloud, no compromise, no black box."*

---

## Target Users

| Persona | Context | Primary Need |
|---------|---------|--------------|
| Software Developers | Individual or team | Self-host on local machine, API integration |
| DevOps / Platform Engineers | Enterprise | Deploy via Docker/Kubernetes, manage infrastructure |
| Privacy-Conscious Organizations | Healthcare, Finance, Legal | Query sensitive documents without external exposure |
| AI/ML Researchers | Academic/Enterprise | Experiment with RAG pipelines, swap models |
| Knowledge Workers | Any org | Query internal documentation, manuals, policies |

---

## Functional Requirements

### FR-AUTH — Authentication & User Management

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-AUTH-01 | Users register with email + password. Passwords stored as bcrypt hashes (cost ≥ 12). | P0 |
| FR-AUTH-02 | JWT issued on login, stored as HTTP-only Secure SameSite=Strict cookie (15-min expiry) + refresh token (7-day expiry in DB). | P0 |
| FR-AUTH-03 | OAuth 2.0 with Google (OpenID Connect) for single-sign-on. | P0 |
| FR-AUTH-04 | Email verification required before document uploads are permitted. | P0 |
| FR-AUTH-05 | Password reset via HMAC-signed, time-limited token (1-hour expiry) emailed to the user. | P0 |
| FR-AUTH-06 | Refresh token rotation — invalidate old token on use and reissue a new one. | P0 |
| FR-AUTH-07 | Logout revokes the current refresh token and clears both cookies. | P0 |
| FR-AUTH-08 | Multiple active refresh tokens per user supported (multi-device login). | P1 |

### FR-DOC — Document Management

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-DOC-01 | Users upload plain-text documents via multipart/form-data POST. | P0 |
| FR-DOC-02 | Documents are chunked with configurable size (default 500 chars) and overlap (default 100 chars). | P0 |
| FR-DOC-03 | Each chunk is embedded via Ollama (nomic-embed-text:v1.5) and stored in Qdrant under the user's namespace. | P0 |
| FR-DOC-04 | Users can list, view metadata for, and delete their documents. | P1 |
| FR-DOC-05 | Document IDs are deterministic slugs derived from the filename. | P1 |
| FR-DOC-06 | Deleting a document removes both the PostgreSQL metadata record and all Qdrant vectors. | P1 |
| FR-DOC-07 | Maximum file size: 10MB. Supported types: .txt (v1.0); .pdf, .docx (v1.1). | P0 |

### FR-CHAT — Chat & Q&A

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-CHAT-01 | Users submit natural-language questions via POST /chat. | P0 |
| FR-CHAT-02 | System retrieves top-K semantically similar chunks from Qdrant (default K=5). | P0 |
| FR-CHAT-03 | Retrieved chunks are assembled into a context-enriched prompt and sent to the LLM. | P0 |
| FR-CHAT-04 | Chat sessions are persisted per user with full message history in PostgreSQL. | P0 |
| FR-CHAT-05 | Streaming responses delivered via Server-Sent Events (GET /chat/stream). | P0 |
| FR-CHAT-06 | Users can filter Q&A to a specific document or query across all documents. | P1 |
| FR-CHAT-07 | Users can create, list, view, and delete chat sessions. | P1 |
| FR-CHAT-08 | Auto-generate a session title from the first question if not provided. | P2 |

### FR-EMAIL — Email Service

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-EMAIL-01 | Transactional email sent for account verification on registration. | P0 |
| FR-EMAIL-02 | Password reset email with 1-hour expiry HMAC token and a secure link. | P0 |
| FR-EMAIL-03 | Welcome email on first OAuth login. | P2 |
| FR-EMAIL-04 | SMTP provider configurable via environment variables (host, port, credentials). | P0 |

### FR-SEARCH — Debug Search

| ID | Requirement | Priority |
|----|-------------|----------|
| FR-SEARCH-01 | POST /search returns raw top-K chunks from Qdrant without LLM generation. | P1 |
| FR-SEARCH-02 | Search results include chunk text, document ID, and similarity score. | P1 |

---

## Non-Functional Requirements

| ID | Category | Attribute | Requirement |
|----|----------|-----------|-------------|
| NFR-PERF-01 | Performance | API latency | P95 < 200ms for all non-LLM endpoints |
| NFR-PERF-02 | Performance | Streaming TTFB | < 1s to first SSE token |
| NFR-PERF-03 | Performance | Ingestion throughput | 1MB file fully ingested in < 10s |
| NFR-SEC-01 | Security | Token storage | JWT in HTTP-only Secure SameSite=Strict cookie |
| NFR-SEC-02 | Security | Password hashing | bcrypt with cost factor ≥ 12 |
| NFR-SEC-03 | Security | Transport | TLS 1.2+ enforced in production |
| NFR-SEC-04 | Security | CSRF | SameSite=Strict cookie + OAuth state param validation |
| NFR-SEC-05 | Security | Rate limiting | Auth endpoints: max 20 req/min per IP |
| NFR-PRIV-01 | Privacy | Data residency | Zero external API calls during document processing |
| NFR-PRIV-02 | Privacy | User isolation | Qdrant collections namespaced per user (`user_{userID}`) |
| NFR-PRIV-03 | Privacy | Logging | No document content logged, only metadata |
| NFR-REL-01 | Reliability | Uptime | 99.5% monthly uptime target |
| NFR-SCAL-01 | Scalability | Concurrency | Support 1,000 concurrent users |
| NFR-SCAL-02 | Scalability | Documents per user | Support up to 10,000 documents per user |
| NFR-MAIN-01 | Maintainability | Test coverage | ≥ 70% unit + integration test coverage |
| NFR-DEPLOY-01 | Deployability | Setup time | Full stack running in < 10 minutes via `docker compose up` |

---

## User Stories

| Story ID | Persona | Story | Acceptance |
|----------|---------|-------|------------|
| US-001 | Developer | As a developer, I want to upload a text document so that I can query its contents in plain language. | Upload returns document_id + chunk count within 10s |
| US-002 | Privacy-conscious user | As a privacy-conscious user, I want all processing to stay on my server so that my documents are never exposed to third parties. | Network monitor shows zero external AI API calls |
| US-003 | Registered user | As a registered user, I want to log in with Google so that I don't need to manage a separate password. | OAuth flow completes, JWT issued, user record created |
| US-004 | User | As a user, I want to receive a streaming response so that I see answers progressively without waiting for the full reply. | SSE events begin within 1 second of request |
| US-005 | Admin | As a system operator, I want users to verify their email before uploading documents so that I can prevent spam and abuse. | Unverified users receive 403 on POST /documents/upload |
| US-006 | User | As a user, I want to reset my password via email so that I can recover access if I forget it. | Reset link works once, expires in 1 hour |
| US-007 | User | As a user, I want to see all my past conversations so that I can refer back to prior answers. | GET /chat/sessions returns all sessions with last message timestamp |
| US-008 | Developer | As a developer, I want a debug search endpoint so that I can inspect exactly what chunks are retrieved for a given query. | POST /search returns chunks with scores without calling the LLM |
| US-009 | User | As a user, I want to delete a document so that its content is no longer searchable and is fully removed from the system. | DELETE removes PostgreSQL record + all Qdrant vectors |
| US-010 | User | As a user, I want to filter my questions to a specific document so that I get focused answers from that source only. | document_id param restricts Qdrant search to that document's vectors |

---

## Acceptance Criteria

### Authentication
- All FR-AUTH requirements pass integration tests with the full JWT lifecycle (issue → use → refresh → revoke).
- Attempting to use a revoked refresh token returns 401.
- Google OAuth flow creates a new user on first login and reuses the existing user on subsequent logins (matched by email).
- Password reset invalidates all existing refresh tokens for the user.

### Document Management
- Document upload returns correct chunk count and document_id within 10 seconds for files up to 1MB.
- Deleting a document verifies that subsequent search queries return zero results for that document's content.

### Chat & RAG
- Chat responses are grounded — the LLM system prompt instructs it to answer only from retrieved context chunks.
- Streaming endpoint emits valid SSE format (`data: <token>\n\n`) for every token and terminates with `data: [DONE]\n\n`.

### Email
- Email delivery tested in staging via SMTP sandbox (e.g., Mailtrap).
- Verification email link expires after 24 hours; second click returns a clear error.

### Privacy
- No document text appears in application logs at any log level.
- Qdrant collections are correctly scoped — a user querying another user's document_id receives zero results.
