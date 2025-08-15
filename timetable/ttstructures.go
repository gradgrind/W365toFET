// Package timetable deals with timetables. It uses its own data structures,
// which are designed around the need to handle constraints and avoid
// "collisions" (e.g. a teacher in two classes at the same time).
package timetable

import (
	"W365toFET/base"
	"slices"
)

type NodeRef = base.Ref // node reference (UUID)

type ActivityIndex int16
type ResourceIndex = int
type TimeSlot int16

type TtActivity struct {
	Id       ActivityIndex
	Duration int16
	Fixed    bool
	//BasicActivityGroup *BasicActivityGroup
	Resources   []ResourceIndex
	RoomChoices [][]ResourceIndex
	//TODO: Constraints to be applied when placing manually?
	//Constraints []TtConstraint
}

type TtRoom struct {
	Id       NodeRef
	Tag      string
	Resource ResourceIndex
}

type TtVirtualRoom struct {
	RoomIndexes       []ResourceIndex
	RoomChoiceIndexes [][]ResourceIndex
}

type Placement struct {
	Activity ActivityIndex
	Slot     TimeSlot
}

type TimetableUnit struct {
	Activities []ActivityIndex
	Placements [][]TimeSlot
	//TODO: Constraints to be applied after a TimetableUnit has been placed?
	//Constraints []TtConstraint
	Next int
}

// A TtData is the top-level structure for the timetable data.
type TtData struct {
	NDays        int
	NHours       int
	HoursPerWeek int
	// `ActivitySlots` is an array of activities + 1 entries, each entry being
	// the time slot in which the corresponding activity has been placed, or
	// -1 if unplaced. There is no activity with index 0.
	ActivitySlots []TimeSlot
	// `ResourceWeeks` contains the allocations of the "resources" (atomic
	// groups, teachers, rooms) to activities (indexes). This is organized
	// as an array of "week-chunks" (`HoursPerWeek` entries), one for each
	// resource in the `Resources array`.
	ResourceWeeks []ActivityIndex

	// `Resources` is an array mapping resource indexes to their corresponding
	// atomic group, teacher or room nodes (it contains pointers).
	Resources    []any
	RoomIndex    map[NodeRef]ResourceIndex
	TeacherIndex map[NodeRef]ResourceIndex

	// `AtomicGroups` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroups map[NodeRef][]ResourceIndex
	// `ClassDivisions` maps a class NodeRef to its list of divisions, each of
	// these being a list of group NodeRefs.
	ClassDivisions map[NodeRef][][]NodeRef

	// Set up by `MakeActivities`
	Activities []*TtActivity
	//?? ActivityCourses []*TtCourseInfo
	//?? CourseInfo      map[NodeRef]*TtCourseInfo // key is Course or SuperCourse

	/*???
	DayIndex     map[string]int
	HourIndex    map[string]int
	GroupIndexes map[string][]ResourceIndex
	VirtualRooms map[string]TtVirtualRoom
	*/

	//?? basic_activity_groups map[int]*BasicActivityGroup
	//?? CollectedBags         map[*BasicActivityGroup]*BagCollection
}

// BasicSetup performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func BasicSetup(db *base.DbTopLevel) *TtData {
	days := len(db.Days)
	hours := len(db.Hours)
	tt_data := &TtData{
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
		//?? ActivitySlots: slices.Repeat([]TimeSlot{-1}, activities+1),
	}

	course_info := CollectCourses(db)
	class_divisions := FilterDivisions(db, course_info)

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_data.MakeAtomicGroups(db, class_divisions)

	// Add teachers and rooms to resource array
	tt_data.TeacherResources(db)
	tt_data.RoomResources(db)
	tt_data.ResourceWeeks = make([]ActivityIndex,
		(len(tt_data.Resources))*days*hours)

	// Get the activities for the timetable
	tt_data.MakeActivities(db, course_info)
	// ... initially all unplaced
	tt_data.ActivitySlots = slices.Repeat(
		[]TimeSlot{-1},
		len(tt_data.Activities))

	// Add the pseudo activities due to the NotAvailable lists of classes,
	// teachers and rooms.
	tt_data.BlockResources(db)

	/* TODO
	// Get preliminary constraint info – needed for the call to addActivity
	ttinfo.processConstraints()

	// Add the remaining Activity information
	ttinfo.addActivityInfo(t2tt, r2tt, g2ags)
	*/

	return tt_data
}

func (tt_data *TtData) BlockResource(resource ResourceIndex, slot TimeSlot) {
	tt_data.ResourceWeeks[int(resource)*tt_data.HoursPerWeek+int(slot)] = -1
}

func (tt_data *TtData) TeacherResources(db *base.DbTopLevel) {
	tt_data.TeacherIndex = map[NodeRef]ResourceIndex{}
	for _, t := range db.Teachers {
		i := len(tt_data.Resources)
		tt_data.TeacherIndex[t.Id] = i
		tt_data.Resources = append(tt_data.Resources, t)
	}
}

func (tt_data *TtData) RoomResources(db *base.DbTopLevel) {
	tt_data.RoomIndex = map[NodeRef]ResourceIndex{}
	for _, r := range db.Rooms {
		i := len(tt_data.Resources)
		tt_data.RoomIndex[r.Id] = i
		tt_data.Resources = append(tt_data.Resources, r)
	}
}
