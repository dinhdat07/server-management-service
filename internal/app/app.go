package app

import (
	"net/http"

	"server-management-service/internal/shared/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"server-management-service/internal/modules/server_management/handler/grpcserver"
	reportinggrpc "server-management-service/internal/modules/reporting/handler/grpcserver"
	reportingsvc "server-management-service/internal/modules/reporting/service"
	authgrpc "server-management-service/internal/modules/identity/handler/grpcserver"
	notificationsvc "server-management-service/internal/modules/notification/service"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB

	GRPCServer *grpc.Server
	HTTPServer *http.Server

	RedisClient   redis.UniversalClient
	ESClient      *elasticsearch.TypedClient
	KafkaBrokers  []string
	ServerHandler    *grpcserver.ServerManagementServer
	ReportingHandler *reportinggrpc.ReportingGrpcHandler
	ReportingWorker  reportingsvc.ReportingWorker
	AuthHandler      *authgrpc.AuthServer
	NotificationService *notificationsvc.NotificationService
}
