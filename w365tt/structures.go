package w365tt

import (
	"W365toFET/base"
	"encoding/json"
)

// The structures used for the "database", adapted to read from W365
//TODO: Currently dealing only with the elements needed for the timetable

type Ref = base.Ref // Element reference

type Info struct {
	Institution        string `json:"schoolName"`
	FirstAfternoonHour int    `json:"firstAfternoonHour"`
	MiddayBreak        []int  `json:"middayBreak"`
	Reference          string `json:"scenario"`
}

type Day struct {
	Id   Ref    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Tag  string `json:"shortcut"`
}

type Hour struct {
	Id    Ref    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Tag   string `json:"shortcut"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type TimeSlot struct {
	Day  int `json:"day"`
	Hour int `json:"hour"`
}

type Teacher struct {
	Id               Ref        `json:"id"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	Tag              string     `json:"shortcut"`
	Firstname        string     `json:"firstname"`
	NotAvailable     []TimeSlot `json:"absences"`
	MinLessonsPerDay int        `json:"minLessonsPerDay"`
	MaxLessonsPerDay int        `json:"maxLessonsPerDay"`
	MaxDays          int        `json:"maxDays"`
	MaxGapsPerDay    int        `json:"maxGapsPerDay"`
	MaxGapsPerWeek   int        `json:"maxGapsPerWeek"`
	MaxAfternoons    int        `json:"maxAfternoons"`
	LunchBreak       bool       `json:"lunchBreak"`
}

func (t *Teacher) UnmarshalJSON(data []byte) error {
	// Customize defaults for Teacher
	t.MinLessonsPerDay = -1
	t.MaxLessonsPerDay = -1
	t.MaxDays = -1
	t.MaxGapsPerDay = -1
	t.MaxGapsPerWeek = -1
	t.MaxAfternoons = -1

	type tempT Teacher
	return json.Unmarshal(data, (*tempT)(t))
}

type Subject struct {
	Id   Ref    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Tag  string `json:"shortcut"`
}

type Room struct {
	Id           Ref        `json:"id"`
	Type         string     `json:"type"`
	Name         string     `json:"name"`
	Tag          string     `json:"shortcut"`
	NotAvailable []TimeSlot `json:"absences"`
}

type RoomGroup struct {
	Id    Ref    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Tag   string `json:"shortcut"`
	Rooms []Ref  `json:"rooms"`
}

type RoomChoiceGroup struct {
	Id    Ref    `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Tag   string `json:"shortcut"`
	Rooms []Ref  `json:"rooms"`
}

type Class struct {
	Id               Ref        `json:"id"`
	Type             string     `json:"type"`
	Name             string     `json:"name"`
	Tag              string     `json:"shortcut"`
	Year             int        `json:"level"`
	Letter           string     `json:"letter"`
	NotAvailable     []TimeSlot `json:"absences"`
	Divisions        []Division `json:"divisions"`
	MinLessonsPerDay int        `json:"minLessonsPerDay"`
	MaxLessonsPerDay int        `json:"maxLessonsPerDay"`
	MaxGapsPerDay    int        `json:"maxGapsPerDay"`
	MaxGapsPerWeek   int        `json:"maxGapsPerWeek"`
	MaxAfternoons    int        `json:"maxAfternoons"`
	LunchBreak       bool       `json:"lunchBreak"`
	ForceFirstHour   bool       `json:"forceFirstHour"`
}

func (t *Class) UnmarshalJSON(data []byte) error {
	// Customize defaults for Teacher
	t.MinLessonsPerDay = -1
	t.MaxLessonsPerDay = -1
	t.MaxGapsPerDay = -1
	t.MaxGapsPerWeek = -1
	t.MaxAfternoons = -1

	type tempT Class
	return json.Unmarshal(data, (*tempT)(t))
}

type Group struct {
	Id   Ref    `json:"id"`
	Type string `json:"type"`
	Tag  string `json:"shortcut"`
}

type Division struct {
	Id     Ref    `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Groups []Ref  `json:"groups"`
}

type Course struct {
	Id             Ref    `json:"id"`
	Type           string `json:"type"`
	Subjects       []Ref  `json:"subjects"`
	Groups         []Ref  `json:"groups"`
	Teachers       []Ref  `json:"teachers"`
	PreferredRooms []Ref  `json:"preferredRooms,omitempty"`
}

