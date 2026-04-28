-- ============ 重建完整模块体系 ============
-- 设计原则:
-- 小型PG 5-15, CPU 10-30 (护卫/驱逐用)
-- 中型PG 20-50, CPU 30-60 (巡洋用)
-- 大型PG 80-200, CPU 50-100 (战列用)
-- 每种舰船的PG/CPU刚好能装满对应尺寸的模块(装满需要取舍)

-- ======== 修复: 重建被删除的小型能量武器 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type)
SELECT 5001, '小型脉冲激光 I', 'module', '短程能量武器，热能伤害，高追踪', 5, 8000, false, 1,
    'weapon', 'small', 35, 'thermal', 8000, 4000, 0.45, 1, 8, 15, 6, 'high'
WHERE NOT EXISTS (SELECT 1 FROM item_defs WHERE id=5001);

INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type)
SELECT 5002, '小型聚焦激光 I', 'module', '中程能量武器，热能伤害', 5, 12000, false, 1,
    'weapon', 'small', 30, 'thermal', 18000, 8000, 0.30, 1, 10, 20, 8, 'high'
WHERE NOT EXISTS (SELECT 1 FROM item_defs WHERE id=5002);

INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type)
SELECT 5003, '小型脉冲激光 II', 'module', 'T2短程能量武器', 5, 50000, false, 2,
    'weapon', 'small', 45, 'thermal', 9000, 5000, 0.48, 1, 10, 18, 7, 'high'
WHERE NOT EXISTS (SELECT 1 FROM item_defs WHERE id=5003);

-- ======== 大型武器 (战列舰用) ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5150, '大型脉冲激光 I', 'module', '大型短程能量武器', 20, 500000, false, 5,
    'weapon', 'large', 220, 'thermal', 18000, 10000, 0.06, 1, 100, 60, 40, 'high'),
(5151, '大型聚焦激光 I', 'module', '大型远程能量武器', 20, 600000, false, 5,
    'weapon', 'large', 180, 'thermal', 40000, 20000, 0.04, 1, 120, 70, 50, 'high'),
(5160, '大型磁轨炮 I', 'module', '大型远程动能武器', 20, 650000, false, 5,
    'weapon', 'large', 250, 'kinetic', 50000, 25000, 0.03, 1, 90, 80, 20, 'high'),
(5161, '大型离子炮 I', 'module', '大型近距高伤动能武器', 20, 550000, false, 5,
    'weapon', 'large', 320, 'kinetic', 15000, 8000, 0.05, 1, 150, 65, 25, 'high'),
(5170, '鱼雷发射器 I', 'module', '大型鱼雷，极高单发伤害', 20, 700000, false, 5,
    'weapon', 'large', 400, 'explosive', 30000, 0, 99.0, 3, 130, 90, 0, 'high'),
(5180, '大型电磁脉冲器 I', 'module', '大型电磁武器', 20, 600000, false, 5,
    'weapon', 'large', 160, 'em', 35000, 15000, 0.04, 1, 80, 100, 60, 'high');

-- ======== 补全中槽模块 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, pg_cost, cpu_cost, cap_cost, bonus_type, bonus_value, slot_type) VALUES
-- 补回小型护盾增强器
(5300, '小型护盾增强器 I', 'module', '每Tick恢复40护盾', 5, 6000, false, 1,
    'shield', 'small', 8, 15, 12, 'shield_boost', 40, 'mid')
ON CONFLICT (id) DO NOTHING;

-- 大型护盾/推进
(5302, '大型护盾增强器 I', 'module', '每Tick恢复250护盾', 20, 400000, false, 5,
    'shield', 'large', 80, 50, 80, 'shield_boost', 250, 'mid'),
(5303, '护盾扩展器 I', 'module', '+20%护盾总量', 5, 8000, false, 1,
    'shield', 'small', 5, 20, 0, 'shield_hp_bonus', 0.20, 'mid'),
(5304, '中型护盾扩展器 I', 'module', '+25%护盾总量', 10, 70000, false, 3,
    'shield', 'medium', 15, 35, 0, 'shield_hp_bonus', 0.25, 'mid'),
(5312, '100MN加力推进器 I', 'module', '+35%亚光速(战列用)', 20, 300000, false, 5,
    'propulsion', 'large', 60, 40, 50, 'speed_bonus', 0.35, 'mid'),
(5313, '微跃引擎 I', 'module', '3秒后瞬移100km', 5, 25000, false, 1,
    'propulsion', 'small', 8, 30, 50, 'microjump', 100000, 'mid'),
-- 电容充能
(5330, '小型电容充能器 I', 'module', '每Tick恢复15电容', 5, 5000, false, 1,
    'capacitor', 'small', 3, 10, 0, 'cap_boost', 15, 'mid'),
(5331, '中型电容充能器 I', 'module', '每Tick恢复35电容', 10, 50000, false, 3,
    'capacitor', 'medium', 10, 25, 0, 'cap_boost', 35, 'mid'),
-- 感应增强
(5340, '感应增强器 I', 'module', '+30%锁定距离和传感器强度', 5, 10000, false, 1,
    'sensor', 'small', 2, 20, 5, 'sensor_boost', 0.30, 'mid'),
