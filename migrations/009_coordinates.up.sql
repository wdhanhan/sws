-- ============ 星系内坐标系 ============
-- 单位：米(m)，恒星在原点(0,0,0)
-- 1 AU ≈ 1.496e11 m，简化为 1AU = 150,000,000 km = 1.5e11 m

-- 行星加入3D坐标
ALTER TABLE planets ADD COLUMN pos_x DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE planets ADD COLUMN pos_y DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE planets ADD COLUMN pos_z DOUBLE PRECISION NOT NULL DEFAULT 0;

-- 矿带加入3D坐标
ALTER TABLE asteroid_belts ADD COLUMN pos_x DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE asteroid_belts ADD COLUMN pos_y DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE asteroid_belts ADD COLUMN pos_z DOUBLE PRECISION NOT NULL DEFAULT 0;

-- 星门加入本星系内的3D坐标
ALTER TABLE stargates ADD COLUMN pos_x DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE stargates ADD COLUMN pos_y DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE stargates ADD COLUMN pos_z DOUBLE PRECISION NOT NULL DEFAULT 0;

-- 角色/舰船加入星系内3D坐标
ALTER TABLE characters ADD COLUMN pos_x DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN pos_y DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE characters ADD COLUMN pos_z DOUBLE PRECISION NOT NULL DEFAULT 0;

-- 空间站表(新建)
CREATE TABLE IF NOT EXISTS stations (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL REFERENCES star_systems(id),
    planet_id BIGINT REFERENCES planets(id),
    name VARCHAR(100) NOT NULL,
    owner_type VARCHAR(20) NOT NULL DEFAULT 'npc',  -- npc / corp / alliance / nation
    owner_id BIGINT,
    pos_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    pos_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    pos_z DOUBLE PRECISION NOT NULL DEFAULT 0,
    has_market BOOLEAN NOT NULL DEFAULT TRUE,
    has_refinery BOOLEAN NOT NULL DEFAULT TRUE,
    has_factory BOOLEAN NOT NULL DEFAULT FALSE,
    has_clone_bay BOOLEAN NOT NULL DEFAULT TRUE,
    has_repair BOOLEAN NOT NULL DEFAULT TRUE,
    docking_fee BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stations_system ON stations(system_id);

-- 根据orbit_au生成行星的3D坐标（环绕恒星的不同角度）
-- 用伪随机角度分配
DO $$
DECLARE
    r RECORD;
    angle DOUBLE PRECISION;
    dist DOUBLE PRECISION;
BEGIN
    FOR r IN SELECT id, orbit_au, system_id FROM planets ORDER BY system_id, orbit_au
    LOOP
        -- 每颗行星在不同角度，轻微z偏移
        angle := (r.id * 137.508) * PI() / 180.0;  -- 黄金角分布
        dist := r.orbit_au * 1.496e11;  -- AU转米
        UPDATE planets SET
            pos_x = dist * COS(angle),
            pos_y = dist * SIN(angle),
            pos_z = (RANDOM() - 0.5) * dist * 0.05  -- 轻微偏离黄道面
        WHERE id = r.id;
    END LOOP;
END $$;

-- 矿带坐标：在对应轨道上随机分布
DO $$
DECLARE
    r RECORD;
    angle DOUBLE PRECISION;
    dist DOUBLE PRECISION;
BEGIN
    FOR r IN SELECT id, orbit_au, system_id FROM asteroid_belts ORDER BY system_id, orbit_au
    LOOP
        angle := (r.id * 222.5) * PI() / 180.0;
        dist := r.orbit_au * 1.496e11;
        UPDATE asteroid_belts SET
            pos_x = dist * COS(angle),
            pos_y = dist * SIN(angle),
            pos_z = (RANDOM() - 0.5) * dist * 0.02
        WHERE id = r.id;
    END LOOP;
END $$;

-- 星门坐标：放在星系外围(约50AU处)
DO $$
DECLARE
    r RECORD;
    angle DOUBLE PRECISION;
    dist DOUBLE PRECISION;
BEGIN
    FOR r IN SELECT id, from_system_id FROM stargates ORDER BY id
    LOOP
        angle := (r.id * 97.3) * PI() / 180.0;
        dist := 50.0 * 1.496e11;
        UPDATE stargates SET
            pos_x = dist * COS(angle),
            pos_y = dist * SIN(angle),
            pos_z = (RANDOM() - 0.5) * 1e10
        WHERE id = r.id;
    END LOOP;
END $$;

-- 在有空间站的行星旁边生成NPC空间站
INSERT INTO stations (system_id, planet_id, name, owner_type, pos_x, pos_y, pos_z)
SELECT
    p.system_id,
    p.id,
    CONCAT(ss.name, ' - ', p.name, ' 空间站'),
    'npc',
    p.pos_x + 50000,  -- 行星旁边50km
    p.pos_y + 50000,
    p.pos_z
FROM planets p
JOIN star_systems ss ON ss.id = p.system_id
WHERE p.has_station = TRUE;

-- 在每个种族起源星系确保有空间站
INSERT INTO stations (system_id, name, owner_type, pos_x, pos_y, pos_z, has_market, has_refinery, has_factory, has_clone_bay, has_repair)
SELECT
    id,
    CONCAT(name, ' - 起源空间站'),
    'npc',
    1e11, 0, 0,
    TRUE, TRUE, TRUE, TRUE, TRUE
FROM star_systems
WHERE id IN (1001, 1002, 1003, 2001, 2002, 2003, 3001, 3002, 3003, 4001, 4002, 4003)
ON CONFLICT DO NOTHING;
