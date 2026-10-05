# 🛡️ CloudSentinel — AI-Powered Distributed WAF & Threat Remediation Engine

**CloudSentinel** es un firewall de aplicaciones web (WAF) distribuido y de alto rendimiento escrito en **Go**, diseñado para interceptar tráfico TCP/HTTP, inspeccionar amenazas OWASP en tiempo real, gestionar control de tasa con **Redis**, persistir firmas de ataques con vectores en **PostgreSQL (pgvector)** y ejecutar remediación de seguridad de forma autónoma mediante un **Agente de IA (RAG + MCP)**.

---

## 📐 Arquitectura del Sistema

```mermaid
flowchart TD
    Client([Cliente / Atacante]) -->|Petición HTTP/TCP| Socket[Servidor TCP Native - net.Listen]
    
    subgraph Ingestion_Engine [Módulo 1: Engine de Ingesta & Concurrencia]
        Socket --> LRU[LRU Cache Thread-Safe]
        Socket --> Parser[HTTP Request Parser]
    end

    subgraph Security_Layer [Módulo 2: WAF & Rate Limiting]
        Parser --> RateLimiter{Redis Rate Limiter}
        RateLimiter -->|Supera límite 429| BlockRate[Bloqueo HTTP 429]
        RateLimiter -->|Permitido| WAF{WAF Regex Scanner}
        WAF -->|SQLi / XSS 403| BlockWAF[Bloqueo HTTP 403]
    end

    subgraph Event_Driven [Módulo 3: Persistencia Asíncrona]
        WAF -->|Evento Asíncrono| Chan[Go Channel Buffer]
        Chan --> DBWorker[Background Worker]
        DBWorker --> Postgres[(PostgreSQL + pgvector)]
    end

    subgraph Agentic_AI [Módulo 5: Agente Autónomo de Ciberseguridad]
        BlockWAF -.->|Trigger Threat| Agent[Agente IA]
        Agent -->|1. Consulta RAG| Postgres
        Agent -->|2. Tool Calling| EmergencyBlock[EmergencyBlockIP Tool]
        EmergencyBlock -->|Aísla IP por 1h| RateLimiter
    end
✨ Características Clave
Motor TCP de Bajo Nivel: Servidor de sockets concurrente en Go para procesar peticiones HTTP/1.1 usando goroutines.

LRU Cache Concurrente: Estructura en memoria protegida contra Race Conditions con sync.RWMutex.

WAF (OWASP Top 10): Inspección profunda de rutas decodificadas (URL Unescape) para mitigar SQLi, XSS y Path Traversal.

Rate Limiting Distribuido: Algoritmo Token Bucket en Redis con TTL automático.

Persistencia Vectorial (pgvector): Canal de eventos asíncrono con generación de embeddings e índice HNSW.

Despliegue Cloud-Native: Multi-Stage Dockerfile (<20MB) y manifiestos de Kubernetes con reglas Zero-Trust (NetworkPolicy).

DevSecOps Ready: Escaneo de vulnerabilidades sin hallazgos en binarios mediante Trivy.

Remediación Agéntica por IA: Agente autónomo con capacidad RAG e invocación de herramientas (Tool Calling).

🛠️ Tech Stack
Lenguaje: Go (Golang 1.22+)

Caché & In-Memory: Redis 7

Database & Vector Search: PostgreSQL 16 + pgvector

Contenedores & Orquestación: Docker, Kubernetes

Seguridad: Trivy Container Scanner

Patrones: Event-Driven, Microservicios, Zero-Trust, Agentic AI (RAG / MCP)

🚀 Inicio Rápido (Quickstart)

# 1. Iniciar Redis y PostgreSQL + pgvector
docker run -d --name redis-sentinel -p 6379:6379 redis:alpine
docker run -d --name postgres-sentinel -e POSTGRES_PASSWORD=sentinelpass -e POSTGRES_DB=cloudsentinel_db -p 5432:5432 ankane/pgvector

# 2. Inicializar DB
docker exec -i postgres-sentinel psql -U postgres -d cloudsentinel_db < schema.sql

# 3. Ejecutar el WAF
go run .

