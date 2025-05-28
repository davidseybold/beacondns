CREATE TABLE
    response_policies (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        name TEXT NOT NULL,
        description TEXT,
        priority INTEGER NOT NULL DEFAULT 0,
        enabled BOOLEAN NOT NULL DEFAULT TRUE,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
    );

CREATE TABLE
    response_policy_rules (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
        response_policy_id UUID NOT NULL,
        name TEXT NOT NULL,
        trigger_type TEXT NOT NULL CHECK (
            trigger_type IN ('QNAME', 'IP', 'CLIENT_IP', 'NSDNAME', 'NSIP')
        ),
        trigger_value TEXT NOT NULL,
        action_type TEXT NOT NULL CHECK (
            action_type IN ('NXDOMAIN', 'NODATA', 'PASSTHRU', 'LOCALDATA')
        ),
        local_data JSONB,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        updated_at TIMESTAMPTZ NOT NULL DEFAULT now (),
        UNIQUE (response_policy_id, trigger_type, trigger_value),
        FOREIGN KEY (response_policy_id) REFERENCES response_policies (id) ON DELETE CASCADE
    );