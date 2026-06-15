// System Level Relationships
admin -> sms "Uses"
sms -> smtp "Sends emails using"
sms -> externalServers "Pings and manages"

// Container Level Relationships
admin -> frontend "Visits [HTTPS]"
frontend -> backend "Makes API calls to [JSON/gRPC]"
scheduler -> db "Reads/Writes [Embedded Reporting]"
scheduler -> elastic "Reads logs [Embedded Reporting]"
scheduler -> smtp "Sends emails [Embedded Notification]"
monitorWorker -> redis "Reads targets & state, acquires distributed lock"
monitorWorker -> db "Updates server status in"
monitorWorker -> elastic "Writes observation logs to"
monitorWorker -> externalServers "Sends ICMP Ping to"
backend -> smtp "Sends notifications via"
backend -> db "Reads/Writes"
backend -> redis "Reads/Writes (Sessions, Server Cache, Rate Limit)"
backend -> elastic "Reads uptime logs from"

// Level 3: Component Relationships (High-Level)
frontend -> identityComp "Login, Logout & Verify Token"
identityComp -> db "Queries/Updates users & sessions"
identityComp -> redis "Reads/Writes Revoked Sessions"

frontend -> serverComp "CRUD & Import/Export"
serverComp -> db "Queries servers"
serverComp -> redis "Dual-Write Server State Cache"

frontend -> reportingComp "gRPC requests"
reportingComp -> db "Queries data"
reportingComp -> elastic "Queries uptime logs"
reportingComp -> notificationComp "Triggers email sending"

schedulerComp -> db "Queries aggregated data"
schedulerComp -> elastic "Queries uptime"
schedulerComp -> smtp "Sends email to"

notificationComp -> smtp "Sends email to"

monitoringComp -> redis "Fetches target IPs & acquires distributed lock"
monitoringComp -> externalServers "ICMP Ping"
monitoringComp -> elastic "Saves observation logs (buffered bulk)"
monitoringComp -> db "Updates server status (only on state change)"

// Level 4: Identity Module Relationships
frontend -> authHandler "REST & gRPC Calls"
authHandler -> authService "Delegates business logic to"
authService -> userRepo "Reads/Writes Users"
authService -> sessionRepo "Reads/Writes Sessions"
authService -> refreshRepo "Reads/Writes Refresh Tokens"
userRepo -> db "Executes SQL"
sessionRepo -> db "Executes SQL"
refreshRepo -> db "Executes SQL"
authService -> redis "Reads/Writes Revoked Sessions"

// Level 4: Server Management Module Relationships
frontend -> serverHandler "REST & gRPC Calls"
serverHandler -> serverService "Delegates business logic to"
serverService -> serverRepo "Reads/Writes Servers"
serverService -> serverCache "Dual-Write Server State"
serverRepo -> db "Executes SQL"
serverCache -> redis "HSET/SADD/DEL Server Info"

// Level 4: Reporting Module (Backend) Relationships
frontend -> reportingHandler "gRPC Calls"
reportingHandler -> reportingService "Delegates business logic to"
reportingService -> reportingRepo "Creates Report Requests"
reportingService -> reportingWorker "Enqueues async report job"
reportingWorker -> reportingRepo "Reads Server Metadata & Updates Status"
reportingWorker -> esUptimeCalcBackend "Calculates Uptime"
esUptimeCalcBackend -> elastic "Queries Observation Logs"
reportingRepo -> db "Executes SQL"
reportingWorker -> notificationService "Triggers Email Rendering"

// Level 4: Notification Module (Backend) Relationships
notificationService -> smtpClient "Delegates Email Sending"
smtpClient -> smtp "Sends Email Protocol"

// Level 4: Monitoring Module Relationships
monitoringWorkerPool -> redis "Acquires Distributed Lock & Fetches Target Server IDs"
monitoringWorkerPool -> icmpPinger "Delegates ICMP Ping"
icmpPinger -> externalServers "Sends ICMP Ping"
monitoringWorkerPool -> monitoringService "Delegates State Evaluation"
monitoringService -> serverStateStore "Reads/Writes Server State"
serverStateStore -> redis "Executes Redis Commands (HGET/HSET)"
monitoringService -> monitoringRepo "Writes Server Status (only on state change)"
monitoringRepo -> db "Executes SQL"
monitoringService -> observationLogger "Fire-and-forget observation log"
observationLogger -> elastic "Buffered Bulk Write"

// Level 4: Scheduler Module Relationships
dailyScheduler -> reportingServiceSch "Triggers Daily Reporting"
reportingServiceSch -> reportingRepoSch "Reads Server Metadata"
reportingRepoSch -> db "Executes SQL"
reportingServiceSch -> esUptimeCalc "Calculates Uptime"
esUptimeCalc -> elastic "Queries Uptime Stats"
reportingServiceSch -> notificationServiceSch "Triggers Email Rendering"
notificationServiceSch -> smtpClientSch "Delegates Email Sending"
smtpClientSch -> smtp "Sends Email Protocol"
