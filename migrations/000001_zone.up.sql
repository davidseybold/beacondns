CREATE TABLE
    zones (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL UNIQUE
    );

CREATE TABLE
    resource_record_sets (
        id UUID PRIMARY KEY,
        zone_id UUID NOT NULL,
        name VARCHAR(255) NOT NULL,
        record_type VARCHAR(10) NOT NULL,
        ttl INT DEFAULT 3600,
        UNIQUE (zone_id, name, record_type),
        FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE CASCADE
    );

CREATE TABLE
    resource_records (
        resource_record_set_id UUID NOT NULL,
        value TEXT NOT NULL,
        FOREIGN KEY (resource_record_set_id) REFERENCES resource_record_sets (id) ON DELETE CASCADE
    );

CREATE TABLE
    changes (
        id UUID PRIMARY KEY,
        zone_id UUID NOT NULL,
        actions JSONB NOT NULL,
        status TEXT NOT NULL CHECK (status IN ('PENDING', 'DONE')),
        submitted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE CASCADE
    );

CREATE TABLE
    events (
        id UUID PRIMARY KEY,
        type TEXT NOT NULL,
        payload JSONB NOT NULL,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        lock_expires TIMESTAMPTZ
    );