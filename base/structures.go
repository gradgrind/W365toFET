package base

import (
	"slices"
)

// The structures used for the "database"
//TODO: Currently dealing only with the elements needed for the timetable

//+++++ These structures may be used by other database representations.

type Ref string // Element reference

type TimeSlot struct {
	Day  int
	Hour int
}

type Division struct {
	Name   string
	Groups []*Group
}

//------

type Info struct {
	Institution        string
	FirstAfternoonHour int
	MiddayBreak        []int
	Reference          string
}

type Day struct {
	Id   Ref
	Name string
	Tag  string
}

type Hour struct {
	Id    Ref
	Name  string
	Tag   string
	Start string
	End   string
}

type Element interface {
	GetReferrers() []Element
}

type Teacher struct {
	Referrers        []Element
	Id               Ref
	Name             string
	Tag              string
	Firstname        string
	NotAvailable     []TimeSlot
	MinLessonsPerDay int // default = -1
	MaxLessonsPerDay int // default = -1
	MaxDays          int // default = -1
	MaxGapsPerDay    int // default = -1
	MaxGapsPerWeek   int // default = -1
	MaxAfternoons    int // default = -1
	LunchBreak       bool
}

func (e *Teacher) GetReferrers() []Element {
	return e.Referrers
}

type Subject struct {
	Referrers []Element
	Id        Ref
	Name      string
	Tag       string
}

func (e *Subject) GetReferrers() []Element {
	return e.Referrers
}

type Room struct {
	Referrers    []Element
	Id           Ref
	Name         string
	Tag          string
	NotAvailable []TimeSlot
}

func (e *Room) GetReferrers() []Element {
	return e.Referrers
}

func (r *Room) IsReal() bool {
	return true
}

type RoomGroup struct {
	Referrers []Element
	Id        Ref
	Name      string
	Tag       string
	Rooms     []*Room
}

func (e *RoomGroup) GetReferrers() []Element {
	return e.Referrers
}

func (r *RoomGroup) IsReal() bool {
	return false
}

type RoomChoiceGroup struct {
	Referrers []Element
	Id        Ref
	Name      string
	Tag       string
	Rooms     []*Room
}

func (e *RoomChoiceGroup) GetReferrers() []Element {
	return e.Referrers
}

func (r *RoomChoiceGroup) IsReal() bool {
	return false
}

type Class struct {
	Referrers        []Element
	Id               Ref
	Name             string
	Tag              string
	Year             int
	Letter           string
	NotAvailable     []TimeSlot
	Divisions        []*Division
	MinLessonsPerDay int // default = -1
	MaxLessonsPerDay int // default = -1
	MaxGapsPerDay    int // default = -1
	MaxGapsPerWeek   int // default = -1
	MaxAfternoons    int // default = -1
	LunchBreak       bool
	ForceFirstHour   bool
	ClassGroup       *Group
}

func (e *Class) GetReferrers() []Element {
	return e.Referrers
}

type Group struct {
	Referrers []Element // not exported
	Id        Ref
	Tag       string
	Class     *Class    // not exported
	Division  *Division // not exported
}

func (e *Group) GetReferrers() []Element {
	return e.Referrers
}

type Course struct {
	Referrers []Element
	Id        Ref
	Subject   *Subject
	Groups    []*Group
	Teachers  []*Teacher
	Room      GeneralRoom // Room, RoomGroup or RoomChoiceGroup Element
	Lessons   []*Lesson
}

func (e *Course) GetReferrers() []Element {
	return e.Referrers
}

func (c *Course) IsSuperCourse() bool {
	return false
}

func (c *Course) AddLesson(l *Lesson) {
	c.Lessons = append(c.Lessons, l)
	l.Referrers = append(l.Referrers, c)
}

type SuperCourse struct {
	Referrers  []Element
	Id         Ref
	Subject    *Subject
	SubCourses []*SubCourse
	Lessons    []*Lesson
}

func (e *SuperCourse) GetReferrers() []Element {
	return e.Referrers
}

func (c *SuperCourse) IsSuperCourse() bool {
	return true
}

func (c *SuperCourse) AddLesson(l *Lesson) {
	c.Lessons = append(c.Lessons, l)
	l.Referrers = append(l.Referrers, c)
}

type SubCourse struct {
	Referrers   []Element
	Id          Ref
	SuperCourse *SuperCourse
	Subject     *Subject
	Groups      []*Group
	Teachers    []*Teacher
	Room        GeneralRoom // Room, RoomGroup or RoomChoiceGroup Element
}

