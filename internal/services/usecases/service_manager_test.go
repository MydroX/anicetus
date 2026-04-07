package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"MydroX/anicetus/internal/services/mocks"
	"MydroX/anicetus/pkg/errs"
	"MydroX/anicetus/pkg/logger"

	valkeymock "github.com/valkey-io/valkey-go/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupManagerTest(t *testing.T) (*mocks.MockServiceStore, *valkeymock.Client, *ServiceManager) {
	ctrl := gomock.NewController(t)
	log, _ := logger.New("TEST")

	store := mocks.NewMockServiceStore(ctrl)
	vClient := valkeymock.NewClient(ctrl)

	manager := NewServiceManager(log, store, vClient)
	return store, vClient, manager
}

func TestCacheAllowedServices_Success(t *testing.T) {
	store, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	store.EXPECT().GetAllowedServices(gomock.Any()).Return([]string{"svc-a", "svc-b"}, nil)
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).Return(valkeymock.Result(valkeymock.ValkeyBlobString("OK")))

	err := manager.CacheAllowedServices(ctx)
	assert.NoError(t, err)
}

func TestCacheAllowedServices_RepoError(t *testing.T) {
	store, _, manager := setupManagerTest(t)
	ctx := context.Background()

	store.EXPECT().GetAllowedServices(gomock.Any()).
		Return(nil, errs.New(errs.ErrorInternal, "db err", nil))

	err := manager.CacheAllowedServices(ctx)
	assert.Error(t, err)
}

func TestGetAllowedServices_CacheHit(t *testing.T) {
	_, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	cached, _ := json.Marshal([]string{"svc-cached"})
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.Result(valkeymock.ValkeyBlobString(string(cached))))

	services, err := manager.GetAllowedServices(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"svc-cached"}, services)
}

func TestGetAllowedServices_CacheMiss(t *testing.T) {
	store, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	// First call: GET cache miss
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.ErrorResult(errors.New("cache miss")))

	// Repo fallback
	store.EXPECT().GetAllowedServices(gomock.Any()).Return([]string{"svc-from-db"}, nil)

	// Background cache update: GetAllowedServices + SET
	store.EXPECT().GetAllowedServices(gomock.Any()).Return([]string{"svc-from-db"}, nil).AnyTimes()
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).Return(valkeymock.Result(valkeymock.ValkeyBlobString("OK"))).AnyTimes()

	services, err := manager.GetAllowedServices(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"svc-from-db"}, services)
}

func TestGetAllowedServices_RepoError(t *testing.T) {
	store, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.ErrorResult(errors.New("cache miss")))
	store.EXPECT().GetAllowedServices(gomock.Any()).
		Return(nil, errs.New(errs.ErrorInternal, "db err", nil))

	_, err := manager.GetAllowedServices(ctx)
	assert.Error(t, err)
}

func TestGetUserServices_CacheHit(t *testing.T) {
	_, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	cached, _ := json.Marshal([]string{"user-svc"})
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.Result(valkeymock.ValkeyBlobString(string(cached))))

	services, err := manager.GetUserServices(ctx, "user-uuid")
	require.NoError(t, err)
	assert.Equal(t, []string{"user-svc"}, services)
}

func TestGetUserServices_CacheMiss(t *testing.T) {
	store, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	// Cache miss
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.ErrorResult(errors.New("cache miss")))
	// Repo fallback
	store.EXPECT().GetUserServices(gomock.Any(), "user-uuid").Return([]string{"user-svc"}, nil)
	// Cache SET
	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.Result(valkeymock.ValkeyBlobString("OK")))

	services, err := manager.GetUserServices(ctx, "user-uuid")
	require.NoError(t, err)
	assert.Equal(t, []string{"user-svc"}, services)
}

func TestGetUserServices_RepoError(t *testing.T) {
	store, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.ErrorResult(errors.New("cache miss")))
	store.EXPECT().GetUserServices(gomock.Any(), "user-uuid").
		Return(nil, errs.New(errs.ErrorInternal, "err", nil))

	_, err := manager.GetUserServices(ctx, "user-uuid")
	assert.Error(t, err)
}

func TestInvalidateUserServicesCache(t *testing.T) {
	_, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.Result(valkeymock.ValkeyInt64(1)))

	manager.InvalidateUserServicesCache(ctx, "user-uuid")
}

func TestInvalidateAllServicesCache(t *testing.T) {
	_, vClient, manager := setupManagerTest(t)
	ctx := context.Background()

	vClient.EXPECT().Do(gomock.Any(), gomock.Any()).
		Return(valkeymock.Result(valkeymock.ValkeyInt64(1)))

	manager.InvalidateAllServicesCache(ctx)
}
