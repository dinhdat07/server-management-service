# HƯỚNG DẪN SỬ DỤNG VÀ VẬN HÀNH

## Server Management Service — `server-management-service`

> **Đối tượng:** Kỹ sư backend, DevOps, hoặc Admin mới onboarding vào dự án.
> **Phiên bản:** Go 1.22 / Nhánh `unit-test`.

---

## 1. Điều Kiện Tiên Quyết (Prerequisites)

### 1.1 Yêu Cầu Phần Mềm

| Công cụ | Phiên bản tối thiểu | Mục đích |
|---|---|---|
| **Go** | 1.22 | Build và chạy 3 binary |
| **Docker** | 24.0 | Chạy infrastructure local |
| **Docker Compose** | 2.0 | Orchestrate PostgreSQL, Redis, ES, MailHog |
| **Git** | 2.x | Clone repository |

**Không cần cài đặt thủ công:** PostgreSQL, Redis, Elasticsearch, MailHog — tất cả được quản lý bởi `docker-compose.yml`.

### 1.2 Yêu Cầu Hệ Thống (Khuyến nghị)

| Tài nguyên | Tối thiểu | Ghi chú |
|---|---|---|
| **RAM** | 4 GB | Elasticsearch chiếm ~512 MB (cấu hình `-Xms512m`) |
| **CPU** | 2 nhân | ICMP Ping pool 100 goroutines concurrent |
| **Disk** | 5 GB | ES data + PostgreSQL data |

---

## 2. Thiết Lập Môi Trường

### 2.1 Clone và Cấu Hình `.env`

```bash
# 1. Clone repository
git clone <repository-url>
cd backend/server-management-service

# 2. Sao chép file mẫu
cp .env.example .env

# 3. Chỉnh sửa .env (xem hướng dẫn chi tiết bên dưới)
# Windows: notepad .env
# Linux/macOS: vim .env hoặc nano .env
```

### 2.2 Biến Môi Trường — Giải Thích Từng Nhóm

#### 2.2.1 Application & Auth

```env
ENV=development
GRPC_PORT=50051          # Cổng gRPC nội bộ (Monitoring Worker, Daily Scheduler kết nối vào đây)
HTTP_PORT=8000           # Cổng HTTP công khai (Swagger, REST API)
FRONTEND_BASE_URL=http://localhost:4200

JWT_SECRET=super-secret-key-change-in-production
JWT_ACCESS_TTL=3600      # Access token hết hạn sau 3600 giây (1 giờ)
JWT_REFRESH_TTL=604800   # Refresh token hết hạn sau 604800 giây (7 ngày)

# Tài khoản được seed tự động vào DB khi API Server khởi động lần đầu
ADMIN_EMAIL=admin@portal.local
ADMIN_PASSWORD=Admin@123456
USER_EMAIL=user@portal.local
USER_PASSWORD=User@123456
```

> [!CAUTION]
> Biến `JWT_SECRET` **phải được thay đổi bằng một chuỗi ngẫu nhiên, bảo mật mạnh** trước khi triển khai lên môi trường Production. Không commit file `.env` chứa secret thật lên git.

#### 2.2.2 Database & Cache

```env
# DB_URL: dùng cho API Server và Daily Scheduler
DB_URL=postgresql://postgres:postgres@localhost:5432/sms

# DATABASE_URL: dùng riêng cho Monitoring Worker (format tương tự nhưng biến khác)
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/sms

REDIS_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=           # Để trống nếu Redis không có password
REDIS_DB=0

ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_SERVER_INDEX=sms_observation_logs
```

> [!IMPORTANT]
> Nếu bạn thay đổi port mapping trong `docker-compose.yml` (ví dụ đổi 5432 thành 5433 để tránh đụng độ với postgres local), bạn **phải cập nhật tương ứng** giá trị port trong các biến `DB_URL` và `DATABASE_URL`.

#### 2.2.3 SMTP (Email)

```env
# Mặc định trỏ vào MailHog (chỉ cho môi trường dev)
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM=no-reply@sms.com
SMTP_FROM_NAME=SMS Server Management
SMTP_USE_AUTH=false    # MailHog không yêu cầu auth
SMTP_USE_TLS=false

# Để dùng Gmail SMTP thật (production):
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USE_AUTH=true
# SMTP_USE_TLS=true
# SMTP_USERNAME=your-email@gmail.com
# SMTP_PASSWORD=your-app-password
```

#### 2.2.4 Monitoring Worker

