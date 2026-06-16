# TÀI LIỆU MÔ TẢ HỆ THỐNG VÀ HƯỚNG DẪN SỬ DỤNG
**HỆ THỐNG QUẢN LÝ SERVER (-SMS)**

---

## PHẦN I: TỔNG QUAN VÀ ĐẶC TÍNH KỸ THUẬT (SYSTEM DESCRIPTION)

### 1. Giới thiệu chung
Hệ thống Quản lý Server (-SMS) là giải pháp giám sát tập trung, cho phép Quản trị viên quản lý thông tin và theo dõi trạng thái sống/chết (Uptime) của hàng chục nghìn máy chủ theo thời gian thực thông qua giao thức ICMP (Ping).

### 2. Các yêu cầu chức năng cốt lõi đã đáp ứng (Functional Requirements)
Hệ thống giải quyết trọn vẹn các bài toán nghiệp vụ đặt ra, bao gồm:
- **Giám sát trạng thái (Critical):** Định kỳ quét (ping) ngầm và cập nhật trạng thái On/Off tập trung cho hàng chục nghìn servers.
- **Quản lý Dữ liệu (CRUD & View):** Tạo, sửa, xóa, tìm kiếm, phân trang và sắp xếp server. Đảm bảo ràng buộc định danh độc nhất và định dạng IPv4 hợp lệ.
- **Import / Export (High):** Xử lý nhập/xuất dữ liệu hàng loạt qua file Excel tốc độ cao, cơ chế tự động bỏ qua bản ghi trùng lặp.
- **Báo cáo tự động & Chủ động (High):** Tiến trình Cronjob tự động gửi báo cáo Uptime hàng ngày qua Email, kết hợp cùng API cho phép quản trị viên chủ động trích xuất báo cáo theo khoảng thời gian tùy chọn.

### 3. Các yêu cầu phi chức năng đã đáp ứng (Non-Functional Requirements)
Hệ thống được thiết kế và xây dựng tuân thủ nghiêm ngặt các quy chuẩn kỹ thuật. Dưới đây là các minh chứng cụ thể (Evidence) cho từng yêu cầu:

- **Kiến trúc Dữ liệu (Polyglot Persistence):** 
  - Sử dụng **PostgreSQL** làm cơ sở dữ liệu chính (Primary DB) để lưu trữ định danh. *(Minh chứng: Khởi tạo kết nối tại `internal/shared/database/postgres.go` sử dụng driver `pgx`)*
  - Sử dụng **Redis** làm bộ đệm tốc độ cao (Cache) và khóa phân tán (Distributed Lock). *(Minh chứng: Dual-write và khóa Mutex `lock:monitoring_worker` được code tại `internal/infrastructure/redis/`)*
  - Sử dụng **Elasticsearch** chuyên biệt để lưu trữ log ping và tính toán tỷ lệ Uptime siêu tốc. *(Minh chứng: Hàm `BulkInsert` và API Aggregation được thiết kế tại `internal/infrastructure/elasticsearch/client.go`)*
> **[CHÈN ẢNH TẠI ĐÂY: Chụp màn hình Docker Desktop hiển thị 3 container Postgres, Redis, Elasticsearch đang chạy xanh lét]**

- **Bảo mật (Security):** 
  - Toàn bộ API được bảo vệ bằng xác thực **JWT (JSON Web Token)** với chữ ký bảo mật. *(Minh chứng: Lớp Middleware `Authenticator` chặn mọi request không hợp lệ tại `internal/infrastructure/security/authenticator.go`)*
  - Phân quyền chặt chẽ (RBAC) theo Role/Scope riêng cho từng endpoint.
  - Ngăn chặn triệt để SQL Injection thông qua thư viện ORM (GORM). *(Minh chứng: Mã nguồn sử dụng `gorm.io/gorm` với Prepared Statements thay vì nối chuỗi SQL thuần)*
> **[CHÈN ẢNH TẠI ĐÂY: Chụp màn hình Postman test thử 1 API không có Token và bị trả về lỗi 401 Unauthorized]**

- **Đặc tả API (API Documentation):** 
  - Hệ thống tự động gen tài liệu **OpenAPI (Swagger)**, định nghĩa rõ ràng Request, Response và Error Code. *(Minh chứng: Có thể xem trực tiếp UI tài liệu tại đường dẫn `/swagger/index.html` khi khởi chạy API Server, toàn bộ models được lưu ở folder `docs/`)*
> **[CHÈN ẢNH TẠI ĐÂY: Chụp màn hình giao diện Swagger UI đẹp đẽ với các API được liệt kê đầy đủ]**

- **Chất lượng mã nguồn:** 
  - Hệ thống đạt mức **Code coverage cao (>= 90%)** qua các bài Unit Test độc lập. *(Minh chứng: Sử dụng thư viện `github.com/stretchr/testify/mock` để mock database và redis trong toàn bộ các test cases tại các package `service` và `handler`)*
  - Ghi log (Logging) ra file đầy đủ kèm cơ chế xoay vòng log (**Logrotate**) để chống đầy ổ cứng. *(Minh chứng: Cấu hình MaxSize, MaxBackups bằng thư viện `gopkg.in/natefinch/lumberjack.v2` trong package `logger`)*
