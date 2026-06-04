# **TÀI LIỆU THIẾT KẾ HƯỚNG TÊN MIỀN (DDD) VÀ ĐẶC TẢ EVENT STORMING**

## **HỆ THỐNG QUẢN LÝ SERVER (ACME-SMS) \- CHƯƠNG TRÌNH ĐÀO TẠO Acme PASSPORT**


# **PHẦN I: DOMAIN MODELING (MÔ HÌNH MIỀN)**

## **1\. Bản đồ ngữ cảnh (Context Mapping) & Phân rã Subdomain**

Để chuẩn hóa kiến trúc cho hệ thống quản lý hạ tầng quy mô lớn, tên miền tổng quát **Server Management System (ACME-SMS)** được bóc tách và phân rã thành các miền con (Subdomains) và các ngữ cảnh bị giới hạn (Bounded Contexts) biệt lập. Mô hình áp dụng nguyên tắc cô lập dữ liệu (Data Isolation) nhằm đảm bảo khả năng mở rộng quy mô lớn (Horizontal Scaling) mà không gây ra xung đột hay thắt nút cổ chai tại cơ sở dữ liệu.

### **1.1 Phân rã Subdomain (Domain Decomposition)**

* **Core Domain (Miền cốt lõi):**   
  * **Server Management:** Quản lý vòng đời thực thể thiết bị (Khai sinh, thay đổi, định danh IP, và xóa server).  
  * **Monitoring:** Chịu trách nhiệm về công cụ quét (Scheduler Engine), thực thi kiểm tra trạng thái uptime.  
  * **Reporting:** Phụ trách tổng hợp phân tích (Analytics Aggregation), xử lý các tác vụ truy vấn chuỗi thời gian nặng để kết xuất tri thức hạ tầng.  
* **Supporting Domain (Miền hỗ trợ):**  
  * **Notification:** Chịu trách nhiệm kết nối hạ tầng trung gian SMTP, quản lý biểu mẫu (Templates) và thực thi việc truyền thông điệp (Email) bất đồng bộ.

### **1.2 Bản đồ mối quan hệ giữa các Bounded Contexts**

Hệ thống Portal Core kế thừa mối quan hệ hướng sự kiện thông qua trục tin nhắn phân tán Apache Kafka:

* **Server Management Context \=\> Monitoring & Status History Context (Upstream/Downstream \- Pub/Sub):** Khi thực thể cấu hình thiết bị thay đổi (Thêm/Xóa cứng), các sự kiện biên dịch lập tức được phát ra. Chiến lược đồng bộ log WAL (Debezium CDC) đóng vai trò chuyển dịch trạng thái này sang Redis Cache và Elasticsearch mà không làm ô nhiễm Command Model tại PostgreSQL.  
* **Reporting Context \=\> Notification Context (Customer/Supplier):** Reporting đóng vai trò như khách hàng đưa ra yêu cầu; Notification Context cung cấp dịch vụ gửi thông điệp dựa trên cấu trúc Payload bất đồng bộ.

## 

## **2\. Bảng thuật ngữ miền (Ubiquitous Language)**

Việc thống nhất ngôn ngữ thống nhất là bắt buộc để triệt tiêu sự mơ hồ giữa đội ngũ phát triển sản phẩm (Product) và kỹ sư kiến trúc (Engineers). Các thuật ngữ trong hệ thống ACME-SMS được định nghĩa tường minh dưới đây:

Dưới đây là toàn bộ thông tin về các thực thể nghiệp vụ hệ thống của bạn đã được chuyển đổi thành dạng bảng Markdown để dễ dàng theo dõi, đối chiếu và tra cứu:

