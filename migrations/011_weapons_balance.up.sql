-- ============ 模块参数扩展 ============
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS module_type VARCHAR(50) DEFAULT '';     -- weapon, shield, armor, propulsion, ewar, mining, rig
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS size_class VARCHAR(10) DEFAULT '';       -- small, medium, large
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS damage_per_tick INT DEFAULT 0;
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS damage_type VARCHAR(20) DEFAULT '';
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS optimal_range INT DEFAULT 0;             -- 最佳射程(m)
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS falloff_range INT DEFAULT 0;             -- 衰减射程(m)
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS tracking_speed DOUBLE PRECISION DEFAULT 0; -- 追踪速度(rad/s)
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS rate_of_fire INT DEFAULT 1;              -- 几个Tick开一次火
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS pg_cost INT DEFAULT 0;                   -- 能量栅格消耗
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS cpu_cost INT DEFAULT 0;                  -- CPU消耗
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS cap_cost INT DEFAULT 0;                  -- 每次激活电容消耗
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS bonus_type VARCHAR(50) DEFAULT '';       -- 被动加成类型
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS bonus_value DOUBLE PRECISION DEFAULT 0;  -- 被动加成数值
ALTER TABLE item_defs ADD COLUMN IF NOT EXISTS slot_type VARCHAR(10) DEFAULT '';        -- high, mid, low, rig

-- ============ 数值设计原则 ============
-- 伤害基准: 小型武器DPT=30-60, 中型=80-150, 大型=200-400
-- 射程基准: 短程<10km, 中程10-30km, 远程30km+
-- 追踪基准: 小型0.3-0.5, 中型0.1-0.25, 大型0.03-0.08 (越大越难打小船)
-- PG/CPU: 小型 5-15PG/10-30CPU, 中型 20-50PG/30-60CPU, 大型 80-200PG/50-100CPU

-- ============ 清理旧模块数据并重建 ============
DELETE FROM item_defs WHERE id BETWEEN 5001 AND 5999;

-- ======== 小型武器 (T1护卫/驱逐用) ========

-- 能量武器(热能伤害) —— 射程中等，追踪中等，耗电容
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5001, '小型脉冲激光 I', 'module', '短程能量武器，热能伤害，高追踪', 5, 8000, false, 1,
    'weapon', 'small', 35, 'thermal', 8000, 4000, 0.45, 1, 8, 15, 6, 'high'),
(5002, '小型聚焦激光 I', 'module', '中程能量武器，热能伤害', 5, 12000, false, 1,
    'weapon', 'small', 30, 'thermal', 18000, 8000, 0.30, 1, 10, 20, 8, 'high'),
(5003, '小型脉冲激光 II', 'module', '短程能量武器强化版', 5, 50000, false, 2,
    'weapon', 'small', 45, 'thermal', 9000, 5000, 0.48, 1, 10, 18, 7, 'high');

-- 混合武器(动能伤害) —— 射程较远，追踪较低
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5010, '小型磁轨炮 I', 'module', '中远程动能武器', 5, 10000, false, 1,
    'weapon', 'small', 40, 'kinetic', 20000, 10000, 0.25, 1, 6, 25, 3, 'high'),
(5011, '小型离子炮 I', 'module', '短程高伤动能武器', 5, 9000, false, 1,
    'weapon', 'small', 50, 'kinetic', 6000, 3000, 0.40, 1, 10, 18, 4, 'high');

-- 导弹(爆炸伤害) —— 不受追踪影响(100%命中)但有飞行时间延迟
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5020, '小型导弹发射器 I', 'module', '通用导弹，爆炸伤害，命中率高但伤害一般', 5, 11000, false, 1,
    'weapon', 'small', 28, 'explosive', 25000, 0, 99.0, 2, 12, 22, 0, 'high'),
(5021, '小型火箭发射器 I', 'module', '短程火箭，高伤害高射速', 5, 9000, false, 1,
    'weapon', 'small', 38, 'explosive', 8000, 0, 99.0, 1, 8, 15, 0, 'high');

-- 电磁武器(电磁伤害) —— 特殊：降低目标护盾抗性
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5030, '小型电磁脉冲器 I', 'module', '电磁伤害，对护盾特别有效', 5, 13000, false, 1,
    'weapon', 'small', 25, 'em', 12000, 6000, 0.35, 1, 5, 30, 10, 'high');

-- ======== 中型武器 (T3巡洋用) ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, slot_type) VALUES
(5101, '中型脉冲激光 I', 'module', '中型短程能量武器', 10, 80000, false, 3,
    'weapon', 'medium', 90, 'thermal', 12000, 6000, 0.18, 1, 30, 35, 15, 'high'),
