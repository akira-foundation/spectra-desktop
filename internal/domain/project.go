package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusConnected    ProjectStatus = "connected"
	ProjectStatusDisconnected ProjectStatus = "disconnected"
	ProjectStatusSyncing      ProjectStatus = "syncing"
	ProjectStatusError        ProjectStatus = "error"
)

const (
	APIFilterModeAuto       = "auto"
	APIFilterModeMiddleware = "middleware"
	APIFilterModePrefix     = "prefix"
	APIFilterModeAll        = "all"
)

type Project struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Path                string        `json:"path"`
	Framework           string        `json:"framework"`
	FrameworkVersion    string        `json:"frameworkVersion"`
	Status              ProjectStatus `json:"status"`
	APIFilterMode       string        `json:"apiFilterMode"`
	APIFilterValue      string        `json:"apiFilterValue"`
	BaseURL             string        `json:"baseUrl"`
	LoginEndpointID     string        `json:"loginEndpointId,omitempty"`
	LoginTokenPath      string        `json:"loginTokenPath,omitempty"`
	LogoutEndpointID    string        `json:"logoutEndpointId,omitempty"`
	ActiveEnvironmentID string        `json:"activeEnvironmentId,omitempty"`
	CreatedAt           time.Time     `json:"createdAt"`
	UpdatedAt           time.Time     `json:"updatedAt"`
	LastSyncedAt        *time.Time    `json:"lastSyncedAt,omitempty"`
}

type ProjectInput struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Path             string `json:"path"`
	Framework        string `json:"framework"`
	FrameworkVersion string `json:"frameworkVersion"`
	APIFilterMode    string `json:"apiFilterMode"`
	APIFilterValue   string `json:"apiFilterValue"`
	BaseURL          string `json:"baseUrl"`
}
