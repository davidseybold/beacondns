CREATE TABLE
    domain_lists (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    );

CREATE TABLE
    domain_list_domains (
        domain_list_id UUID NOT NULL,
        domain TEXT NOT NULL,
        PRIMARY KEY (domain_list_id, domain),
        FOREIGN KEY (domain_list_id) REFERENCES domain_lists (id) ON DELETE CASCADE
    );

CREATE TABLE
    firewall_rules (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name TEXT NOT NULL,
        domain_list_id UUID NOT NULL,
        action TEXT NOT NULL,
        block_response_type TEXT,
        block_response JSONB,
        priority INTEGER NOT NULL DEFAULT 0,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        UNIQUE (domain_list_id),
        FOREIGN KEY (domain_list_id) REFERENCES domain_lists (id) ON DELETE CASCADE
    )