package updaters_test

import (
	"context"
	"testing"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// --- Mocks for AutoProductionUpdater ---

type mockAutoMarkersRepo struct {
	markers []*models.StockpileMarker
	err     error
}

func (m *mockAutoMarkersRepo) GetAutoProductionMarkers(ctx context.Context) ([]*models.StockpileMarker, error) {
	return m.markers, m.err
}

type mockAutoAssetsRepo struct {
	response *repositories.StockpilesResponse
	err      error
}

func (m *mockAutoAssetsRepo) GetStockpileDeficits(ctx context.Context, user int64) (*repositories.StockpilesResponse, error) {
	return m.response, m.err
}

type mockAutoPlansRepo struct {
	plan map[int64]*models.ProductionPlan
	err  error
}

func (m *mockAutoPlansRepo) GetByID(ctx context.Context, id, userID int64) (*models.ProductionPlan, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.plan[id], nil
}

type mockAutoPlanRunsRepo struct {
	created        *models.ProductionPlanRun
	createErr      error
	pendingOutput  int64
	pendingErr     error
	createdRuns    []*models.ProductionPlanRun
}

func (m *mockAutoPlanRunsRepo) Create(ctx context.Context, run *models.ProductionPlanRun) (*models.ProductionPlanRun, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdRuns = append(m.createdRuns, run)
	if m.created != nil {
		return m.created, nil
	}
	r := *run
	r.ID = 999
	return &r, nil
}

func (m *mockAutoPlanRunsRepo) GetPendingOutputForPlan(ctx context.Context, planID, userID int64) (int64, error) {
	return m.pendingOutput, m.pendingErr
}

type mockAutoMarketRepo struct {
	jitaPrices     map[int64]*models.MarketPrice
	jitaErr        error
	adjustedPrices map[int64]float64
	adjustedErr    error
}

func (m *mockAutoMarketRepo) GetAllJitaPrices(ctx context.Context) (map[int64]*models.MarketPrice, error) {
	return m.jitaPrices, m.jitaErr
}

func (m *mockAutoMarketRepo) GetAllAdjustedPrices(ctx context.Context) (map[int64]float64, error) {
	return m.adjustedPrices, m.adjustedErr
}

type mockAutoQueueRepo struct {
	created    []*models.IndustryJobQueueEntry
	createErr  error
	slotUsage  map[int64]map[string]int
	slotErr    error
}

func (m *mockAutoQueueRepo) Create(ctx context.Context, entry *models.IndustryJobQueueEntry) (*models.IndustryJobQueueEntry, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.created = append(m.created, entry)
	e := *entry
	e.ID = int64(len(m.created))
	return &e, nil
}

func (m *mockAutoQueueRepo) GetSlotUsage(ctx context.Context, userID int64) (map[int64]map[string]int, error) {
	return m.slotUsage, m.slotErr
}

type mockAutoCharRepo struct {
	names    map[int64]string
	namesErr error
}

func (m *mockAutoCharRepo) GetNames(ctx context.Context, userID int64) (map[int64]string, error) {
	return m.names, m.namesErr
}

type mockAutoSkillsRepo struct {
	skills    []*models.CharacterSkill
	skillsErr error
}

func (m *mockAutoSkillsRepo) GetSkillsForUser(ctx context.Context, userID int64) ([]*models.CharacterSkill, error) {
	return m.skills, m.skillsErr
}

type mockAutoSdeRepo struct {
	blueprint *repositories.ManufacturingBlueprintRow
	bpErr     error
	materials []*repositories.ManufacturingMaterialRow
	matErr    error
}

func (m *mockAutoSdeRepo) GetBlueprintForActivity(ctx context.Context, blueprintTypeID int64, activity string) (*repositories.ManufacturingBlueprintRow, error) {
	return m.blueprint, m.bpErr
}

func (m *mockAutoSdeRepo) GetBlueprintMaterialsForActivity(ctx context.Context, blueprintTypeID int64, activity string) ([]*repositories.ManufacturingMaterialRow, error) {
	return m.materials, m.matErr
}

// --- Helpers ---

func planIDPtr(id int64) *int64 {
	return &id
}

func makeMinimalPlan(userID, planID int64) *models.ProductionPlan {
	return &models.ProductionPlan{
		ID:            planID,
		UserID:        userID,
		Name:          "Test Plan",
		ProductTypeID: 587,
		Steps: []*models.ProductionPlanStep{
			{
				ID:              1,
				PlanID:          planID,
				BlueprintTypeID: 700,
				Activity:        "manufacturing",
				ProductTypeID:   587,
				ProductName:     "Rifter",
				MELevel:         0,
				TELevel:         0,
				FacilityTax:     0.0,
				Structure:       "none",
				Rig:             "none",
				Security:        "null",
			},
		},
	}
}

func makeMinimalBlueprint() *repositories.ManufacturingBlueprintRow {
	return &repositories.ManufacturingBlueprintRow{
		BlueprintTypeID: 700,
		ProductTypeID:   587,
		ProductName:     "Rifter",
		ProductQuantity: 1,
		Time:            3600,
		ProductVolume:   27289.0,
	}
}

func setupAutoProductionUpdater(
	markers []*models.StockpileMarker,
	deficits *repositories.StockpilesResponse,
	plan *models.ProductionPlan,
	pendingOutput int64,
	jitaPrices map[int64]*models.MarketPrice,
	adjustedPrices map[int64]float64,
) (*updaters.AutoProductionUpdater, *mockAutoPlanRunsRepo, *mockAutoQueueRepo) {
	markersRepo := &mockAutoMarkersRepo{markers: markers}
	assetsRepo := &mockAutoAssetsRepo{response: deficits}

	plansMap := map[int64]*models.ProductionPlan{}
	if plan != nil {
		plansMap[plan.ID] = plan
	}
	plansRepo := &mockAutoPlansRepo{plan: plansMap}

	if jitaPrices == nil {
		jitaPrices = map[int64]*models.MarketPrice{}
	}
	if adjustedPrices == nil {
		adjustedPrices = map[int64]float64{}
	}

	runsRepo := &mockAutoPlanRunsRepo{pendingOutput: pendingOutput}
	marketRepo := &mockAutoMarketRepo{jitaPrices: jitaPrices, adjustedPrices: adjustedPrices}
	queueRepo := &mockAutoQueueRepo{}
	charRepo := &mockAutoCharRepo{names: map[int64]string{}}
	skillsRepo := &mockAutoSkillsRepo{skills: []*models.CharacterSkill{}}
	sdeRepo := &mockAutoSdeRepo{
		blueprint: makeMinimalBlueprint(),
		materials: []*repositories.ManufacturingMaterialRow{},
	}

	u := updaters.NewAutoProductionUpdater(
		markersRepo,
		assetsRepo,
		plansRepo,
		runsRepo,
		marketRepo,
		queueRepo,
		charRepo,
		skillsRepo,
		sdeRepo,
	)

	return u, runsRepo, queueRepo
}

// --- Tests ---

func Test_AutoProduction_NoMarkersDoesNothing(t *testing.T) {
	u, runsRepo, queueRepo := setupAutoProductionUpdater(
		[]*models.StockpileMarker{},
		nil,
		nil,
		0,
		nil,
		nil,
	)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
	assert.Len(t, queueRepo.created, 0)
}

func Test_AutoProduction_MarkersRepoErrorPropagated(t *testing.T) {
	markersRepo := &mockAutoMarkersRepo{err: errors.New("db failure")}
	assetsRepo := &mockAutoAssetsRepo{}
	plansRepo := &mockAutoPlansRepo{plan: map[int64]*models.ProductionPlan{}}
	runsRepo := &mockAutoPlanRunsRepo{}
	marketRepo := &mockAutoMarketRepo{jitaPrices: map[int64]*models.MarketPrice{}, adjustedPrices: map[int64]float64{}}
	queueRepo := &mockAutoQueueRepo{}
	charRepo := &mockAutoCharRepo{names: map[int64]string{}}
	skillsRepo := &mockAutoSkillsRepo{}
	sdeRepo := &mockAutoSdeRepo{}

	u := updaters.NewAutoProductionUpdater(markersRepo, assetsRepo, plansRepo, runsRepo, marketRepo, queueRepo, charRepo, skillsRepo, sdeRepo)

	err := u.RunAll(context.Background())

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auto-production markers")
}

func Test_AutoProduction_MarkerWithNoPlanIDSkipped(t *testing.T) {
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 100, PlanID: nil, AutoProductionEnabled: true},
	}
	u, runsRepo, queueRepo := setupAutoProductionUpdater(
		markers, nil, nil, 0, nil, nil,
	)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
	assert.Len(t, queueRepo.created, 0)
}

