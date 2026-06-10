package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"server-management-service/internal/modules/identity/domain"
)

// Mocks

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

type mockSessionRepo struct {
	mock.Mock
}

func (m *mockSessionRepo) Create(ctx context.Context, session *domain.AuthSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockSessionRepo) FindActiveByID(ctx context.Context, id uuid.UUID) (*domain.AuthSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthSession), args.Error(1)
}

func (m *mockSessionRepo) ListActiveByUserID(ctx context.Context, userID uint) ([]*domain.AuthSession, error) {
	args := m.Called(ctx, userID)
	var list []*domain.AuthSession
	if args.Get(0) != nil {
		list = args.Get(0).([]*domain.AuthSession)
	}
	return list, args.Error(1)
}

func (m *mockSessionRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSessionRepo) RevokeAllByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockRefreshRepo struct {
	mock.Mock
}

func (m *mockRefreshRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockRefreshRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockRefreshRepo) RevokeByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRefreshRepo) RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockRefreshRepo) RevokeByUserID(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockRefreshRepo) MarkReplacement(ctx context.Context, oldID uuid.UUID, newID uuid.UUID) error {
	args := m.Called(ctx, oldID, newID)
	return args.Error(0)
}

type mockRevoStore struct {
	mock.Mock
}

func (m *mockRevoStore) MarkRevoked(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	args := m.Called(ctx, sessionID, expiresAt)
	return args.Error(0)
}

type mockTokenMgr struct {
	mock.Mock
}

func (m *mockTokenMgr) GenerateAccessToken(userID uint, role domain.RoleCode, sessionID uuid.UUID) (string, error) {
	args := m.Called(userID, role, sessionID)
	return args.String(0), args.Error(1)
}

func (m *mockTokenMgr) GenerateRefreshToken() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockTokenMgr) HashToken(token string) string {
	args := m.Called(token)
	return args.String(0)
}

func (m *mockTokenMgr) ValidateAccessToken(tokenString string) (uint, string, string, error) {
	args := m.Called(tokenString)
	return args.Get(0).(uint), args.String(1), args.String(2), args.Error(3)
}

// Tests

func TestAuthService_Login(t *testing.T) {
	pass, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &domain.User{Model: gorm.Model{ID: 1}, Email: "test@test.com", Password: string(pass), RoleCode: "ADMIN"}

	t.Run("success", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		uRepo.On("FindByEmail", mock.Anything, "test@test.com").Return(user, nil).Once()
		tm.On("GenerateRefreshToken").Return("raw-refresh-token").Once()
		tm.On("HashToken", "raw-refresh-token").Return("hashed-token").Once()
		sRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.AuthSession")).Return(nil).Once()
		rRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()
		tm.On("GenerateAccessToken", uint(1), domain.RoleCode("ADMIN"), mock.Anything).Return("access-token", nil).Once()

		res, err := svc.Login(context.Background(), "test@test.com", "password", "127.0.0.1", "agent")
		assert.NoError(t, err)
		assert.Equal(t, "access-token", res.AccessToken)
		assert.Equal(t, "raw-refresh-token", res.RefreshToken)
	})

	t.Run("invalid password", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		uRepo.On("FindByEmail", mock.Anything, "test@test.com").Return(user, nil).Once()

		_, err := svc.Login(context.Background(), "test@test.com", "wrong", "127.0.0.1", "agent")
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("user not found", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		uRepo.On("FindByEmail", mock.Anything, "test@test.com").Return(nil, nil).Once()

		_, err := svc.Login(context.Background(), "test@test.com", "password", "127.0.0.1", "agent")
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})
}

