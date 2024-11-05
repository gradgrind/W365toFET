package w365tt

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
)

// The structures used for the "database", adapted to read from W365
//TODO: Currently dealing only with the elements needed for the timetable

type Ref string // Element reference

type Info struct {
	Institution        string `json:"SchoolName"`
	FirstAfternoonHour int
	MiddayBreak        []int
	Reference          string `json:"Scenario"`
}

type Day struct {
	Id   Ref
	Name string
	Tag  string `json:"Shortcut"`
}

type Hour struct {
	Id    Ref
	Name  string
	Tag   string `json:"Shortcut"`
	Start string
	End   string
	// These are for W365 only, optional, with default = False:
	FirstAfternoonHour bool `json:",omitempty"`
	MiddayBreak        bool `json:",omitempty"`
}

type TimeSlot struct {
	Day  int
	Hour int
}

type Teacher struct {
	Id               Ref
	Name             string
	Tag              string `json:"Shortcut"`
	Firstname        string
	NotAvailable     []TimeSlot `json:"Absences"`
	MinLessonsPerDay interface{}
	MaxLessonsPerDay interface{}
	MaxDays          interface{}
	MaxGapsPerDay    interface{}
	MaxGapsPerWeek   interface{}
	MaxAfternoons    interface{}
	LunchBreak       bool
}

type Subject struct {
	Id   Ref
	Name string
	Tag  string `json:"Shortcut"`
}

type Room struct {
	Id           Ref
	Name         string
	Tag          string     `json:"Shortcut"`
	NotAvailable []TimeSlot `json:"Absences"`
}

type RoomGroup struct {
	Id    Ref
	Name  string
	Tag   string `json:"Shortcut"`
	Rooms []Ref
}

type RoomChoiceGroup struct {
	Id    Ref
	Name  string
	Tag   string `json:"Shortcut"`
	Rooms []Ref
}

type Class struct {
	Id               Ref
	Name             string
	Tag              string `json:"Shortcut"`
	Year             int    `json:"Level"`
	Letter           string
	NotAvailable     []TimeSlot `json:"Absences"`
	Divisions        []Division
	MinLessonsPerDay interface{}
	MaxLessonsPerDay interface{}
	MaxGapsPerDay    interface{}
	MaxGapsPerWeek   interface{}
	MaxAfternoons    interface{}
	LunchBreak       bool
	ForceFirstHour   bool
}

type Group struct {
	Id  Ref
	Tag string `json:"Shortcut"`
}

type Division struct {
	Name   string
	Groups []Ref
}

/*
type Division struct {
	Id     Ref
	Name   string
	Groups []Ref
}
*/

type Course struct {
	Id             Ref
	Subjects       []Ref `json:",omitempty"`
	Subject        Ref
	Groups         []Ref
	Teachers       []Ref
	PreferredRooms []Ref `json:",omitempty"`
	// Not in W365:
	Room Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type SuperCourse struct {
	Id      Ref
	Subject Ref
}

type SubCourse struct {
	Id             Ref
	SuperCourse    Ref
	Subjects       []Ref `json:",omitempty"`
	Subject        Ref
	Groups         []Ref
	Teachers       []Ref
	PreferredRooms []Ref `json:",omitempty"`
	// Not in W365:
	Room Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type Lesson struct {
	Id       Ref
	Course   Ref // Course or SuperCourse Elements
	Duration int
	Day      int
	Hour     int
	Fixed    bool
	Rooms    []Ref `json:"LocalRooms"` // only Room Elements
}

type DbTopLevel struct {
	Info             Info `json:"W365TT"`
	Days             []Day
	Hours            []Hour
	Teachers         []Teacher
	Subjects         []Subject
	Rooms            []Room
	RoomGroups       []RoomGroup
	RoomChoiceGroups []RoomChoiceGroup
	Classes          []Class
	Groups           []Group
	Courses          []Course
	SuperCourses     []SuperCourse
	SubCourses       []SubCourse
	Lessons          []Lesson
	Constraints      map[string]interface{}

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
		log.Fatalf("*ERROR* Element Id defined more than once:\n  %s\n", ref)
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
			log.Fatalln("*ERROR* MiddayBreak hours not contiguous")
		}

	}
	db.SubjectTags = map[string]Ref{}
	db.SubjectNames = map[string]string{}
	db.RoomTags = map[string]Ref{}
	db.RoomChoiceNames = map[string]Ref{}
	// Initialize the Ref -> Element mapping
	db.Elements = make(map[Ref]interface{})
	if len(db.Days) == 0 {
		log.Fatalln("*ERROR* No Days")
	}
	if len(db.Hours) == 0 {
		log.Fatalln("*ERROR* No Hours")
	}
	if len(db.Teachers) == 0 {
		log.Fatalln("*ERROR* No Teachers")
	}
	if len(db.Subjects) == 0 {
		log.Fatalln("*ERROR* No Subjects")
	}
	if len(db.Rooms) == 0 {
		log.Fatalln("*ERROR* No Rooms")
	}
	if len(db.Classes) == 0 {
		log.Fatalln("*ERROR* No Classes")
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
			db.AddElement(n.Id, &db.SubCourses[i])
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