type SuperCourse struct {
	Id         Ref         `json:"id"`
	Type       string      `json:"type"`
	EpochPlan  Ref         `json:"epochPlan"`
	SubCourses []SubCourse `json:"subCourses"`
}

type SubCourse struct {
	Id             Ref    `json:"id"`
	Type           string `json:"type"`
	Subjects       []Ref  `json:"subjects"`
	Groups         []Ref  `json:"groups"`
	Teachers       []Ref  `json:"teachers"`
	PreferredRooms []Ref  `json:"preferredRooms"`
}

type Lesson struct {
	Id       Ref      `json:"id"`
	Type     string   `json:"type"`
	Course   Ref      `json:"course"` // Course or SuperCourse Elements
	Duration int      `json:"duration"`
	Day      int      `json:"day"`
	Hour     int      `json:"hour"`
	Fixed    bool     `json:"fixed"`
	Rooms    []Ref    `json:"localRooms"` // only Room Elements
	Flags    []string `json:"flags"`
}

type EpochPlan struct {
	Id   Ref    `json:"id"`
	Type string `json:"type"`
	Tag  string `json:"shortcut"`
	Name string `json:"name"`
}

type PrintOptions struct {
	Title              string           `json:"title"`
	Subtitle           string           `json:"subtitle"`
	PageHeadingClass   string           `json:"pageHeadingClass"`
	PageHeadingTeacher string           `json:"pageHeadingTeacher"`
	PageHeadingRoom    string           `json:"pageHeadingRoom"`
	WithTimes          bool             `json:"withTimes"`
	WithBreaks         bool             `json:"withBreaks"`
	BoxTextClass       []map[string]any `json:"boxTextClass"`
	BoxTextTeacher     []map[string]any `json:"boxTextTeacher"`
	BoxTextRoom        []map[string]any `json:"boxTextRoom"`
	PrintTables        []string         `json:"printTables"`
	PrintId            base.Ref         `json:"printId"`
}

type DbTopLevel struct {
	Info         Info             `json:"w365TT"`
	PrintOptions PrintOptions     `json:"printOptions"`
	Days         []*Day           `json:"days"`
	Hours        []*Hour          `json:"hours"`
	Teachers     []*Teacher       `json:"teachers"`
	Subjects     []*Subject       `json:"subjects"`
	Rooms        []*Room          `json:"rooms"`
	RoomGroups   []*RoomGroup     `json:"roomGroups"`
	Classes      []*Class         `json:"classes"`
	Groups       []*Group         `json:"groups"`
	Courses      []*Course        `json:"courses"`
	SuperCourses []*SuperCourse   `json:"superCourses"`
	Lessons      []*Lesson        `json:"lessons"`
	EpochPlans   []*EpochPlan     `json:"epochPlans"`
	Constraints  []map[string]any `json:"constraints"`

	// These fields do not belong in the JSON object.
	RealRooms       map[Ref]*base.Room      `json:"-"`
	RoomGroupMap    map[Ref]*base.RoomGroup `json:"-"`
	SubjectMap      map[Ref]*base.Subject   `json:"-"`
	GroupRefMap     map[Ref]base.Ref        `json:"-"`
	TeacherMap      map[Ref]bool            `json:"-"`
	CourseMap       map[Ref]bool            `json:"-"`
	SubjectTags     map[string]Ref          `json:"-"`
	RoomTags        map[string]Ref          `json:"-"`
	RoomChoiceNames map[string]Ref          `json:"-"`
}

// Block all afternoons if nAfternnons == 0.
func (dbp *DbTopLevel) handleZeroAfternoons(
	notAvailable []TimeSlot,
	nAfternoons int,
) []base.TimeSlot {
	// Make a bool array and fill this in two passes, then remake list
	namap := make([][]bool, len(dbp.Days))
	nhours := len(dbp.Hours)
	// In the first pass, conditionally block afternoons
	for i := range namap {
		namap[i] = make([]bool, nhours)
		if nAfternoons == 0 {
			for h := dbp.Info.FirstAfternoonHour; h < nhours; h++ {
				namap[i][h] = true
			}
		}
	}
	// In the second pass, include existing blocked hours.
	for _, ts := range notAvailable {
		namap[ts.Day][ts.Hour] = true
	}
	// Build a new base.TimeSlot list
	na := []base.TimeSlot{}
	for d, naday := range namap {
		for h, nahour := range naday {
			if nahour {
				na = append(na, base.TimeSlot{Day: d, Hour: h})
			}
		}
	}
	return na
}