func Test_AutoProduction_NoDeficitDoesNotCreateRun(t *testing.T) {
	planID := int64(10)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}

	// deficit = 0 (StockpileDelta >= 0)
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: 0},
		},
	}

	plan := makeMinimalPlan(1, planID)
	u, runsRepo, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 0, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
	assert.Len(t, queueRepo.created, 0)
}

func Test_AutoProduction_PendingOutputCoversDeficit(t *testing.T) {
	planID := int64(11)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true},
	}

	// gross deficit = 5 (delta = -5)
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -5},
		},
	}

	plan := makeMinimalPlan(1, planID)
	// pending output >= gross deficit, so net = 0
	u, runsRepo, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 10, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
	assert.Len(t, queueRepo.created, 0)
}

func Test_AutoProduction_CreatesRunAndJobsForNetDeficit(t *testing.T) {
	planID := int64(20)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}

	// gross deficit = 3 (delta = -3)
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -3},
		},
	}

	plan := makeMinimalPlan(1, planID)
	// pending output = 0, so net deficit = 3
	u, runsRepo, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 0, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 1)
	assert.Equal(t, int(3), runsRepo.createdRuns[0].Quantity)
	assert.Equal(t, planID, runsRepo.createdRuns[0].PlanID)
	// One merged job for the single-step plan
	assert.Len(t, queueRepo.created, 1)
	assert.Equal(t, int64(700), queueRepo.created[0].BlueprintTypeID)
	assert.Equal(t, "manufacturing", queueRepo.created[0].Activity)
	assert.Equal(t, 3, queueRepo.created[0].Runs)
}

