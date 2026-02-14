package sqlite_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/example/orc/internal/adapters/sqlite"
	"github.com/example/orc/internal/db"
	"github.com/example/orc/internal/ports/secondary"
)

// setupFileDB creates a file-backed SQLite database with WAL mode and busy_timeout
// configured, matching production init from db.go. Uses t.TempDir() for cleanup.
func setupFileDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	testDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open file db: %v", err)
	}

	// Apply pragmas matching production (db.go)
	if _, err := testDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	if _, err := testDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		t.Fatalf("failed to enable WAL: %v", err)
	}
	if _, err := testDB.Exec("PRAGMA busy_timeout=5000"); err != nil {
		t.Fatalf("failed to set busy_timeout: %v", err)
	}

	// Load authoritative schema
	if _, err := testDB.Exec(db.GetSchemaSQL()); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	t.Cleanup(func() { testDB.Close() })
	return testDB
}

// TestWALModeActive verifies that WAL mode is active after database init.
func TestWALModeActive(t *testing.T) {
	testDB := setupFileDB(t)

	var journalMode string
	err := testDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode=wal, got %q", journalMode)
	}
}

// TestBusyTimeoutSet verifies that busy_timeout is set.
func TestBusyTimeoutSet(t *testing.T) {
	testDB := setupFileDB(t)

	var busyTimeout int
	err := testDB.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout)
	if err != nil {
		t.Fatalf("failed to query busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Errorf("expected busy_timeout=5000, got %d", busyTimeout)
	}
}

// TestAtomicIDGeneration_Commission verifies that concurrent commission creates
// via goroutines never produce duplicate IDs. This validates that BEGIN IMMEDIATE
// transactions serialize the GetNextID + Create pair.
func TestAtomicIDGeneration_Commission(t *testing.T) {
	testDB := setupFileDB(t)
	repo := sqlite.NewCommissionRepository(testDB, nil)
	transactor := sqlite.NewTransactor(testDB)

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	ids := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx := context.Background()

			err := transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
				nextID, err := repo.GetNextID(txCtx)
				if err != nil {
					return fmt.Errorf("goroutine %d: GetNextID failed: %w", n, err)
				}

				record := &secondary.CommissionRecord{
					ID:     nextID,
					Title:  fmt.Sprintf("Commission %d", n),
					Status: "active",
				}
				if err := repo.Create(txCtx, record); err != nil {
					return fmt.Errorf("goroutine %d: Create failed: %w", n, err)
				}

				ids <- nextID
				return nil
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	close(ids)

	// Check for errors
	for err := range errs {
		t.Error(err)
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != goroutines {
		t.Errorf("expected %d unique IDs, got %d", goroutines, len(seen))
	}
}

// TestAtomicIDGeneration_Task verifies that concurrent task creates never
// produce duplicate IDs when using the transactor.
func TestAtomicIDGeneration_Task(t *testing.T) {
	testDB := setupFileDB(t)
	repo := sqlite.NewTaskRepository(testDB, nil)
	transactor := sqlite.NewTransactor(testDB)

	// Seed a commission for the FK constraint
	_, err := testDB.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-001', 'Test', 'active')")
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	ids := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx := context.Background()

			err := transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
				nextID, err := repo.GetNextID(txCtx)
				if err != nil {
					return fmt.Errorf("goroutine %d: GetNextID failed: %w", n, err)
				}

				record := &secondary.TaskRecord{
					ID:           nextID,
					CommissionID: "COMM-001",
					Title:        fmt.Sprintf("Task %d", n),
					Status:       "open",
				}
				if err := repo.Create(txCtx, record); err != nil {
					return fmt.Errorf("goroutine %d: Create failed: %w", n, err)
				}

				ids <- nextID
				return nil
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	close(ids)

	// Check for errors
	for err := range errs {
		t.Error(err)
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != goroutines {
		t.Errorf("expected %d unique IDs, got %d", goroutines, len(seen))
	}
}

// TestAtomicIDGeneration_Note verifies that concurrent note creates never
// produce duplicate IDs. Notes are high-frequency entities in ORC so this
// exercises a realistic contention scenario.
func TestAtomicIDGeneration_Note(t *testing.T) {
	testDB := setupFileDB(t)
	repo := sqlite.NewNoteRepository(testDB, nil)
	transactor := sqlite.NewTransactor(testDB)

	// Seed a commission for the FK constraint
	_, err := testDB.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-001', 'Test', 'active')")
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)
	ids := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx := context.Background()

			err := transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
				nextID, err := repo.GetNextID(txCtx)
				if err != nil {
					return fmt.Errorf("goroutine %d: GetNextID failed: %w", n, err)
				}

				record := &secondary.NoteRecord{
					ID:           nextID,
					CommissionID: "COMM-001",
					Title:        fmt.Sprintf("Note %d", n),
					Type:         "exploration",
				}
				if err := repo.Create(txCtx, record); err != nil {
					return fmt.Errorf("goroutine %d: Create failed: %w", n, err)
				}

				ids <- nextID
				return nil
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	close(ids)

	// Check for errors
	for err := range errs {
		t.Error(err)
	}

	// Verify all IDs are unique
	seen := make(map[string]bool)
	for id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}

	if len(seen) != goroutines {
		t.Errorf("expected %d unique IDs, got %d", goroutines, len(seen))
	}
}

