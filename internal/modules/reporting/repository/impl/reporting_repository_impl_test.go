package impl

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"server-management-service/internal/modules/reporting/domain"
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

func TestGormReportingRepository_CreateReportRequest(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGormReportingRepository(db)

	req := &domain.ReportRequest{
		ID:             uuid.New(),
		RequestorEmail: "admin@example.com",
		StartTime:      time.Now(),
		EndTime:        time.Now(),
		Status:         domain.ReportStatusPending,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO .*report_requests.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateReportRequest(context.Background(), req)
		assert.NoError(t, err)
	})
}

func TestGormReportingRepository_UpdateReportStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGormReportingRepository(db)

	id := uuid.New().String()

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE .*report_requests.*`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateReportStatus(context.Background(), id, domain.ReportStatusProcessing)
		assert.NoError(t, err)
	})
}

func TestGormReportingRepository_GetServerCountByStatus(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := NewGormReportingRepository(db)

	t.Run("success with status", func(t *testing.T) {
		mock.ExpectQuery(`SELECT count.*FROM .*servers.* WHERE current_status = \$1`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(150))

		count, err := repo.GetServerCountByStatus(context.Background(), "ONLINE")
		assert.NoError(t, err)
		assert.Equal(t, int64(150), count)
	})

	t.Run("success without status", func(t *testing.T) {
		mock.ExpectQuery(`SELECT count.*FROM .*servers.*`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(200))

		count, err := repo.GetServerCountByStatus(context.Background(), "")
		assert.NoError(t, err)
		assert.Equal(t, int64(200), count)
	})
}
