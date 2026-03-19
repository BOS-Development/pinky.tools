package repositories

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type JobSlotRentals struct {
	db *sql.DB
}

func NewJobSlotRentals(db *sql.DB) *JobSlotRentals {
	return &JobSlotRentals{db: db}
}

// CalculateSlotInventory computes available slots per character per activity.
// Queries character_skills, esi_industry_jobs, industry_job_queue, and existing listings.
func (r *JobSlotRentals) CalculateSlotInventory(ctx context.Context, userID int64) ([]*models.CharacterSlotInventory, error) {
	// Get all characters for this user with their slot capacities from skills
	charQuery := `
		SELECT DISTINCT c.id, c.name
		FROM characters c
		WHERE c.user_id = $1
		ORDER BY c.name
	`
	rows, err := r.db.QueryContext(ctx, charQuery, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query characters")
	}
	defer rows.Close()

	type charInfo struct {
		id   int64
		name string
	}
	var chars []*charInfo
	for rows.Next() {
		var c charInfo
		if err := rows.Scan(&c.id, &c.name); err != nil {
			return nil, errors.Wrap(err, "failed to scan character")
		}
		chars = append(chars, &c)
	}

	if len(chars) == 0 {
		return []*models.CharacterSlotInventory{}, nil
	}

	// Get skills for all characters
	skillsQuery := `
		SELECT character_id, skill_id, trained_level
		FROM character_skills
		WHERE user_id = $1 AND skill_id = ANY($2)
	`
	skillIDs := []int64{3380, 3387, 24625, 45746, 45748, 45749, 3402, 3406, 24624}
	skillRows, err := r.db.QueryContext(ctx, skillsQuery, userID, pq.Array(skillIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query character skills")
	}
	defer skillRows.Close()

	skillsByChar := make(map[int64]map[int64]int)
	for skillRows.Next() {
		var charID, skillID int64
		var level int
		if err := skillRows.Scan(&charID, &skillID, &level); err != nil {
			return nil, errors.Wrap(err, "failed to scan skill")
		}
		if skillsByChar[charID] == nil {
			skillsByChar[charID] = make(map[int64]int)
		}
		skillsByChar[charID][skillID] = level
	}

	// Calculate max slots per character per activity
	type slotMax struct {
		mfg     int
		react   int
		science int
	}
	maxByChar := make(map[int64]slotMax)
	for _, c := range chars {
		skills := skillsByChar[c.id]
		if skills == nil {
			skills = make(map[int64]int)
		}

		// Manufacturing: 1 + MassProduction + AdvMassProduction
		mfg := 1 + skills[3387] + skills[24625]

		// Reactions: require Reactions >= 1, then 1 + MassReactions + AdvMassReactions
		react := 0
		if skills[45746] >= 1 {
			react = 1 + skills[45748] + skills[45749]
		}

		// Science: require Science >= 1, then 1 + LabOp + AdvLabOp
		science := 0
		if skills[3402] >= 1 {
			science = 1 + skills[3406] + skills[24624]
		}

		maxByChar[c.id] = slotMax{mfg: mfg, react: react, science: science}
	}

	// Get in-use slots from ESI industry jobs (active jobs)
	esiJobsQuery := `
		SELECT installer_id, activity_id, COUNT(*)
		FROM esi_industry_jobs
		WHERE user_id = $1 AND status IN ('active', 'paused')
		GROUP BY installer_id, activity_id
	`
	esiRows, err := r.db.QueryContext(ctx, esiJobsQuery, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query ESI industry jobs")
	}
	defer esiRows.Close()

	inUseByChar := make(map[int64]map[string]int)
	for esiRows.Next() {
		var charID int64
		var activityID int
		var count int
		if err := esiRows.Scan(&charID, &activityID, &count); err != nil {
			return nil, errors.Wrap(err, "failed to scan ESI job")
		}

		if inUseByChar[charID] == nil {
			inUseByChar[charID] = make(map[string]int)
		}

		// Map ESI activity_id to our activity names
		// 1→manufacturing, 3→te_research, 4→me_research, 5→copying, 8→invention, 9→reaction
		switch activityID {
		case 1:
			inUseByChar[charID]["manufacturing"] += count
		case 3:
			inUseByChar[charID]["te_research"] += count
		case 4:
			inUseByChar[charID]["me_research"] += count
		case 5:
			inUseByChar[charID]["copying"] += count
		case 8:
			inUseByChar[charID]["invention"] += count
		case 9:
			inUseByChar[charID]["reaction"] += count
		}
	}

	// Get reserved slots from industry_job_queue (planned jobs)
	queueQuery := `
		SELECT character_id, activity, COUNT(*)
		FROM industry_job_queue
		WHERE user_id = $1 AND status IN ('planned', 'active')
		GROUP BY character_id, activity
	`
	queueRows, err := r.db.QueryContext(ctx, queueQuery, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query job queue")
	}
	defer queueRows.Close()

	reservedByChar := make(map[int64]map[string]int)
	for queueRows.Next() {
		var charID sql.NullInt64
		var activity string
		var count int
		if err := queueRows.Scan(&charID, &activity, &count); err != nil {
			return nil, errors.Wrap(err, "failed to scan queue entry")
		}

		// Skip entries with null character_id
		if !charID.Valid {
			continue
		}

		if reservedByChar[charID.Int64] == nil {
			reservedByChar[charID.Int64] = make(map[string]int)
		}
		reservedByChar[charID.Int64][activity] += count
	}

	// Get listed slots from job_slot_rental_listings
	listingsQuery := `
		SELECT character_id, activity_type, SUM(slots_listed)
		FROM job_slot_rental_listings
		WHERE user_id = $1 AND is_active = true
		GROUP BY character_id, activity_type
	`
	listingRows, err := r.db.QueryContext(ctx, listingsQuery, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query job slot listings")
	}
	defer listingRows.Close()

	listedByChar := make(map[int64]map[string]int)
	for listingRows.Next() {
		var charID int64
		var activityType string
		var slotsListed int
		if err := listingRows.Scan(&charID, &activityType, &slotsListed); err != nil {
			return nil, errors.Wrap(err, "failed to scan listing")
		}

		if listedByChar[charID] == nil {
			listedByChar[charID] = make(map[string]int)
		}
		listedByChar[charID][activityType] += slotsListed
	}

	// Build inventory for each character
	result := []*models.CharacterSlotInventory{}
	for _, c := range chars {
		max := maxByChar[c.id]
		inUse := inUseByChar[c.id]
		if inUse == nil {
			inUse = make(map[string]int)
		}
		reserved := reservedByChar[c.id]
		if reserved == nil {
			reserved = make(map[string]int)
		}
		listed := listedByChar[c.id]
		if listed == nil {
			listed = make(map[string]int)
		}

		// Science activities share a single slot pool
		scienceInUse := inUse["te_research"] + inUse["me_research"] + inUse["copying"] + inUse["invention"]
		scienceReserved := reserved["te_research"] + reserved["me_research"] + reserved["copying"] + reserved["invention"]

		slotsByActivity := make(map[string]*models.ActivitySlotInfo)

		// Manufacturing
		mfgAvail := max.mfg - inUse["manufacturing"] - reserved["manufacturing"] - listed["manufacturing"]
		if mfgAvail < 0 {
			mfgAvail = 0
		}
		slotsByActivity["manufacturing"] = &models.ActivitySlotInfo{
			ActivityType:   "manufacturing",
			SlotsMax:       max.mfg,
			SlotsInUse:     inUse["manufacturing"],
			SlotsReserved:  reserved["manufacturing"],
			SlotsAvailable: mfgAvail,
			SlotsListed:    listed["manufacturing"],
		}

		// Reaction
		reactAvail := max.react - inUse["reaction"] - reserved["reaction"] - listed["reaction"]
		if reactAvail < 0 {
			reactAvail = 0
		}
		slotsByActivity["reaction"] = &models.ActivitySlotInfo{
			ActivityType:   "reaction",
			SlotsMax:       max.react,
			SlotsInUse:     inUse["reaction"],
			SlotsReserved:  reserved["reaction"],
			SlotsAvailable: reactAvail,
			SlotsListed:    listed["reaction"],
		}

		// Science activities - all share the same pool
		sciActivities := []string{"te_research", "me_research", "copying", "invention"}
		for _, act := range sciActivities {
			sciListedTotal := listed["te_research"] + listed["me_research"] + listed["copying"] + listed["invention"]
			sciAvail := max.science - scienceInUse - scienceReserved - sciListedTotal
			if sciAvail < 0 {
				sciAvail = 0
			}

			slotsByActivity[act] = &models.ActivitySlotInfo{
				ActivityType:   act,
				SlotsMax:       max.science,
				SlotsInUse:     scienceInUse,
				SlotsReserved:  scienceReserved,
				SlotsAvailable: sciAvail,
				SlotsListed:    sciListedTotal,
			}
		}

		result = append(result, &models.CharacterSlotInventory{
			CharacterID:     c.id,
			CharacterName:   c.name,
			SlotsByActivity: slotsByActivity,
		})
	}

	return result, nil
}

// GetByUser returns all listings for a user
func (r *JobSlotRentals) GetByUser(ctx context.Context, userID int64) ([]*models.JobSlotRentalListing, error) {
	query := `
		SELECT
			l.id,
			l.user_id,
			l.character_id,
			c.name AS character_name,
			l.activity_type,
			l.slots_listed,
			l.price_amount,
			l.pricing_unit,
			l.location_id,
			COALESCE(s.name, '') AS location_name,
			l.notes,
			l.is_active,
			l.created_at,
			l.updated_at
		FROM job_slot_rental_listings l
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		LEFT JOIN stations s ON s.station_id = l.location_id
		WHERE l.user_id = $1 AND l.is_active = true
		ORDER BY l.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query job slot listings")
	}
	defer rows.Close()

	listings := []*models.JobSlotRentalListing{}
	for rows.Next() {
		var listing models.JobSlotRentalListing
		err = rows.Scan(
			&listing.ID,
			&listing.UserID,
			&listing.CharacterID,
			&listing.CharacterName,
			&listing.ActivityType,
			&listing.SlotsListed,
			&listing.PriceAmount,
			&listing.PricingUnit,
			&listing.LocationID,
			&listing.LocationName,
			&listing.Notes,
			&listing.IsActive,
			&listing.CreatedAt,
			&listing.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan job slot listing")
		}
		listings = append(listings, &listing)
	}

	return listings, nil
}

// GetBrowsableListings returns listings from permitted sellers
func (r *JobSlotRentals) GetBrowsableListings(ctx context.Context, viewerUserID int64, sellerUserIDs []int64) ([]*models.JobSlotRentalListing, error) {
	if len(sellerUserIDs) == 0 {
		return []*models.JobSlotRentalListing{}, nil
	}

	query := `
		SELECT
			l.id,
			l.user_id,
			l.character_id,
			c.name AS character_name,
			l.activity_type,
			l.slots_listed,
			l.price_amount,
			l.pricing_unit,
			l.location_id,
			COALESCE(s.name, '') AS location_name,
			l.notes,
			l.is_active,
			l.created_at,
			l.updated_at
		FROM job_slot_rental_listings l
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		LEFT JOIN stations s ON s.station_id = l.location_id
		WHERE l.user_id = ANY($1) AND l.is_active = true
		ORDER BY l.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(sellerUserIDs))
	if err != nil {
		return nil, errors.Wrap(err, "failed to query browsable job slot listings")
	}
	defer rows.Close()

	listings := []*models.JobSlotRentalListing{}
	for rows.Next() {
		var listing models.JobSlotRentalListing
		err = rows.Scan(
			&listing.ID,
			&listing.UserID,
			&listing.CharacterID,
			&listing.CharacterName,
			&listing.ActivityType,
			&listing.SlotsListed,
			&listing.PriceAmount,
			&listing.PricingUnit,
			&listing.LocationID,
			&listing.LocationName,
			&listing.Notes,
			&listing.IsActive,
			&listing.CreatedAt,
			&listing.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan browsable job slot listing")
		}
		listings = append(listings, &listing)
	}

	return listings, nil
}

