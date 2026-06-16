# Server Management Service (SMS)

A core backend service responsible for managing server lifecycles, monitoring real-time connection statuses, and processing asynchronous reporting. The system is designed using a **Modular Monolith** architecture, focusing on high performance, scalability, and ease of maintenance.

## 🛠 Tech Stack

- **Language:** Go 1.22
- **Communication:** gRPC, REST (via `grpc-gateway v2`)
- **Storage (OLTP):** PostgreSQL 15
- **Cache & Distributed Lock:** Redis 7
- **Log & Analytics:** Elasticsearch 8

---

## 🚀 Installation & Local Development (Developer Guide)

### 1. System Requirements
- Go 1.22+
- Docker & Docker Compose
- Make (On Windows, this can be installed via MinGW, or you can run the commands inside the `Makefile` manually).

### 2. Clone the Repository
```bash
git clone <REPO_URL>
cd server-management-service
```

### 3. Environment Configuration
Create a `.env` configuration file based on the provided example:
```bash
cp .env.example .env
```
*(Note: The default values in `.env.example` are already mapped to the default ports of the local infrastructure in step 4).*

### 4. Start Infrastructure
Use Docker Compose to start up the required dependencies, including PostgreSQL, Redis, Elasticsearch, and MailHog:
```bash
make infra-up
```
*(To stop the infrastructure: `make infra-down`)*

### 5. Run the Application
The SMS system consists of 3 independent processes running in parallel. To quickly spin up the entire Dev environment on Windows:
```bash
make dev
```
*(This command will automatically open 3 new Terminal windows corresponding to the 3 services).*

**If you wish to run each process individually, use the following commands:**
- `make run-api`: Starts the API Server.
- `make run-monitor`: Starts the Monitoring Worker (responsible for pinging servers).
- `make run-scheduler`: Starts the Daily Scheduler (responsible for triggering the report cronjob).

---

## 🏛 High-Level Architecture

```mermaid
flowchart LR
    classDef client fill:#08427b,color:#fff,stroke:#073b6f
    classDef api fill:#438dd5,color:#fff,stroke:#3c7fc0
    classDef worker fill:#85bbf0,color:#000,stroke:#5a91c8
    classDef db fill:#2e7d32,color:#fff,stroke:#1b5e20

    Client(["Client (Admin)"]):::client

    subgraph Modulith["server-management-service (Multi-Process)"]
        API["API Server"]:::api
        MON["Monitoring Worker"]:::worker
        SCH["Daily Scheduler"]:::worker
    end

    PG[("PostgreSQL")]:::db
    RD[("Redis")]:::db
    ES[("Elasticsearch")]:::db

    Client -->|"gRPC / REST\n(Port :8000)"| API
    SCH -.->|"gRPC Internal\n(Port :50051)"| API
    
    API --> PG & RD & ES
    MON -->|"Read / Update State"| PG & RD
    MON -->|"Log Ping"| ES
```

---

## 🧪 Testing & Quality

The project enforces strict Unit Testing using Mocking tools (`mockery`). The target code coverage for core modules must always be **> 90%**.

To run tests locally, use the following commands:
```bash
# Run the entire test suite and generate a coverage report in the terminal
make test-coverage

# Open the visual coverage report in your HTML browser
make test-coverage-html
```