(5110, '中型磁轨炮 I', 'module', '中型中远程动能武器', 10, 100000, false, 3,
    'weapon', 'medium', 100, 'kinetic', 30000, 15000, 0.10, 1, 25, 45, 8, 'high'),
(5111, '中型离子炮 I', 'module', '中型短程高伤动能', 10, 90000, false, 3,
    'weapon', 'medium', 130, 'kinetic', 10000, 5000, 0.16, 1, 40, 35, 10, 'high'),
(5120, '中型导弹发射器 I', 'module', '中型通用导弹', 10, 110000, false, 3,
    'weapon', 'medium', 80, 'explosive', 40000, 0, 99.0, 2, 35, 50, 0, 'high'),
(5130, '中型电磁脉冲器 I', 'module', '中型电磁武器', 10, 120000, false, 3,
    'weapon', 'medium', 65, 'em', 20000, 10000, 0.14, 1, 20, 55, 25, 'high');

-- ======== 采矿设备 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, damage_per_tick, damage_type, optimal_range, falloff_range, tracking_speed,
    rate_of_fire, pg_cost, cpu_cost, cap_cost, bonus_type, bonus_value, slot_type) VALUES
(5200, '小型采矿激光 I', 'module', '基础采矿装备，每周期产出100单位', 5, 5000, false, 1,
    'mining', 'small', 0, '', 5000, 0, 0, 1, 5, 10, 3, 'mining_yield', 100, 'high'),
(5201, '小型采矿激光 II', 'module', '强化采矿，每周期150单位', 5, 40000, false, 2,
    'mining', 'small', 0, '', 6000, 0, 0, 1, 7, 15, 4, 'mining_yield', 150, 'high'),
(5202, '中型采矿激光 I', 'module', '中型采矿，每周期250单位', 10, 120000, false, 3,
    'mining', 'medium', 0, '', 8000, 0, 0, 1, 15, 25, 8, 'mining_yield', 250, 'high');

-- ======== 中槽模块 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, pg_cost, cpu_cost, cap_cost, bonus_type, bonus_value, slot_type) VALUES
-- 护盾
(5300, '小型护盾增强器 I', 'module', '每Tick恢复40护盾', 5, 6000, false, 1,
    'shield', 'small', 8, 15, 12, 'shield_boost', 40, 'mid'),
(5301, '中型护盾增强器 I', 'module', '每Tick恢复100护盾', 10, 60000, false, 3,
    'shield', 'medium', 25, 30, 30, 'shield_boost', 100, 'mid'),
-- 推进
(5310, '1MN加力推进器 I', 'module', '+50%亚光速速度', 5, 4000, false, 1,
    'propulsion', 'small', 5, 10, 8, 'speed_bonus', 0.50, 'mid'),
(5311, '10MN加力推进器 I', 'module', '+40%亚光速速度(巡洋用)', 10, 40000, false, 3,
    'propulsion', 'medium', 20, 25, 20, 'speed_bonus', 0.40, 'mid'),
-- 电子战
(5320, '跃迁干扰器 I', 'module', '阻止目标跃迁', 5, 15000, false, 1,
    'ewar', 'small', 3, 25, 10, 'warp_disrupt', 1, 'mid'),
(5321, '停滞网 I', 'module', '降低目标速度60%', 5, 12000, false, 1,
    'ewar', 'small', 2, 20, 8, 'web_strength', 0.60, 'mid'),
(5322, 'ECM干扰器 I', 'module', '概率使目标丢失锁定', 5, 18000, false, 1,
    'ewar', 'small', 3, 30, 12, 'ecm_strength', 3.0, 'mid'),
(5323, '目标标记器 I', 'module', '增大目标信号半径30%', 5, 10000, false, 1,
    'ewar', 'small', 2, 15, 6, 'sig_increase', 0.30, 'mid');

-- ======== 低槽模块 ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, pg_cost, cpu_cost, bonus_type, bonus_value, slot_type) VALUES
-- 装甲
(5400, '小型装甲板 I', 'module', '+200装甲HP，降低速度5%', 5, 5000, false, 1,
    'armor', 'small', 10, 5, 'armor_hp', 200, 'low'),
(5401, '中型装甲板 I', 'module', '+600装甲HP，降低速度8%', 10, 50000, false, 3,
    'armor', 'medium', 30, 10, 'armor_hp', 600, 'low'),
(5402, '小型装甲维修器 I', 'module', '每Tick修复30装甲', 5, 7000, false, 1,
    'armor_rep', 'small', 8, 12, 'armor_repair', 30, 'low'),
