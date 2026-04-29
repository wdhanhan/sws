-- ============================================================
-- 024: 残骸系统 + 星系扩展到 10000
-- ============================================================

-- 残骸表
CREATE TABLE IF NOT EXISTS wrecks (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL,
    owner_name VARCHAR(100) NOT NULL DEFAULT '',
    ship_def_id BIGINT,
    ship_name VARCHAR(100) DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ DEFAULT NOW() + interval '2 hours',
    is_looted BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS wreck_items (
    id BIGSERIAL PRIMARY KEY,
    wreck_id BIGINT NOT NULL REFERENCES wrecks(id) ON DELETE CASCADE,
    item_def_id BIGINT NOT NULL REFERENCES item_defs(id),
    quantity INT NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_wrecks_system ON wrecks(system_id, is_looted);
CREATE INDEX IF NOT EXISTS idx_wreck_items_wreck ON wreck_items(wreck_id);

-- ============================================================
-- 星系扩展: 从 1000 扩展到 10000
-- 7个旋臂, 安等从 -1.0 到 1.0
-- 内圈高安(0.5~1.0), 中圈低安(0.1~0.5), 外圈零安(0~-0.5), 深渊(-0.5~-1.0)
-- ============================================================

-- 中文星系名前缀和后缀
INSERT INTO star_systems (name, arm_id, coord_x, coord_y, coord_z, security_level, star_type, planet_count, belt_count, seed)
SELECT
  -- 生成星系名: 旋臂名 + 编号
  CASE (s % 7) + 1
    WHEN 1 THEN '焚天'
    WHEN 2 THEN '厚土'
    WHEN 3 THEN '罡风'
    WHEN 4 THEN '渊水'
    WHEN 5 THEN '核心'
    WHEN 6 THEN '虚空'
    WHEN 7 THEN '外缘'
  END || '-' ||
  CASE (s % 20)
    WHEN 0 THEN '星渊' WHEN 1 THEN '天枢' WHEN 2 THEN '瑶光' WHEN 3 THEN '开阳'
    WHEN 4 THEN '玉衡' WHEN 5 THEN '天权' WHEN 6 THEN '天玑' WHEN 7 THEN '摇光'
    WHEN 8 THEN '紫微' WHEN 9 THEN '太乙' WHEN 10 THEN '天罡' WHEN 11 THEN '地煞'
    WHEN 12 THEN '破军' WHEN 13 THEN '贪狼' WHEN 14 THEN '巨门' WHEN 15 THEN '禄存'
    WHEN 16 THEN '文曲' WHEN 17 THEN '廉贞' WHEN 18 THEN '武曲' WHEN 19 THEN '左辅'
  END || ' ' || (s / 20 + 1)::text,
  -- arm_id: 1-7
  (s % 7) + 1,
  -- coordinates: spiral arm pattern
  (COS(s * 0.1 + (s % 7) * 0.9) * (50 + s * 0.02))::int,
  (SIN(s * 0.1 + (s % 7) * 0.9) * (50 + s * 0.02))::int,
  ((s % 50) - 25),
  -- security_level: based on distance from center
  ROUND(CAST(
    CASE
      WHEN s < 2250 THEN 0.5 + RANDOM() * 0.5          -- 高安 0.5~1.0
      WHEN s < 4500 THEN 0.1 + RANDOM() * 0.4           -- 低安 0.1~0.5
      WHEN s < 6750 THEN -0.5 + RANDOM() * 0.5          -- 零安 -0.5~0.0
      ELSE -1.0 + RANDOM() * 0.5                         -- 深渊 -1.0~-0.5
    END AS NUMERIC), 3),
  -- star_type
  (ARRAY['O','B','A','F','G','K','M'])[((s % 7) + 1)],
  -- planet_count: 1-12
  (s % 12) + 1,
  -- belt_count: 0-8
  (s % 9),
  -- seed
  s * 17 + 42
FROM generate_series(0, 8999) AS s
WHERE NOT EXISTS (SELECT 1 FROM star_systems WHERE id = 1001 + s);

-- Update sequence
SELECT setval('star_systems_id_seq', GREATEST((SELECT MAX(id) FROM star_systems), 10000));
