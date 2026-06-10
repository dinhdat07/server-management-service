package impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	serverDomain "server-management-service/internal/modules/server_management/domain"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return gormDB, mock
}

func TestGormMonitoringRepository_UpdateServerStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGormMonitoringRepository(db)

	id := "srv-1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE .*servers.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateServerStatus(context.Background(), id, serverDomain.ServerStatusOffline, 3)
		assert.NoError(t, err)
	})
}
