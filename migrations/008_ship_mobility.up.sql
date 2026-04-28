-- 补充舰船移动参数
ALTER TABLE ship_defs ADD COLUMN warp_speed DOUBLE PRECISION NOT NULL DEFAULT 3.0;  -- 跃迁速度(AU/s)
ALTER TABLE ship_defs ADD COLUMN warp_cap_cost INT NOT NULL DEFAULT 50;              -- 每次跃迁消耗电容
ALTER TABLE ship_defs ADD COLUMN jump_range DOUBLE PRECISION NOT NULL DEFAULT 0;     -- 跳跃引擎范围(光年)，0=无跳跃能力
ALTER TABLE ship_defs ADD COLUMN mass BIGINT NOT NULL DEFAULT 10000000;              -- 质量(kg)，影响对齐速度和推进

-- 不同级别的对齐时间已有(align_ticks)，现在设置合理值
-- align_ticks = 对齐到跃迁需要的Tick数，与质量正相关

-- 白羊座: 速度快，跃迁快，质量轻
UPDATE ship_defs SET max_speed=450, warp_speed=5.0, align_ticks=2, mass=1200000, warp_cap_cost=30 WHERE id=1;  -- 锐矛(突击护卫)
UPDATE ship_defs SET max_speed=380, warp_speed=4.5, align_ticks=3, mass=1400000, warp_cap_cost=35 WHERE id=2;  -- 怒目(电子护卫)
UPDATE ship_defs SET max_speed=350, warp_speed=4.0, align_ticks=4, mass=3500000, warp_cap_cost=60 WHERE id=3;  -- 焚城(驱逐)
UPDATE ship_defs SET max_speed=250, warp_speed=3.0, align_ticks=6, mass=12000000, warp_cap_cost=120 WHERE id=4; -- 战嚎(巡洋)

-- 金牛座: 速度慢，但货仓大
UPDATE ship_defs SET max_speed=380, warp_speed=4.0, align_ticks=3, mass=1500000, warp_cap_cost=35 WHERE id=5;  -- 铁蹄(突击护卫)
UPDATE ship_defs SET max_speed=250, warp_speed=3.0, align_ticks=5, mass=2500000, warp_cap_cost=50 WHERE id=6;  -- 犁刃(工业护卫)
UPDATE ship_defs SET max_speed=200, warp_speed=2.5, align_ticks=7, mass=15000000, warp_cap_cost=150 WHERE id=7; -- 弥诺(巡洋)

-- 双子座: 最快的对齐和跃迁
UPDATE ship_defs SET max_speed=500, warp_speed=6.0, align_ticks=2, mass=1000000, warp_cap_cost=25 WHERE id=8;  -- 影刺(突击护卫)
UPDATE ship_defs SET max_speed=300, warp_speed=4.5, align_ticks=4, mass=11000000, warp_cap_cost=100 WHERE id=9; -- 虚影(隐形巡洋)

-- NPC也需要移动参数
ALTER TABLE npc_defs ADD COLUMN warp_speed DOUBLE PRECISION NOT NULL DEFAULT 2.0;
ALTER TABLE npc_defs ADD COLUMN align_ticks INT NOT NULL DEFAULT 5;

-- 失控机器: 慢速笨重
UPDATE npc_defs SET warp_speed=1.5, align_ticks=8 WHERE npc_type='machine';
-- 异星生物: 不会跃迁
UPDATE npc_defs SET warp_speed=0, align_ticks=99 WHERE npc_type='alien';
-- 维度体: 不用跃迁（维度滑移）
UPDATE npc_defs SET warp_speed=0, align_ticks=1 WHERE npc_type='dimension';