// Create inserts a new listing
func (r *JobSlotRentals) Create(ctx context.Context, listing *models.JobSlotRentalListing) error {
	query := `
		INSERT INTO job_slot_rental_listings
		(user_id, character_id, activity_type, slots_listed, price_amount, pricing_unit, location_id, notes, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		listing.UserID,
		listing.CharacterID,
		listing.ActivityType,
		listing.SlotsListed,
		listing.PriceAmount,
		listing.PricingUnit,
		listing.LocationID,
		listing.Notes,
		listing.IsActive,
	).Scan(&listing.ID, &listing.CreatedAt, &listing.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create job slot listing")
	}

	return nil
}

// Update modifies an existing listing
func (r *JobSlotRentals) Update(ctx context.Context, listing *models.JobSlotRentalListing) error {
	query := `
		UPDATE job_slot_rental_listings
		SET slots_listed = $1, price_amount = $2, pricing_unit = $3, location_id = $4, notes = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7 AND user_id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		listing.SlotsListed,
		listing.PriceAmount,
		listing.PricingUnit,
		listing.LocationID,
		listing.Notes,
		listing.IsActive,
		listing.ID,
		listing.UserID,
	)

	if err != nil {
		return errors.Wrap(err, "failed to update job slot listing")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("job slot listing not found or user is not the owner")
	}

	return nil
}

