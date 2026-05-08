package model

import (
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
)

type Project struct {
	bun.BaseModel `bun:"table:projects"`

	ID               string     `bun:"id,pk"`
	Name             string     `bun:"name,notnull"`
	Path             string     `bun:"path,notnull,unique"`
	Framework        string     `bun:"framework,notnull"`
	FrameworkVersion string     `bun:"framework_version,notnull,default:''"`
	Status           string     `bun:"status,notnull,default:'disconnected'"`
	APIFilterMode    string     `bun:"api_filter_mode,notnull,default:'auto'"`
	APIFilterValue   string     `bun:"api_filter_value,notnull,default:''"`
	BaseURL          string     `bun:"base_url,notnull,default:''"`
	CreatedAt        time.Time  `bun:"created_at,notnull"`
	UpdatedAt        time.Time  `bun:"updated_at,notnull"`
	LastSyncedAt     *time.Time `bun:"last_synced_at"`
}

func (p Project) ToDomain() domain.Project {
	return domain.Project{
		ID:               p.ID,
		Name:             p.Name,
		Path:             p.Path,
		Framework:        p.Framework,
		FrameworkVersion: p.FrameworkVersion,
		Status:           domain.ProjectStatus(p.Status),
		APIFilterMode:    p.APIFilterMode,
		APIFilterValue:   p.APIFilterValue,
		BaseURL:          p.BaseURL,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		LastSyncedAt:     p.LastSyncedAt,
	}
}

func FromDomain(p domain.Project) Project {
	return Project{
		ID:               p.ID,
		Name:             p.Name,
		Path:             p.Path,
		Framework:        p.Framework,
		FrameworkVersion: p.FrameworkVersion,
		Status:           string(p.Status),
		APIFilterMode:    p.APIFilterMode,
		APIFilterValue:   p.APIFilterValue,
		BaseURL:          p.BaseURL,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		LastSyncedAt:     p.LastSyncedAt,
	}
}
