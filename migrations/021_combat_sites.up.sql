-- ============ 战斗地点系统 ============

CREATE TABLE IF NOT EXISTS combat_sites (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL,
    dungeon_def_id BIGINT NOT NULL REFERENCES dungeon_defs(id),
    site_type VARCHAR(30) NOT NULL, -- small_anomaly, medium_anomaly, large_anomaly, signal, expedition
    name VARCHAR(150) NOT NULL,
    difficulty INT NOT NULL DEFAULT 1,
    is_scanned BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, in_progress, completed, cooldown
    occupied_by BIGINT DEFAULT 0,
    spawned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    cooldown_until TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_combat_sites_system ON combat_sites(system_id, status);
CREATE INDEX idx_combat_sites_status ON combat_sites(status);

-- ============ 按难度缩放的NPC模板 ============
-- 为每个难度级别创建基础NPC（不依赖已有的T1-T5 NPC）

INSERT INTO npc_defs (id, name, npc_type, tier, shield_hp, armor_hp, structure_hp, shield_recharge,
    damage_per_tick, damage_type, optimal_range, speed, signature, bounty, ai_behavior,
    shield_res_kinetic, shield_res_thermal, shield_res_em, shield_res_explosive,
    armor_res_kinetic, armor_res_thermal, armor_res_em, armor_res_explosive)
VALUES
-- T1 小型异常用
(101, '巡逻无人机', 'machine', 1, 200, 150, 100, 5, 15, 'kinetic', 8000, 200, 40, 300, 'aggressive',
    0.15, 0.15, 0.05, 0.10, 0.25, 0.15, 0.05, 0.15),
(102, '哨戒浮游体', 'alien', 1, 150, 100, 120, 4, 12, 'thermal', 6000, 150, 35, 200, 'passive',
    0.10, 0.30, 0.15, 0.05, 0.10, 0.30, 0.10, 0.10),
-- T2 小/中型异常用
(103, '强化哨兵', 'machine', 2, 500, 400, 300, 12, 35, 'kinetic', 12000, 180, 60, 1000, 'aggressive',
    0.20, 0.15, 0.05, 0.15, 0.30, 0.15, 0.05, 0.20),
(104, '猎食水母群', 'alien', 2, 400, 300, 350, 10, 28, 'em', 8000, 250, 50, 800, 'aggressive',
    0.10, 0.25, 0.30, 0.10, 0.10, 0.25, 0.20, 0.10),
-- T3 中型异常用
(105, '歼击编队长', 'machine', 3, 1200, 900, 600, 22, 70, 'thermal', 15000, 160, 90, 4000, 'aggressive',
    0.20, 0.25, 0.10, 0.15, 0.30, 0.25, 0.10, 0.20),
(106, '维度渗透者', 'dimension', 3, 800, 600, 700, 18, 55, 'em', 18000, 300, 45, 3500, 'evasive',
    0.15, 0.10, 0.35, 0.30, 0.10, 0.05, 0.30, 0.35),
-- T4 中/大型异常+信号用
(107, '重型堡垒炮台', 'machine', 4, 3000, 2500, 1500, 40, 120, 'explosive', 20000, 80, 200, 12000, 'defensive',
    0.25, 0.20, 0.10, 0.20, 0.35, 0.20, 0.10, 0.30),
-- T5 大型异常+信号用
(108, '精英指挥舰', 'machine', 5, 6000, 5000, 3000, 60, 200, 'thermal', 25000, 120, 300, 40000, 'aggressive',
    0.25, 0.30, 0.15, 0.15, 0.35, 0.30, 0.10, 0.25),
-- Boss模板(各难度共用，战斗时按boss_hp_override缩放)
(199, '远征Boss模板', 'machine', 5, 10000, 8000, 5000, 80, 150, 'kinetic', 20000, 100, 250, 50000, 'aggressive',
    0.25, 0.25, 0.15, 0.15, 0.30, 0.25, 0.15, 0.25)
ON CONFLICT (id) DO NOTHING;

-- 为NPC掉落表添加新NPC的掉落
INSERT INTO npc_loot_table (npc_def_id, item_def_id, quantity_min, quantity_max, drop_chance) VALUES
(101, 1001, 5, 20, 0.9),   -- 巡逻无人机掉铁陨石
(102, 1101, 5, 15, 0.8),   -- 浮游体掉氢雾
(103, 1002, 10, 40, 0.8),  -- 强化哨兵掉铜辉矿
(104, 1103, 8, 25, 0.7),   -- 猎食水母掉碳氢星云
(105, 1003, 15, 50, 0.7),  -- 歼击编队掉钛铁矿
(106, 1005, 8, 25, 0.5),   -- 维度渗透者掉钨锰矿
(107, 1006, 10, 30, 0.6),  -- 重型堡垒掉铌钽铁矿
(108, 1008, 10, 40, 0.7)   -- 精英指挥掉铂族砂矿
ON CONFLICT DO NOTHING;