```env
ICMP_PRIVILEGED=false                      # false = UDP ICMP (không cần root), true = raw socket (cần root)
MONITORING_FAILURE_THRESHOLD=2             # Số lần ping thất bại liên tiếp trước khi chuyển OFFLINE
MONITORING_WORKER_CONCURRENCY=100          # Số goroutines ping song song
MONITORING_WORKER_PING_TIMEOUT=3s          # Timeout mỗi lần ping
MONITORING_WORKER_TICK_INTERVAL=30s        # Chu kỳ quét
MONITORING_WORKER_LOCK_KEY=lock:monitoring_worker
MONITORING_WORKER_LOCK_EXPIRATION=25s      # Phải nhỏ hơn TICK_INTERVAL để tránh deadlock
```

> [!WARNING]
> **Về `ICMP_PRIVILEGED`:**
> - `false` (mặc định): Dùng UDP ICMP. Không cần quyền root nhưng một số môi trường container/cloud (như AWS EC2) mặc định block UDP ICMP.
> - `true`: Dùng raw socket. Cần chạy ứng dụng với quyền `root` (dùng `sudo`) hoặc cấp capability `NET_RAW` cho container. Hãy chọn đúng với môi trường hạ tầng của bạn.

> [!WARNING]
> **Về `MONITORING_WORKER_LOCK_EXPIRATION`:**
> Giá trị này **BẮT BUỘC** phải nhỏ hơn `MONITORING_WORKER_TICK_INTERVAL`. Nếu set lớn hơn, khi worker crash, lock chưa kịp hết hạn ở tick tiếp theo sẽ dẫn đến deadlock tạm thời, bỏ lỡ nhịp ping.

#### 2.2.5 Reporting Worker

```env
REPORTING_WORKER_COUNT=5       # Số goroutines xử lý report song song
REPORTING_JOB_QUEUE_SIZE=100   # Dung lượng buffered channel chứa các report request
```

### 2.3 Khởi Động Infrastructure

```bash
# Chạy từ thư mục server-management-service/
docker-compose up -d
```

Lệnh này khởi động 4 container:

| Container | Image | Cổng |
|---|---|---|
| `sms_postgres` | postgres:15 | `5432` |
| `sms_redis` | redis:7-alpine | `6379` |
| `sms_elasticsearch` | elasticsearch:8.17.4 | `9200`, `9300` |
| `sms_mailhog` | mailhog/mailhog | `1025` (SMTP), `8025` (Web UI) |

### 2.4 Kiểm Tra Infrastructure

```bash
# Kiểm tra tất cả container đang UP và HEALTHY
docker-compose ps

# Ping Elasticsearch (phải trả về status green/yellow)
curl http://localhost:9200/_cluster/health

# Ping Redis
docker exec sms_redis redis-cli ping
# Expected: PONG
```

> [!TIP]
> **Elasticsearch khởi động khá chậm (~30-60 giây).** Nếu `curl` trả về `connection refused`, hãy kiên nhẫn đợi thêm và thử lại trước khi chạy API Server.

---

## 3. Khởi Chạy Ứng Dụng

### 3.1 Thứ Tự Khởi Động

```
Infrastructure (docker-compose) → API Server → Monitoring Worker → Daily Scheduler
```

**Tại sao theo thứ tự này?**
- API Server sẽ thực hiện `AutoMigrate` (tạo các bảng DB) và `SeedUsers` (tạo tài khoản mặc định) khi khởi động.
- Monitoring Worker và Daily Scheduler kết nối trực tiếp vào DB, nên chúng phụ thuộc vào việc các bảng đã được API Server khởi tạo.

### 3.2 API Server

```bash
# Mở terminal 1
go run cmd/api/main.go
```

Khi khởi động thành công, log hiển thị:
```text
grpc listening on :50051
gateway listening on :8000
```

**Swagger UI:** Truy cập `http://localhost:8000/docs` để xem và test toàn bộ endpoint một cách trực quan (Swagger tự động xử lý lấy Token và Cookie).

**Health check:**
```bash
curl http://localhost:8000/health
# Expected: OK
```

### 3.3 Monitoring Worker

```bash
# Mở terminal 2
go run cmd/monitoring-worker/main.go
```

Khi khởi động thành công, log hiển thị:
```text
[MonitoringWorker] Starting pool with 100 workers
[Worker] Cron tick started, interval: 30s
```

### 3.4 Daily Scheduler

```bash
# Mở terminal 3
go run cmd/daily-scheduler/main.go
```

Scheduler này chỉ kích hoạt gửi request lúc đúng 01:00 sáng mỗi ngày.

---

## 4. Xác Thực và Phân Quyền

### 4.1 Tài Khoản Mặc Định

