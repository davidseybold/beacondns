package repository

import (
	"context"
	"encoding/json"

	"github.com/davidseybold/beacondns/internal/controller/domain"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BeaconDB struct {
	db postgres.PgxPool
}

var _ BeaconDBRepository = (*BeaconDB)(nil)

func NewBeaconDB(pgx postgres.PgxPool) (*BeaconDB, error) {
	return &BeaconDB{
		db: pgx,
	}, nil
}

func (b *BeaconDB) AddNameServer(ctx context.Context, ns *domain.NameServer) (*domain.NameServer, error) {
	txn, err := b.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback(ctx)

	row := txn.QueryRow(ctx, insertNameserverQuery, ns.ID, ns.Name, ns.RouteKey, ns.IPAddress)

	dbNs, err := scanNameServer(row)
	if err != nil {
		return nil, err
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}

	return dbNs, nil
}

func (b *BeaconDB) ListNameServers(ctx context.Context) ([]domain.NameServer, error) {
	rows, err := b.db.Query(ctx, listNameserversQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNameServers(rows)
}

func (b *BeaconDB) CreateZone(ctx context.Context, params CreateZoneParams) (*domain.ChangeInfo, error) {
	txn, err := b.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback(ctx)

	if err := insertDelegationSet(ctx, txn, params.DelegationSet); err != nil {
		return nil, err
	}

	if _, err := txn.Exec(ctx, insertZoneQuery, params.Zone.ID, params.Zone.Name, params.DelegationSet.ID); err != nil {
		return nil, err
	}

	if err := insertResourceRecordSets(ctx, txn, params.Zone.ID, []domain.ResourceRecordSet{*params.SOA, *params.NS}); err != nil {
		return nil, err
	}

	changeInfo, err := createZoneChange(ctx, txn, params.Change)
	if err != nil {
		return nil, err
	}

	for _, sync := range params.Syncs {
		if _, err := txn.Exec(ctx, insertZoneChangeSyncQuery, sync.ZoneChangeID, sync.NameServerID, sync.Status); err != nil {
			return nil, err
		}
	}

	for _, msg := range params.OutboxMessages {
		if err := insertOutboxMessage(ctx, txn, msg); err != nil {
			return nil, err
		}
	}

	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}

	return changeInfo, nil
}

func insertDelegationSet(ctx context.Context, q postgres.Queryer, ds *domain.DelegationSet) error {
	if _, err := q.Exec(ctx, insertDelegationSetQuery, ds.ID); err != nil {
		return err
	}

	for _, ns := range ds.NameServers {
		if _, err := q.Exec(ctx, insertDelegationSetNameServersQuery, ds.ID, ns.ID); err != nil {
			return err
		}
	}

	return nil
}

func insertResourceRecordSets(ctx context.Context, q postgres.Queryer, zoneID uuid.UUID, recordSets []domain.ResourceRecordSet) error {
	for _, recordSet := range recordSets {
		if err := insertResourceRecordSet(ctx, q, zoneID, &recordSet); err != nil {
			return err
		}
	}
	return nil
}

func insertResourceRecordSet(ctx context.Context, q postgres.Queryer, zoneID uuid.UUID, recordSet *domain.ResourceRecordSet) error {
	if _, err := q.Exec(ctx, insertResourceRecordSetQuery, recordSet.ID, zoneID, recordSet.Name, recordSet.Type, recordSet.TTL); err != nil {
		return err
	}

	for _, rr := range recordSet.ResourceRecords {
		if _, err := q.Exec(ctx, insertResourceRecordQuery, recordSet.ID, rr.Value); err != nil {
			return err
		}
	}

	return nil
}

func createZoneChange(ctx context.Context, q postgres.Queryer, change *domain.ZoneChange) (*domain.ChangeInfo, error) {
	row := q.QueryRow(ctx, insertZoneChangeQuery, change.ID, change.ZoneID, change.Action)

	changeInfo := &domain.ChangeInfo{
		ID:     change.ID,
		Status: domain.ChangeSyncStatusPending,
	}
	if err := row.Scan(&changeInfo.SubmittedAt); err != nil {
		return nil, err
	}

	for i, rrcChange := range change.Changes {
		values := getRecordSetValues(rrcChange.ResourceRecordSet)
		b, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}

		if _, err := q.Exec(
			ctx,
			insertResourceRecordSetChangeQuery,
			change.ID,
			rrcChange.Action,
			rrcChange.ResourceRecordSet.Name,
			rrcChange.ResourceRecordSet.Type,
			rrcChange.ResourceRecordSet.TTL,
			b, // record values
			i+1,
		); err != nil {
			return nil, err
		}
	}

	return changeInfo, nil
}

func getRecordSetValues(rrs domain.ResourceRecordSet) []string {
	values := make([]string, len(rrs.ResourceRecords))
	for i, rr := range rrs.ResourceRecords {
		values[i] = rr.Value
	}
	return values
}

func insertOutboxMessage(ctx context.Context, q postgres.Queryer, msg domain.OutboxMessage) error {
	if _, err := q.Exec(ctx, insertOutboxMessageQuery, msg.ID, msg.RouteKey, msg.Payload); err != nil {
		return err
	}
	return nil
}

func (b *BeaconDB) GetRandomNameServers(ctx context.Context, count int) ([]domain.NameServer, error) {
	rows, err := b.db.Query(ctx, selectRandomNameServersQuery, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNameServers(rows)
}

func scanNameServers(rows pgx.Rows) ([]domain.NameServer, error) {
	var nameServers []domain.NameServer
	for rows.Next() {
		ns, err := scanNameServer(rows)
		if err != nil {
			return nil, err
		}
		nameServers = append(nameServers, *ns)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nameServers, nil
}

func scanNameServer(row pgx.Row) (*domain.NameServer, error) {
	var ns domain.NameServer
	if err := row.Scan(&ns.ID, &ns.Name, &ns.RouteKey, &ns.IPAddress); err != nil {
		return nil, err
	}

	return &ns, nil
}
