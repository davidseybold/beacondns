package model

import "github.com/google/uuid"

type ServerType string

const (
	ServerTypeResolver ServerType = "resolver"
)

type Server struct {
	ID       uuid.UUID  `json:"id"`
	Type     ServerType `json:"type"`
	HostName string     `json:"hostName"`
}
