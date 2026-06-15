package rest

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"server-management-service/internal/modules/server_management/domain"
	"server-management-service/internal/modules/server_management/repository"
	"server-management-service/internal/modules/server_management/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSvc struct {
	mock.Mock
}

func (m *mockSvc) CreateServer(ctx context.Context, input service.CreateServerInput) (*domain.Server, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Server), args.Error(1)
}

func (m *mockSvc) UpdateServer(ctx context.Context, id string, input service.UpdateServerInput) (*domain.Server, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Server), args.Error(1)
}

func (m *mockSvc) DeleteServer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSvc) SearchServers(ctx context.Context, filter repository.ServerListFilter) ([]*domain.Server, int32, error) {
	args := m.Called(ctx, filter)
	var servers []*domain.Server
	if args.Get(0) != nil {
		servers = args.Get(0).([]*domain.Server)
	}
	return servers, args.Get(1).(int32), args.Error(2)
}

func (m *mockSvc) ImportServers(ctx context.Context, fileBytes []byte) (*service.ImportResult, error) {
	args := m.Called(ctx, fileBytes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ImportResult), args.Error(1)
}

func (m *mockSvc) ExportServers(ctx context.Context, filter repository.ServerListFilter) ([]byte, string, error) {
	args := m.Called(ctx, filter)
	var b []byte
	if args.Get(0) != nil {
		b = args.Get(0).([]byte)
	}
	return b, args.String(1), args.Error(2)
}

func TestImportExportHandler_HandleImport(t *testing.T) {
	mockSvc := new(mockSvc)
	handler := NewImportExportHandler(mockSvc)

	t.Run("Wrong Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/import", nil)
		rr := httptest.NewRecorder()
		handler.HandleImport(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Success Import", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		part, _ := w.CreateFormFile("file", "test.xlsx")
		part.Write([]byte("dummy excel data"))
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/import", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rr := httptest.NewRecorder()

		mockResult := &service.ImportResult{SuccessCount: 1, FailCount: 0}
		mockSvc.On("ImportServers", mock.Anything, mock.Anything).Return(mockResult, nil).Once()

		handler.HandleImport(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "successCount")
	})

	t.Run("Service Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte("raw data")))
		req.Header.Set("Content-Type", "application/octet-stream")
		rr := httptest.NewRecorder()

		mockSvc.On("ImportServers", mock.Anything, mock.Anything).Return(nil, errors.New("db err")).Once()

		handler.HandleImport(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("File Too Large", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte("raw data")))
		req.Header.Set("Content-Type", "application/octet-stream")
		rr := httptest.NewRecorder()

		mockSvc.On("ImportServers", mock.Anything, mock.Anything).Return(nil, service.ErrFileTooLarge).Once()

		handler.HandleImport(rr, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
	})

	t.Run("Bad Request Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/import", bytes.NewBuffer([]byte("raw data")))
		req.Header.Set("Content-Type", "application/octet-stream")
		rr := httptest.NewRecorder()

		mockSvc.On("ImportServers", mock.Anything, mock.Anything).Return(nil, service.ErrInvalidFormat).Once()

		handler.HandleImport(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestImportExportHandler_HandleExport(t *testing.T) {
	mockSvc := new(mockSvc)
	handler := NewImportExportHandler(mockSvc)

	t.Run("Wrong Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/export", nil)
		rr := httptest.NewRecorder()
		handler.HandleExport(rr, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("Success Export", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/export?page=1&limit=10", nil)
		rr := httptest.NewRecorder()

		mockSvc.On("ExportServers", mock.Anything, mock.Anything).Return([]byte("excel file"), "export.xlsx", nil).Once()

		handler.HandleExport(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "excel file", rr.Body.String())
	})

	t.Run("Service Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/export", nil)
		rr := httptest.NewRecorder()

		mockSvc.On("ExportServers", mock.Anything, mock.Anything).Return(nil, "", errors.New("db error")).Once()

		handler.HandleExport(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Default Limits", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/export?page=-1&limit=0", nil)
		rr := httptest.NewRecorder()

		mockSvc.On("ExportServers", mock.Anything, mock.Anything).
			Return([]byte("excel file default"), "export.xlsx", nil).Once()

		handler.HandleExport(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "excel file default", rr.Body.String())
	})
}

func TestExtractBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token123")
	assert.Equal(t, "token123", extractBearer(req))

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", extractBearer(req2))
}

func TestExtractBoundary(t *testing.T) {
	assert.Equal(t, "123", extractBoundary("multipart/form-data; boundary=123"))
	assert.Equal(t, "", extractBoundary("application/json"))
}

func TestExtractFileFromMultipart(t *testing.T) {
	mockSvc := new(mockSvc)
	handler := NewImportExportHandler(mockSvc)

	assert.Nil(t, handler.extractFileFromMultipart("application/json", []byte("bad")))
	assert.Nil(t, handler.extractFileFromMultipart("multipart/form-data; boundary=123", []byte("--123\r\nContent-Disposition: form-data; name=\"notfile\"\r\n\r\nhi\r\n--123--")))
}
