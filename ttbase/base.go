package ttbase

import (
	"W365toFET/base"
	"fmt"
	"strings"
)

// TODO: Would it be better to use internal references – small integers?
// There could be a list to map them (directly) to the referenced items.
type Ref = base.Ref

const CLASS_GROUP_SEP = "."
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"
const LUNCH_BREAK_TAG = "-lb-"
const LUNCH_BREAK_NAME = "Lunch Break"

type VirtualRoom struct {
	Rooms       []Ref   // only Rooms
	RoomChoices [][]Ref // list of ("real") Room lists
}

// Possibly helpful when testing
func (ttinfo *TtInfo) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, ttinfo.Ref2Tag[t])
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, ttinfo.Ref2Tag[g])
	}

	return fmt.Sprintf("<Course %s/%s:%s>\n",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		ttinfo.Ref2Tag[cinfo.Subject],
	)
}

type idMap struct {
	activityId int
	baseId     Ref
}

// TODO: Some of this might be better placed in DbTopLevel?
type TtInfo struct {
	Db      *base.DbTopLevel
	Ref2Tag map[Ref]string

	Days  []string //TODO: rather NDays?
	Hours []string //TODO: rather NHours?
	// These cover only courses and groups with lessons:
	ONLY_FIXED bool // normally true, false allows generation of
	// placement constraints for non-fixed lessons
	WITHOUT_ROOM_PLACEMENTS bool // ?

	// Set up by "gatherCourseInfo"
	SuperSubs  map[Ref][]Ref
	CourseInfo map[Ref]*CourseInfo // Key can be Course or SuperCourse
	TtLessons  []TtLesson

	// Set by filterDivisions
	ClassDivisions map[Ref][][]Ref // value is list of list of groups

	// Set by makeAtomicGroups
	AtomicGroups map[Ref][]*AtomicGroup

	//?
	ttVirtualRooms map[string]string // cache for tt virtual rooms,
	// "hash" -> tt-virtual-room tag
	ttVirtualRoomN          map[string]int // tt-virtual-room tag -> number of room sets
	differentDayConstraints map[Ref][]int  // Retain the indexes of the entries
	// in the ConstraintMinDaysBetweenActivities list for each course. This
	// allows the default constraints to be modified later.
}

func MakeTtInfo(db *base.DbTopLevel) *TtInfo {
	ttinfo := &TtInfo{
		Db:                      db,
		ONLY_FIXED:              true,
		WITHOUT_ROOM_PLACEMENTS: true,
	}
	gatherCourseInfo(ttinfo)

	// Build Ref -> Tag mapping for subjects, teachers, rooms, classes
	// and groups.
	Ref2Tag := map[Ref]string{}
	for _, r := range db.Subjects {
		Ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Rooms {
		Ref2Tag[r.Id] = r.Tag
	}
	for _, r := range db.Teachers {
		Ref2Tag[r.Id] = r.Tag
	}

	// Get filtered divisions (only those with lessons)
	filterDivisions(ttinfo)

	// Handle the classes and groups (used for lessons)
	for cref, divs := range ttinfo.ClassDivisions {
		c := db.Elements[cref].(*base.Class)
		ctag := c.Tag
		Ref2Tag[c.Id] = ctag
		Ref2Tag[c.ClassGroup] = ctag
		for _, d := range divs {
			for _, gref := range d {
				gtag := db.Elements[gref].(*base.Group).Tag
				Ref2Tag[gref] = ctag + CLASS_GROUP_SEP + gtag
			}
		}
	}

	//fmt.Printf("Ref2Tag: %v\n", Ref2Tag)

	// Get "atomic" groups
	makeAtomicGroups(ttinfo)

	//TODO--
	/*
		getDays(&ttinfo)
		getHours(&ttinfo)
		getTeachers(&ttinfo)
		getSubjects(&ttinfo)
		getRooms(&ttinfo)
		fmt.Println("=====================================")
		gatherCourseInfo(&ttinfo)

		//readCourseIndexes(&ttinfo)
		makeAtomicGroups(&ttinfo)
		//fmt.Println("\n +++++++++++++++++++++++++++")
		//printAtomicGroups(&ttinfo)
		getClasses(&ttinfo)
		lessonIdMap := getActivities(&ttinfo)

		addTeacherConstraints(&ttinfo)
		addClassConstraints(&ttinfo)

		getExtraConstraints(&ttinfo)

		// Convert lessonIdMap to string
		idmlines := []string{}
		for _, idm := range lessonIdMap {
			idmlines = append(idmlines,
				strconv.Itoa(idm.activityId)+":"+string(idm.baseId))
		}
		lidmap := strings.Join(idmlines, "\n")

		return xml.Header + makeXML(ttinfo.ttdata, 0), lidmap
	*/
	return ttinfo
}

//TODO--?
/*
func getString(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		b, _ := json.Marshal(val)
		s = string(b)
	}
	return s
}
*/
