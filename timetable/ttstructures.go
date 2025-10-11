// Package timetable deals with timetables. It uses its own data structures,
// which are designed around the need to handle constraints and avoid
// "collisions" (e.g. a teacher in two classes at the same time).
package timetable

import (
	"W365toFET/base"
)

type NodeRef = base.Ref // node reference (UUID)

type ActivityIndex int
type TeacherIndex int
type RoomIndex int
type TtSlot int

type TtBackend struct {
	Run   func(tt_data *TtData, testing bool)
	Abort func(tt_data *TtData)
	Tick  func(tt_data *TtData)
	Clear func(tt_data *TtData)

	//TODO: This will probably need to return a more elaborate structure
	Results func(tt_data *TtData) []ActivityPlacement
}

var BACKEND TtBackend

type TtSharedData struct {
	Db           *base.DbTopLevel
	NDays        int
	NHours       int
	HoursPerWeek int

	// Directory which can be freely used by the timetable generator back-end
	WorkingDir string

	// `xxxNodes` are arrays mapping resource indexes to their corresponding
	// atomic group, teacher or room nodes (they contains pointers).
	AtomicNodes []base.Resource

	TeacherIndex map[NodeRef]TeacherIndex
	RoomIndex    map[NodeRef]RoomIndex

	// `AtomicGroups` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroups map[NodeRef][]AtomicIndex
	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Set up by `CollectCourses`, which calls `makeActivities`
	// Note that activity 0 is invalid, the first activity has index 1.
	Activities     []*Activity
	CourseInfoList []*CourseInfo
	Ref2CourseInfo map[NodeRef]*CourseInfo
}

// A TtData is the top-level structure for the timetable data.
type TtData struct {
	Description string

	SharedData *TtSharedData

	//TODO: In SharedData? Probably only if these are really not changed.
	// Each teacher, class and room has a matrix of days * hours cells
	// containing true in blocked slots, indexing: [item-index][day][hour].
	TeacherNotAvailable [][][]bool
	ClassNotAvailable   [][][]bool
	RoomNotAvailable    [][][]bool

	HardConstraints map[ConstraintType][]any
	SoftConstraints map[ConstraintType][]any

	WITHOUT_ROOM_PLACEMENTS bool // ignore room allocation constraints

	// `State` values:
	//		-1: not started (yet)
	//		 0: running
	//     	 1: finished successfully
	//		 2: failed (with errors or aborted)
	State    int
	Progress int // percent
	Ticks    int
	LastTime int // ticks at last Progress change
	Message  string

	BackEndData any // for use by the timetable generator itself

	/* TODO: These are not currently used. They are intended for keeping
	// an internal record of activity placements.

	// `ActivitySlots` is an array of activities + 1 entries, each entry being
	// the time slot in which the corresponding activity has been placed, or
	// -1 if unplaced. There is no activity with index 0.
	ActivitySlots []TtSlot

	// `ResourceWeeks` contains the allocations of the "resources" (atomic
	// groups, teachers, rooms) to activities (indexes). This is organized
	// as an array of "week-chunks" (`HoursPerWeek` entries), one for each
	// resource in the `Resources array`.
	ResourceWeeks []ActivityIndex
	*/
}

type ClassDivision struct {
	Class     *base.Class
	Divisions [][]NodeRef
}

// BasicSetup performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func BasicSetup(db *base.DbTopLevel, workingdir string) *TtData {
	days := len(db.Days)
	hours := len(db.Hours)
	tt_shared_data := &TtSharedData{
		Db:           db,
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
		WorkingDir:   workingdir,
	}
	tt_data := &TtData{
		SharedData: tt_shared_data,
	}

	// Collect ClassDivisions
	tt_shared_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_shared_data.MakeAtomicGroups()

	// Add teachers and rooms to resource array
	tt_shared_data.TeacherResources()
	tt_shared_data.RoomResources()

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_shared_data.CollectCourses()

	/* ... initially all activities unplaced
	tt_data.ActivitySlots = slices.Repeat(
		[]TtSlot{-1},
		len(tt_shared_data.Activities))
	*/

	tt_data.HardConstraints = map[ConstraintType][]any{}
	tt_data.SoftConstraints = map[ConstraintType][]any{}
	tt_data.preprocessConstraints()
	return tt_data
}

func (tt_shared_data *TtSharedData) TeacherResources() {
	tt_shared_data.TeacherIndex = map[NodeRef]TeacherIndex{}
	for i, t := range tt_shared_data.Db.Teachers {
		tt_shared_data.TeacherIndex[t.Id] = TeacherIndex(i)
	}
}

func (tt_shared_data *TtSharedData) RoomResources() {
	tt_shared_data.RoomIndex = map[NodeRef]RoomIndex{}
	for i, r := range tt_shared_data.Db.Rooms {
		tt_shared_data.RoomIndex[r.Id] = RoomIndex(i)
	}
}

type MinDaysBetweenActivities struct {
	// Result of processing constraints DifferentDays and DaysBetween
	Weight               int
	ConsecutiveIfSameDay bool
	Activities           []ActivityIndex
	MinDays              int
}

type ParallelLessons struct {
	Weight         int
	ActivityGroups [][]ActivityIndex
}

// This structure is used to return the placement results from the
// timetable back-end.
type ActivityPlacement struct {
	Id    ActivityIndex
	Day   int
	Hour  int
	Rooms []RoomIndex
}
