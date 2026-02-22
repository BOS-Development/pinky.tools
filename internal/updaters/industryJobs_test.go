package updaters_test

import (
	"context"
	"testing"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/stretchr/testify/assert"
)

// --- Mocks ---

type mockIndustryJobsRepo struct {
	upsertedJobs map[int64][]*models.IndustryJob // keyed by userID
	upsertErr    error
	activeJobs   []*models.IndustryJob
	jobByID      map[int64]*models.IndustryJob
}

func (m *mockIndustryJobsRepo) UpsertJobs(ctx context.Context, userID int64, jobs []*models.IndustryJob) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	if m.upsertedJobs == nil {
		m.upsertedJobs = map[int64][]*models.IndustryJob{}
	}
	m.upsertedJobs[userID] = jobs
	return nil
}

func (m *mockIndustryJobsRepo) GetActiveJobsForMatching(ctx context.Context, userID int64) ([]*models.IndustryJob, error) {
	return m.activeJobs, nil
}

func (m *mockIndustryJobsRepo) GetJobByID(ctx context.Context, jobID int64) (*models.IndustryJob, error) {
	if m.jobByID != nil {
		return m.jobByID[jobID], nil
	}
	return nil, nil
}

type mockJobQueueRepo struct {
	plannedJobs    []*models.IndustryJobQueueEntry
	linkedJobs     []*models.IndustryJobQueueEntry
	linkedCalls    []linkCall
	completedCalls []int64
}

type linkCall struct {
	QueueID  int64
	EsiJobID int64
}

func (m *mockJobQueueRepo) GetPlannedJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	return m.plannedJobs, nil
}

func (m *mockJobQueueRepo) GetLinkedActiveJobs(ctx context.Context, userID int64) ([]*models.IndustryJobQueueEntry, error) {
	return m.linkedJobs, nil
}

func (m *mockJobQueueRepo) LinkToEsiJob(ctx context.Context, queueID, esiJobID int64) error {
	m.linkedCalls = append(m.linkedCalls, linkCall{QueueID: queueID, EsiJobID: esiJobID})
	return nil
}

func (m *mockJobQueueRepo) CompleteJob(ctx context.Context, queueID int64) error {
	m.completedCalls = append(m.completedCalls, queueID)
	return nil
}

type mockIndustryEsiClient struct {
	jobsByChar     map[int64][]*client.EsiIndustryJob
	jobsByCorp     map[int64][]*client.EsiIndustryJob
	jobsErr        error
	refreshedToken *client.RefreshedToken
	refreshErr     error
}

func (m *mockIndustryEsiClient) GetCharacterIndustryJobs(ctx context.Context, characterID int64, token string, includeCompleted bool) ([]*client.EsiIndustryJob, error) {
	if m.jobsErr != nil {
		return nil, m.jobsErr
	}
	return m.jobsByChar[characterID], nil
}

func (m *mockIndustryEsiClient) GetCorporationIndustryJobs(ctx context.Context, corporationID int64, token string, includeCompleted bool) ([]*client.EsiIndustryJob, error) {
	if m.jobsErr != nil {
		return nil, m.jobsErr
	}
	if m.jobsByCorp != nil {
		return m.jobsByCorp[corporationID], nil
	}
	return nil, nil
}

func (m *mockIndustryEsiClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.refreshedToken, nil
}

type mockIndustryUserRepo struct {
	userIDs []int64
}

func (m *mockIndustryUserRepo) GetAllIDs(ctx context.Context) ([]int64, error) {
	return m.userIDs, nil
}

type mockIndustryCharRepo struct {
	charactersByUser map[int64][]*repositories.Character
}

func (m *mockIndustryCharRepo) GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error) {
	return m.charactersByUser[baseUserID], nil
}

func (m *mockIndustryCharRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return nil
}

type mockIndustryCorpRepo struct {
	corpsByUser map[int64][]repositories.PlayerCorporation
}

func (m *mockIndustryCorpRepo) Get(ctx context.Context, user int64) ([]repositories.PlayerCorporation, error) {
	if m.corpsByUser != nil {
		return m.corpsByUser[user], nil
	}
	return nil, nil
}

func (m *mockIndustryCorpRepo) UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error {
	return nil
}

// --- Tests ---

