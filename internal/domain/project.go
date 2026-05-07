package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusConnected    ProjectStatus = "connected"
	ProjectStatusDisconnected ProjectStatus = "disconnected"
	ProjectStatusSyncing      ProjectStatus = "syncing"
	ProjectStatusError        ProjectStatus = "error"
)

type Project struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Path             string        `json:"path"`
	Framework        string        `json:"framework"`
	FrameworkVersion string        `json:"frameworkVersion"`
	Status           ProjectStatus `json:"status"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	LastSyncedAt     *time.Time    `json:"lastSyncedAt,omitempty"`
}

type ProjectInput struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	Framework        string `json:"framework"`
	FrameworkVersion string `json:"frameworkVersion"`
}
