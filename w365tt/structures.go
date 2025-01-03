package w365tt

import (
	"W365toFET/base"
	"encoding/json"
)

// The structures used for the "database", adapted to read from W365
//TODO: Currently dealing only with the elements needed for the timetable

type Ref = base.Ref // Element reference

type Info struct {
	Institution        string
	FirstAfternoonHour int
	MiddayBreak        []int
	Reference          string `json:"Scenario"`
}

type Day struct {
	Id   Ref
	Type string
	Name string
	Tag  string `json:"Shortcut"`
}

type Hour struct {
	Id    Ref
	Type  string
	Name  string
	Tag   string `json:"Shortcut"`
	Start string
	End   string
}

type TimeSlot struct {
	Day  int
	Hour int
}

type Teacher struct {
	Id               Ref
	Type             string
	Name             string
	Tag              string `json:"Shortcut"`
	Firstname        string
	NotAvailable     []TimeSlot `json:"Absences"`
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxDays          int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
	LunchBreak       bool
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
	Id   Ref
	Type string
	Name string
	Tag  string `json:"Shortcut"`
}

type Room struct {
	Id           Ref
	Type         string
	Name         string
	Tag          string     `json:"Shortcut"`
	NotAvailable []TimeSlot `json:"Absences"`
}

type RoomGroup struct {
	Id    Ref
	Type  string
	Name  string
	Tag   string `json:"Shortcut"`
	Rooms []Ref
}

type RoomChoiceGroup struct {
	Id    Ref
	Type  string
	Name  string
	Tag   string `json:"Shortcut"`
	Rooms []Ref
}

type Class struct {
	Id               Ref
	Type             string
	Name             string
	Tag              string `json:"Shortcut"`
	Year             int    `json:"Level"`
	Letter           string
	NotAvailable     []TimeSlot `json:"Absences"`
	Divisions        []Division
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
	LunchBreak       bool
	ForceFirstHour   bool
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
	Id   Ref
	Type string
	Tag  string `json:"Shortcut"`
}

type Division struct {
	Id     Ref
	Type   string
	Name   string
	Groups []Ref
}

type Course struct {
	Id             Ref
	Type           string
	Subjects       []Ref
	Groups         []Ref
	Teachers       []Ref
	PreferredRooms []Ref
}

type SuperCourse struct {
	Id         Ref
	Type       string
	EpochPlan  Ref
	SubCourses []SubCourse
}

type SubCourse struct {
	Id             Ref
	Type           string
	Subjects       []Ref
	Groups         []Ref
	Teachers       []Ref
	PreferredRooms []Ref
}

type Lesson struct {
	Id         Ref
	Type       string
	Course     Ref // Course or SuperCourse Elements
	Duration   int
	Day        int
	Hour       int
	Fixed      bool
	Rooms      []Ref `json:"LocalRooms"` // only Room Elements
	Flags      []string
	Background string
	Footnote   string
}

type EpochPlan struct {
	Id   Ref
	Type string
	Tag  string `json:"Shortcut"`
	Name string
}

type DbTopLevel struct {
	Info         Info `json:"W365TT"`
	PrintTables  []*base.PrintTable
	Days         []*Day
	Hours        []*Hour
	Teachers     []*Teacher
	Subjects     []*Subject
	Rooms        []*Room
	RoomGroups   []*RoomGroup
	Classes      []*Class
	Groups       []*Group
	Courses      []*Course
	SuperCourses []*SuperCourse
	Lessons      []*Lesson
	EpochPlans   []*EpochPlan
	Constraints  []map[string]any

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