func Test_IndustryJobsUpdater_FetchesAndUpsertsJobs(t *testing.T) {
	now := time.Now().UTC()
	cost := 1500000.0
	jobsRepo := &mockIndustryJobsRepo{}
	queueRepo := &mockJobQueueRepo{}
	esiClient := &mockIndustryEsiClient{
		jobsByChar: map[int64][]*client.EsiIndustryJob{
			1001: {
				{
					JobID: 100001, InstallerID: 1001, FacilityID: 60003760, StationID: 60003760,
					ActivityID: 1, BlueprintID: 9876, BlueprintTypeID: 787,
					BlueprintLocationID: 60003760, OutputLocationID: 60003760,
					Runs: 10, Cost: &cost, Status: "active", Duration: 3600,
					StartDate: now.Format(time.RFC3339), EndDate: now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
	}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "Industry Char", UserID: 100, EsiToken: "token", EsiScopes: "esi-industry.read_character_jobs.v1", EsiTokenExpiresOn: now.Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	assert.Len(t, jobsRepo.upsertedJobs[100], 1)
	assert.Equal(t, int64(100001), jobsRepo.upsertedJobs[100][0].JobID)
	assert.Equal(t, "active", jobsRepo.upsertedJobs[100][0].Status)
	assert.Equal(t, "character", jobsRepo.upsertedJobs[100][0].Source)
	assert.Equal(t, int64(60003760), jobsRepo.upsertedJobs[100][0].StationID)
}

func Test_IndustryJobsUpdater_MatchesQueueEntries(t *testing.T) {
	now := time.Now().UTC()
	createdAt := now.Add(-time.Hour)

	jobsRepo := &mockIndustryJobsRepo{
		activeJobs: []*models.IndustryJob{
			{
				JobID: 100001, BlueprintTypeID: 787, ActivityID: 1, Runs: 10,
				StartDate: now, Status: "active",
			},
		},
	}
	queueRepo := &mockJobQueueRepo{
		plannedJobs: []*models.IndustryJobQueueEntry{
			{
				ID: 1, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 10,
				CreatedAt: createdAt, Status: "planned",
			},
		},
	}
	esiClient := &mockIndustryEsiClient{
		jobsByChar: map[int64][]*client.EsiIndustryJob{},
	}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {},
		},
	}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	assert.Len(t, queueRepo.linkedCalls, 1)
	assert.Equal(t, int64(1), queueRepo.linkedCalls[0].QueueID)
	assert.Equal(t, int64(100001), queueRepo.linkedCalls[0].EsiJobID)
}

func Test_IndustryJobsUpdater_DoesNotMatchWrongActivity(t *testing.T) {
	now := time.Now().UTC()
	createdAt := now.Add(-time.Hour)

	jobsRepo := &mockIndustryJobsRepo{
		activeJobs: []*models.IndustryJob{
			{
				JobID: 100001, BlueprintTypeID: 787, ActivityID: 9, Runs: 10,
				StartDate: now, Status: "active",
			},
		},
	}
	queueRepo := &mockJobQueueRepo{
		plannedJobs: []*models.IndustryJobQueueEntry{
			{
				ID: 1, BlueprintTypeID: 787, Activity: "manufacturing", Runs: 10,
				CreatedAt: createdAt, Status: "planned",
			},
		},
	}
	esiClient := &mockIndustryEsiClient{jobsByChar: map[int64][]*client.EsiIndustryJob{}}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{charactersByUser: map[int64][]*repositories.Character{100: {}}}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	assert.Len(t, queueRepo.linkedCalls, 0)
}

func Test_IndustryJobsUpdater_CompletesDeliveredJobs(t *testing.T) {
	esiJobID := int64(100001)

	jobsRepo := &mockIndustryJobsRepo{
		activeJobs: []*models.IndustryJob{},
		jobByID: map[int64]*models.IndustryJob{
			100001: {JobID: 100001, Status: "delivered"},
		},
	}
	queueRepo := &mockJobQueueRepo{
		plannedJobs: []*models.IndustryJobQueueEntry{},
		linkedJobs: []*models.IndustryJobQueueEntry{
			{
				ID: 5, EsiJobID: &esiJobID, Status: "active",
			},
		},
	}
	esiClient := &mockIndustryEsiClient{jobsByChar: map[int64][]*client.EsiIndustryJob{}}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{charactersByUser: map[int64][]*repositories.Character{100: {}}}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	assert.Len(t, queueRepo.completedCalls, 1)
	assert.Equal(t, int64(5), queueRepo.completedCalls[0])
}

func Test_IndustryJobsUpdater_DoesNotCompleteActiveJobs(t *testing.T) {
	esiJobID := int64(100001)

	jobsRepo := &mockIndustryJobsRepo{
		activeJobs: []*models.IndustryJob{},
		jobByID: map[int64]*models.IndustryJob{
			100001: {JobID: 100001, Status: "active"},
		},
	}
	queueRepo := &mockJobQueueRepo{
		plannedJobs: []*models.IndustryJobQueueEntry{},
		linkedJobs: []*models.IndustryJobQueueEntry{
			{
				ID: 5, EsiJobID: &esiJobID, Status: "active",
			},
		},
	}
	esiClient := &mockIndustryEsiClient{jobsByChar: map[int64][]*client.EsiIndustryJob{}}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{charactersByUser: map[int64][]*repositories.Character{100: {}}}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	assert.Len(t, queueRepo.completedCalls, 0)
}

func Test_IndustryJobsUpdater_SkipsWithoutScope(t *testing.T) {
	jobsRepo := &mockIndustryJobsRepo{}
	queueRepo := &mockJobQueueRepo{}
	esiClient := &mockIndustryEsiClient{}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{
		charactersByUser: map[int64][]*repositories.Character{
			100: {
				{ID: 1001, Name: "No Scope Char", UserID: 100, EsiToken: "token", EsiScopes: "esi-assets.read_assets.v1", EsiTokenExpiresOn: time.Now().Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, &mockIndustryCorpRepo{}, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateAllUsers(context.Background())
	assert.NoError(t, err)

	assert.Empty(t, jobsRepo.upsertedJobs)
}

func Test_IndustryJobsUpdater_FetchesCorporationJobs(t *testing.T) {
	now := time.Now().UTC()
	cost := 2500000.0

	jobsRepo := &mockIndustryJobsRepo{}
	queueRepo := &mockJobQueueRepo{}
	esiClient := &mockIndustryEsiClient{
		jobsByChar: map[int64][]*client.EsiIndustryJob{},
		jobsByCorp: map[int64][]*client.EsiIndustryJob{
			98000001: {
				{
					JobID: 200001, InstallerID: 1001, FacilityID: 60003760, LocationID: 60003760,
					ActivityID: 9, BlueprintID: 5555, BlueprintTypeID: 46166,
					BlueprintLocationID: 60003760, OutputLocationID: 60003760,
					Runs: 100, Cost: &cost, Status: "active", Duration: 7200,
					StartDate: now.Format(time.RFC3339), EndDate: now.Add(2 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{charactersByUser: map[int64][]*repositories.Character{100: {}}}
	corpRepo := &mockIndustryCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 98000001, UserID: 100, Name: "Test Corp", EsiToken: "corp-token", EsiScopes: "esi-industry.read_corporation_jobs.v1", EsiExpiresOn: now.Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, corpRepo, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	assert.Len(t, jobsRepo.upsertedJobs[100], 1)
	assert.Equal(t, int64(200001), jobsRepo.upsertedJobs[100][0].JobID)
	assert.Equal(t, "active", jobsRepo.upsertedJobs[100][0].Status)
	assert.Equal(t, 9, jobsRepo.upsertedJobs[100][0].ActivityID)
	assert.Equal(t, "corporation", jobsRepo.upsertedJobs[100][0].Source)
	assert.Equal(t, int64(60003760), jobsRepo.upsertedJobs[100][0].StationID, "location_id should be normalized to station_id for corp jobs")
}

func Test_IndustryJobsUpdater_SkipsCorpsWithoutScope(t *testing.T) {
	now := time.Now().UTC()

	jobsRepo := &mockIndustryJobsRepo{}
	queueRepo := &mockJobQueueRepo{}
	esiClient := &mockIndustryEsiClient{
		jobsByChar: map[int64][]*client.EsiIndustryJob{},
		jobsByCorp: map[int64][]*client.EsiIndustryJob{
			98000001: {{JobID: 200001, InstallerID: 1001, FacilityID: 60003760, StationID: 60003760, ActivityID: 1, BlueprintID: 1, BlueprintTypeID: 787, BlueprintLocationID: 60003760, OutputLocationID: 60003760, Runs: 1, Status: "active", Duration: 3600, StartDate: now.Format(time.RFC3339), EndDate: now.Add(time.Hour).Format(time.RFC3339)}},
		},
	}
	userRepo := &mockIndustryUserRepo{userIDs: []int64{100}}
	charRepo := &mockIndustryCharRepo{charactersByUser: map[int64][]*repositories.Character{100: {}}}
	corpRepo := &mockIndustryCorpRepo{
		corpsByUser: map[int64][]repositories.PlayerCorporation{
			100: {
				{ID: 98000001, UserID: 100, Name: "No Scope Corp", EsiToken: "corp-token", EsiScopes: "esi-corporations.read_structures.v1", EsiExpiresOn: now.Add(time.Hour)},
			},
		},
	}

	updater := updaters.NewIndustryJobsUpdater(userRepo, charRepo, corpRepo, jobsRepo, queueRepo, esiClient)
	err := updater.UpdateUserJobs(context.Background(), 100)
	assert.NoError(t, err)

	// Should not have upserted any jobs since corp lacks the industry scope
	assert.Empty(t, jobsRepo.upsertedJobs)
}
