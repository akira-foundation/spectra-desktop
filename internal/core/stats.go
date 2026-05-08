package core

import "context"

type StatCardKind string

const (
	StatRoutes       StatCardKind = "routes"
	StatControllers  StatCardKind = "controllers"
	StatMiddleware   StatCardKind = "middleware"
	StatModels       StatCardKind = "models"
	StatFormRequests StatCardKind = "form_requests"
	StatJobs         StatCardKind = "jobs"
	StatMailers      StatCardKind = "mailers"
	StatServices     StatCardKind = "services"
	StatErrors       StatCardKind = "errors"
)

type StatCard struct {
	Key   string       `json:"key"`
	Kind  StatCardKind `json:"kind"`
	Label string       `json:"label"`
	Value int          `json:"value"`
	Hint  string       `json:"hint,omitempty"`
}

type StatsReport struct {
	Cards []StatCard `json:"cards"`
}

type StatsCapable interface {
	Stats(ctx context.Context, projectPath string) (StatsReport, error)
}
