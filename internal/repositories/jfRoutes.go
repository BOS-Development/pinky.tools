package repositories

import (
	"context"
	"database/sql"
	"math"

	"github.com/annymsMthd/industry-tool/internal/models"
	"github.com/pkg/errors"
)

type JFRoutes struct {
	db *sql.DB
}

func NewJFRoutes(db *sql.DB) *JFRoutes {
	return &JFRoutes{db: db}
}

// CalculateDistanceLY calculates the light-year distance between two solar systems
// from their SDE coordinates (in meters).
func CalculateDistanceLY(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	distanceMeters := math.Sqrt(dx*dx + dy*dy + dz*dz)
	return distanceMeters / 9.461e+15
}

func (r *JFRoutes) GetByUser(ctx context.Context, userID int64) ([]*models.JFRoute, error) {
	query := `
		select r.id, r.user_id, r.name, r.origin_system_id, r.destination_system_id,
		       r.total_distance_ly, r.created_at,
		       COALESCE(os.name, ''), COALESCE(ds.name, '')
		from jf_routes r
		left join solar_systems os on os.solar_system_id = r.origin_system_id
		left join solar_systems ds on ds.solar_system_id = r.destination_system_id
		where r.user_id = $1
		order by r.created_at desc
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query JF routes")
	}
	defer rows.Close()

	routes := []*models.JFRoute{}
	for rows.Next() {
		var route models.JFRoute
		if err := rows.Scan(
			&route.ID, &route.UserID, &route.Name,
			&route.OriginSystemID, &route.DestinationSystemID,
			&route.TotalDistanceLY, &route.CreatedAt,
			&route.OriginSystemName, &route.DestinationSystemName,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan JF route")
		}
		route.Waypoints = []*models.JFRouteWaypoint{}
		routes = append(routes, &route)
	}

	// Load waypoints for all routes
	for _, route := range routes {
		waypoints, err := r.getWaypoints(ctx, route.ID)
		if err != nil {
			return nil, err
		}
		route.Waypoints = waypoints
	}

	return routes, nil
}

func (r *JFRoutes) GetByID(ctx context.Context, id, userID int64) (*models.JFRoute, error) {
	query := `
		select r.id, r.user_id, r.name, r.origin_system_id, r.destination_system_id,
		       r.total_distance_ly, r.created_at,
		       COALESCE(os.name, ''), COALESCE(ds.name, '')
		from jf_routes r
		left join solar_systems os on os.solar_system_id = r.origin_system_id
		left join solar_systems ds on ds.solar_system_id = r.destination_system_id
		where r.id = $1 and r.user_id = $2
	`

	var route models.JFRoute
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&route.ID, &route.UserID, &route.Name,
		&route.OriginSystemID, &route.DestinationSystemID,
		&route.TotalDistanceLY, &route.CreatedAt,
		&route.OriginSystemName, &route.DestinationSystemName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get JF route by ID")
	}

	waypoints, err := r.getWaypoints(ctx, route.ID)
	if err != nil {
		return nil, err
	}
	route.Waypoints = waypoints

	return &route, nil
}

func (r *JFRoutes) Create(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Calculate waypoint distances from coordinates
	totalDistance := 0.0
	for i, wp := range route.Waypoints {
		if i == 0 {
			wp.DistanceLY = 0 // First waypoint is the origin
		} else {
			prevWP := route.Waypoints[i-1]
			prevSys := systemCoords[prevWP.SystemID]
			curSys := systemCoords[wp.SystemID]
			if prevSys != nil && curSys != nil && prevSys.X != nil && curSys.X != nil {
				wp.DistanceLY = CalculateDistanceLY(*prevSys.X, *prevSys.Y, *prevSys.Z, *curSys.X, *curSys.Y, *curSys.Z)
			}
		}
		totalDistance += wp.DistanceLY
	}

	// Insert route
	var created models.JFRoute
	err = tx.QueryRowContext(ctx, `
		insert into jf_routes (user_id, name, origin_system_id, destination_system_id, total_distance_ly)
		values ($1, $2, $3, $4, $5)
		returning id, user_id, name, origin_system_id, destination_system_id, total_distance_ly, created_at
	`, route.UserID, route.Name, route.OriginSystemID, route.DestinationSystemID, totalDistance,
	).Scan(
		&created.ID, &created.UserID, &created.Name,
		&created.OriginSystemID, &created.DestinationSystemID,
		&created.TotalDistanceLY, &created.CreatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert JF route")
	}

	// Insert waypoints
	created.Waypoints = []*models.JFRouteWaypoint{}
	for _, wp := range route.Waypoints {
		var createdWP models.JFRouteWaypoint
		err = tx.QueryRowContext(ctx, `
			insert into jf_route_waypoints (route_id, sequence, system_id, distance_ly)
			values ($1, $2, $3, $4)
			returning id, route_id, sequence, system_id, distance_ly
		`, created.ID, wp.Sequence, wp.SystemID, wp.DistanceLY,
		).Scan(
			&createdWP.ID, &createdWP.RouteID, &createdWP.Sequence,
			&createdWP.SystemID, &createdWP.DistanceLY,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert JF route waypoint")
		}
		created.Waypoints = append(created.Waypoints, &createdWP)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit JF route")
	}

	return &created, nil
}

func (r *JFRoutes) Update(ctx context.Context, route *models.JFRoute, systemCoords map[int64]*models.SolarSystem) (*models.JFRoute, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	// Calculate waypoint distances from coordinates
	totalDistance := 0.0
	for i, wp := range route.Waypoints {
		if i == 0 {
			wp.DistanceLY = 0
		} else {
			prevWP := route.Waypoints[i-1]
			prevSys := systemCoords[prevWP.SystemID]
			curSys := systemCoords[wp.SystemID]
			if prevSys != nil && curSys != nil && prevSys.X != nil && curSys.X != nil {
				wp.DistanceLY = CalculateDistanceLY(*prevSys.X, *prevSys.Y, *prevSys.Z, *curSys.X, *curSys.Y, *curSys.Z)
			}
		}
		totalDistance += wp.DistanceLY
	}

	// Update route
	var updated models.JFRoute
	err = tx.QueryRowContext(ctx, `
		update jf_routes
		set name = $3, origin_system_id = $4, destination_system_id = $5, total_distance_ly = $6
		where id = $1 and user_id = $2
		returning id, user_id, name, origin_system_id, destination_system_id, total_distance_ly, created_at
	`, route.ID, route.UserID, route.Name, route.OriginSystemID, route.DestinationSystemID, totalDistance,
	).Scan(
		&updated.ID, &updated.UserID, &updated.Name,
		&updated.OriginSystemID, &updated.DestinationSystemID,
		&updated.TotalDistanceLY, &updated.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update JF route")
	}

	// Delete old waypoints and insert new ones
	_, err = tx.ExecContext(ctx, `delete from jf_route_waypoints where route_id = $1`, route.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete old waypoints")
	}

	updated.Waypoints = []*models.JFRouteWaypoint{}
	for _, wp := range route.Waypoints {
		var createdWP models.JFRouteWaypoint
		err = tx.QueryRowContext(ctx, `
			insert into jf_route_waypoints (route_id, sequence, system_id, distance_ly)
			values ($1, $2, $3, $4)
			returning id, route_id, sequence, system_id, distance_ly
		`, updated.ID, wp.Sequence, wp.SystemID, wp.DistanceLY,
		).Scan(
			&createdWP.ID, &createdWP.RouteID, &createdWP.Sequence,
			&createdWP.SystemID, &createdWP.DistanceLY,
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert JF route waypoint")
		}
		updated.Waypoints = append(updated.Waypoints, &createdWP)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "failed to commit JF route update")
	}

	return &updated, nil
}

func (r *JFRoutes) Delete(ctx context.Context, id, userID int64) error {
	result, err := r.db.ExecContext(ctx, `
		delete from jf_routes where id = $1 and user_id = $2
	`, id, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete JF route")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for delete")
	}
	if rows == 0 {
		return errors.New("JF route not found")
	}

	return nil
}

func (r *JFRoutes) getWaypoints(ctx context.Context, routeID int64) ([]*models.JFRouteWaypoint, error) {
	query := `
		select w.id, w.route_id, w.sequence, w.system_id, w.distance_ly,
		       COALESCE(ss.name, '')
		from jf_route_waypoints w
		left join solar_systems ss on ss.solar_system_id = w.system_id
		where w.route_id = $1
		order by w.sequence asc
	`

	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query JF route waypoints")
	}
	defer rows.Close()

	waypoints := []*models.JFRouteWaypoint{}
	for rows.Next() {
		var wp models.JFRouteWaypoint
		if err := rows.Scan(
			&wp.ID, &wp.RouteID, &wp.Sequence, &wp.SystemID, &wp.DistanceLY,
			&wp.SystemName,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan JF route waypoint")
		}
		waypoints = append(waypoints, &wp)
	}

	return waypoints, nil
}