func Test_AutoProduction_PartialPendingOutputReducesNetDeficit(t *testing.T) {
	planID := int64(21)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}

	// gross deficit = 10
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -10},
		},
	}

	plan := makeMinimalPlan(1, planID)
	// pending = 6, so net = 4
	u, runsRepo, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 6, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 1)
	assert.Equal(t, 4, runsRepo.createdRuns[0].Quantity)
	assert.Len(t, queueRepo.created, 1)
	assert.Equal(t, 4, queueRepo.created[0].Runs)
}

func Test_AutoProduction_MultipleTypeIDsInGroupSumsDeficits(t *testing.T) {
	planID := int64(22)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
		{UserID: 1, TypeID: 588, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}

	// Both types have deficits
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -4},
			{TypeID: 588, StockpileDelta: -6},
			{TypeID: 999, StockpileDelta: -100}, // unrelated type, should be ignored
		},
	}

	plan := makeMinimalPlan(1, planID)
	u, runsRepo, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 0, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 1)
	// gross = 4 + 6 = 10, pending = 0, net = 10
	assert.Equal(t, 10, runsRepo.createdRuns[0].Quantity)
	assert.Len(t, queueRepo.created, 1)
	assert.Equal(t, 10, queueRepo.created[0].Runs)
}

func Test_AutoProduction_NilPlanDoesNotCrash(t *testing.T) {
	planID := int64(30)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true},
	}
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -5},
		},
	}

	// Plan repo returns nil (plan not found)
	markersRepo := &mockAutoMarkersRepo{markers: markers}
	assetsRepo := &mockAutoAssetsRepo{response: deficits}
	plansRepo := &mockAutoPlansRepo{plan: map[int64]*models.ProductionPlan{}} // empty map = nil returned
	runsRepo := &mockAutoPlanRunsRepo{}
	marketRepo := &mockAutoMarketRepo{jitaPrices: map[int64]*models.MarketPrice{}, adjustedPrices: map[int64]float64{}}
	queueRepo := &mockAutoQueueRepo{}
	charRepo := &mockAutoCharRepo{names: map[int64]string{}}
	skillsRepo := &mockAutoSkillsRepo{}
	sdeRepo := &mockAutoSdeRepo{}

	u := updaters.NewAutoProductionUpdater(markersRepo, assetsRepo, plansRepo, runsRepo, marketRepo, queueRepo, charRepo, skillsRepo, sdeRepo)

	err := u.RunAll(context.Background())

	// Should not error — missing plan is handled gracefully via logging
	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
}

