-- ============================================================
-- 022: 大批量补充装备模块
-- ============================================================

-- ==================== 低槽 (low) ====================

-- 装甲板 (小型/中型 + II级)
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8001,'小型装甲板 I','module','low','armor','small',15,8,'armor_hp',400,'增加400装甲值'),
(8002,'小型装甲板 II','module','low','armor','small',18,10,'armor_hp',520,'增加520装甲值'),
(8003,'中型装甲板 I','module','low','armor','medium',40,12,'armor_hp',800,'增加800装甲值'),
(8004,'中型装甲板 II','module','low','armor','medium',48,15,'armor_hp',1040,'增加1040装甲值'),
(8005,'大型装甲板 II','module','low','armor','large',96,18,'armor_hp',1950,'增加1950装甲值')
ON CONFLICT (name) DO NOTHING;

-- 装甲维修器 (小型 I)
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8010,'小型装甲维修器 I','module','low','armor_rep','small',8,12,'armor_repair',30,'每Tick修复30装甲')
ON CONFLICT (name) DO NOTHING;

-- 装甲抗性 II级 + 中型
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8020,'动能装甲强化 II','module','low','armor_resist','small',1,15,'armor_kinetic_resist',0.20,'装甲动能抗性+20%'),
(8021,'热能装甲强化 II','module','low','armor_resist','small',1,15,'armor_thermal_resist',0.20,'装甲热能抗性+20%'),
(8022,'电磁装甲强化 II','module','low','armor_resist','small',1,15,'armor_em_resist',0.20,'装甲电磁抗性+20%'),
(8023,'爆炸装甲强化 II','module','low','armor_resist','small',1,15,'armor_explosive_resist',0.20,'装甲爆炸抗性+20%'),
(8024,'多光谱装甲强化 II','module','low','armor_resist','small',3,22,'armor_omni_resist',0.12,'装甲全抗+12%'),
(8025,'中型动能装甲强化 I','module','low','armor_resist','medium',3,20,'armor_kinetic_resist',0.20,'装甲动能抗性+20%'),
(8026,'中型热能装甲强化 I','module','low','armor_resist','medium',3,20,'armor_thermal_resist',0.20,'装甲热能抗性+20%'),
(8027,'中型电磁装甲强化 I','module','low','armor_resist','medium',3,20,'armor_em_resist',0.20,'装甲电磁抗性+20%'),
(8028,'中型爆炸装甲强化 I','module','low','armor_resist','medium',3,20,'armor_explosive_resist',0.20,'装甲爆炸抗性+20%'),
(8029,'中型多光谱装甲强化 I','module','low','armor_resist','medium',5,28,'armor_omni_resist',0.15,'装甲全抗+15%')
ON CONFLICT (name) DO NOTHING;

-- 伤害增强模块 I级
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8030,'磁稳定器 I','module','low','damage_mod','small',2,14,'kinetic_damage_bonus',0.08,'动能武器伤害+8%'),
(8031,'散热器 I','module','low','damage_mod','small',2,14,'thermal_damage_bonus',0.08,'热能武器伤害+8%'),
(8032,'弹道控制系统 I','module','low','damage_mod','small',2,14,'explosive_damage_bonus',0.08,'爆炸武器伤害+8%'),
(8033,'全武器伤害控制 I','module','low','damage_mod','small',3,18,'all_damage_bonus',0.05,'全武器伤害+5%'),
(8034,'全武器伤害控制 II','module','low','damage_mod','small',4,22,'all_damage_bonus',0.08,'全武器伤害+8%')
ON CONFLICT (name) DO NOTHING;

-- 纳米纤维 / 惯性稳定器 / 跟踪增强
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8040,'纳米纤维内衬 I','module','low','engineering','small',1,5,'speed_bonus',0.08,'速度+8%'),
(8041,'纳米纤维内衬 II','module','low','engineering','small',1,6,'speed_bonus',0.12,'速度+12%'),
(8042,'惯性稳定器 I','module','low','engineering','small',1,8,'agility_bonus',0.12,'灵活性+12%(减少对齐时间)'),
(8043,'惯性稳定器 II','module','low','engineering','small',1,10,'agility_bonus',0.16,'灵活性+16%'),
(8044,'跟踪增强器 I','module','low','engineering','small',2,12,'tracking_bonus',0.15,'追踪速度+15%'),
(8045,'跟踪增强器 II','module','low','engineering','small',3,15,'tracking_bonus',0.20,'追踪速度+20%'),
(8046,'信号增幅器 I','module','low','engineering','small',1,10,'signature_reduction',0.05,'信号半径-5%'),
(8047,'信号增幅器 II','module','low','engineering','small',1,12,'signature_reduction',0.08,'信号半径-8%')
ON CONFLICT (name) DO NOTHING;

