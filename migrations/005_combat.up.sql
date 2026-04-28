-- 舰船定义表
CREATE TABLE IF NOT EXISTS ship_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    race_id INT NOT NULL,
    tier INT NOT NULL,
    ship_class VARCHAR(50) NOT NULL,
    ship_role VARCHAR(50) NOT NULL,
    -- 基础属性
    shield_hp INT NOT NULL DEFAULT 1000,
    armor_hp INT NOT NULL DEFAULT 800,
    structure_hp INT NOT NULL DEFAULT 500,
    shield_recharge INT NOT NULL DEFAULT 20,
    capacitor INT NOT NULL DEFAULT 500,
    cap_recharge INT NOT NULL DEFAULT 15,
    -- 机动性
    max_speed INT NOT NULL DEFAULT 300,
    align_ticks INT NOT NULL DEFAULT 5,
    signature INT NOT NULL DEFAULT 100,
    -- 槽位
    high_slots INT NOT NULL DEFAULT 3,
    mid_slots INT NOT NULL DEFAULT 3,
    low_slots INT NOT NULL DEFAULT 2,
    -- 容量
    cargo_m3 DOUBLE PRECISION NOT NULL DEFAULT 200,
    drone_bay_m3 DOUBLE PRECISION NOT NULL DEFAULT 0,
    -- 限制
    powergrid INT NOT NULL DEFAULT 100,
    cpu INT NOT NULL DEFAULT 150
);

-- 玩家舰船实例
CREATE TABLE IF NOT EXISTS ships (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    ship_def_id BIGINT NOT NULL REFERENCES ship_defs(id),
    name VARCHAR(100) NOT NULL DEFAULT '',
    -- 当前状态
    shield_current INT NOT NULL,
    armor_current INT NOT NULL,
    structure_current INT NOT NULL,
    cap_current INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    location_system_id BIGINT NOT NULL,
    is_destroyed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 舰船装配的模块
CREATE TABLE IF NOT EXISTS ship_fittings (
    id BIGSERIAL PRIMARY KEY,
    ship_id BIGINT NOT NULL REFERENCES ships(id) ON DELETE CASCADE,
    slot_type VARCHAR(10) NOT NULL, -- high, mid, low
    slot_index INT NOT NULL,
    module_item_id BIGINT NOT NULL REFERENCES item_defs(id),
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    is_overloaded BOOLEAN NOT NULL DEFAULT FALSE
);

-- NPC敌人定义
CREATE TABLE IF NOT EXISTS npc_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    npc_type VARCHAR(50) NOT NULL, -- machine, alien, dimension, guardian
    tier INT NOT NULL DEFAULT 1,
    shield_hp INT NOT NULL DEFAULT 500,
    armor_hp INT NOT NULL DEFAULT 300,
    structure_hp INT NOT NULL DEFAULT 200,
    shield_recharge INT NOT NULL DEFAULT 10,
    damage_per_tick INT NOT NULL DEFAULT 50,
    damage_type VARCHAR(20) NOT NULL DEFAULT 'kinetic',
    optimal_range INT NOT NULL DEFAULT 15000,
    speed INT NOT NULL DEFAULT 200,
    signature INT NOT NULL DEFAULT 80,
    bounty BIGINT NOT NULL DEFAULT 1000,
    ai_behavior VARCHAR(50) NOT NULL DEFAULT 'aggressive'
);

-- NPC掉落表
CREATE TABLE IF NOT EXISTS npc_loot_table (
    id BIGSERIAL PRIMARY KEY,
    npc_def_id BIGINT NOT NULL REFERENCES npc_defs(id),
    item_def_id BIGINT NOT NULL REFERENCES item_defs(id),
    quantity_min INT NOT NULL DEFAULT 1,
    quantity_max INT NOT NULL DEFAULT 1,
    drop_chance DOUBLE PRECISION NOT NULL DEFAULT 0.5
);

-- 战斗实例
CREATE TABLE IF NOT EXISTS combat_instances (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_tick INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE
);

-- 战斗参与者
CREATE TABLE IF NOT EXISTS combat_participants (
    id BIGSERIAL PRIMARY KEY,
    combat_id BIGINT NOT NULL REFERENCES combat_instances(id),
    participant_type VARCHAR(10) NOT NULL, -- player, npc
    character_id BIGINT,
    npc_def_id BIGINT,
    ship_id BIGINT,
    team VARCHAR(10) NOT NULL DEFAULT 'a',
    -- 战斗中的实时状态
    shield_current INT NOT NULL,
    armor_current INT NOT NULL,
    structure_current INT NOT NULL,
    cap_current INT NOT NULL,
    distance INT NOT NULL DEFAULT 20000,
    is_destroyed BOOLEAN NOT NULL DEFAULT FALSE,
    target_id BIGINT
);

-- 战斗日志
CREATE TABLE IF NOT EXISTS combat_logs (
    id BIGSERIAL PRIMARY KEY,
    combat_id BIGINT NOT NULL REFERENCES combat_instances(id),
    tick INT NOT NULL,
    log_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ships_char ON ships(character_id);
CREATE INDEX idx_ship_fittings_ship ON ship_fittings(ship_id);
CREATE INDEX idx_combat_participants ON combat_participants(combat_id);
CREATE INDEX idx_combat_logs_combat ON combat_logs(combat_id, tick);
