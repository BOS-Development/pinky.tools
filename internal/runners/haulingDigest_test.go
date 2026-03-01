package runners_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHaulingDigestUserRepo mocks HaulingDigestUserRepository.
type MockHaulingDigestUserRepo struct {
	mock.Mock
}

func (m *MockHaulingDigestUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

// MockHaulingDigestRunsRepo mocks HaulingDigestRunsRepository.
type MockHaulingDigestRunsRepo struct {
	mock.Mock
}

func (m *MockHaulingDigestRunsRepo) ListDigestRunsByUser(ctx context.Context, userID int64) ([]*models.HaulingRun, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.HaulingRun), args.Error(1)
}

// MockHaulingDigestNotifier mocks HaulingDigestNotifier.
type MockHaulingDigestNotifier struct {
	mock.Mock
}

func (m *MockHaulingDigestNotifier) SendHaulingDailyDigest(ctx context.Context, userID int64, runs []*models.HaulingRun) {
	m.Called(ctx, userID, runs)
}

func setupDigestRunner() (*runners.HaulingDigestRunner, *MockHaulingDigestNotifier, *MockHaulingDigestRunsRepo, *MockHaulingDigestUserRepo) {
	notifier := new(MockHaulingDigestNotifier)
	runsRepo := new(MockHaulingDigestRunsRepo)
	userRepo := new(MockHaulingDigestUserRepo)
	runner := runners.NewHaulingDigestRunner(notifier, runsRepo, userRepo, 24*time.Hour)
	return runner, notifier, runsRepo, userRepo
}

func Test_HaulingDigestRunner_RunsOnStartup(t *testing.T) {
	runner, notifier, runsRepo, userRepo := setupDigestRunner()
	mockTicker := NewMockTicker()

	runner = runner.WithTickerFactory(func(d time.Duration) runners.Ticker {
		return mockTicker
	})

	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{int64(1)}, nil).Once()
	runs := []*models.HaulingRun{{ID: int64(1), Name: "Run 1", Status: "PLANNING"}}
	runsRepo.On("ListDigestRunsByUser", mock.Anything, int64(1)).Return(runs, nil).Once()
	notifier.On("SendHaulingDailyDigest", mock.Anything, int64(1), runs).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	runsRepo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func Test_HaulingDigestRunner_SkipsUsersWithNoDigestRuns(t *testing.T) {
	runner, notifier, runsRepo, userRepo := setupDigestRunner()
	mockTicker := NewMockTicker()

	runner = runner.WithTickerFactory(func(d time.Duration) runners.Ticker {
		return mockTicker
	})

	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{int64(1), int64(2)}, nil).Once()
	runsRepo.On("ListDigestRunsByUser", mock.Anything, int64(1)).Return([]*models.HaulingRun{}, nil).Once()
	runsRepo.On("ListDigestRunsByUser", mock.Anything, int64(2)).Return([]*models.HaulingRun{
		{ID: int64(5), Name: "Active Run", Status: "ACCUMULATING"},
	}, nil).Once()
	notifier.On("SendHaulingDailyDigest", mock.Anything, int64(2), mock.Anything).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	runsRepo.AssertExpectations(t)
	notifier.AssertExpectations(t)
	// Notifier should not be called for user 1 (no runs)
	notifier.AssertNotCalled(t, "SendHaulingDailyDigest", mock.Anything, int64(1), mock.Anything)
}

func Test_HaulingDigestRunner_HandlesGetAllIDsError(t *testing.T) {
	runner, notifier, _, userRepo := setupDigestRunner()
	mockTicker := NewMockTicker()

	runner = runner.WithTickerFactory(func(d time.Duration) runners.Ticker {
		return mockTicker
	})

	userRepo.On("GetAllIDs", mock.Anything).Return(nil, errors.New("db error")).Once()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := runner.Run(ctx)
	assert.NoError(t, err)
	notifier.AssertNotCalled(t, "SendHaulingDailyDigest", mock.Anything, mock.Anything, mock.Anything)
}

func Test_HaulingDigestRunner_RunsPeriodically(t *testing.T) {
	runner, _, runsRepo, userRepo := setupDigestRunner()
	mockTicker := NewMockTicker()

	runner = runner.WithTickerFactory(func(d time.Duration) runners.Ticker {
		return mockTicker
	})

	// Startup + 1 tick = 2 rounds
	userRepo.On("GetAllIDs", mock.Anything).Return([]int64{int64(1)}, nil).Times(2)
	runsRepo.On("ListDigestRunsByUser", mock.Anything, int64(1)).Return([]*models.HaulingRun{}, nil).Times(2)

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
	userRepo.AssertExpectations(t)
}

func Test_HaulingDigestRunner_Constructor(t *testing.T) {
	notifier := new(MockHaulingDigestNotifier)
	runsRepo := new(MockHaulingDigestRunsRepo)
	userRepo := new(MockHaulingDigestUserRepo)
	runner := runners.NewHaulingDigestRunner(notifier, runsRepo, userRepo, 24*time.Hour)
	assert.NotNil(t, runner)
}
