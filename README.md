<div align="center">

<img src="https://img.shields.io/badge/dispatch-v1.0.0-F59E0B?style=for-the-badge&labelColor=111111" alt="version"/>
<img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=111111" alt="go"/>
<img src="https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=white&labelColor=111111" alt="react"/>
<img src="https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white&labelColor=111111" alt="redis"/>
<img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white&labelColor=111111" alt="postgres"/>
<img src="https://img.shields.io/badge/Docker-ready-2496ED?style=for-the-badge&logo=docker&logoColor=white&labelColor=111111" alt="docker"/>
<img src="https://img.shields.io/badge/license-MIT-22C55E?style=for-the-badge&labelColor=111111" alt="license"/>

<br/>
<br/>

```
██████╗ ██╗███████╗██████╗  █████╗ ████████╗ ██████╗██╗  ██╗
██╔══██╗██║██╔════╝██╔══██╗██╔══██╗╚══██╔══╝██╔════╝██║  ██║
██║  ██║██║███████╗██████╔╝███████║   ██║   ██║     ███████║
██║  ██║██║╚════██║██╔═══╝ ██╔══██║   ██║   ██║     ██╔══██║
██████╔╝██║███████║██║     ██║  ██║   ██║   ╚██████╗██║  ██║
╚═════╝ ╚═╝╚══════╝╚═╝     ╚═╝  ╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝
```

### **Distributed Task Queue Engine**
*High-throughput, fault-tolerant background job processing — built from scratch in Go*

