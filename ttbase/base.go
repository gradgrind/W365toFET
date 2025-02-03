// Package ttbase extracts data from the base db (DbTopLevel) and processes
// or reorganized it to be generally useful for timetable handling.
// It provides functions/methods for using this data.
package ttbase

import (
	"W365toFET/base"
	"slices"
	"strings"
)

type Ref = base.Ref

// Use alias types for indexes to aid documentation. Making them distinct
// types might help avoid errors, but it would necessitate conversions at
// certain points, which has pros and cons ...
type TtIndex = int // use this as the basic index type

const CLASS_GROUP_SEP = "."
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"

type VirtualRoom struct {
	Rooms       []Ref   // only ("real") Rooms
	RoomChoices [][]Ref // list of ("real") Room lists
}

// A TtInfo contains the core structures for the timetable
type TtInfo struct {
	NDays  int // number of days in the school week
	NHours int // number of timetable-hours in the school day
	// DayLength is longer than NHours – "dummy" time-slots are added at the
	// end of each day to ease handling of constraints which relate to days.
	// It should be (at least) 2 * NHours - 1.
	DayLength    int
	SlotsPerWeek int // NDays * DayLength
	PMStart      int // index (0-based) of first afternoon hour
	// LunchTimes is a contiguous, ordered list of hours (0-based indexes)
	// in which a lunch break can be taken
	LunchTimes []int
	// Resources provides indexed access to all resources (atomic student
	// groups, teachers, rooms), via pointers. Type "any" is used
	// rather than an interface because the resources are partly from
	// another package. The atomic student groups start at index 0.
	Resources []any // pointers to resource elements
	// Placements is used for lesson (in the form of TtLesson items) placement
	Placements *TtPlacement

	// "Convenience" data

	// Db is a reference to the underlying school data
	Db *base.DbTopLevel
	// Ref2Tag is a mapping, Ref -> Tag, for subjects, teachers,
	// rooms, classes and groups
	Ref2Tag map[Ref]string

	// ResourceOrder is a map set up by [orderResources] and used by
	// [SortList] for ordering resource lists
	ResourceOrder map[Ref]int

	// LessonCourses is an array of pointers to the [CourseInfo] items
	// and is set up by [gatherCourseInfo]
	LessonCourses []*CourseInfo
	// CourseInfo maps the [base.Course] or [base.SuperCourse] references
	// to their [CourseInfo] items
	CourseInfo map[Ref]*CourseInfo

	// ClassDivision maps a [base.Class] reference to a list of lists of
	// [base.Group] elements. It is set by [filterDivisions]
	ClassDivisions map[Ref][][]Ref

	// AtomicGroupIndexes maps a [base.Group] reference to a list of
	// resource indexes (for the AtomicGroup elements)
	AtomicGroupIndexes map[Ref][]ResourceIndex
	// TeacherIndexes maps a [base.Teacher] reference to the resource index
	// (for the Teacher elements)
	TeacherIndexes map[Ref]ResourceIndex
	// RoomIndexes maps a [base.Room] reference to the resource index
	// (for the Room elements)
	RoomIndexes map[Ref]ResourceIndex
	// NAtomicGroups ist the total number of [AtomicGroup] items.
	NAtomicGroups int

	Constraints       map[string][]any
	DayGapConstraints *DayGapConstraints
	// HardParallelCourses gives a list of parallel courses for each course,
	// set in method [addParallelCoursesConstraint]
	HardParallelCourses map[Ref][]Ref
	// SoftParallelCourses maps each course to a list of its soft
	// parallel constraints, set in method [addParallelCoursesConstraint]
	SoftParallelCourses []*base.ParallelCourses

	//TODO
	WITHOUT_ROOM_PLACEMENTS bool // ignore initial room placements
}

// pResources returns a string representation of a list of "resources",
// using their Tag fields.
func (ttinfo *TtInfo) pResources(rlist []Ref) string {
	slist := make([]string, len(rlist))
	for i, r := range rlist {
		slist[i] = ttinfo.Ref2Tag[r]
	}
	return strings.Join(slist, ",")
}

