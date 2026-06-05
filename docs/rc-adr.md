# **TÀI LIỆU PHÂN TÍCH YÊU CẦU VÀ HỒ SƠ QUYẾT ĐỊNH KIẾN TRÚC (ADR)**

## **HỆ THỐNG QUẢN LÝ SERVER (SMS)**


# **PHẦN I: REQUIREMENT CLARIFICATION** 

## **1\. Tổng quan dự án (Project Overview) & Mục tiêu kinh doanh**

### **1.1 Bối cảnh dự án**

Trong hạ tầng công nghệ thông tin quy mô lớn của Công ty, việc giám sát trạng thái hoạt động của hạ tầng thiết bị là nhiệm vụ sống còn. Hệ thống **Server Management System (SMS)** được định hướng phát triển nhằm mục đích quản lý tập trung và kiểm tra tự động trạng thái hoạt động (On/Off) của toàn bộ danh sách server thuộc công ty.

Hệ thống được xây dựng dựa trên việc kế thừa và mở rộng từ nền tảng **Portal Backend** hiện tại – một kiến trúc nguyên khối lõi (Monolithic Core) kết hợp với các dịch vụ vệ tinh hướng sự kiện (Event-Driven Satellite Services) sử dụng ngôn ngữ Go.

### **1.2 Mục tiêu kinh doanh (Business Goals)**

* **Tối ưu hóa năng lực giám sát (Operational Efficiency):** Tự động hóa hoàn toàn quy trình kiểm tra trạng thái của ít nhất 10,000 server, giảm thiểu thời gian phát hiện sự cố downtime từ hàng giờ xuống dưới 1 phút.  
* **Nâng cao tính sẵn sàng của hạ tầng (Infrastructure Availability):** Cung cấp các chỉ số tính toán thời gian hoạt động liên tục (Uptime %) chính xác , giúp đội ngũ Admin tối ưu hóa tài nguyên và phát hiện sớm các server có tỷ lệ lỗi cao để xử lý kịp thời.  
* **Tái sử dụng tài nguyên (Asset Reuse):** Khai thác tối đa cấu trúc công nghệ sẵn có của dự án Portal (PostgreSQL, Redis, Kafka, Debezium, Elasticsearch, Notification Service) để giảm chi phí phát triển và đảm bảo tính nhất quán của hệ thống.

## **2\. Phạm vi hệ thống (System Scope)**

Để đảm bảo dự án Checkpoint được hoàn thành đúng hạn và tránh bẫy phình to phạm vi (Scope Creep), ranh giới hệ thống được phân định rõ ràng như sau:

### **2.1 Phạm vi xử lý (In-Scope)**

* Quản lý vòng đời (CRUD) thông tin cơ bản của Server.  
* Cơ chế kiểm tra trạng thái song song định kỳ đối với 10,000 server dựa trên địa chỉ IPv4 bằng giao thức ICMP Ping.  
* Cơ chế đồng bộ dữ liệu thay đổi (CDC) từ PostgreSQL sang Elasticsearch và Redis Cache nhằm tối ưu hiệu năng.  
* Hệ thống tính toán Uptime % dựa trên các sự kiện chuyển đổi trạng thái (Status Transition Events) lưu trữ tại Elasticsearch.  
* Hệ thống báo cáo tự động định kỳ hàng ngày và API yêu cầu báo cáo chủ động qua Email HTML.  
* Xác thực và phân quyền truy cập API bằng JWT thông qua HttpOnly Cookie.

### **2.2 Nằm ngoài phạm vi xử lý (Out-of-Scope)**

* Giám sát chuyên sâu mức ứng dụng (Application-level metrics) như dung lượng CPU, RAM, Disk, IOPS.  
* Tự động kích hoạt các kịch bản khắc phục sự cố (Auto-healing / Auto-remediation actions).  
* Cơ chế cấu hình động các bộ quy tắc cảnh báo tinh vi (Alert Rules Engine) gửi qua SMS, Telegram, hay Slack, trước hết chỉ tập trung vào kênh Email.

## **3\. Yêu cầu chức năng (Functional Requirements)**

Hệ thống được phân rã thành các nhóm tính năng cốt lõi theo bảng ma trận User Story dưới đây