| Thực thể / Khái niệm | Định nghĩa Nghiệp vụ | Bounded Context Sở hữu | Quy tắc & Ràng buộc Kỹ thuật |
| :---- | :---- | :---- | :---- |
| **Server** *(Máy chủ)* | Đối tượng hạ tầng vật lý hoặc máy chủ ảo độc lập cần được quản lý và giám sát. | **Server Management Context** | \- Mỗi Server đại diện cho một thực thể hạ tầng có IPv4 duy nhất.  \- Bắt buộc phải có mã định danh duy nhất dạng UUIDv4 (server\_id) và tên không trùng lặp (server\_name). |
| **Monitoring Check** *(Lượt kiểm tra)* | Hành động thực thi một lệnh kiểm tra kết nối kỹ thuật đơn lẻ tại một mốc thời điểm cố định. | **Monitoring Context** | \- Được kích hoạt tự động theo chu kỳ mỗi 30 giây.  \- Phiên bản V1 sử dụng giao thức ICMP Ping với thời gian kết nối chờ (Timeout) là 3 giây. |
| **Monitoring Result** *(Kết quả kiểm tra)* | Kết quả phản hồi thô tức thì thu được từ một hành động *Monitoring Check*. | **Monitoring Context** | \- Chỉ có 2 trạng thái: SUCCESS (Thành công) hoặc FAIL (Thất bại/Thử lại). \- *Lưu ý quan trọng:* Kết quả này chưa phải là Trạng thái nghiệp vụ chính thức của máy chủ. |
| **Server Status** *(Trạng thái nghiệp vụ)* | Trạng thái hoạt động chính thức của một Server được hệ thống ghi nhận tập trung để hiển thị cho người dùng. | **Status History Context** | \- Chỉ bao gồm hai trạng thái: ONLINE (Bật) hoặc OFFLINE (Tắt). \- Trạng thái này được quyết định dựa trên bộ lọc chiến lược lỗi (*Failure/Recovery Threshold*). |
| **Status Transition** *(Chuyển đổi trạng thái)* | Sự thay đổi mốc trạng thái nghiệp vụ mang tính bước ngoặt của một Server | **Status History Context** | \- Là dữ liệu chuỗi thời gian bất biến (Immutable), không bao giờ được chỉnh sửa sau khi ghi nhận.\- Chỉ ghi nhận sự kiện khi trạng thái mới thực sự khác trạng thái cũ (Tuyệt đối không ghi chuỗi log thừa ONLINE \-\> ONLINE). |
| **Uptime (%)** *(Tỷ lệ hoạt động)* | Tỷ lệ phần trăm thời gian máy chủ ở trạng thái sẵn sàng hoạt động (ONLINE) trong một khoảng thời gian quan sát. | **Reporting Context** | \- Tính toán bằng cách gộp dữ liệu (Aggregation) từ lịch sử *Status Transition* trên Elasticsearch. \- Mẫu số thời gian tự động khấu trừ mốc thời gian trước khi server được tạo hoặc sau khi server bị xóa cứng (Hard Delete). |
| **Report Request** *(Yêu cầu báo cáo)* | Yêu cầu tổng hợp dữ liệu báo cáo chủ động do một tài khoản Admin khởi tạo từ API. | **Reporting Context** | \- Xử lý hoàn toàn bất đồng bộ thông qua hàng đợi tin nhắn Apache Kafka. \- Mỗi yêu cầu sinh ra một mã CorrelationId riêng biệt để phục vụ việc truy vết lỗi hệ thống phân tán. |
| **Report** *(Báo cáo)* | Kết quả tổng hợp dữ liệu hạ tầng đã qua xử lý phân tích từ Elasticsearch Aggregation. | **Reporting Context** | \- Chứa các chỉ số: Tổng số server, số lượng On, số lượng Off, tỷ lệ Uptime trung bình và danh sách top server downtime cao nhất. \- Kết xuất dưới định dạng Email HTML gửi đích danh đến hòm thư quản trị. |

## **3\. Chi tiết các thành phần Domain (Domain Artifacts)**

### 

### **3.1 Bounded Context: Server Management**

Ngữ cảnh này chịu trách nhiệm quản lý cấu hình tĩnh và vòng đời thực thể Server trong hệ thống cơ sở dữ liệu PostgreSQL (Source of Truth).

