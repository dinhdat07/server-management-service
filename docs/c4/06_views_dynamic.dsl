// Identity Module Flows
dynamic backend "Identity_Dynamic" "Detailed sequence diagram for user login (Level 4)" {
    admin -> frontend "1. Clicks Login"
    frontend -> authHandler "2. POST /login"
    authHandler -> authService "3. Login(email, password)"
    authService -> userRepo "4. FindByEmail(email)"
    userRepo -> db "5. SELECT FROM users"
    authService -> authHandler "6. [If Invalid Credentials] Return 401 Unauthenticated"
    authService -> sessionRepo "7. Create(session)"
    sessionRepo -> db "8. INSERT INTO auth_sessions"
    authService -> refreshRepo "9. Create(refreshToken)"
    refreshRepo -> db "10. INSERT INTO refresh_tokens"
    authService -> authHandler "11. Return LoginResult (AccessToken + RefreshToken)"
    authHandler -> frontend "12. Return 200 OK + JWT Tokens + [Set-Cookie]"
    autoLayout
}

dynamic backend "Identity_VerifyToken" "Detailed sequence diagram for Authenticator Middleware (Level 4)" {
    frontend -> authHandler "1. Make API Request (with JWT in Cookie/Header)"
    authHandler -> authService "2. Middleware: ParseJWT(token)"
    authService -> authHandler "3. [If missing/invalid/expired] Return 401"
    authService -> redis "4. EXISTS revoked_session:{session_id}"
    authService -> authHandler "5. [If revoked] Return 401"
    authService -> authHandler "6. Return Principal (userID, roleCode, sessionID)"
    authHandler -> frontend "7. Allow Request → proceed to handler"
    autoLayout
}

dynamic backend "Identity_Logout" "Detailed sequence diagram for user logout (Level 4)" {
    admin -> frontend "1. Clicks Logout"
    frontend -> authHandler "2. POST /logout (with JWT)"
    authHandler -> authService "3. Logout(sessionID)"
    authService -> sessionRepo "4. RevokeByID(sessionID)"
    sessionRepo -> db "5. UPDATE auth_sessions SET revoked_at=NOW()"
    authService -> refreshRepo "6. RevokeBySessionID(sessionID)"
    refreshRepo -> db "7. UPDATE refresh_tokens SET revoked_at=NOW()"
    authService -> redis "8. SET revoked_session:{session_id} with TTL"
    authService -> authHandler "9. Return Success"
    authHandler -> frontend "10. Return 200 OK + [Clear-Cookie]"
    autoLayout
}

dynamic backend "Identity_Refresh" "Detailed sequence diagram for refreshing access token (Level 4)" {
    frontend -> authHandler "1. POST /refresh"
    authHandler -> authService "2. Refresh(refreshToken)"
    authService -> refreshRepo "3. FindByTokenHash(SHA256(token))"
    refreshRepo -> db "4. SELECT FROM refresh_tokens"
    authService -> redis "5. [Security] If Token Reuse Detected (already revoked): LogoutAll & Return 401"
    authService -> authHandler "6. [If Expired] Return 401"
    authService -> sessionRepo "7. FindActiveByID(sessionID)"
    sessionRepo -> db "8. SELECT FROM auth_sessions"
    authService -> userRepo "9. FindByID(userID)"
    userRepo -> db "10. SELECT FROM users"
    authService -> refreshRepo "11. RevokeByID(old) & Create(new) & MarkReplacement"
    refreshRepo -> db "12. UPDATE & INSERT INTO refresh_tokens"
    authService -> authHandler "13. Return RefreshResult (new AccessToken + RefreshToken)"
    authHandler -> frontend "14. Return 200 OK + New JWT Tokens + [Set-Cookie]"
    autoLayout
}

// Server Management Module Flows
dynamic backend "ServerManagement_List" "Detailed sequence diagram for querying servers (Level 4)" {
    admin -> frontend "1. Navigates to Servers List"
    frontend -> serverHandler "2. GET /api/v1/servers"
    serverHandler -> serverService "3. SearchServers(filter)"
    serverService -> serverRepo "4. Search(filter)"
    serverRepo -> db "5. SELECT FROM management_schema.servers WHERE status=? LIMIT OFFSET"
    serverService -> serverHandler "6. Return List + Total Count"
    serverHandler -> frontend "7. Return paginated server list"
    autoLayout
}

