package updaters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/mock"
)

// MockJobSlotNotificationRepo mocks the JobSlotNotificationRepo interface.
type MockJobSlotNotificationRepo struct {
	mock.Mock
}

func (m *MockJobSlotNotificationRepo) GetActiveAgreementCharacters(ctx context.Context) ([]*repositories.ActiveAgreementCharacter, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.ActiveAgreementCharacter), args.Error(1)
}

func (m *MockJobSlotNotificationRepo) HasJobBeenNotified(ctx context.Context, characterID, jobID int64) (bool, error) {
	args := m.Called(ctx, characterID, jobID)
	return args.Bool(0), args.Error(1)
}

func (m *MockJobSlotNotificationRepo) MarkJobNotified(ctx context.Context, characterID, jobID int64) error {
	args := m.Called(ctx, characterID, jobID)
	return args.Error(0)
}

// MockIndustryJobsForNotificationRepo mocks the IndustryJobsForNotificationRepo interface.
type MockIndustryJobsForNotificationRepo struct {
	mock.Mock
}

func (m *MockIndustryJobsForNotificationRepo) GetDeliveredJobsForCharacter(ctx context.Context, characterID int64) ([]*models.IndustryJob, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.IndustryJob), args.Error(1)
}

// MockJobSlotJobCompletedNotifier mocks the JobSlotJobCompletedNotifier interface.
type MockJobSlotJobCompletedNotifier struct {
	mock.Mock
}

func (m *MockJobSlotJobCompletedNotifier) NotifyJobSlotJobCompleted(ctx context.Context, renterUserID int64, characterName string, activityType string, productName string, runs int, endDate time.Time) {
	m.Called(ctx, renterUserID, characterName, activityType, productName, runs, endDate)
}

