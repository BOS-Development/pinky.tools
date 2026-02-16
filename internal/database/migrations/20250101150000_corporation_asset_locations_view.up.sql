BEGIN;

-- View to resolve corporation asset locations with container recursion
-- Handles: direct station placement, container in station, container in office in station
CREATE OR REPLACE VIEW corporation_asset_locations AS
SELECT
    ca.corporation_id,
    ca.user_id,
    ca.item_id,
    ca.type_id,
    ca.location_id,
    ca.location_type,
    ca.location_flag,
    -- Container information (if asset is in a container)
    containers.item_id as container_id,
    containers.type_id as container_type_id,
    containers.location_flag as container_location_flag,
    containers.location_type as container_location_type,
    -- Division number extracted from container or direct location
    CASE
        WHEN ca.location_type = 'station' AND ca.location_flag LIKE 'CorpSAG%'
            THEN SUBSTRING(ca.location_flag, 8, 1)::int
        WHEN containers.location_flag LIKE 'CorpSAG%'
            THEN SUBSTRING(containers.location_flag, 8, 1)::int
        ELSE NULL
    END as division_number,
    -- Resolved station_id (handling all recursion scenarios)
    CASE
        -- Direct station placement
        WHEN ca.location_type = 'station' THEN ca.location_id
        -- Container at station
        WHEN containers.location_type = 'station' THEN containers.location_id
        -- Container in office/folder at station
        WHEN containers.location_type = 'item' THEN container_location.location_id
        -- Fallback to container's location_id
        ELSE containers.location_id
    END as station_id,
    -- Location details
    stations.name as station_name,
    stations.corporation_id as station_corporation_id,
    stations.is_npc_station,
    systems.solar_system_id,
    systems.name as solar_system_name,
    systems.security,
    constellations.constellation_id,
    constellations.name as constellation_name,
    regions.region_id,
    regions.name as region_name
FROM corporation_assets ca
-- Join to container (if asset is inside a container)
LEFT JOIN corporation_assets containers ON (
    ca.location_type = 'item'
    AND containers.item_id = ca.location_id
    AND containers.corporation_id = ca.corporation_id
    AND containers.user_id = ca.user_id
)
-- Join to container's location (if container is inside an office/folder)
LEFT JOIN corporation_assets container_location ON (
    containers.location_type = 'item'
    AND container_location.item_id = containers.location_id
    AND container_location.corporation_id = ca.corporation_id
    AND container_location.user_id = ca.user_id
)
-- Join to station using resolved station_id
LEFT JOIN stations ON stations.station_id = (
    CASE
        WHEN ca.location_type = 'station' THEN ca.location_id
        WHEN containers.location_type = 'station' THEN containers.location_id
        WHEN containers.location_type = 'item' THEN container_location.location_id
        ELSE containers.location_id
    END
)
-- Join location hierarchy
LEFT JOIN solar_systems systems ON systems.solar_system_id = stations.solar_system_id
LEFT JOIN constellations ON constellations.constellation_id = systems.constellation_id
LEFT JOIN regions ON regions.region_id = constellations.region_id;

COMMIT;
