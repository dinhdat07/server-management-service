# SYSTEM DESCRIPTION AND USER MANUAL
**SERVER MANAGEMENT SYSTEM (SMS)**

---

## PART I: SYSTEM DESCRIPTION & TECHNICAL SPECIFICATIONS

### 1. General Introduction
The Server Management System (SMS) is a centralized monitoring solution that allows Administrators to manage information and track the uptime status of tens of thousands of servers in real-time via the ICMP (Ping) protocol.

### 2. Fulfilled Functional Requirements
The system completely solves the required business use-cases, including:
- **Status Monitoring (Critical):** Periodically runs background pings and updates the On/Off status centrally for tens of thousands of servers.
- **Data Management (CRUD & View):** Create, update, delete, search, paginate, and sort servers. Ensures unique identification constraints and valid IPv4 formatting.
- **Import / Export (High):** Handles high-speed bulk import/export via Excel files, with a mechanism to automatically skip duplicate records.
- **Automated & Manual Reporting (High):** A Cronjob process automatically sends daily Uptime reports via Email. Additionally, APIs allow administrators to proactively extract reports for custom timeframes.

### 3. Fulfilled Non-Functional Requirements
The system is designed and built in strict adherence to technical standards. Below are specific evidences for each requirement:

- **Data Architecture (Polyglot Persistence):**
  - Uses **PostgreSQL** as the primary database (Primary DB) to store identities. *(Evidence: Connection initialized at `internal/shared/database/postgres.go` using the `pgx` driver)*
  - Uses **Redis** as a high-speed cache and for distributed locking. *(Evidence: Dual-write and Mutex lock `lock:monitoring_worker` are implemented in `internal/infrastructure/redis/`)*
  - Uses specialized **Elasticsearch** to store ping logs and calculate Uptime ratios at ultra-high speeds. *(Evidence: `BulkInsert` function and Aggregation APIs designed at `internal/infrastructure/elasticsearch/client.go`)*
> **[INSERT IMAGE HERE: Screenshot of Docker Desktop showing 3 green/running containers for Postgres, Redis, Elasticsearch]**

- **Security:**
  - All APIs are protected by **JWT (JSON Web Token)** authentication with secure signatures. *(Evidence: The `Authenticator` Middleware class blocks any invalid request at `internal/infrastructure/security/authenticator.go`)*
  - Strict Role-Based Access Control (RBAC) with unique Role/Scope per endpoint.
  - Complete prevention of SQL Injection through an ORM library (GORM). *(Evidence: Source code uses `gorm.io/gorm` with Prepared Statements instead of raw SQL concatenation)*
> **[INSERT IMAGE HERE: Screenshot of Postman testing an API without a Token and receiving a 401 Unauthorized error]**

- **API Documentation:**
  - The system automatically generates **OpenAPI (Swagger)** documentation, clearly defining Requests, Responses, and Error Codes. *(Evidence: You can view the UI documentation directly at `/swagger/index.html` when running the API Server, all models are stored in the `docs/` folder)*
> **[INSERT IMAGE HERE: Screenshot of a beautiful Swagger UI listing all APIs comprehensively]**

- **Code Quality:**
  - The system achieves **high Code Coverage (>= 90%)** through independent Unit Tests. *(Evidence: Uses `github.com/stretchr/testify/mock` to mock database and redis in all test cases within the `service` and `handler` packages)*
  - Full logging to files with a log rotation mechanism (**Logrotate**) to prevent disk space exhaustion. *(Evidence: MaxSize, MaxBackups configuration using the `gopkg.in/natefinch/lumberjack.v2` library in the `logger` package)*
> **[INSERT IMAGE HERE: Screenshot of Terminal/Console after running `go test ./...` showing 100% green coverage lines]**

---

## PART II: DETAILED USER MANUAL

### 1. Login and Security (Authentication)
All users must be granted an account to access the system.
- Enter **Email** and **Password** at the login screen.
- If the credentials are valid, the system grants a JWT Token and redirects to the Administration dashboard.

