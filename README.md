# vai

An AI Knowledge system. Upload documents, ask questions, get AI-generated answers — all running locally with no external API required.

---

## Features

- Upload and query PDF, Markdown, and plain text documents
- Semantic search over document content using vector embeddings
- AI answers grounded in your documents via a RAG pipeline
- Streaming responses for a real-time chat experience
- Fully containerized — runs locally with Docker

---

## System Architecture

```
Frontend (React + Vite + Tailwind)
        ↓
Backend API (Go)
        ↓
Vector DB (Qdrant)
        ↓
LLM service (Ollama)
        ↓
AI model (Llama 3)
```

### Document Flow

```
Upload document
        ↓
Split into chunks
        ↓
Generate embeddings
        ↓
Store in vector DB
        ↓
User asks question
        ↓
Embed question → search vector DB
        ↓
Retrieve relevant chunks
        ↓
Inject into LLM prompt
        ↓
Stream answer to user
```

---

## Tech Stack

| Layer | Tool |
|---|---|
| Frontend | React, Vite, Tailwind CSS |
| Backend | Go |
| Local LLM | [Ollama](https://ollama.com) + Llama 3 |
| Embeddings | [Hugging Face](https://huggingface.co) sentence transformers |
| Vector DB | [Qdrant](https://qdrant.tech) |
| Deployment | Docker + Docker Compose |

---

## Prerequisites

- Docker and Docker Compose
- Go 1.22+
- Node.js 18+
- 16GB+ RAM

---

## Quick Start

```bash
git clone https://github.com/your-username/vai.git
cd vai
ollama pull qwen3.5:4b
docker compose up
```

Visit `http://localhost:5173`.

---

## Docker Compose

```yaml
services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      - qdrant
      - ollama

  qdrant:
    image: qdrant/qdrant
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage

  ollama:
    image: ollama/ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama

volumes:
  qdrant_data:
  ollama_data:
```

---

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/upload` | Upload a document |
| `POST` | `/api/chat` | Ask a question |
| `GET` | `/api/stream` | Stream a response via SSE |

### Example

```json
POST /api/chat
{
  "question": "How does authentication work?",
  "history": []
}
```

---

## Development

### Backend (Go)

```bash
cd backend
go mod tidy
go run main.go
```

### Frontend (React + Vite)

```bash
cd frontend
npm install
npm run dev
```

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `OLLAMA_HOST` | `http://ollama:11434` | Ollama service URL |
| `QDRANT_HOST` | `http://qdrant:6333` | Qdrant service URL |
| `EMBEDDING_MODEL` | `all-MiniLM-L6-v2` | Embedding model |
| `LLM_MODEL` | `llama3` | Ollama model |
| `CHUNK_SIZE` | `500` | Tokens per chunk |
| `CHUNK_OVERLAP` | `50` | Overlap between chunks |
| `TOP_K` | `5` | Chunks retrieved per query |

---

## License

MIT