-- II级工程模块
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8050,'功率诊断系统 II','module','low','engineering','small',1,10,'power_diagnostic',0.08,'综合能量诊断+8%'),
(8051,'辅助能量栅格 II','module','low','engineering','small',0,18,'pg_bonus',0.12,'能量栅格+12%'),
(8052,'协处理器 II','module','low','engineering','small',4,0,'cpu_bonus',0.12,'CPU+12%')
ON CONFLICT (name) DO NOTHING;

-- 护盾抗性 (低槽) — 补充II级
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8060,'多光谱护盾强化 II','module','low','shield_resist','small',2,18,'shield_omni_resist',0.12,'护盾全抗+12%')
ON CONFLICT (name) DO NOTHING;


-- ==================== 中槽 (mid) ====================

-- 护盾增强器补充
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8100,'小型护盾增强器 I','module','mid','shield','small',6,14,10,'shield_boost',35,'每次激活回复35护盾'),
(8101,'大型护盾增强器 II','module','mid','shield','large',96,60,96,'shield_boost',325,'每次激活回复325护盾'),
(8102,'大型护盾扩展器 I','module','mid','shield','large',30,45,0,'shield_hp_bonus',0.35,'护盾总量+35%'),
(8103,'护盾扩展器 II','module','mid','shield','small',6,24,0,'shield_hp_bonus',0.26,'护盾总量+26%'),
(8104,'中型护盾扩展器 II','module','mid','shield','medium',18,42,0,'shield_hp_bonus',0.33,'护盾总量+33%')
ON CONFLICT (name) DO NOTHING;

-- 护盾抗性 (中槽)
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8110,'动能护盾偏转力场 I','module','mid','shield_resist','small',2,15,5,'shield_kinetic_resist',0.20,'护盾动能抗+20%'),
(8111,'热能护盾偏转力场 I','module','mid','shield_resist','small',2,15,5,'shield_thermal_resist',0.20,'护盾热能抗+20%'),
(8112,'电磁护盾偏转力场 I','module','mid','shield_resist','small',2,15,5,'shield_em_resist',0.20,'护盾电磁抗+20%'),
(8113,'爆炸护盾偏转力场 I','module','mid','shield_resist','small',2,15,5,'shield_explosive_resist',0.20,'护盾爆炸抗+20%'),
(8114,'多光谱护盾偏转力场 I','module','mid','shield_resist','small',3,20,8,'shield_omni_resist',0.12,'护盾全抗+12%'),
(8115,'动能护盾偏转力场 II','module','mid','shield_resist','small',3,18,6,'shield_kinetic_resist',0.25,'护盾动能抗+25%'),
(8116,'热能护盾偏转力场 II','module','mid','shield_resist','small',3,18,6,'shield_thermal_resist',0.25,'护盾热能抗+25%'),
(8117,'电磁护盾偏转力场 II','module','mid','shield_resist','small',3,18,6,'shield_em_resist',0.25,'护盾电磁抗+25%'),
(8118,'爆炸护盾偏转力场 II','module','mid','shield_resist','small',3,18,6,'shield_explosive_resist',0.25,'护盾爆炸抗+25%'),
(8119,'多光谱护盾偏转力场 II','module','mid','shield_resist','small',4,24,10,'shield_omni_resist',0.15,'护盾全抗+15%')
ON CONFLICT (name) DO NOTHING;

-- 电容模块补充
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8120,'大型电容充能器 I','module','mid','capacitor','large',25,40,0,'cap_boost',80,'电容+80'),
(8121,'小型电容充能器 II','module','mid','capacitor','small',4,12,0,'cap_boost',20,'电容+20'),
(8122,'中型电容充能器 II','module','mid','capacitor','medium',12,30,0,'cap_boost',45,'电容+45'),
(8123,'电容回充增强器 I','module','mid','capacitor','small',3,15,0,'cap_recharge_bonus',0.15,'电容回充+15%'),
(8124,'电容回充增强器 II','module','mid','capacitor','small',4,18,0,'cap_recharge_bonus',0.20,'电容回充+20%')
ON CONFLICT (name) DO NOTHING;

