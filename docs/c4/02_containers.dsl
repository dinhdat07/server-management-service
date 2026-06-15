// Containers within SMS System
frontend = container "Web Application" "Provides the UI for administrators." "Angular SPA" "WebBrowser"

db = container "Database" "Stores system state, servers, and configurations." "PostgreSQL 15" "Database"
redis = container "Cache & Lock" "Caches server state, sessions, distributed locks, and rate-limit counters." "Redis 7" "Database"
elastic = container "Log Storage" "Stores time-series ping observation logs for uptime calculation." "Elasticsearch 8" "Database"

// Backend API
backend = container "Backend Application (API)" "Provides REST/gRPC endpoints and handles core business logic." "Go" {
    // High-Level Components (For Level 3)
    identityComp = component "Identity Component" "Handles user authentication and authorization." "Go Module"
    serverComp = component "Server Management Component" "Handles server CRUD and Import/Export." "Go Module"
    reportingComp = component "Reporting Component" "Generates HTML uptime reports asynchronously and sends via email." "Go Module"
    notificationComp = component "Notification Component" "Handles email rendering and sending." "Go Module"

    // Low-Level Components (For Level 4 Dynamic Views)
    group "Identity Module (Code Level)" {
        authHandler = component "Auth Handler" "Provides gRPC/REST Endpoints" "Go"
        authService = component "Auth Service" "Handles Business Logic" "Go"
        userRepo = component "User Repository" "Postgres Data Access" "Go"
        sessionRepo = component "Session Repository" "Postgres Data Access" "Go"
        refreshRepo = component "Refresh Token Repository" "Postgres Data Access" "Go"
    }
    
    group "Server Management Module (Code Level)" {
        serverHandler = component "Server Handler" "Provides gRPC/REST Endpoints" "Go"
        serverService = component "Server Service" "Handles Business Logic, Excel Parsing & Redis Dual-Write" "Go"
        serverRepo = component "Server Repository" "Postgres Data Access" "Go"
        serverCache = component "Server Cache" "Redis Dual-Write for Server State" "Go"
    }
    
    group "Reporting Module (Code Level)" {
        reportingHandler = component "Reporting Handler" "Provides gRPC/REST Endpoints" "Go"
        reportingService = component "Reporting Service" "Handles Business Logic" "Go"
        reportingWorker = component "Reporting Worker" "Async Goroutine Pool for Report Generation" "Go"
        reportingRepo = component "Reporting Repository" "Postgres Data Access" "Go"
        esUptimeCalcBackend = component "ES Uptime Calculator" "Calculates uptime from Elasticsearch" "Go"
    }

    group "Notification Module (Code Level)" {
        notificationService = component "Notification Service" "Handles email rendering" "Go"
        smtpClient = component "SMTP Client" "Sends emails" "Go"
    }
}

// Monitoring Worker
monitorWorker = container "Monitoring Worker" "Background process that pings servers continuously." "Go Worker" {
    // High-Level
    monitoringComp = component "Monitoring Component" "Runs continuous ping loops and stores logs." "Go Module"
    
    // Low-Level
    group "Monitoring Module (Code Level)" {
        monitoringWorkerPool = component "Monitoring Worker Pool" "Manages goroutines for pinging" "Go"
        icmpPinger = component "ICMP Pinger" "Sends ICMP ping to target servers via pro-bing" "Go"
        monitoringService = component "Monitoring Service" "Evaluates ping results using State Machine (FSM)" "Go"
        monitoringRepo = component "Monitoring Repository" "Postgres Data Access" "Go"
        serverStateStore = component "Server State Store" "Redis Data Access" "Go"
        observationLogger = component "Observation Logger" "Buffered async bulk logger to Elasticsearch" "Go"
    }
}

// Daily Scheduler
scheduler = container "Daily Scheduler" "Background process that triggers periodic jobs." "Go Cron" {
    // High-Level
    schedulerComp = component "Scheduler Component" "Triggers daily reports and other cron jobs." "Go Module"

    // Low-Level
    group "Reporting Module (Embedded Code Level)" {
        dailyScheduler = component "Daily Scheduler Worker" "Triggers periodic reporting tasks" "Go"
        esUptimeCalc = component "ES Uptime Calculator" "Calculates uptime from Elasticsearch" "Go"
        reportingServiceSch = component "Reporting Service (Scheduler)" "Aggregates report data" "Go"
        reportingRepoSch = component "Reporting Repository (Scheduler)" "Postgres Data Access" "Go"
    }
    group "Notification Module (Embedded Code Level)" {
        notificationServiceSch = component "Notification Service (Scheduler)" "Handles email rendering" "Go"
        smtpClientSch = component "SMTP Client (Scheduler)" "Sends emails" "Go"
    }
}
