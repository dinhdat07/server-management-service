// Identity Module Flows
dynamic backend "Identity_Dynamic" "Detailed sequence diagram for user login (Level 4)" {
    admin -> frontend "1. Clicks Login"
    frontend -> authHandler "2. POST /login"
    authHandler -> authService "3. Login(email, password)"
    authService -> userRepo "4. FindByEmail(email)"
    userRepo -> db "5. SELECT FROM users"
    authService -> authHandler "6. [If Invalid Credentials] Return 401 Unauthenticated"
    authService -> sessionRepo "7. Create(session)"
    sessionRepo -> db "8. INSERT INTO sessions"
    authService -> refreshRepo "9. Create(refreshToken)"
    refreshRepo -> db "10. INSERT INTO refresh_tokens"
    authService -> authHandler "11. Return LoginResult"
    authHandler -> frontend "12. Return 200 OK + JWT Tokens + [Set-Cookie]"
    autoLayout
}

dynamic backend "Identity_VerifyToken" "Detailed sequence diagram for Authenticator Middleware (Level 4)" {
    frontend -> authHandler "1. Make API Request (with JWT)"
    authHandler -> authService "2. Authenticate(token)"
    authService -> authHandler "3. [If missing/invalid] Return 401"
    authService -> redis "4. Check session revocation status"
    authService -> authHandler "5. [If revoked] Return 401"
    authService -> authHandler "6. Return Principal"
    authHandler -> frontend "7. Allow Request"
    autoLayout
}

dynamic backend "Identity_Logout" "Detailed sequence diagram for user logout (Level 4)" {
    admin -> frontend "1. Clicks Logout"
    frontend -> authHandler "2. POST /logout (with JWT)"
    authHandler -> authService "3. Logout(sessionID)"
    authService -> sessionRepo "4. RevokeByID(sessionID)"
    sessionRepo -> db "5. UPDATE sessions SET revoked_at=NOW()"
    authService -> refreshRepo "6. RevokeBySessionID(sessionID)"
    refreshRepo -> db "7. UPDATE refresh_tokens SET revoked_at=NOW()"
    authService -> redis "8. Set Revocation Key with TTL"
    authService -> authHandler "9. Return Success"
    authHandler -> frontend "10. Return 200 OK + [Clear-Cookie]"
    autoLayout
}

dynamic backend "Identity_Refresh" "Detailed sequence diagram for refreshing access token (Level 4)" {
    frontend -> authHandler "1. POST /refresh"
    authHandler -> authService "2. Refresh(refreshToken)"
    authService -> refreshRepo "3. FindByTokenHash(hash)"
    refreshRepo -> db "4. SELECT FROM refresh_tokens"
    authService -> redis "5. [Security] If Token Reuse Detected: Block all sessions & Return 401"
    authService -> authHandler "6. [If Expired] Return 401"
    authService -> sessionRepo "7. FindActiveByID(sessionID)"
    sessionRepo -> db "8. SELECT FROM sessions"
    authService -> userRepo "9. FindByID(userID)"
    userRepo -> db "10. SELECT FROM users"
    authService -> refreshRepo "11. RevokeByID & Create New"
    refreshRepo -> db "12. UPDATE & INSERT INTO refresh_tokens"
    authService -> authHandler "13. Return RefreshResult"
    authHandler -> frontend "14. Return 200 OK + New JWT Tokens + [Set-Cookie]"
    autoLayout
}

// Server Management Module Flows
dynamic backend "ServerManagement_List" "Detailed sequence diagram for querying servers (Level 4)" {
    admin -> frontend "1. Navigates to Servers List"
    frontend -> serverHandler "2. GET /servers?page=1"
    serverHandler -> serverService "3. SearchServers(filter)"
    serverService -> serverRepo "4. FetchWithPagination(filter)"
    serverRepo -> db "5. SELECT FROM servers WHERE status=? LIMIT OFFSET"
    serverService -> serverHandler "6. Return List"
    serverHandler -> frontend "7. Return JSON List"
    autoLayout
}

