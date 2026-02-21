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

type PiPlanetsRepository interface {
	UpsertPlanets(ctx context.Context, characterID, userID int64, planets []*models.PiPlanet) error
	UpsertColony(ctx context.Context, characterID, planetID int64, pins []*models.PiPin, contents []*models.PiPinContent, routes []*models.PiRoute) error
	GetPlanetsForUser(ctx context.Context, userID int64) ([]*models.PiPlanet, error)
	GetPinsForPlanets(ctx context.Context, userID int64) ([]*models.PiPin, error)
	UpdateStallNotifiedAt(ctx context.Context, characterID, planetID int64, notifiedAt *time.Time) error
}

type PiEsiClient interface {
	GetCharacterPlanets(ctx context.Context, characterID int64, token string) ([]*client.EsiPiPlanet, error)
	GetCharacterPlanetDetails(ctx context.Context, characterID, planetID int64, token string) (*client.EsiPiColony, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*client.RefreshedToken, error)
}

type PiUserRepository interface {
	GetAllIDs(ctx context.Context) ([]int64, error)
}

type PiCharacterRepository interface {
	GetAll(ctx context.Context, baseUserID int64) ([]*repositories.Character, error)
	UpdateTokens(ctx context.Context, id, userID int64, token, refreshToken string, expiresOn time.Time) error
}

type PiSolarSystemRepository interface {
	GetNames(ctx context.Context, ids []int64) (map[int64]string, error)
}

type PiSchematicRepository interface {
	GetAllSchematics(ctx context.Context) ([]*models.SdePlanetSchematic, error)
}

type PiUpdater struct {
	userRepo      PiUserRepository
	characterRepo PiCharacterRepository
	piRepo        PiPlanetsRepository
	esiClient     PiEsiClient
	systemRepo    PiSolarSystemRepository
	schematicRepo PiSchematicRepository
	notifier      PiStallNotifier
}

func NewPiUpdater(
	userRepo PiUserRepository,
	characterRepo PiCharacterRepository,
	piRepo PiPlanetsRepository,
	esiClient PiEsiClient,
	systemRepo PiSolarSystemRepository,
	schematicRepo PiSchematicRepository,
) *PiUpdater {
	return &PiUpdater{
		userRepo:      userRepo,
		characterRepo: characterRepo,
		piRepo:        piRepo,
		esiClient:     esiClient,
		systemRepo:    systemRepo,
		schematicRepo: schematicRepo,
	}
}

// WithStallNotifier sets an optional notifier for PI stall alerts.
func (u *PiUpdater) WithStallNotifier(notifier PiStallNotifier) {
	u.notifier = notifier
}

// UpdateAllUsers refreshes PI data for every user in the system.
func (u *PiUpdater) UpdateAllUsers(ctx context.Context) error {
	userIDs, err := u.userRepo.GetAllIDs(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get user IDs for PI update")
	}

	for _, userID := range userIDs {
		if err := u.UpdateUserPlanets(ctx, userID); err != nil {
			log.Error("failed to update PI planets for user", "userID", userID, "error", err)
		}
	}

	return nil
}

// UpdateUserPlanets refreshes PI data for all characters belonging to a user.
func (u *PiUpdater) UpdateUserPlanets(ctx context.Context, userID int64) error {
	characters, err := u.characterRepo.GetAll(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user characters for PI update")
	}

	charNamesByID := map[int64]string{}
	for _, char := range characters {
		charNamesByID[char.ID] = char.Name

		if !strings.Contains(char.EsiScopes, "esi-planets.manage_planets.v1") {
			continue
		}

		if err := u.updateCharacterPlanets(ctx, char, userID); err != nil {
			log.Error("failed to update PI for character", "characterID", char.ID, "error", err)
		}
	}

	// Check for stalls after all data is updated
	if u.notifier != nil {
		u.checkStalls(ctx, userID, charNamesByID)
	}

	return nil
}