| Role | Email | Password | Permissions |
|---|---|---|---|
| **ADMIN** | `admin@portal.local` | `Admin@123456` | Toàn quyền: CRUD, Import, Export, Yêu cầu Report |
| **USER** | `user@portal.local` | `User@123456` | Chỉ đọc: `server:read` (chỉ được GET danh sách) |

### 4.2 Login — Lấy JWT & CSRF Token

**Endpoint:** `POST /api/v1/auth/login`

```bash
curl -i -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "identifier": "admin@portal.local",
    "password": "Admin@123456"
  }'
```

**Response thành công (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1...",
  "refresh_token": "eyJhbGciOiJIUzI1...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

> [!IMPORTANT]
> **Lưu ý về Header trả về:** Response từ Login sẽ gửi các token qua HTTP Headers `grpc-metadata-set-cookie-access-token`, `grpc-metadata-set-cookie-refresh-token`, và đặc biệt là **`grpc-metadata-set-cookie-csrf-token`**. Các ví dụ bên dưới sẽ gọi biến này là `<csrf_token_from_login>`.
> Nếu dùng cURL tự do, hãy trích xuất thủ công. Nếu dùng Swagger UI, hệ thống sẽ tự động gán.

### 4.3 Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

```bash
curl -X POST http://localhost:8000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -b cookies.txt -c cookies.txt \
  -d '{
    "refresh_token": "<refresh_token_from_login>"
  }'
```

> [!NOTE]
> Hệ thống áp dụng cơ chế **Refresh Token Rotation**. Khi được refresh thành công, token cũ sẽ bị đưa vào **Redis Blacklist** (với TTL bằng chính hạn sử dụng còn lại của token). Bất kỳ ai dùng lại token cũ sẽ bị chặn bởi Interceptor Chain (lỗi 401).

### 4.4 Logout

```bash
# Logout phiên hiện tại
curl -X POST http://localhost:8000/api/v1/auth/logout \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt -c cookies.txt
```

---

## 5. Quản Lý Server (Core Feature)

> [!IMPORTANT]
> Tất cả các endpoint thực hiện mutate dữ liệu (`POST`, `PUT`, `DELETE`) **bắt buộc** phải truyền header `-H "X-CSRF-Token: <csrf_token_from_login>"` để vượt qua `CSRFInterceptor`.

### 5.1 Xem Danh Sách Server

**Endpoint:** `GET /api/v1/servers`  
**Quyền:** `server:read` (cả ADMIN và USER)

```bash
# Filter theo status (chỉ lấy OFFLINE) và sort giảm dần theo tên
curl "http://localhost:8000/api/v1/servers?filter_status=OFFLINE&sort_by=server_name&sort_direction=desc&page=1&limit=20" \
  -b cookies.txt
```

**Response thành công (200 OK):**
```json
{
  "total_count": 1,
  "servers": [
    {
      "server_id": "550e8400-e29b-41d4-a716-446655440000",
      "server_name": "web-server-01",
      "ipv4": "192.168.1.100",
      "current_status": "OFFLINE",
      "created_at": "2025-06-01T08:00:00Z",
      "updated_at": "2025-06-10T14:30:00Z"
    }
  ]
}
```

### 5.2 Thêm Server Thủ Công

**Endpoint:** `POST /api/v1/servers`  
**Quyền:** `server:create` (chỉ ADMIN)

```bash
curl -X POST http://localhost:8000/api/v1/servers \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt \
  -d '{
    "server_name": "web-server-02",
    "ipv4": "192.168.1.102"
  }'
```

**Response lỗi khi validation fail (400 Bad Request):**
```json
{
  "code": 3,
  "message": "validation error:\n - server_name: string value must have a length greater than or equal to 3\n - ipv4: value must be a valid IP address",
  "details": []
}
```
*(Lưu ý: `protovalidate` sẽ trả về danh sách liệt kê tất cả các trường bị vi phạm trong thuộc tính `message`)*.

### 5.3 Cập Nhật Server

**Endpoint:** `PUT /api/v1/servers/{server_id}`  
**Quyền:** `server:update` (chỉ ADMIN)

```bash
curl -X PUT "http://localhost:8000/api/v1/servers/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt \
  -d '{
    "server_name": "web-server-01-renamed",
    "ipv4": "10.0.0.99"
  }'
```

### 5.4 Xóa Server

**Endpoint:** `DELETE /api/v1/servers/{server_id}`  
**Quyền:** `server:delete` (chỉ ADMIN)

```bash
curl -X DELETE "http://localhost:8000/api/v1/servers/550e8400-e29b-41d4-a716-446655440000" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt
```

### 5.5 Import Hàng Loạt từ Excel

**Endpoint:** `POST /api/v1/servers/import`  
**Quyền:** `server:import` (chỉ ADMIN)

#### Cấu Trúc File Excel Bắt Buộc

File Excel (`.xlsx`) phải có sheet chứa data từ dòng 1 với 2 cột bắt buộc (không quan trọng thứ tự): `Server Name` và `IPv4`.

```bash
curl -X POST http://localhost:8000/api/v1/servers/import \
  -H "Content-Type: application/octet-stream" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt \
  --data-binary @servers.xlsx
```

**Giới Hạn và Edge Cases (Luật xử lý Import):**
- **File > 2MB:** Trả lỗi 400.
- **Dòng bị trùng IP/Name đã có trong DB:** Sẽ bị bỏ qua và liệt kê vào mảng `failed_servers`.
- **Lỗi cục bộ:** Bất kỳ dòng nào sai định dạng IP đều bị đẩy vào `failed_servers`, các dòng đúng khác vẫn được hệ thống ghi nhận và thêm mới bình thường (Fail-safe, không rollback toàn bộ file).

### 5.6 Export ra Excel

**Endpoint:** `GET /api/v1/servers/export`  
**Quyền:** `server:export` (chỉ ADMIN)

```bash
# Export chỉ server OFFLINE tải thẳng ra file
curl "http://localhost:8000/api/v1/servers/export?filter_status=OFFLINE" \
  -b cookies.txt \
  -o offline_servers.xlsx
```

---

## 6. Giám Sát Trạng Thái (Monitoring)

### 6.1 Cách Monitoring Worker Hoạt Động

```text
[Cron tick: 30s]
    ↓
Redis: SMembers("server:all_ids")  →  Lấy toàn bộ Server ID
    ↓
[100 Goroutines song song]
  Mỗi goroutine (với 1 Server ID):
    Redis HGet("server:info:<id>", "ipv4")
    ICMP Ping(ipv4, timeout=3s)
    Evaluate FSM:
      - Ping OK → reset retryCount=0
      - Ping FAIL × 2 (Threshold=2) → đổi status=OFFLINE (cập nhật PG & Redis)
    Log Observation → Elasticsearch (mọi lượt ping đều đẩy log)
```

### 6.2 Kiểm Tra Log & Data Trực Tiếp

**Xem Uptime thô trực tiếp từ Elasticsearch:**
```bash
curl -X POST "http://localhost:9200/sms_observation_logs/_count" \
  -H "Content-Type: application/json" \
  -d '{
    "query": {
      "bool": {
        "must": [
          { "term": { "server_id": "<server_id>" } },
          { "term": { "is_success": true } }
        ]
      }
    }
  }'
```

---

## 7. Báo Cáo Uptime (Reporting)

### 7.1 Yêu Cầu Báo Cáo Thủ Công (On-demand)

**Endpoint:** `POST /api/v1/reports`  
**Quyền:** `report:request` (chỉ ADMIN)

```bash
curl -X POST http://localhost:8000/api/v1/reports \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: <csrf_token_from_login>" \
  -b cookies.txt \
  -d '{
    "target_email": "admin@portal.local",
    "start_date": "2025-06-01T00:00:00Z",
    "end_date": "2025-06-10T23:59:59Z"
  }'
```

**Quá Trình Xử Lý Bất Đồng Bộ (Async Lifecycle):**

| Bước | Hành động | Trạng Thái DB | Thời gian kỳ vọng |
|---|---|---|---|
| 1 | API lưu request vào DB và push vào Buffered Channel | `PENDING` | < 100ms |
| 2 | Worker Goroutine bốc request khỏi Channel | `PROCESSING` | < 5s |
| 3 | Worker truy vấn COUNT (3 queries tới PostgreSQL) | `PROCESSING` | < 500ms |
| 4 | Worker truy vấn Uptime (2 queries tới Elasticsearch) | `PROCESSING` | < 2s |
| 5 | Worker render HTML Template & gửi email qua SMTP | `PROCESSING` | < 5s |
| 6 | Thành công: cập nhật trạng thái | `COMPLETED` | < 50ms |
| 6b| Thất bại (SMTP lỗi, ES timeout): cập nhật trạng thái | `FAILED` | < 50ms |

### 7.2 Kiểm Tra Email

1. Truy cập Web UI của MailHog tại `http://localhost:8025`.
2. Kiểm tra Inbox sẽ nhận được HTML Report thông báo tổng lượng Server và tỷ lệ Uptime.

---

## 8. Tham Khảo Mã Lỗi API

| HTTP Status | gRPC Code | Nguyên nhân phổ biến | Cách xử lý |
|---|---|---|---|
| `400 Bad Request` | `INVALID_ARGUMENT` | Bắt Validation Error (field trống, sai chuẩn format email/IP) | Đọc key `message` để biết chính xác field nào bị reject và điều chỉnh Request. |
| `401 Unauthorized` | `UNAUTHENTICATED` | Hết hạn token, Cookie chưa truyền, Token nằm trong Blacklist | Gọi `/auth/refresh` lấy token mới hoặc Login lại. Nhớ đính kèm cookie. |
| `403 Forbidden` | `PERMISSION_DENIED` | Thiếu CSRF Header hoặc Role không đủ Scope truy cập endpoint | Đăng nhập bằng `admin@portal.local` & kiểm tra `-H "X-CSRF-Token"`. |
| `404 Not Found` | `NOT_FOUND` | Server ID, Request ID không tồn tại trên hệ thống DB | Kiểm tra ID đã copy chính xác chưa. |
| `409 Conflict` | `ALREADY_EXISTS` | Tạo/Sửa server mà bị trùng lặp `server_name` hoặc `ipv4` | Chuyển sang tên/IP khác. |
| `500 Internal Error`| `INTERNAL` | Mất kết nối Database, Redis hoặc panic code (Recovered) | Kiểm tra ngay container docker (log từ `sms_postgres`, `sms_redis`). |

---

## 9. Troubleshooting (Xử Lý Sự Cố)

### 9.1 Infrastructure Không Khởi Động Được

**Triệu chứng:** `docker-compose ps` báo có container bị `Exit` hoặc `unhealthy`.
**Checklist debug:**
1. Lấy nguyên nhân: `docker-compose logs elasticsearch`
2. Kiểm tra cổng: (Windows) `netstat -ano | findstr 9200` / (Linux) `lsof -i :9200`
3. Restart sâu (xóa cả volume nếu cần reset data trắng): `docker-compose down -v && docker-compose up -d`

### 9.2 API Trả Lỗi 403 Forbidden Khi POST/PUT/DELETE

**Nguyên nhân:** Thiếu CSRF Token.
**Checklist debug:**
1. Mở lại response của hàm Login, tìm header mang tên cấu trúc `Set-Cookie-Csrf-Token` hoặc tương đương.
2. Sao chép giá trị của nó.
3. Đính vào lệnh cURL qua cờ: `-H "X-CSRF-Token: <giá-trị>"`.

### 9.3 Monitoring Worker Không Ping Mặc Dù Có Data (Deadlock)

**Triệu chứng:** Worker log không báo lỗi, nhưng Redis/ES không có thêm log ping mới nào cả sau vài phút. Rất có thể Distributed Lock chưa được giải phóng sau một đợt worker bị crash.
**Checklist debug:**
1. Kiểm tra Lock còn kẹt không:
   ```bash
   docker exec sms_redis redis-cli GET lock:monitoring_worker
   ```
   *(Nếu trả về "1" liên tục mà không tự giải phóng)*
2. Xóa thủ công key Lock:
   ```bash
   docker exec sms_redis redis-cli DEL lock:monitoring_worker
   ```
3. Restart container worker hoặc tắt/mở lại process `go run cmd/monitoring-worker/main.go`.

### 9.4 Server Đứt Mạng Nhưng Không Bị Chuyển OFFLINE

**Checklist debug:**
1. Server có trong cache `server:all_ids`?
   ```bash
   docker exec sms_redis redis-cli SMEMBERS server:all_ids
   ```
2. Thử Ping bằng tay chính máy host đang chạy Worker: `ping <ipv4>`. Nếu không ping được, tức là bản thân máy ảo/host đó bị chặn Firewall ICMP outbound, hoặc Server target chặn inbound ICMP. 
3. Xem log Worker, xem chỉ số `consecutive_failures` đã tích lũy chạm mức `MONITORING_FAILURE_THRESHOLD` chưa.

### 9.5 Email Báo Cáo Không Được Gửi (Job Async Treo)

**Checklist debug:**
1. Xem trạng thái Async trong DB:
   ```bash
   docker exec -it sms_postgres psql -U postgres -d sms \
     -c "SELECT id, status, requestor_email, created_at FROM reporting_schema.report_requests ORDER BY created_at DESC LIMIT 5;"
   ```
   Nếu đang `PENDING` quá 5 phút → Queue bị đầy hoặc tiến trình Reporting Worker đang tắt.
   Nếu `FAILED` → Xem log terminal của process API Server (`cmd/api/main.go`), tìm filter `[ReportingWorker]` để đọc stack trace lỗi SMTP hoặc Elasticsearch Error.
