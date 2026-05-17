package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"spectra-desktop/internal/auth"
	"spectra-desktop/internal/billing"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/drivers/laravel"
	"spectra-desktop/internal/httpclient"
	"spectra-desktop/internal/mock"
	"spectra-desktop/internal/repository"
	"spectra-desktop/internal/secrets"
	"spectra-desktop/internal/storage"
	"spectra-desktop/internal/watcher"
	"spectra-desktop/internal/workspace"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	scanner     *core.Scanner
	workspace   *workspace.Service
	storage     *storage.Storage
	projects    domain.ProjectRepository
	settings    domain.SettingsRepository
	endpoints   domain.EndpointRepository
	auth        domain.AuthRepository
	accounts    domain.AccountRepository
	mockRepo    domain.MockRepository
	mock        *mock.Manager
	vault       *secrets.Vault
	authResolve *auth.Resolver
	licenseRepo domain.LicenseRepository
	usageBuffer domain.UsageBufferRepository
	billing     *billing.Client
	billingGate *billing.Gate
	usage       *billing.UsageTracker
	machineID   *billing.MachineIdentity
	history     domain.HistoryRepository
	envs        domain.EnvironmentRepository
	snapshots   domain.SnapshotRepository
	tests       domain.TestRepository
	captures    domain.CaptureRepository
	captured    *capturedStore
	collections domain.CollectionRepository
	datasets    *repository.DatasetRepository
	scratch     *repository.ScratchRepository
	metrics     *repository.MetricsRepository
	http        *httpclient.Client
	watcher     *watcher.Watcher
}

func New() (*App, error) {
	scanner := core.NewScanner()
	scanner.Register(laravel.New())

	if applied, err := storage.ApplyPendingRestoreIfAny(); err != nil {
		log.Printf("apply pending restore: %v", err)
	} else if applied {
		log.Printf("storage: pending restore applied")
	}

	store := storage.New()
	if err := store.Open(""); err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrate storage: %w", err)
	}

	a := &App{
		scanner:     scanner,
		workspace:   workspace.NewService(),
		storage:     store,
		projects:    repository.NewProjectRepository(store.DB),
		settings:    repository.NewSettingsRepository(store.DB),
		endpoints:   repository.NewEndpointRepository(store.DB),
		auth:        repository.NewAuthRepository(store.DB),
		accounts:    repository.NewAccountRepository(store.DB),
		mockRepo:    repository.NewMockRepository(store.DB),
		licenseRepo: repository.NewLicenseRepository(store.DB),
		usageBuffer: repository.NewUsageBufferRepository(store.DB),
		history:     repository.NewHistoryRepository(store.DB),
		envs:        repository.NewEnvironmentRepository(store.DB),
		snapshots:   repository.NewSnapshotRepository(store.DB),
		tests:       repository.NewTestRepository(store.DB),
		captures:    repository.NewCaptureRepository(store.DB),
		collections: repository.NewCollectionRepository(store.DB),
		datasets:    repository.NewDatasetRepository(store.DB),
		scratch:     repository.NewScratchRepository(store.DB),
		metrics:     repository.NewMetricsRepository(store.DB),
		http:        httpclient.New(),
		watcher:     watcher.New(),
	}
	a.captured = newCapturedStore(repository.NewCapturedValuesRepository(store.DB), func() context.Context {
		if a.ctx != nil {
			return a.ctx
		}
		return context.Background()
	})

	vault, err := secrets.Default()
	if err != nil {
		log.Printf("secrets vault init failed (continuing without encryption): %v", err)
	} else {
		a.vault = vault
		a.authResolve = auth.NewResolver(a.accounts, vault)
	}

	a.mock = mock.NewManager(a.endpoints, a.history, a.mockRepo, a.newWailsEventEmitter())

	if billing.IsConfigured() && a.vault != nil {
		billingClient, err := billing.NewClient(a.licenseRepo, a.vault)
		if err != nil {
			log.Printf("billing init failed (continuing without): %v", err)
		} else {
			a.billing = billingClient
			a.billingGate = billing.NewGate(billingClient, a.usageBuffer)
		}
	}

	configDir, err := os.UserConfigDir()
	if err == nil {
		identity, err := billing.GetOrCreateMachineIdentity(configDir)
		if err != nil {
			log.Printf("machine identity init failed: %v", err)
		} else {
			a.machineID = identity
			if a.billing != nil {
				a.usage = billing.NewUsageTracker(a.billing, a.usageBuffer, a.licenseRepo, identity.ID)
			}
		}
	}

	return a, nil
}

func (a *App) emitUpsell(feature string, err error) {
	if a.ctx == nil {
		return
	}
	denied, ok := err.(*billing.GateDenied)
	payload := map[string]any{"feature": feature}
	if ok {
		payload["reason"] = denied.Reason
		payload["plan"] = denied.Plan
	}
	runtime.EventsEmit(a.ctx, "billing:upsell-required", payload)
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.applyPHPBinaryOverrideFromSettings()
	go a.migrateAuthRolesIfNeeded()
	if a.billing != nil {
		if err := a.billing.LoadSession(ctx); err != nil {
			log.Printf("billing load session: %v", err)
		}
	}
	if a.usage != nil {
		a.usage.StartFlusher(ctx, 5*time.Minute)
	}
}

func (a *App) applyPHPBinaryOverrideFromSettings() {
	if a.settings == nil {
		return
	}
	value, err := a.settings.Get(a.ctx, domain.SettingPHPBinaryPath)
	if err != nil {
		return
	}
	laravel.SetPHPBinaryOverride(value)
}

func (a *App) migrateAuthRolesIfNeeded() {
	projects, err := a.projects.List(a.ctx)
	if err != nil {
		return
	}
	for _, p := range projects {
		eps, err := a.endpoints.List(a.ctx, p.ID)
		if err != nil || len(eps) == 0 {
			continue
		}
		hasRole := false
		for _, e := range eps {
			if e.AuthRole != "" {
				hasRole = true
				break
			}
		}
		if hasRole {
			continue
		}
		log.Printf("auth migrate: rescanning project %s (%s)", p.Name, p.ID)
		if _, err := a.ScanWorkspace(p.ID); err != nil {
			log.Printf("auth migrate scan failed %s: %v", p.ID, err)
		}
	}
}

func (a *App) Shutdown(_ context.Context) {
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			log.Printf("close storage: %v", err)
		}
	}
}