func (u *PiUpdater) updateCharacterPlanets(ctx context.Context, char *repositories.Character, userID int64) error {
	token, refresh, expire := char.EsiToken, char.EsiRefreshToken, char.EsiTokenExpiresOn

	if time.Now().After(expire) {
		refreshed, err := u.esiClient.RefreshAccessToken(ctx, refresh)
		if err != nil {
			return errors.Wrapf(err, "failed to refresh token for character %d", char.ID)
		}
		token = refreshed.AccessToken
		refresh = refreshed.RefreshToken
		expire = refreshed.Expiry

		err = u.characterRepo.UpdateTokens(ctx, char.ID, char.UserID, token, refresh, expire)
		if err != nil {
			return errors.Wrapf(err, "failed to persist refreshed token for character %d", char.ID)
		}
		log.Info("refreshed ESI token for character (PI)", "characterID", char.ID)
	}

	esiPlanets, err := u.esiClient.GetCharacterPlanets(ctx, char.ID, token)
	if err != nil {
		return errors.Wrap(err, "failed to get character planets from ESI")
	}

	planets := []*models.PiPlanet{}
	for _, ep := range esiPlanets {
		lastUpdate, err := time.Parse(time.RFC3339, ep.LastUpdate)
		if err != nil {
			return errors.Wrapf(err, "failed to parse last_update for planet %d", ep.PlanetID)
		}

		planets = append(planets, &models.PiPlanet{
			CharacterID:   char.ID,
			UserID:        userID,
			PlanetID:      ep.PlanetID,
			PlanetType:    ep.PlanetType,
			SolarSystemID: ep.SolarSystemID,
			UpgradeLevel:  ep.UpgradeLevel,
			NumPins:       ep.NumPins,
			LastUpdate:    lastUpdate,
		})
	}

	err = u.piRepo.UpsertPlanets(ctx, char.ID, userID, planets)
	if err != nil {
		return errors.Wrap(err, "failed to upsert PI planets")
	}

	for _, ep := range esiPlanets {
		colony, err := u.esiClient.GetCharacterPlanetDetails(ctx, char.ID, ep.PlanetID, token)
		if err != nil {
			log.Error("failed to get planet details from ESI", "characterID", char.ID, "planetID", ep.PlanetID, "error", err)
			continue
		}

		pins, contents := convertColonyPins(char.ID, ep.PlanetID, colony)
		routes := convertColonyRoutes(char.ID, ep.PlanetID, colony)

		err = u.piRepo.UpsertColony(ctx, char.ID, ep.PlanetID, pins, contents, routes)
		if err != nil {
			log.Error("failed to upsert colony data", "characterID", char.ID, "planetID", ep.PlanetID, "error", err)
		}
	}

	return nil
}

// classifyPin determines the category of a PI pin based on its ESI data.
// Extractors and factories are identified by their detail fields; other types
// are classified by well-known type_id sets from the EVE SDE.
func classifyPin(pin *client.EsiPiPin) string {
	if pin.ExtractorDetails != nil {
		return "extractor"
	}
	if pin.FactoryDetails != nil {
		return "factory"
	}
	// Factories may lack factory_details but still have a top-level schematic_id
	if pin.SchematicID != nil {
		return "factory"
	}
	switch pin.TypeID {
	case 2256, 2542, 2543, 2544:
		return "launchpad"
	case 2257, 2535, 2536, 2541:
		return "storage"
	case 2524, 2525, 2526, 2527, 2528, 2529, 2530, 2531:
		return "command_center"
	case 2473, 2480, 2481:
		return "factory"
	default:
		return "unknown"
	}
}

func convertColonyPins(characterID, planetID int64, colony *client.EsiPiColony) ([]*models.PiPin, []*models.PiPinContent) {
	pins := []*models.PiPin{}
	contents := []*models.PiPinContent{}

	for i := range colony.Pins {
		esiPin := &colony.Pins[i]

		pin := &models.PiPin{
			CharacterID: characterID,
			PlanetID:    planetID,
			PinID:       esiPin.PinID,
			TypeID:      esiPin.TypeID,
			Latitude:    &esiPin.Latitude,
			Longitude:   &esiPin.Longitude,
			PinCategory: classifyPin(esiPin),
		}

		if esiPin.InstallTime != nil {
			t, err := time.Parse(time.RFC3339, *esiPin.InstallTime)
			if err == nil {
				pin.InstallTime = &t
			}
		}
		if esiPin.ExpiryTime != nil {
			t, err := time.Parse(time.RFC3339, *esiPin.ExpiryTime)
			if err == nil {
				pin.ExpiryTime = &t
			}
		}
		if esiPin.LastCycleStart != nil {
			t, err := time.Parse(time.RFC3339, *esiPin.LastCycleStart)
			if err == nil {
				pin.LastCycleStart = &t
			}
		}

		if esiPin.SchematicID != nil {
			s := *esiPin.SchematicID
			pin.SchematicID = &s
		}
		if esiPin.FactoryDetails != nil {
			s := esiPin.FactoryDetails.SchematicID
			pin.SchematicID = &s
		}

		if esiPin.ExtractorDetails != nil {
			cycleTime := esiPin.ExtractorDetails.CycleTime
			pin.ExtractorCycleTime = &cycleTime

			headRadius := esiPin.ExtractorDetails.HeadRadius
			pin.ExtractorHeadRadius = &headRadius

			productTypeID := esiPin.ExtractorDetails.ProductTypeID
			pin.ExtractorProductTypeID = &productTypeID

			qtyPerCycle := esiPin.ExtractorDetails.QtyPerCycle
			pin.ExtractorQtyPerCycle = &qtyPerCycle

			numHeads := len(esiPin.ExtractorDetails.Heads)
			pin.ExtractorNumHeads = &numHeads
		}

		pins = append(pins, pin)

		for _, c := range esiPin.Contents {
			contents = append(contents, &models.PiPinContent{
				CharacterID: characterID,
				PlanetID:    planetID,
				PinID:       esiPin.PinID,
				TypeID:      c.TypeID,
				Amount:      int64(c.Amount),
			})
		}
	}

	return pins, contents
}