-- 推进模块补充
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8130,'1MN加力推进器 I','module','mid','propulsion','small',4,10,8,'speed_bonus',0.40,'速度+40%'),
(8131,'10MN加力推进器 I','module','mid','propulsion','medium',18,24,18,'speed_bonus',0.35,'速度+35%'),
(8132,'100MN加力推进器 II','module','mid','propulsion','large',72,48,60,'speed_bonus',0.45,'速度+45%'),
(8133,'1MN微曲引擎 I','module','mid','propulsion','small',10,15,15,'speed_bonus',0.80,'速度+80%，增大信号半径'),
(8134,'10MN微曲引擎 I','module','mid','propulsion','medium',30,30,35,'speed_bonus',0.80,'速度+80%，增大信号半径'),
(8135,'100MN微曲引擎 I','module','mid','propulsion','large',100,55,80,'speed_bonus',0.80,'速度+80%，增大信号半径')
ON CONFLICT (name) DO NOTHING;

-- 电子战模块
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8140,'跃迁扰频器 I','module','mid','ewar','small',1,25,20,'warp_disrupt',1,'阻止目标跃迁(1点)'),
(8141,'跃迁干扰器 I','module','mid','ewar','small',1,30,25,'warp_scramble',2,'阻止目标跃迁(2点)+关闭微曲'),
(8142,'感应抑制器 I','module','mid','ewar','small',1,25,15,'sensor_damp',0.25,'降低目标锁定距离25%'),
(8143,'目标标记器 I','module','mid','ewar','small',1,20,10,'target_paint',0.20,'增大目标信号半径20%'),
(8144,'目标标记器 II','module','mid','ewar','small',2,24,12,'target_paint',0.28,'增大目标信号半径28%'),
(8145,'ECM干扰器 I','module','mid','ewar','small',1,30,20,'ecm',0.20,'20%概率干扰目标锁定'),
(8146,'追踪干扰器 I','module','mid','ewar','small',1,20,12,'tracking_disrupt',0.25,'降低目标追踪速度25%'),
(8147,'追踪干扰器 II','module','mid','ewar','small',2,24,15,'tracking_disrupt',0.32,'降低目标追踪速度32%'),
(8148,'能量吸附器 I','module','mid','ewar','medium',5,25,0,'energy_nosferatu',15,'每Tick吸取15电容'),
(8149,'能量中和器 I','module','mid','ewar','medium',5,30,20,'energy_neutralizer',25,'每Tick中和25电容')
ON CONFLICT (name) DO NOTHING;

-- 感应模块补充
INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,cap_cost,bonus_type,bonus_value,description)
VALUES
(8150,'感应增强器 II','module','mid','sensor','small',3,24,6,'sensor_boost',0.40,'感应强度+40%'),
(8151,'信号放大器 II','module','mid','sensor','small',2,18,0,'lock_range',0.28,'锁定距离+28%'),
(8152,'中型感应增强器 I','module','mid','sensor','medium',5,30,10,'sensor_boost',0.40,'感应强度+40%')
ON CONFLICT (name) DO NOTHING;


-- ==================== 改装件 (rig) 补充 ====================

INSERT INTO item_defs (id,name,category,slot_type,module_type,size_class,pg_cost,cpu_cost,bonus_type,bonus_value,description)
VALUES
(8200,'中型辅助推进器 I','module','rig','rig','medium',0,0,'speed_pct',0.15,'速度+15%'),
(8201,'中型火力控制回路 I','module','rig','rig','medium',0,0,'tracking_pct',0.15,'追踪速度+15%'),
(8202,'中型碰撞加速器 I','module','rig','rig','medium',0,0,'damage_pct',0.15,'武器伤害+15%'),
(8203,'小型电容控制回路 I','module','rig','rig','small',0,0,'cap_recharge_pct',0.15,'电容回充+15%'),
(8204,'中型电容控制回路 I','module','rig','rig','medium',0,0,'cap_recharge_pct',0.20,'电容回充+20%'),
(8205,'小型抗性增幅器 I','module','rig','rig','small',0,0,'omni_resist_pct',0.05,'全抗+5%'),
(8206,'中型抗性增幅器 I','module','rig','rig','medium',0,0,'omni_resist_pct',0.08,'全抗+8%'),
(8207,'中型采矿器强化 I','module','rig','rig','medium',0,0,'mining_yield_pct',0.15,'采矿效率+15%'),
(8208,'中型宇航扩展 I','module','rig','rig','medium',0,0,'cargo_pct',0.20,'货舱容量+20%')
ON CONFLICT (name) DO NOTHING;

-- 更新ID序列
SELECT setval('item_defs_id_seq', (SELECT MAX(id) FROM item_defs));
