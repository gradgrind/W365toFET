package w365tt

import (
	"W365toFET/logging"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// The structures used for the "database", adapted to read from W365
//TODO: Currently dealing only with the elements needed for the timetable

type Ref string // Element reference

type Info struct {
	Institution        string `json:"schoolName"`
	FirstAfternoonHour int    `json:"firstAfternoonHour"`
	MiddayBreak        []int  `json:"middayBreak"`
	Reference          string `json:"scenario"`
}

type Day struct {
	Id   Ref    `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"shortcut"`
}

type Hour struct {
	Id    Ref    `json:"id"`
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
	Id               Ref         `json:"id"`
	Name             string      `json:"name"`
	Tag              string      `json:"shortcut"`
	Firstname        string      `json:"firstname"`
	NotAvailable     []TimeSlot  `json:"absences"`
	MinLessonsPerDay interface{} `json:"minLessonsPerDay"`
	MaxLessonsPerDay interface{} `json:"maxLessonsPerDay"`
	MaxDays          interface{} `json:"maxDays"`
	MaxGapsPerDay    interface{} `json:"maxGapsPerDay"`
	MaxGapsPerWeek   interface{} `json:"maxGapsPerWeek"`
	MaxAfternoons    interface{} `json:"maxAfternoons"`
	LunchBreak       bool        `json:"lunchBreak"`
}

type Subject struct {
	Id   Ref    `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"shortcut"`
}

type Room struct {
	Id           Ref        `json:"id"`
	Name         string     `json:"name"`
	Tag          string     `json:"shortcut"`
	NotAvailable []TimeSlot `json:"absences"`
}

type RoomGroup struct {
	Id    Ref    `json:"id"`
	Name  string `json:"name"`
	Tag   string `json:"shortcut"`
	Rooms []Ref  `json:"rooms"`
}

type RoomChoiceGroup struct {
	Id    Ref    `json:"id"`
	Name  string `json:"name"`
	Tag   string `json:"shortcut"`
	Rooms []Ref  `json:"rooms"`
}

type Class struct {
	Id               Ref         `json:"id"`
	Name             string      `json:"name"`
	Tag              string      `json:"shortcut"`
	Year             int         `json:"level"`
	Letter           string      `json:"letter"`
	NotAvailable     []TimeSlot  `json:"absences"`
	Divisions        []Division  `json:"divisions"`
	MinLessonsPerDay interface{} `json:"minLessonsPerDay"`
	MaxLessonsPerDay interface{} `json:"maxLessonsPerDay"`
	MaxGapsPerDay    interface{} `json:"maxGapsPerDay"`
	MaxGapsPerWeek   interface{} `json:"maxGapsPerWeek"`
	MaxAfternoons    interface{} `json:"maxAfternoons"`
	LunchBreak       bool        `json:"lunchBreak"`
	ForceFirstHour   bool        `json:"forceFirstHour"`
}

type Group struct {
	Id  Ref    `json:"id"`
	Tag string `json:"shortcut"`
}

type Division struct {
	Id     Ref    `json:"id"`
	Name   string `json:"name"`
	Groups []Ref  `json:"groups"`
}

type Course struct {
	Id             Ref   `json:"id"`
	Subjects       []Ref `json:"subjects,omitempty"`
	Subject        Ref   `json:"subject"`
	Groups         []Ref `json:"groups"`
	Teachers       []Ref `json:"teachers"`
	PreferredRooms []Ref `json:"preferredRooms,omitempty"`
	// Not in W365:
	Room Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type SuperCourse struct {
	Id      Ref `json:"id"`
	Subject Ref `json:"subject"`
}

type SubCourse struct {
	Id0            Ref `json:"id"`
	Id             Ref
	SuperCourse    Ref   `json:"superCourse"`
	Subjects       []Ref `json:"subjects,omitempty"`
	Subject        Ref   `json:"subject"`
	Groups         []Ref `json:"groups"`
	Teachers       []Ref `json:"teachers"`
	PreferredRooms []Ref `json:"preferredRooms,omitempty"`
	// Not in W365:
	Room Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type Lesson struct {
	Id       Ref   `json:"id"`
	Course   Ref   `json:"course"` // Course or SuperCourse Elements
	Duration int   `json:"duration"`
	Day      int   `json:"day"`
	Hour     int   `json:"hour"`
	Fixed    bool  `json:"fixed"`
	Rooms    []Ref `json:"localRooms"` // only Room Elements
}

type DbTopLevel struct {
	Info             Info                   `json:"w365TT"`
	Days             []Day                  `json:"days"`
	Hours            []Hour                 `json:"hours"`
	Teachers         []Teacher              `json:"teachers"`
	Subjects         []Subject              `json:"subjects"`
	Rooms            []Room                 `json:"rooms"`
	RoomGroups       []RoomGroup            `json:"roomGroups"`
	RoomChoiceGroups []RoomChoiceGroup      `json:"roomChoiceGroups"`
	Classes          []Class                `json:"classes"`
	Groups           []Group                `json:"groups"`
	Courses          []Course               `json:"courses"`
	SuperCourses     []SuperCourse          `json:"superCourses"`
	SubCourses       []SubCourse            `json:"subCourses"`
	Lessons          []Lesson               `json:"lessons"`
	Constraints      map[string]interface{} `json:"constraints"`

	// These fields do not belong in the JSON object.
	Elements        map[Ref]interface{} `json:"-"`
	MaxId           int                 `json:"-"` // for "indexed" Ids only
	SubjectTags     map[string]Ref      `json:"-"`
	SubjectNames    map[string]string   `json:"-"`
	RoomTags        map[string]Ref      `json:"-"`
	RoomChoiceNames map[string]Ref      `json:"-"`
}

func (db *DbTopLevel) NewId() Ref {
	return Ref(fmt.Sprintf("#%d", db.MaxId+1))
}

func (db *DbTopLevel) AddElement(ref Ref, element interface{}) {
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
	// Initializations
	if db.Info.MiddayBreak == nil {
		db.Info.MiddayBreak = []int{}
	} else {
		// Sort and check contiguity.
		slices.Sort(db.Info.MiddayBreak)
		mb := db.Info.MiddayBreak
		if mb[len(mb)-1]-mb[0] >= len(mb) {
			logging.Error.Fatalln("MiddayBreak hours not contiguous")
		}

	}
	db.SubjectTags = map[string]Ref{}
	db.SubjectNames = map[string]string{}
	db.RoomTags = map[string]Ref{}
	db.RoomChoiceNames = map[string]Ref{}
	// Initialize the Ref -> Element mapping
	db.Elements = make(map[Ref]interface{})
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
	for i, n := range db.Days {
		db.AddElement(n.Id, &db.Days[i])
	}
	for i, n := range db.Hours {
		db.AddElement(n.Id, &db.Hours[i])
	}
	for i, n := range db.Teachers {
		db.AddElement(n.Id, &db.Teachers[i])
	}
	for i, n := range db.Subjects {
		db.AddElement(n.Id, &db.Subjects[i])
	}
	for i, n := range db.Rooms {
		db.AddElement(n.Id, &db.Rooms[i])
	}
	for i, n := range db.Classes {
		db.AddElement(n.Id, &db.Classes[i])
	}
	if db.RoomGroups == nil {
		db.RoomGroups = []RoomGroup{}
	} else {
		for i, n := range db.RoomGroups {
			db.AddElement(n.Id, &db.RoomGroups[i])
		}
	}
	if db.RoomChoiceGroups == nil {
		db.RoomChoiceGroups = []RoomChoiceGroup{}
	} else {
		for i, n := range db.RoomChoiceGroups {
			db.AddElement(n.Id, &db.RoomChoiceGroups[i])
		}
	}
	if db.Groups == nil {
		db.Groups = []Group{}
	} else {
		for i, n := range db.Groups {
			db.AddElement(n.Id, &db.Groups[i])
		}
	}
	if db.Courses == nil {
		db.Courses = []Course{}
	} else {
		for i, n := range db.Courses {
			db.AddElement(n.Id, &db.Courses[i])
		}
	}
	if db.SuperCourses == nil {
		db.SuperCourses = []SuperCourse{}
	} else {
		for i, n := range db.SuperCourses {
			db.AddElement(n.Id, &db.SuperCourses[i])
		}
	}
	if db.SubCourses == nil {
		db.SubCourses = []SubCourse{}
	} else {
		for i, n := range db.SubCourses {
			nid := "$$" + n.Id0
			db.SubCourses[i].Id = nid
			db.AddElement(nid, &db.SubCourses[i])
		}
	}
	if db.Lessons == nil {
		db.Lessons = []Lesson{}
	} else {
		for i, n := range db.Lessons {
			db.AddElement(n.Id, &db.Lessons[i])
		}
	}
	if db.Constraints == nil {
		db.Constraints = make(map[string]interface{})
	}
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
