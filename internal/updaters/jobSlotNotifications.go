package updaters

import (
	"context"
	"time"

	"github.com/annymsMthd/industry-tool/internal/models"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/repositories"
)

// JobSlotNotificationRepo is the interface for querying agreement/notification state.
type JobSlotNotificationRepo interface {
	GetActiveAgreementCharacters(ctx context.Context) ([]*repositories.ActiveAgreementCharacter, error)
	HasJobBeenNotified(ctx context.Context, characterID, jobID int64) (bool, error)
	MarkJobNotified(ctx context.Context, characterID, jobID int64) error
}

// IndustryJobsForNotificationRepo is the interface for querying delivered jobs.
type IndustryJobsForNotificationRepo interface {
	GetDeliveredJobsForCharacter(ctx context.Context, characterID int64) ([]*models.IndustryJob, error)
}

// JobSlotNotificationsUpdater checks for completed jobs on rental characters and notifies renters.
type JobSlotNotificationsUpdater struct {
	agreementRepo JobSlotNotificationRepo
	jobsRepo      IndustryJobsForNotificationRepo
	notifier      JobSlotJobCompletedNotifier
}

// NewJobSlotNotificationsUpdater creates a new JobSlotNotificationsUpdater.
func NewJobSlotNotificationsUpdater(
	agreementRepo JobSlotNotificationRepo,
	jobsRepo IndustryJobsForNotificationRepo,
	notifier JobSlotJobCompletedNotifier,
) *JobSlotNotificationsUpdater {
	return &JobSlotNotificationsUpdater{
		agreementRepo: agreementRepo,
		jobsRepo:      jobsRepo,
		notifier:      notifier,
	}
}

// CheckAndNotifyCompletedJobs scans active rental agreements, finds newly delivered jobs,
// and notifies the renter via Discord.
func (u *JobSlotNotificationsUpdater) CheckAndNotifyCompletedJobs(ctx context.Context) {
	chars, err := u.agreementRepo.GetActiveAgreementCharacters(ctx)
	if err != nil {
		log.Error("job slot notifications: failed to get active agreement characters", "error", err)
		return
	}

	if len(chars) == 0 {
		return
	}

	for _, aac := range chars {
		jobs, err := u.jobsRepo.GetDeliveredJobsForCharacter(ctx, aac.CharacterID)
		if err != nil {
			log.Error("job slot notifications: failed to get delivered jobs", "character_id", aac.CharacterID, "error", err)
			continue
		}

		for _, job := range jobs {
			already, err := u.agreementRepo.HasJobBeenNotified(ctx, aac.CharacterID, job.JobID)
			if err != nil {
				log.Error("job slot notifications: failed to check job notification", "character_id", aac.CharacterID, "job_id", job.JobID, "error", err)
				continue
			}
			if already {
				continue
			}

			productName := job.ProductName
			if productName == "" {
				productName = job.BlueprintName
			}

			var endDate time.Time
			if job.CompletedDate != nil {
				endDate = *job.CompletedDate
			} else {
				endDate = job.EndDate
			}

			u.notifier.NotifyJobSlotJobCompleted(
				ctx,
				aac.RenterUserID,
				aac.CharacterName,
				aac.ActivityType,
				productName,
				job.Runs,
				endDate,
			)

			if err := u.agreementRepo.MarkJobNotified(ctx, aac.CharacterID, job.JobID); err != nil {
				log.Error("job slot notifications: failed to mark job notified", "character_id", aac.CharacterID, "job_id", job.JobID, "error", err)
			}
		}
	}
}
