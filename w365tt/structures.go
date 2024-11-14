package w365tt

import (
	"W365toFET/base"
	"W365toFET/logging"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
	Subject    Ref         `json:"subject"`
	EpochPlan  Ref         `json:"epochPlan"`
	SubCourses []SubCourse `json:"subCourses"`
}

type SubCourse struct {
	Id0            Ref    `json:"id"`
	Id             Ref    `json:"-"`
	Type           string `json:"type"`
	SuperCourses   []Ref  `json:"superCourses"`
	Subjects       []Ref  `json:"subjects"`
	Subject        Ref    `json:"subject"`
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

type DbTopLevel struct {
	Info         Info           `json:"w365TT"`
	Days         []*Day         `json:"days"`
	Hours        []*Hour        `json:"hours"`
	Teachers     []*Teacher     `json:"teachers"`
	Subjects     []*Subject     `json:"subjects"`
	Rooms        []*Room        `json:"rooms"`
	RoomGroups   []*RoomGroup   `json:"roomGroups"`
	Classes      []*Class       `json:"classes"`
	Groups       []*Group       `json:"groups"`
	Courses      []*Course      `json:"courses"`
	SuperCourses []*SuperCourse `json:"superCourses"`
	Lessons      []*Lesson      `json:"lessons"`
	EpochPlans   []*EpochPlan   `json:"epochPlans"`
	Constraints  map[string]any `json:"constraints"`

	// These fields do not belong in the JSON object.
	RealRooms  map[Ref]*base.Room    `json:"-"`
	SubjectMap map[Ref]*base.Subject `json:"-"`

	//??
	//Elements        map[Ref]any       `json:"-"`
	MaxId           int               `json:"-"` // for "indexed" Ids only
	SubjectTags     map[string]Ref    `json:"-"`
	SubjectNames    map[string]string `json:"-"`
	RoomTags        map[string]Ref    `json:"-"`
	RoomChoiceNames map[string]Ref    `json:"-"`
}

// TODO: At present I am not maintaining  db.MaxId ...
func (db *DbTopLevel) NewId() Ref {
	return Ref(fmt.Sprintf("#%d", db.MaxId+1))
}

func (db *DbTopLevel) AddElement(ref Ref, element any) {
	_, nok := db.Elements[ref]
	if nok {
		logging.Error.Printf("Element Id defined more than once:\n  %s\n", ref)
		return
	}
	db.Elements[ref] = element
	// Special handling if it is an "indexed" Id.
	if strings.HasPrefix(string(ref), "#") {
		s := strings.TrimPrefix(string(ref), "#")
		i, err := strconv.Atoi(s)
		if err == nil {
			if i > db.MaxId {
				db.MaxId = i
			}
		}
	}
}

func (db *DbTopLevel) checkDb() {
	if len(db.Days) == 0 {
		logging.Error.Fatalln("No Days")
	}
	if len(db.Hours) == 0 {
		logging.Error.Fatalln("No Hours")
	}
	if len(db.Teachers) == 0 {
		logging.Error.Fatalln("No Teachers")
	}
	if len(db.Subjects) == 0 {
		logging.Error.Fatalln("No Subjects")
	}
	if len(db.Rooms) == 0 {
		logging.Error.Fatalln("No Rooms")
	}
	if len(db.Classes) == 0 {
		logging.Error.Fatalln("No Classes")
	}
}

// Block all afternoons if nAfternnons == 0.
func (dbp *DbTopLevel) handleZeroAfternoons(
	notAvailable []TimeSlot,
	nAfternoons int,
) []base.TimeSlot {
	// Make a bool array and fill this in two passes, then remake list
	namap := make([][]bool, len(dbp.Days))
	nhours := len(dbp.Hours)
	// In the first pass, conditionally blocak afternoons
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

// Interface for Course and SubCourse elements
type CourseInterface interface {
	GetId() Ref
	GetGroups() []Ref
	GetTeachers() []Ref
	GetSubject() Ref
	getSubjects() []Ref       // not available externally
	getPreferredRooms() []Ref // not available externally
	GetRoom() Ref
	setSubject(Ref)
	setSubjects([]Ref)
	setPreferredRooms([]Ref)
	setRoom(Ref)
}

func (c *Course) GetId() Ref                    { return c.Id }
func (c *SubCourse) GetId() Ref                 { return c.Id }
func (c *Course) GetGroups() []Ref              { return c.Groups }
func (c *SubCourse) GetGroups() []Ref           { return c.Groups }
func (c *Course) GetTeachers() []Ref            { return c.Teachers }
func (c *SubCourse) GetTeachers() []Ref         { return c.Teachers }
func (c *Course) GetSubject() Ref               { return c.Subject }
func (c *SubCourse) GetSubject() Ref            { return c.Subject }
func (c *Course) getSubjects() []Ref            { return c.Subjects }
func (c *SubCourse) getSubjects() []Ref         { return c.Subjects }
func (c *Course) getPreferredRooms() []Ref      { return c.PreferredRooms }
func (c *SubCourse) getPreferredRooms() []Ref   { return c.PreferredRooms }
func (c *Course) GetRoom() Ref                  { return c.Room }
func (c *SubCourse) GetRoom() Ref               { return c.Room }
func (c *Course) setSubject(r Ref)              { c.Subject = r }
func (c *SubCourse) setSubject(r Ref)           { c.Subject = r }
func (c *Course) setSubjects(rr []Ref)          { c.Subjects = rr }
func (c *SubCourse) setSubjects(rr []Ref)       { c.Subjects = rr }
func (c *Course) setPreferredRooms(rr []Ref)    { c.PreferredRooms = rr }
func (c *SubCourse) setPreferredRooms(rr []Ref) { c.PreferredRooms = rr }
func (c *Course) setRoom(r Ref)                 { c.Room = r }
func (c *SubCourse) setRoom(r Ref)              { c.Room = r }
