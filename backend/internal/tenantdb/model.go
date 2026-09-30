package tenantdb

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusVerified Status = "verified"
	StatusReady Status = "ready"
	StatusError Status = "error"
)

type Connection struct {
	TenantID int64
	Host string
	Port string
	DatabaseName string
	Username string
	EncryptedPassword string
	Status Status
	VerifiedAt *time.Time
	LastError *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Credentials struct {
	Host string
	Port string
	Database string
	Username string
	Password string
}
