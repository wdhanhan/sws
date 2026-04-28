-- ============ 技能系统 ============
CREATE TABLE IF NOT EXISTS skill_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL, -- piloting, gunnery, defense, industry, research, navigation, leadership, trade, special
    description TEXT NOT NULL DEFAULT '',
    rank INT NOT NULL DEFAULT 1,  -- 难度等级1-16，影响训练时间倍率
    primary_attr VARCHAR(20) NOT NULL DEFAULT 'intelligence', -- intelligence, perception, willpower, memory, charisma
    secondary_attr VARCHAR(20) NOT NULL DEFAULT 'memory',
    prereq_skill_id BIGINT, -- 前置技能
    prereq_level INT DEFAULT 0  -- 前置技能需要的等级
);

-- 角色已学技能
CREATE TABLE IF NOT EXISTS character_skills (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill_def_id BIGINT NOT NULL REFERENCES skill_defs(id),
    level INT NOT NULL DEFAULT 0,  -- 当前等级 0-5
    skill_points BIGINT NOT NULL DEFAULT 0, -- 当前累积技能点
    UNIQUE(character_id, skill_def_id)
);

-- 技能训练队列
CREATE TABLE IF NOT EXISTS skill_queue (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    skill_def_id BIGINT NOT NULL REFERENCES skill_defs(id),
    target_level INT NOT NULL, -- 目标等级
    queue_position INT NOT NULL, -- 队列位置 0=当前训练
    start_time TIMESTAMP WITH TIME ZONE,
    finish_time TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT FALSE
);

-- ============ 植入体系统 ============
CREATE TABLE IF NOT EXISTS implant_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slot_type VARCHAR(20) NOT NULL, -- cortex(大脑皮层), spine(脊髓), gene(基因), neural(神经接口)
    grade VARCHAR(20) NOT NULL DEFAULT 'white', -- white, green, blue, purple, orange
    description TEXT NOT NULL DEFAULT '',
    effect_type VARCHAR(50) NOT NULL, -- train_speed, mining_yield, damage_bonus, shield_boost, etc.
    effect_value DOUBLE PRECISION NOT NULL DEFAULT 0.05, -- 加成幅度
    set_name VARCHAR(50) DEFAULT '', -- 套装名(空=不属于套装)
    stability INT NOT NULL DEFAULT 100 -- 稳定度 0-100
);