func (e *SubCourse) GetReferrers() []Element {
	return e.Referrers
}

type GeneralRoom interface {
	IsReal() bool
}

type Lesson struct {
	Referrers []Element
	Id        Ref
	Course    LessonCourse // *Course or *SuperCourse Elements
	Duration  int
	Day       int
	Hour      int
	Fixed     bool
	Rooms     []*Room
}

func (e *Lesson) GetReferrers() []Element {
	return e.Referrers
}

type LessonCourse interface {
	IsSuperCourse() bool
	AddLesson(*Lesson)
}

type DbTopLevel struct {
	Info             Info
	Days             []*Day
	Hours            []*Hour
	Teachers         []*Teacher
	Subjects         []*Subject
	Rooms            []*Room
	RoomGroups       []*RoomGroup
	RoomChoiceGroups []*RoomChoiceGroup
	Classes          []*Class
	Groups           []*Group
	Courses          []*Course
	SuperCourses     []*SuperCourse
	SubCourses       []*SubCourse
	Lessons          []*Lesson
	//TODO:
	Constraints map[string]any
}

/* TODO
func (db *DbTopLevel) NewId() Ref {
	return Ref(fmt.Sprintf("#%d", db.MaxId+1))
}
*/

func (db *DbTopLevel) CheckDb() {
	// Checks
	//TODO: Should these really be fatal?
	if len(db.Days) == 0 {
		Error.Fatalln("No Days")
	}
	if len(db.Hours) == 0 {
		Error.Fatalln("No Hours")
	}
	if len(db.Teachers) == 0 {
		Error.Fatalln("No Teachers")
	}
	if len(db.Subjects) == 0 {
		Error.Fatalln("No Subjects")
	}
	if len(db.Rooms) == 0 {
		Error.Fatalln("No Rooms")
	}
	if len(db.Classes) == 0 {
		Error.Fatalln("No Classes")
	}

	// Initializations
	if db.Info.MiddayBreak == nil {
		db.Info.MiddayBreak = []int{}
	} else {
		// Sort and check contiguity.
		slices.Sort(db.Info.MiddayBreak)
		mb := db.Info.MiddayBreak
		if mb[len(mb)-1]-mb[0] >= len(mb) {
			Error.Fatalln("MiddayBreak hours not contiguous")
		}
	}

	// Collect the SubCourses for each SuperCourse
	for i := 0; i < len(db.SubCourses); i++ {
		n := db.SubCourses[i]
		supref := n.SuperCourse
		supref.SubCourses = append(supref.SubCourses, n)
		n.Referrers = append(n.Referrers, supref)
	}

	// Collect the Lessons for each Course and SuperCourse
	for i := 0; i < len(db.Lessons); i++ {
		n := db.Lessons[i]
		n.Course.AddLesson(n)
	}

	if db.Constraints == nil {
		db.Constraints = make(map[string]any)
	}

	// Expand Group information
	for ic := 0; ic < len(db.Classes); ic++ {
		c := db.Classes[ic]
		c.ClassGroup.Class = c // Tag and Division are empty.
		for id := 0; id < len(c.Divisions); id++ {
			d := c.Divisions[id]
			for _, g := range d.Groups {
				g.Class = c
				g.Division = d
			}
		}
	}
	// Check that all groups belong to a class
	for _, g := range db.Groups {
		if g.Class == nil {
			Error.Printf("Group not in Class: %s\n", g.Id)
		}
	}
}

// Interface for Course and SubCourse elements
type CourseInterface interface {
	GetId() Ref
	GetGroups() []*Group
	GetTeachers() []*Teacher
	GetSubject() *Subject
	GetRoom() GeneralRoom
}

func (c *Course) GetId() Ref                 { return c.Id }
func (c *SubCourse) GetId() Ref              { return c.Id }
func (c *Course) GetGroups() []*Group        { return c.Groups }
func (c *SubCourse) GetGroups() []*Group     { return c.Groups }
func (c *Course) GetTeachers() []*Teacher    { return c.Teachers }
func (c *SubCourse) GetTeachers() []*Teacher { return c.Teachers }
func (c *Course) GetSubject() *Subject       { return c.Subject }
func (c *SubCourse) GetSubject() *Subject    { return c.Subject }
func (c *Course) GetRoom() GeneralRoom       { return c.Room }
func (c *SubCourse) GetRoom() GeneralRoom    { return c.Room }
