// Package workshop contains the pure business logic for workshop operations.
package workshop

// WorkshopNames is the pool of atmospheric names for workshops.
// These names evoke a sense of craftsmanship and industry.
var WorkshopNames = []string{
	"Ironmoss Forge",
	"Blackpine Foundry",
	"Mosslight Mill",
	"Copperwood Works",
	"Ashvale Smithy",
	"Stonefern Studio",
	"Bramblegate Bench",
	"Cinderhollow Shop",
	"Driftwood Den",
	"Embervale Workshop",
	"Frostpeak Forge",
	"Glimmershade Guild",
	"Hearthstone Hall",
	"Ivywood Instruments",
	"Jadecrest Junction",
	"Kindlewood Keep",
	"Lanternwick Loft",
	"Moonvale Manufactory",
	"Nightshade Nook",
	"Oakenshield Outpost",
	"Pinecrest Parlor",
	"Quartzdale Quarters",
	"Ravenshadow Refuge",
	"Silverbark Station",
	"Thornwood Terrace",
	"Undermoss Utility",
	"Verdanthollow Vault",
	"Willowmere Works",
	"Xylem Exchange",
	"Yarrowfield Yard",
	"Zenithstone Zone",
}

// GetWorkshopName returns a workshop name from the pool based on index.
// Uses modulo to wrap around if index exceeds pool size.
func GetWorkshopName(index int) string {
	if index < 0 {
		index = 0
	}
	return WorkshopNames[index%len(WorkshopNames)]
}

// GetNextWorkshopName returns the next available workshop name given the count of existing workshops.
func GetNextWorkshopName(existingCount int) string {
	return GetWorkshopName(existingCount)
}
