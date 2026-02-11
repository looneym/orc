package cli

import (
	"context"

	"github.com/example/orc/internal/config"
)

// MigrateGoblinConfigIfNeeded is a legacy stub. Goblin configs are no longer supported.
// Returns the config as-is.
func MigrateGoblinConfigIfNeeded(_ context.Context, dir string) (*config.Config, error) {
	return config.LoadConfig(dir)
}
