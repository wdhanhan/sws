-- ============ 建筑系统 ============
CREATE TABLE IF NOT EXISTS building_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(30) NOT NULL, -- personal, corp, alliance, nation
    building_type VARCHAR(50) NOT NULL,
    description TEXT DEFAULT '',
    -- 建造需求
    build_time_hours INT NOT NULL DEFAULT 24,
    fuel_type_id BIGINT, -- 消耗的燃料物品ID
    fuel_per_hour INT NOT NULL DEFAULT 1,
    -- 属性
    shield_hp INT NOT NULL DEFAULT 1000,
    armor_hp INT NOT NULL DEFAULT 1000,
    structure_hp INT NOT NULL DEFAULT 500,
    -- 功能标记
    has_market BOOLEAN NOT NULL DEFAULT FALSE,
    has_refinery BOOLEAN NOT NULL DEFAULT FALSE,
    has_factory BOOLEAN NOT NULL DEFAULT FALSE,
    has_clone_bay BOOLEAN NOT NULL DEFAULT FALSE,
    has_repair BOOLEAN NOT NULL DEFAULT FALSE,
    has_defense BOOLEAN NOT NULL DEFAULT FALSE,
    defense_dps INT NOT NULL DEFAULT 0,
    cargo_capacity DOUBLE PRECISION NOT NULL DEFAULT 1000
);

-- 建造材料需求
CREATE TABLE IF NOT EXISTS building_materials (
    id BIGSERIAL PRIMARY KEY,
    building_def_id BIGINT NOT NULL REFERENCES building_defs(id),
    item_def_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1
);

