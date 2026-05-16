package model

import (
	"time"

	"github.com/uptrace/bun"
)

type License struct {
	bun.BaseModel `bun:"table:license"`

	ID                string    `bun:"id,pk"`
	CustomerID        string    `bun:"customer_id,notnull,default:''"`
	CustomerEmail     string    `bun:"customer_email,notnull,default:''"`
	CustomerName      string    `bun:"customer_name,notnull,default:''"`
	AccessTokenEnc    string    `bun:"access_token_enc,notnull,default:''"`
	Plan              string    `bun:"plan,notnull,default:''"`
	Cycle             string    `bun:"cycle,notnull,default:''"`
	Status            string    `bun:"status,notnull,default:'inactive'"`
	ValidUntil        string    `bun:"valid_until,notnull,default:''"`
	ActivatedAt       string    `bun:"activated_at,notnull,default:''"`
	LastVerifiedAt    string    `bun:"last_verified_at,notnull,default:''"`
	LicenseKeyID      string    `bun:"license_key_id,notnull,default:''"`
	LicenseAlgorithm  string    `bun:"license_algorithm,notnull,default:''"`
	LicensePayload    string    `bun:"license_payload,notnull,default:''"`
	LicenseSignature  string    `bun:"license_signature,notnull,default:''"`
	FeaturesJSON      string    `bun:"features_json,notnull,default:'{}'"`
	DeviceID          string    `bun:"device_id,notnull,default:''"`
	CancelAtPeriodEnd bool      `bun:"cancel_at_period_end,notnull,default:false"`
	CancelAt          string    `bun:"cancel_at,notnull,default:''"`
	TargetPlan        string    `bun:"target_plan,notnull,default:''"`
	GracePeriod       bool      `bun:"grace_period,notnull,default:false"`
	UpdatedAt         time.Time `bun:"updated_at,notnull"`
}

type UsageBufferEntry struct {
	bun.BaseModel `bun:"table:billing_usage_buffer"`

	ID         string    `bun:"id,pk"`
	Feature    string    `bun:"feature,notnull"`
	Amount     int       `bun:"amount,notnull,default:1"`
	OccurredAt time.Time `bun:"occurred_at,notnull"`
	Flushed    bool      `bun:"flushed,notnull,default:false"`
	CreatedAt  time.Time `bun:"created_at,notnull"`
}
