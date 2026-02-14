// Package wire provides dependency injection for the ORC application.
// It creates singleton services with lazy initialization.
package wire

import (
	"io"
	"log"
	"os"
	"sync"

	cliadapter "github.com/example/orc/internal/adapters/cli"
	"github.com/example/orc/internal/adapters/filesystem"
	"github.com/example/orc/internal/adapters/persistence"
	"github.com/example/orc/internal/adapters/sqlite"
	tmuxadapter "github.com/example/orc/internal/adapters/tmux"
	"github.com/example/orc/internal/app"
	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/ports/secondary"
	"github.com/example/orc/internal/version"
)

var (
	commissionService              primary.CommissionService
	shipmentService                primary.ShipmentService
	taskService                    primary.TaskService
	noteService                    primary.NoteService
	tomeService                    primary.TomeService
	planService                    primary.PlanService
	tagService                     primary.TagService
	repoService                    primary.RepoService
	prService                      primary.PRService
	factoryService                 primary.FactoryService
	workshopService                primary.WorkshopService
	workbenchService               primary.WorkbenchService
	summaryService                 primary.SummaryService
	eventService                   primary.EventService
	hookEventService               primary.HookEventService
	commissionOrchestrationService *app.CommissionOrchestrationService
	tmuxService                    secondary.TMuxAdapter
	shipmentRepo                   secondary.ShipmentRepository
	eventWriterInstance            secondary.EventWriter
	once                           sync.Once
)

// CommissionService returns the singleton CommissionService instance.
func CommissionService() primary.CommissionService {
	once.Do(initServices)
	return commissionService
}

// ShipmentService returns the singleton ShipmentService instance.
func ShipmentService() primary.ShipmentService {
	once.Do(initServices)
	return shipmentService
}

// TaskService returns the singleton TaskService instance.
func TaskService() primary.TaskService {
	once.Do(initServices)
	return taskService
}

// NoteService returns the singleton NoteService instance.
func NoteService() primary.NoteService {
	once.Do(initServices)
	return noteService
}

// TomeService returns the singleton TomeService instance.
func TomeService() primary.TomeService {
	once.Do(initServices)
	return tomeService
}

// PlanService returns the singleton PlanService instance.
func PlanService() primary.PlanService {
	once.Do(initServices)
	return planService
}

// TagService returns the singleton TagService instance.
func TagService() primary.TagService {
	once.Do(initServices)
	return tagService
}

// RepoService returns the singleton RepoService instance.
func RepoService() primary.RepoService {
	once.Do(initServices)
	return repoService
}

// PRService returns the singleton PRService instance.
func PRService() primary.PRService {
	once.Do(initServices)
	return prService
}

// FactoryService returns the singleton FactoryService instance.
func FactoryService() primary.FactoryService {
	once.Do(initServices)
	return factoryService
}

// WorkshopService returns the singleton WorkshopService instance.
func WorkshopService() primary.WorkshopService {
	once.Do(initServices)
	return workshopService
}

// WorkbenchService returns the singleton WorkbenchService instance.
func WorkbenchService() primary.WorkbenchService {
	once.Do(initServices)
	return workbenchService
}

// SummaryService returns the singleton SummaryService instance.
func SummaryService() primary.SummaryService {
	once.Do(initServices)
	return summaryService
}

// EventService returns the singleton EventService instance.
func EventService() primary.EventService {
	once.Do(initServices)
	return eventService
}

// HookEventService returns the singleton HookEventService instance.
func HookEventService() primary.HookEventService {
	once.Do(initServices)
	return hookEventService
}

// CommissionOrchestrationService returns the singleton CommissionOrchestrationService instance.
func CommissionOrchestrationService() *app.CommissionOrchestrationService {
	once.Do(initServices)
	return commissionOrchestrationService
}

// TMuxAdapter returns the singleton TMuxAdapter instance.
func TMuxAdapter() secondary.TMuxAdapter {
	once.Do(initServices)
	return tmuxService
}