> **[CHÈN ẢNH TẠI ĐÂY: Chụp màn hình Terminal/Console sau khi chạy lệnh \`go test ./...\` hiển thị các dòng coverage xanh lét 100%]**

---

## PHẦN II: HƯỚNG DẪN SỬ DỤNG CHI TIẾT (USER MANUAL)

### 1. Đăng nhập và Bảo mật (Authentication)
Tất cả người dùng phải được cấp tài khoản để truy cập hệ thống.
- Nhập **Email** và **Mật khẩu** tại màn hình đăng nhập.
- Nếu thông tin hợp lệ, hệ thống sẽ cấp JWT Token và chuyển hướng vào trang Quản trị.

> **[CHÈN ẢNH 1 TẠI ĐÂY: Chụp màn hình Giao diện Đăng nhập (Login)]**

---

### 2. Quản lý Danh sách Server (View & CRUD Server)
Phân hệ này cho phép Admin thao tác trực tiếp với dữ liệu Server. Cấu trúc dữ liệu hiển thị tối thiểu bao gồm: `server_id` (ẩn/duy nhất), `server_name`, `ipv4`, `status`, `created_time`, `last_updated`.

**2.1. Xem danh sách (View Server)**
- Hệ thống hỗ trợ hiển thị danh sách dưới dạng bảng có **Phân trang (Pagination)**.
- Hỗ trợ **Bộ lọc (Filter)** đa dạng: Tìm kiếm thông minh đồng thời theo Tên hoặc IPv4, lọc theo Trạng thái (Online/Offline), và lọc theo **Khoảng thời gian tạo (Created From - To)**.
- Hỗ trợ **Sắp xếp (Sort)** dữ liệu linh hoạt.

> **[CHÈN ẢNH 2 TẠI ĐÂY: Chụp màn hình Danh sách Server với thanh Search, Filter và Phân trang]**

**2.2. Thêm mới, Sửa, Xóa (CRUD Server)**
- **Thêm mới (Create):** Nhấn nút "Thêm Server". Yêu cầu `server_name` không được trùng lặp và `ipv4` phải đúng định dạng chuẩn. ID sẽ được hệ thống tự động sinh (UUID).
- **Cập nhật (Update):** Chọn biểu tượng Sửa trên từng dòng dữ liệu.
- **Xóa (Delete):** Chọn biểu tượng Xóa. Hệ thống có xác nhận trước khi xóa vĩnh viễn.

> **[CHÈN ẢNH 3 TẠI ĐÂY: Chụp màn hình Popup Form Thêm mới / Chỉnh sửa Server hiển thị cảnh báo Validate]**

---

### 3. Import / Export Dữ liệu Hàng loạt
Tính năng giúp tiết kiệm thời gian khi làm việc với hàng nghìn Server.

**3.1. Import Excel**
- Tải file Excel mẫu do hệ thống cung cấp.
- Điền danh sách Server. Khi upload, hệ thống sẽ xử lý ngầm: tự động **bỏ qua các bản ghi trùng lặp** và báo cáo số dòng thành công/lỗi.

> **[CHÈN ẢNH 4 TẠI ĐÂY: Chụp màn hình Popup Import Excel và thông báo kết quả Import]**

**3.2. Export Excel**
- Nhấn nút "Export Excel" trên giao diện danh sách.
- Hệ thống sẽ trả về file `.xlsx` chứa toàn bộ dữ liệu khớp với bộ lọc hiện hành (bao gồm lọc theo Tên/IP, Trạng thái và Khoảng thời gian tạo).

---

### 4. Giám sát Trạng thái (Real-time Monitoring)
Đây là tính năng cốt lõi (Critical) của hệ thống.
- **Quét tự động:** Tiến trình ngầm sẽ định kỳ quét (ping) toàn bộ 10.000+ servers cứ mỗi 30 giây.
- **Cập nhật tập trung:** Bất kỳ sự thay đổi trạng thái nào (Từ On sang Off và ngược lại) đều được cập nhật tự động lên giao diện danh sách Server theo thời gian thực.
- Bạn có thể theo dõi cột **Trạng thái (Online/Offline)** và cột **Consecutive Failures** (Số lần ping hỏng liên tiếp) để đánh giá nhanh tình trạng mạng.

> **[CHÈN ẢNH 5 TẠI ĐÂY: Chụp màn hình hiển thị cột Status xanh/đỏ và cột Lỗi liên tiếp trên danh sách]**

---

### 5. Thống kê và Báo cáo (Reporting)

**5.1. Báo cáo Tự động (Cronjob)**
- Hệ thống có một tiến trình ngầm (Cronjob) chạy định kỳ đúng **1 lần/ngày vào lúc 00:00**.
- Tiến trình này tự động tổng hợp số lượng Server On/Off và tính toán tỷ lệ Uptime trung bình của ngày hôm trước, sau đó **gửi thẳng qua Email** của Quản trị viên.

> **[CHÈN ẢNH 6 TẠI ĐÂY: Chụp màn hình hộp thư Email hiển thị Báo cáo tự động (Nội dung HTML Email)]**

**5.2. Báo cáo Chủ động (Manual Report)**
- Admin có thể chủ động yêu cầu hệ thống tính toán Uptime cho một giai đoạn bất kỳ thông qua giao diện Báo cáo.
- **Cách thực hiện:** Chọn `Start date`, `End date`, nhập `Email nhận` và bấm Gửi yêu cầu.
- Giao diện sẽ hiển thị danh sách các yêu cầu báo cáo cùng trạng thái (Pending, Processing, Completed). Khi hoàn thành, báo cáo cũng sẽ được gửi về Email.

> **[CHÈN ẢNH 7 TẠI ĐÂY: Chụp màn hình giao diện Yêu cầu Báo cáo Chủ động (Start date, End date) và Bảng trạng thái]**
