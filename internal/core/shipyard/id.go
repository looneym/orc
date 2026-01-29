// Package shipyard contains domain logic for shipyard entities.
package shipyard

import "fmt"

// GenerateShipyardID creates a new shipyard ID from the current maximum.
func GenerateShipyardID(currentMax int) string {
	return fmt.Sprintf("YARD-%03d", currentMax+1)
}

// ParseShipyardNumber extracts the numeric part from a shipyard ID.
func ParseShipyardNumber(id string) int {
	var num int
	_, err := fmt.Sscanf(id, "YARD-%d", &num)
	if err != nil {
		return -1
	}
	return num
}