* **Entities (Thực thể):**  
  * **Server \[Aggregate Root\]:** Quản lý trạng thái snapshot của cấu hình thiết bị.  
    * serverId (Mã định danh server \- UUIDv4).  
    * serverName (Tên hiển thị độc nhất của máy chủ).  
    * ipv4 (Địa chỉ giao tiếp mạng IPv4 chuẩn).  
    * currentStatus (Snapshot trạng thái nghiệp vụ hiện tại phục vụ API View Read nhanh: ONLINE / OFFLINE).  
    * consecutiveFailures (Bộ đếm số lần thất bại liên tiếp phục vụ kiểm soát lỗi).  
    * createdAt (Mốc thời gian khởi tạo).  
    * updatedAt (Mốc thời gian cập nhật cấu hình cuối cùng).  
* **Value Objects (Đối tượng giá trị):**  
  * **IPv4Address:** Đóng gói logic kiểm tra tính hợp lệ của chuỗi ký tự theo chuẩn định dạng IP mạng toàn cầu thông qua thư viện buf.build/go/protovalidate.

### 

### **3.2 Bounded Context: Monitoring & Status History**

Ngữ cảnh chịu trách nhiệm vận hành công cụ quét song song (Worker Pool), ghi nhận các biến động và lưu trữ chuỗi lịch sử chuỗi thời gian (Time-series Events).

* **Entities (Thực thể):**  
  * **MonitoringState \[Aggregate Root\]:** Trạng thái running-time nằm trên RAM/Redis của bộ máy quét.  
    * serverId (UUIDv4 liên kết).  
    * currentRetryCount (Số lần ping lỗi tích lũy trong chu kỳ hiện tại).  
    * lastCheckTimestamp (Thời điểm thực thi kiểm tra gần nhất).  
* **Entities độc lập / Lưu trữ tại Read Model (Elasticsearch):**  
  * **StatusTransitionEvent \[Aggregate Root\]:** Bản ghi sự kiện chuyển đổi trạng thái mang tính bất biến (Immutable), phục vụ phân tích Uptime.  
    * eventId (UUIDv4 độc nhất của sự kiện).  
    * serverId (Mã server mục tiêu).  
    * previousStatus (Trạng thái nghiệp vụ cũ).  
    * currentStatus (Trạng thái nghiệp vụ mới).  
    * reason (Nguyên nhân kỹ thuật chi tiết: e.g., "ICMP Ping Timeout after 3000ms").  
    * occurredAt (Timestamp chính xác ghi nhận sự kiện chuyển đổi).

### 

### 

### **3.3 Bounded Context: Reporting**

Quản lý các yêu cầu trích xuất tri thức hạ tầng, điều phối luồng xử lý dữ liệu nặng bất đồng bộ xuyên suốt qua Apache Kafka.

* **Entities (Thực thể):**  
  * **ReportRequest \[Aggregate Root\]:** Đại diện cho một phiên làm việc xử lý báo cáo.  
    * requestId (Mã định danh phiên yêu cầu \- UUIDv4).  
    * correlationId (Mã định danh truy vết xuyên suốt hệ thống phân tán).  
    * requestedBy (Tài khoản Admin thực hiện lệnh).  
    * reportStatus (Trạng thái vòng đời tác vụ: PENDING / PROCESSING / COMPLETED / FAILED).  
    * startDate / endDate (Khoảng mốc thời gian Admin muốn quét dữ liệu trên Elasticsearch).  
    * targetEmail (Địa chỉ hòm thư đích nhận kết quả báo cáo dạng HTML).  
    * requestedAt (Mốc thời gian tiếp nhận yêu cầu).

# **PHẦN II: EVENT STORMING SPECIFICATION (BẢN TẢ SỰ KIỆN BIÊN DỊCH)**

Dưới đây là bảng đặc tả chi tiết toàn bộ các luồng tiến trình nghiệp vụ cốt lõi của hệ thống ACME-SMS

