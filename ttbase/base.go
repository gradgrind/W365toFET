package ttbase

import (
	"W365toFET/base"
	"encoding/json"
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

type virtualRoom struct {
	rooms       []Ref   // only Rooms
	roomChoices [][]Ref // list of pure Room lists
}

// Possibly helpful when testing
func (ttinfo *TtInfo) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, ttinfo.ref2tt[t])
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, ttinfo.ref2tt[g])
	}

	return fmt.Sprintf("<Course %s/%s:%s>\n",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		ttinfo.ref2tt[cinfo.Subject],
	)
}

type idMap struct {
	activityId int
	baseId     Ref
}

// Some of this might be better placed in DbTopLevel?
type TtInfo struct {
	db            *base.DbTopLevel
	ref2tt        map[Ref]string
	ref2grouponly map[Ref]string // includes only genuine groups
	days          []string
	hours         []string
	// These cover only courses and groups with lessons:
	ONLY_FIXED bool // normally true, false allows generation of
	// placement constraints for non-fixed lessons
	WITHOUT_ROOM_PLACEMENTS bool
	superSubs               map[Ref][]Ref
	courseInfo              map[Ref]CourseInfo // Key can be Course or SuperCourse
	classDivisions          map[Ref][][]Ref    // value is list of list of groups
	atomicGroups            map[Ref][]AtomicGroup
	ttVirtualRooms          map[string]string // cache for tt virtual rooms,
	// "hash" -> tt-virtual-room tag
	ttVirtualRoomN          map[string]int // tt-virtual-room tag -> number of room sets
	differentDayConstraints map[Ref][]int  // Retain the indexes of the entries
	// in the ConstraintMinDaysBetweenActivities list for each course. This
	// allows the default constraints to be modified later.
}

func MakeTtFile(dbdata *base.DbTopLevel) (string, string) {
	// Build ref-index -> tt-key mapping
	ref2tt := map[Ref]string{}
	for _, r := range dbdata.Subjects {
		ref2tt[r.Id] = r.Tag
	}
	for _, r := range dbdata.Rooms {
		ref2tt[r.Id] = r.Tag
	}
	for _, r := range dbdata.Teachers {
		ref2tt[r.Id] = r.Tag
	}
	ref2grouponly := map[Ref]string{}
	for _, r := range dbdata.Groups {
		if r.Tag != "" {
			ref2grouponly[r.Id] = r.Tag
		}
	}
	for _, r := range dbdata.Classes {
		ref2tt[r.Id] = r.Tag
		ref2tt[r.ClassGroup] = r.Tag
		// Handle the groups
		for _, d := range r.Divisions {
			for _, g := range d.Groups {
				ref2tt[g] = r.Tag + CLASS_GROUP_SEP + ref2grouponly[g]
			}
		}
	}

	//fmt.Printf("ref2tt: %v\n", ref2tt)

	/*
		ttinfo := ttInfo{
			db:                      dbdata,
			ref2tt:                 ref2tt,
			ref2grouponly:           ref2grouponly,
			ONLY_FIXED:              true,
			WITHOUT_ROOM_PLACEMENTS: true,
			ttVirtualRooms:         map[string]string{},
			ttVirtualRoomN:         map[string]int{},
			differentDayConstraints: map[Ref][]int{},
		}
	*/

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
	return "", ""
}

func getString(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		b, _ := json.Marshal(val)
		s = string(b)
	}
	return s
}