dynamic backend "ServerManagement_Create" "Detailed sequence diagram for creating a server (Level 4)" {
    admin -> frontend "1. Submits New Server Form"
    frontend -> serverHandler "2. POST /api/v1/servers"
    serverHandler -> serverService "3. CreateServer(input)"
    serverService -> serverRepo "4. GetByName(name) & GetByIPv4(ipv4)"
    serverRepo -> db "5. SELECT FROM servers (uniqueness check)"
    serverService -> serverHandler "6. [If Name/IPv4 Exists] Return 409 Conflict"
    serverService -> serverRepo "7. Create(server)"
    serverRepo -> db "8. INSERT INTO management_schema.servers"
    serverService -> serverCache "9. Dual-Write: Upsert(id, ipv4, status, 0)"
    serverCache -> redis "10. HSET server:info:{id} & SADD server:all_ids"
    serverService -> serverHandler "11. Return Server"
    serverHandler -> frontend "12. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Update" "Detailed sequence diagram for updating a server (Level 4)" {
    admin -> frontend "1. Submits Edit Server Form"
    frontend -> serverHandler "2. PUT /api/v1/servers/{id}"
    serverHandler -> serverService "3. UpdateServer(id, input)"
    serverService -> serverRepo "4. GetByID(id)"
    serverRepo -> db "5. SELECT FROM servers"
    serverService -> serverHandler "6. [If Not Found] Return 404"
    serverService -> serverRepo "7. [If Name/IPv4 Changed] Check uniqueness & Update(server)"
    serverRepo -> db "8. UPDATE management_schema.servers"
    serverService -> serverCache "9. Dual-Write: Upsert(id, ipv4, status, retryCount)"
    serverCache -> redis "10. HSET server:info:{id}"
    serverService -> serverHandler "11. Return Server"
    serverHandler -> frontend "12. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Delete" "Detailed sequence diagram for deleting a server (Level 4)" {
    admin -> frontend "1. Clicks Delete Server"
    frontend -> serverHandler "2. DELETE /api/v1/servers/{id}"
    serverHandler -> serverService "3. DeleteServer(id)"
    serverService -> serverRepo "4. GetByID(id)"
    serverRepo -> db "5. SELECT FROM servers (existence check)"
    serverService -> serverHandler "6. [If Not Found] Return 404"
    serverService -> serverRepo "7. Delete(id)"
    serverRepo -> db "8. DELETE FROM management_schema.servers"
    serverService -> serverCache "9. Dual-Write: Delete(id)"
    serverCache -> redis "10. DEL server:info:{id} & SREM server:all_ids"
    serverService -> serverHandler "11. Return Success"
    serverHandler -> frontend "12. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Dynamic" "Detailed sequence diagram for importing servers via Excel (Level 4)" {
    admin -> frontend "1. Uploads Excel file"
    frontend -> serverHandler "2. POST /api/v1/servers/import (multipart)"
    serverHandler -> serverService "3. ImportServers(fileBytes)"
    serverService -> serverHandler "4. [If >2MB or Invalid Excel] Return 400 Bad Request"
    serverService -> serverRepo "5. FindByNamesOrIPv4s(names, ips) per batch(100)"
    serverRepo -> db "6. SELECT FROM servers WHERE name IN(...) OR ipv4 IN(...)"
    serverService -> serverRepo "7. BatchCreate(validServers)"
    serverRepo -> db "8. Batch INSERT INTO management_schema.servers"
    serverService -> serverCache "9. Dual-Write: BatchUpsert(cacheItems) via Redis Pipeline"
    serverCache -> redis "10. Pipeline: HSET + SADD for each server"
    serverService -> serverHandler "11. Return ImportResult (successCount, failCount, details)"
    serverHandler -> frontend "12. Return 200 OK"
    autoLayout
}

dynamic backend "ServerManagement_Export" "Detailed sequence diagram for exporting servers to Excel (Level 4)" {
    admin -> frontend "1. Clicks Export Servers"
    frontend -> serverHandler "2. GET /api/v1/servers/export"
    serverHandler -> serverService "3. ExportServers(filter)"
    serverService -> serverRepo "4. Search(filter)"
    serverRepo -> db "5. SELECT FROM management_schema.servers WHERE status=? LIMIT OFFSET"
    serverService -> serverHandler "6. Generate Excel File & Return File Bytes"
    serverHandler -> frontend "7. Return 200 OK + application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
    autoLayout
}

// Monitoring Module Flows
dynamic monitorWorker "Monitoring_Dynamic" "Detailed sequence diagram for continuous monitoring ping cycle (Level 4)" {
    monitoringWorkerPool -> redis "1. Acquire Distributed Lock (SET lock:monitoring_worker NX EX 25)"
    monitoringWorkerPool -> redis "2. SMEMBERS server:all_ids (Fetch Active Server IDs)"
    monitoringWorkerPool -> icmpPinger "3. Dispatch to Goroutine Pool (100 workers)"
    icmpPinger -> externalServers "4. ICMP Ping(ip, timeout=3s)"
    monitoringWorkerPool -> monitoringService "5. Evaluate(serverID, ip, pingSuccess)"
    monitoringService -> observationLogger "6. LogObservation(serverID, isSuccess) [fire-and-forget]"
    observationLogger -> elastic "7. Buffered Bulk Write observation logs"
    monitoringService -> serverStateStore "8. GetServerState(serverID)"
    serverStateStore -> redis "9. HGET server:info:{id} (status, retry_count)"
    monitoringService -> serverStateStore "10. SetServerState(serverID, newStatus, retryCount)"
    serverStateStore -> redis "11. HSET server:info:{id}"
    monitoringService -> monitoringRepo "12. [Only if state changed] UpdateServerStatus(id, status, retryCount)"
    monitoringRepo -> db "13. UPDATE management_schema.servers SET current_status=?, consecutive_failures=?"
    autoLayout
}

// Reporting Module Flows
dynamic scheduler "Reporting_Scheduled" "Detailed sequence diagram for automated daily reporting (Level 4)" {
    dailyScheduler -> reportingServiceSch "1. RequestReport(adminEmail, yesterday, yesterday)"
    reportingServiceSch -> reportingRepoSch "2. CreateReportRequest(req) [Status=PENDING]"
    reportingRepoSch -> db "3. INSERT INTO reporting_schema.report_requests"
    reportingServiceSch -> reportingRepoSch "4. Worker dequeues: UpdateReportStatus(PROCESSING)"
    reportingRepoSch -> db "5. UPDATE report_requests SET status=PROCESSING"
    reportingServiceSch -> reportingRepoSch "6. GetServerCountByStatus(total/online/offline)"
    reportingRepoSch -> db "7. SELECT count FROM management_schema.servers"
    reportingServiceSch -> esUptimeCalc "8. CalculateUptime(startTime, endTime)"
    esUptimeCalc -> elastic "9. Count total & success observations in time range"
    reportingServiceSch -> notificationServiceSch "10. SendReportEmail(email, subject, htmlBody)"
    notificationServiceSch -> smtpClientSch "11. Render HTML Template & Send"
    smtpClientSch -> smtp "12. SMTP Protocol"
    reportingServiceSch -> reportingRepoSch "13. UpdateReportStatus(COMPLETED)"
    reportingRepoSch -> db "14. UPDATE report_requests SET status=COMPLETED"
    autoLayout
}

dynamic backend "Reporting_Manual" "Detailed sequence diagram for manual report requests (Level 4 - Asynchronous)" {
    admin -> frontend "1. Request Report Generation"
    frontend -> reportingHandler "2. POST /api/v1/reports"
    reportingHandler -> reportingService "3. RequestReport(email, dates)"
    reportingService -> reportingRepo "4. CreateReportRequest(req) [Status=PENDING]"
    reportingRepo -> db "5. INSERT INTO reporting_schema.report_requests"
    reportingService -> reportingWorker "6. EnqueueReport(req) → Buffered Channel"
    reportingService -> reportingHandler "7. Return immediately (async)"
    reportingHandler -> frontend "8. Return 200 OK (Accepted)"
    autoLayout
}

dynamic backend "Reporting_Worker_Process" "Detailed sequence diagram for background report generation (Level 4)" {
    reportingWorker -> reportingRepo "1. Worker dequeues job → UpdateReportStatus(PROCESSING)"
    reportingRepo -> db "2. UPDATE report_requests SET status=PROCESSING"
    reportingWorker -> reportingRepo "3. GetServerCountByStatus(total/online/offline)"
    reportingRepo -> db "4. SELECT count FROM management_schema.servers"
    reportingWorker -> esUptimeCalcBackend "5. CalculateUptime(startTime, endTime)"
    esUptimeCalcBackend -> elastic "6. Count total & success observations"
    reportingWorker -> notificationService "7. SendReportEmail(email, subject, htmlBody)"
    notificationService -> smtpClient "8. Render HTML Template & Send"
    smtpClient -> smtp "9. SMTP Protocol"
    reportingWorker -> reportingRepo "10. UpdateReportStatus(COMPLETED or FAILED)"
    reportingRepo -> db "11. UPDATE report_requests SET status=?"
    autoLayout
}