// MakeTtInfo makes a new TtInfo object and initializes some of its
// properties.
func MakeTtInfo(db *base.DbTopLevel) *TtInfo {
	ndays := len(db.Days)
	nhours := len(db.Hours)
	daylength := nhours*2 - 1
	ttinfo := &TtInfo{
		Db: db,
		//
		NDays:        ndays,
		NHours:       nhours,
		DayLength:    daylength,
		SlotsPerWeek: ndays * daylength,
	}

	// Build Ref -> Tag mapping for subjects, teachers, rooms, classes
	// and groups. Set up this mapping for subjects, rooms and teachers.
	ref2Tag := map[Ref]string{}
	ttinfo.Ref2Tag = ref2Tag
	for _, r := range db.Subjects {
		ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Rooms {
		ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Teachers {
		ref2Tag[r.Id] = r.Tag
	}

	// Generate the [CourseInfo] items and set up the references
	// [TtInfo.CourseInfo] and [TtInfo.LessonCourses].
	ttinfo.collectCourses()

	// Get filtered class divisions (only those with lessons). This uses
	// the filtered courses (only those with lessons) set up by the call to
	// [gatherCourseInfo].
	ttinfo.filterDivisions()

	// Handle the classes and groups (those used for lessons)
	for cref, divs := range ttinfo.ClassDivisions {
		c := db.Elements[cref].(*base.Class)
		ctag := c.Tag
		ref2Tag[c.Id] = ctag         // reference -> tag for Class item
		ref2Tag[c.ClassGroup] = ctag // reference -> tag for class-Group item
		// Add map entries for the other Group items
		for _, d := range divs {
			for _, gref := range d {
				gtag := db.Elements[gref].(*base.Group).Tag
				ref2Tag[gref] = ctag + CLASS_GROUP_SEP + gtag
			}
		}
	}

	// Prepare ordered list for the teachers, groups and rooms (used
	// for printing tag lists)
	ttinfo.orderResources()

	// Get "atomic" groups. The first elements of the resources list
	// [TtInfo.Resources] are allocated to the atomic groups.
	ttinfo.makeAtomicGroups()

	// Get preliminary constraint info, performing some checks
	ttinfo.processConstraints()

	return ttinfo
}

// PrepareCoreData adds teachers and (real) rooms to the resources list
// ([TtInfo.Resources]). It also sets the Resources field of the [CourseInfo]
// items. For the rooms that is only the "necessary" rooms, the rooms where
// a choice can be made are not included.
//
// The data structures for keeping track of placements are set up at
// [TtInfo.Placements]. This includes an array of pointers to all the
// [TtLesson] items, keeping the first entry free (0 should be an invalid
// lesson index).
// Also an array of time-slot-weeks is set up. Each resource has a block of
// time-slots representing a timetable week, arranged as an array, the
// resource order being the same as in the [TtInfo.Resource] array.
// Each time-slot contains the index of the lesson claiming that resource in
// the time-slot in question. The value may also be 0 (resource free) or -1
// (resource blocked – "not available").
func (ttinfo *TtInfo) PrepareCoreData() {
	db := ttinfo.Db

	// The AtomicGroup pointers are already at the beginning of the resources
	// list. Add the teachers and rooms

	ttinfo.TeacherIndexes = map[Ref]ResourceIndex{}
	for _, t := range db.Teachers {
		ttinfo.TeacherIndexes[t.Id] = len(ttinfo.Resources)
		ttinfo.Resources = append(ttinfo.Resources, t)
	}
	ttinfo.RoomIndexes = map[Ref]ResourceIndex{}
	for _, r := range db.Rooms {
		ttinfo.RoomIndexes[r.Id] = len(ttinfo.Resources)
		ttinfo.Resources = append(ttinfo.Resources, r)
	}

	// Check the allocated rooms at Lesson.Rooms
	ttinfo.checkAllocatedRooms()

	// Set the Resources field of the [CourseInfo] items
	ttinfo.collectCourseResources()

	// Set up the lesson-placement mechanism with [TtInfo.Placements]
	ttinfo.PrepareActivityGroups()

	ttinfo.processDaysBetweenConstraints()

	// Place fixed lessons
	ttinfo.initialFixedPlacement()
	// Set up possible slots for non-fixed lessons
	ttinfo.makePossibleSlots()
	// Place any non-fixed lessons which have a placement
	ttinfo.initialNonFixedPlacement()
	// Allocate supplied room choices as far as possible
	ttinfo.initialRoomChoices()
}

// orderResources generates an ordering index for each of the resources,
// saving the result at [TtInfo.ResourceOrder].
func (ttinfo *TtInfo) orderResources() {
	// Needed for sorting teachers, groups and rooms
	db := ttinfo.Db
	i := 0
	olist := map[base.Ref]int{}
	for _, t := range db.Teachers {
		olist[t.Id] = i
		i++
	}
	for _, r := range db.Rooms {
		olist[r.Id] = i
		i++
	}
	for _, c := range db.Classes {
		olist[c.ClassGroup] = i
		i++
		for _, div := range ttinfo.ClassDivisions[c.Id] {
			for _, gref := range div {
				olist[gref] = i
				i++
			}
		}
	}
	ttinfo.ResourceOrder = olist
}

// SortList sorts a list of resource references according to the order
// in [TtInfo.ResourceOrder]. It returns a list of tags (short names).
func (ttinfo *TtInfo) SortList(list []Ref) []string {
	ordering := ttinfo.ResourceOrder
	ref2tag := ttinfo.Ref2Tag
	olist := []string{}
	if len(list) > 1 {
		slices.SortFunc(list, func(a, b Ref) int {
			if ordering[a] < ordering[b] {
				return -1
			}
			return 1
		})
		for _, ref := range list {
			olist = append(olist, ref2tag[ref])
		}
	} else if len(list) == 1 {
		olist = append(olist, ref2tag[list[0]])
	}
	return olist
}
