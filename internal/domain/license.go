package domain

import (
	"context"
	"time"
)

type License struct {
	ID                string
	CustomerID        string
	CustomerEmail     string
	CustomerName      string
	AccessTokenEnc    string
	Plan              string
	Cycle             string
	Status            string
	ValidUntil        string
	ActivatedAt       string
	LastVerifiedAt    string
	LicenseKeyID      string
	LicenseAlgorithm  string
	LicensePayload    string
	LicenseSignature  string
	FeaturesJSON      string
	DeviceID          string
	CancelAtPeriodEnd bool
	CancelAt          string
	TargetPlan        string
	GracePeriod       bool
	UpdatedAt         time.Time
}

type UsageBufferEntry struct {
	ID         string
	Feature    string
	Amount     int
	OccurredAt time.Time
	Flushed    bool
	CreatedAt  time.Time
}

type LicenseRepository interface {
	Get(ctx context.Context) (*License, error)
	Save(ctx context.Context, license License) error
	Clear(ctx context.Context) error
}

type UsageBufferRepository interface {
	Append(ctx context.Context, entry UsageBufferEntry) error
	PendingBatch(ctx context.Context, limit int) ([]UsageBufferEntry, error)
	MarkFlushed(ctx context.Context, ids []string) error
}