// ShipmentRepository returns the singleton ShipmentRepository instance.
func ShipmentRepository() secondary.ShipmentRepository {
	once.Do(initServices)
	return shipmentRepo
}

// EventWriter returns the singleton EventWriter instance.
func EventWriter() secondary.EventWriter {
	once.Do(initServices)
	return eventWriterInstance
}

// initServices initializes all services and their dependencies.
// This is called once via sync.Once.
func initServices() {
	// Get database connection
	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Create EventWriter infrastructure early (needed by most repositories)
	// Order matters: workshopEventRepo needs DB, workbenchRepo needs DB, eventWriter needs both
	workshopEventRepo := sqlite.NewWorkshopEventRepository(database)
	operationalEventRepo := sqlite.NewOperationalEventRepository(database)
	workbenchRepo := sqlite.NewWorkbenchRepository(database, nil) // nil EventWriter: circular dependency (EventWriter needs workbenchRepo)
	transactor := sqlite.NewTransactor(database)
	eventWriter := sqlite.NewEventWriterAdapter(workshopEventRepo, operationalEventRepo, workbenchRepo, transactor, version.Commit)
	eventWriterInstance = eventWriter

	// Create repository adapters (secondary ports) - sqlite adapters with injected DB
	commissionRepo := sqlite.NewCommissionRepository(database, eventWriter)
	agentProvider := persistence.NewAgentIdentityProvider()
	tmuxAdapter := tmuxadapter.NewAdapter("") // default factory → default tmux server
	tmuxService = tmuxAdapter                 // Store for getter

	// Create workspace adapter (needed by effect executor and workshop service)
	home, _ := os.UserHomeDir()
	workspaceAdapter, err := filesystem.NewWorkspaceAdapter(home+"/wb", home+"/src") // ~/wb for worktrees, ~/src for repos
	if err != nil {
		log.Fatalf("failed to create workspace adapter: %v", err)
	}

	// Create effect executor with injected repositories and adapters
	executor := app.NewEffectExecutor(commissionRepo, tmuxAdapter, workspaceAdapter)

	// Create services (primary ports implementation)
	commissionService = app.NewCommissionService(commissionRepo, agentProvider, executor, transactor)

	// Create shipment and task services
	shipmentRepo = sqlite.NewShipmentRepository(database, eventWriter)
	taskRepo := sqlite.NewTaskRepository(database, eventWriter)
	tagRepo := sqlite.NewTagRepository(database)
	taskService = app.NewTaskService(taskRepo, tagRepo, shipmentRepo, transactor)

	// Create note and tome services
	noteRepo := sqlite.NewNoteRepository(database, eventWriter)
	tomeRepo := sqlite.NewTomeRepository(database, eventWriter)
	noteService = app.NewNoteService(noteRepo, transactor)

	// Create tome and shipment services
	tomeService = app.NewTomeService(tomeRepo, noteService, transactor)
	shipmentService = app.NewShipmentService(shipmentRepo, taskRepo, noteService, transactor)

	// Create plan repository
	planRepo := sqlite.NewPlanRepository(database, eventWriter)

	// Create tag service
	tagService = app.NewTagService(tagRepo, transactor)

	// Create repo and PR services
	repoRepo := sqlite.NewRepoRepository(database)
	prRepo := sqlite.NewPRRepository(database)
	repoService = app.NewRepoService(repoRepo, transactor)
	prService = app.NewPRService(prRepo, shipmentService, transactor)

	// Create factory, workshop, and workbench services
	factoryRepo := sqlite.NewFactoryRepository(database)
	workshopRepo := sqlite.NewWorkshopRepository(database)
	// workbenchRepo already created early for EventWriter (with nil EventWriter due to circular dependency)
	factoryService = app.NewFactoryService(factoryRepo, transactor)
	workshopService = app.NewWorkshopService(factoryRepo, workshopRepo, workbenchRepo, repoRepo, tmuxService, workspaceAdapter, executor, transactor)
	workbenchService = app.NewWorkbenchService(workbenchRepo, workshopRepo, repoRepo, agentProvider, executor, workspaceAdapter, transactor)

	// Create plan service
	planService = app.NewPlanService(planRepo, transactor)

	// Create event service (unified audit + operational events)
	eventService = app.NewEventService(workshopEventRepo, operationalEventRepo)

	// Create hook event service for hook invocation tracking
	hookEventRepo := sqlite.NewHookEventRepository(database)
	hookEventService = app.NewHookEventService(hookEventRepo, transactor)

	// Create orchestration services
	commissionOrchestrationService = app.NewCommissionOrchestrationService(commissionService, agentProvider)

	// Create summary service (depends on most other services)
	summaryService = app.NewSummaryService(
		commissionService,
		tomeService,
		shipmentService,
		taskService,
		noteService,
		workbenchService,
		planService,
	)
}

