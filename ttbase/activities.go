package ttbase

import (
	"W365toFET/base"
	"slices"
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

// addActivityInfo completes the initialization of the Activities. This
// includes the placement in the timetable structures of all the lessons
// (fixed, and also non-fixed with a placement). In this way various errors
// can be checked for.
func (ttinfo *TtInfo) addActivityInfo() {
	//g2tt := ttinfo.AtomicGroupIndexes

	// Complete the Activity items for each course
	r2tt := ttinfo.RoomIndexes
	for _, cinfo := range ttinfo.LessonCourses {
		// Get the room-choices. Check against room choices in course.

		//TODO: This needs testing with data that provides choice allocations.

		crooms := cinfo.Room.Rooms       // "necessary" rooms
		xrooms := cinfo.Room.RoomChoices // list of room choices
		//fmt.Printf("++ COURSE: %s\n", ttinfo.View(cinfo))
		for _, aix := range cinfo.Lessons {
			a := ttinfo.Activities[aix]
			// Check each actually allocated room
			rlist := []Ref{}
			for _, rref := range a.Lesson.Rooms {
				if slices.Contains(crooms, rref) {
					// Ignore "necessary" rooms
					continue
				}
				rlist = append(rlist, rref)
			}
			var xr []Ref
			if len(rlist) != 0 {
				// Try to match rooms in rlist to the choice list, xrooms
				xr = rclfunc(xrooms, rlist)
				if xr == nil {
					base.Error.Printf("Rooms (%s) used for lesson of"+
						" course %s:\n  rooms don't match.\n",
						ttinfo.pResources(rlist), ttinfo.View(cinfo))
				}
			}
			a.XRooms = make([]ResourceIndex, len(xrooms))
			var ri ResourceIndex
			for i := 0; i < len(xrooms); i++ {
				ri = -1
				if i < len(xr) {
					r := xr[i]
					if r != "" {
						ri = r2tt[r]
					}
				}
				a.XRooms[i] = ri
			}
			//fmt.Printf("  -- %+v\n", a.XRooms)
		}

		//

	}

}

// rclfunc uses recursion to match the rooms to the room-choice list,
// using "" where a choice has no room allocated
func rclfunc(rclist [][]Ref, rlist []Ref) []Ref {
	if len(rclist) == 0 {
		if len(rlist) == 0 {
			return []Ref{}
		} else {
			return nil
		}
	}
	rc := rclist[0]
	rlx := make([]Ref, len(rlist)-1)
	for i, r := range rlist {
		if slices.Contains(rc, r) {
			// Remove the room from the list
			rlx = rlx[:i]
			copy(rlx, rlist)
			rlx = append(rlx, rlist[i+1:]...)
			rl := rclfunc(rclist[1:], rlx)
			if rl != nil {
				return append([]Ref{r}, rl...)
			}
		}
	}
	// Assume there was no room supplied for this choice
	rl := rclfunc(rclist, rlx)
	if rl != nil {
		return append([]Ref{""}, rl...)
	}
	return nil
}
