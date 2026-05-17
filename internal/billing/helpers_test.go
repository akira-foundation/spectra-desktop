package billing

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"testing"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

type stubLicenseRepo struct {
	license *domain.License
	getErr  error
	saveErr error
}

func (s *stubLicenseRepo) Get(ctx context.Context) (*domain.License, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.license == nil {
		return nil, nil
	}
	cp := *s.license
	return &cp, nil
}

func (s *stubLicenseRepo) Save(ctx context.Context, license domain.License) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	cp := license
	s.license = &cp
	return nil
}

func (s *stubLicenseRepo) Clear(ctx context.Context) error {
	s.license = nil
	return nil
}

type stubUsageRepo struct {
	rows    []domain.UsageBufferEntry
	appendE error
	pendErr error
	flushed []string
}

func (s *stubUsageRepo) Append(ctx context.Context, entry domain.UsageBufferEntry) error {
	if s.appendE != nil {
		return s.appendE
	}
	s.rows = append(s.rows, entry)
	return nil
}

func (s *stubUsageRepo) PendingBatch(ctx context.Context, limit int) ([]domain.UsageBufferEntry, error) {
	if s.pendErr != nil {
		return nil, s.pendErr
	}
	out := make([]domain.UsageBufferEntry, 0, len(s.rows))
	for _, r := range s.rows {
		if !r.Flushed {
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (s *stubUsageRepo) MarkFlushed(ctx context.Context, ids []string) error {
	s.flushed = append(s.flushed, ids...)
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	for i := range s.rows {
		if wanted[s.rows[i].ID] {
			s.rows[i].Flushed = true
		}
	}
	return nil
}

type signedHarness struct {
	pub  ed25519.PublicKey
	priv ed25519.PrivateKey
}

func newSignedHarness(t *testing.T) signedHarness {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return signedHarness{pub: pub, priv: priv}
}

func (h signedHarness) sign(t *testing.T, payload billingsdk.LicenseSnapshotPayload) billingsdk.SignedLicense {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	sig := ed25519.Sign(h.priv, raw)
	return billingsdk.SignedLicense{
		KeyID:      "test",
		Algorithm:  "ed25519",
		Payload:    base64.StdEncoding.EncodeToString(raw),
		Signature:  base64.StdEncoding.EncodeToString(sig),
		ValidUntil: payload.ValidUntil,
	}
}
