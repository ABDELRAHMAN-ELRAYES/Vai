# Vai

**Vai** is a self-hosted, privacy-first AI document assistant. Upload your documents, ask questions in plain language, and get accurate answers grounded in your own content — no cloud APIs, no data leaving your machine.

Built with Go, Ollama(llama2.3:3b), Qdrant, and open-source embedding model(nomic-embed-text:v1.5), Vai gives you the full power of a retrieval-augmented generation (RAG) system that you own and control entirely.

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [How It Works](#how-it-works)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the Stack](#running-the-stack)
- [API Reference](#api-reference)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Most AI assistants send your data to third-party servers. Vai is different. Every component — the language model, the embedding model, and the vector database — runs locally on your own infrastructure. Your documents never leave your environment.

Vai is designed for developers, teams, and organizations that need AI-powered document search and Q&A without compromising on privacy or data sovereignty.

---

## Features

- **Fully self-hosted** — no external API calls, no data sent to third parties
- **Semantic search** — find relevant content by meaning, not just keywords
- **RAG pipeline** — answers are grounded in your documents, not hallucinated
- **Streaming responses** — real-time token streaming via Server-Sent Events
- **Multi-document support** — upload multiple documents and query across all of them or filter to a specific one
- **Overlap-aware chunking** — intelligent document splitting that preserves context at boundaries
- **REST API** — clean HTTP interface, easy to integrate with any frontend
- **Docker-ready** — full containerized deployment with a single command

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend / Client                     │
│                  (chat UI, curl, any HTTP client)            │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                        Backend API (Go)                      │
│         Upload · Chat · Stream · Search endpoints            │
└────────┬─────────────────────┬───────────────────┬──────────┘
         │                     │                   │
         ▼                     ▼                   ▼
┌─────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│  Chunker        │  │  Embedding Model │  │  LLM (Ollama)    │
│  (text splitter)│  │  (nomic-embed-text:v1.5)   │  │  (llama2.3:3b)        │
└────────┬────────┘  └────────┬─────────┘  └──────────────────┘
         │                    │
         ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                     Qdrant Vector Database                   │
│               (stores and searches embeddings)               │
└─────────────────────────────────────────────────────────────┘
```

---

## How It Works

Vai operates across two distinct workflows.

### Document Ingestion

When you upload a document, Vai processes it through a pipeline before anything is stored:

```
Raw document
    ↓
Split into overlapping chunks (~500 chars each, 100-char overlap)
    ↓
Each chunk passed through the embedding model
    ↓
Each chunk becomes a vector of 768 numbers representing its meaning
    ↓
Vectors stored in Qdrant alongside the original chunk text and metadata
```

The overlap between chunks is intentional — a sentence at the boundary of two chunks appears in both, so no context is lost at split points.

### Question Answering

When you ask a question, Vai runs a shorter version of the same process:

```
User question
    ↓
Question embedded using the same model (produces a vector)
    ↓
Qdrant finds the top-K most semantically similar stored chunks
    ↓
Chunks assembled into a prompt alongside the question
    ↓
LLM generates an answer grounded in those chunks
    ↓
Response streamed token-by-token back to the client
```

The LLM never sees the full document — only the retrieved chunks most relevant to the question. This keeps responses fast, accurate, and within the model's context window.

---

## Tech Stack

| Component  | Technology                     | Purpose                            |
| ---------- | ------------------------------ | ---------------------------------- |
| Backend    | Go                             | API server, pipeline orchestration |
| LLM        | Ollama + llama2.3:3b           | Answer generation                  |
| Embeddings | Ollama + nomic-embed-text:v1.5 | Semantic vector generation         |
| Vector DB  | Qdrant                         | Similarity search                  |
| Deployment | Docker + Docker Compose        | Containerized infrastructure       |

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Go 1.22+](https://go.dev/dl/) (for local development)
- [Ollama](https://ollama.com/) installed and running

### Installation

```bash
# Clone the repository
git clone https://github.com/yourname/vai.git
cd vai

# Install Go dependencies
go mod tidy
```

### Running the Stack

**1. Start Qdrant**

```bash
docker run -p 6333:6333 qdrant/qdrant
```

**2. Start Ollama and pull the required models**

```bash
ollama serve

# In a separate terminal:
ollama pull llama2.3:3b              # language model for generating answers
ollama pull nomic-embed-text:v1.5    # embedding model for semantic search
```

**3. Start the Vai server**

```bash
air
```

The server will be available at `http://localhost:8080`.

**Or run everything with Docker Compose**

```bash
docker compose up
```

---

## API Reference

### Upload a Document

```
POST /documents/upload
Content-Type: multipart/form-data
```

```bash
curl -X POST http://localhost:8080/documents/upload \
  -F "file=@/path/to/document.txt"
```

**Response**

```json
{
  "document_id": "document",
  "chunks": 14,
  "source": "document.txt"
}
```

---

### Ask a Question

```
POST /chat
Content-Type: application/json
```

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How does authentication work?",
    "top_k": 5
  }'
```

**Request Body**

| Field         | Type   | Required | Description                               |
| ------------- | ------ | -------- | ----------------------------------------- |
| `question`    | string | yes      | The question to answer                    |
| `top_k`       | int    | no       | Number of chunks to retrieve (default: 5) |
| `document_id` | string | no       | Filter search to a specific document      |

**Response**

```json
{
  "answer": "Based on the documents, authentication works by..."
}
```

---

### Stream a Response

```
GET /chat/stream?question=...&top_k=5&document_id=...
```

```bash
curl -N "http://localhost:8080/chat/stream?question=How+does+auth+work&top_k=5"
```

Returns a stream of Server-Sent Events (SSE):

```
data: Based
data:  on
data:  the
data:  documents...
data: [DONE]
```

---

### Search (Debug)

Returns raw retrieved chunks without LLM generation — useful for inspecting what the retrieval layer finds.

```
POST /search
Content-Type: application/json
```

```bash
curl -X POST http://localhost:8080/search \
  -H "Content-Type: application/json" \
  -d '{"query": "authentication", "top_k": 3}'
```

---

## Configuration

Configuration is set in `main.go` via the `rag.Config` struct. All values have sensible defaults.

| Parameter        | Default                  | Description                                      |
| ---------------- | ------------------------ | ------------------------------------------------ |
| `ChunkSize`      | `500`                    | Target character count per chunk                 |
| `ChunkOverlap`   | `100`                    | Characters of overlap between consecutive chunks |
| `EmbeddingModel` | `nomic-embed-text:v1.5`  | Ollama embedding model                           |
| `ChatModel`      | `llama2.3:3b`            | Ollama language model                            |
| `QdrantURL`      | `http://localhost:6333`  | Qdrant server address                            |
| `OllamaURL`      | `http://localhost:11434` | Ollama server address                            |
| `Collection`     | `documents`              | Qdrant collection name                           |
| `VectorSize`     | `768`                    | Must match your embedding model output dimension |

---

## Project Structure

```
vai/
├── main.go                 # Entry point — wires all layers together
├── go.mod
├── go.sum
│
├── chunker/
│   └── chunker.go          # Splits raw text into overlapping chunks
│
├── embeddings/
│   └── embeddings.go       # Generates vectors via Ollama or HuggingFace
│
├── vectorstore/
│   └── qdrant.go           # Qdrant client — upsert, search, delete
│
├── rag/
│   └── pipeline.go         # Full RAG pipeline — ingest, search, answer, stream
│
└── handlers/
    └── handlers.go         # HTTP handlers — upload, chat, stream, search
```

---

## Roadmap

- [ ] PDF and DOCX file support
- [ ] Web UI (chat interface)
- [ ] Conversation history and multi-turn context
- [ ] Multiple embedding model backends (HuggingFace, OpenAI-compatible)
- [ ] Document management endpoints (list, delete, re-index)
- [ ] Authentication and API key support
- [ ] Response caching for repeated questions
- [ ] Monitoring dashboard (latency, error rates, memory usage)
- [ ] Helm chart for Kubernetes deployment

---

## Contributing

Contributions are welcome. Please open an issue first to discuss what you would like to change, then submit a pull request against the `main` branch.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'add your feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Open a Pull Request

---

## License

MIT License. See [LICENSE](LICENSE) for details.

---

> Built with Go · Powered by open-source AI · Runs entirely on your machine
