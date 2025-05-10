CREATE TABLE
    servers (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    zones (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE
    );

CREATE TABLE
    resource_record_sets (
        id SERIAL PRIMARY KEY,
        zone_id UUID NOT NULL,
        name VARCHAR(255) NOT NULL,
        record_type VARCHAR(10) NOT NULL,
        ttl INT DEFAULT 3600,
        UNIQUE (zone_id, name, record_type),
        FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE CASCADE
    );

CREATE TABLE
    resource_records (
        resource_record_set_id INTEGER NOT NULL,
        value TEXT NOT NULL,
        FOREIGN KEY (resource_record_set_id) REFERENCES resource_record_sets (id) ON DELETE CASCADE
    );

CREATE TABLE
    changes (
        id UUID PRIMARY KEY,
        type TEXT NOT NULL CHECK (type IN ('ZONE')),
        data JSONB NOT NULL,
        submitted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE
    change_targets (
        id UUID PRIMARY KEY,
        change_id UUID NOT NULL,
        server_id UUID NOT NULL,
        status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'INSYNC')),
        updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        synced_at TIMESTAMPTZ,
        FOREIGN KEY (change_id) REFERENCES changes (id) ON DELETE CASCADE,
        FOREIGN KEY (server_id) REFERENCES servers (id) ON DELETE CASCADE
    );