// Delete soft-deletes a listing
func (r *JobSlotRentals) Delete(ctx context.Context, listingID int64, userID int64) error {
	query := `
		UPDATE job_slot_rental_listings
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, listingID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete job slot listing")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("job slot listing not found or user is not the owner")
	}

	return nil
}

// GetByID returns a specific listing
func (r *JobSlotRentals) GetByID(ctx context.Context, listingID int64) (*models.JobSlotRentalListing, error) {
	query := `
		SELECT
			l.id,
			l.user_id,
			l.character_id,
			c.name AS character_name,
			l.activity_type,
			l.slots_listed,
			l.price_amount,
			l.pricing_unit,
			l.location_id,
			COALESCE(s.name, '') AS location_name,
			l.notes,
			l.is_active,
			l.created_at,
			l.updated_at
		FROM job_slot_rental_listings l
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		LEFT JOIN stations s ON s.station_id = l.location_id
		WHERE l.id = $1
	`

	var listing models.JobSlotRentalListing
	err := r.db.QueryRowContext(ctx, query, listingID).Scan(
		&listing.ID,
		&listing.UserID,
		&listing.CharacterID,
		&listing.CharacterName,
		&listing.ActivityType,
		&listing.SlotsListed,
		&listing.PriceAmount,
		&listing.PricingUnit,
		&listing.LocationID,
		&listing.LocationName,
		&listing.Notes,
		&listing.IsActive,
		&listing.CreatedAt,
		&listing.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("job slot listing not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get job slot listing")
	}

	return &listing, nil
}

// CreateInterest inserts a new interest request
func (r *JobSlotRentals) CreateInterest(ctx context.Context, interest *models.JobSlotInterestRequest) error {
	query := `
		INSERT INTO job_slot_interest_requests
		(listing_id, requester_user_id, slots_requested, duration_days, message, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		interest.ListingID,
		interest.RequesterUserID,
		interest.SlotsRequested,
		interest.DurationDays,
		interest.Message,
		interest.Status,
	).Scan(&interest.ID, &interest.CreatedAt, &interest.UpdatedAt)

	if err != nil {
		return errors.Wrap(err, "failed to create interest request")
	}

	return nil
}

