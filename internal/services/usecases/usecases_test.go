package usecases

import (
	"context"
	"errors"
	"testing"

	"MydroX/anicetus/internal/services/dto"
	"MydroX/anicetus/internal/services/mocks"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	valkeymock "github.com/valkey-io/valkey-go/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testDeps struct {
	mockStore *mocks.MockServiceStore
	usecases  ServicesUsecases
}

func setupTest(t *testing.T) testDeps {
	ctrl := gomock.NewController(t)
	log, _ := logger.New("TEST")

	mockStore := mocks.NewMockServiceStore(ctrl)

	valkeyClient := valkeymock.NewClient(ctrl)
	valkeyClient.EXPECT().Do(gomock.Any(), gomock.Any()).Return(valkeymock.ErrorResult(errors.New("cache miss"))).AnyTimes()

	manager := NewServiceManager(log, mockStore, valkeyClient)
	uc := New(log, mockStore, manager)

	return testDeps{mockStore: mockStore, usecases: uc}
}

func TestRegisterService(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deps.mockStore.EXPECT().RegisterService(gomock.Any(), "my-svc", gomock.Any()).Return(nil)

		err := deps.usecases.RegisterService(ctx, &dto.RegisterServiceRequest{
			Audience:    "my-svc",
			ServiceName: "My Service",
		})
		assert.NoError(t, err)
	})

	t.Run("store error", func(t *testing.T) {
		deps.mockStore.EXPECT().RegisterService(gomock.Any(), "dup", gomock.Any()).
			Return(errs.New(errs.ErrorUniqueViolation, "exists", nil))

		err := deps.usecases.RegisterService(ctx, &dto.RegisterServiceRequest{
			Audience:    "dup",
			ServiceName: "Dup",
		})
		assert.Error(t, err)
	})
}

func TestRevokeService(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deps.mockStore.EXPECT().RevokeService(gomock.Any(), "my-svc").Return(nil)

		assert.NoError(t, deps.usecases.RevokeService(ctx, "my-svc"))
	})

	t.Run("store error", func(t *testing.T) {
		deps.mockStore.EXPECT().RevokeService(gomock.Any(), "unknown").
			Return(errs.New(errs.ErrorNotFound, "not found", nil))

		assert.Error(t, deps.usecases.RevokeService(ctx, "unknown"))
	})
}

func TestGetAllServices_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	// Cache miss -> falls back to repo (may be called again by background goroutine)
	deps.mockStore.EXPECT().GetAllowedServices(gomock.Any()).Return([]string{"svc-a", "svc-b"}, nil).AnyTimes()

	services, err := deps.usecases.GetAllServices(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"svc-a", "svc-b"}, services)
}

func TestGetAllServices_StoreError(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	deps.mockStore.EXPECT().GetAllowedServices(gomock.Any()).
		Return(nil, errs.New(errs.ErrorInternal, "db error", nil)).AnyTimes()

	_, err := deps.usecases.GetAllServices(ctx)
	assert.Error(t, err)
}

func TestGetUserServices_Success(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	deps.mockStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid").Return([]string{"svc-1"}, nil).AnyTimes()

	services, err := deps.usecases.GetUserServices(ctx, "user-uuid")
	require.NoError(t, err)
	assert.Equal(t, []string{"svc-1"}, services)
}

func TestGetUserServices_StoreError(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	deps.mockStore.EXPECT().GetUserServices(gomock.Any(), "user-uuid").
		Return(nil, errs.New(errs.ErrorInternal, "err", nil)).AnyTimes()

	_, err := deps.usecases.GetUserServices(ctx, "user-uuid")
	assert.Error(t, err)
}

func TestAssignServiceToUser(t *testing.T) {
	deps := setupTest(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		deps.mockStore.EXPECT().AssignServiceToUser(gomock.Any(), "user-uuid", "my-svc").Return(nil)

		err := deps.usecases.AssignServiceToUser(ctx, "user-uuid", &dto.AssignServiceRequest{Audience: "my-svc"})
		assert.NoError(t, err)
	})

	t.Run("store error", func(t *testing.T) {
		deps.mockStore.EXPECT().AssignServiceToUser(gomock.Any(), "user-uuid", "bad").
			Return(errs.New(errs.ErrorNotFound, "not found", nil))

		err := deps.usecases.AssignServiceToUser(ctx, "user-uuid", &dto.AssignServiceRequest{Audience: "bad"})
		assert.Error(t, err)
	})
}
