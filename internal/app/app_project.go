package app

import (
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/workspace"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) OpenProject(path string) (*workspace.Workspace, error) {
	return a.workspace.Open(path)
}

func (a *App) DetectFramework(path string) (string, core.DetectionResult, error) {
	driver, result, err := a.scanner.Resolve(path)
	if err != nil {
		return "", core.DetectionResult{}, err
	}
	return driver.Name(), result, nil
}

func (a *App) SelectProjectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select Project Folder",
		CanCreateDirectories: false,
	})
}

type ProjectInfo struct {
	Path             string               `json:"path"`
	Name             string               `json:"name"`
	Framework        string               `json:"framework"`
	FrameworkVersion string               `json:"frameworkVersion"`
	Detection        core.DetectionResult `json:"detection"`
	APIDetection     APIDetection         `json:"apiDetection"`
	DefaultBaseURL   string               `json:"defaultBaseUrl"`
	DefaultPorts     []int                `json:"defaultPorts,omitempty"`
}

type APIDetection struct {
	Mode        string `json:"mode"`
	Value       string `json:"value"`
	Count       int    `json:"count"`
	TotalCount  int    `json:"totalCount"`
	ScanError   string `json:"scanError,omitempty"`
	LoginRoute  string `json:"loginRoute,omitempty"`
	LogoutRoute string `json:"logoutRoute,omitempty"`
}

func summarizeAuth(endpoints []core.Endpoint) (string, string) {
	type cand struct {
		ep    core.Endpoint
		score int
	}
	var bestLogin, bestLogout cand
	for _, ep := range endpoints {
		if ep.AuthRole == core.AuthRoleLogin {
			score := scoreAuthPath(string(ep.Path), core.AuthRoleLogin) + 1
			if score > bestLogin.score {
				bestLogin = cand{ep: ep, score: score}
			}
		} else if ep.AuthRole == core.AuthRoleLogout {
			score := scoreAuthPath(string(ep.Path), core.AuthRoleLogout) + 1
			if score > bestLogout.score {
				bestLogout = cand{ep: ep, score: score}
			}
		}
	}
	loginRoute := ""
	if bestLogin.score > 0 {
		loginRoute = string(bestLogin.ep.Method) + " " + bestLogin.ep.Path
	}
	logoutRoute := ""
	if bestLogout.score > 0 {
		logoutRoute = string(bestLogout.ep.Method) + " " + bestLogout.ep.Path
	}
	return loginRoute, logoutRoute
}

func scoreAuthPath(path string, role core.AuthRole) int {
	p := strings.ToLower(path)
	switch role {
	case core.AuthRoleLogin:
		switch {
		case strings.HasSuffix(p, "/login"):
			return 5
		case strings.Contains(p, "/login"):
			return 3
		case strings.Contains(p, "/signin") || strings.Contains(p, "/sign-in"):
			return 2
		case strings.Contains(p, "/authenticate") || strings.Contains(p, "/auth/token"):
			return 2
		}
	case core.AuthRoleLogout:
		switch {
		case strings.HasSuffix(p, "/logout"):
			return 5
		case strings.Contains(p, "/logout"):
			return 3
		case strings.Contains(p, "/signout") || strings.Contains(p, "/sign-out"):
			return 2
		}
	}
	return 0
}

func (a *App) InspectProject(path string) (ProjectInfo, error) {
	ws, err := a.workspace.Open(path)
	if err != nil {
		return ProjectInfo{}, err
	}
	info := ProjectInfo{
		Path:      ws.Path,
		Name:      ws.Name,
		Framework: "other",
	}
	driver, det, err := a.scanner.Resolve(ws.Path)
	if err != nil {
		return info, nil
	}
	info.Framework = driver.Name()
	info.Detection = det
	defaults := driver.Defaults()
	info.DefaultBaseURL = defaults.BaseURL
	if envURL := readDotenvAppURL(ws.Path); envURL != "" {
		info.DefaultBaseURL = envURL
	}
	info.DefaultPorts = defaults.Ports

	endpoints, scanErr := driver.Scan(a.ctx, ws.Path)
	if scanErr != nil {
		info.APIDetection = APIDetection{
			Mode:      core.FilterModeAuto,
			ScanError: scanErr.Error(),
		}
		return info, nil
	}
	result := core.ApplyFilter(endpoints, core.FilterModeAuto, "")
	loginRoute, logoutRoute := summarizeAuth(result.Endpoints)
	info.APIDetection = APIDetection{
		Mode:        result.Mode,
		Value:       result.Value,
		Count:       len(result.Endpoints),
		TotalCount:  len(endpoints),
		LoginRoute:  loginRoute,
		LogoutRoute: logoutRoute,
	}
	return info, nil
}

func (a *App) PreviewAPIRoutes(path, mode, value string) (APIDetection, error) {
	driver, _, err := a.scanner.Resolve(path)
	if err != nil {
		return APIDetection{}, err
	}
	endpoints, err := driver.Scan(a.ctx, path)
	if err != nil {
		return APIDetection{Mode: mode, Value: value, ScanError: err.Error()}, nil
	}
	result := core.ApplyFilter(endpoints, mode, value)
	loginRoute, logoutRoute := summarizeAuth(result.Endpoints)
	return APIDetection{
		Mode:        result.Mode,
		Value:       result.Value,
		Count:       len(result.Endpoints),
		TotalCount:  len(endpoints),
		LoginRoute:  loginRoute,
		LogoutRoute: logoutRoute,
	}, nil
}

func (a *App) Drivers() []string {
	names := make([]string, 0)
	for _, d := range a.scanner.Drivers() {
		names = append(names, d.Name())
	}
	return names
}

func (a *App) ListProjects() ([]domain.Project, error) {
	return a.projects.List(a.ctx)
}

func (a *App) SaveProject(input domain.ProjectInput) (*domain.Project, error) {
	if strings.TrimSpace(input.BaseURL) == "" {
		if driver, _, err := a.scanner.Resolve(input.Path); err == nil {
			input.BaseURL = driver.Defaults().BaseURL
		}
	}
	return a.projects.Save(a.ctx, input)
}

func (a *App) UpdateProjectBaseURL(projectID, baseURL string) error {
	return a.projects.UpdateBaseURL(a.ctx, projectID, baseURL)
}
