-- ============ 白羊座舰船（T1-T3示例）============
INSERT INTO ship_defs (id, name, race_id, tier, ship_class, ship_role, shield_hp, armor_hp, structure_hp, shield_recharge, capacitor, cap_recharge, max_speed, align_ticks, signature, high_slots, mid_slots, low_slots, cargo_m3, powergrid, cpu) VALUES
(1, '锐矛', 1, 1, '护卫舰', '突击', 800, 500, 400, 25, 400, 12, 450, 3, 40, 3, 2, 2, 150, 60, 120),
(2, '怒目', 1, 1, '护卫舰', '电子', 600, 400, 350, 20, 500, 15, 380, 3, 35, 2, 3, 2, 120, 50, 160),
(3, '焚城', 1, 2, '驱逐舰', '火力', 1500, 1000, 700, 30, 600, 15, 350, 4, 80, 5, 3, 2, 300, 90, 180),
(4, '战嚎', 1, 3, '巡洋舰', '主战', 4000, 3000, 2000, 50, 1200, 25, 250, 5, 150, 4, 4, 3, 450, 150, 250);

-- ============ 金牛座舰船 ============
INSERT INTO ship_defs (id, name, race_id, tier, ship_class, ship_role, shield_hp, armor_hp, structure_hp, shield_recharge, capacitor, cap_recharge, max_speed, align_ticks, signature, high_slots, mid_slots, low_slots, cargo_m3, powergrid, cpu) VALUES
(5, '铁蹄', 2, 1, '护卫舰', '突击', 700, 700, 500, 20, 350, 10, 380, 4, 45, 3, 2, 3, 200, 55, 110),
(6, '犁刃', 2, 1, '护卫舰', '工业', 500, 600, 400, 15, 300, 10, 300, 5, 50, 2, 2, 2, 500, 40, 100),
(7, '弥诺', 2, 3, '巡洋舰', '主战', 3500, 4500, 2500, 40, 1000, 20, 220, 6, 180, 4, 3, 4, 600, 180, 220);

-- ============ 双子座舰船 ============
INSERT INTO ship_defs (id, name, race_id, tier, ship_class, ship_role, shield_hp, armor_hp, structure_hp, shield_recharge, capacitor, cap_recharge, max_speed, align_ticks, signature, high_slots, mid_slots, low_slots, cargo_m3, powergrid, cpu) VALUES
(8, '影刺', 3, 1, '护卫舰', '突击', 600, 400, 300, 20, 450, 14, 500, 2, 30, 3, 3, 2, 130, 50, 150),
(9, '虚影', 3, 3, '巡洋舰', '隐形', 3000, 2000, 1500, 35, 1400, 30, 280, 4, 100, 3, 5, 3, 350, 120, 300);

-- ============ NPC敌人 ============
-- A类: 失控机器
INSERT INTO npc_defs (id, name, npc_type, tier, shield_hp, armor_hp, structure_hp, shield_recharge, damage_per_tick, damage_type, optimal_range, speed, signature, bounty, ai_behavior) VALUES
(1, '失控清扫者', 'machine', 1, 300, 200, 150, 8, 30, 'kinetic', 10000, 250, 50, 500, 'aggressive'),
(2, '游荡哨兵', 'machine', 2, 800, 600, 400, 15, 60, 'kinetic', 15000, 200, 80, 2000, 'aggressive'),
(3, '歼灭者', 'machine', 3, 2000, 1500, 1000, 25, 120, 'thermal', 20000, 180, 120, 8000, 'aggressive'),
(4, '自走堡垒', 'machine', 4, 5000, 4000, 3000, 40, 200, 'thermal', 25000, 100, 250, 25000, 'defensive'),
(5, '毁灭核心', 'machine', 5, 15000, 10000, 8000, 60, 400, 'explosive', 30000, 80, 400, 100000, 'aggressive');

-- B类: 异星生物
INSERT INTO npc_defs (id, name, npc_type, tier, shield_hp, armor_hp, structure_hp, shield_recharge, damage_per_tick, damage_type, optimal_range, speed, signature, bounty, ai_behavior) VALUES
(10, '真空水母', 'alien', 1, 200, 100, 300, 5, 20, 'em', 5000, 150, 60, 300, 'passive'),
(11, '星际鲨', 'alien', 2, 500, 800, 600, 10, 80, 'kinetic', 8000, 400, 70, 3000, 'aggressive'),
(12, '虫巢体', 'alien', 3, 1000, 2000, 1500, 15, 50, 'thermal', 12000, 120, 200, 10000, 'spawner');

-- C类: 维度入侵者
INSERT INTO npc_defs (id, name, npc_type, tier, shield_hp, armor_hp, structure_hp, shield_recharge, damage_per_tick, damage_type, optimal_range, speed, signature, bounty, ai_behavior) VALUES
(20, '相位幽灵', 'dimension', 2, 600, 300, 500, 20, 70, 'em', 18000, 350, 40, 5000, 'evasive'),
(21, '空间折叠兽', 'dimension', 3, 1500, 1000, 800, 25, 100, 'explosive', 15000, 250, 100, 12000, 'aggressive');

-- NPC掉落表
INSERT INTO npc_loot_table (npc_def_id, item_def_id, quantity_min, quantity_max, drop_chance) VALUES
(1, 1001, 10, 50, 0.8),    -- 清扫者掉铁陨石
(2, 1002, 20, 80, 0.7),    -- 哨兵掉铜辉矿
(2, 5002, 1, 1, 0.1),      -- 哨兵小概率掉武器
(3, 1003, 30, 100, 0.6),   -- 歼灭者掉钛铁矿
(3, 5006, 1, 1, 0.15),     -- 歼灭者掉磁轨炮
(4, 1006, 20, 60, 0.5),    -- 堡垒掉铌钽铁矿
(4, 5007, 1, 1, 0.2),      -- 堡垒掉导弹
(5, 1008, 30, 100, 0.7),   -- 核心掉铂族砂矿
(10, 1101, 20, 50, 0.9),   -- 水母掉氢雾
(11, 1103, 10, 40, 0.6),   -- 星际鲨掉碳氢星云
(20, 1005, 10, 30, 0.5),   -- 幽灵掉钨锰矿
(21, 1009, 5, 15, 0.3);    -- 折叠兽掉铪锆英石