-- 角色已装备的植入体
CREATE TABLE IF NOT EXISTS character_implants (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    implant_def_id BIGINT NOT NULL REFERENCES implant_defs(id),
    slot_index INT NOT NULL, -- 槽位编号
    stability INT NOT NULL DEFAULT 100, -- 当前稳定度
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    installed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ 克隆系统 ============
CREATE TABLE IF NOT EXISTS clones (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    clone_type VARCHAR(20) NOT NULL DEFAULT 'main', -- main, branch, jump
    station_id BIGINT NOT NULL REFERENCES stations(id),
    system_id BIGINT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_jump_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 死亡回响
CREATE TABLE IF NOT EXISTS death_echoes (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    system_id BIGINT NOT NULL,
    pos_x DOUBLE PRECISION NOT NULL,
    pos_y DOUBLE PRECISION NOT NULL,
    pos_z DOUBLE PRECISION NOT NULL,
    killed_by VARCHAR(100) NOT NULL DEFAULT 'unknown',
    combat_log TEXT,
    is_collected BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_char_skills ON character_skills(character_id);
CREATE INDEX idx_skill_queue ON skill_queue(character_id, queue_position);
CREATE INDEX idx_char_implants ON character_implants(character_id);
CREATE INDEX idx_clones ON clones(character_id);
CREATE INDEX idx_death_echoes ON death_echoes(character_id, is_collected);

-- ============ 技能种子数据 ============
-- 驾驶技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(1, '护卫舰驾驶', 'piloting', '驾驶T1护卫舰的基础能力', 1, 'perception', 'willpower'),
(2, '驱逐舰驾驶', 'piloting', '驾驶T2驱逐舰', 2, 'perception', 'willpower'),
(3, '巡洋舰驾驶', 'piloting', '驾驶T3巡洋舰', 4, 'perception', 'willpower'),
(4, '战列巡洋舰驾驶', 'piloting', '驾驶T4战列巡洋舰', 6, 'perception', 'willpower'),
(5, '战列舰驾驶', 'piloting', '驾驶T5战列舰', 8, 'perception', 'willpower');

-- 武器技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(10, '小型能量武器', 'gunnery', '操作小型激光/等离子武器', 1, 'perception', 'memory'),
(11, '小型混合武器', 'gunnery', '操作小型磁轨/离子武器', 1, 'perception', 'memory'),
(12, '小型导弹', 'gunnery', '操作小型导弹发射器', 1, 'perception', 'memory'),
(13, '中型能量武器', 'gunnery', '操作中型能量武器', 3, 'perception', 'memory'),
(14, '中型混合武器', 'gunnery', '操作中型磁轨武器', 3, 'perception', 'memory'),
(15, '武器精准', 'gunnery', '提升所有武器命中率，每级+3%', 2, 'perception', 'memory');

-- 防御技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(20, '护盾管理', 'defense', '每级提升5%护盾容量', 2, 'intelligence', 'memory'),
(21, '护盾操作', 'defense', '每级提升5%护盾回充速度', 2, 'intelligence', 'memory'),
(22, '装甲强化', 'defense', '每级提升5%装甲总量', 2, 'intelligence', 'memory'),
(23, '损伤控制', 'defense', '每级提升3%全抗性', 3, 'intelligence', 'memory');

-- 工业技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(30, '采矿', 'industry', '每级提升5%采矿产出', 1, 'memory', 'intelligence'),
(31, '精炼', 'industry', '每级提升2%精炼效率', 2, 'memory', 'intelligence'),
(32, '制造', 'industry', '每级减少3%制造时间', 2, 'memory', 'intelligence'),
(33, '蓝图研究', 'industry', '每级减少5%研究时间', 3, 'memory', 'intelligence');

-- 导航技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(40, '加速控制', 'navigation', '每级提升5%亚光速速度', 1, 'intelligence', 'perception'),
(41, '跃迁操作', 'navigation', '每级减少5%跃迁对齐时间', 2, 'intelligence', 'perception'),
(42, '燃料节省', 'navigation', '每级减少5%跃迁电容消耗', 2, 'intelligence', 'perception');

-- 特殊技能
INSERT INTO skill_defs (id, name, category, description, rank, primary_attr, secondary_attr) VALUES
(50, '多线操控', 'special', '每级+1可同时操控角色数，底线效率+3%', 5, 'charisma', 'willpower'),
(51, '生物适配', 'special', '每级解锁更多植入体槽位', 4, 'intelligence', 'memory'),
(52, '克隆网络', 'special', '每级+1可部署的分支克隆数', 4, 'intelligence', 'memory'),
(53, '科研效率', 'research', '每级提升5%对组织科研的贡献量', 3, 'intelligence', 'memory'),
(54, '交易学', 'trade', '每级减少5%市场手续费', 2, 'charisma', 'memory');

-- 设定前置条件
UPDATE skill_defs SET prereq_skill_id=1, prereq_level=3 WHERE id=2;  -- 驱逐需要护卫III
UPDATE skill_defs SET prereq_skill_id=2, prereq_level=3 WHERE id=3;  -- 巡洋需要驱逐III
UPDATE skill_defs SET prereq_skill_id=3, prereq_level=3 WHERE id=4;  -- 战巡需要巡洋III
UPDATE skill_defs SET prereq_skill_id=4, prereq_level=4 WHERE id=5;  -- 战列需要战巡IV
UPDATE skill_defs SET prereq_skill_id=10, prereq_level=3 WHERE id=13; -- 中型能量需要小型III
UPDATE skill_defs SET prereq_skill_id=11, prereq_level=3 WHERE id=14; -- 中型混合需要小型III

-- ============ 植入体种子数据 ============
INSERT INTO implant_defs (id, name, slot_type, grade, description, effect_type, effect_value, set_name) VALUES
-- 大脑皮层槽
(1, '记忆增幅器 I', 'cortex', 'white', '提升技能训练速度', 'train_speed', 0.03, ''),
(2, '记忆增幅器 II', 'cortex', 'green', '提升技能训练速度', 'train_speed', 0.05, ''),
(3, '战术模拟器 I', 'cortex', 'green', '战斗中显示敌方下1Tick行动概率', 'combat_predict', 0.05, '战争之魂'),
(4, '科研加速器 I', 'cortex', 'blue', '提升组织科研贡献效率', 'research_speed', 0.08, ''),
-- 脊髓链路槽
(10, '反应加速器 I', 'spine', 'white', '减少跃迁对齐时间', 'align_speed', 0.03, ''),
(11, '精密操控体 I', 'spine', 'green', '提升武器追踪速度', 'tracking_speed', 0.05, '战争之魂'),
(12, '采矿优化体 I', 'spine', 'green', '减少采矿周期时间', 'mining_cycle', 0.05, '锻造之心'),
-- 基因编辑槽
(20, '体质强化序列 I', 'gene', 'white', '提升意识完整度恢复速度', 'consciousness_regen', 0.03, ''),
(21, '疲劳抑制因子 I', 'gene', 'green', '提升疲劳值恢复速度', 'fatigue_regen', 0.05, ''),
(22, '神经再生序列 I', 'gene', 'blue', '减少死亡时意识完整度损失', 'death_protection', 0.08, ''),
-- 神经接口槽
(30, '舰船融合接口 I', 'neural', 'blue', '与舰载AI协调度+，解锁高级自动化', 'ai_coordination', 0.08, ''),
(31, '意识广播接口 I', 'neural', 'purple', '指挥加成范围大幅提升', 'command_range', 0.15, '');
