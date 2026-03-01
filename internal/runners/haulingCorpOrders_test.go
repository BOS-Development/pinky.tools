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

// MockHaulingCorpOrdersUpdater mocks the HaulingCorpOrdersUpdaterInterface.
type MockHaulingCorpOrdersUpdater struct {
	mock.Mock
}

func (m *MockHaulingCorpOrdersUpdater) UpdateAllUsers(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func Test_HaulingCorpOrdersRunner_RunsOnStartup(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_HaulingCorpOrdersRunner_RunsOnStartupError(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(errors.New("startup error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_HaulingCorpOrdersRunner_RunsPeriodically(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Startup + 2 scheduled ticks
	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(nil).Times(3)

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

func Test_HaulingCorpOrdersRunner_ContinuesOnError(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(nil).Once()
	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(errors.New("update error")).Times(2)

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

func Test_HaulingCorpOrdersRunner_StopsOnContextCancellation(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("UpdateAllUsers", mock.Anything).Return(nil).Once()

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

func Test_HaulingCorpOrdersRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockHaulingCorpOrdersUpdater)
	runner := runners.NewHaulingCorpOrdersRunner(mockUpdater, 15*time.Minute)
	assert.NotNil(t, runner)
}