-- 已建造的建筑实例
CREATE TABLE IF NOT EXISTS buildings (
    id BIGSERIAL PRIMARY KEY,
    building_def_id BIGINT NOT NULL REFERENCES building_defs(id),
    owner_type VARCHAR(20) NOT NULL, -- character, corp, alliance, nation
    owner_id BIGINT NOT NULL,
    owner_name VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    system_id BIGINT NOT NULL,
    pos_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    pos_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    pos_z DOUBLE PRECISION NOT NULL DEFAULT 0,
    -- 状态
    status VARCHAR(20) NOT NULL DEFAULT 'anchoring', -- anchoring, building, online, offline, destroyed
    build_progress INT NOT NULL DEFAULT 0, -- 0-100
    shield_current INT NOT NULL,
    armor_current INT NOT NULL,
    structure_current INT NOT NULL,
    fuel_remaining INT NOT NULL DEFAULT 0,
    is_powered BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ 奇遇系统 ============
CREATE TABLE IF NOT EXISTS encounter_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    encounter_type VARCHAR(20) NOT NULL, -- fragment, choice, event, epic
    description TEXT NOT NULL,
    -- 触发条件
    min_security DOUBLE PRECISION DEFAULT -1,
    max_security DOUBLE PRECISION DEFAULT 1,
    required_arm_id INT DEFAULT 0, -- 0=任意旋臂
    min_consciousness INT DEFAULT 0,
    base_probability DOUBLE PRECISION NOT NULL DEFAULT 0.05,
    cooldown_hours INT NOT NULL DEFAULT 6,
    -- 内容
    intro_text TEXT NOT NULL, -- 开场文字
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- 奇遇选项
CREATE TABLE IF NOT EXISTS encounter_choices (
    id BIGSERIAL PRIMARY KEY,
    encounter_id BIGINT NOT NULL REFERENCES encounter_defs(id),
    choice_index INT NOT NULL,
    choice_text VARCHAR(200) NOT NULL,
    result_text TEXT NOT NULL,
    -- 奖励/惩罚
    reward_item_id BIGINT DEFAULT 0,
    reward_quantity INT DEFAULT 0,
    reward_credits BIGINT DEFAULT 0,
    consciousness_change INT DEFAULT 0,
    standing_faction VARCHAR(50) DEFAULT '',
    standing_change DOUBLE PRECISION DEFAULT 0,
    trigger_combat_npc_id BIGINT DEFAULT 0, -- 触发战斗
    next_encounter_id BIGINT DEFAULT 0 -- 连锁奇遇
);

-- 角色奇遇记录
CREATE TABLE IF NOT EXISTS character_encounters (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    encounter_id BIGINT NOT NULL REFERENCES encounter_defs(id),
    choice_made INT,
    system_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ 自动化系统 ============
CREATE TABLE IF NOT EXISTS automation_plans (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    plan_name VARCHAR(100) NOT NULL,
    plan_type VARCHAR(30) NOT NULL, -- mining, combat, trade, patrol
    status VARCHAR(20) NOT NULL DEFAULT 'stopped', -- running, stopped, paused
    -- 配置(JSON)
    config JSONB NOT NULL DEFAULT '{}',
    -- 统计
    total_earned BIGINT NOT NULL DEFAULT 0,
    total_runs INT NOT NULL DEFAULT 0,
    last_run_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_buildings_system ON buildings(system_id);
CREATE INDEX idx_buildings_owner ON buildings(owner_type, owner_id);
CREATE INDEX idx_encounter_type ON encounter_defs(encounter_type);
CREATE INDEX idx_char_encounters ON character_encounters(character_id);
CREATE INDEX idx_automation_char ON automation_plans(character_id);

-- ============ 建筑种子数据 ============
INSERT INTO building_defs (id, name, category, building_type, description, build_time_hours, fuel_per_hour,
    shield_hp, armor_hp, structure_hp, has_market, has_refinery, has_factory, has_clone_bay, has_repair, has_defense, defense_dps, cargo_capacity) VALUES
-- 个人建筑
(1, '移动仓库', 'personal', 'storage', '太空中的私人储物箱', 2, 0, 500, 300, 200, false, false, false, false, false, false, 0, 5000),
(2, '个人前哨', 'personal', 'outpost', '小型停靠点：充能/换弹', 6, 1, 1000, 800, 500, false, false, false, false, true, false, 0, 2000),
(3, '采矿平台', 'personal', 'mining', '自动采矿设施', 8, 1, 800, 600, 400, false, false, false, false, false, false, 0, 10000),
(4, '观测哨', 'personal', 'scanner', '预警探测设施', 4, 0, 300, 200, 150, false, false, false, false, false, false, 0, 100),
(5, '隐蔽据点', 'personal', 'hidden', '隐形私人基地', 12, 2, 600, 400, 300, false, false, false, false, false, false, 0, 3000),
-- 军团建筑
(10, '工程站', 'corp', 'factory', '制造/精炼设施', 72, 5, 8000, 6000, 4000, false, true, true, false, false, false, 0, 50000),
(11, '军团船坞', 'corp', 'shipyard', '存放/维修大量舰船', 120, 8, 12000, 10000, 6000, false, false, false, false, true, false, 0, 100000),
(12, '精炼厂', 'corp', 'refinery', '高效矿石精炼', 72, 4, 6000, 5000, 3000, false, true, false, false, false, false, 0, 30000),
(13, '市场站', 'corp', 'market', '玩家市场', 168, 6, 10000, 8000, 5000, true, false, false, false, false, false, 0, 200000),
(14, '防御炮台阵列', 'corp', 'defense', '自动化防御', 48, 3, 5000, 8000, 4000, false, false, false, false, false, true, 200, 1000),
(15, '克隆湾', 'corp', 'clone', '克隆服务', 72, 4, 4000, 3000, 2000, false, false, false, true, false, false, 0, 5000),
-- 联盟建筑
(20, '空间站', 'alliance', 'station', '完整一体化空间站', 720, 20, 80000, 60000, 40000, true, true, true, true, true, true, 500, 1000000),
(21, '跳跃门', 'alliance', 'jump_gate', '两星系间永久通道', 336, 10, 20000, 15000, 10000, false, false, false, false, false, false, 0, 0),
(22, '超级船坞', 'alliance', 'super_shipyard', '建造无畏/航母级', 480, 15, 40000, 30000, 20000, false, false, true, false, true, false, 0, 500000),
-- 国家建筑
(30, '戴森云阵列', 'nation', 'dyson', '收集恒星能量', 2160, 0, 100000, 80000, 50000, false, false, false, false, false, false, 0, 0),
(31, '首都堡垒', 'nation', 'capital_fortress', '国家首都超级防御', 1440, 30, 200000, 150000, 100000, true, true, true, true, true, true, 2000, 5000000),
(32, '天穹实验室', 'nation', 'research_lab', '研究第7-9层科技', 2160, 25, 30000, 25000, 15000, false, false, false, false, false, false, 0, 50000);

-- ============ 奇遇种子数据 ============
INSERT INTO encounter_defs (id, name, encounter_type, description, min_security, max_security, base_probability, cooldown_hours, intro_text) VALUES
-- 碎片类
(1, '先驱者的低语', 'fragment', '接收到先驱者数据碎片', -1, 1, 0.08, 6,
 '你的扫描仪突然捕捉到一段微弱的加密信号。信号来自一个未知的频率，似乎是先驱者时代的数据编码。一个模糊的画面闪过你的意识接口——一座悬浮在黑洞边缘的巨大建筑。画面消失了，只留下一串坐标的残片。'),
(2, '漂浮的残骸', 'fragment', '发现一艘废弃飞船', -1, 0.5, 0.10, 6,
 '跃迁途中你的传感器探测到一个微弱的金属回波。靠近后发现是一艘严重损坏的飞船残骸，船体上的标志已经无法辨认。残骸的货仓似乎还有少量物品。'),
(3, '时空涟漪', 'fragment', '短暂的时空异常', -0.5, 0, 0.05, 12,
 '你的曲率场探测器突然发出警报——附近空间出现了微弱的时空涟漪。这种现象在先驱者崩溃后偶尔出现，被认为是天穹工程残留的回波。涟漪中似乎带有微弱的信息...'),
-- 选择类
(10, '遗弃运输舰', 'choice', '发现满载的运输舰', -1, 0.5, 0.06, 6,
 '你的传感器探测到一艘漂浮的运输舰。无生命迹象，但货仓似乎满载。舰体上有战斗痕迹，附近可能有海盗出没。你必须做出选择。'),
(11, '求救信号', 'choice', '接收到求救信号', -1, 1, 0.07, 6,
 '你收到了一个标准求救频率的信号。发送者自称是一名被海盗伏击的商人，请求护航到最近的空间站。但你无法确认这不是一个陷阱...'),
(12, '守墓者的提问', 'choice', '遭遇先驱者AI', -0.5, 0, 0.03, 24,
 '被击毁的守墓者核心发出最后一道光脉冲。你的意识接口被强制连接——一个苍老的AI意识出现在你脑海："后来者...你们为何而来？是为了力量？知识？还是...你也听到了那边的声音？"它在等你回答。'),
-- 事件类
(20, '维度裂隙涌动', 'event', '维度裂隙突然扩大', -0.5, 0, 0.04, 24,
 '你所在星系的空间结构突然开始震颤。一道维度裂隙在不远处撕开，维度入侵者开始涌出！附近的所有船只都面临威胁。你必须战斗或逃跑。'),
(21, '先驱遗迹激活', 'event', '附近遗迹突然自行启动', -1, 0, 0.02, 48,
 '你附近的一座先驱者遗迹建筑突然发出了蓝白色的光芒——它在启动！沉睡万年的先驱者系统正在苏醒。遗迹周围出现了强烈的能量场，守墓者也开始异常活跃...');

-- 奇遇选项
INSERT INTO encounter_choices (encounter_id, choice_index, choice_text, result_text, reward_item_id, reward_quantity, reward_credits) VALUES
-- 遗弃运输舰的选项
(10, 1, '靠近打捞货物', '你小心翼翼地靠近残骸，打开了货仓舱门...里面有一些有价值的矿石。', 1003, 50, 0),
(10, 2, '扫描后记录位置', '你记录下了残骸的精确坐标，可以稍后带人来。书签已保存。', 0, 0, 1000),
(10, 3, '不管它继续赶路', '你决定不冒险，继续你的旅程。', 0, 0, 0),
-- 求救信号的选项
(11, 1, '前往救援', '你赶到求救地点，发现确实是一名商人被困。你帮助他击退了海盗残余，他给了你一笔报酬。', 0, 0, 5000),
(11, 2, '要求预付报酬', '"先转账再说。"商人犹豫了一下，转给了你一笔星币。你前去救援。', 0, 0, 3000),
(11, 3, '忽略信号', '你关闭了通信频道，继续自己的事。不是所有人都能被拯救。', 0, 0, 0),
-- 守墓者的选项
(12, 1, '我来寻找力量', 'AI沉默片刻："力量...先驱者也曾如此说。"一束能量注入了你的舰船。', 5011, 1, 0),
(12, 2, '我来寻找知识', 'AI叹息："你比大多数人诚实。"一段加密数据传入了你的系统。', 0, 0, 10000),
(12, 3, '什么声音？', 'AI震动："你...也能听到？那不应该...这太早了..."一段异常频率记录被留下。', 0, 0, 0),
(12, 4, '切断连接', '你强行断开了意识接口。安全第一。', 0, 0, 0);