// GetInterestsByListing returns all interest requests for a listing (seller view)
func (r *JobSlotRentals) GetInterestsByListing(ctx context.Context, listingID int64, sellerUserID int64) ([]*models.JobSlotInterestRequest, error) {
	query := `
		SELECT
			i.id,
			i.listing_id,
			i.requester_user_id,
			u.name AS requester_name,
			i.slots_requested,
			i.duration_days,
			i.message,
			i.status,
			i.created_at,
			i.updated_at
		FROM job_slot_interest_requests i
		JOIN job_slot_rental_listings l ON i.listing_id = l.id
		JOIN users u ON i.requester_user_id = u.id
		WHERE i.listing_id = $1 AND l.user_id = $2
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, listingID, sellerUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query interests by listing")
	}
	defer rows.Close()

	interests := []*models.JobSlotInterestRequest{}
	for rows.Next() {
		var interest models.JobSlotInterestRequest
		err = rows.Scan(
			&interest.ID,
			&interest.ListingID,
			&interest.RequesterUserID,
			&interest.RequesterName,
			&interest.SlotsRequested,
			&interest.DurationDays,
			&interest.Message,
			&interest.Status,
			&interest.CreatedAt,
			&interest.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan interest request")
		}
		interests = append(interests, &interest)
	}

	return interests, nil
}

// GetInterestsByRequester returns all interest requests sent by a user (buyer view)
func (r *JobSlotRentals) GetInterestsByRequester(ctx context.Context, requesterUserID int64) ([]*models.JobSlotInterestRequest, error) {
	query := `
		SELECT
			i.id,
			i.listing_id,
			i.requester_user_id,
			u1.name AS requester_name,
			i.slots_requested,
			i.duration_days,
			i.message,
			i.status,
			i.created_at,
			i.updated_at,
			l.activity_type AS listing_activity_type,
			c.name AS listing_character_name,
			u2.name AS listing_owner_name,
			l.price_amount AS listing_price_amount,
			l.pricing_unit AS listing_pricing_unit
		FROM job_slot_interest_requests i
		JOIN users u1 ON i.requester_user_id = u1.id
		JOIN job_slot_rental_listings l ON i.listing_id = l.id
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		JOIN users u2 ON l.user_id = u2.id
		WHERE i.requester_user_id = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, requesterUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query interests by requester")
	}
	defer rows.Close()

	interests := []*models.JobSlotInterestRequest{}
	for rows.Next() {
		var interest models.JobSlotInterestRequest
		err = rows.Scan(
			&interest.ID,
			&interest.ListingID,
			&interest.RequesterUserID,
			&interest.RequesterName,
			&interest.SlotsRequested,
			&interest.DurationDays,
			&interest.Message,
			&interest.Status,
			&interest.CreatedAt,
			&interest.UpdatedAt,
			&interest.ListingActivityType,
			&interest.ListingCharacterName,
			&interest.ListingOwnerName,
			&interest.ListingPriceAmount,
			&interest.ListingPricingUnit,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan interest request with listing details")
		}
		interests = append(interests, &interest)
	}

	return interests, nil
}

