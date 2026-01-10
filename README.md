# GraphWeaver

[![Go Report Card](https://goreportcard.com/badge/github.com/suyw-0123/graphweaver)](https://goreportcard.com/report/github.com/suyw-0123/graphweaver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**GraphWeaver** is a cloud-native, high-performance **Graph RAG (Retrieval-Augmented Generation)** knowledge engine. It leverages the power of "Entity-Relationship" modeling and "Semantic Vector" search to solve complex multi-hop reasoning challenges that traditional RAG systems often struggle with.

#### [ Click here to try the Demo](https://download-directory.github.io/?url=https://github.com/Suyw-0123/GraphWeaver/tree/main/demo)

---

## Key Features

- **Hybrid Retrieval Engine**: Combines **Vector Search** (for semantic entry points) with **Graph Diffusion** (for logical relationship reasoning).
- **Containerized Deployment**: Easy setup with **Docker Compose** for local development and testing.
- **High Performance**: Backend implemented in **Go (Golang)** for efficient concurrency and low-latency processing.
- **Modern Knowledge Management**: Automated entity and relationship extraction from unstructured data (PDF/Markdown).
- **Interactive UI**: Sleek and responsive dashboard built with **React**, **TypeScript**, and **Tailwind CSS**.

---

## Technology Stack

### 💻 Frontend & Styling

![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![TypeScript](https://img.shields.io/badge/TypeScript-007ACC?style=for-the-badge&logo=typescript&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=FFD62E)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white)

### ⚙️ Backend & Core

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-008080?style=for-the-badge&logo=gin&logoColor=white)
![Google Gemini](https://img.shields.io/badge/Google_Gemini-8E75B2?style=for-the-badge&logo=googlegemini&logoColor=white)

### 🗄️ Databases & Storage

![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Neo4j](https://img.shields.io/badge/Neo4j-008CC1?style=for-the-badge&logo=neo4j&logoColor=white)
![Qdrant](https://img.shields.io/badge/Qdrant-FF4B4B?style=for-the-badge&logo=qdrant&logoColor=white)

### 🚀 Infrastructure & DevOps

![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-2088FF?style=for-the-badge&logo=github-actions&logoColor=white)

---

## System Architecture

GraphWeaver follows a monolithic architecture with clear separation of concerns, designed for ease of development and deployment via Docker Compose.

![diagram](graphweaver_architecture.png)

---

## Quick Start

### Prerequisites

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/) 1.24+
- [Make](https://www.gnu.org/software/make/) (optional, for convenience)

### Running with Docker Compose

1.  Clone the repository:
    ```bash
    git clone https://github.com/suyw-0123/GraphWeaver.git
    cd GraphWeaver
    ```
2.  Configure your environment:
    ```bash
    cp .env.example .env
    # Edit .env and add your GEMINI_API_KEY
    ```
3.  Start the services:
    ```bash
    docker-compose up -d
    ```
4.  Access the web interface at `http://localhost:80`.

---

## Development

We use a `Makefile` to automate common development tasks:

- `make dev-tools`: Install necessary Go development tools.
- `make deps`: Download dependencies.
- `make fmt`: Format code.
- `make lint`: Run linters.
- `make test`: Run unit and integration tests.
- `make docker-build`: Build Docker images locally.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
