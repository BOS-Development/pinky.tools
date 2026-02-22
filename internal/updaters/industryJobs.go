package updaters

import (
	"context"
	"strings"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"

	"github.com/pkg/errors"
)

// Activity ID to activity name mapping for queue matching
var activityIDToName = map[int]string{
	1: "manufacturing",
	3: "te_research",
	4: "me_research",
	5: "copying",
	8: "invention",
	9: "reaction",
}

type IndustryJobsRepository interface {
	UpsertJobs(ctx context.Context, userID int64, jobs []*models.IndustryJob) error
	GetActiveJobsForMatching(ctx context.Context, userID int64) ([]*models.IndustryJob, error)
	GetJobByID(ctx context.Context, jobID int64) (*models.IndustryJob, error)
}

type IndustryJobQueueRepository interface {
	GetPlannedJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error)
	GetLinkedActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error)
	LinkToEsiJob(ctx context.Context, queueID, esiJobID int64) error
	CompleteJob(ctx context.Context, queueID int64) error
}

type IndustryEsiClient interface {
	GetCharacterIndustryJobs(ctx context.Context, characterID int64, token string, includeCompleted bool) ([]*client.EsiIndustryJob, error)
	GetCorporationIndustryJobs(ctx context.Context, corporationID int64, token string, includeCompleted bool) ([]*client.EsiIndustryJob, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type IndustryUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

type IndustryCharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type IndustryCorporationRepository interface {
	Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type IndustryJobsUpdater struct {
	userRepo      IndustryUserRepository
	characterRepo IndustryCharacterRepository
	corpRepo      IndustryCorporationRepository
	jobsRepo      IndustryJobsRepository
	queueRepo     IndustryJobQueueRepository
	esiClient     IndustryEsiClient
}

func NewIndustryJobsUpdater(
	userRepo IndustryUserRepository,
	characterRepo IndustryCharacterRepository,
	corpRepo IndustryCorporationRepository,
	jobsRepo IndustryJobsRepository,
	queueRepo IndustryJobQueueRepository,
	esiClient IndustryEsiClient,
) *IndustryJobsUpdater {
	return &IndustryJobsUpdater{
		userRepo:      userRepo,
		characterRepo: characterRepo,
		corpRepo:      corpRepo,
		jobsRepo:      jobsRepo,
		queueRepo:     queueRepo,
		esiClient:     esiClient,
	}
}

func (u *IndustryJobsUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for industry jobs update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserJobs(ctx, userID); err != nil {
			log.Error("failed to update industry jobs for user", "userID", userID, "error", err)
		}
	}

	return nil
}

func (u *IndustryJobsUpdater) UpdateUserJobs(ctx context.Context, userID int64) error {
	characters, err := u.characterRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user characters for industry jobs update")
	}

	for _, char := range characters {
		if !strings.Contains(char.EsiScopes, "esi-industry.read_character_jobs.v1") {
			continue
		}

		if err := u.updateCharacterJobs(ctx, char, userID); err != nil {
			log.Error("failed to update industry jobs for character", "characterID", char.ID, "error", err)
		}
	}

	// Fetch corporation industry jobs
	corporations, err := u.corpRepo.Get(ctx, userID)
	if err != nil {
		log.Error("failed to get user corporations for industry jobs update", "userID", userID, "error", err)
	} else {
		for _, corp := range corporations {
			if !strings.Contains(corp.EsiScopes, "esi-industry.read_corporation_jobs.v1") {
				continue
			}

			if err := u.updateCorporationJobs(ctx, corp, userID); err != nil {
				log.Error("failed to update industry jobs for corporation", "corporationID", corp.ID, "error", err)
			}
		}
	}

	// After all characters and corporations are fetched, match queue entries
	if err := u.matchQueueEntries(ctx, userID); err != nil {
		log.Error("failed to match queue entries", "userID", userID, "error", err)
	}

	return nil
}

func (u *IndustryJobsUpdater) updateCharacterJobs(ctx context.Context, char *repositories.Character, userID int64) error {
	token := char.EsiToken

	if time.Now().After(char.EsiTokenExpiresOn) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, char.EsiRefreshToken)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken

		err = u.characterRepo.UpdateTokens(ctx, char.ID, char.UserID, refreshed.AccessToken, refreshed.RefreshToken, refreshed.Expiry)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for character %d", char.ID)
		}
		log.Info("refreshed ESI token for character (industry)", "characterID", char.ID)
	}

	esiJobs, err := u.esiClient.GetCharacterIndustryJobs(ctx, char.ID, token, true)
	if err != nil {
		return errors.Wrap(err, "failed to get character industry jobs from ESI")
	}

	jobs, err := convertEsiJobs(esiJobs, userID, "character")
	if err != nil {
		return err
	}

	err = u.jobsRepo.UpsertJobs(ctx, userID, jobs)
	if err != nil {
		return errors.Wrap(err, "failed to upsert industry jobs")
	}

	log.Info("updated industry jobs", "characterID", char.ID, "count", len(jobs))
	return nil
}

