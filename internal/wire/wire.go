// Package wire provides dependency injection for the ORC application.
// It creates singleton services with lazy initialization.
package wire

import (
	"io"
	"os"
	"sync"

	cliadapter "github.com/example/orc/internal/adapters/cli"
	"github.com/example/orc/internal/adapters/persistence"
	"github.com/example/orc/internal/app"
	"github.com/example/orc/internal/ports/primary"
)

var (
	missionService primary.MissionService
	groveService   primary.GroveService
	once           sync.Once
)

// MissionService returns the singleton MissionService instance.
func MissionService() primary.MissionService {
	once.Do(initServices)
	return missionService
}

// GroveService returns the singleton GroveService instance.
func GroveService() primary.GroveService {
	once.Do(initServices)
	return groveService
}

// initServices initializes all services and their dependencies.
// This is called once via sync.Once.
func initServices() {
	// Create repository adapters (secondary ports)
	missionRepo := persistence.NewMissionRepository()
	groveRepo := persistence.NewGroveRepository()
	agentProvider := persistence.NewAgentIdentityProvider()

	// Create effect executor
	executor := app.NewEffectExecutor()

	// Create services (primary ports implementation)
	missionService = app.NewMissionService(missionRepo, groveRepo, agentProvider, executor)
	groveService = app.NewGroveService(groveRepo, missionRepo, agentProvider, executor)
}

// MissionAdapter returns a new MissionAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func MissionAdapter() *cliadapter.MissionAdapter {
	return MissionAdapterWithOutput(os.Stdout)
}

// MissionAdapterWithOutput returns a new MissionAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func MissionAdapterWithOutput(out io.Writer) *cliadapter.MissionAdapter {
	once.Do(initServices)
	return cliadapter.NewMissionAdapter(missionService, out)
}

// GroveAdapter returns a new GroveAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func GroveAdapter() *cliadapter.GroveAdapter {
	return GroveAdapterWithOutput(os.Stdout)
}

// GroveAdapterWithOutput returns a new GroveAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func GroveAdapterWithOutput(out io.Writer) *cliadapter.GroveAdapter {
	once.Do(initServices)
	return cliadapter.NewGroveAdapter(groveService, out)
}
