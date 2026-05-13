package app

import (
	"strings"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/drivers/laravel"
)

func (a *App) GetPHPBinaryPath() (string, error) {
	if a.settings == nil {
		return "", nil
	}
	value, err := a.settings.Get(a.ctx, domain.SettingPHPBinaryPath)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (a *App) SetPHPBinaryPath(path string) error {
	clean := strings.TrimSpace(path)
	if a.settings != nil {
		if clean == "" {
			_ = a.settings.Delete(a.ctx, domain.SettingPHPBinaryPath)
		} else if err := a.settings.Set(a.ctx, domain.SettingPHPBinaryPath, clean); err != nil {
			return err
		}
	}
	laravel.SetPHPBinaryOverride(clean)
	return nil
}

func (a *App) DetectPHPBinary() string {
	path, err := laravel.ResolvePHPBinaryPath()
	if err != nil {
		return ""
	}
	return path
}
