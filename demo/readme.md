# 🚀 GraphWeaver Quick Start Demo

This folder contains everything you need to run GraphWeaver locally using Docker. No coding required!

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Windows/Mac) or Docker Engine (Linux)
- Docker Compose (usually included with Docker Desktop)

## How to Run

1. **Download Files**
   Download the files in this folder:
   - `docker-compose.yml`
   - `.env.example`

2. **Configure Environment**
   - Rename `.env.example` to `.env`.
   - Open `.env` with a text editor.
   - Set your `GEMINI_API_KEY` (Get one [here](https://aistudio.google.com/app/apikey)).
   - Set a secure password for `DB_PASSWORD`.

3. **Start the App**
   Open a terminal in this folder and run:
   ```bash
   docker compose up -d
   ```

4. **Access**
   - Frontend: [http://localhost](http://localhost)
   - Backend API: [http://localhost:8080](http://localhost:8080)

## Troubleshooting

- **Database Error**: If you see logs about "password authentication failed", make sure your `.env` file is named correctly and contains `DB_PASSWORD`.
- **Port Conflict**: If port 80 or 8080 is taken, edit `docker-compose.yml` to change the ports (e.g., `8081:80`).

## Stopping the App

```bash
docker compose down
```
