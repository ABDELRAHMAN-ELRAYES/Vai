# Project Charter · Roadmap · OKRs
## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Date:** June 2025  
**Project Sponsor:** Engineering Leadership  
**Project Manager:** Lead Software Architect

---

## Table of Contents

1. [Project Charter](#project-charter)
2. [Product Roadmap](#product-roadmap)
3. [OKRs](#okrs)

---

## Project Charter

### Project Overview

| Field | Value |
|-------|-------|
| **Project Name** | Vai — Privacy-First AI Document Assistant |
| **Project Sponsor** | Engineering Leadership |
| **Project Manager** | Lead Software Architect |
| **Start Date** | January 2025 |
| **MVP Target** | Q2 2025 |
| **v1.0 GA Release** | Q3 2025 |
| **Budget** | Internal engineering time + infrastructure |
| **Priority** | High |

### Project Purpose

Vai is chartered to deliver a production-grade, self-hosted Retrieval-Augmented Generation (RAG) platform that enables individuals and organizations to query their own documents using AI — without transmitting any data to external servers. The project addresses a critical gap in the AI tooling market: capable document intelligence that respects data sovereignty.

### Scope

#### In Scope (v1.0)

- Go-based HTTP API server with all document, chat, and auth endpoints
- JWT + HTTP-only cookie authentication with refresh token rotation
- OAuth 2.0 / Google Sign-In integration (OpenID Connect)
- Email verification and password reset flows (SMTP)
- RAG pipeline: text chunking, embedding, Qdrant storage, LLM generation + streaming
- PostgreSQL schema for users, sessions, tokens, chat history, document metadata
- Qdrant for vector storage and semantic similarity search
- Per-user data isolation (namespaced Qdrant collections)
- Docker Compose deployment (single-command)
- REST API documentation (OpenAPI 3.0 spec)
- Unit and integration test suite

#### Out of Scope (Deferred)

| Feature | Target Version |
|---------|---------------|
| Web UI / frontend chat application | v1.2 |
| PDF and DOCX file parsing | v1.1 |
| Kubernetes Helm chart | v1.2 |
| Multi-model support (HuggingFace, OpenAI-compatible) | v1.2 |
| Response caching | v1.3 |
| Multi-tenant billing / SaaS mode | v2.0 |
| Admin dashboard | v2.0 |

### Deliverables

| # | Deliverable | Owner | Due |
|---|-------------|-------|-----|
| D-01 | Go backend service (all API endpoints implemented) | Backend Team | Q2 2025 |
| D-02 | PostgreSQL migrations (all tables, indexes) | Backend Eng 2 | Feb 2025 |
| D-03 | Docker Compose file (all 4 services) | DevOps | Apr 2025 |
| D-04 | OpenAPI 3.0 specification | Lead Architect | Jun 2025 |
| D-05 | README (setup, config, usage guide) | All | Jun 2025 |
| D-06 | Test suite (≥ 70% coverage) | Backend Team | May 2025 |
| D-07 | Architecture + engineering documentation | Lead Architect | Jun 2025 |

### Project Timeline

| Phase | Month | Duration | Focus |
|-------|-------|----------|-------|
| **Phase 0 — Design** | Jan 2025 | 2 weeks | Architecture decisions, DB schema, API contract, tech stack finalization |
| **Phase 1 — RAG Core** | Feb 2025 | 4 weeks | Chunker, embeddings client, Qdrant integration, RAG pipeline (ingest + query) |
| **Phase 2 — Auth** | Mar 2025 | 4 weeks | JWT, cookie handling, OAuth Google, email verification, password reset |
| **Phase 3 — Users & Chat** | Apr 2025 | 3 weeks | User model, chat sessions, message persistence, conversation history |
| **Phase 4 — Deployment** | May 2025 | 2 weeks | Docker Compose, integration tests, environment configuration, hardening |
| **Phase 5 — Release** | Jun 2025 | 2 weeks | API spec, security review, documentation, v1.0 tag and release |

### Team & Responsibilities

| Role | Responsibilities |
|------|-----------------|
| **Lead Architect** | System design, code review, architecture decisions, cross-cutting concerns, documentation |
| **Backend Engineer 1** | RAG pipeline, embedding service, Qdrant client, chunker module |
| **Backend Engineer 2** | Auth system (JWT, OAuth, email service), PostgreSQL user/token models |
| **Backend Engineer 3** | API handlers, chat service, session management, conversation history |
| **DevOps Engineer** | Docker Compose, CI/CD pipeline, deployment scripts, environment configuration |

### Success Criteria

The project is considered successful when:
1. All functional requirements (FR-AUTH, FR-DOC, FR-CHAT, FR-EMAIL) pass their integration tests
2. A new user can go from `docker compose up` to first successful chat query in under 15 minutes
3. All AI inference is confirmed to run with zero external network calls
4. Test coverage is ≥ 70%
5. OpenAPI spec covers 100% of public endpoints

### Risks & Escalation

| Risk | Owner | Escalation |
|------|-------|------------|
| Model quality insufficient | Backend Eng 1 | Swap model, re-evaluate at Phase 3 |
| Auth security vulnerability | Backend Eng 2 | External security review, delay release |
| Hardware resource issues | DevOps | Document requirements, add lighter model option |

---

## Product Roadmap

### Version History & Milestones

```
Q1 2025         Q2 2025         Q3 2025         Q4 2025         Q1 2026
   |               |               |               |               |
[v0.9 MVP]──────[v1.0 GA]───────[v1.1 Files]────[v1.2 UI]───────[v1.3 Scale]
```

### Version Details

| Version | Target | Status | Key Features |
|---------|--------|--------|--------------|
| **v0.9 — MVP** | Q1 2025 | ✅ Done | Core RAG pipeline, basic upload + chat API, Qdrant + Ollama integration |
| **v1.0 — GA** | Q2 2025 | 🔄 In Progress | Auth (JWT + OAuth), email, multi-user isolation, chat history, Docker Compose |
| **v1.1 — Files** | Q3 2025 | 📋 Planned | PDF/DOCX parsing, document list/delete endpoints, re-indexing |
| **v1.2 — UI** | Q4 2025 | 📋 Planned | Web chat UI, conversation view, document library, Kubernetes Helm chart |
| **v1.3 — Scale** | Q1 2026 | 🗺️ Roadmap | Multi-embedding backends (HuggingFace), response caching, monitoring dashboard |
| **v2.0 — Platform** | Q2 2026 | 🗺️ Roadmap | Multi-tenant SaaS mode, API keys, usage billing, admin dashboard |

### Feature Backlog by Priority

#### P0 — Must Have (v1.0)
- [ ] JWT authentication with HTTP-only cookie
- [ ] Refresh token rotation
- [ ] Google OAuth 2.0 sign-in
- [ ] Email verification on registration
- [ ] Password reset via email
- [ ] User-isolated Qdrant collections (`user_{userID}`)
- [ ] Chat session persistence
- [ ] Chat message history

#### P1 — Should Have (v1.1)
- [ ] PDF ingestion (pdftotext / pdfminer)
- [ ] DOCX ingestion (mammoth / docx2txt)
- [ ] Document list endpoint (`GET /documents`)
- [ ] Document delete endpoint (`DELETE /documents/{id}`)
- [ ] Rate limiting on auth endpoints
- [ ] Request ID middleware + structured logging

#### P2 — Nice to Have (v1.2)
- [ ] Web chat UI (React or HTMX)
- [ ] Multi-model embedding support (HuggingFace Inference API)
- [ ] Response caching for repeated questions
- [ ] Monitoring dashboard (latency, errors, token usage)
- [ ] Kubernetes Helm chart

#### P3 — Future (v2.0+)
- [ ] Multi-tenant billing
- [ ] API key management
- [ ] Admin user panel
- [ ] SAML / LDAP SSO
- [ ] Audit logging

---

## OKRs

### Q2 2025 OKRs

---

#### Objective 1: Ship a production-ready, secure authentication system

*Why it matters: Auth is the foundation of multi-user trust and privacy guarantees.*

| Key Result | Target | Measurement |
|------------|--------|-------------|
| KR1.1 | JWT + refresh token flow passes 100% of auth integration tests | CI pipeline green |
| KR1.2 | Zero known security vulnerabilities in auth implementation | Security review sign-off |
| KR1.3 | Google OAuth login completes end-to-end in < 3 seconds | Timed test on standard hardware |
| KR1.4 | Email verification and password reset tested with 3+ SMTP providers | QA test report |

---

#### Objective 2: Deliver complete multi-user RAG system with strict data isolation

*Why it matters: User data isolation is the core privacy promise of Vai.*

| Key Result | Target | Measurement |
|------------|--------|-------------|
| KR2.1 | Zero cross-user data leakage verified by test suite | Dedicated isolation tests pass |
| KR2.2 | Chat history retrieved in < 50ms for up to 1,000 messages | Performance benchmark |
| KR2.3 | 10 concurrent document uploads complete without error | Load test |
| KR2.4 | Each user's Qdrant collection correctly namespaced and scoped | Integration test |

---

#### Objective 3: Achieve developer-friendly deployment and documentation

*Why it matters: Adoption depends on how quickly a new user can go from zero to value.*

| Key Result | Target | Measurement |
|------------|--------|-------------|
| KR3.1 | `docker compose up` deploys full stack in under 5 minutes | Timed from fresh machine |
| KR3.2 | OpenAPI 3.0 spec covers 100% of public endpoints with examples | Spec coverage audit |
| KR3.3 | New developer achieves first chat query in < 15 minutes following README | User test with 3 developers |
| KR3.4 | Test coverage ≥ 70% | `go test -cover` output |

---

### Q3 2025 OKRs

---

#### Objective 4: Expand document format support to cover common enterprise files

*Why it matters: Most enterprise documents are PDFs or Word files, not plain text.*

| Key Result | Target | Measurement |
|------------|--------|-------------|
| KR4.1 | PDF text extraction accuracy > 95% vs. manual extraction on 20-doc benchmark | Automated accuracy test |
| KR4.2 | DOCX ingestion working for standard Word documents | Integration test suite |
| KR4.3 | Document management endpoints (list, delete, re-index) respond in < 100ms | API benchmark |

---

#### Objective 5: Build open-source community traction

*Why it matters: Community adoption validates product-market fit and accelerates development.*

| Key Result | Target | Measurement |
|------------|--------|-------------|
| KR5.1 | 500+ GitHub stars within 60 days of v1.0 release | GitHub metrics |
| KR5.2 | 10+ community contributors open pull requests | GitHub contributor graph |
| KR5.3 | Published technical blog post achieves 2,000+ reads | Analytics |
| KR5.4 | Featured in at least 2 developer newsletters / communities | Media mentions |