// TestEventWriteWithinEntityTransaction verifies that event writes work
// correctly when an entity transaction is active. The EventWriterAdapter
// detects the existing transaction and piggybacks on it rather than
// starting its own BEGIN IMMEDIATE.
func TestEventWriteWithinEntityTransaction(t *testing.T) {
	testDB := setupFileDB(t)

	// Create repos and event writer infrastructure
	workshopEventRepo := sqlite.NewWorkshopEventRepository(testDB)
	operationalEventRepo := sqlite.NewOperationalEventRepository(testDB)
	workbenchRepo := sqlite.NewWorkbenchRepository(testDB, nil)
	transactor := sqlite.NewTransactor(testDB)

	eventWriter := sqlite.NewEventWriterAdapter(
		workshopEventRepo,
		operationalEventRepo,
		workbenchRepo,
		transactor,
		"test-version",
	)

	// Seed factory, workshop, and workbench for actor resolution
	_, err := testDB.Exec("INSERT INTO factories (id, name, status) VALUES ('FACT-001', 'test', 'active')")
	if err != nil {
		t.Fatalf("failed to seed factory: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO workshops (id, factory_id, name, status) VALUES ('SHOP-001', 'FACT-001', 'test', 'active')")
	if err != nil {
		t.Fatalf("failed to seed workshop: %v", err)
	}

	// Create a commission repo that emits events on Create
	commissionRepo := sqlite.NewCommissionRepository(testDB, eventWriter)

	ctx := context.Background()

	// Create a commission within a transaction — the event writer should
	// detect the active tx and not start a nested one
	err = transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
		nextID, err := commissionRepo.GetNextID(txCtx)
		if err != nil {
			return fmt.Errorf("GetNextID failed: %w", err)
		}

		record := &secondary.CommissionRecord{
			ID:     nextID,
			Title:  "Test Commission",
			Status: "active",
		}
		return commissionRepo.Create(txCtx, record)
	})
	if err != nil {
		t.Fatalf("transactional create failed: %v", err)
	}

	// Verify commission was created
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM commissions").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count commissions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 commission, got %d", count)
	}

	// Verify operational events can also be emitted independently
	err = eventWriter.EmitOperational(ctx, "test", "info", "standalone event", nil)
	if err != nil {
		t.Fatalf("standalone EmitOperational failed: %v", err)
	}

	var opsCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM operational_events").Scan(&opsCount)
	if err != nil {
		t.Fatalf("failed to count operational events: %v", err)
	}
	if opsCount != 1 {
		t.Errorf("expected 1 operational event, got %d", opsCount)
	}
}

// TestConcurrentWritesSucceedWithBusyTimeout verifies that concurrent writers
// on the same file-backed DB succeed rather than getting SQLITE_BUSY errors,
// thanks to WAL mode and busy_timeout working together.
func TestConcurrentWritesSucceedWithBusyTimeout(t *testing.T) {
	testDB := setupFileDB(t)

	// Seed a commission for FK constraints
	_, err := testDB.Exec("INSERT INTO commissions (id, title, status) VALUES ('COMM-001', 'Test', 'active')")
	if err != nil {
		t.Fatalf("failed to seed commission: %v", err)
	}

	// Create multiple repos that write to different tables concurrently
	commissionRepo := sqlite.NewCommissionRepository(testDB, nil)
	noteRepo := sqlite.NewNoteRepository(testDB, nil)
	transactor := sqlite.NewTransactor(testDB)

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, goroutines*2)

	// Goroutines creating commissions
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx := context.Background()
			err := transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
				id, err := commissionRepo.GetNextID(txCtx)
				if err != nil {
					return err
				}
				return commissionRepo.Create(txCtx, &secondary.CommissionRecord{
					ID:     id,
					Title:  fmt.Sprintf("Concurrent Commission %d", n),
					Status: "active",
				})
			})
			if err != nil {
				errs <- fmt.Errorf("commission goroutine %d: %w", n, err)
			}
		}(i)
	}

	// Goroutines creating notes concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ctx := context.Background()
			err := transactor.WithImmediateTx(ctx, func(txCtx context.Context) error {
				id, err := noteRepo.GetNextID(txCtx)
				if err != nil {
					return err
				}
				return noteRepo.Create(txCtx, &secondary.NoteRecord{
					ID:           id,
					CommissionID: "COMM-001",
					Title:        fmt.Sprintf("Concurrent Note %d", n),
					Type:         "exploration",
				})
			})
			if err != nil {
				errs <- fmt.Errorf("note goroutine %d: %w", n, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}

	// Verify all writes succeeded
	var commCount, noteCount int
	testDB.QueryRow("SELECT COUNT(*) FROM commissions").Scan(&commCount)
	testDB.QueryRow("SELECT COUNT(*) FROM notes").Scan(&noteCount)

	// commCount includes the seeded commission
	if commCount != goroutines+1 {
		t.Errorf("expected %d commissions, got %d", goroutines+1, commCount)
	}
	if noteCount != goroutines {
		t.Errorf("expected %d notes, got %d", goroutines, noteCount)
	}
}
