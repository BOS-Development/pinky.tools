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

// MockMarketPricesUpdater mocks the MarketPricesUpdater interface
type MockMarketPricesUpdater struct {
	mock.Mock
}

func (m *MockMarketPricesUpdater) UpdateJitaMarket(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockTicker allows controlling when ticks occur for testing
type MockTicker struct {
	ch     chan time.Time
	stopped bool
}

func NewMockTicker() *MockTicker {
	return &MockTicker{
		ch: make(chan time.Time, 10),
	}
}

func (t *MockTicker) C() <-chan time.Time {
	return t.ch
}

func (t *MockTicker) Stop() {
	t.stopped = true
	close(t.ch)
}

func (t *MockTicker) Tick() {
	if !t.stopped {
		t.ch <- time.Now()
	}
}

func Test_MarketPricesRunner_UpdatesOnStartup(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Once()

	// Cancel context immediately after startup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_UpdatesOnStartupError(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Startup update fails but runner should continue
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(errors.New("startup error")).Once()

	// Cancel context immediately after startup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	// Runner should not return error even if startup update fails
	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_UpdatesPeriodically(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Expect 3 calls: 1 on startup + 2 scheduled
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Times(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Wait briefly for startup call
	time.Sleep(10 * time.Millisecond)

	// Trigger 2 ticks manually
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	// Cancel and wait
	cancel()
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_ContinuesOnScheduledError(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// First call succeeds, subsequent calls fail
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Once()
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(errors.New("update error")).Times(2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Wait for startup
	time.Sleep(10 * time.Millisecond)

	// Trigger 2 ticks that will fail
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)
	mockTicker.Tick()
	time.Sleep(10 * time.Millisecond)

	// Cancel and wait
	cancel()
	err := <-done

	// Runner should not return error even if updates fail
	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_StopsOnContextCancellation(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Only startup call should happen
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Once()

	ctx, cancel := context.WithCancel(context.Background())

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Give it time to start and call startup update
	time.Sleep(10 * time.Millisecond)
	cancel()

	// Wait for runner to finish
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_RespectsCancelledContextOnStartup(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Startup update should still be called
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Once()

	// Create already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_MultipleTickIntervals(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewMarketPricesRunner(mockUpdater, 1*time.Hour).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	// Expect startup + 5 scheduled updates
	mockUpdater.On("UpdateJitaMarket", mock.Anything).Return(nil).Times(6)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- runner.Run(ctx)
	}()

	// Wait for startup
	time.Sleep(10 * time.Millisecond)

	// Trigger 5 ticks manually
	for i := 0; i < 5; i++ {
		mockTicker.Tick()
		time.Sleep(5 * time.Millisecond)
	}

	// Cancel and wait
	cancel()
	err := <-done

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t)
}

func Test_MarketPricesRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockMarketPricesUpdater)
	interval := 5 * time.Minute

	runner := runners.NewMarketPricesRunner(mockUpdater, interval)

	assert.NotNil(t, runner)
}