| Phân loại Luồng | Hành động / Lệnh (Command) | Tác nhân & Luật (Actor & Policy) | Sự kiện Tạo ra (Event) | Dữ liệu / Giao diện Đầu vào (Data/UI) |
| :---- | :---- | :---- | :---- | :---- |
| **Giám sát Định kỳ & Phát hiện Sự cố (Failure Flow)** | Kích Hoạt Chu Kỳ Quét Định Kỳ TriggerMonitoringSchedule | **Policy (Luật tự động):** Hệ thống Cron chạy ngầm mỗi 30 giây một lần. | Lượt Kiểm Tra Trạng Thái Đã Thực Thi MonitoringCheckExecuted | Mảng Tập Hợp Danh Sách ID Server hoạt động Redis Set: server:all\_ids |
|  | Đánh Giá Kết Quả Quét Lỗi EvaluateCheckResult | **Actor:** System Worker Pool **Policy:** Nhận kết quả ICMP Ping trả về trạng thái FAIL / Timeout. | Lượt Kiểm Tra Trạng Thái Bị Thất Bại MonitoringCheckFailed | Thông Tin Cấu Hình Hạt Nhân Server Redis Hash: server:info:\<id\> |
|  | Chuyển Đổi Trạng Thái Nghiệp Vụ Sang Offline TransitionToOffline | **Policy (Luật hệ thống):** Bộ đếm thất bại liên tiếp đạt ngưỡng giới hạn ConsecutiveFailures \== 2. | Máy Chủ Đã Mất Kết Nối Chính Thức ServerWentOffline | Lịch sử Biến Động Trạng Thái Server Elasticsearch Index |
| **Giám sát Định kỳ & Tự động Phục hồi (Recovery Flow)** | Đánh Giá Kết Quả Quét Thành Công EvaluateSuccessResult | **Actor:** System Worker Pool **Policy:** Nhận kết quả ICMP Ping trả về trạng thái thành công (Success). | Lượt Kiểm Tra Trạng Thái Đã Thành Công MonitoringCheckSucceeded | Thông Tin Cấu Hình Hạt Nhân Server Redis Hash: server:info:\<id\> |
|  | Khôi Phục Trạng Thái Nghiệp Vụ Sang Online TransitionToOnline | **Policy (Luật hệ thống):** Trạng thái hiện tại đang là OFFLINE nhưng nhận về một lượt check Thành công (Recovery Threshold \= 1). | Máy Chủ Đã Phục Hồi Hoạt Động ServerRecovered | Giao Diện Snapshot Trạng Thái Server Hiện Tại PostgreSQL / Redis Snapshot |
| **Hệ thống Báo cáo & Khởi tạo Tri thức (Reporting Pipeline)** | Yêu Cầu Xuất Báo Cáo Hạ Tầng RequestInfrastructureReport | **Actor:** Admin (Có scope server:report) **Policy:** Hệ thống ghi nhận tác vụ vào Outbox Table, lập tức trả về mã HTTP 202 cho Client. | Yêu Cầu Tạo Báo Cáo Đã Tiếp Nhận ReportRequested | Form cấu hình tham số nhập liệu ngày thống kê trên Swagger UI / Postman |
|  | Thực Thi Tổng Hợp Phân Tích Dữ Liệu Ngầm AggregateHistoricalData | **Policy (Luật tự động):** Outbox Worker quét bảng dữ liệu, đẩy tin nhắn qua Kafka, kích hoạt Notification Service tiêu thụ bất đồng bộ. | Dữ Liệu Báo Cáo Đã Tổng Hợp Thành Công ReportGenerated | Dữ Liệu Lịch Sử Sự Kiện Chuỗi Thời Gian Elasticsearch Aggregations Engine |
|  | Kết Xuất Giao Diện HTML & Truyền Thư Điện Tử RenderAndSendEmail | **Actor:** Standalone Notification Service Microservice | Báo Cáo Đã Được Gửi Tới Thư Điện Tử Admin ReportEmailSent | Biểu mẫu HTML Email Template định dạng chuyên nghiệp sẵn có |
|  | Xử Lý Sự Cố Kết Nối Hệ Thống Thư Điện Tử HandleSmtpFailure | **Policy (Luật tự động):** Máy chủ SMTP bị mất kết nối mạng bất ngờ, kích hoạt Circuit Breaker và luồng thử lại Exponential Backoff ngầm. | Gửi Thư Điện Tử Báo Cáo Bị Thất Bại ReportEmailFailed | Hàng đợi tin nhắn Dead Letter Queue (DLQ) nội bộ để kiểm tra vết lỗi |

  \<td row

# 

# 

# 

