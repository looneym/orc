// Package workbench contains the pure business logic for workbench operations.
// This is part of the Functional Core - no I/O, only pure functions.
package workbench

import "fmt"

// GenerateWorkbenchID generates a workbench ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is BENCH-XXX where XXX is a zero-padded 3-digit number.
func GenerateWorkbenchID(currentMax int) string {
	return fmt.Sprintf("BENCH-%03d", currentMax+1)
}

// ParseWorkbenchNumber extracts the numeric portion from a workbench ID.
// Returns -1 if the ID format is invalid.
func ParseWorkbenchNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "BENCH-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
