// Package fet handles interaction with the fet timetabling program.
package fet

import (
	"W365toFET/base"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Ref = base.Ref

const CLASS_GROUP_SEP = "."
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"
const LUNCH_BREAK_TAG = "-lb-"
const LUNCH_BREAK_NAME = "Lunch Break"

const fet_version = "6.25.2"

// Function makeXML produces a chunk of pretty-printed XML output from
// the input data.
func makeXML(data interface{}, indent_level int) string {
	const indent = "  "
	prefix := strings.Repeat(indent, indent_level)
	xmlData, err := xml.MarshalIndent(data, prefix, indent)
	if err != nil {
		base.Error.Fatalf("%v\n", err)
	}
	return string(xmlData)
}

type fet struct {
	Version          string `xml:"version,attr"`
	Mode             string
	Institution_Name string
	Comments         string // this can be a source reference
	Days_List        fetDaysList
	Hours_List       fetHoursList
	Teachers_List    fetTeachersList
	Subjects_List    fetSubjectsList
	Rooms_List       fetRoomsList
	Students_List    fetStudentsList
	//Buildings_List
	Activity_Tags_List     fetActivityTags
	Activities_List        fetActivitiesList
	Time_Constraints_List  timeConstraints
	Space_Constraints_List spaceConstraints
}

type virtualRoom struct {
	rooms       []Ref   // only Rooms
	roomChoices [][]Ref // list of pure Room lists
}

type courseInfo struct {
	subject    Ref
	groups     []Ref
	teachers   []Ref
	room       virtualRoom
	lessons    []*base.Lesson
	activities []int
}

func weight2fet(w int) string {
	if w == 0 {
		return "0"
	}
	if w == 100 {
		return "100"
	}
	wf := float64(w)
	n := wf + math.Pow(2, wf/12)
	wfet := 100.0 - 100.0/n
	return strconv.FormatFloat(wfet, 'f', 3, 64)
}

// Possibly helpful when testing
func (fetinfo *fetInfo) View(cinfo *courseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.teachers {
		tlist = append(tlist, fetinfo.ref2fet[t])
	}
	glist := []string{}
	for _, g := range cinfo.groups {
		glist = append(glist, fetinfo.ref2fet[g])
	}

	return fmt.Sprintf("<Course %s/%s:%s>\n",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		fetinfo.ref2fet[cinfo.subject],
	)
}

type idMap struct {
	activityId int
	baseId     Ref
}

type fetInfo struct {
	db            *base.DbTopLevel
	ref2fet       map[Ref]string
	ref2grouponly map[Ref]string
	days          []string
	hours         []string
	fetdata       fet
	// These cover only courses and groups with lessons:
	ONLY_FIXED bool // normally true, false allows generation of
	// placement constraints for non-fixed lessons
	WITHOUT_ROOM_PLACEMENTS bool
	superSubs               map[Ref][]Ref
	courseInfo              map[Ref]courseInfo // Key can be Course or SuperCourse
	classDivisions          map[Ref][][]Ref
	atomicGroups            map[Ref][]AtomicGroup
	fetVirtualRooms         map[string]string // cache for FET virtual rooms,
	// "hash" -> FET-virtual-room tag
	fetVirtualRoomN   map[string]int // FET-virtual-room tag -> number of room sets
	autoDifferentDays *base.AutomaticDifferentDays
	daysBetween       map[Ref][]*base.DaysBetween
}

type timeConstraints struct {
	XMLName xml.Name `xml:"Time_Constraints_List"`
	//
	ConstraintBasicCompulsoryTime          basicTimeConstraint
	ConstraintStudentsSetNotAvailableTimes []studentsNotAvailable
	ConstraintTeacherNotAvailableTimes     []teacherNotAvailable

	ConstraintActivityPreferredStartingTime []startingTime
	ConstraintActivityPreferredTimeSlots    []activityPreferredTimes
	ConstraintActivitiesPreferredTimeSlots  []preferredSlots
	ConstraintMinDaysBetweenActivities      []minDaysBetweenActivities
	ConstraintActivityEndsStudentsDay       []lessonEndsDay

	ConstraintStudentsSetMaxGapsPerDay                  []maxGapsPerDay
	ConstraintStudentsSetMaxGapsPerWeek                 []maxGapsPerWeek
	ConstraintStudentsSetMinHoursDaily                  []minLessonsPerDay
	ConstraintStudentsSetMaxHoursDaily                  []maxLessonsPerDay
	ConstraintStudentsSetIntervalMaxDaysPerWeek         []maxDaysinIntervalPerWeek
	ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour []maxLateStarts
	ConstraintStudentsSetMaxHoursDailyInInterval        []lunchBreak

	ConstraintTeacherMaxDaysPerWeek          []maxDaysT
	ConstraintTeacherMaxGapsPerDay           []maxGapsPerDayT
	ConstraintTeacherMaxGapsPerWeek          []maxGapsPerWeekT
	ConstraintTeacherMaxHoursDailyInInterval []lunchBreakT
	ConstraintTeacherMinHoursDaily           []minLessonsPerDayT
	ConstraintTeacherMaxHoursDaily           []maxLessonsPerDayT
	ConstraintTeacherIntervalMaxDaysPerWeek  []maxDaysinIntervalPerWeekT
}

type basicTimeConstraint struct {
	XMLName           xml.Name `xml:"ConstraintBasicCompulsoryTime"`
	Weight_Percentage int
	Active            bool
}

type spaceConstraints struct {
	XMLName                          xml.Name `xml:"Space_Constraints_List"`
	ConstraintBasicCompulsorySpace   basicSpaceConstraint
	ConstraintActivityPreferredRooms []roomChoice
	ConstraintActivityPreferredRoom  []placedRoom
}

type basicSpaceConstraint struct {
	XMLName           xml.Name `xml:"ConstraintBasicCompulsorySpace"`
	Weight_Percentage int
	Active            bool
}

func MakeFetFile(dbdata *base.DbTopLevel) (string, string) {
	// Build ref-index -> fet-key mapping
	ref2fet := map[Ref]string{}
	for _, r := range dbdata.Subjects {
		ref2fet[r.Id] = r.Tag
	}
	for _, r := range dbdata.Rooms {
		ref2fet[r.Id] = r.Tag
	}
	for _, r := range dbdata.Teachers {
		ref2fet[r.Id] = r.Tag
	}
	ref2grouponly := map[Ref]string{}
	for _, r := range dbdata.Groups {
		if r.Tag != "" {
			ref2grouponly[r.Id] = r.Tag
		}
	}
	for _, r := range dbdata.Classes {
		ref2fet[r.Id] = r.Tag
		ref2fet[r.ClassGroup] = r.Tag
		// Handle the groups
		for _, d := range r.Divisions {
			for _, g := range d.Groups {
				ref2fet[g] = r.Tag + CLASS_GROUP_SEP + ref2grouponly[g]
			}
		}
	}

	//fmt.Printf("ref2fet: %v\n", ref2fet)

	fetinfo := fetInfo{
		db:            dbdata,
		ref2fet:       ref2fet,
		ref2grouponly: ref2grouponly,
		fetdata: fet{
			Version:          fet_version,
			Mode:             "Official",
			Institution_Name: dbdata.Info.Institution,
			Comments:         getString(dbdata.Info.Reference),
			Time_Constraints_List: timeConstraints{
				ConstraintBasicCompulsoryTime: basicTimeConstraint{
					Weight_Percentage: 100, Active: true},
			},
			Space_Constraints_List: spaceConstraints{
				ConstraintBasicCompulsorySpace: basicSpaceConstraint{
					Weight_Percentage: 100, Active: true},
			},
		},
		ONLY_FIXED:              true,
		WITHOUT_ROOM_PLACEMENTS: true,
		fetVirtualRooms:         map[string]string{},
		fetVirtualRoomN:         map[string]int{},
		daysBetween:             map[Ref][]*base.DaysBetween{},
	}

	getDays(&fetinfo)
	getHours(&fetinfo)
	getTeachers(&fetinfo)
	getSubjects(&fetinfo)
	getRooms(&fetinfo)
	fmt.Println("=====================================")
	gatherCourseInfo(&fetinfo)

	//readCourseIndexes(&fetinfo)
	makeAtomicGroups(&fetinfo)
	//fmt.Println("\n +++++++++++++++++++++++++++")
	//printAtomicGroups(&fetinfo)
	getClasses(&fetinfo)
	lessonIdMap := getActivities(&fetinfo)

	addTeacherConstraints(&fetinfo)
	addClassConstraints(&fetinfo)

	getExtraConstraints(&fetinfo)
	addDifferentDaysConstraints(&fetinfo) // after getExtraConstraints!

	// Convert lessonIdMap to string
	idmlines := []string{}
	for _, idm := range lessonIdMap {
		idmlines = append(idmlines,
			strconv.Itoa(idm.activityId)+":"+string(idm.baseId))
	}
	lidmap := strings.Join(idmlines, "\n")

	return xml.Header + makeXML(fetinfo.fetdata, 0), lidmap
}

func getString(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		b, _ := json.Marshal(val)
		s = string(b)
	}
	return s
}
