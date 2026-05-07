package laravel

import (
	"context"

	"spectra-desktop/internal/core"
)

func scanRoutes(_ context.Context, _ string) ([]core.Endpoint, error) {
	return []core.Endpoint{}, nil
}