func Test_AutoProduction_DeficitsRepoErrorLogsAndContinues(t *testing.T) {
	planID := int64(31)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true},
	}

	plan := makeMinimalPlan(1, planID)
	markersRepo := &mockAutoMarkersRepo{markers: markers}
	// assets repo returns error
	assetsRepo := &mockAutoAssetsRepo{err: errors.New("assets db down")}
	plansMap := map[int64]*models.ProductionPlan{planID: plan}
	plansRepo := &mockAutoPlansRepo{plan: plansMap}
	runsRepo := &mockAutoPlanRunsRepo{}
	marketRepo := &mockAutoMarketRepo{jitaPrices: map[int64]*models.MarketPrice{}, adjustedPrices: map[int64]float64{}}
	queueRepo := &mockAutoQueueRepo{}
	charRepo := &mockAutoCharRepo{names: map[int64]string{}}
	skillsRepo := &mockAutoSkillsRepo{}
	sdeRepo := &mockAutoSdeRepo{}

	u := updaters.NewAutoProductionUpdater(markersRepo, assetsRepo, plansRepo, runsRepo, marketRepo, queueRepo, charRepo, skillsRepo, sdeRepo)

	err := u.RunAll(context.Background())

	// RunAll logs per-group errors and continues — returns nil
	assert.NoError(t, err)
	assert.Len(t, runsRepo.createdRuns, 0)
}

func Test_AutoProduction_RunCreationErrorLogsAndContinues(t *testing.T) {
	planID := int64(32)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -5},
		},
	}

	plan := makeMinimalPlan(1, planID)
	markersRepo := &mockAutoMarkersRepo{markers: markers}
	assetsRepo := &mockAutoAssetsRepo{response: deficits}
	plansRepo := &mockAutoPlansRepo{plan: map[int64]*models.ProductionPlan{planID: plan}}
	runsRepo := &mockAutoPlanRunsRepo{createErr: errors.New("run create failed")}
	marketRepo := &mockAutoMarketRepo{jitaPrices: map[int64]*models.MarketPrice{}, adjustedPrices: map[int64]float64{}}
	queueRepo := &mockAutoQueueRepo{}
	charRepo := &mockAutoCharRepo{names: map[int64]string{}}
	skillsRepo := &mockAutoSkillsRepo{}
	sdeRepo := &mockAutoSdeRepo{
		blueprint: makeMinimalBlueprint(),
		materials: []*repositories.ManufacturingMaterialRow{},
	}

	u := updaters.NewAutoProductionUpdater(markersRepo, assetsRepo, plansRepo, runsRepo, marketRepo, queueRepo, charRepo, skillsRepo, sdeRepo)

	err := u.RunAll(context.Background())

	// RunAll logs per-group errors and continues — returns nil
	assert.NoError(t, err)
	assert.Len(t, queueRepo.created, 0)
}

func Test_AutoProduction_NoteContainsPlanName(t *testing.T) {
	planID := int64(40)
	markers := []*models.StockpileMarker{
		{UserID: 1, TypeID: 587, PlanID: &planID, AutoProductionEnabled: true, AutoProductionParallelism: 0},
	}
	deficits := &repositories.StockpilesResponse{
		Items: []*repositories.StockpileItem{
			{TypeID: 587, StockpileDelta: -2},
		},
	}

	plan := makeMinimalPlan(1, planID)
	plan.Name = "My Rifter Plan"
	u, _, queueRepo := setupAutoProductionUpdater(markers, deficits, plan, 0, nil, nil)

	err := u.RunAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, queueRepo.created, 1)
	assert.NotNil(t, queueRepo.created[0].Notes)
	assert.Contains(t, *queueRepo.created[0].Notes, "My Rifter Plan")
}
