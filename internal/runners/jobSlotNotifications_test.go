package runners_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobSlotNotificationsUpdater mocks the JobSlotNotificationsUpdaterInterface.
type MockJobSlotNotificationsUpdater struct {
	mock.Mock
}

func (m *MockJobSlotNotificationsUpdater) CheckAndNotifyCompletedJobs(ctx context.Context) {
	m.Called(ctx)
}

func Test_JobSlotNotificationsRunner_DoesNotRunOnStartup(t *testing.T) {
	mockUpdater := new(MockJobSlotNotificationsUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewJobSlotNotificationsRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)

	assert.NoError(t, err)
	mockUpdater.AssertExpectations(t) // zero calls expected
}

func Test_JobSlotNotificationsRunner_RunsOnScheduledTick(t *testing.T) {
	mockUpdater := new(MockJobSlotNotificationsUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewJobSlotNotificationsRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("CheckAndNotifyCompletedJobs", mock.Anything).Return().Once()

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

func Test_JobSlotNotificationsRunner_RunsMultipleTicks(t *testing.T) {
	mockUpdater := new(MockJobSlotNotificationsUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewJobSlotNotificationsRunner(mockUpdater, 15*time.Minute).
		WithTickerFactory(func(d time.Duration) runners.Ticker {
			return mockTicker
		})

	mockUpdater.On("CheckAndNotifyCompletedJobs", mock.Anything).Return().Times(2)

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

func Test_JobSlotNotificationsRunner_StopsOnContextCancellation(t *testing.T) {
	mockUpdater := new(MockJobSlotNotificationsUpdater)
	mockTicker := NewMockTicker()

	runner := runners.NewJobSlotNotificationsRunner(mockUpdater, 15*time.Minute).
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

func Test_JobSlotNotificationsRunner_Constructor(t *testing.T) {
	mockUpdater := new(MockJobSlotNotificationsUpdater)
	interval := 15 * time.Minute

	runner := runners.NewJobSlotNotificationsRunner(mockUpdater, interval)

	assert.NotNil(t, runner)
}