func Test_JobSlotNotifications_NoActiveAgreements(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return([]*repositories.ActiveAgreementCharacter{}, nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertNotCalled(t, "GetDeliveredJobsForCharacter")
	notifier.AssertNotCalled(t, "NotifyJobSlotJobCompleted")
}

func Test_JobSlotNotifications_GetActiveAgreementsError(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(nil, errors.New("db error"))

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertNotCalled(t, "GetDeliveredJobsForCharacter")
	notifier.AssertNotCalled(t, "NotifyJobSlotJobCompleted")
}

func Test_JobSlotNotifications_NoDeliveredJobs(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	chars := []*repositories.ActiveAgreementCharacter{
		{CharacterID: 1001, CharacterName: "Alpha", RenterUserID: 500, ActivityType: "manufacturing"},
	}
	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(chars, nil)
	jobsRepo.On("GetDeliveredJobsForCharacter", mock.Anything, int64(1001)).Return([]*models.IndustryJob{}, nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertExpectations(t)
	notifier.AssertNotCalled(t, "NotifyJobSlotJobCompleted")
}

func Test_JobSlotNotifications_NotifiesUnnotifiedJob(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	chars := []*repositories.ActiveAgreementCharacter{
		{CharacterID: 1002, CharacterName: "Beta", RenterUserID: 501, ActivityType: "manufacturing"},
	}

	now := time.Now().UTC()
	job := &models.IndustryJob{
		JobID:       200001,
		InstallerID: 1002,
		ActivityID:  1,
		Runs:        10,
		ProductName: "Tritanium",
		EndDate:     now,
		Status:      "delivered",
	}

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(chars, nil)
	jobsRepo.On("GetDeliveredJobsForCharacter", mock.Anything, int64(1002)).Return([]*models.IndustryJob{job}, nil)
	agreementRepo.On("HasJobBeenNotified", mock.Anything, int64(1002), int64(200001)).Return(false, nil)
	notifier.On("NotifyJobSlotJobCompleted", mock.Anything, int64(501), "Beta", "manufacturing", "Tritanium", 10, mock.AnythingOfType("time.Time")).Return()
	agreementRepo.On("MarkJobNotified", mock.Anything, int64(1002), int64(200001)).Return(nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func Test_JobSlotNotifications_SkipsAlreadyNotifiedJob(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	chars := []*repositories.ActiveAgreementCharacter{
		{CharacterID: 1003, CharacterName: "Gamma", RenterUserID: 502, ActivityType: "reaction"},
	}

	now := time.Now().UTC()
	job := &models.IndustryJob{
		JobID:       200002,
		InstallerID: 1003,
		ActivityID:  9,
		Runs:        100,
		ProductName: "Helium Isotopes",
		EndDate:     now,
		Status:      "delivered",
	}

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(chars, nil)
	jobsRepo.On("GetDeliveredJobsForCharacter", mock.Anything, int64(1003)).Return([]*models.IndustryJob{job}, nil)
	agreementRepo.On("HasJobBeenNotified", mock.Anything, int64(1003), int64(200002)).Return(true, nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertExpectations(t)
	notifier.AssertNotCalled(t, "NotifyJobSlotJobCompleted")
	agreementRepo.AssertNotCalled(t, "MarkJobNotified")
}

func Test_JobSlotNotifications_UsesBlueprintNameWhenNoProductName(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	chars := []*repositories.ActiveAgreementCharacter{
		{CharacterID: 1004, CharacterName: "Delta", RenterUserID: 503, ActivityType: "copying"},
	}

	now := time.Now().UTC()
	job := &models.IndustryJob{
		JobID:         200003,
		InstallerID:   1004,
		ActivityID:    5,
		Runs:          1,
		BlueprintName: "Raven Blueprint",
		ProductName:   "", // no product name for copying jobs
		EndDate:       now,
		Status:        "delivered",
	}

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(chars, nil)
	jobsRepo.On("GetDeliveredJobsForCharacter", mock.Anything, int64(1004)).Return([]*models.IndustryJob{job}, nil)
	agreementRepo.On("HasJobBeenNotified", mock.Anything, int64(1004), int64(200003)).Return(false, nil)
	notifier.On("NotifyJobSlotJobCompleted", mock.Anything, int64(503), "Delta", "copying", "Raven Blueprint", 1, mock.AnythingOfType("time.Time")).Return()
	agreementRepo.On("MarkJobNotified", mock.Anything, int64(1004), int64(200003)).Return(nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func Test_JobSlotNotifications_UsesCompletedDateWhenAvailable(t *testing.T) {
	agreementRepo := new(MockJobSlotNotificationRepo)
	jobsRepo := new(MockIndustryJobsForNotificationRepo)
	notifier := new(MockJobSlotJobCompletedNotifier)

	chars := []*repositories.ActiveAgreementCharacter{
		{CharacterID: 1005, CharacterName: "Epsilon", RenterUserID: 504, ActivityType: "manufacturing"},
	}

	endDate := time.Now().UTC().Add(-2 * time.Hour)
	completedDate := time.Now().UTC().Add(-time.Hour)
	job := &models.IndustryJob{
		JobID:         200004,
		InstallerID:   1005,
		ActivityID:    1,
		Runs:          5,
		ProductName:   "Some Module",
		EndDate:       endDate,
		CompletedDate: &completedDate,
		Status:        "delivered",
	}

	agreementRepo.On("GetActiveAgreementCharacters", mock.Anything).Return(chars, nil)
	jobsRepo.On("GetDeliveredJobsForCharacter", mock.Anything, int64(1005)).Return([]*models.IndustryJob{job}, nil)
	agreementRepo.On("HasJobBeenNotified", mock.Anything, int64(1005), int64(200004)).Return(false, nil)
	// The notifier should receive completedDate, not endDate
	notifier.On("NotifyJobSlotJobCompleted", mock.Anything, int64(504), "Epsilon", "manufacturing", "Some Module", 5, completedDate).Return()
	agreementRepo.On("MarkJobNotified", mock.Anything, int64(1005), int64(200004)).Return(nil)

	updater := updaters.NewJobSlotNotificationsUpdater(agreementRepo, jobsRepo, notifier)
	updater.CheckAndNotifyCompletedJobs(context.Background())

	agreementRepo.AssertExpectations(t)
	jobsRepo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}
