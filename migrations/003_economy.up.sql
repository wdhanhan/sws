-- 精炼配方表
CREATE TABLE IF NOT EXISTS refine_recipes (
    id BIGSERIAL PRIMARY KEY,
    input_item_id BIGINT NOT NULL,
    input_quantity INT NOT NULL DEFAULT 100,
    output_item_id BIGINT NOT NULL,
    output_quantity INT NOT NULL DEFAULT 1,
    skill_required VARCHAR(100) DEFAULT ''
);

-- 制造配方表
CREATE TABLE IF NOT EXISTS manufacture_recipes (
    id BIGSERIAL PRIMARY KEY,
    output_item_id BIGINT NOT NULL,
    output_quantity INT NOT NULL DEFAULT 1,
    build_time_sec INT NOT NULL DEFAULT 60,
    tech_level INT NOT NULL DEFAULT 1
);

-- 制造配方所需材料
CREATE TABLE IF NOT EXISTS manufacture_materials (
    id BIGSERIAL PRIMARY KEY,
    recipe_id BIGINT NOT NULL REFERENCES manufacture_recipes(id),
    item_def_id BIGINT NOT NULL,
    quantity INT NOT NULL DEFAULT 1
);

-- 蓝图表
CREATE TABLE IF NOT EXISTS blueprints (
    id BIGSERIAL PRIMARY KEY,
    owner_type VARCHAR(20) NOT NULL,
    owner_id BIGINT NOT NULL,
    recipe_id BIGINT NOT NULL REFERENCES manufacture_recipes(id),
    is_original BOOLEAN NOT NULL DEFAULT FALSE,
    runs_remaining INT NOT NULL DEFAULT -1,
    material_efficiency INT NOT NULL DEFAULT 0,
    time_efficiency INT NOT NULL DEFAULT 0,
    location_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 制造队列
CREATE TABLE IF NOT EXISTS manufacture_jobs (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    blueprint_id BIGINT NOT NULL REFERENCES blueprints(id),
    recipe_id BIGINT NOT NULL REFERENCES manufacture_recipes(id),
    status VARCHAR(20) NOT NULL DEFAULT 'running',
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    location_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 市场订单表
CREATE TABLE IF NOT EXISTS market_orders (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    item_def_id BIGINT NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    price BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    quantity_filled BIGINT NOT NULL DEFAULT 0,
    station_id BIGINT NOT NULL,
    system_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 市场交易历史
CREATE TABLE IF NOT EXISTS market_transactions (
    id BIGSERIAL PRIMARY KEY,
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    item_def_id BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    price BIGINT NOT NULL,
    total BIGINT NOT NULL,
    station_id BIGINT NOT NULL,
    system_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 采矿会话表
CREATE TABLE IF NOT EXISTS mining_sessions (
    id BIGSERIAL PRIMARY KEY,
    character_id BIGINT NOT NULL REFERENCES characters(id),
    belt_id BIGINT NOT NULL REFERENCES asteroid_belts(id),
    ore_item_id BIGINT NOT NULL,
    yield_per_cycle INT NOT NULL DEFAULT 100,
    cycle_time_sec INT NOT NULL DEFAULT 30,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_cycle_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_blueprints_owner ON blueprints(owner_type, owner_id);
CREATE INDEX idx_manufacture_jobs_char ON manufacture_jobs(character_id);
CREATE INDEX idx_manufacture_jobs_status ON manufacture_jobs(status);
CREATE INDEX idx_market_orders_item ON market_orders(item_def_id, status);
CREATE INDEX idx_market_orders_station ON market_orders(station_id, status);
CREATE INDEX idx_market_orders_char ON market_orders(character_id);
CREATE INDEX idx_market_tx_item ON market_transactions(item_def_id);
CREATE INDEX idx_mining_sessions_char ON mining_sessions(character_id, status);
