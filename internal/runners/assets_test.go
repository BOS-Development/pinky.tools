package runners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAssetsUpdater mocks the AssetsUpdater interface
type MockAssetsUpdater struct {
	mock.Mock
}

func (m *MockAssetsUpdater) UpdateUserAssets(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockUserIDsProvider mocks the UserIDsProvider interface
type MockUserIDsProvider struct {
	mock.Mock
}

func (m *MockUserIDsProvider) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func Test_AssetsRunner_UpdatesOnStartup(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	mockTicker := NewMockTicker()

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockProvider.On("GetAllIDs", mock.Anything).Return([]int64{1, 2}, nil).Once()
	mockUpdater.On("UpdateUserAssets", mock.Anything, int64(1)).Return(nil).Once()
	mockUpdater.On("UpdateUserAssets", mock.Anything, int64(2)).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
	mockUpdater.AssertExpectations(t)
}

func Test_AssetsRunner_ContinuesOnUserError(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	mockTicker := NewMockTicker()

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockProvider.On("GetAllIDs", mock.Anything).Return([]int64{1, 2}, nil).Once()
	// First user fails, second succeeds — runner should continue
	mockUpdater.On("UpdateUserAssets", mock.Anything, int64(1)).Return(errors.New("user 1 error")).Once()
	mockUpdater.On("UpdateUserAssets", mock.Anything, int64(2)).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
	mockUpdater.AssertExpectations(t)
}

func Test_AssetsRunner_HandlesGetAllIDsError(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	mockTicker := NewMockTicker()

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockProvider.On("GetAllIDs", mock.Anything).Return(nil, errors.New("database error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
	// UpdateUserAssets should not be called when GetAllIDs fails
	mockUpdater.AssertNotCalled(t, "UpdateUserAssets", mock.Anything, mock.Anything)
}

func Test_AssetsRunner_UpdatesPeriodically(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	mockTicker := NewMockTicker()

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Expect 2 rounds: startup + 1 scheduled tick, each with 1 user
	mockProvider.On("GetAllIDs", mock.Anything).Return([]int64{1}, nil).Times(2)
	mockUpdater.On("UpdateUserAssets", mock.Anything, int64(1)).Return(nil).Times(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Wait for startup call
	time.Sleep(10 * time.Millisecond)

	// Trigger 1 scheduled tick
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	cancel()
	err := <-done

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
	mockUpdater.AssertExpectations(t)
}

func Test_AssetsRunner_NoUsersSkipsUpdate(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	mockTicker := NewMockTicker()

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockProvider.On("GetAllIDs", mock.Anything).Return([]int64{}, nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockProvider.AssertExpectations(t)
	mockUpdater.AssertNotCalled(t, "UpdateUserAssets", mock.Anything, mock.Anything)
}

func Test_AssetsRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockAssetsUpdater)
	mockProvider := new(MockUserIDsProvider)
	interval := 5 * time.Minute

	runner := runners.NewAssetsRunner(mockUpdater, mockProvider, interval)

	assert.NotNil(t, runner)
}