| Mã Tính Năng | Tên Tính Năng | Actor | Mô tả Chi tiết Nghiệp vụ | Mức độ Ưu tiên |
| :---- | :---- | :---- | :---- | :---- |
| **REQ-F001** | Kiểm tra trạng thái tự động | System Worker | Hệ thống tự động kích hoạt Worker luồng song song (Concurrency) định kỳ mỗi 30 giây một lần. Worker lấy danh sách IP từ Redis Set , thực hiện kiểm tra bằng ICMP Ping với cơ chế thử lại (Failure Threshold \= 2). Trạng thái (On/Off) và mốc thời gian cập nhật cuối cùng (last\_updated) của server phải được ghi nhận tập trung. | Critical |
| **REQ-F002** | Tạo mới Server | Admin | Cho phép nhập thông tin: server\_name (Unique), ipv4 (Unique/Valid format). Trường server\_id được sinh tự động (UUIDv4) để đảm bảo tính duy nhất. Trả về kết quả thành công hoặc mã lỗi nếu trùng lặp thông tin.**Mở rộng**: Thêm trường check\_interval và cron\_expression nếu cần thiết | High |
| **REQ-F003** | Xem danh sách Server (View) | Admin / User | Truy vấn danh sách server hỗ trợ phân trang (Pagination), bộ lọc theo trạng thái/tên (Filter), và sắp xếp theo các trường dữ liệu tùy chọn (Sort). Dữ liệu trả ra bao gồm tổng số lượng (Total count) và mảng danh sách bản ghi. Truy vấn bắt buộc phải đọc từ Read Model (Elasticsearch) để tối ưu tốc độ. | High |
| **REQ-F004** | Cập nhật Server | Admin | Cho phép chỉnh sửa thông tin cấu hình (server\_name, ipv4) dựa trên server\_id. Hệ thống **tuyệt đối không cho phép** thay đổi trường server\_id. Kết quả trả về thông tin sau cập nhật hoặc thông báo lỗi cụ thể. | Medium |
| **REQ-F005** | Xóa Server | Admin | Cho phép xóa hoàn toàn cấu hình server khỏi PostgreSQL để duy trì tính toàn vẹn dữ liệu. Hệ thống tự động hủy kích hoạt trong Redis Cache để dừng luồng quét ICMP Ping ngay lập tức. ***\[Đặc điểm Kiến trúc\]:*** Để phục vụ tính toán Analytics chính xác, các log sự kiện quá khứ tại Elasticsearch được giữ lại và đánh dấu bằng một sự kiện kết thúc `TERMINATED`. Chỉ số Uptime % của server này chỉ được tính từ lúc tạo cho đến mốc thời gian bị xóa cứng. | Medium |
| **REQ-F006** | Import danh sách từ Excel | Admin | Cho phép tải lên file Excel (.xlsx) chứa danh sách hàng loạt server. Hệ thống phân tích cú pháp dữ liệu, kiểm tra trùng lặp server\_id hoặc server\_name đối với database hiện tại để bỏ qua (Skip) bản ghi trùng. Kết quả đầu ra phải hiển thị rõ ràng: Số lượng thành công, số lượng thất bại kèm danh sách ID chi tiết và lý do lỗi. | High |
| **REQ-F007** | Export danh sách ra Excel | Admin | Nhận các tham số lọc và sắp xếp tương tự API View Server. Kết xuất toàn bộ danh sách server phù hợp ra file định dạng Excel chuẩn cho phép Admin tải về trực tiếp từ Postman/Swagger. | High |
| **REQ-F008** | Báo cáo định kỳ tự động | System Cron | Hệ thống chạy ngầm một Cron Job tự động vào mốc giờ cố định mỗi ngày (ví dụ: 01:00 AM) cấu hình qua file .env. Thực hiện truy vấn Elasticsearch Aggregation tính toán các chỉ số của **ngày hôm trước** bao gồm: Tổng server, số lượng On, số lượng Off, và tỷ lệ Uptime trung bình. Email dạng HTML được gửi tới danh sách địa chỉ quản trị định sẵn trong hệ thống. | High |
| **REQ-F009** | API Yêu cầu Báo cáo chủ động | Admin | API POST /api/v1/servers/report tiếp nhận Input gồm khoảng ngày (start\_date, end\_date) và địa chỉ email. Hệ thống phản hồi tức thì mã 202 Accepted cho Client và đẩy tác vụ tính toán bất đồng bộ qua Kafka. Email HTML kết quả sẽ được gửi đích danh đến địa chỉ email được truyền trong Request Body. | High |

## 

## 

## **4\. Yêu cầu phi chức năng (Non-Functional Requirements)**

Các tiêu chí chất lượng của hệ thống được quy định nghiêm ngặt nhằm đáp ứng tiêu chuẩn hạ tầng doanh nghiệp lớn và khung điểm tối đa của bài Checkpoint:

### **4.1 Hiệu năng & Khả năng xử lý song song (Performance & Concurrency)**

* **Tần suất quét:** Khả năng hoàn thành việc kiểm tra trạng thái hoạt động của 10,000 mục tiêu trong vòng tối đa 5 giây kể từ khi kích hoạt chu kỳ kiểm tra, đảm bảo không xảy ra nghẽn cổ chai (Bottleneck).  
* **Độ trễ API:** 95% số lượng request truy vấn danh sách (API View Server) phải có thời gian phản hồi (Response Time) \< 100ms nhờ việc tận dụng Elasticsearch Index và Redis Cache.  
* **Quản lý lưu lượng (Throttling):** Áp dụng Distributed Rate Limiting thông qua Redis Token Bucket/Fixed Window (kế thừa từ Portal) để giới hạn tối đa 100 requests/phút đối với mỗi tài khoản Admin nhằm chống tấn công từ chối dịch vụ (DoS).