// UpdateInterestStatus updates the status of an interest request
func (r *JobSlotRentals) UpdateInterestStatus(ctx context.Context, interestID int64, userID int64, status string) error {
	// Verify that the user is either the requester (can withdraw) or listing owner (can accept/decline)
	query := `
		UPDATE job_slot_interest_requests i
		SET status = $1, updated_at = NOW()
		FROM job_slot_rental_listings l
		WHERE i.id = $2
			AND i.listing_id = l.id
			AND (i.requester_user_id = $3 OR l.user_id = $3)
	`

	result, err := r.db.ExecContext(ctx, query, status, interestID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to update interest status")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return errors.New("interest request not found or user not authorized")
	}

	return nil
}

// GetInterestByID returns a single interest request enriched with requester and listing details
func (r *JobSlotRentals) GetInterestByID(ctx context.Context, interestID int64) (*models.JobSlotInterestRequest, error) {
	query := `
		SELECT
			i.id,
			i.listing_id,
			i.requester_user_id,
			u1.name AS requester_name,
			i.slots_requested,
			i.duration_days,
			i.message,
			i.status,
			i.created_at,
			i.updated_at,
			l.activity_type AS listing_activity_type,
			c.name AS listing_character_name,
			u2.name AS listing_owner_name,
			l.price_amount AS listing_price_amount,
			l.pricing_unit AS listing_pricing_unit
		FROM job_slot_interest_requests i
		JOIN users u1 ON i.requester_user_id = u1.id
		JOIN job_slot_rental_listings l ON i.listing_id = l.id
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		JOIN users u2 ON l.user_id = u2.id
		WHERE i.id = $1
	`

	var interest models.JobSlotInterestRequest
	err := r.db.QueryRowContext(ctx, query, interestID).Scan(
		&interest.ID,
		&interest.ListingID,
		&interest.RequesterUserID,
		&interest.RequesterName,
		&interest.SlotsRequested,
		&interest.DurationDays,
		&interest.Message,
		&interest.Status,
		&interest.CreatedAt,
		&interest.UpdatedAt,
		&interest.ListingActivityType,
		&interest.ListingCharacterName,
		&interest.ListingOwnerName,
		&interest.ListingPriceAmount,
		&interest.ListingPricingUnit,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("interest request not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get interest request by id")
	}

	return &interest, nil
}