# **PHẦN III: INVARIANTS & READ MODELS (LUẬT NGHIỆP VỤ & MÔ HÌNH ĐỌC)**

## **1\. Quy tắc bất biến hệ thống (Business Invariants)**

Luật nghiệp vụ là các quy tắc luôn luôn phải đúng trong suốt vòng đời hoạt động của hệ thống, được thực thi nghiêm ngặt tại tầng Core Domain lớp nghiệp vụ:

* **Invariant 1 (Duy nhất định danh thực thể):** Trường serverId (UUIDv4) và serverName phải là độc nhất tuyệt đối trong toàn bộ hệ thống cơ sở dữ liệu Postgres tại cùng một thời điểm máy chủ đang hoạt động.  
* **Invariant 2 (Bảo toàn IP hạ tầng):** Một địa chỉ ipv4 chỉ được cấp phát duy nhất cho một serverId ở trạng thái Active. Khi một server bị xóa cứng (Hard Delete), địa chỉ IP này mới được giải phóng hoàn toàn và cho phép gán cho một thực thể ID mới tiếp theo trong tương lai mà không làm ảnh hưởng đến tính cô lập dữ liệu lịch sử trên Elasticsearch.  
* **Invariant 3 (Ranh giới trạng thái nghiệp vụ):** Thực thể Server chỉ được phép tồn tại ở một trong hai trạng thái nghiệp vụ duy nhất: ONLINE hoặc OFFLINE. Các trạng thái thô kỹ thuật (e.g., Timeout, Connection Refused) bắt buộc phải được dịch mã thông qua bộ quy tắc lọc (Failure Threshold Strategy) trước khi gán về trạng thái nghiệp vụ.  
* **Invariant 4 (Quy tắc chuyển dịch trạng thái mất kết nối):** Hệ thống chỉ được phép chuyển trạng thái nghiệp vụ của máy chủ sang OFFLINE và phát ra sự kiện ServerWentOffline khi và chỉ khi bộ đếm lỗi của node đó ghi nhận liên tiếp đủ 2 chu kỳ quét thất bại (consecutiveFailures \== 2). Tất cả các lượt check lỗi đơn lẻ lần thứ 1 chỉ được cập nhật bộ đếm phụ, trạng thái nghiệp vụ bắt buộc giữ nguyên là ONLINE để triệt tiêu báo động sai ảo (Network Jitter).  
* **Invariant 5 (Quy tắc chuyển dịch trạng thái khôi phục):** Khi một máy chủ đang ở trạng thái OFFLINE, hệ thống chỉ cần ghi nhận duy nhất 1 lượt quét ICMP Ping trả về kết quả thành công (SUCCESS) để lập tức đưa trạng thái nghiệp vụ trở lại ONLINE và phát ra sự kiện ServerRecovered.  
* **Invariant 6 (Tối ưu hóa ghi dữ liệu lịch sử \- Hạn chế trùng lặp log):** Hệ thống nghiêm cấm hành vi ghi log sự kiện lặp lại vào Elasticsearch nếu trạng thái kiểm tra của chu kỳ mới trùng khít với trạng thái hiện hữu của chu kỳ cũ (Tuyệt đối không ghi nhận các chuỗi log dư thừa dạng ONLINE → ONLINE hoặc OFFLINE → OFFLINE). Lịch sử biến động chỉ được lưu khi và chỉ khi xảy ra sự chuyển đổi trạng thái thực sự (Status Transition).  
* **Invariant 7 (Tính bất biến của dữ liệu quá khứ):** Toàn bộ các tài liệu sự kiện chuyển đổi trạng thái (StatusTransitionEvent) sau khi đã được CDC Consumer đẩy vào index của Elasticsearch là dữ liệu lịch sử bất biến (Immutable), hệ thống tuyệt đối không cung cấp bất kỳ API nào cho phép sửa đổi hay cập nhật các log quá khứ này để đảm bảo tính khách quan tối cao khi audit hệ thống an ninh.

## 

## **2\. Đặc tả các Mô hình Đọc chuyên biệt (Read Models Design)**