### **4.2 Khả năng mở rộng & Lưu trữ (Scalability & Data Retention)**

* **Scale-out Ready:** Thiết kế Worker Pool xử lý phi trạng thái (Stateless), sẵn sàng nhân rộng số lượng node chạy ngầm để nâng cấp khả năng giám sát từ 10,000 server lên 50,000+ server mà không cần chỉnh sửa kiến trúc cốt lõi.  
* **Chiến lược lưu trữ (Data Retention):** Dữ liệu log chuỗi thời gian (Time-series log) lưu tại Elasticsearch có tốc độ tăng trưởng rất nhanh. Do đó, hệ thống áp dụng chính sách lưu giữ dữ liệu (Retention Policy) trong vòng **90 ngày** nhằm phục vụ phân tích vận hành. Quá thời gian này, dữ liệu sẽ được tự động dọn dẹp thông qua Elasticsearch Index Lifecycle Management (ILM) để tối ưu tài nguyên lưu trữ.

### **4.3 Độ tin cậy & Tính toàn vẹn (Availability & Data Integrity)**

* **Nguyên tắc dữ liệu duy nhất (Source of Truth):** Cơ sở dữ liệu PostgreSQL là nguồn sự thật tối cao đối với thông tin cấu hình và trạng thái thực thể Server. Hệ thống Elasticsearch hoạt động độc lập như một Read Model và hoàn toàn có thể tái xây dựng chỉ mục (Rebuild Index) bất cứ lúc nào từ PostgreSQL thông qua Kafka CDC pipeline nếu xảy ra sự cố mất dữ liệu đột ngột.  
* **Xử lý bất đồng bộ (Fault Tolerance):** Toàn bộ luồng phát email và xử lý báo cáo nặng bắt buộc phải đi qua hàng đợi tin nhắn Apache Kafka , áp dụng mẫu thiết kế **Transactional Outbox Pattern** tại Core API nhằm triệt tiêu hoàn toàn rủi ro dual-write failure (ghi DB thành công nhưng lỗi kết nối mạng không bắn được message đi).

### **4.4 Bảo mật & Tuân thủ (Security & Compliance)**

* **Chống mã độc & SQL Injection:** Kiểm soát nghiêm ngặt toàn bộ tham số đầu vào. Sử dụng thư viện GORM thực hiện Parameterized Queries mặc định để triệt tiêu lỗ hổng SQL Injection. Toàn bộ IPv4 truyền vào bắt buộc đi qua interceptor validate định dạng IP thông qua buf.build/go/protovalidate.  
* **Xác thực và Phân quyền (Authentication & RBAC):** Toàn bộ hệ thống REST/gRPC API ngoại trừ luồng Login phải được bảo vệ bởi Middleware xác thực JWT token truyền qua HttpOnly Cookie chống tấn công XSS. Hệ thống phân quyền dựa trên Scope chi tiết:  
  * server:write Thêm, sửa, xóa, import server.  
  * server:read Truy vấn danh sách, hiển thị chi tiết, export server.  
  * server:report Yêu cầu báo cáo, cấu hình gửi email báo cáo định kỳ.

### **4.5 Chất lượng mã nguồn (Code Quality)**

* **Unit Testing:** Toàn bộ logic core nghiệp vụ (nhất là hàm kiểm tra trạng thái song song của Worker và hàm Aggregation tính toán Uptime % của Elasticsearch) bắt buộc phải viết Unit Test cô lập (sử dụng Mocking Framework cho các ngoại vi như DB/Cache/Kafka). Tỷ lệ bao phủ mã nguồn (Code Coverage) bắt buộc phải đạt \>= 90%.  
* **Ghi log vận hành:** Sử dụng thư viện Zap Logger ghi log cấu trúc (Structured Log dạng JSON) ra file hệ thống, tích hợp công cụ Lumberjack để thực hiện cơ chế tự động xoay vòng log (Logrotate) theo dung lượng (max 100MB/file) và số lượng file lưu trữ tối đa nhằm tránh làm tràn ổ cứng hệ thống.

# **PHẦN II: ARCHITECTURAL DECISION RECORD (ADR)**

Hồ sơ các quyết định kiến trúc lớn dưới đây đóng vai trò định hình khung kỹ thuật cho dự án SMS:

## **ADR-001: LỰA CHỌN PHƯƠNG THỨC ĐỊNH NGHĨA TRẠNG THÁI SERVER (HEALTH CHECK TYPE)**

* **Mã quyết định:** ADR-001  
  
