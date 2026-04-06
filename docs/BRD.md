# Business Requirements Document (BRD)

## Vai — Privacy-First AI Document Assistant

**Version:** 1.0  
**Status:** Approved  
**Date:** June 2025  
**Sponsor:** Engineering Leadership

---

## Table of Contents

1. [Business Purpose](#business-purpose)
2. [Business Objectives](#business-objectives)
3. [Business Drivers](#business-drivers)
4. [Stakeholders](#stakeholders)
5. [Business Constraints](#business-constraints)
6. [Business Assumptions](#business-assumptions)
7. [Business Risks](#business-risks)
8. [Success Metrics](#success-metrics)
9. [Cost-Benefit Summary](#cost-benefit-summary)

---

## Business Purpose

The Vai project addresses a growing market demand for AI document intelligence that does not compromise organizational data privacy. As AI adoption accelerates, enterprises in regulated sectors are blocked from using public AI APIs due to compliance requirements (HIPAA, GDPR, SOC 2, ISO 27001). Vai removes this blocker by delivering the full capability of a RAG pipeline entirely on-premises.

The business case is straightforward: organizations that need AI document Q&A today must choose between capability and compliance. Vai eliminates that tradeoff.

---

## Business Objectives

| #     | Objective                                                                | Target Date   | Owner        |
| ----- | ------------------------------------------------------------------------ | ------------- | ------------ |
| BO-01 | Deliver a production-ready self-hosted RAG platform                      | Q3 2025       | Engineering  |
| BO-02 | Enable AI document Q&A with zero third-party data exposure               | Q2 2025 (MVP) | Architecture |
| BO-03 | Provide a RESTful API that integrates with existing developer workflows  | Q2 2025       | Backend      |
| BO-04 | Support multi-user environments with per-user isolation and chat history | Q2 2025       | Backend      |
| BO-05 | Achieve single-command Docker deployment                                 | Q1 2025       | DevOps       |
| BO-06 | Establish an open-source foundation with community extensibility         | Q3 2025       | Product      |

---

## Business Drivers

| Driver                    | Description                                                                                                                                                                   | Priority |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| **Privacy Regulation**    | GDPR, HIPAA, SOC 2, and ISO 27001 compliance requirements block cloud AI for sensitive data. Organizations face fines and legal liability for unauthorized data transmission. | Critical |
| **Data Sovereignty**      | Organizations need full, auditable control over where their data is stored and processed. This is especially relevant for government, defense, and financial institutions.    | Critical |
| **Cost Optimization**     | Cloud LLM API costs scale with usage (per-token pricing). Self-hosted inference is a fixed infrastructure cost. At scale, the ROI is highly favorable.                        | High     |
| **AI Hallucination Risk** | RAG grounding reduces incorrect, fabricated answers compared to pure LLM responses, increasing trust in AI-assisted workflows.                                                | High     |
| **Developer Experience**  | Clean REST API, Docker deployment, and comprehensive documentation lower the adoption barrier and reduce integration time.                                                    | Medium   |
| **Vendor Independence**   | No dependency on a single AI provider. Models can be swapped via configuration, reducing lock-in risk.                                                                        | Medium   |

---

## Stakeholders

| Stakeholder                            | Role                                                    | Interest                                  | Influence |
| -------------------------------------- | ------------------------------------------------------- | ----------------------------------------- | --------- |
| **Product Owner**                      | Defines priorities, accepts deliverables, owns roadmap  | Feature completeness, delivery timeline   | High      |
| **Software Architect**                 | Designs system, reviews decisions, writes documentation | Technical correctness, scalability        | High      |
| **Backend Engineers (3)**              | Implement Go services, API handlers, RAG pipeline       | Clear requirements, technical autonomy    | High      |
| **DevOps Engineer**                    | Manages Docker, CI/CD, deployment                       | Infrastructure simplicity, reliability    | Medium    |
| **End Users — Developers**             | Integrate Vai via REST API                              | API quality, documentation, performance   | Medium    |
| **End Users — Organizations**          | Deploy internally for document Q&A                      | Privacy guarantees, ease of deployment    | Medium    |
| **Open Source Community**              | Contribute features, report bugs                        | Code quality, contribution guidelines     | Low       |
| **Legal / Compliance (if enterprise)** | Review data handling                                    | Zero external data transmission guarantee | High      |

---

## Business Constraints

| ID    | Constraint                                                                                | Rationale                                     |
| ----- | ----------------------------------------------------------------------------------------- | --------------------------------------------- |
| BC-01 | All AI inference must run locally — no calls to OpenAI, Anthropic, or any external AI API | Core privacy value proposition                |
| BC-02 | Initial release targets Linux/macOS Docker environments                                   | Broadest developer adoption, simplest testing |
| BC-03 | MVP must be deployable with a single `docker compose up` command                          | Reduce time-to-value for new users            |
| BC-04 | Ollama is the sole LLM/embedding provider in v1.0                                         | Simplicity; multi-provider planned for v1.2   |
| BC-05 | No paid external services in the critical path                                            | Self-hosted means fully self-contained        |
| BC-06 | Go is the required backend language                                                       | Performance, concurrency, team expertise      |
| BC-07 | PostgreSQL and Qdrant are the only databases in v1.0                                      | Proven stack, excellent Go clients            |

---

## Business Assumptions

| ID    | Assumption                                                                                               |
| ----- | -------------------------------------------------------------------------------------------------------- |
| BA-01 | Users have Docker Desktop or Docker Engine installed and available                                       |
| BA-02 | Users have Ollama installed and can pull the required models (llama2.3:3b, nomic-embed-text:v1.5)        |
| BA-03 | The target deployment environment has at minimum 8GB RAM and 20GB available disk space                   |
| BA-04 | Email delivery uses an SMTP provider (self-hosted or third-party) configurable via environment variables |
| BA-05 | Google OAuth credentials (client ID + secret) are provisioned separately by the deploying team           |
| BA-06 | The primary document format in v1.0 is plain text (.txt); binary format support is deferred              |
| BA-07 | Users are responsible for the security of their own infrastructure (network, OS hardening)               |
| BA-08 | Qdrant and PostgreSQL data persistence is managed via Docker volumes by the operator                     |

---

## Business Risks

| Risk ID | Description                                              | Likelihood | Impact   | Mitigation Strategy                                                             |
| ------- | -------------------------------------------------------- | ---------- | -------- | ------------------------------------------------------------------------------- |
| BR-001  | LLM output quality insufficient for production use cases | Medium     | High     | Allow model swap via config; document recommended models and prompts            |
| BR-002  | Hardware requirements too high for target users          | Medium     | Medium   | Document minimum specs; test on CPU-only; provide lighter model options         |
| BR-003  | Qdrant or Ollama API changes break integration           | Low        | High     | Pin dependency versions in Docker Compose; maintain integration test suite      |
| BR-004  | Email deliverability issues in production                | Medium     | Medium   | Support multiple SMTP providers; document SPF/DKIM setup                        |
| BR-005  | JWT/auth security vulnerability                          | Low        | Critical | Security code review; automated auth tests; follow OWASP JWT best practices     |
| BR-006  | PostgreSQL data loss (no backup strategy)                | Low        | High     | Document backup procedures; operator responsibility; future: automated backups  |
| BR-007  | Open-source repo attracts malicious pull requests        | Low        | Medium   | Require signed commits; PR review policy; branch protection rules               |
| BR-008  | User uploads excessively large documents causing OOM     | Medium     | Medium   | Enforce 10MB file size limit; stream processing rather than full in-memory load |

---

## Success Metrics

| Metric                                                | Target       | Measurement Method                       |
| ----------------------------------------------------- | ------------ | ---------------------------------------- |
| Time-to-first-answer (new user: upload → ask)         | < 3 minutes  | Manual QA testing                        |
| Document retrieval precision (correct chunk in top-3) | > 80%        | Benchmark on test corpus                 |
| Zero external network calls during ingestion + Q&A    | 100%         | Network capture during integration tests |
| Setup time (zero to running service)                  | < 10 minutes | Timed walkthrough with new developer     |
| Test coverage                                         | ≥ 70%        | `go test -cover`                         |
| GitHub stars (60 days post v1.0 release)              | ≥ 500        | GitHub repository metrics                |
| Community contributors (PRs opened)                   | ≥ 10         | GitHub contributor graph                 |

---

## Cost-Benefit Summary

### Cost (Infrastructure, Self-Hosted)

| Item                                             | Estimated Cost      |
| ------------------------------------------------ | ------------------- |
| VPS / bare-metal server (16GB RAM, GPU optional) | $20–$80/month       |
| Storage (PostgreSQL + Qdrant volumes)            | $5–$20/month        |
| SMTP service (e.g., Mailgun free tier)           | $0–$10/month        |
| Development time (initial)                       | 4–6 engineer-months |
| **Total operational**                            | **$25–$110/month**  |

### Cost Avoided (vs. Cloud AI)

| Scenario                                     | Cloud AI Cost (est.) | Vai Cost        |
| -------------------------------------------- | -------------------- | --------------- |
| 100 users × 50 queries/day × 1K tokens avg   | ~$150/month (GPT-4o) | $0 additional   |
| 1,000 users × 50 queries/day × 1K tokens avg | ~$1,500/month        | $0 additional   |
| Enterprise (10K queries/day)                 | ~$15,000/month       | $80/month infra |

**Break-even:** Approximately 50–200 active users depending on query volume and model choice.
