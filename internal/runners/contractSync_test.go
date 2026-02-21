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

// MockContractSyncUpdater mocks the ContractSyncUpdater interface
type MockContractSyncUpdater struct {
	mock.Mock
}

func (m *MockContractSyncUpdater) SyncAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func Test_ContractSyncRunner_SyncsOnStartup(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewContractSyncRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("SyncAll", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_ContractSyncRunner_SyncsOnStartupError(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewContractSyncRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("SyncAll", mock.Anything).Return(errors.New("startup error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_ContractSyncRunner_SyncsPeriodically(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewContractSyncRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Expect 3 calls: 1 on startup + 2 scheduled
	mockUpdater.On("SyncAll", mock.Anything).Return(nil).Times(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)

	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	cancel()
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_ContractSyncRunner_ContinuesOnScheduledError(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewContractSyncRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("SyncAll", mock.Anything).Return(nil).Once()
	mockUpdater.On("SyncAll", mock.Anything).Return(errors.New("sync error")).Times(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)

	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	cancel()
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_ContractSyncRunner_StopsOnContextCancellation(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewContractSyncRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("SyncAll", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_ContractSyncRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockContractSyncUpdater)
	interval := 15 * time.Minute

	runner := runners.NewContractSyncRunner(mockUpdater, interval)

	assert.NotNil(t, runner)
}
