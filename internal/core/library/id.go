// Package library contains domain logic for library entities.
package library

import "fmt"

// GenerateLibraryID creates a new library ID from the current maximum.
func GenerateLibraryID(currentMax int) string {
	return fmt.Sprintf("LIB-%03d", currentMax+1)
}

// ParseLibraryNumber extracts the numeric part from a library ID.
func ParseLibraryNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "LIB-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
