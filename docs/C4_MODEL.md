# THIẾT KẾ KIẾN TRÚC HỆ THỐNG (C4 MODEL)

Tài liệu mô tả kiến trúc Quản lý Server (SMS) theo phương pháp C4. Các biểu đồ mô tả ranh giới hệ thống, ứng dụng, và giao tiếp nội bộ trong kiến trúc Modular Monolith.

## 1. Level 1: System Context Diagram

Xác định ranh giới ngoài của hệ thống SMS, các diễn viên (Actors) thao tác với hệ thống và các hệ thống phụ trợ bên ngoài.

```mermaid
C4Context
    title C4 Level 1: System Context Diagram - Server Management System

    Person(admin, "System Admin", "Quản trị viên hạ tầng. Quản lý danh sách thiết bị, xem báo cáo Uptime.")
    
    System(sms, "Server Management System", "Quản lý danh sách máy chủ, tự động giám sát sức khỏe mạng (ICMP), tính toán Uptime và gửi báo cáo HTML.")
    
    System_Ext(smtp, "SMTP Mail Server", "Dịch vụ Mail Relay trung gian (MailHog/SendGrid) phân phối email đến Admin.")
    System_Ext(target_servers, "Target Servers", "Hàng chục ngàn máy chủ ảo/vật lý cần giám sát kết nối.")

    Rel(admin, sms, "Thao tác quản lý & Yêu cầu báo cáo", "HTTPS / JSON")
    Rel(sms, target_servers, "Thăm dò trạng thái kết nối mạng", "ICMP Ping")
    Rel(sms, smtp, "Đẩy luồng email báo cáo", "SMTP")
    Rel(smtp, admin, "Phân phối email", "Email")
```

## 2. Level 2: Container Diagram

Bóc tách các khối tiến trình thực thi độc lập (Containers) và hạ tầng lưu trữ bên trong Server Management System.

```mermaid
C4Container
    title C4 Level 2: Container Diagram - Server Management System

    Person(admin, "System Admin", "Quản trị viên hạ tầng")

    System_Boundary(sms_boundary, "Server Management System (Modulith Architecture)") {
        Container(api_app, "API Server Application", "Go 1.22", "Phục vụ gRPC/REST API cho Identity, Server Management, Reporting. Chạy trên cổng 8000/50051.")
        Container(monitoring_worker, "Monitoring Worker Application", "Go 1.22", "Background process. Quét Redis định kỳ, gửi ICMP Ping và ghi Log.")
        Container(scheduler_app, "Daily Scheduler Application", "Go 1.22", "Cron process. Gọi gRPC báo cáo định kỳ hàng ngày.")

        ContainerDb(postgres, "PostgreSQL Database", "PostgreSQL 15", "Database gốc (Source of Truth). Lưu Users, Server Inventory, Report Status.")
        ContainerDb(redis, "Redis Cache", "Redis 7", "Lưu cache IP Servers phục vụ Monitor Worker, Distributed Locks.")
        ContainerDb(elasticsearch, "Elasticsearch", "Elasticsearch 8", "Lưu Observation Logs và tính Time-Series Uptime (CQRS Read Model).")
    }

    System_Ext(smtp, "SMTP Mail Server", "Mail Relay")
    System_Ext(target_servers, "Target Servers", "Mạng hạ tầng")

    Rel(admin, api_app, "Gọi API CRUD & Báo cáo", "JSON/REST/HTTPS")
    
    Rel(api_app, postgres, "Truy xuất/Cập nhật dữ liệu", "TCP/5432")
    Rel(api_app, redis, "Cập nhật cache IP khi đổi cấu hình", "TCP/6379")
    Rel(api_app, elasticsearch, "Truy vấn Analytics để lấy Uptime", "HTTP/9200")
    Rel(api_app, smtp, "Gửi MIME Email", "TCP/1025")

    Rel(monitoring_worker, redis, "Đọc nhanh danh sách IP cần quét", "TCP/6379")
    Rel(monitoring_worker, target_servers, "Thăm dò gói tin", "ICMP Ping")
    Rel(monitoring_worker, postgres, "Cập nhật trạng thái", "TCP/5432")
    Rel(monitoring_worker, elasticsearch, "Ghi Observation Log", "HTTP/9200")

    Rel(scheduler_app, api_app, "Gửi tín hiệu Trigger Report", "gRPC")
```

## 3. Level 3: Component Diagram

Chi tiết hóa `API Server Application`, làm rõ các ranh giới module chức năng nội bộ (Identity, Server Management, Reporting) theo phong cách Modular Monolith. Cấm tuyệt đối truy xuất DB chéo module.

```mermaid
C4Component
    title C4 Level 3: Component Diagram - API Server Application

    Container_Boundary(api_app, "API Server Application (Go Modulith)") {
        Component(identity, "Identity Module", "internal/modules/identity", "Xác thực JWT, Phân quyền RBAC.")
        Component(server_mgmt, "Server Management Module", "internal/modules/server_management", "CRUD Máy chủ, Import/Export Excel nền. Giữ đồng bộ dữ liệu với Cache.")
        Component(reporting, "Reporting Module", "internal/modules/reporting", "Xử lý hàng đợi báo cáo. Gọi Elasticsearch để tính Uptime.")
        Component(notification, "Notification Module", "internal/modules/notification", "Đóng gói dữ liệu thành giao diện HTML Template và kết nối SMTP.")
    }

    ContainerDb(postgres, "PostgreSQL", "Database Schema theo module")
    ContainerDb(redis, "Redis Cache", "Data Cache")
    ContainerDb(elasticsearch, "Elasticsearch", "Log Data")
    System_Ext(smtp, "SMTP Server", "Mail Provider")

    Rel(server_mgmt, identity, "Gọi Middleware xác thực JWT", "In-Process Call")
    Rel(reporting, identity, "Gọi Middleware xác thực JWT", "In-Process Call")
    
    Rel(reporting, server_mgmt, "Truy vấn tổng số máy chủ thông qua grpcctx", "In-Process Call")
    Rel(reporting, notification, "Gọi hàm SendReportEmail", "In-Process Call")

    Rel(identity, postgres, "Đọc/Ghi Users (GORM)", "TCP/5432")
    Rel(server_mgmt, postgres, "Đọc/Ghi Servers (GORM)", "TCP/5432")
    Rel(server_mgmt, redis, "Ghi IP vào Cache", "go-redis")
    
    Rel(reporting, postgres, "Cập nhật trạng thái Report_Requests", "TCP/5432")
    Rel(reporting, elasticsearch, "Date Histogram Aggregations", "ES Client API")
    
    Rel(notification, smtp, "Kết nối & Gửi thư", "net/smtp")
```

## 4. Bổ sung: Database Schema Boundaries

Sử dụng nguyên lý chia tách logic (Logical Schema Isolation) cho Database để giữ ranh giới sạch giữa các module. Các bảng không kết nối khóa ngoại (Foreign Key) cứng qua lại giữa các domain.

| Tên Module | Schema / Table Name | Khóa chính (PK) | Chức năng lưu trữ |
|---|---|---|---|
| **Identity** | `public.users` | `id` | Thông tin đăng nhập Admin, Role. |
| **Server Mgmt** | `management_schema.servers` | `server_id` | Cấu hình máy chủ cốt lõi (Tên, IPv4, Status). |
| **Reporting** | `reporting_schema.report_requests` | `id` | Hàng đợi yêu cầu kết xuất báo cáo (Pending/Done). |
| **Monitoring** | *(Lưu Log tại Elasticsearch)* | `_id` | Document-DB chứa Time-Series Log trạng thái mạng. |