> **[INSERT IMAGE 1 HERE: Screenshot of the Login Interface]**

---

### 2. Server List Management (View & CRUD Server)
This module allows the Admin to directly manipulate Server data. The displayed data structure includes at minimum: `server_id` (hidden/unique), `server_name`, `ipv4`, `status`, `created_time`, `last_updated`.

**2.1. View Server List**
- The system supports displaying lists in a table format with **Pagination**.
- Diverse **Filters** are supported: Smart search simultaneously by Name or IPv4, filter by Status (Online/Offline), and filter by **Creation Time Range (Created From - To)**.
- Flexible data **Sorting** is supported.

> **[INSERT IMAGE 2 HERE: Screenshot of the Server List with Search bar, Filters, and Pagination]**

**2.2. Create, Update, Delete (CRUD Server)**
- **Create:** Click the "Add Server" button. Requires `server_name` to be unique and `ipv4` to be properly formatted. The ID will be automatically generated by the system (UUID).
- **Update:** Click the Edit icon on any data row.
- **Delete:** Click the Delete icon. The system asks for confirmation before permanently deleting.

> **[INSERT IMAGE 3 HERE: Screenshot of the Create / Edit Server Popup Form showing Validation warnings]**

---

### 3. Bulk Data Import / Export
A feature to save time when working with thousands of Servers.

**3.1. Import Excel**
- Download the sample Excel file provided by the system.
- Fill in the list of Servers. Upon upload, the system will process in the background: automatically **skipping duplicate records** and reporting the number of successful/failed rows.

> **[INSERT IMAGE 4 HERE: Screenshot of the Excel Import Popup and Import results notification]**

**3.2. Export Excel**
- Click the "Export Excel" button on the list interface.
- The system will return an `.xlsx` file containing all data matching the current filters (including Name/IP, Status, and Creation Date Range filters).

---

### 4. Real-time Monitoring
This is the core (Critical) feature of the system.
- **Auto Scanning:** A background process periodically scans (pings) all 10,000+ servers every 30 seconds.
- **Centralized Updates:** Any status change (From On to Off and vice versa) is automatically updated on the Server list UI in real-time.
- You can monitor the **Status (Online/Offline)** column and the **Consecutive Failures** column to quickly assess network health.

> **[INSERT IMAGE 5 HERE: Screenshot displaying the green/red Status column and the Consecutive Failures column on the list]**

---

### 5. Statistics and Reporting

**5.1. Automated Report (Cronjob)**
- The system has a background process (Cronjob) that runs periodically exactly **once a day at 00:00**.
- This process automatically aggregates the number of On/Off Servers and calculates the average Uptime ratio for the previous day, then **sends it directly via Email** to Administrators.

> **[INSERT IMAGE 6 HERE: Screenshot of an Email inbox showing the Automated Report (HTML Email content)]**

**5.2. Manual Report**
- Admins can proactively request the system to calculate Uptime for any specific period via the Reporting interface.
- **How to execute:** Select `Start date`, `End date`, enter `Recipient Email` and click Send Request.
- The interface will display a list of report requests along with their status (Pending, Processing, Completed). Upon completion, the report will also be sent to the Email.

> **[INSERT IMAGE 7 HERE: Screenshot of the Manual Report Request interface (Start date, End date) and Status Table]**

---

### 6. Rate Limiting
To protect the system against Denial of Service (DDoS) attacks and ensure fair resource allocation, strict API limits are enforced. If you interact too quickly, the system will return a **429 Too Many Requests** error.

**Specific limits (calculated per 1 minute):**
- **Login / Import Server:** Max **5 requests/min**.
- **Request Report:** Max **10 requests/min**.
- **Delete Server:** Max **20 requests/min**.
- **Refresh Token:** Max **30 requests/min**.
- **Create / Update Server:** Max **60 requests/min**.
- **View Server List:** Max **120 requests/min**.
- **Other operations:** Max **300 requests/min**.

> [!NOTE]
> The countdown timer will automatically reset after each minute. If you encounter this error, please wait a moment and try again!
