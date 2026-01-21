// Package factory contains the pure business logic for factory operations.
// This is part of the Functional Core - no I/O, only pure functions.
package factory

import "fmt"

// GenerateFactoryID generates a factory ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is FACT-XXX where XXX is a zero-padded 3-digit number.
func GenerateFactoryID(currentMax int) string {
	return fmt.Sprintf("FACT-%03d", currentMax+1)
}

// ParseFactoryNumber extracts the numeric portion from a factory ID.
// Returns -1 if the ID format is invalid.
func ParseFactoryNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "FACT-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
