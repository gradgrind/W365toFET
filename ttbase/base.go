package ttbase

import (
	"W365toFET/base"
	"slices"
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
	SlotsPerWeek int // NDays * NHours
	PMStart      int // index (0-based) of first afternoon hour
	// LunchTimes is a contiguous, ordered list of hours (0-based indexes)
	// in which a lunch break can be taken
	LunchTimes []int
	// Activities provides indexed access to all the [Activity] items, via
	// pointers. The first item should be free, as index 0 is used to
	// indicate "no activity".
	Activities []*Activity
	// Resources provides indexed access to all resources (teachers,
	// atomic student groups, rooms), via pointers. Type any is used
	// rather than an interface because the resources are partly from
	// another package.
	Resources []any // pointers to resource elements
	// TtSlots contains a full week of time-slots for each resource,
	// with the same indexing as Resources
	TtSlots []ActivityIndex

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

	// AtomicGroups maps a [base.Group] reference to a list of pointers
	// to the group's [AtomicGroup] items. It is set by [makeAtomicGroups]
	AtomicGroups map[Ref][]*AtomicGroup
	// AtomicGroupIndexes maps a [base.Group] reference to a list of
	// resource indexes
	AtomicGroupIndexes map[Ref][]ResourceIndex
	// NAtomicGroups ist the total number of [AtomicGroup] items.
	NAtomicGroups int

	Constraints       map[string][]any
	DayGapConstraints *DayGapConstraints

	MinDaysBetweenLessons []MinDaysBetweenLessons
	ParallelLessons       []ParallelLessons

	WITHOUT_ROOM_PLACEMENTS bool // ignore initial room placements
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
	// and groups. Set up the mapping for subjects, rooms and teachers.
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

	// Get the course info and generate the Activities list – though some
	// fields will not yet be properly set.
	gatherCourseInfo(ttinfo) // must be before call to filterDivisions

	// Get filtered class divisions (only those with lessons)
	filterDivisions(ttinfo)

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

	// Get "atomic" groups. The resources list (ttinfo.Resources) is begun
	// with the atomic groups.
	ttinfo.makeAtomicGroups()

	return ttinfo
}

// PrepareCoreData adds teachers and (real) rooms to the resources list
// (ttinfo.Resources).
// Also an array of pointers to all the Activities is set up, keeping the
// first entry free (0 should be an invalid activity index).
// Also an array of time-slot-weeks is set up. Each resource has a block of
// time-slots representing a timetable week, arranged as an array in
// [TtInfo.TimeSlots], the resource order being the same as in the
// [TtInfo.Resource] array. Each time-slot contains the index of the
// activity (in [TtInfo.Activities]) claiming that resource in the
// time-slot in question. The value may also be 0 (resource free) or -1
// (resource blocked – "not available").
func (ttinfo *TtInfo) PrepareCoreData() {
	db := ttinfo.Db

	lt := len(db.Teachers)
	lr := len(db.Rooms)

	resix := ttinfo.NAtomicGroups
	// Size the time-slot-array:
	ttinfo.TtSlots = make([]ActivityIndex, (lt+lr+resix)*ttinfo.SlotsPerWeek)

	// The AtomicGroup pointers are already at the beginning of the resources
	// list. Add the teachers and rooms

	t2tt := map[Ref]ResourceIndex{}
	for _, t := range db.Teachers {
		t2tt[t.Id] = resix
		ttinfo.Resources = append(ttinfo.Resources, t)
		resix++
	}
	r2tt := map[Ref]ResourceIndex{}
	for _, r := range db.Rooms {
		r2tt[r.Id] = resix
		ttinfo.Resources = append(ttinfo.Resources, r)
		resix++
	}

	// Add the pseudo-activities arising from the NotAvailable lists
	ttinfo.addBlockers(t2tt, r2tt)

	// Get preliminary constraint info – needed for the call to addActivity
	ttinfo.processConstraints()

	// Add the remaining Activity information
	ttinfo.addActivityInfo(t2tt, r2tt)
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
