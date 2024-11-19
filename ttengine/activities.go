package ttengine

type SlotIndex int
type ResourceIndex int

type Activity struct {
	Index         int
	Duration      int
	Resources     []ResourceIndex
	PossibleSlots []SlotIndex
	Fixed         bool
	Placement     int // day * nhours + hour, or -1 if unplaced
}