-- 伤害增强
(5410, '磁稳定器 I', 'module', '混合武器伤害+8%', 5, 10000, false, 1,
    'damage_mod', 'small', 2, 15, 'kinetic_damage_bonus', 0.08, 'low'),
(5411, '散热器 I', 'module', '能量武器伤害+8%', 5, 10000, false, 1,
    'damage_mod', 'small', 2, 15, 'thermal_damage_bonus', 0.08, 'low'),
(5412, '弹道控制系统 I', 'module', '导弹伤害+8%', 5, 10000, false, 1,
    'damage_mod', 'small', 2, 15, 'explosive_damage_bonus', 0.08, 'low'),
-- 速度/灵活
(5420, '惯性稳定器 I', 'module', '对齐时间-15%，信号半径+10%', 5, 6000, false, 1,
    'navigation', 'small', 1, 10, 'align_bonus', -0.15, 'low'),
(5421, '纳米纤维结构 I', 'module', '速度+8%，结构HP-8%', 5, 8000, false, 1,
    'navigation', 'small', 1, 8, 'speed_bonus', 0.08, 'low');

-- ======== 改装件(永久安装，拆除摧毁) ========
INSERT INTO item_defs (id, name, category, description, volume, base_price, stackable, tech_level,
    module_type, size_class, bonus_type, bonus_value, slot_type) VALUES
(5500, '小型三角装甲泵 I', 'module', '装甲HP+10%', 5, 20000, false, 1,
    'rig', 'small', 'armor_hp_pct', 0.10, 'rig'),
(5501, '小型核心防御力场扩展 I', 'module', '护盾HP+10%', 5, 20000, false, 1,
    'rig', 'small', 'shield_hp_pct', 0.10, 'rig'),
(5502, '小型辅助推进器 I', 'module', '亚光速速度+10%', 5, 18000, false, 1,
    'rig', 'small', 'speed_pct', 0.10, 'rig'),
(5503, '小型火力控制回路 I', 'module', '武器追踪速度+10%', 5, 25000, false, 1,
    'rig', 'small', 'tracking_pct', 0.10, 'rig'),
(5504, '小型碰撞加速器 I', 'module', '武器伤害+10%', 5, 30000, false, 1,
    'rig', 'small', 'damage_pct', 0.10, 'rig'),
(5505, '小型采矿器强化 I', 'module', '采矿产出+10%', 5, 15000, false, 1,
    'rig', 'small', 'mining_yield_pct', 0.10, 'rig'),
(5506, '小型宇航扩展 I', 'module', '货仓容量+15%', 5, 12000, false, 1,
    'rig', 'small', 'cargo_pct', 0.15, 'rig'),
(5510, '中型三角装甲泵 I', 'module', '装甲HP+15%', 10, 150000, false, 3,
    'rig', 'medium', 'armor_hp_pct', 0.15, 'rig'),
(5511, '中型核心防御力场扩展 I', 'module', '护盾HP+15%', 10, 150000, false, 3,
    'rig', 'medium', 'shield_hp_pct', 0.15, 'rig');

-- ============ 完善所有9艘舰船的详细属性 ============
-- 设计原则:
-- T1护卫: PG 50-70, CPU 100-170, 总HP 1200-1800
-- T2驱逐: PG 80-100, CPU 160-200, 总HP 2500-3500
-- T3巡洋: PG 140-200, CPU 220-300, 总HP 7000-12000

-- 白羊座(火象): 高伤害，中等防御，速度快
UPDATE ship_defs SET
  shield_hp=800, armor_hp=600, structure_hp=400, shield_recharge=25,
  capacitor=400, cap_recharge=12,
  max_speed=450, warp_speed=5.0, align_ticks=2, signature=40,
  high_slots=3, mid_slots=2, low_slots=2,
  cargo_m3=150, powergrid=60, cpu=130, mass=1200000
WHERE id=1; -- 锐矛(突击护卫): 高速高伤，薄皮

UPDATE ship_defs SET
  shield_hp=600, armor_hp=400, structure_hp=350, shield_recharge=20,
  capacitor=500, cap_recharge=15,
  max_speed=380, warp_speed=4.5, align_ticks=3, signature=35,
  high_slots=2, mid_slots=3, low_slots=2,
  cargo_m3=120, powergrid=50, cpu=160, mass=1400000
WHERE id=2; -- 怒目(电子护卫): CPU高，中槽多

UPDATE ship_defs SET
  shield_hp=1500, armor_hp=1200, structure_hp=800, shield_recharge=30,
  capacitor=600, cap_recharge=15,
  max_speed=350, warp_speed=4.0, align_ticks=4, signature=80,
  high_slots=5, mid_slots=3, low_slots=2,
  cargo_m3=300, powergrid=95, cpu=190, mass=3500000