* **Bối cảnh (Context):** Đề bài yêu cầu hệ thống định kỳ kiểm tra trạng thái On/Off của 10,000 server nhưng không quy định rõ ràng phương thức kỹ thuật cụ thể để định nghĩa một server đang sống hay chết. Việc lựa chọn sai phương thức có thể dẫn đến quá tải hạ tầng mạng hoặc yêu cầu mở rộng domain model quá sớm khi thông tin đầu vào ban đầu rất hạn chế (chỉ cung cấp trường ipv4).  
* **Quyết định (Decision):** Trong phiên bản V1, hệ thống thống nhất sử dụng giải pháp **ICMP Ping** trực tiếp đến địa chỉ IPv4 của Server để định nghĩa trạng thái hoạt động. Hệ thống cấu hình thuộc tính trừu tượng hóa thông qua một Interface HealthChecker trong Go để chuẩn bị cho khả năng mở rộng đa giao thức trong tương lai.  
* **Biện minh (Justification):** Lựa chọn ICMP Ping giải quyết triệt để bài toán Checkpoint một cách gọn gàng, tuân thủ nghiêm ngặt nguyên lý YAGNI (You Aren't Gonna Need It), tránh thiết kế quá mức (Over-engineering). Khung Abstraction `HealthChecker` đảm bảo việc bổ sung TCP/HTTP Check ở phiên bản V2 hoàn toàn cô lập, không ảnh hưởng đến kiến trúc Core chạy ngầm của Worker Pool.  
* **Hệ quả (Consequences)**:  
  * *Tích cực:* Tốc độ phát triển cực nhanh, chiếm dụng tài nguyên hệ thống tối thiểu, đáp ứng 100% yêu cầu đề bài.  
  * *Tiêu cực:* Chấp nhận rủi ro sai số "False Positive" nếu hệ điều hành của server mục tiêu cấu hình chặn gói tin ICMP (Firewall Drop Ping).  
  * *Biện pháp giảm thiểu:* Hướng dẫn đội vận hành hạ tầng whitelist cấu hình cho phép IP của hệ thống SMS đi qua tường lửa của các server thành viên.

## **ADR-002: CHIẾN LƯỢC KIỂM TRA TRẠNG THÁI SONG SONG VÀ GIẢM THIỂU FALSE POSITIVE**

* **Mã quyết định:** ADR-002  
  
* **Bối cảnh (Context):** Việc quét tuần tự (Sequential) 10,000 server với thời gian chờ phản hồi (Timeout) mặc định sẽ khiến một chu kỳ quét kéo dài hàng tiếng đồng hồ, gây nghẽn cổ chai nghiêm trọng và không thể đảm bảo tính thời gian thực của dữ liệu. Ngoài ra, mạng chập chờn nhất thời (Network Jitter) dễ dẫn đến tình trạng báo động sai (False Positive) nếu chỉ dựa vào một lần kiểm tra thất bại duy nhất.  
* **Quyết định (Decision):** Hệ thống chốt các tham số cấu hình vận hành tối ưu như sau:  
  1. **Check Interval:** 30 giây (Mỗi 30 giây chạy một chu kỳ quét mới).  
  2. **Connection Timeout:** 3 giây (Quá 3 giây không phản hồi coi như lỗi kết nối).  
  3. **Failure Threshold:** 2 (Phải lỗi liên tiếp 2 lần trong 2 chu kỳ quét mới chính thức chuyển đổi trạng thái sang **OFFLINE**).  
  4. **Recovery Threshold:** 1 (Chỉ cần 1 lần ping thành công lập tức đưa trạng thái trở lại **ONLINE**).  
  5. **Cơ chế thực thi:** Sử dụng mẫu thiết kế **Worker Pool Pattern** phối hợp với các tính năng nâng cao của Go bao gồm Goroutines, Channels, và sync.WaitGroup để kích hoạt việc kiểm tra trạng thái song song (Concurrency).  
* **Biện minh (Justification):** Việc kết hợp Check Interval 30 giây và Failure Threshold \= 2 đảm bảo tính chính xác cao. Trạng thái OFFLINE chỉ được xác nhận sau mốc thời gian ít nhất là:

Thời gian phát hiện \= Interval (30s) \+ Timeout (3s) \= 33 giây

Điều này loại bỏ hoàn toàn các cảnh báo ảo gây nhiễu do hiện tượng mất gói tin tạm thời trên hạ tầng mạng.

* **Hệ quả (Consequences):**  
  * *Tích cực:* Giám sát thời gian thực ổn định, triệt tiêu báo động sai, bảo vệ hiệu năng hệ thống lõi.  
  * *Tiêu cực:* Khối lượng log sinh ra tương đối nhiều nếu lưu trữ toàn bộ lịch sử quét.  
  * *Biện pháp giảm thiểu:* Sử dụng chiến lược lưu trữ Status Transition Event (Chi tiết tại ADR-005) để tối ưu hóa tài nguyên.

## **ADR-003: LỰA CHỌN CƠ SỞ DỮ LIỆU VÀ KIẾN TRÚC ĐỒNG BỘ DỮ LIỆU (CQRS & CDC PIPELINE)**

* **Mã quyết định:** ADR-003  
  
* **Bối cảnh (Context):** Chức năng View Server yêu cầu các tính năng tìm kiếm phức tạp, phân trang và sắp xếp đa tiêu chí. Nếu tất cả 10,000 server liên tục bị quét trạng thái và cập nhật trường status, last\_updated trực tiếp vào bảng operational database (PostgreSQL), việc thực hiện đồng thời các câu lệnh Query phức tạp (Read-Heavy) từ Admin sẽ lập tức gây ra hiện tượng khóa bảng (Table Locking), nghẽn kết nối và làm sập cơ sở dữ liệu hệ thống.  
* **Quyết định (Decision):** Hệ thống triển khai triệt để mô hình **CQRS (Command Query Responsibility Segregation)** tách biệt hoàn toàn luồng ghi và luồng đọc, kết hợp kiến trúc **CDC (Change Data Capture)** tự động kế thừa từ nền tảng Portal:  
  1. **Write Model (Source of Truth):** PostgreSQL xử lý các thao tác Create/Update/Delete cấu hình và cập nhật trạng thái core.  
  2. **Read Model:** Elasticsearch đảm nhận duy nhất vai trò phục vụ API View Server, bộ lọc nâng cao và tính toán analytics.  
  3. **Pipeline đồng bộ:** Tránh áp dụng giải pháp Dual-Write nguy hiểm từ tầng Application. API chỉ viết vào Postgres. Hệ thống sử dụng công cụ **Debezium** bám sát file log WAL (Write-Ahead Log) của Postgres để bắt trọn mọi biến động của bảng servers, tự động đẩy sự kiện qua **Apache Kafka**, và một dịch vụ **CDC Consumer** chạy ngầm sẽ tiêu thụ tin nhắn để đồng bộ hóa dữ liệu thời gian thực sang Elasticsearch.  
* **Biện minh (Justification):** Tận dụng tối đa 100% hạ tầng Event-Driven sẵn có của hệ thống Portal giúp việc hiện thực hóa mô hình này không tốn thêm chi phí hạ tầng mới mà mang lại độ an toàn dữ liệu tuyệt đối ở quy mô doanh nghiệp.  
* **Hệ quả (Consequences):**  
  * *Tích cực:* API truy vấn View đạt tốc độ phản hồi cực cao, cơ sở dữ liệu Postgres được bảo vệ an toàn tuyệt đối khỏi các tác vụ đọc nặng.  
  * *Tiêu cực:* Phát sinh gánh nặng quản lý phiên bản tài liệu (Document Versioning) trên Elasticsearch để phòng ngừa trường hợp các bản tin CDC từ Kafka bị tiêu thụ sai thứ tự (Out-of-order execution).  
  * *Biện pháp giảm thiểu:* CDC Consumer sử dụng trường phiên mã định danh hệ thống hoặc mốc last\_updated làm số hiệu phiên bản phiên dịch (Version field) khi đẩy dữ liệu vào Elasticsearch để đảm bảo bản ghi cũ hơn không thể ghi đè lên bản ghi mới hơn.

## **ADR-004: CHIẾN LƯỢC QUẢN LÝ BỘ NHỚ ĐỆM (REDIS CACHE DESIGN)**

* **Mã quyết định:** ADR-004  
  
* **Bối cảnh (Context):** Mỗi 30 giây, Worker Pool cần lấy toàn bộ thông tin của 10,000 server để chia việc cho các luồng con xử lý. Nếu Worker thực hiện truy vấn trực tiếp bằng câu lệnh SELECT xuống PostgreSQL, cơ sở dữ liệu sẽ bị quá tải kết nối liên tục. Ngược lại, nếu lưu cache toàn bộ danh sách server thành một cục chuỗi JSON lớn trong Redis, mỗi lần có hành động thêm/sửa/xóa một server đơn lẻ, hệ thống sẽ phải chịu chi phí rất lớn để phân tích cú pháp (Parse JSON), chỉnh sửa và ghi đè lại, đồng thời đối mặt với rủi ro Stale Data nặng nề khi nhiều luồng cập nhật cùng lúc.  
* **Quyết định (Decision):** Hệ thống thiết kế giải pháp lưu trữ phân tách cấu trúc trên Redis như sau:  
  1. **Redis Set (Quản lý tập hợp danh sách ID):** Sử dụng duy nhất một Key mang tên server:all\_ids. Các thành viên (Members) bên trong Set là danh sách chuỗi mã định danh duy nhất của server: \["SV-001", "SV-002", ..., "SV-10000"\].  
  2. **Redis Hash (Quản lý thông tin chi tiết từng Server):** Mỗi server được lưu riêng biệt dưới một Key định dạng: server:info:\<server\_id\>. Các trường dữ liệu bên trong Hash bao gồm: ipv4, status, retry\_count.  
  3. **Cơ chế cập nhật Cache:** Không dual-write từ API. **CDC Consumer** sau khi tiêu thụ bản tin thay đổi từ Kafka sẽ chịu trách nhiệm cập nhật trực tiếp cấu trúc bộ nhớ đệm Redis song song với việc index vào Elasticsearch.  
* **Biện minh (Justification):** Việc tách biệt thành Set và Hash giúp tối ưu hóa hiệu năng một cách triệt để. Khi Cron Job chạy, luồng chính chỉ cần gọi duy nhất lệnh SMEMBERS server:all\_ids để lấy về mảng 10,000 ID với thời gian xử lý cực nhanh \< 5ms mà hoàn toàn không chạm vào Postgres. Sau đó, các Worker con khi nhận việc sẽ gọi HGETALL server:info:\<server\_id\> để lấy riêng địa chỉ IP của server đó, đảm bảo tính độc lập và phân tán tài nguyên tuyệt vời.  
* **Hệ quả (Consequences):**  
  * *Tích cực:* Tiết kiệm dung lượng bộ nhớ RAM của Redis, loại bỏ hoàn toàn việc clear cache mù quáng toàn hệ thống, tốc độ truy cập dữ liệu cấu trúc đạt mức tối đa.  
  * *Tiêu cực:* Phát sinh kịch bản lệch dữ liệu bộ nhớ đệm nếu hệ thống Redis bị khởi động lại đột ngột và mất dữ liệu trong RAM (trong trường hợp cấu hình AOF/RDB chưa kịp ghi xuống đĩa).  
  * *Biện pháp giảm thiểu:* Xây dựng một hàm quản trị ẩn (Admin script) cho phép kích hoạt luồng quét toàn bộ bảng dữ liệu Postgres để thực hiện tác vụ làm ấm lại bộ nhớ đệm (Cache Warming) khi hệ thống khởi động lại từ sự cố mất điện (Cold Start).

## **ADR-005: CHIẾN LƯỢC LƯU TRỮ LỊCH SỬ VÀ CÔNG THỨC TÍNH TOÁN TỶ LỆ UPTIME %**

* **Mã quyết định:** ADR-005  
  
* **Bối cảnh (Context):** Để phục vụ hệ thống báo cáo, bài toán bắt buộc phải tính toán chính xác tỷ lệ thời gian hoạt động liên tục (Uptime %) của từng server cũng như toàn hạ tầng trong một khoảng thời gian tùy chọn bất kỳ. Nếu thực hiện ghi nhận log trạng thái xuống database sau mỗi lần quét (Cứ 30 giây ghi 1 dòng cho mỗi server), thì với 10,000 server, hệ thống sẽ sinh ra lượng dữ liệu khổng lồ:

  Số lượng bản ghi/ngày \= 86,400/30 \* 10,000 \= 28,800,000 bản ghi/ngày  
  Khối lượng này sẽ nhanh chóng làm sập và cạn kiệt dung lượng lưu trữ của Elasticsearch chỉ sau vài tuần vận hành.  
* **Quyết định (Decision):** Hệ thống áp dụng chiến lược lưu trữ tinh gọn mang tên **Status Transition Event Logging** phối hợp với thuật toán tính toán khoảng thời gian chi tiết:  
  1. **Cơ chế ghi log:** Hệ thống **chỉ ghi nhận log sự kiện vào Elasticsearch khi và chỉ khi có sự chuyển đổi trạng thái thực sự** của server (Ví dụ: Chuyển đổi từ ONLINE \-\> OFFLINE hoặc ngược lại từ OFFLINE \-\> ONLINE). Nếu server giữ nguyên trạng thái so với lần check trước đó, hệ thống sẽ hoàn toàn không ghi log mới.  
  2. **Cấu trúc bản ghi sự kiện (Event Model):**  
     JSON  
     {  
       "server\_id": "string",  
       "previous\_status": "string",  
       "current\_status": "string",  
       "occurred\_at": "timestamp",  
       "reason": "string"  
     }

3. **Công thức tính toán:** Tỷ lệ Uptime % trong một khoảng thời gian báo cáo được xác định theo công thức toán học sau:  
   Uptime (%) \= Tổng khoảng thời gian ONLINE / Tổng thời gian Quản lý thực tế

   Trong đó, **Tổng thời gian Quản lý thực tế (Observation Duration)** được định nghĩa bằng hàm giới hạn biên để xử lý triệt để hai trường hợp đặc biệt (Edge-cases): **Server được tạo mới ở giữa kỳ báo cáo** và **Server bị xóa cứng (Hard Delete) khỏi hệ thống trước khi kỳ báo cáo kết thúc**, mẫu số sẽ được giới hạn biên tự động dựa trên mốc sự kiện kết thúc `TERMINATED` lưu tại Elasticsearch: 

   Observation Duration \= min(terminate\_at, report\_end) \- max(server\_created\_at, report\_start)  
* **Biện minh (Justification):** Giải pháp này giúp giảm thiểu hơn 99% khối lượng dữ liệu lưu trữ phát sinh tại Elasticsearch. Trong điều kiện hạ tầng hoạt động ổn định bình thường, số lượng sự kiện chuyển đổi trạng thái sinh ra là rất ít, giúp hệ thống hoạt động vô cùng nhẹ nhàng và duy trì khả năng lưu giữ dữ liệu 90 ngày một cách dễ dàng. Khi cần tính toán Uptime, Elasticsearch sử dụng các tập lệnh truy vấn gộp nâng cao **Elasticsearch Aggregations (Date Histogram kết hợp Range Aggregation)** để quét mốc thời gian giữa các sự kiện chuyển đổi nhằm tính tổng thời gian ONLINE một cách chính xác tuyệt đối.  
* **Hệ quả (Consequences):**  
  * *Tích cực:* Tiết kiệm chi phí phần cứng lưu trữ hạ tầng tối đa, thuật toán tính toán thông minh xử lý hoàn hảo mọi trường hợp đặc biệt liên quan đến vòng đời thực thể Server.  
  * *Tiêu cực:* Logic viết câu lệnh Elasticsearch Aggregation tương đối phức tạp do phải giả định trạng thái ban đầu của server tại mốc thời điểm report\_start nếu trước đó không có sự kiện chuyển đổi nào được ghi nhận gần đó.  
  * *Biện pháp giảm thiểu:* Khi thực hiện tính toán, hệ thống sẽ chạy một câu lệnh query phụ (Sub-query) tìm kiếm bản ghi sự kiện có mốc thời gian lớn nhất nhưng nhỏ hơn report\_start để lấy trạng thái nền làm điểm xuất phát tính toán cho chuỗi thời gian tiếp theo.

## **ADR-006: THIẾT KẾ KIẾN TRÚC HỆ THỐNG BÁO CÁO BẤT ĐỒNG BỘ QUA EMAIL (ASYNC REPORTING PIPELINE)**

* **Trạng thái:** Đã phê duyệt (Approved) - *Đã cập nhật tuân thủ Bounded Context*.  
* **Bối cảnh:** Việc tổng hợp dữ liệu, quét lịch sử sự kiện trên Elasticsearch và kết xuất định dạng báo cáo cho 10,000 server là một tác vụ tiêu tốn rất nhiều tài nguyên. Nếu API thiết kế theo dạng đồng bộ (Synchronous), HTTP Connection sẽ bị giữ lại quá lâu gây Gateway Timeout. Ngoài ra, việc giao toàn bộ logic tính toán Uptime cho `Notification Service` vi phạm nghiêm trọng nguyên lý Bounded Context.  
* **Quyết định:** Triển khai giải pháp kiến trúc **Bất đồng bộ hoàn toàn (Asynchronous Processing)** dựa trên cơ chế **Transactional Outbox Pattern** kết hợp với **Report Consumer** nội bộ và **Notification Service**:  
  1. **Phản hồi Client:** API tiếp nhận yêu cầu báo cáo chủ động tại Endpoint POST /api/v1/servers/report lập tức thực hiện ghi nhận tác vụ vào bảng `outbox_events` bên trong cùng một Database Transaction của Postgres. Trả về ngay lập tức mã trạng thái **200 OK** kèm thông điệp:  
  2. JSON  
     {  
       "status": "processing",  
       "code": 200,  
       "message": "Report generation is processing in background"      
     }  
  3. **Cơ chế Outbox:** Dịch vụ **Outbox Worker** liên tục quét bảng dữ liệu, đóng gói payload và thực hiện đẩy tin nhắn (Publish) vào một Kafka Topic chuyên biệt mang tên `server.report.requested`.  
  4. **Tính toán Số liệu (Report Consumer):** Một Consumer nội bộ của `server-management-service` tiêu thụ tin nhắn từ `server.report.requested`, đảm nhận vai trò thực thi câu lệnh truy vấn Aggregation Date Histogram lên Elasticsearch để lấy số liệu Uptime %.  
  5. **Gửi Mail (Notification Service):** Sau khi tính toán xong, Report Consumer đóng gói số liệu và đẩy tin nhắn vào topic `portal.notification.requested`. Dịch vụ vệ tinh **Notification Service** chỉ nhận Data thuần, thực hiện điền dữ liệu (Render) vào biểu mẫu HTML Email `server_report` và gửi thư đến máy chủ SMTP trung gian.  
* **Biện minh (Justification):** Thiết kế này đảm bảo tuyệt đối nguyên lý **Bounded Context**, giữ cho `Notification Service` hoàn toàn generic (không biết Elasticsearch là gì). Core API luôn giữ vững trạng thái ổn định với độ trễ tối thiểu nhờ Outbox Pattern.  
* **Hệ quả (Consequences):**  
  * *Tích cực:* Triệt tiêu điểm nghẽn cổ chai ở API, dễ dàng mở rộng và bảo trì do các Domain được phân rạch ròi.  
  * *Tiêu cực:* Việc debug trở nên phức tạp do luồng xử lý đi qua 2 topic Kafka và 3 dịch vụ.  
  * *Biện pháp giảm thiểu:* Áp dụng cơ chế truyền mã định danh theo vết độc nhất **Correlation ID** xuyên suốt giúp dễ dàng tra cứu log tập trung khi có sự cố xảy ra.

## **ADR-007: QUẢN LÝ TRANSACTION CHO CÁC TÁC VỤ ĐA BƯỚC (TX MANAGER)**

* **Mã quyết định:** ADR-007

* **Bối cảnh (Context):** GORM cung cấp cơ chế auto-transaction cho các tác vụ CRUD đơn lẻ. Tuy nhiên, đối với các nghiệp vụ phức tạp liên đới nhiều bảng (Ví dụ: Lưu Report Request và Ghi Outbox Event cùng lúc, hoặc Import Excel hàng loạt), nếu xảy ra lỗi giữa chừng mà không có cơ chế rollback toàn vẹn, dữ liệu sẽ rơi vào trạng thái rác (Inconsistent).
* **Quyết định (Decision):** Áp dụng pattern `TxManager` (tương tự như `identity-service`) cho `server-management-service`. Mọi nghiệp vụ đa bước ở tầng Service bắt buộc phải được bọc trong hàm `txManager.WithTx(ctx, func(txCtx context.Context) error)`.
* **Biện minh (Justification):** Mô hình này cho phép tầng Service điều khiển luồng Transaction một cách trong suốt (Transparent) mà không bị phụ thuộc cứng vào GORM (GORM DB instance chỉ được inject ngầm qua `context.Context`), tuân thủ tuyệt đối nguyên lý Clean Architecture.
* **Hệ quả (Consequences):** Có sự lặp lại mã nguồn nhỏ (Duplicate code) so với `identity-service`, nhưng đổi lại là sự độc lập hoàn toàn (Decoupled) giữa 2 microservices.

## **ADR-008: KIẾN TRÚC XÁC THỰC CHO MICROSERVICES**

* **Mã quyết định:** ADR-008
* **Bối cảnh (Context):** `identity-service` chịu trách nhiệm cấp phát JWT. Khi người dùng gọi API quản lý Server, hệ thống cần cơ chế kiểm tra quyền hạn. Ngoài ra, việc lưu vết "Ai tạo Server này" đòi hỏi thông tin Admin.
* **Quyết định (Decision):** 
  1. **Stateless Auth:** `server-management-service` tự định nghĩa một `Auth Interceptor` ở tầng gRPC để tự động giải mã và verify JWT sử dụng chung `JWT_SECRET`, tuyệt đối không gọi đồng bộ (Synchronous Call) sang `identity-service` để hỏi quyền. Không thiết lập API Gateway tĩnh để tránh thắt cổ chai.
  2. **Database Isolation:** Bảng `servers` chỉ lưu trường `created_by_id` (UUID). Việc ghép Tên Admin vào danh sách Server sẽ do Frontend hoặc API Gateway/BFF thực hiện qua kỹ thuật **API Composition**.
* **Biện minh (Justification):** Giữ cho `server-management-service` có tốc độ phản hồi tính bằng mili-giây, chống lỗi dây chuyền (Cascading Failure) nếu `identity-service` bị sập. Việc lưu `created_by_id` đảm bảo tính chất cô lập cơ sở dữ liệu (Database per service).
* **Hệ quả và Đánh đổi (Consequences & Trade-offs):** 
  - Về SessionID: Việc xác thực không gọi chéo (Stateless) dẫn đến nguy cơ Access Token vẫn khả dụng dù tài khoản đã bị khóa. **Đã khắc phục** bằng phương pháp **Shared Redis Blacklist** (Kiểm tra chéo danh sách thu hồi trực tiếp từ Redis chung với identity-service).
  - Về Phân quyền (Authorization): Bỏ qua cơ chế kiểm tra quyền hạn chi tiết (Fine-grained Permissions) bằng cách tra cứu Database. Thay vào đó, thiết kế ưu tiên sử dụng cơ chế **Role-Based Access Control (RBAC)** thông qua trường `RoleCode` được mã hóa sẵn trong JWT. Điều này phù hợp với mô hình nghiệp vụ đơn giản ("Chỉ ADMIN được quản lý Server"), đảm bảo an toàn tuyệt đối nhờ chữ ký JWT mà không làm chậm hệ thống. Trong tương lai, nếu phát sinh nghiệp vụ phân quyền phức tạp (ABAC), có thể kiến nghị `identity-service` đính kèm mảng Permissions thẳng vào payload JWT.