dynamic backend "ServerManagement_Create" "Detailed sequence diagram for creating a server (Level 4)" {
    admin -> frontend "1. Submits New Server Form"
    frontend -> serverHandler "2. POST /servers"
    serverHandler -> serverService "3. CreateServer(input)"
    serverService -> serverRepo "4. CheckExists(ip, name)"
    serverRepo -> db "5. SELECT COUNT"
    serverService -> serverHandler "6. [If Exists] Return 409 Conflict"
    serverService -> serverRepo "7. Create(server)"
    serverRepo -> db "8. INSERT INTO servers"
    serverService -> serverHandler "9. Return Server"
    serverHandler -> frontend "10. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Update" "Detailed sequence diagram for updating a server (Level 4)" {
    admin -> frontend "1. Submits Edit Server Form"
    frontend -> serverHandler "2. PUT /servers/{id}"
    serverHandler -> serverService "3. UpdateServer(id, input)"
    serverService -> serverRepo "4. FindByID(id)"
    serverRepo -> db "5. SELECT FROM servers"
    serverService -> serverHandler "6. [If Not Found] Return 404"
    serverService -> serverRepo "7. Update(server)"
    serverRepo -> db "8. UPDATE servers"
    serverService -> serverHandler "9. Return Server"
    serverHandler -> frontend "10. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Delete" "Detailed sequence diagram for deleting a server (Level 4)" {
    admin -> frontend "1. Clicks Delete Server"
    frontend -> serverHandler "2. DELETE /servers/{id}"
    serverHandler -> serverService "3. DeleteServer(id)"
    serverService -> serverRepo "4. Delete(id)"
    serverRepo -> db "5. DELETE FROM servers"
    serverService -> serverHandler "6. Return Success"
    serverHandler -> frontend "7. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Dynamic" "Detailed sequence diagram for importing servers (Level 4)" {
    admin -> frontend "1. Uploads Excel"
    frontend -> serverHandler "2. POST /import"
    serverHandler -> serverService "3. ImportServers(file)"
    serverService -> serverHandler "4. [If Invalid Excel Format] Return 400 Bad Request"
    serverService -> serverRepo "5. BatchCreate(servers)"
    serverRepo -> db "6. Batch INSERT INTO servers"
    serverService -> serverHandler "7. Return Import Result"
    serverHandler -> frontend "8. Return 200 OK"
    autoLayout
}

// Monitoring Module Flows
dynamic monitorWorker "Monitoring_Dynamic" "Detailed sequence diagram for continuous monitoring ping (Level 4)" {
    monitoringWorkerPool -> redis "1. SMEMBERS server_ids (Fetch Active Servers)"
    monitoringWorkerPool -> externalServers "2. Ping(ip)"
    monitoringWorkerPool -> monitoringService "3. EvaluatePingResult(result)"
    monitoringService -> serverStateStore "4. GetServerState(id)"
    serverStateStore -> redis "5. HGET server_state"
    monitoringService -> monitoringRepo "6. [If Ping Fails] UpdateServerStatus(id, offline)"
    monitoringService -> monitoringRepo "7. UpdateServerStatus(id, online)"
    monitoringRepo -> db "8. UPDATE servers SET status=?"
    monitoringService -> elastic "9. LogObservation(ping_latency, status)"
    autoLayout
}

// Reporting Module Flows
dynamic scheduler "Reporting_Scheduled" "Detailed sequence diagram for automated daily reporting (Level 4)" {
    dailyScheduler -> reportingServiceSch "1. GenerateDailyReport()"
    reportingServiceSch -> reportingRepoSch "2. GetServerCountByStatus()"
    reportingRepoSch -> db "3. SELECT count FROM servers GROUP BY status"
    reportingServiceSch -> esUptimeCalc "4. CalculateUptime()"
    esUptimeCalc -> elastic "5. Aggregation Query on ping logs"
    reportingServiceSch -> notificationServiceSch "6. SendReport(metadata, uptime)"
    notificationServiceSch -> smtpClientSch "7. Execute HTML Template & SendMail"
    smtpClientSch -> smtp "8. Send SMTP Protocol"
    autoLayout
}

dynamic backend "Reporting_Manual" "Detailed sequence diagram for manual report requests (Level 4 - Asynchronous)" {
    admin -> frontend "1. Request Report Generation"
    frontend -> reportingHandler "2. POST /reports (gRPC-Web)"
    reportingHandler -> reportingService "3. RequestReport(email, dates)"
    reportingService -> reportingRepo "4. CreateReportRequest(req)"
    reportingRepo -> db "5. INSERT INTO report_requests"
    reportingService -> reportingHandler "6. EnqueueReport(Channel)"
    reportingHandler -> frontend "7. Return 200 OK (Accepted)"
    autoLayout
}

dynamic backend "Reporting_Worker_Process" "Detailed sequence diagram for background report generation (Level 4)" {
    reportingService -> reportingRepo "1. Worker Dequeues Job from Channel"
    reportingRepo -> db "2. GetServerCountByStatus()"
    reportingService -> elastic "3. Aggregation Query on ping logs (ESUptimeCalculator)"
    reportingService -> notificationService "4. SendReport(metadata, uptime)"
    notificationService -> smtpClient "5. Execute HTML Template & SendMail"
    smtpClient -> smtp "6. Send SMTP Protocol"
    reportingService -> reportingRepo "7. UpdateReportRequestStatus(completed)"
    reportingRepo -> db "8. UPDATE report_requests"
    autoLayout
}
