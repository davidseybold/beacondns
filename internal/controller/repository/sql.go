package repository

const (
	insertNameserverQuery = "INSERT INTO nameservers (id, name, route_key, ip_address) VALUES ($1, $2, $3, $4) RETURNING id, name, route_key, ip_address;"
	listNameserversQuery  = "SELECT id, name, route_key, ip_address FROM nameservers ORDER BY name;"

	selectRandomNameServersQuery = "SELECT id, name, route_key, ip_address FROM nameservers ORDER BY RANDOM() LIMIT $1;"

	insertDelegationSetQuery            = "INSERT INTO delegation_sets (id) VALUES ($1) RETURNING id;"
	insertDelegationSetNameServersQuery = "INSERT INTO delegation_set_nameservers (delegation_set_id, nameserver_id) VALUES($1, $2);"

	insertZoneQuery = "INSERT INTO zones(id, name, delegation_set_id) VALUES ($1, $2, $3);"

	insertResourceRecordSetQuery = "INSERT INTO resource_record_sets (id, zone_id, name, record_type, ttl) VALUES ($1, $2, $3, $4, $5);"
	insertResourceRecordQuery    = "INSERT INTO resource_records (resource_record_set_id, value) VALUES ($1, $2);"

	// TODO: Update created_at to submitted_at
	insertZoneChangeQuery = `
	INSERT INTO zone_changes (id, zone_id, action)
	VALUES ($1, $2, $3) RETURNING created_at
	`
	insertResourceRecordSetChangeQuery = `
	INSERT INTO resource_record_set_changes (
    	zone_change_id, action, name, record_type, ttl, record_values, ordering
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	insertZoneChangeSyncQuery = `
	INSERT INTO zone_change_syncs (zone_change_id, nameserver_id, status)
	VALUES ($1, $2, $3)
	`

	insertOutboxMessageQuery = `
	INSERT INTO outbox (id, route_key, payload)
	VALUES ($1, $2, $3);
	`
)
