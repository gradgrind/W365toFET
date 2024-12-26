// Package fet handles interaction with the fet timetabling program.
package fet

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/xml"
	"math"
	"strconv"
	"strings"
)

type Ref = base.Ref

const CLASS_GROUP_SEP = "."
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

const fet_version = "6.28.2"

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

func weight2fet(w int) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	wf := float64(w)
	n := wf + math.Pow(2, wf/12)
	wfet := 100.0 - 100.0/n
	return strconv.FormatFloat(wfet, 'f', 3, 64)
}

type idMap struct {
	activityId int
	baseId     string
}

type fetInfo struct {
	ttinfo        *ttbase.TtInfo
	ref2grouponly map[Ref]string
	fetdata       fet

	fetVirtualRooms map[string]string // cache for FET virtual rooms,
	// "hash" -> FET-virtual-room tag
	fetVirtualRoomN map[string]int // FET-virtual-room tag -> number of room sets
}

type timeConstraints struct {
	XMLName xml.Name `xml:"Time_Constraints_List"`
	//
	ConstraintBasicCompulsoryTime          basicTimeConstraint
	ConstraintStudentsSetNotAvailableTimes []studentsNotAvailable
	ConstraintTeacherNotAvailableTimes     []teacherNotAvailable

	ConstraintActivityPreferredStartingTime    []startingTime
	ConstraintActivityPreferredTimeSlots       []activityPreferredTimes
	ConstraintActivitiesPreferredTimeSlots     []preferredSlots
	ConstraintActivitiesPreferredStartingTimes []preferredStarts
	ConstraintMinDaysBetweenActivities         []minDaysBetweenActivities
	ConstraintActivityEndsStudentsDay          []lessonEndsDay
	ConstraintActivitiesSameStartingTime       []sameStartingTime

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
	ConstraintRoomNotAvailableTimes  []roomNotAvailable
}

type basicSpaceConstraint struct {
	XMLName           xml.Name `xml:"ConstraintBasicCompulsorySpace"`
	Weight_Percentage int
	Active            bool
}

func MakeFetFile(ttinfo *ttbase.TtInfo) (string, string) {
	dbdata := ttinfo.Db

	// Build ref-index -> fet-key mapping
	ref2grouponly := map[Ref]string{}
	for _, r := range dbdata.Groups {
		if r.Tag != "" {
			ref2grouponly[r.Id] = r.Tag
		}
	}

	//fmt.Printf("ref2fet: %v\n", ref2fet)

	fetinfo := fetInfo{
		ttinfo:        ttinfo,
		ref2grouponly: ref2grouponly,
		fetdata: fet{
			Version:          fet_version,
			Mode:             "Official",
			Institution_Name: dbdata.Info.Institution,
			Comments:         dbdata.Info.Reference,
			Time_Constraints_List: timeConstraints{
				ConstraintBasicCompulsoryTime: basicTimeConstraint{
					Weight_Percentage: 100, Active: true},
			},
			Space_Constraints_List: spaceConstraints{
				ConstraintBasicCompulsorySpace: basicSpaceConstraint{
					Weight_Percentage: 100, Active: true},
			},
		},
		fetVirtualRooms: map[string]string{},
		fetVirtualRoomN: map[string]int{},

		//ONLY_FIXED:              true,
		//WITHOUT_ROOM_PLACEMENTS: true,
		//daysBetween:             map[Ref][]*base.DaysBetween{},
	}

	getDays(&fetinfo)
	getHours(&fetinfo)
	getTeachers(&fetinfo)
	getSubjects(&fetinfo)
	getRooms(&fetinfo)

	//TODO--
	//fmt.Println("=====================================")
	//gatherCourseInfo(&fetinfo)

	//readCourseIndexes(&fetinfo)
	//makeAtomicGroups(&fetinfo)
	//fmt.Println("\n +++++++++++++++++++++++++++")
	//printAtomicGroups(&fetinfo)

	getClasses(&fetinfo)
	lessonIdMap := getActivities(&fetinfo)

	addTeacherConstraints(&fetinfo)
	addClassConstraints(&fetinfo)
	getExtraConstraints(&fetinfo)

	// Convert lessonIdMap to string
	idmlines := []string{}
	for _, idm := range lessonIdMap {
		idmlines = append(idmlines,
			strconv.Itoa(idm.activityId)+":"+string(idm.baseId))
	}
	lidmap := strings.Join(idmlines, "\n")

	return xml.Header + makeXML(fetinfo.fetdata, 0), lidmap
}

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
