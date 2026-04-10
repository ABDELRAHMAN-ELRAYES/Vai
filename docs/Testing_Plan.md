# Testing Plan — Vai (Privacy-First AI Assistant)
**Version:** 1.0  
**Date:** April 2026  

## 1. Executive Summary
This document outlines the testing strategy for Vai, a self-hosted AI document assistant. As the project currently lacks automated tests, this plan establishes a foundation for achieving >= 70% coverage as per NFR requirements, ensuring reliability, security, and performance.

---

## 2. Test Strategy & Levels

### 2.1. Unit Testing
**Objective:** Verify individual functions and logic in isolation.
- **Backend (Go):**
    - **Tools:** Standard `testing` package, `testify`.
    - **Scope:**
        - Domain modules (`internal/modules/*`).
        - Core engine (`internal/rag-engine`, `internal/chunker`).
        - Service layers: Business logic with mocked repositories.
- **Frontend (React):**
    - **Tools:** `Vitest`, `React Testing Library`.
    - **Scope:**
        - UI Components (Buttons, Inputs, Modals).
        - Custom Hooks (`useChat`, [useAuth](file:///home/abdelrahman/elrayes/2026/projects/Vai/web/src/hooks/auth/use-auth.ts#5-10)).
        - Form validation logic (Zod).

### 2.2. Integration Testing
**Objective:** Verify interaction between components and external systems (PostgreSQL, Qdrant, Ollama).
- **Tooling:** `httptest` for API, `testcontainers-go` (optional) or a dedicated test database environment.
- **Key Scenarios:**
    - **Authentication Flow:** Registration -> Email Verification -> Login -> Token Rotation.
    - **Document Pipeline:** Multipart upload -> Text extraction -> Chunking -> Qdrant upsert -> Metadata persistence.
    - **RAG Flow:** Query embedding -> Vector search -> Prompt assembly -> LLM streaming (using actual local Ollama or mock).
    - **Database Repositories:** Verifying SQL queries and migrations.

### 2.3. End-to-End (E2E) Testing
**Objective:** Validate full user journeys in a real browser environment.
- **Tools:** `Playwright`.
- **Primary Flows:**
    1. **Onboarding:** Register new account -> Confirm email (mocked) -> Login.
    2. **Document Management:** Upload a `.txt` file -> Verify it appears in the list -> Delete it.
    3. **AI Chat:** Upload doc -> Start session -> Ask a question -> Verify streaming response content.
    4. **Session Persistence:** Refresh page -> Verify chat history is restored.

---

## 3. Specialized Testing

### 3.1. RAG & AI Testing
- **Retrieval Accuracy:** Test if `top-K` search returns expected chunks for known queries.
- **Prompt Injection:** Verify that system prompts prevent the LLM from answering outside the provided context.
- **Streaming Reliability:** Verify SSE format compliance and connection resilience.

### 3.2. Security Testing
- **User Isolation:** Attempt to access/delete documents using a valid JWT from a *different* user.
- **JWT Protection:** Verify `HttpOnly` and `SameSite=Strict` flags.
- **Rate Limiting:** Scripted bombardment of `/auth/login` to confirm `429` responses.
- **CSRF:** Verify state parameter validation in Google OAuth flow.
- **Advanced Access Control (Enterprise/Future):**
    - **Team Mode Isolation:** Verify that "Managers" can see team data while "Operators" are restricted to their own.
    - **RBAC:** Test role-based permissions for admin actions (e.g., system-wide document management).

### 3.3. Performance Testing (NFR Verification)
- **Ingestion Latency:** Benchmark 1MB and 10MB file processing times (Target: 1MB < 10s).
- **Streaming TTFB:** Measure time from request to first SSE token (Target: < 1s).
- **Concurrent Load:** Use `k6` to simulate 1,000 concurrent users on the API.

---

## 4. Test Environment & Automation
### 4.1. Local Development
- **Backend:** `make test` (targeted `go test -v ./...`).
- **Frontend:** `npm run test` (Vitest).

### 4.2. CI/CD Pipeline (GitHub Actions)
1. **Linting:** `golangci-lint` and `eslint`.
2. **Unit/Integration:** Parallel jobs for backend and frontend tests.
3. **Security Scan:** `gosec` for Go and `npm audit` for frontend dependencies.
4. **E2E Trace:** Playwright snapshots on failure.

### 4.3. Continuous Performance Monitoring
- Integration of `k6` in the pipeline for regression testing of core API latencies.

---

---

## 5. Test Case Inventory (Sample)

| ID | Module | Scenario | Expected Result |
|----|--------|----------|-----------------|
| TC-01 | Auth | Login with incorrect password | Returns `401 Unauthorized` |
| TC-02 | Doc | Upload > 10MB file | Returns `413 Request Entity Too Large` |
| TC-03 | Chat | Ask question without documents | Returns grounded "I don't know" or similar |
| TC-04 | User | Delete account | All related Docs and Chat history purged from DB + Qdrant |
| TC-05 | AI | SSE stream interruption | Client can gracefully handle or reconnect |