-- 目标锁定
(5341, '信号放大器 I', 'module', '+20%最大锁定距离', 5, 8000, false, 1,
    'sensor', 'small', 1, 15, 0, 'lock_range', 0.20, 'mid');

-- ======== 补全低槽模块 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, pg_cost, cpu_cost, bonus_type, bonus_value, slot_type) VALUES
-- 大型装甲
(5403, '大型装甲板 I', 'module', '+1500装甲HP，速度-12%', 20, 350000, false, 5,
    'armor', 'large', 80, 15, 'armor_hp', 1500, 'low'),
-- 抗性强化
(5430, '动能装甲强化 I', 'module', '装甲动能抗性+15%', 5, 12000, false, 1,
    'armor_resist', 'small', 1, 12, 'armor_kinetic_resist', 0.15, 'low'),
(5431, '热能装甲强化 I', 'module', '装甲热能抗性+15%', 5, 12000, false, 1,
    'armor_resist', 'small', 1, 12, 'armor_thermal_resist', 0.15, 'low'),
(5432, '电磁装甲强化 I', 'module', '装甲电磁抗性+15%', 5, 12000, false, 1,
    'armor_resist', 'small', 1, 12, 'armor_em_resist', 0.15, 'low'),
(5433, '爆炸装甲强化 I', 'module', '装甲爆炸抗性+15%', 5, 12000, false, 1,
    'armor_resist', 'small', 1, 12, 'armor_explosive_resist', 0.15, 'low'),
(5434, '多光谱装甲强化 I', 'module', '装甲全抗性+8%', 5, 25000, false, 1,
    'armor_resist', 'small', 2, 18, 'armor_omni_resist', 0.08, 'low'),
-- 护盾抗性(被动低槽)
(5435, '多光谱护盾强化 I', 'module', '护盾全抗性+8%', 5, 25000, false, 1,
    'shield_resist', 'small', 1, 15, 'shield_omni_resist', 0.08, 'low'),
-- 能量管理
(5440, '功率诊断系统 I', 'module', '护盾+5% 电容+5% PG+5%', 5, 8000, false, 1,
    'engineering', 'small', 1, 8, 'power_diagnostic', 0.05, 'low'),
(5441, '辅助能量栅格 I', 'module', 'PG上限+8%', 5, 10000, false, 1,
    'engineering', 'small', 0, 15, 'pg_bonus', 0.08, 'low'),
(5442, '协处理器 I', 'module', 'CPU上限+8%', 5, 10000, false, 1,
    'engineering', 'small', 3, 0, 'cpu_bonus', 0.08, 'low'),
-- 损伤控制(独特模块，只能装一个)
(5450, '损伤控制 I', 'module', '全层+12%抗性(只能装1个)', 5, 15000, false, 1,
    'damage_control', 'small', 1, 10, 'all_resist', 0.12, 'low'),
-- 中型装甲维修
(5404, '中型装甲维修器 I', 'module', '每Tick修复80装甲', 10, 60000, false, 3,
    'armor_rep', 'medium', 25, 20, 'armor_repair', 80, 'low'),
-- 大型装甲维修
(5405, '大型装甲维修器 I', 'module', '每Tick修复200装甲', 20, 400000, false, 5,
    'armor_rep', 'large', 70, 35, 'armor_repair', 200, 'low');

-- ============ 装配示例验证 ============
-- 白羊锐矛(PG65 CPU160 高3 中2 低2):
--   高: 3x小型脉冲激光I (PG8*3=24, CPU15*3=45) → 剩余PG41 CPU115
--   中: 1x小型护盾增强器I (PG8,CPU15) + 1x1MN推进器(PG5,CPU10) → 剩余PG28 CPU90
--   低: 1x小型装甲板I (PG10,CPU5) + 1x损伤控制I (PG1,CPU10) → 剩余PG17 CPU75
--   ✓ 刚好能装满，还有少量余量

-- 金牛弥诺(PG190 CPU230 高4 中3 低5):
--   高: 4x中型磁轨炮I (PG25*4=100, CPU45*4=180) → 剩余PG90 CPU50
--   中: 1x中型护盾增强器I(PG25,CPU30) → 剩余PG65 CPU20 → CPU不够再装了
--   低: 2x中型装甲板I(PG30*2=60) + 1x损伤控制 → 剩余PG4
--   ✓ 甲坦配置，PG/CPU都很紧张，需要取舍

-- 双子虚影(PG130 CPU320 高3 中5 低3):
--   高: 3x中型电磁脉冲(PG20*3=60, CPU55*3=165) → 剩余PG70 CPU155
--   中: 2x跃迁干扰(PG3*2=6, CPU25*2=50) + 2x停滞网(PG2*2=4, CPU20*2=40) + 1x推进(PG20,CPU25) → 剩余PG40 CPU40
--   低: 2x惯性稳定(PG1*2, CPU10*2) + 1x纳米纤维 → 
--   ✓ 电子战配置，CPU高适合堆电子战模块