[Live Demo](https://dispatch-beryl.vercel.app) · [Features](#-features) · [Architecture](#-architecture) · [Quick Start](#-quick-start) · [API Reference](#-api-reference) · [Dashboard](#-dashboard) · [Benchmarks](#-benchmarks)

</div>

---

## What is Dispatch?

Dispatch is a production-grade distributed task queue engine built from scratch in Go. It handles async job processing across distributed worker pools with priority scheduling, automatic retries, dead-letter queues, and a real-time monitoring dashboard — the kind of infrastructure every backend system needs, built transparently so you understand exactly how it works.

Think **Celery** or **BullMQ**, but written in Go with full visibility into every design decision.

---

## Features

### Core Engine
- **Priority Queues** — `CRITICAL` / `HIGH` / `NORMAL` / `LOW` lanes with strict priority ordering
- **Worker Pool** — configurable concurrency with graceful shutdown and in-flight job protection
- **Retry Logic** — exponential backoff, configurable max attempts per job type
- **Dead-Letter Queue** — failed jobs routed to DLQ after exhausting retries, retryable from UI
- **Delayed Jobs** — schedule jobs to run at a future time via Redis sorted sets
- **Queue Pause/Resume** — backpressure control, pause any queue without losing jobs

### Observability
- **Real-Time Dashboard** — live throughput graphs, worker health, queue depth via WebSockets
- **Job Inspector** — filter by status/priority, cancel pending jobs, retry dead jobs
- **Metrics API** — jobs/sec throughput, worker utilization, DLQ count
- **Structured Logging** — per-request logs with duration and status

### Reliability
- **PostgreSQL Persistence** — full job history, execution logs, retry records
- **Redis Pub/Sub** — real-time events broadcast to all connected dashboard clients
- **Graceful Shutdown** — drains in-flight jobs before stopping workers
- **Health Check** — `/health` endpoint for container orchestration

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT / API                            │
│              REST API  ·  WebSocket  ·  Go SDK                  │
└──────────────────────────────┬──────────────────────────────────┘
                               │  Enqueue Jobs
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                       DISPATCH ENGINE                           │
│                                                                 │
│   ┌─────────────┐  ┌─────────────┐  ┌──────────────────────┐   │
│   │  CRITICAL   │  │    HIGH     │  │       NORMAL         │   │
│   │   Queue     │  │   Queue     │  │       Queue          │   │
│   └──────┬──────┘  └──────┬──────┘  └──────────┬───────────┘   │
│          │                │                    │               │
│          └────────────────┴────────────────────┘               │
│                           │  Priority Dequeue                  │
│                           ▼                                     │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                   WORKER POOL                           │   │
│   │   Worker 1  ·  Worker 2  ·  Worker 3  ·  Worker N      │   │
│   └──────────────────────────┬──────────────────────────────┘   │
│                              │                                  │
│           ┌──────────────────┼──────────────────┐              │
│           ▼                  ▼                  ▼              │
│      Job Success        Job Failed         Job Timeout         │
│           │             (retry?)               │               │
│           │           ┌────┴────┐              │               │
│           │      retry│         │max retries   │               │
│           │           ▼         ▼              │               │
│           │      Re-queue      DLQ ◄───────────┘               │
│           ▼                                                     │
│     PostgreSQL ◄──── Full History + Logs                        │
└─────────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    REAL-TIME DASHBOARD                          │
│         React · WebSockets · Live Charts · Job Inspector        │
└─────────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.26 | Core engine, worker pool, REST API |
| **Queue State** | Redis 7 | Priority queues, pub/sub, job leases |
| **Persistence** | PostgreSQL 16 | Job history, execution logs, DLQ |
| **Real-Time** | WebSockets | Live dashboard updates |
| **Frontend** | React 18 + Vite | Monitoring dashboard |
| **Charts** | Recharts | Throughput and latency graphs |
| **Containers** | Docker + Compose | One-command local setup |
| **CI/CD** | GitHub Actions | Test, lint, build on push |

---

## Quick Start

### Prerequisites
- Docker + Docker Compose

### Run in 60 seconds

```bash
git clone https://github.com/gautammehendale/dispatch.git
cd dispatch
docker-compose up --build
```

- **Dashboard** → [http://localhost:3000](http://localhost:3000)
- **API** → [http://localhost:8080](http://localhost:8080)
- **Health** → [http://localhost:8080/health](http://localhost:8080/health)

### Enqueue your first job

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "send_email",
    "priority": "HIGH",
    "payload": { "to": "user@example.com", "subject": "Welcome!" },
    "max_retries": 3,
    "run_at": "now"
  }'
```

---

## API Reference

### Jobs

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/jobs` | Enqueue a new job |
| `GET` | `/api/v1/jobs` | List jobs with filters |
| `GET` | `/api/v1/jobs/:id` | Get job details + execution log |
| `POST` | `/api/v1/jobs/:id/cancel` | Cancel a pending or retrying job |
| `POST` | `/api/v1/jobs/:id/retry` | Manually retry a failed job |

### Queues

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/queues` | List all queues and depths |
| `POST` | `/api/v1/queues/:name/pause` | Pause a queue |
| `POST` | `/api/v1/queues/:name/resume` | Resume a paused queue |
| `GET` | `/api/v1/dlq` | List dead-letter queue jobs |

### Workers

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/workers` | List active workers + status |
| `GET` | `/api/v1/metrics` | Throughput, latency, error rates |

### WebSocket

```
ws://localhost:8080/ws
```

Streams real-time events: `job.enqueued`, `job.started`, `job.completed`, `job.failed`, `worker.heartbeat`

---

## Dashboard

The Dispatch dashboard gives you full visibility into your job pipeline in real-time:

- **Throughput Graph** — live jobs/sec chart powered by WebSocket events
- **Queue Depth Monitor** — real-time depth per queue
- **Worker Grid** — each worker's status, current job, and jobs processed
- **Job Inspector** — filter by status/priority, cancel or retry individual jobs
- **DLQ Viewer** — inspect dead jobs, retry them back into the queue

---

## Benchmarks

> Tested locally — Docker Compose, 4 workers, Apple M1/M2

| Metric | Value |
|--------|-------|
| Enqueue throughput | **~1,000 jobs/sec** |
| Dequeue + execute (no-op handler) | **~1,000 jobs/sec** |
| p50 engine latency | **~2ms** |
| p99 engine latency | **~6ms** |
| Graceful shutdown drain | **< 500ms** |
| Memory footprint (backend) | **~30MB** |

### Verified Test Output

```
=== RUN   TestThroughput
  Throughput benchmark (4 workers, 5s)
  Jobs processed : 5236
  Elapsed        : 5.00s
  Jobs/sec       : 1047
--- PASS: TestThroughput (5.01s)

=== RUN   TestLatency
  Latency benchmark (200 jobs)
  p50 : 1.605ms
  p95 : 1.948ms
  p99 : 2.893ms
--- PASS: TestLatency (0.42s)
```

> Run with: `cd backend && go test ./internal/queue/... -v -run "TestThroughput|TestLatency" -timeout 60s`

---

## Project Structure

```
dispatch/
├── backend/
│   ├── cmd/server/         # Entry point
│   ├── internal/
│   │   ├── queue/          # Priority queue engine
│   │   ├── worker/         # Worker pool + lifecycle
│   │   ├── scheduler/      # Delayed job scheduling
│   │   ├── api/            # REST handlers + WebSocket
│   │   ├── store/          # Redis + PostgreSQL clients
│   │   └── models/         # Shared types
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/     # Dashboard UI components
│   │   ├── pages/          # Route pages
│   │   └── hooks/          # WebSocket + data hooks
│   └── Dockerfile
├── docker-compose.yml
└── .github/workflows/      # CI pipeline
```

---

## License

MIT — see [LICENSE](LICENSE)

---

<div align="center">
  Built by <a href="https://github.com/gautammehendale">Gautam Mehendale</a>
</div>
