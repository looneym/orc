// Package commission contains the pure business logic for commission operations.
// This is part of the Functional Core - no I/O, only pure functions.
package commission

import "fmt"

// GenerateCommissionID generates a commission ID from the current max number.
// This is a pure function that defines the ID format as a business rule.
// The format is COMM-XXX where XXX is a zero-padded 3-digit number.
func GenerateCommissionID(currentMax int) string {
	return fmt.Sprintf("COMM-%03d", currentMax+1)
}

// ParseCommissionNumber extracts the numeric portion from a commission ID.
// Returns -1 if the ID format is invalid.
func ParseCommissionNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "COMM-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
