package billing

import (
	"context"
	"errors"
	"fmt"
	"sync"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/secrets"
)

type Client struct {
	mu      sync.RWMutex
	sdk     *billingsdk.Client
	repo    domain.LicenseRepository
	vault   *secrets.Vault
	keyring *Keyring
}

func NewClient(repo domain.LicenseRepository, vault *secrets.Vault) (*Client, error) {
	if repo == nil {
		return nil, errors.New("billing: license repository required")
	}
	if vault == nil {
		return nil, errors.New("billing: secrets vault required")
	}
	keyring, err := NewKeyringFromEnv()
	if err != nil {
		return nil, err
	}
	c := &Client{
		sdk:     billingsdk.NewClient(BillingURL, ProductSlug, BillingSecret),
		repo:    repo,
		vault:   vault,
		keyring: keyring,
	}
	return c, nil
}

func (c *Client) SDK() *billingsdk.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sdk
}

func (c *Client) Keyring() *Keyring {
	return c.keyring
}

func (c *Client) LoadSession(ctx context.Context) error {
	license, err := c.repo.Get(ctx)
	if err != nil {
		return err
	}
	if license == nil || license.AccessTokenEnc == "" {
		return nil
	}
	token, err := c.vault.Decrypt(license.AccessTokenEnc)
	if err != nil {
		return fmt.Errorf("billing: decrypt token: %w", err)
	}
	c.mu.Lock()
	c.sdk.SetCustomerToken(token)
	c.mu.Unlock()
	return nil
}

func (c *Client) PersistAccessToken(ctx context.Context, token string) error {
	enc, err := c.vault.Encrypt(token)
	if err != nil {
		return fmt.Errorf("billing: encrypt token: %w", err)
	}
	license, err := c.repo.Get(ctx)
	if err != nil {
		return err
	}
	if license == nil {
		license = &domain.License{ID: "local", Status: "inactive", FeaturesJSON: "{}"}
	}
	license.AccessTokenEnc = enc
	c.mu.Lock()
	c.sdk.SetCustomerToken(token)
	c.mu.Unlock()
	return c.repo.Save(ctx, *license)
}

func (c *Client) HasCustomerToken() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sdk.CustomerToken != ""
}

func (c *Client) ClearSession(ctx context.Context) error {
	c.mu.Lock()
	c.sdk.SetCustomerToken("")
	c.mu.Unlock()
	return c.repo.Clear(ctx)
}
