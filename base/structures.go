package base

// The structures used for the "database"
//TODO: Currently dealing only with the elements needed for the timetable

type Ref string // Element Id

type TimeSlot struct {
	Day  int
	Hour int
}

type Division struct {
	Name   string
	Groups []Ref
}

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

type Teacher struct {
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

type Subject struct {
	Id   Ref
	Name string
	Tag  string
}

type Room struct {
	Id           Ref
	Name         string
	Tag          string
	NotAvailable []TimeSlot
}

func (r *Room) IsReal() bool {
	return true
}

type RoomGroup struct {
	Id    Ref
	Name  string
	Tag   string
	Rooms []Ref
}

func (r *RoomGroup) IsReal() bool {
	return false
}

type RoomChoiceGroup struct {
	Id    Ref
	Name  string
	Tag   string
	Rooms []Ref
}

func (r *RoomChoiceGroup) IsReal() bool {
	return false
}

type Class struct {
	Id               Ref
	Name             string
	Tag              string
	Year             int
	Letter           string
	NotAvailable     []TimeSlot
	Divisions        []Division
	MinLessonsPerDay int // default = -1
	MaxLessonsPerDay int // default = -1
	MaxGapsPerDay    int // default = -1
	MaxGapsPerWeek   int // default = -1
	MaxAfternoons    int // default = -1
	LunchBreak       bool
	ForceFirstHour   bool
	ClassGroup       Ref
}

type Group struct {
	Id  Ref
	Tag string
	// These fields do not belong in the JSON object:
	Class Ref `json:"-"`
}

type Course struct {
	Id       Ref
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     Ref // Room, RoomGroup or RoomChoiceGroup Element
	// These fields do not belong in the JSON object:
	Lessons []Ref `json:"-"`
}

func (c *Course) IsSuperCourse() bool {
	return false
}

func (c *Course) AddLesson(lref Ref) {
	c.Lessons = append(c.Lessons, lref)
}

type SuperCourse struct {
	Id      Ref
	Subject Ref
	// These fields do not belong in the JSON object:
	SubCourses []Ref `json:"-"`
	Lessons    []Ref `json:"-"`
}

func (c *SuperCourse) IsSuperCourse() bool {
	return true
}

func (c *SuperCourse) AddLesson(lref Ref) {
	c.Lessons = append(c.Lessons, lref)
}

type SubCourse struct {
	Id          Ref
	SuperCourse Ref
	Subject     Ref
	Groups      []Ref
	Teachers    []Ref
	Room        Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type GeneralRoom interface {
	IsReal() bool
}

type Lesson struct {
	Id       Ref
	Course   Ref // Course or SuperCourse Elements
	Duration int
	Day      int
	Hour     int
	Fixed    bool
	Rooms    []Ref
}

type LessonCourse interface {
	IsSuperCourse() bool
	AddLesson(Ref)
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

	// These fields do not belong in the JSON object:
	Elements map[Ref]any `json:"-"`
}