func convertColonyRoutes(characterID, planetID int64, colony *client.EsiPiColony) []*models.PiRoute {
	routes := []*models.PiRoute{}

	for _, r := range colony.Routes {
		routes = append(routes, &models.PiRoute{
			CharacterID:      characterID,
			PlanetID:         planetID,
			RouteID:          r.RouteID,
			SourcePinID:      r.SourcePinID,
			DestinationPinID: r.DestinationPinID,
			ContentTypeID:    r.ContentTypeID,
			Quantity:         int64(r.Quantity),
		})
	}

	return routes
}

// Stall detection thresholds (same as controller)
const (
	piStaleDataThreshold  = 48 * time.Hour
	piFactoryIdleMultiple = 2
)

// checkStalls detects newly stalled planets and sends notifications.
// Only notifies on state transitions (running → stalled), not repeatedly.
func (u *PiUpdater) checkStalls(ctx context.Context, userID int64, charNames map[int64]string) {
	planets, err := u.piRepo.GetPlanetsForUser(ctx, userID)
	if err != nil {
		log.Error("failed to get planets for stall check", "userID", userID, "error", err)
		return
	}

	if len(planets) == 0 {
		return
	}

	pins, err := u.piRepo.GetPinsForPlanets(ctx, userID)
	if err != nil {
		log.Error("failed to get pins for stall check", "userID", userID, "error", err)
		return
	}

	schematics, err := u.schematicRepo.GetAllSchematics(ctx)
	if err != nil {
		log.Error("failed to get schematics for stall check", "userID", userID, "error", err)
		return
	}
	schematicMap := map[int]*models.SdePlanetSchematic{}
	for _, s := range schematics {
		schematicMap[int(s.SchematicID)] = s
	}

	// Group pins by planet
	type planetKey struct {
		characterID int64
		planetID    int64
	}
	pinsByPlanet := map[planetKey][]*models.PiPin{}
	for _, pin := range pins {
		k := planetKey{pin.CharacterID, pin.PlanetID}
		pinsByPlanet[k] = append(pinsByPlanet[k], pin)
	}

	// Collect solar system IDs for name resolution
	systemIDs := []int64{}
	for _, p := range planets {
		systemIDs = append(systemIDs, p.SolarSystemID)
	}
	systemNames, err := u.systemRepo.GetNames(ctx, systemIDs)
	if err != nil {
		log.Error("failed to get system names for stall check", "userID", userID, "error", err)
		systemNames = map[int64]string{}
	}

	now := time.Now()

	// Collect all new stall alerts, then send as one notification
	var newAlerts []*PiStallAlert
	type stalledPlanet struct {
		characterID int64
		planetID    int64
	}
	var newlyStalled []stalledPlanet

	for _, planet := range planets {
		k := planetKey{planet.CharacterID, planet.PlanetID}
		planetPins := pinsByPlanet[k]

		stalledPins := []PiStalledPin{}

		for _, pin := range planetPins {
			switch pin.PinCategory {
			case "extractor":
				if pin.ExpiryTime != nil && pin.ExpiryTime.Before(now) {
					stalledPins = append(stalledPins, PiStalledPin{
						PinCategory: "extractor",
						Reason:      "expired",
					})
				}
			case "factory":
				if pin.SchematicID != nil && pin.LastCycleStart != nil {
					schematic := schematicMap[*pin.SchematicID]
					if schematic != nil && schematic.CycleTime > 0 {
						expectedNext := pin.LastCycleStart.Add(time.Duration(schematic.CycleTime*piFactoryIdleMultiple) * time.Second)
						if now.After(expectedNext) {
							stalledPins = append(stalledPins, PiStalledPin{
								PinCategory: "factory",
								Reason:      "stalled",
							})
						}
					}
				}
			}
		}

		isStalled := len(stalledPins) > 0
		wasNotified := planet.LastStallNotifiedAt != nil

		if isStalled && !wasNotified {
			newAlerts = append(newAlerts, &PiStallAlert{
				CharacterName:   charNames[planet.CharacterID],
				PlanetType:      planet.PlanetType,
				SolarSystemName: systemNames[planet.SolarSystemID],
				StalledPins:     stalledPins,
			})
			newlyStalled = append(newlyStalled, stalledPlanet{planet.CharacterID, planet.PlanetID})
		} else if !isStalled && wasNotified {
			// Planet recovered — clear the notification timestamp
			if err := u.piRepo.UpdateStallNotifiedAt(ctx, planet.CharacterID, planet.PlanetID, nil); err != nil {
				log.Error("failed to clear stall notified timestamp", "characterID", planet.CharacterID, "planetID", planet.PlanetID, "error", err)
			}
		}
	}

	// Send one batched notification for all newly stalled planets
	if len(newAlerts) > 0 {
		u.notifier.NotifyPiStalls(ctx, userID, newAlerts)

		notifiedAt := now
		for _, sp := range newlyStalled {
			if err := u.piRepo.UpdateStallNotifiedAt(ctx, sp.characterID, sp.planetID, &notifiedAt); err != nil {
				log.Error("failed to update stall notified timestamp", "characterID", sp.characterID, "planetID", sp.planetID, "error", err)
			}
		}
	}
}
