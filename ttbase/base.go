package ttbase

import (
	"W365toFET/base"
)

type Ref = base.Ref

const CLASS_GROUP_SEP = "."
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"
const LUNCH_BREAK_TAG = "-lb-"
const LUNCH_BREAK_NAME = "Lunch Break"

type VirtualRoom struct {
	Rooms       []Ref   // only ("real") Rooms
	RoomChoices [][]Ref // list of ("real") Room lists
}

// TODO: Some of this might be better placed in DbTopLevel?
type TtInfo struct {
	Db      *base.DbTopLevel
	Ref2Tag map[Ref]string // Ref -> Tag mapping for subjects, teachers,
	// rooms, classes and groups

	// Set up by "gatherCourseInfo"
	SuperSubs  map[Ref][]Ref       // SuperCourse -> list of its SubCourses
	CourseInfo map[Ref]*CourseInfo // Key can be Course or SuperCourse
	TtLessons  []TtLesson

	// Set by filterDivisions
	ClassDivisions map[Ref][][]Ref // Class -> list of list of Groups

	// Set by makeAtomicGroups
	AtomicGroups map[Ref][]*AtomicGroup // Group -> list of AtomicGroups
}

func MakeTtInfo(db *base.DbTopLevel) *TtInfo {
	ttinfo := &TtInfo{
		Db: db,
	}
	gatherCourseInfo(ttinfo)

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

	// Get "atomic" groups
	makeAtomicGroups(ttinfo)

	return ttinfo
}
