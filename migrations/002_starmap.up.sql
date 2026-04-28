CREATE TABLE IF NOT EXISTS star_systems (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    arm_id INT NOT NULL,
    coord_x DOUBLE PRECISION NOT NULL,
    coord_y DOUBLE PRECISION NOT NULL,
    coord_z DOUBLE PRECISION NOT NULL,
    security_level DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    star_type INT NOT NULL,
    planet_count INT NOT NULL DEFAULT 0,
    belt_count INT NOT NULL DEFAULT 0,
    has_anomaly BOOLEAN NOT NULL DEFAULT FALSE,
    has_ruins BOOLEAN NOT NULL DEFAULT FALSE,
    owner_id BIGINT,
    seed BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stargates (
    id BIGSERIAL PRIMARY KEY,
    from_system_id BIGINT NOT NULL REFERENCES star_systems(id),
    to_system_id BIGINT NOT NULL REFERENCES star_systems(id),
    is_natural BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS planets (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL REFERENCES star_systems(id),
    name VARCHAR(50) NOT NULL,
    planet_type VARCHAR(50) NOT NULL,
    orbit_au DOUBLE PRECISION NOT NULL,
    moon_count INT NOT NULL DEFAULT 0,
    has_station BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS asteroid_belts (
    id BIGSERIAL PRIMARY KEY,
    system_id BIGINT NOT NULL REFERENCES star_systems(id),
    name VARCHAR(50) NOT NULL,
    belt_type VARCHAR(50) NOT NULL,
    orbit_au DOUBLE PRECISION NOT NULL,
    remaining_pct INT NOT NULL DEFAULT 100
);

CREATE TABLE IF NOT EXISTS item_defs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    volume DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    base_price BIGINT NOT NULL DEFAULT 0,
    stackable BOOLEAN NOT NULL DEFAULT TRUE,
    tech_level INT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS inventory (
    id BIGSERIAL PRIMARY KEY,
    owner_type VARCHAR(20) NOT NULL,
    owner_id BIGINT NOT NULL,
    item_def_id BIGINT NOT NULL REFERENCES item_defs(id),
    quantity BIGINT NOT NULL DEFAULT 1,
    location_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stargates_from ON stargates(from_system_id);
CREATE INDEX idx_stargates_to ON stargates(to_system_id);
CREATE INDEX idx_planets_system ON planets(system_id);
CREATE INDEX idx_belts_system ON asteroid_belts(system_id);
CREATE INDEX idx_inventory_owner ON inventory(owner_type, owner_id);
CREATE INDEX idx_inventory_location ON inventory(location_id);