// AcceptInterestWithAgreement atomically accepts an interest request and creates an agreement.
// Only the seller (listing owner) can call this. Returns the created agreement.
func (r *JobSlotRentals) AcceptInterestWithAgreement(ctx context.Context, interestID int64, sellerUserID int64) (*models.JobSlotAgreement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Fetch interest and verify it belongs to sellerUserID's listing
	var interest struct {
		id              int64
		listingID       int64
		requesterUserID int64
		slotsRequested  int
		status          string
	}
	fetchInterestQuery := `
		SELECT i.id, i.listing_id, i.requester_user_id, i.slots_requested, i.status
		FROM job_slot_interest_requests i
		JOIN job_slot_rental_listings l ON i.listing_id = l.id
		WHERE i.id = $1 AND l.user_id = $2
	`
	err = tx.QueryRowContext(ctx, fetchInterestQuery, interestID, sellerUserID).Scan(
		&interest.id,
		&interest.listingID,
		&interest.requesterUserID,
		&interest.slotsRequested,
		&interest.status,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("interest request not found or user not authorized")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch interest request")
	}

	if interest.status != "pending" {
		return nil, errors.New("interest request is not pending")
	}

	// Fetch listing for price/pricing_unit/character_id
	var listing struct {
		priceAmount float64
		pricingUnit string
		characterID int64
	}
	fetchListingQuery := `
		SELECT price_amount, pricing_unit, character_id
		FROM job_slot_rental_listings
		WHERE id = $1
	`
	err = tx.QueryRowContext(ctx, fetchListingQuery, interest.listingID).Scan(
		&listing.priceAmount,
		&listing.pricingUnit,
		&listing.characterID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch listing")
	}

	// Update interest status to accepted
	_, err = tx.ExecContext(ctx,
		`UPDATE job_slot_interest_requests SET status = 'accepted', updated_at = now() WHERE id = $1`,
		interestID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update interest status")
	}

	// Insert agreement
	insertQuery := `
		INSERT INTO job_slot_agreements
		(interest_request_id, listing_id, seller_user_id, renter_user_id, slots_agreed, price_amount, pricing_unit)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, agreed_at, created_at, updated_at
	`
	var agreement models.JobSlotAgreement
	err = tx.QueryRowContext(ctx, insertQuery,
		interestID,
		interest.listingID,
		sellerUserID,
		interest.requesterUserID,
		interest.slotsRequested,
		listing.priceAmount,
		listing.pricingUnit,
	).Scan(&agreement.ID, &agreement.AgreedAt, &agreement.CreatedAt, &agreement.UpdatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert agreement")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit agreement transaction")
	}

	// Fetch the enriched agreement (with user names, character info)
	enriched, err := r.getAgreementByID(ctx, agreement.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch enriched agreement")
	}

	return enriched, nil
}

// getAgreementByID fetches a single agreement with all enriched fields
func (r *JobSlotRentals) getAgreementByID(ctx context.Context, agreementID int64) (*models.JobSlotAgreement, error) {
	query := `
		SELECT
			a.id,
			a.interest_request_id,
			a.listing_id,
			a.seller_user_id,
			seller_u.name AS seller_name,
			a.renter_user_id,
			renter_u.name AS renter_name,
			a.slots_agreed,
			a.price_amount,
			a.pricing_unit,
			a.agreed_at,
			a.expected_end_at,
			a.status,
			a.cancellation_reason,
			l.activity_type,
			l.character_id,
			c.name AS character_name,
			a.created_at,
			a.updated_at
		FROM job_slot_agreements a
		JOIN users seller_u ON a.seller_user_id = seller_u.id
		JOIN users renter_u ON a.renter_user_id = renter_u.id
		JOIN job_slot_rental_listings l ON a.listing_id = l.id
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		WHERE a.id = $1
	`

	var ag models.JobSlotAgreement
	err := r.db.QueryRowContext(ctx, query, agreementID).Scan(
		&ag.ID,
		&ag.InterestRequestID,
		&ag.ListingID,
		&ag.SellerUserID,
		&ag.SellerName,
		&ag.RenterUserID,
		&ag.RenterName,
		&ag.SlotsAgreed,
		&ag.PriceAmount,
		&ag.PricingUnit,
		&ag.AgreedAt,
		&ag.ExpectedEndAt,
		&ag.Status,
		&ag.CancellationReason,
		&ag.ActivityType,
		&ag.CharacterID,
		&ag.CharacterName,
		&ag.CreatedAt,
		&ag.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("agreement not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch agreement")
	}

	return &ag, nil
}

