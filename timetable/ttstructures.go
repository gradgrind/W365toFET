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
type TtSlot int16

// A TtData is the top-level structure for the timetable data.
type TtData struct {
	Db           *base.DbTopLevel
	NDays        int
	NHours       int
	HoursPerWeek int
	// `ActivitySlots` is an array of activities + 1 entries, each entry being
	// the time slot in which the corresponding activity has been placed, or
	// -1 if unplaced. There is no activity with index 0.
	ActivitySlots []TtSlot
	// `ResourceWeeks` contains the allocations of the "resources" (atomic
	// groups, teachers, rooms) to activities (indexes). This is organized
	// as an array of "week-chunks" (`HoursPerWeek` entries), one for each
	// resource in the `Resources array`.
	ResourceWeeks []ActivityIndex

	// `Resources` is an array mapping resource indexes to their corresponding
	// atomic group, teacher or room nodes (it contains pointers).
	Resources    []base.Resource //TODO: or any?
	RoomIndex    map[NodeRef]ResourceIndex
	TeacherIndex map[NodeRef]ResourceIndex

	// `AtomicGroups` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroups map[NodeRef][]ResourceIndex
	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Set up by `MakeActivities`
	Activities []*Activity
	//?? ActivityCourses []*TtCourseInfo
	CourseInfoList []*CourseInfo
	Ref2CourseInfo map[NodeRef]*CourseInfo

	Constraints map[string][]any

	MinDaysBetweenLessons []MinDaysBetweenLessons
	ParallelLessons       []ParallelLessons

	WITHOUT_ROOM_PLACEMENTS bool // ignore room allocation constraints
}

type ClassDivision struct {
	Class     *base.Class
	Divisions [][]NodeRef
}

// BasicSetup performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func BasicSetup(db *base.DbTopLevel) *TtData {
	days := len(db.Days)
	hours := len(db.Hours)
	tt_data := &TtData{
		Db:           db,
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
	}

	// Collect ClassDivisions
	tt_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_data.MakeAtomicGroups()

	// Add teachers and rooms to resource array
	tt_data.TeacherResources()
	tt_data.RoomResources()
	tt_data.ResourceWeeks = make([]ActivityIndex,
		(len(tt_data.Resources))*days*hours)

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_data.CollectCourses()

	// ... initially all activities unplaced
	tt_data.ActivitySlots = slices.Repeat(
		[]TtSlot{-1},
		len(tt_data.Activities))

	tt_data.processConstraints()

	//for _, mdbl := range tt_data.MinDaysBetweenLessons {
	//	fmt.Printf("§§§ %v\n", mdbl)
	//}

	return tt_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.TeacherIndex = map[NodeRef]ResourceIndex{}
	for _, t := range tt_data.Db.Teachers {
		i := len(tt_data.Resources)
		tt_data.TeacherIndex[t.Id] = i
		tt_data.Resources = append(tt_data.Resources, t)
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.RoomIndex = map[NodeRef]ResourceIndex{}
	for _, r := range tt_data.Db.Rooms {
		i := len(tt_data.Resources)
		tt_data.RoomIndex[r.Id] = i
		tt_data.Resources = append(tt_data.Resources, r)
	}
}

type MinDaysBetweenLessons struct {
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

// Structures and methods used in connection with the timetable "back-end"

type TtHandler interface {
	Update(*TtInstance)
}

type TtChainedFunc struct {
	// `Delay` specifies the number of ticks after which the function should
	// be called. 0 specifies an infinite delay (i.e. the function must be
	// started in some other way), -1 specifies that the function will never
	// be called – possibly because it has already been started.
	Delay int
	Func  func(*TtRunData)
}

type TtInstance struct {
	Id    int
	Ticks int
	// `LastState` values:
	//    0-100: progress in percent
	//       -1: finished successfully
	//       -2: failed
	//       -3: cancelled
	LastState int
	// `LastTime` is some representation of the elapsed time at which
	// the `LastState` field was last updated.
	LastTime string
	// `HandlerData` provides a field to be used by the timetable "back-end".
	Message          string
	HandlerData      any
	UpdateHandler    func(instance *TtInstance)
	Abort            func(any) // pass HandlerData
	SuccessPath      TtChainedFunc
	SuccessInstances []*TtInstance
	FailurePath      TtChainedFunc
	FailureInstances []*TtInstance
	OtherPath        []TtChainedFunc
	OtherInstances   []*TtInstance
}

type TtRunData struct {
	TtData_0   *TtData // original data
	TtData     *TtData // current (modified) data
	WorkingDir string

	Stop        chan bool
	NewInstance chan *TtInstance

	//TODO: Use len(Instances) instead of RunCounter?
	//RunCounter int // count subprocesses, for indexing
	Instances []*TtInstance
	Active    map[int]struct{}
}

func (rundata *TtRunData) GetHandlerData(i int) any {
	return rundata.Instances[i].HandlerData
}

func (rundata *TtRunData) SetHandlerData(i int, data any) {
	rundata.Instances[i].HandlerData = data
}

func (rundata *TtRunData) InstanceUpdate(i int, cc int, t string) {
	tti := rundata.Instances[i]
	tti.LastState = cc
	tti.LastTime = t
}

func (rundata *TtRunData) InstanceTerminated(i int, msg string) {
	rundata.Instances[i].Message = msg
	delete(rundata.Active, i)
}