WHERE id=3; -- 焚城(火力驱逐): 5高槽=高火力

UPDATE ship_defs SET
  shield_hp=4000, armor_hp=3000, structure_hp=2000, shield_recharge=50,
  capacitor=1200, cap_recharge=25,
  max_speed=250, warp_speed=3.0, align_ticks=6, signature=150,
  high_slots=4, mid_slots=4, low_slots=3,
  cargo_m3=450, powergrid=160, cpu=260, mass=12000000
WHERE id=4; -- 战嚎(主战巡洋): 均衡

-- 金牛座(土象): 高装甲，慢速，大货仓
UPDATE ship_defs SET
  shield_hp=700, armor_hp=800, structure_hp=500, shield_recharge=20,
  capacitor=350, cap_recharge=10,
  max_speed=380, warp_speed=4.0, align_ticks=3, signature=45,
  high_slots=3, mid_slots=2, low_slots=3,
  cargo_m3=200, powergrid=55, cpu=120, mass=1500000
WHERE id=5; -- 铁蹄(突击护卫): 装甲厚，低槽多

UPDATE ship_defs SET
  shield_hp=500, armor_hp=700, structure_hp=400, shield_recharge=15,
  capacitor=300, cap_recharge=10,
  max_speed=250, warp_speed=3.0, align_ticks=5, signature=55,
  high_slots=2, mid_slots=2, low_slots=2,
  cargo_m3=600, powergrid=45, cpu=100, mass=2500000
WHERE id=6; -- 犁刃(工业护卫): 大货仓，慢

UPDATE ship_defs SET
  shield_hp=3500, armor_hp=5000, structure_hp=2800, shield_recharge=40,
  capacitor=1000, cap_recharge=20,
  max_speed=200, warp_speed=2.5, align_ticks=7, signature=180,
  high_slots=4, mid_slots=3, low_slots=5,
  cargo_m3=700, powergrid=190, cpu=230, mass=15000000
WHERE id=7; -- 弥诺(主战巡洋): 装甲王，5低槽

-- 双子座(风象): 高速，电子战强，薄皮
UPDATE ship_defs SET
  shield_hp=600, armor_hp=400, structure_hp=300, shield_recharge=20,
  capacitor=450, cap_recharge=14,
  max_speed=500, warp_speed=6.0, align_ticks=2, signature=30,
  high_slots=3, mid_slots=3, low_slots=2,
  cargo_m3=130, powergrid=50, cpu=170, mass=1000000
WHERE id=8; -- 影刺(突击护卫): 最快，最薄

UPDATE ship_defs SET
  shield_hp=3000, armor_hp=2000, structure_hp=1500, shield_recharge=35,
  capacitor=1400, cap_recharge=30,
  max_speed=300, warp_speed=4.5, align_ticks=4, signature=100,
  high_slots=3, mid_slots=5, low_slots=3,
  cargo_m3=350, powergrid=130, cpu=320, mass=11000000
WHERE id=9; -- 虚影(隐形巡洋): CPU最高，5中槽

-- ============ 舰船特性(种族加成) ============
-- 作为描述存入ship_defs的一个新字段
ALTER TABLE ship_defs ADD COLUMN IF NOT EXISTS race_bonus TEXT DEFAULT '';

UPDATE ship_defs SET race_bonus = '白羊座加成: 每级小型能量武器技能+5%伤害' WHERE id IN (1, 2);
UPDATE ship_defs SET race_bonus = '白羊座加成: 每级小型能量武器技能+5%伤害, 驱逐舰驾驶每级+10%最佳射程' WHERE id = 3;
UPDATE ship_defs SET race_bonus = '白羊座加成: 每级中型武器+5%伤害, 巡洋舰驾驶每级+7.5%最佳射程' WHERE id = 4;
UPDATE ship_defs SET race_bonus = '金牛座加成: 每级护卫舰驾驶+5%装甲HP, 每级采矿+5%产出' WHERE id IN (5, 6);
UPDATE ship_defs SET race_bonus = '金牛座加成: 每级巡洋舰驾驶+7.5%装甲HP, 每级装甲维修+5%效率' WHERE id = 7;
UPDATE ship_defs SET race_bonus = '双子座加成: 每级护卫舰驾驶-5%信号半径, 每级电子战+5%效果' WHERE id = 8;
UPDATE ship_defs SET race_bonus = '双子座加成: 每级巡洋舰驾驶-7.5%信号半径, 可装备隐形装置' WHERE id = 9;
