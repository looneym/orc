package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/orc/internal/ports/primary"
	"github.com/example/orc/internal/wire"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View unified event stream",
	Long:  "View, search, and manage audit and operational events",
}

var eventsTailCmd = &cobra.Command{
	Use:   "tail",
	Short: "Show recent events",
	Long:  "Show recent events (default 50). Interleaves audit and operational events by timestamp.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		limit, _ := cmd.Flags().GetInt("limit")
		workshopID, _ := cmd.Flags().GetString("workshop")
		actorID, _ := cmd.Flags().GetString("actor")
		source, _ := cmd.Flags().GetString("source")
		level, _ := cmd.Flags().GetString("level")
		eventType, _ := cmd.Flags().GetString("type")
		follow, _ := cmd.Flags().GetBool("follow")

		if limit <= 0 {
			limit = 50
		}

		filters := primary.EventFilters{
			WorkshopID: workshopID,
			ActorID:    actorID,
			Source:     source,
			Level:      level,
			EventType:  eventType,
			Limit:      limit,
		}

		// Initial fetch
		events, err := wire.EventService().ListEvents(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to fetch events: %w", err)
		}

		// Apply default level filter (info+): hide debug unless explicitly requested
		events = filterByLevel(events, level)

		printEventEntries(events)

		// If --follow, poll for new entries
		if follow {
			var lastTimestamp string
			if len(events) > 0 {
				lastTimestamp = events[0].Timestamp
			}

			for {
				time.Sleep(1 * time.Second)

				newEvents, err := wire.EventService().ListEvents(ctx, filters)
				if err != nil {
					fmt.Printf("Error fetching events: %v\n", err)
					continue
				}

				newEvents = filterByLevel(newEvents, level)

				// Print only entries newer than lastTimestamp
				for i := len(newEvents) - 1; i >= 0; i-- {
					event := newEvents[i]
					if lastTimestamp == "" || event.Timestamp > lastTimestamp {
						printEventEntry(event)
						if event.Timestamp > lastTimestamp {
							lastTimestamp = event.Timestamp
						}
					}
				}
			}
		}

		return nil
	},
}

var eventsShowCmd = &cobra.Command{
	Use:   "show [id-or-entity]",
	Short: "Show events for a specific entity or by ID",
	Long:  "Show event history for a specific entity (e.g., SHIP-243, TASK-001) or a single event by ID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		actorID, _ := cmd.Flags().GetString("actor")
		limit, _ := cmd.Flags().GetInt("limit")

		filters := primary.EventFilters{
			ActorID: actorID,
			Limit:   limit,
		}

		// If entity ID provided, filter by it
		if len(args) > 0 {
			filters.EntityID = args[0]
		}

		events, err := wire.EventService().ListEvents(ctx, filters)
		if err != nil {
			return fmt.Errorf("failed to fetch events: %w", err)
		}

		if len(events) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		printEventEntries(events)
		return nil
	},
}

var eventsPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old events",
	Long:  "Delete events older than the specified number of days (default 30)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := NewContext()
		days, _ := cmd.Flags().GetInt("days")

		if days <= 0 {
			days = 30
		}

		count, err := wire.EventService().PruneEvents(ctx, days)
		if err != nil {
			return fmt.Errorf("failed to prune events: %w", err)
		}

		if count == 0 {
			fmt.Printf("No events older than %d days found.\n", days)
		} else {
			fmt.Printf("Pruned %d events older than %d days.\n", count, days)
		}
		return nil
	},
}

// filterByLevel applies the default level filter.
// If no explicit level was passed, hide debug events.
func filterByLevel(events []*primary.Event, explicitLevel string) []*primary.Event {
	if explicitLevel != "" {
		// Explicit level filter — already handled by repo query
		return events
	}

	// Default: info+ (hide debug)
	var filtered []*primary.Event
	for _, e := range events {
		if e.Level == "debug" {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

func printEventEntries(events []*primary.Event) {
	if len(events) == 0 {
		fmt.Println("No events found.")
		return
	}

	fmt.Printf("Found %d events:\n\n", len(events))

	// Print in reverse order (oldest first) for tail view
	for i := len(events) - 1; i >= 0; i-- {
		printEventEntry(events[i])
	}
}

func printEventEntry(event *primary.Event) {
	actorStr := event.ActorID
	if actorStr == "" {
		actorStr = "-"
	}

	ts := formatEventTimestamp(event.Timestamp)

	if event.EntityType != "" {
		// Audit event: entity CRUD format
		actionIcon := getEventActionIcon(event.Action)
		fmt.Printf("%s | %-12s | %s %s | %s/%s",
			ts,
			actorStr,
			actionIcon,
			event.Action,
			event.EntityType,
			event.EntityID,
		)
		if event.Action == "update" && event.FieldName != "" {
			fmt.Printf(" | %s: %s -> %s", event.FieldName, event.OldValue, event.NewValue)
		}
		fmt.Println()
	} else {
		// Operational event: [source] level: message format
		fmt.Printf("%s | %-12s | [%s] %s: %s\n",
			ts,
			actorStr,
			event.Source,
			event.Level,
			event.Message,
		)
	}
}

func getEventActionIcon(action string) string {
	switch action {
	case "create":
		return "+"
	case "update":
		return "~"
	case "delete":
		return "-"
	default:
		return "?"
	}
}

func formatEventTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("2006-01-02 15:04:05")
}

// EventsCmd returns the events command with all subcommands attached.
func EventsCmd() *cobra.Command {
	// events tail
	eventsTailCmd.Flags().IntP("limit", "n", 50, "Number of entries to show")
	eventsTailCmd.Flags().String("workshop", "", "Filter by workshop ID")
	eventsTailCmd.Flags().String("actor", "", "Filter by actor ID")
	eventsTailCmd.Flags().String("source", "", "Filter by event source")
	eventsTailCmd.Flags().String("level", "", "Filter by level (debug, info, warn, error)")
	eventsTailCmd.Flags().String("type", "", "Filter by event type (audit, ops)")
	eventsTailCmd.Flags().BoolP("follow", "f", false, "Follow mode: poll for new entries")

	// events show
	eventsShowCmd.Flags().String("actor", "", "Filter by actor ID")
	eventsShowCmd.Flags().IntP("limit", "n", 100, "Maximum entries to show")

	// events prune
	eventsPruneCmd.Flags().Int("days", 30, "Delete entries older than N days")

	eventsCmd.AddCommand(eventsTailCmd)
	eventsCmd.AddCommand(eventsShowCmd)
	eventsCmd.AddCommand(eventsPruneCmd)

	return eventsCmd
}
