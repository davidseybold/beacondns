CREATE TABLE
    nameservers (
        id UUID PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        routing_key VARCHAR(255) NOT NULL,
        UNIQUE (name)
    );

CREATE TABLE
    delegation_sets (id UUID PRIMARY KEY);

CREATE TABLE
    delegation_set_nameservers (
        delegation_set_id UUID NOT NULL,
        nameserver_id UUID NOT NULL,
        PRIMARY KEY (delegation_set_id, nameserver_id),
        FOREIGN KEY (delegation_set_id) REFERENCES delegation_sets (id) ON DELETE CASCADE,
        FOREIGN KEY (nameserver_id) REFERENCES nameservers (id) ON DELETE CASCADE
    );

CREATE TABLE
    zones (
        id UUID PRIMARY KEY,
        delegation_set_id UUID,
        name VARCHAR(255) NOT NULL UNIQUE,
        is_private BOOLEAN NOT NULL DEFAULT TRUE,
        FOREIGN KEY (delegation_set_id) REFERENCES delegation_sets (id) ON DELETE CASCADE
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
    zone_changes (
        id UUID PRIMARY KEY,
        zone_id UUID NOT NULL,
        action TEXT NOT NULL CHECK (
            action IN ('CREATE_ZONE', 'UPDATE_ZONE', 'DELETE_ZONE')
        ),
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (zone_id) REFERENCES zones (id) ON DELETE CASCADE
    );

CREATE TABLE
    resource_record_set_changes (
        zone_change_id UUID NOT NULL,
        action TEXT NOT NULL CHECK (action IN ('CREATE', 'UPSERT', 'DELETE')),
        name TEXT NOT NULL,
        record_type TEXT NOT NULL,
        ttl INT,
        record_values JSONB,
        ordering INT,
        PRIMARY KEY (zone_change_id, name, record_type, ordering),
        FOREIGN KEY (zone_change_id) REFERENCES zone_changes (id) ON DELETE CASCADE
    );

CREATE TABLE
    zone_change_syncs (
        zone_change_id UUID NOT NULL,
        nameserver_id UUID NOT NULL,
        status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'INSYNC')),
        synced_at TIMESTAMPTZ,
        PRIMARY KEY (zone_change_id, nameserver_id),
        FOREIGN KEY (zone_change_id) REFERENCES zone_changes (id) ON DELETE CASCADE,
        FOREIGN KEY (nameserver_id) REFERENCES nameservers (id) ON DELETE CASCADE
    );

CREATE TABLE
    outbox (
        id UUID PRIMARY KEY,
        routing_key VARCHAR(255) NOT NULL,
        payload BYTEA NOT NULL,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );