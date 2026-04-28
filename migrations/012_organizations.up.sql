-- ============ 军团 ============
CREATE TABLE IF NOT EXISTS corporations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    ticker VARCHAR(10) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    ceo_character_id BIGINT NOT NULL REFERENCES characters(id),
    member_count INT NOT NULL DEFAULT 1,
    tax_rate DOUBLE PRECISION NOT NULL DEFAULT 0.05,
    balance BIGINT NOT NULL DEFAULT 0,
    home_system_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 军团成员
CREATE TABLE IF NOT EXISTS corp_members (
    id BIGSERIAL PRIMARY KEY,
    corp_id BIGINT NOT NULL REFERENCES corporations(id) ON DELETE CASCADE,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    role VARCHAR(30) NOT NULL DEFAULT 'member', -- ceo, director, officer, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(character_id)
);

-- 军团仓库
CREATE TABLE IF NOT EXISTS corp_hangars (
    id BIGSERIAL PRIMARY KEY,
    corp_id BIGINT NOT NULL REFERENCES corporations(id) ON DELETE CASCADE,
    item_def_id BIGINT NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 0,
    station_id BIGINT NOT NULL
);

-- ============ 联盟 ============
CREATE TABLE IF NOT EXISTS alliances (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    ticker VARCHAR(10) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    leader_corp_id BIGINT NOT NULL REFERENCES corporations(id),
    member_corp_count INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 联盟成员军团
CREATE TABLE IF NOT EXISTS alliance_members (
    id BIGSERIAL PRIMARY KEY,
    alliance_id BIGINT NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    corp_id BIGINT NOT NULL REFERENCES corporations(id),
    role VARCHAR(30) NOT NULL DEFAULT 'member', -- executor, diplomat, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(corp_id)
);

-- ============ 国家 ============
CREATE TABLE IF NOT EXISTS nations (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    leader_alliance_id BIGINT NOT NULL REFERENCES alliances(id),
    government_type VARCHAR(30) NOT NULL DEFAULT 'republic',
    -- 体制追踪维度
    power_centralization DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    military_tendency DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    economic_freedom DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    research_priority DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    ideology_control DOUBLE PRECISION NOT NULL DEFAULT 0.1,
    diplomacy_openness DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    -- 效果
    research_bonus DOUBLE PRECISION NOT NULL DEFAULT 0,
    military_bonus DOUBLE PRECISION NOT NULL DEFAULT 0,
    economy_bonus DOUBLE PRECISION NOT NULL DEFAULT 0,
    capital_system_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 国家成员联盟
CREATE TABLE IF NOT EXISTS nation_members (
    id BIGSERIAL PRIMARY KEY,
    nation_id BIGINT NOT NULL REFERENCES nations(id) ON DELETE CASCADE,
    alliance_id BIGINT NOT NULL REFERENCES alliances(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(alliance_id)
);

-- ============ 主权 ============
CREATE TABLE IF NOT EXISTS sovereignty (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL REFERENCES star_systems(id),
    owner_type VARCHAR(20) NOT NULL, -- corp, alliance, nation
    owner_id BIGINT NOT NULL,
    owner_name VARCHAR(100) NOT NULL,
    tcu_level INT NOT NULL DEFAULT 1,
    tcu_shield INT NOT NULL DEFAULT 10000,
    tcu_armor INT NOT NULL DEFAULT 8000,
    tcu_structure INT NOT NULL DEFAULT 5000,
    tcu_shield_current INT NOT NULL DEFAULT 10000,
    tcu_armor_current INT NOT NULL DEFAULT 8000,
    tcu_structure_current INT NOT NULL DEFAULT 5000,
    reinforce_timer TIMESTAMP WITH TIME ZONE,
    vulnerable_start TIME,  -- 防方设定的脆弱窗口开始时间
    vulnerable_end TIME,
    claimed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(system_id)
);

-- ============ 战争 ============
CREATE TABLE IF NOT EXISTS wars (
    id BIGSERIAL PRIMARY KEY,
    attacker_type VARCHAR(20) NOT NULL,
    attacker_id BIGINT NOT NULL,
    attacker_name VARCHAR(100) NOT NULL,
    defender_type VARCHAR(20) NOT NULL,
    defender_id BIGINT NOT NULL,
    defender_name VARCHAR(100) NOT NULL,
    target_system_id BIGINT NOT NULL REFERENCES star_systems(id),
    status VARCHAR(20) NOT NULL DEFAULT 'preparing', -- preparing, active, siege, vulnerable, finished, cancelled
    deposit BIGINT NOT NULL DEFAULT 0,
    phase VARCHAR(20) NOT NULL DEFAULT 'preparation',
    preparation_ends TIMESTAMP WITH TIME ZONE,
    siege_started TIMESTAMP WITH TIME ZONE,
    vulnerable_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    winner VARCHAR(20), -- attacker, defender, draw
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============ 聊天/邮件 ============
CREATE TABLE IF NOT EXISTS chat_messages (
    id BIGSERIAL PRIMARY KEY,
    channel VARCHAR(50) NOT NULL, -- local:systemID, corp:corpID, alliance:allianceID, trade, private:charID
    sender_id BIGINT NOT NULL REFERENCES characters(id),
    sender_name VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS mails (
    id BIGSERIAL PRIMARY KEY,
    from_id BIGINT NOT NULL REFERENCES characters(id),
    from_name VARCHAR(50) NOT NULL,
    to_id BIGINT NOT NULL REFERENCES characters(id),
    subject VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 合同
CREATE TABLE IF NOT EXISTS contracts (
    id BIGSERIAL PRIMARY KEY,
    issuer_id BIGINT NOT NULL REFERENCES characters(id),
    assignee_id BIGINT,
    contract_type VARCHAR(30) NOT NULL, -- item_exchange, courier, auction, bounty, loan
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    title VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    price BIGINT NOT NULL DEFAULT 0,
    collateral BIGINT NOT NULL DEFAULT 0,
    reward BIGINT NOT NULL DEFAULT 0,
    system_id BIGINT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 好友/观察列表
CREATE TABLE IF NOT EXISTS contacts (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    target_id BIGINT NOT NULL REFERENCES characters(id),
    relation VARCHAR(20) NOT NULL DEFAULT 'friend', -- friend, blocked, watchlist
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(character_id, target_id)
);

-- 声望
CREATE TABLE IF NOT EXISTS standings (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    faction_type VARCHAR(20) NOT NULL, -- npc_race, corp, alliance
    faction_id BIGINT NOT NULL,
    standing DOUBLE PRECISION NOT NULL DEFAULT 0, -- -10.0 to 10.0
    UNIQUE(character_id, faction_type, faction_id)
);

CREATE INDEX idx_corp_members_corp ON corp_members(corp_id);
CREATE INDEX idx_alliance_members_alliance ON alliance_members(alliance_id);
CREATE INDEX idx_sovereignty_system ON sovereignty(system_id);
CREATE INDEX idx_wars_status ON wars(status);
CREATE INDEX idx_chat_channel ON chat_messages(channel, created_at);
CREATE INDEX idx_mails_to ON mails(to_id, is_read);
CREATE INDEX idx_contracts_status ON contracts(status, system_id);
CREATE INDEX idx_contacts_char ON contacts(character_id);
CREATE INDEX idx_standings_char ON standings(character_id);
