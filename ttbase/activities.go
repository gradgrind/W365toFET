package ttbase

//TODO--

import (
	"W365toFET/base"
)

// These types are defined primarily for documentation, to make the types
// of the corresponding items a bit clearer.
type SlotIndex = TtIndex     // index of a time-slot within a week
type ResourceIndex = TtIndex // index into [TtInfo.Resources]
type ActivityIndex = TtIndex // index into [TtInfo.Activities]

type Activity struct {
	// Index is the offset of this activity in [TtInfo.Activities]
	Index ActivityIndex
	// Duration specifies the length in timetable-hours of this activity
	Duration int
	// Resources are teachers, student (atomic!) groups and rooms. They are
	// referenced using indexes into [TtInfo.Resources].
	// [Activity.Resources] lists those required by the Activity.
	Resources []ResourceIndex
	// XRooms is a list of chosen rooms, distinct from the required rooms
	XRooms []ResourceIndex
	// ExtendedGroups is a list of atomic group indexes for those groups
	// in the activity's class(es) which are NOT involved in the activity.
	//TODO: Do I really need this? It might rather confuse things
	ExtendedGroups []ResourceIndex
	// Fixed specifies whether the activity must remain in its current slot
	Fixed bool
	// Placement specifies the time-slot in which this activity has been
	// placed (day * daylength + hour), or -1 if unplaced
	Placement SlotIndex
	// PossibleSlots is a list of non-blocked time-slots for this activity
	PossibleSlots []SlotIndex
	// DifferentDays lists activities which must definitely be placed on a
	// different day to the present activity
	DifferentDays []ActivityIndex // hard constraint only
	// Parallel lists activities which must start at the same time as the
	// present activity.
	Parallel []ActivityIndex // hard constraint only

	// Access to basic information about this activity
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
}
