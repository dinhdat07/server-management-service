# TÀI LIỆU HƯỚNG DẪN SỬ DỤNG VÀ VẬN HÀNH (USER GUIDE)

## 1. Yêu Cầu Môi Trường (Prerequisites)

| Thành phần | Phiên bản | Ghi chú |
|---|---|---|
| Go | >= 1.22 | Môi trường build và chạy mã nguồn |
| Docker | >= 24.0 | Cung cấp hạ tầng Local |
| Docker Compose | >= 2.0 | Orchestration hạ tầng Local |
| PostgreSQL | 15.x | Database lưu trữ cấu hình (OLTP) |
| Redis | 7.x | Cache dữ liệu Server & Distributed Lock |
| Elasticsearch | 8.x | Lưu Observation Log & tính Uptime (OLAP) |
| MailHog | latest | Test bắt Email SMTP Local |

## 2. Cấu Hình Hệ Thống

Các biến môi trường được quản lý tại file `.env` gốc dự án (`backend/server-management-service/.env`).

*   **Database:** `DB_URL` (cho API) và `DATABASE_URL` (cho Monitoring Worker).
*   **Redis:** Cấu hình `REDIS_ADDR`, `REDIS_DB`.
*   **Elasticsearch:** `ELASTICSEARCH_URL`, `ELASTICSEARCH_SERVER_INDEX`.
*   **SMTP:** `SMTP_HOST=localhost`, `SMTP_PORT=1025` (Mặc định cho MailHog).

## 3. Hướng Dẫn Khởi Chạy

### 3.1 Khởi chạy hạ tầng Local (Dependencies)

Mở terminal tại `backend/server-management-service` và chạy:

```bash
docker-compose up -d
```

Các cổng dịch vụ sẽ được mở:
*   Postgres: `5432`
*   Redis: `6379`
*   Elasticsearch: `9200`
*   MailHog Web UI: `http://localhost:8025`

### 3.2 Khởi chạy ứng dụng (Modulith Processes)

Hệ thống cung cấp 3 tiến trình thực thi độc lập (có thể mở 3 terminal riêng):

1.  **API Server (REST/gRPC):**
    ```bash
    go run cmd/api/main.go
    ```
2.  **Monitoring Worker (Background Ping):**
    ```bash
    go run cmd/monitoring-worker/main.go
    ```
3.  **Daily Scheduler (Cron Trigger):**
    ```bash
    go run cmd/daily-scheduler/main.go
    ```

## 4. Vận Hành Tính Năng

### 4.1 Quản Lý Server (CRUD & Import/Export)

*   **Giao diện API:** Truy cập `http://localhost:8000/api/v1/swagger-ui` để thao tác trực tiếp qua Swagger.
*   **Xác thực (JWT):** Gọi API `/api/v1/auth/login`. Thông tin mặc định:
    *   Email: `admin@portal.local`
    *   Password: `Admin@123456`
*   **Import Excel:** Upload file mẫu với cấu trúc 2 cột chính: `ServerName`, `IPv4`.

*[Placeholder: Ảnh chụp màn hình Swagger UI thao tác Import Excel]*

### 4.2 Giám Sát Trạng Thái (Monitoring)

*   Worker chạy nền, quét Redis 30 giây/lần để lấy danh sách IP.
*   Sử dụng ICMP Ping. Trạng thái tự động cập nhật `OFFLINE` nếu trượt 2 lần liên tiếp.
*   Log trạng thái chi tiết được đẩy sang Elasticsearch (`sms_observation_logs`).

*[Placeholder: Ảnh chụp màn hình Console log của tiến trình Monitoring Worker đang ping]*

### 4.3 Báo Cáo & Email (Reporting)

*   **Báo cáo thủ công (On-demand):** Gọi API `/api/v1/reports` qua Swagger, truyền `start_time`, `end_time` và `email`.
*   **Báo cáo tự động:** Do tiến trình `daily-scheduler` gọi gRPC nội bộ kích hoạt hàng ngày.
*   **Kiểm tra Email:** Truy cập Web UI của MailHog tại `http://localhost:8025` để xem HTML report gửi thành công.

*[Placeholder: Ảnh chụp màn hình MailHog chứa template HTML báo cáo Uptime]*
