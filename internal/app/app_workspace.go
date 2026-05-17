package app

import (
	"fmt"
	"log"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"strings"
)

func (a *App) DeleteProject(id string) error {
	return a.projects.Delete(a.ctx, id)
}

func (a *App) MarkProjectSynced(id string) error {
	return a.projects.MarkSynced(a.ctx, id)
}

func (a *App) GetActiveProjectID() (string, error) {
	return a.settings.Get(a.ctx, domain.SettingActiveProjectID)
}

func (a *App) SetActiveProjectID(id string) error {
	if id == "" {
		return a.settings.Delete(a.ctx, domain.SettingActiveProjectID)
	}
	return a.settings.Set(a.ctx, domain.SettingActiveProjectID, id)
}

func (a *App) ScanWorkspace(projectID string) ([]core.Endpoint, error) {
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	driver, _, err := a.scanner.Resolve(project.Path)
	if err != nil {
		return nil, err
	}
	all, err := driver.Scan(a.ctx, project.Path)
	if err != nil {
		return nil, err
	}
	filtered := core.ApplyFilter(all, project.APIFilterMode, project.APIFilterValue).Endpoints
	if err := a.endpoints.Replace(a.ctx, projectID, filtered); err != nil {
		return nil, fmt.Errorf("persist endpoints: %w", err)
	}
	if err := a.recordSnapshot(projectID, filtered); err != nil {
		log.Printf("record snapshot: %v", err)
	}
	if err := a.projects.MarkSynced(a.ctx, projectID); err != nil {
		log.Printf("mark synced: %v", err)
	}
	if project.LoginEndpointID == "" || project.LogoutEndpointID == "" {
		stored, err := a.endpoints.List(a.ctx, projectID)
		if err == nil {
			loginID := project.LoginEndpointID
			logoutID := project.LogoutEndpointID
			if loginID == "" {
				loginID = pickAuthEndpoint(stored, core.AuthRoleLogin)
			}
			if logoutID == "" {
				logoutID = pickAuthEndpoint(stored, core.AuthRoleLogout)
			}
			if loginID != project.LoginEndpointID || logoutID != project.LogoutEndpointID {
				if err := a.projects.UpdateAuthRoutes(a.ctx, projectID, loginID, logoutID, project.LoginTokenPath); err != nil {
					log.Printf("auto-set auth routes: %v", err)
				}
			}
		}
	}
	return filtered, nil
}

func pickAuthEndpoint(endpoints []core.Endpoint, target core.AuthRole) string {
	type candidate struct {
		id    string
		score int
	}
	var best candidate
	for _, ep := range endpoints {
		if ep.AuthRole != target {
			continue
		}
		score := 1
		path := strings.ToLower(ep.Path)
		switch target {
		case core.AuthRoleLogin:
			switch {
			case strings.HasSuffix(path, "/login"):
				score += 5
			case strings.Contains(path, "/login"):
				score += 3
			case strings.Contains(path, "/signin") || strings.Contains(path, "/sign-in"):
				score += 2
			case strings.Contains(path, "/authenticate") || strings.Contains(path, "/auth/token"):
				score += 2
			}
		case core.AuthRoleLogout:
			switch {
			case strings.HasSuffix(path, "/logout"):
				score += 5
			case strings.Contains(path, "/logout"):
				score += 3
			case strings.Contains(path, "/signout") || strings.Contains(path, "/sign-out"):
				score += 2
			}
		}
		if score > best.score {
			best = candidate{id: ep.ID, score: score}
		}
	}
	return best.id
}

func (a *App) ScanActiveProject() ([]core.Endpoint, error) {
	id, err := a.settings.Get(a.ctx, domain.SettingActiveProjectID)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return []core.Endpoint{}, nil
	}
	return a.ScanWorkspace(id)
}

func (a *App) ListEndpoints(projectID string) ([]core.Endpoint, error) {
	if projectID == "" {
		return []core.Endpoint{}, nil
	}
	return a.endpoints.List(a.ctx, projectID)
}

func (a *App) GetProjectStats(projectID string) (domain.ProjectStats, error) {
	if projectID == "" {
		return domain.ProjectStats{}, nil
	}
	return a.endpoints.Stats(a.ctx, projectID)
}

func (a *App) GetProjectStatsReport(projectID string) (core.StatsReport, error) {
	if projectID == "" {
		return core.StatsReport{}, nil
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return core.StatsReport{}, err
	}
	driver, err := a.scanner.ResolveByName(project.Framework)
	if err != nil {
		return core.StatsReport{}, nil
	}
	cap, ok := driver.(core.StatsCapable)
	if !ok {
		return a.fallbackStatsReport(projectID)
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return core.StatsReport{}, err
	}
	return cap.Stats(a.ctx, project.Path, endpoints)
}

func (a *App) fallbackStatsReport(projectID string) (core.StatsReport, error) {
	stats, err := a.endpoints.Stats(a.ctx, projectID)
	if err != nil {
		return core.StatsReport{}, err
	}
	return core.StatsReport{
		Cards: []core.StatCard{
			{Key: "routes", Kind: core.StatRoutes, Label: "Routes", Value: stats.Routes},
			{Key: "controllers", Kind: core.StatControllers, Label: "Controllers", Value: stats.Controllers},
			{Key: "middleware", Kind: core.StatMiddleware, Label: "Middleware", Value: stats.Middleware},
		},
	}, nil
}

func (a *App) DetectProject(id string) (core.DetectionResult, error) {
	project, err := a.projects.GetByID(a.ctx, id)
	if err != nil {
		return core.DetectionResult{}, err
	}
	for _, d := range a.scanner.Drivers() {
		result := d.Detect(project.Path)
		if result.Detected {
			return result, nil
		}
	}
	return core.DetectionResult{}, nil
}