func (u *IndustryJobsUpdater) updateCorporationJobs(ctx context.Context, corp repositories.PlayerCorporation, userID int64) error {
	token := corp.EsiToken

	if time.Now().After(corp.EsiExpiresOn) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, corp.EsiRefreshToken)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for corporation %d", corp.ID)
		}
		token = refreshed.AccessToken

		err = u.corpRepo.UpdateTokens(ctx, corp.ID, corp.UserID, token, refreshed.RefreshToken, refreshed.Expiry)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for corporation %d", corp.ID)
		}
		log.Info("refreshed ESI token for corporation (industry)", "corporationID", corp.ID)
	}

	esiJobs, err := u.esiClient.GetCorporationIndustryJobs(ctx, corp.ID, token, true)
	if err != nil {
		return errors.Wrap(err, "failed to get corporation industry jobs from ESI")
	}

	jobs, err := convertEsiJobs(esiJobs, userID, "corporation")
	if err != nil {
		return err
	}

	err = u.jobsRepo.UpsertJobs(ctx, userID, jobs)
	if err != nil {
		return errors.Wrap(err, "failed to upsert corporation industry jobs")
	}

	log.Info("updated corporation industry jobs", "corporationID", corp.ID, "count", len(jobs))
	return nil
}

// convertEsiJobs converts ESI industry job responses into domain models.
// source should be "character" or "corporation".
func convertEsiJobs(esiJobs []*client.EsiIndustryJob, userID int64, source string) ([]*models.IndustryJob, error) {
	jobs := []*models.IndustryJob{}
	for _, ej := range esiJobs {
		// Corporation endpoint uses location_id instead of station_id
		stationID := ej.StationID
		if stationID == 0 && ej.LocationID != 0 {
			stationID = ej.LocationID
		}

		job := &models.IndustryJob{
			JobID:                ej.JobID,
			InstallerID:          ej.InstallerID,
			UserID:               userID,
			FacilityID:           ej.FacilityID,
			StationID:            stationID,
			ActivityID:           ej.ActivityID,
			BlueprintID:          ej.BlueprintID,
			BlueprintTypeID:      ej.BlueprintTypeID,
			BlueprintLocationID:  ej.BlueprintLocationID,
			OutputLocationID:     ej.OutputLocationID,
			Runs:                 ej.Runs,
			Cost:                 ej.Cost,
			LicensedRuns:         ej.LicensedRuns,
			Probability:          ej.Probability,
			ProductTypeID:        ej.ProductTypeID,
			Status:               ej.Status,
			Duration:             ej.Duration,
			CompletedCharacterID: ej.CompletedCharacterID,
			SuccessfulRuns:       ej.SuccessfulRuns,
			Source:               source,
		}

		startDate, err := time.Parse(time.RFC3339, ej.StartDate)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse start_date for job %d", ej.JobID)
		}
		job.StartDate = startDate

		endDate, err := time.Parse(time.RFC3339, ej.EndDate)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse end_date for job %d", ej.JobID)
		}
		job.EndDate = endDate

		if ej.PauseDate != nil {
			t, err := time.Parse(time.RFC3339, *ej.PauseDate)
			if err == nil {
				job.PauseDate = &t
			}
		}
		if ej.CompletedDate != nil {
			t, err := time.Parse(time.RFC3339, *ej.CompletedDate)
			if err == nil {
				job.CompletedDate = &t
			}
		}

		jobs = append(jobs, job)
	}
	return jobs, nil
}

// matchQueueEntries matches planned queue entries to active ESI jobs and
// checks if linked active entries have been delivered.
func (u *IndustryJobsUpdater) matchQueueEntries(ctx context.Context, userID int64) error {
	// 1. Match planned entries to active ESI jobs
	planned, err := u.queueRepo.GetPlannedJobs(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get planned queue entries")
	}

	if len(planned) > 0 {
		activeEsiJobs, err := u.jobsRepo.GetActiveJobsForMatching(ctx, userID)
		if err != nil {
			return errors.Wrap(err, "failed to get active ESI jobs for matching")
		}

		for _, entry := range planned {
			for _, esiJob := range activeEsiJobs {
				activityName := activityIDToName[esiJob.ActivityID]
				if esiJob.BlueprintTypeID == entry.BlueprintTypeID &&
					activityName == entry.Activity &&
					esiJob.Runs == entry.Runs &&
					esiJob.StartDate.After(entry.CreatedAt) {

					err := u.queueRepo.LinkToEsiJob(ctx, entry.ID, esiJob.JobID)
					if err != nil {
						log.Error("failed to link queue entry to ESI job", "queueID", entry.ID, "jobID", esiJob.JobID, "error", err)
					} else {
						log.Info("linked queue entry to ESI job", "queueID", entry.ID, "jobID", esiJob.JobID)
					}
					break
				}
			}
		}
	}

	// 2. Check if linked active entries have been delivered
	linkedActive, err := u.queueRepo.GetLinkedActiveJobs(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get linked active queue entries")
	}

	for _, entry := range linkedActive {
		if entry.EsiJobID == nil {
			continue
		}

		esiJob, err := u.jobsRepo.GetJobByID(ctx, *entry.EsiJobID)
		if err != nil {
			log.Error("failed to get ESI job for completion check", "jobID", *entry.EsiJobID, "error", err)
			continue
		}

		if esiJob != nil && esiJob.Status == "delivered" {
			err := u.queueRepo.CompleteJob(ctx, entry.ID)
			if err != nil {
				log.Error("failed to complete queue entry", "queueID", entry.ID, "error", err)
			} else {
				log.Info("completed queue entry", "queueID", entry.ID, "jobID", *entry.EsiJobID)
			}
		}
	}

	return nil
}