// IsOrcSession returns true if the current tmux session has ORC_WORKSHOP_ID set,
// indicating this is an ORC-managed workshop session.
func IsOrcSession() bool {
	return tmuxadapter.IsOrcSession()
}

// ApplyGlobalTMuxBindings sets up ORC's global tmux key bindings.
// Safe to call repeatedly (idempotent). Silently ignores errors (tmux may not be running).
// This is called on every orc command invocation to ensure bindings are always current.
func ApplyGlobalTMuxBindings() {
	tmuxadapter.ApplyGlobalBindings()
}

// CommissionAdapter returns a new CommissionAdapter writing to stdout.
// Each call creates a new adapter (adapters are stateless translators).
func CommissionAdapter() *cliadapter.CommissionAdapter {
	return CommissionAdapterWithOutput(os.Stdout)
}

// CommissionAdapterWithOutput returns a new CommissionAdapter writing to the given output.
// This variant allows testing or alternate output destinations.
func CommissionAdapterWithOutput(out io.Writer) *cliadapter.CommissionAdapter {
	once.Do(initServices)
	return cliadapter.NewCommissionAdapter(commissionService, out)
}

// EnrichSession applies ORC enrichment to all windows in a session.
func EnrichSession(sessionName string) error {
	once.Do(initServices)
	return tmuxadapter.EnrichSession(sessionName)
}

// GotmuxAdapter re-exports the gotmux adapter type for CLI use.
type GotmuxAdapter = tmuxadapter.GotmuxAdapter

// DesiredWorkbench re-exports the desired workbench type for plan building.
type DesiredWorkbench = tmuxadapter.DesiredWorkbench

// ApplyPlan re-exports the reconciliation plan type.
type ApplyPlan = tmuxadapter.ApplyPlan

// NewGotmuxAdapter creates a new gotmux adapter for the default tmux server.
func NewGotmuxAdapter() (*GotmuxAdapter, error) {
	return tmuxadapter.NewGotmuxAdapter("")
}

// NewGotmuxAdapterWithSocket creates a gotmux adapter targeting a specific tmux socket.
// An empty socket means the default server; a non-empty socket targets that server.
func NewGotmuxAdapterWithSocket(socket string) (*GotmuxAdapter, error) {
	return tmuxadapter.NewGotmuxAdapter(socket)
}

// FactorySocket derives a tmux socket name from a factory name.
func FactorySocket(factoryName string) string {
	return tmuxadapter.FactorySocket(factoryName)
}

// DeskServerInfo re-exports the desk server info type.
type DeskServerInfo = tmuxadapter.DeskServerInfo

// ListDeskServers scans for *-desk tmux server sockets.
func ListDeskServers() ([]DeskServerInfo, error) {
	return tmuxadapter.ListDeskServers()
}

// KillDeskServer kills a specific desk server by workbench name.
func KillDeskServer(benchName string) error {
	return tmuxadapter.KillDeskServer(benchName)
}

// KillAllDeskServers kills all discoverable desk servers.
func KillAllDeskServers() (int, error) {
	return tmuxadapter.KillAllDeskServers()
}
