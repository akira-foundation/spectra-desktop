package app

import "spectra-desktop/internal/core"

func (a *App) FormatException(projectID string, body string, status int) (*core.FormattedException, bool) {
	driver := a.driverForProject(projectID)
	if driver == nil {
		return nil, false
	}
	formatter, ok := driver.(core.ExceptionFormatter)
	if !ok {
		return nil, false
	}
	parsed, ok := formatter.FormatException(body, status)
	if !ok {
		return nil, false
	}
	return &parsed, true
}
