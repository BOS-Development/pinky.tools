BEGIN;

DROP TABLE IF EXISTS corporation_hanger_divisions;

CREATE TABLE corporation_divisions (
    corporation_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    division_number INT NOT NULL,
    division_type VARCHAR(20) NOT NULL,
    name VARCHAR(500) NOT NULL,
    PRIMARY KEY (corporation_id, user_id, division_number, division_type),
    FOREIGN KEY (corporation_id, user_id) REFERENCES player_corporations(id, user_id)
);

COMMIT;