// GetAgreements returns agreements for a user, optionally filtered by status and role.
func (r *JobSlotRentals) GetAgreements(ctx context.Context, userID int64, statusFilter string, role string) ([]*models.JobSlotAgreement, error) {
	whereClause := ""
	args := []interface{}{userID}

	switch role {
	case "seller":
		whereClause = "WHERE a.seller_user_id = $1"
	case "renter":
		whereClause = "WHERE a.renter_user_id = $1"
	default:
		whereClause = "WHERE (a.seller_user_id = $1 OR a.renter_user_id = $1)"
	}

	if statusFilter != "" {
		args = append(args, statusFilter)
		whereClause += " AND a.status = $" + itoa(len(args))
	}

	query := `
		SELECT
			a.id,
			a.interest_request_id,
			a.listing_id,
			a.seller_user_id,
			seller_u.name AS seller_name,
			a.renter_user_id,
			renter_u.name AS renter_name,
			a.slots_agreed,
			a.price_amount,
			a.pricing_unit,
			a.agreed_at,
			a.expected_end_at,
			a.status,
			a.cancellation_reason,
			l.activity_type,
			l.character_id,
			c.name AS character_name,
			a.created_at,
			a.updated_at
		FROM job_slot_agreements a
		JOIN users seller_u ON a.seller_user_id = seller_u.id
		JOIN users renter_u ON a.renter_user_id = renter_u.id
		JOIN job_slot_rental_listings l ON a.listing_id = l.id
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		` + whereClause + `
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query agreements")
	}
	defer rows.Close()

	agreements := []*models.JobSlotAgreement{}
	for rows.Next() {
		var ag models.JobSlotAgreement
		err = rows.Scan(
			&ag.ID,
			&ag.InterestRequestID,
			&ag.ListingID,
			&ag.SellerUserID,
			&ag.SellerName,
			&ag.RenterUserID,
			&ag.RenterName,
			&ag.SlotsAgreed,
			&ag.PriceAmount,
			&ag.PricingUnit,
			&ag.AgreedAt,
			&ag.ExpectedEndAt,
			&ag.Status,
			&ag.CancellationReason,
			&ag.ActivityType,
			&ag.CharacterID,
			&ag.CharacterName,
			&ag.CreatedAt,
			&ag.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan agreement")
		}
		agreements = append(agreements, &ag)
	}

	return agreements, nil
}

// itoa converts an int to its string decimal representation.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// ActiveAgreementCharacter holds character and renter info for an active rental agreement.
type ActiveAgreementCharacter struct {
	CharacterID   int64
	CharacterName string
	RenterUserID  int64
	ActivityType  string
}

// GetActiveAgreementCharacters returns character IDs that have active rental agreements
// along with the renter user ID and activity type for each agreement.
func (r *JobSlotRentals) GetActiveAgreementCharacters(ctx context.Context) ([]*ActiveAgreementCharacter, error) {
	query := `
		SELECT DISTINCT l.character_id, c.name AS character_name, a.renter_user_id, l.activity_type
		FROM job_slot_agreements a
		JOIN job_slot_rental_listings l ON l.id = a.listing_id
		JOIN characters c ON c.id = l.character_id
		WHERE a.status = 'active'
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query active agreement characters")
	}
	defer rows.Close()

	result := []*ActiveAgreementCharacter{}
	for rows.Next() {
		var aac ActiveAgreementCharacter
		if err := rows.Scan(&aac.CharacterID, &aac.CharacterName, &aac.RenterUserID, &aac.ActivityType); err != nil {
			return nil, errors.Wrap(err, "failed to scan active agreement character")
		}
		result = append(result, &aac)
	}

	return result, nil
}

// HasJobBeenNotified checks if a job has already been notified for a character.
func (r *JobSlotRentals) HasJobBeenNotified(ctx context.Context, characterID, jobID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM job_slot_job_notifications WHERE character_id = $1 AND esi_job_id = $2)`,
		characterID, jobID,
	).Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check job notification")
	}
	return exists, nil
}

// MarkJobNotified records that a job has been notified.
func (r *JobSlotRentals) MarkJobNotified(ctx context.Context, characterID, jobID int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO job_slot_job_notifications (character_id, esi_job_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		characterID, jobID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to mark job notified")
	}
	return nil
}

// UpdateAgreementStatus updates the status of an agreement (seller only).
// Only active → completed or cancelled is permitted.
func (r *JobSlotRentals) UpdateAgreementStatus(ctx context.Context, agreementID int64, userID int64, newStatus string, reason *string) error {
	// Fetch existing agreement, verify seller ownership
	var currentStatus string
	var sellerUserID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT status, seller_user_id FROM job_slot_agreements WHERE id = $1`,
		agreementID,
	).Scan(&currentStatus, &sellerUserID)
	if err == sql.ErrNoRows {
		return errors.New("agreement not found")
	}
	if err != nil {
		return errors.Wrap(err, "failed to fetch agreement")
	}

	if sellerUserID != userID {
		return errors.New("not authorized to update this agreement")
	}

	if currentStatus != "active" {
		return errors.New("only active agreements can be updated")
	}

	_, err = r.db.ExecContext(ctx,
		`UPDATE job_slot_agreements
		SET status = $1, cancellation_reason = $2, updated_at = now()
		WHERE id = $3`,
		newStatus, reason, agreementID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update agreement status")
	}

	return nil
}

