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

type TtInfo struct {

	// Core structures maintaining timetable state

	NDays        int
	NHours       int
	SlotsPerWeek int
	PMStart      int
	LunchTimes   []int
	Activities   []*Activity // first entry (index 0) free!
	Resources    []any       // pointers to Resources
	TtSlots      []ActivityIndex

	// "Convenience" data

	Db      *base.DbTopLevel
	Ref2Tag map[Ref]string // Ref -> Tag mapping for subjects, teachers,
	// rooms, classes and groups

	// Set up by "orderResources",
	// used by "SortList" for ordering resource lists
	ResourceOrder map[Ref]int

	// Set up by "gatherCourseInfo"
	SuperSubs  map[Ref][]Ref       // SuperCourse -> list of its SubCourses
	CourseInfo map[Ref]*CourseInfo // Key can be Course or SuperCourse

	// Set by "filterDivisions"
	ClassDivisions map[Ref][][]Ref // Class -> list of list of Groups

	// Set by "makeAtomicGroups"
	AtomicGroups  map[Ref][]*AtomicGroup // Group -> list of AtomicGroups
	NAtomicGroups int

	Constraints map[string][]any

	MinDaysBetweenLessons []MinDaysBetweenLessons
	ParallelLessons       []ParallelLessons

	WITHOUT_ROOM_PLACEMENTS bool // ignore initial room placements
}

func MakeTtInfo(db *base.DbTopLevel) *TtInfo {
	ttinfo := &TtInfo{
		Db:         db,
		PMStart:    db.Info.FirstAfternoonHour,
		LunchTimes: db.Info.MiddayBreak,
	}

	gatherCourseInfo(ttinfo) // must be before call to filterDivisions

	// Build Ref -> Tag mapping for subjects, teachers, rooms, classes
	// and groups.
	ref2Tag := map[Ref]string{}
	for _, r := range db.Subjects {
		ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Rooms {
		ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Teachers {
		ref2Tag[r.Id] = r.Tag
	}

	// Get filtered divisions (only those with lessons)
	filterDivisions(ttinfo)

	// Handle the classes and groups (used for lessons)
	for cref, divs := range ttinfo.ClassDivisions {
		c := db.Elements[cref].(*base.Class)
		ctag := c.Tag
		ref2Tag[c.Id] = ctag
		ref2Tag[c.ClassGroup] = ctag
		for _, d := range divs {
			for _, gref := range d {
				gtag := db.Elements[gref].(*base.Group).Tag
				ref2Tag[gref] = ctag + CLASS_GROUP_SEP + gtag
			}
		}
	}

	ttinfo.Ref2Tag = ref2Tag
	//fmt.Printf("Ref2Tag: %v\n", ttinfo.Ref2Tag)
	ttinfo.orderResources()

	// Get "atomic" groups
	ttinfo.makeAtomicGroups()

	return ttinfo
}

func (ttinfo *TtInfo) PrepareCoreData() {
	db := ttinfo.Db
	ndays := len(db.Days)
	nhours := len(db.Hours)

	// Allocate a vector for pointers to all Resources: teachers, (atomic)
	// student groups and (real) rooms.
	// Allocate a vector for pointers to all Activities, keeping the first
	// entry free (0 should be an invalid ActivityIndex).
	// Allocate a vector for a week of time slots for each Resource. Each
	// cell represents a timetable slot for a single resource. If it is
	// occupied – by an ActivityIndex – that indicates which Activity is
	// using the Resource at this time. A value of -1 indicates that the
	// time slot is blocked for this Resource.

	lt := len(db.Teachers)
	lr := len(db.Rooms)
	lw := ndays * nhours

	ags := []*AtomicGroup{}
	g2ags := map[Ref][]ResourceIndex{}
	for _, cl := range db.Classes {
		for _, ag := range ttinfo.AtomicGroups[cl.ClassGroup] {
			ags = append(ags, ag)
			// Add to the Group -> index list map
			g2ags[cl.ClassGroup] = append(g2ags[cl.ClassGroup], ag.Index)
			for _, gref := range ag.Groups {
				g2ags[gref] = append(g2ags[gref], ag.Index)
			}
		}
	}

	// Sort the AtomicGroups
	slices.SortFunc(ags, func(a, b *AtomicGroup) int {
		if a.Index < b.Index {
			return -1
		}
		return 1
	})

	lg := len(ags)

	// If using a single vector for all slots:
	ttinfo.NDays = ndays
	ttinfo.NHours = nhours
	ttinfo.SlotsPerWeek = ndays * nhours
	ttinfo.Resources = make([]any, lt+lr+lg)
	ttinfo.TtSlots = make([]ActivityIndex, (lt+lr+lg)*lw)

	// The slice cells are initialized to 0 or nil, according to slice type.

	// Copy the AtomicGroups to the beginning of the Resources slice.
	i := 0
	for _, ag := range ags {
		if ag.Index != i {
			base.Bug.Fatalf("Atomic group index != resource index:\n"+
				"  -- %d: %+v\n", i, ag)
		}
		ttinfo.Resources[i] = ag
		//fmt.Printf(" :: %+v\n", ag)
		i++
	}
	ttinfo.NAtomicGroups = i

	t2tt := map[Ref]ResourceIndex{}
	for _, t := range db.Teachers {
		t2tt[t.Id] = i
		ttinfo.Resources[i] = t
		i++
	}
	r2tt := map[Ref]ResourceIndex{}
	for _, r := range db.Rooms {
		r2tt[r.Id] = i
		ttinfo.Resources[i] = r
		i++
	}

	// Add the pseudo activities due to the NotAvailable lists
	ttinfo.addBlockers(t2tt, r2tt)

	// Add the remaining Activity information
	ttinfo.addActivityInfo(t2tt, r2tt, g2ags)

	// The constraint "preprocessing" needs the Activity info, so it must be
	// placed here.
	ttinfo.processConstraints()
}

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
