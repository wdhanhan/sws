-- ============ 远征副本系统 ============

CREATE TABLE IF NOT EXISTS dungeon_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    race_theme INT NOT NULL,
    difficulty INT NOT NULL,
    wave_count INT NOT NULL DEFAULT 3,
    min_security DOUBLE PRECISION NOT NULL DEFAULT -1,
    reward_credits BIGINT NOT NULL DEFAULT 1000,
    encounter_id BIGINT DEFAULT 0,
    key_item_id BIGINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS dungeon_waves (
    id BIGSERIAL PRIMARY KEY,
    dungeon_id BIGINT NOT NULL REFERENCES dungeon_defs(id),
    wave_number INT NOT NULL,
    npc_def_id BIGINT NOT NULL,
    npc_count INT NOT NULL DEFAULT 1,
    is_boss BOOLEAN NOT NULL DEFAULT FALSE,
    boss_name VARCHAR(100) DEFAULT '',
    boss_hp_override INT DEFAULT 0,
    wave_text TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS dungeon_instances (
    id BIGSERIAL PRIMARY KEY,
    dungeon_def_id BIGINT NOT NULL REFERENCES dungeon_defs(id),
    character_id BIGINT NOT NULL REFERENCES characters(id),
    current_wave INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    total_kills INT NOT NULL DEFAULT 0,
    total_damage_dealt BIGINT NOT NULL DEFAULT 0,
    loot_collected JSONB NOT NULL DEFAULT '[]',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_dungeon_defs_race ON dungeon_defs(race_theme, difficulty);
CREATE INDEX idx_dungeon_waves_dungeon ON dungeon_waves(dungeon_id, wave_number);
CREATE INDEX idx_dungeon_instances_char ON dungeon_instances(character_id, status);
