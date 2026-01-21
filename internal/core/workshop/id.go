// Package workshop contains the pure business logic for workshop operations.
// This is part of the Functional Core - no I/O, only pure functions.
package workshop

import "fmt"

// GenerateWorkshopID generates a workshop ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is WORK-XXX where XXX is a zero-padded 3-digit number.
func GenerateWorkshopID(currentMax int) string {
	return fmt.Sprintf("WORK-%03d", currentMax+1)
}

// ParseWorkshopNumber extracts the numeric portion from a workshop ID.
// Returns -1 if the ID format is invalid.
func ParseWorkshopNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "WORK-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
