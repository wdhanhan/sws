CREATE TABLE IF NOT EXISTS fleets (
    id BIGSERIAL PRIMARY KEY,
    leader_id BIGINT NOT NULL REFERENCES characters(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_members (
    id BIGSERIAL PRIMARY KEY,
    fleet_id BIGINT NOT NULL REFERENCES fleets(id) ON DELETE CASCADE,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(fleet_id, character_id)
);

CREATE INDEX IF NOT EXISTS idx_fleet_members_char ON fleet_members(character_id, status);
CREATE INDEX IF NOT EXISTS idx_fleets_leader ON fleets(leader_id, status);