func TestAuthService_Refresh(t *testing.T) {
	sessionID := uuid.New()
	user := &domain.User{Model: gorm.Model{ID: 1}, RoleCode: "ADMIN"}
	session := &domain.AuthSession{ID: sessionID, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}
	tokenModel := &domain.RefreshToken{ID: uuid.New(), SessionID: sessionID, UserID: 1, ExpiresAt: time.Now().Add(time.Hour)}

	t.Run("success", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		tm.On("HashToken", "token").Return("hash").Once()
		rRepo.On("FindByTokenHash", mock.Anything, "hash").Return(tokenModel, nil).Once()
		sRepo.On("FindActiveByID", mock.Anything, sessionID).Return(session, nil).Once()
		uRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil).Once()
		
		tm.On("GenerateRefreshToken").Return("new-raw").Once()
		tm.On("HashToken", "new-raw").Return("new-hash").Once()
		rRepo.On("RevokeByID", mock.Anything, tokenModel.ID).Return(nil).Once()
		rRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).Return(nil).Once()
		rRepo.On("MarkReplacement", mock.Anything, tokenModel.ID, mock.Anything).Return(nil).Once()
		tm.On("GenerateAccessToken", uint(1), domain.RoleCode("ADMIN"), sessionID).Return("new-access", nil).Once()

		res, err := svc.Refresh(context.Background(), "token")
		assert.NoError(t, err)
		assert.Equal(t, "new-access", res.AccessToken)
		assert.Equal(t, "new-raw", res.RefreshToken)
	})

	t.Run("token reused", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		now := time.Now()
		reusedToken := &domain.RefreshToken{ID: uuid.New(), SessionID: sessionID, UserID: 1, RevokedAt: &now}

		tm.On("HashToken", "token").Return("hash").Once()
		rRepo.On("FindByTokenHash", mock.Anything, "hash").Return(reusedToken, nil).Once()
		sRepo.On("ListActiveByUserID", mock.Anything, uint(1)).Return([]*domain.AuthSession{session}, nil).Once()
		sRepo.On("RevokeAllByUserID", mock.Anything, uint(1)).Return(nil).Once()
		rRepo.On("RevokeByUserID", mock.Anything, uint(1)).Return(nil).Once()
		rs.On("MarkRevoked", mock.Anything, sessionID, session.ExpiresAt).Return(nil).Once()

		_, err := svc.Refresh(context.Background(), "token")
		assert.ErrorIs(t, err, ErrInvalidRefreshToken)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		sessionID := uuid.New()
		session := &domain.AuthSession{ID: sessionID, ExpiresAt: time.Now().Add(time.Hour)}

		sRepo.On("FindActiveByID", mock.Anything, sessionID).Return(session, nil).Once()
		sRepo.On("RevokeByID", mock.Anything, sessionID).Return(nil).Once()
		rRepo.On("RevokeBySessionID", mock.Anything, sessionID).Return(nil).Once()
		rs.On("MarkRevoked", mock.Anything, sessionID, session.ExpiresAt).Return(nil).Once()

		err := svc.Logout(context.Background(), sessionID.String())
		assert.NoError(t, err)
	})
}

func TestAuthService_LogoutAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		uRepo := new(mockUserRepo)
		sRepo := new(mockSessionRepo)
		rRepo := new(mockRefreshRepo)
		rs := new(mockRevoStore)
		tm := new(mockTokenMgr)
		svc := NewAuthService(uRepo, sRepo, rRepo, rs, tm)

		sessionID := uuid.New()
		session := &domain.AuthSession{ID: sessionID, ExpiresAt: time.Now().Add(time.Hour)}

		sRepo.On("ListActiveByUserID", mock.Anything, uint(1)).Return([]*domain.AuthSession{session}, nil).Once()
		sRepo.On("RevokeAllByUserID", mock.Anything, uint(1)).Return(nil).Once()
		rRepo.On("RevokeByUserID", mock.Anything, uint(1)).Return(nil).Once()
		rs.On("MarkRevoked", mock.Anything, sessionID, session.ExpiresAt).Return(nil).Once()

		err := svc.LogoutAll(context.Background(), 1)
		assert.NoError(t, err)
	})
}