Để đảm bảo hiệu năng độ trễ thấp (Response Time \< 100ms) phục vụ các Admin vận hành trong điều kiện hệ thống liên tục chịu tải quét trạng thái từ 10,000 server, các Read Models được thiết kế phân tách chuyên sâu:

### 

### **2.1 Mô hình Đọc Danh Sách Máy Chủ (Server List View Model)**

* **Nguồn dữ liệu gốc:** Đồng bộ thời gian thực từ bảng public.servers của Postgres thông qua Debezium CDC Pipeline sang Elasticsearch.  
* **Cấu trúc dữ liệu thiết kế:** Elasticsearch Document (index: sms\_server\_catalog).  
* **Mục đích phục vụ nghiệp vụ:** Đáp ứng hoàn hảo cho API **REQ-F003 (View Server)**. Hỗ trợ phân trang hiệu năng cao, tìm kiếm tìm gần đúng theo tên server\_name và lọc cực nhanh theo trạng thái hệ thống status.

### 

### **2.2 Mô hình Đọc Lịch Sử Biến Động Thiết Bị (Status History View Model)**

* **Nguồn dữ liệu gốc:** Chuỗi sự kiện chuyển đổi trạng thái (`StatusTransitionEvent`) được sinh ra từ luồng giám sát tự động của Monitoring Worker Pool và được lưu bền vững trong PostgreSQL dưới vai trò là nguồn dữ liệu chuẩn (Source of Truth).  
  Các sự kiện này sau đó được đồng bộ bất đồng bộ thông qua cơ chế Change Data Capture (CDC) sử dụng Debezium. Debezium theo dõi PostgreSQL WAL, phát sinh các CDC event lên Kafka, và Projection Worker tiêu thụ các sự kiện này để xây dựng Read Model trong Elasticsearch  
* **Cấu trúc dữ liệu thiết kế:** Time-series Event Document (index: sms\_status\_transition\_logs).  
* **Mục đích phục vụ nghiệp vụ:** Cung cấp dữ liệu nguồn hiển thị biểu đồ dòng thời gian biến động cho API xem chi tiết lịch sử lỗi của một máy chủ cụ thể theo mã định danh UUIDv4 (GET /api/v1/servers/{id}/history).


### **2.3 Mô hình Đọc Tính Toán Báo Cáo Uptime (Uptime Report Analytics View Model)**

* **Nguồn dữ liệu gốc:** Truy vấn gộp nâng cao (Elasticsearch Aggregations Engine \- Quét đồng thời Date Histogram kết hợp Range Filters) dựa trên dữ liệu của index: sms\_status\_transition\_logs.  
* **Thuật toán thực thi cốt lõi:** Quét chuỗi thời gian được giới hạn biên tự động dựa trên mốc sự kiện kết thúc TERMINATED để xử lý triệt để bài toán xóa cứng thiết bị hạ tầng:  
* **Mục đích phục vụ nghiệp vụ:** Tính toán chính xác tỷ lệ sẵn sàng hoạt động (%) của một thiết bị độc lập hoặc toàn bộ cụm hạ tầng công ty Acme mà không bị tính phạt oan khoảng thời gian sau khi server đã bị xóa cứng ra khỏi hệ thống.


### **2.4 Mô hình Đọc Tổng Hợp Báo Cáo Định Kỳ Hàng Ngày (Daily Summary Report View Model)**

* **Nguồn dữ liệu gốc:** Elasticsearch Metrics Aggregation Scheduler job chạy ngầm.  
* **Cấu trúc dữ liệu thiết kế:** JSON Snapshot Object được lưu trữ tạm thời trong bảng outbox tác vụ hoặc gửi trực tiếp Payload qua Kafka topic server.report.requested.  
* **Mục đích phục vụ nghiệp vụ:** Điền dữ liệu tự động (Auto-render) vào biểu mẫu HTML Email gửi cho toàn bộ danh sách Admin đăng ký trong hệ thống vào đầu giờ sáng mỗi ngày, hiển thị rõ ràng: Tổng máy chủ, tỷ lệ Uptime trung bình toàn công ty, và danh sách đen các server có thời gian downtime cao nhất để đội an ninh mạng kịp thời xử lý hạ tầng kỹ thuật.

