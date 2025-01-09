package storage

import (
	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
)

// GitHubWebhookData presents `github` table fields.
type GitHubWebhookData struct {
	Payload    any
	Event      github.Event
	Target     string
	UserLogin  string
	DeliveryID uuid.UUID
	TargetID   uint64
	HookID     uint64
	UserID     int64
	Offset     int64
	Partition  int32
}
