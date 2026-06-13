package storage

import (
	"server-management-service/internal/modules/identity/domain"
	reportingDomain "server-management-service/internal/modules/reporting/domain"
	serverDomain "server-management-service/internal/modules/server_management/domain"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	db.Exec("CREATE SCHEMA IF NOT EXISTS management_schema")
	db.Exec("CREATE SCHEMA IF NOT EXISTS reporting_schema")

	return db.AutoMigrate(
		&domain.User{},
		&domain.AuthSession{},
		&domain.RefreshToken{},
		&serverDomain.Server{},
		&reportingDomain.ReportRequest{},
	)

}
