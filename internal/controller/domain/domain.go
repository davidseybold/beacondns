package domain

import (
	"time"

	"github.com/google/uuid"

	beacondomain "github.com/davidseybold/beacondns/internal/domain"
)

type ChangeStatus string
type ChangeTargetStatus string
type ChangeType string

const (
	ChangeStatusPending ChangeStatus = "PENDING"
	ChangeStatusInSync  ChangeStatus = "INSYNC"

	ChangeTargetStatusPending ChangeTargetStatus = "PENDING"
	ChangeTargetStatusSent    ChangeTargetStatus = "SENT"
	ChangeTargetStatusInSync  ChangeTargetStatus = "INSYNC"

	ChangeTypeZone ChangeType = "ZONE"
)

type CreateZoneResult struct {
	Zone       Zone       `json:"zone"`
	ChangeInfo ChangeInfo `json:"changeInfo"`
}

type Zone struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ChangeWithTargets struct {
	beacondomain.Change
	Targets []ChangeTarget `json:"targets"`
}

type ChangeTarget struct {
	ID       uuid.UUID          `json:"id"`
	ChangeID uuid.UUID          `json:"changeId"`
	ServerID uuid.UUID          `json:"serverId"`
	Status   ChangeTargetStatus `json:"status"`
	SyncedAt *time.Time         `json:"syncedAt,omitempty"`
}

type ChangeInfo struct {
	ID          uuid.UUID    `json:"id"`
	Status      ChangeStatus `json:"status"`
	SubmittedAt time.Time    `json:"submittedAt"`
}

type Server struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
