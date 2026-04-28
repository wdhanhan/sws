-- 给舰船定义表加入抗性字段
ALTER TABLE ship_defs ADD COLUMN shield_res_kinetic DOUBLE PRECISION NOT NULL DEFAULT 0.15;
ALTER TABLE ship_defs ADD COLUMN shield_res_thermal DOUBLE PRECISION NOT NULL DEFAULT 0.40;
ALTER TABLE ship_defs ADD COLUMN shield_res_em DOUBLE PRECISION NOT NULL DEFAULT 0.30;
ALTER TABLE ship_defs ADD COLUMN shield_res_explosive DOUBLE PRECISION NOT NULL DEFAULT 0.10;
ALTER TABLE ship_defs ADD COLUMN armor_res_kinetic DOUBLE PRECISION NOT NULL DEFAULT 0.40;
ALTER TABLE ship_defs ADD COLUMN armor_res_thermal DOUBLE PRECISION NOT NULL DEFAULT 0.20;
ALTER TABLE ship_defs ADD COLUMN armor_res_em DOUBLE PRECISION NOT NULL DEFAULT 0.10;
ALTER TABLE ship_defs ADD COLUMN armor_res_explosive DOUBLE PRECISION NOT NULL DEFAULT 0.30;

-- 给NPC定义表也加入抗性
ALTER TABLE npc_defs ADD COLUMN shield_res_kinetic DOUBLE PRECISION NOT NULL DEFAULT 0.15;
ALTER TABLE npc_defs ADD COLUMN shield_res_thermal DOUBLE PRECISION NOT NULL DEFAULT 0.30;
ALTER TABLE npc_defs ADD COLUMN shield_res_em DOUBLE PRECISION NOT NULL DEFAULT 0.20;
ALTER TABLE npc_defs ADD COLUMN shield_res_explosive DOUBLE PRECISION NOT NULL DEFAULT 0.10;
ALTER TABLE npc_defs ADD COLUMN armor_res_kinetic DOUBLE PRECISION NOT NULL DEFAULT 0.30;
ALTER TABLE npc_defs ADD COLUMN armor_res_thermal DOUBLE PRECISION NOT NULL DEFAULT 0.20;
ALTER TABLE npc_defs ADD COLUMN armor_res_em DOUBLE PRECISION NOT NULL DEFAULT 0.15;
ALTER TABLE npc_defs ADD COLUMN armor_res_explosive DOUBLE PRECISION NOT NULL DEFAULT 0.20;

-- ============ 设置种族舰船抗性 ============
-- 火象种族(白羊1,狮子5,射手9): 热能高抗(常年在高辐射环境)，爆炸低抗
-- 白羊座: 护盾偏弱但装甲热抗极高(冲锋硬扛)
UPDATE ship_defs SET
  shield_res_kinetic=0.10, shield_res_thermal=0.45, shield_res_em=0.25, shield_res_explosive=0.05,
  armor_res_kinetic=0.35, armor_res_thermal=0.50, armor_res_em=0.10, armor_res_explosive=0.15
WHERE race_id = 1;

-- 金牛座(2): 装甲全面厚实(工业重甲)，护盾一般
UPDATE ship_defs SET
  shield_res_kinetic=0.20, shield_res_thermal=0.25, shield_res_em=0.20, shield_res_explosive=0.15,
  armor_res_kinetic=0.50, armor_res_thermal=0.35, armor_res_em=0.25, armor_res_explosive=0.40
WHERE race_id = 2;

-- 双子座(3): 电磁高抗(电子战专精)，但装甲薄
UPDATE ship_defs SET
  shield_res_kinetic=0.15, shield_res_thermal=0.20, shield_res_em=0.50, shield_res_explosive=0.10,
  armor_res_kinetic=0.20, armor_res_thermal=0.15, armor_res_em=0.40, armor_res_explosive=0.10
WHERE race_id = 3;

-- 失控机器NPC: 动能高抗(金属外壳)，电磁低抗
UPDATE npc_defs SET
  shield_res_kinetic=0.35, shield_res_thermal=0.15, shield_res_em=0.05, shield_res_explosive=0.20,
  armor_res_kinetic=0.45, armor_res_thermal=0.20, armor_res_em=0.05, armor_res_explosive=0.25
WHERE npc_type = 'machine';

-- 异星生物NPC: 热能高抗(生物耐热)，动能低抗
UPDATE npc_defs SET
  shield_res_kinetic=0.05, shield_res_thermal=0.40, shield_res_em=0.20, shield_res_explosive=0.10,
  armor_res_kinetic=0.10, armor_res_thermal=0.45, armor_res_em=0.15, armor_res_explosive=0.15
WHERE npc_type = 'alien';

-- 维度入侵者NPC: 爆炸高抗(维度体不实体)，热能低抗
UPDATE npc_defs SET
  shield_res_kinetic=0.20, shield_res_thermal=0.05, shield_res_em=0.30, shield_res_explosive=0.45,
  armor_res_kinetic=0.15, armor_res_thermal=0.05, armor_res_em=0.35, armor_res_explosive=0.50
WHERE npc_type = 'dimension';
