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

// MockAutoProductionUpdater mocks the AutoProductionUpdaterInterface
type MockAutoProductionUpdater struct {
	mock.Mock
}

func (m *MockAutoProductionUpdater) RunAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func Test_AutoProductionRunner_DoesNotRunOnStartup(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewAutoProductionRunner(mockUpdater, 30*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// No calls expected — runner does not run on startup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_AutoProductionRunner_RunsOnScheduledTick(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewAutoProductionRunner(mockUpdater, 30*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// One scheduled call expected
	mockUpdater.On("RunAll", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	cancel()
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_AutoProductionRunner_RunsMultipleTicks(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewAutoProductionRunner(mockUpdater, 30*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Two scheduled calls expected
	mockUpdater.On("RunAll", mock.Anything).Return(nil).Times(2)

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

func Test_AutoProductionRunner_ContinuesOnError(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewAutoProductionRunner(mockUpdater, 30*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// First tick errors, second tick succeeds
	mockUpdater.On("RunAll", mock.Anything).Return(errors.New("transient error")).Once()
	mockUpdater.On("RunAll", mock.Anything).Return(nil).Once()

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

func Test_AutoProductionRunner_StopsOnContextCancellation(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewAutoProductionRunner(mockUpdater, 30*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t) // zero calls expected
}

func Test_AutoProductionRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockAutoProductionUpdater)
	interval := 30 * time.Minute

	runner := runners.NewAutoProductionRunner(mockUpdater, interval)

	assert.NotNil(t, runner)
}