// GetAgreementJobsByID returns active ESI industry jobs for the character on the listing
// associated with an agreement. Seller-only view.
func (r *JobSlotRentals) GetAgreementJobsByID(ctx context.Context, agreementID int64, userID int64) ([]*models.AgreementJob, error) {
	// Fetch agreement and verify seller ownership
	var sellerUserID int64
	var characterID int64
	err := r.db.QueryRowContext(ctx,
		`SELECT a.seller_user_id, l.character_id
		FROM job_slot_agreements a
		JOIN job_slot_rental_listings l ON a.listing_id = l.id
		WHERE a.id = $1`,
		agreementID,
	).Scan(&sellerUserID, &characterID)
	if err == sql.ErrNoRows {
		return nil, errors.New("agreement not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch agreement")
	}

	if sellerUserID != userID {
		return nil, errors.New("not authorized to view jobs for this agreement")
	}

	query := `
		SELECT
			j.job_id,
			j.activity_id,
			CASE j.activity_id
				WHEN 1 THEN 'Manufacturing'
				WHEN 3 THEN 'TE Research'
				WHEN 4 THEN 'ME Research'
				WHEN 5 THEN 'Copying'
				WHEN 8 THEN 'Invention'
				WHEN 9 THEN 'Reactions'
				ELSE 'Unknown'
			END AS activity_name,
			j.blueprint_type_id,
			COALESCE(bp.type_name, '') AS blueprint_type_name,
			j.product_type_id,
			COALESCE(prod.type_name, '') AS product_type_name,
			j.runs,
			j.start_date,
			j.end_date,
			j.status,
			j.facility_id AS location_id
		FROM esi_industry_jobs j
		JOIN asset_item_types bp ON bp.type_id = j.blueprint_type_id
		LEFT JOIN asset_item_types prod ON prod.type_id = j.product_type_id
		WHERE j.installer_id = $1 AND j.status != 'delivered'
		ORDER BY j.start_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, characterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query agreement jobs")
	}
	defer rows.Close()

	jobs := []*models.AgreementJob{}
	for rows.Next() {
		var job models.AgreementJob
		err = rows.Scan(
			&job.JobID,
			&job.ActivityID,
			&job.ActivityName,
			&job.BlueprintTypeID,
			&job.BlueprintTypeName,
			&job.ProductTypeID,
			&job.ProductTypeName,
			&job.Runs,
			&job.StartDate,
			&job.EndDate,
			&job.Status,
			&job.LocationID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan agreement job")
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetReceivedInterests returns all interests across all of a user's listings
func (r *JobSlotRentals) GetReceivedInterests(ctx context.Context, userID int64) ([]*models.JobSlotInterestRequest, error) {
	query := `
		SELECT
			i.id,
			i.listing_id,
			i.requester_user_id,
			u.name AS requester_name,
			i.slots_requested,
			i.duration_days,
			i.message,
			i.status,
			i.created_at,
			i.updated_at,
			l.activity_type AS listing_activity_type,
			c.name AS listing_character_name
		FROM job_slot_interest_requests i
		JOIN job_slot_rental_listings l ON i.listing_id = l.id
		JOIN users u ON i.requester_user_id = u.id
		JOIN characters c ON l.character_id = c.id AND c.user_id = l.user_id
		WHERE l.user_id = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query received interests")
	}
	defer rows.Close()

	interests := []*models.JobSlotInterestRequest{}
	for rows.Next() {
		var interest models.JobSlotInterestRequest
		err = rows.Scan(
			&interest.ID,
			&interest.ListingID,
			&interest.RequesterUserID,
			&interest.RequesterName,
			&interest.SlotsRequested,
			&interest.DurationDays,
			&interest.Message,
			&interest.Status,
			&interest.CreatedAt,
			&interest.UpdatedAt,
			&interest.ListingActivityType,
			&interest.ListingCharacterName,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan received interest request")
		}
		interests = append(interests, &interest)
	}

	return interests, nil
}
