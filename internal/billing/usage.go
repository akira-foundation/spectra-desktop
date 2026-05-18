package billing

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"
	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

type UsageTracker struct {
	client      *Client
	repo        domain.UsageBufferRepository
	licenseRepo domain.LicenseRepository
	fingerprint string

	buffer *repoBuffer
	inner  *billingsdk.UsageTracker

	stopOnce sync.Once
	stopCh   chan struct{}
}

func NewUsageTracker(client *Client, repo domain.UsageBufferRepository, licenseRepo domain.LicenseRepository, fingerprint string) *UsageTracker {
	u := &UsageTracker{
		client:      client,
		repo:        repo,
		licenseRepo: licenseRepo,
		fingerprint: fingerprint,
		stopCh:      make(chan struct{}),
	}
	u.buffer = &repoBuffer{repo: repo}
	u.inner, _ = billingsdk.NewUsageTracker(billingsdk.TrackerOptions{
		Buffer:        u.buffer,
		Sync:          u.sync,
		Serial:        u.currentSerial,
		OnRefresh:     u.onRefresh,
		FlushInterval: 5 * time.Minute,
	})
	return u
}

func (u *UsageTracker) Track(ctx context.Context, feature string, amount int) error {
	if u.repo == nil || amount <= 0 {
		return nil
	}
	return u.repo.Append(ctx, domain.UsageBufferEntry{
		ID:         uuid.NewString(),
		Feature:    feature,
		Amount:     amount,
		OccurredAt: time.Now().UTC(),
	})
}

func (u *UsageTracker) StartFlusher(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-u.stopCh:
				return
			case <-ticker.C:
				if err := u.Flush(ctx); err != nil {
					log.Printf("usage flush: %v", err)
				}
			}
		}
	}()
}

func (u *UsageTracker) Stop() {
	u.stopOnce.Do(func() {
		close(u.stopCh)
	})
}

func (u *UsageTracker) Flush(ctx context.Context) error {
	if u.client == nil || u.repo == nil || u.fingerprint == "" {
		return nil
	}
	if !u.client.HasCustomerToken() {
		return nil
	}
	return u.inner.Flush(ctx)
}

func (u *UsageTracker) sync(ctx context.Context, deltas map[string]uint64, serial uint64) (*billingsdk.LicenseSyncUsageResponse, error) {
	resp, err := u.client.SDK().LicenseSyncUsage(ctx, billingsdk.LicenseSyncUsagePayload{
		Product:     ProductSlug,
		Fingerprint: u.fingerprint,
		Serial:      serial,
		Deltas:      deltas,
	})
	if err != nil {
		return nil, fmt.Errorf("billing: sync usage: %w", err)
	}
	return resp, nil
}

func (u *UsageTracker) currentSerial(ctx context.Context) (uint64, error) {
	if u.licenseRepo == nil {
		return 0, nil
	}
	license, err := u.licenseRepo.Get(ctx)
	if err != nil || license == nil || license.LicensePayload == "" {
		return 0, err
	}
	signed := billingsdk.SignedLicense{
		KeyID:     license.LicenseKeyID,
		Algorithm: license.LicenseAlgorithm,
		Payload:   license.LicensePayload,
		Signature: license.LicenseSignature,
	}
	decoded, err := billingsdk.DecodeLicense(signed)
	if err != nil {
		return 0, nil
	}
	return decoded.Payload.Serial, nil
}

func (u *UsageTracker) onRefresh(ctx context.Context, resp *billingsdk.LicenseSyncUsageResponse) error {
	if err := u.persistRefreshedSnapshot(ctx, resp); err != nil {
		return err
	}
	ids := u.buffer.takePendingIDs()
	if len(ids) == 0 {
		return nil
	}
	return u.repo.MarkFlushed(ctx, ids)
}

func (u *UsageTracker) persistRefreshedSnapshot(ctx context.Context, resp *billingsdk.LicenseSyncUsageResponse) error {
	if u.client == nil || u.client.keyring == nil {
		return nil
	}
	payload, err := u.client.keyring.Verify(resp.License)
	if err != nil {
		return err
	}
	license, err := u.licenseRepo.Get(ctx)
	if err != nil {
		return err
	}
	if license == nil {
		return nil
	}
	license.LicenseKeyID = resp.License.KeyID
	license.LicenseAlgorithm = resp.License.Algorithm
	license.LicensePayload = resp.License.Payload
	license.LicenseSignature = resp.License.Signature
	license.ValidUntil = payload.ValidUntil
	license.LastVerifiedAt = time.Now().UTC().Format(time.RFC3339)
	return u.licenseRepo.Save(ctx, *license)
}

type repoBuffer struct {
	repo domain.UsageBufferRepository

	mu         sync.Mutex
	pendingIDs []string
}

func (b *repoBuffer) Add(ctx context.Context, feature string, delta uint64) error {
	if delta == 0 {
		return nil
	}
	return b.repo.Append(ctx, domain.UsageBufferEntry{
		ID:         uuid.NewString(),
		Feature:    feature,
		Amount:     int(delta),
		OccurredAt: time.Now().UTC(),
	})
}

func (b *repoBuffer) Drain(ctx context.Context) (map[string]uint64, error) {
	rows, err := b.repo.PendingBatch(ctx, 500)
	if err != nil {
		return nil, err
	}
	deltas := map[string]uint64{}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Amount > 0 {
			deltas[row.Feature] += uint64(row.Amount)
		}
		ids = append(ids, row.ID)
	}
	b.mu.Lock()
	b.pendingIDs = ids
	b.mu.Unlock()
	return deltas, nil
}

func (b *repoBuffer) Restore(ctx context.Context, deltas map[string]uint64) error {
	b.mu.Lock()
	b.pendingIDs = nil
	b.mu.Unlock()
	return nil
}

func (b *repoBuffer) takePendingIDs() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	ids := b.pendingIDs
	b.pendingIDs = nil
	return ids
}
