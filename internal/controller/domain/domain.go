package domain

import (
	"time"

	"github.com/google/uuid"

	beacondomain "github.com/davidseybold/beacondns/internal/domain"
)

type ChangeStatus string
type ChangeTargetStatus string
type ChangeType string
type ServerType string

const (
	ChangeStatusPending ChangeStatus = "PENDING"
	ChangeStatusInSync  ChangeStatus = "INSYNC"

	ChangeTargetStatusPending ChangeTargetStatus = "PENDING"
	ChangeTargetStatusSent    ChangeTargetStatus = "SENT"
	ChangeTargetStatusInSync  ChangeTargetStatus = "INSYNC"

	ChangeTypeZone ChangeType = "ZONE"

	ServerTypeResolver ServerType = "resolver"
)

type CreateZoneResult struct {
	Zone       Zone       `json:"zone"`
	ChangeInfo ChangeInfo `json:"changeInfo"`
}

type Zone struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func NewZone(name string) Zone {
	return Zone{
		ID:   uuid.New(),
		Name: name,
	}
}

type ChangeWithTargets struct {
	beacondomain.Change
	Targets []ChangeTarget `json:"targets"`
}

func NewChangeWithTargets(change beacondomain.Change, targets []ChangeTarget) ChangeWithTargets {
	return ChangeWithTargets{
		Change:  change,
		Targets: targets,
	}
}

type ChangeTarget struct {
	Server   Server             `json:"server"`
	Status   ChangeTargetStatus `json:"status"`
	SyncedAt *time.Time         `json:"syncedAt,omitempty"`
}

type ChangeInfo struct {
	ID          uuid.UUID    `json:"id"`
	Status      ChangeStatus `json:"status"`
	SubmittedAt time.Time    `json:"submittedAt"`
}

type Server struct {
	ID       uuid.UUID  `json:"id"`
	Type     ServerType `json:"type"`
	HostName string     `json:"hostName"`
}
