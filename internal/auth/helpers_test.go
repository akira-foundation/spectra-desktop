package auth

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"testing"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/secrets"
)

func newVault(t *testing.T) *secrets.Vault {
	t.Helper()
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("rand key: %v", err)
	}
	v, err := secrets.NewVault(key)
	if err != nil {
		t.Fatalf("new vault: %v", err)
	}
	return v
}

func encrypt(t *testing.T, v *secrets.Vault, plain string) string {
	t.Helper()
	out, err := v.Encrypt(plain)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	return out
}

type fakeAccountRepo struct {
	mu     sync.Mutex
	stored map[string]domain.ProjectAccount
	saves  int
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{stored: map[string]domain.ProjectAccount{}}
}

func (f *fakeAccountRepo) List(_ context.Context, projectID string) ([]domain.ProjectAccount, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.ProjectAccount, 0, len(f.stored))
	for _, a := range f.stored {
		if a.ProjectID == projectID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeAccountRepo) Get(_ context.Context, id string) (*domain.ProjectAccount, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	a, ok := f.stored[id]
	if !ok {
		return nil, nil
	}
	return &a, nil
}

func (f *fakeAccountRepo) GetDefault(_ context.Context, projectID string) (*domain.ProjectAccount, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.stored {
		if a.ProjectID == projectID && a.IsDefault {
			return &a, nil
		}
	}
	return nil, nil
}

func (f *fakeAccountRepo) Save(_ context.Context, acc domain.ProjectAccount) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saves++
	f.stored[acc.ID] = acc
	return nil
}

func (f *fakeAccountRepo) SetDefault(_ context.Context, projectID, accountID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, a := range f.stored {
		if a.ProjectID == projectID {
			a.IsDefault = id == accountID
			f.stored[id] = a
		}
	}
	return nil
}

func (f *fakeAccountRepo) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.stored, id)
	return nil
}

func (f *fakeAccountRepo) saveCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.saves
}
