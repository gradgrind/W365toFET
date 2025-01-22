// Package base provides data structures for management of a school's data,
// together with some supporting functions and methods.
//
// Initially designed around the data connected with timetabling, it is
// readily extendable. The root element is the [DbTopLevel] struct, which
// is sometimes referred to as the "database".
// No particular method of persistent storage is specified, the data
// structures can be assembled from and saved to any format for which a
// supporting package is defined.
// However, there is basic in-built support for reading from and saving to
// JSON.
package base

// A Ref is used to identify the constituent elements of the database.
type Ref string // Element Id

// A TimeSlot specifies a lesson time period within the school week. The
// school week is divided into days, which are divided into "hours" (lesson
// periods), which are usually not 60 minutes in length. Each day has the
// same number of lessons.
type TimeSlot struct {
	Day  int // index to [DbTopLevel.Days]
	Hour int // index to [DbTopLevel.Hours]
}

// A Division specifies a particular splitting of a school "class" (the
// students, not the lessons) into a number of groups (say, "A" and "B").
//
// In principle, a class may have any number of divisions, each of which
// may have any number of groups, though keeping them to a minimum is
// generally advisable.
//
// Group names must be unique within a class and groups from different
// divisions may not have lessons at the same time.
type Division struct {
	Name   string
	Groups []Ref
}

// An Info (of which there will only be one instance) collects general
// information which doesn't have its own structure.
type Info struct {
	// Institution can be the name of the school. It may be used in printed
	// output, for example.
	Institution string
	// FirstAfternoonHour is the first "hour" (0-based index) which is to
	// be regarded as "afternoon".
	FirstAfternoonHour int
	// MiddayBreak specifies the "hours" (0-based indexes) which are to be
	// regarded as possible lunch breaks. They should be contiguous.
	MiddayBreak []int
	// Reference can be used to distinguish this particular data set from
	// others. It is not used in the code.
	Reference string
}

type Element struct {
	Id Ref
	// Not all Element objects use the Tag field
	Tag string // abbreviation/acronmym
}

type Elem interface {
	getId() Ref
	getTag() string
	setTag(string)
}

func (e *Element) getId() Ref {
	return e.Id
}

func (e *Element) getTag() string {
	return e.Tag
}

func (e *Element) setTag(tag string) {
	e.Tag = tag
}

// A Day represents a day of the timetable's week
type Day struct {
	Element
	Name string
}

// An Hour represents a lesson period ("hour") of a timetable's day
type Hour struct {
	Element
	Name  string
	Start string // start time, format hour:mins, e.g. "13:45"
	End   string // end time, format hour:mins, e.g. "14:30"
}

// A Teacher represents a member of staff, including various constraint
// information relevant for the timetable.
// It can be specified as a recourse for an activity.
type Teacher struct {
	Element
	Name      string
	Firstname string
	// NotAvailable is an ordered list of time-slots in which the teacher
	// is to be regarded as not available for the timetable.
	NotAvailable     []TimeSlot
	MinLessonsPerDay int  // default = -1 (unconstrained)
	MaxLessonsPerDay int  // default = -1 (unconstrained)
	MaxDays          int  // default = -1 (unconstrained)
	MaxGapsPerDay    int  // default = -1 (unconstrained)
	MaxGapsPerWeek   int  // default = -1 (unconstrained)
	MaxAfternoons    int  // default = -1 (unconstrained)
	LunchBreak       bool // whether the teacher should have a lunch break
}

// A Subject represents a taught subject, used for labelling a lesson, but
// it can also be used for any other activities which are timetabled (say,
// conferences).
type Subject struct {
	Element
	Name string
}

// A Room is a resource which can be specified for an activity.
type Room struct {
	Element
	Name string
	// NotAvailable is an ordered list of time-slots in which the room is to
	// be regarded as not available for the timetable.
	NotAvailable []TimeSlot
}

// IsReal reports whether r is an actual [Room], rather than a [RoomGroup] or
// [RoomChoiceGroup].
func (r *Room) IsReal() bool {
	return true
}

// A RoomGroup is a collection of [Room] items, all of which are "required".
type RoomGroup struct {
	Element
	Name  string
	Rooms []Ref
}

func (r *RoomGroup) IsReal() bool {
	return false
}

// A RoomChoiceGroup is a collection of [Room] items, one of which is
// "required".
type RoomChoiceGroup struct {
	Element
	Name  string
	Rooms []Ref
}

func (r *RoomChoiceGroup) IsReal() bool {
	return false
}

// A Class represents a collection of students and will generally correspond
// to a school class (not lesson). It includes various constraint
// information relevant for the timetable.
// See type [Group] (representing a subgroup of a class) for the student
// groups which can be specified as a resourse for an activity.
// A class often has a name which consists of a number and a letter or two.
// The number (Year field) represents the class's "year" (A.E. "grade"), the
// Letter field the text part (it can be more than one letter). The Tag field
// is the combination, e.g. "11A". The Name field can be used for a longer
// description of the class.
type Class struct {
	Element
	Name             string
	Year             int
	Letter           string
	NotAvailable     []TimeSlot
	Divisions        []Division
	MinLessonsPerDay int  // default = -1 (unconstrained)
	MaxLessonsPerDay int  // default = -1 (unconstrained)
	MaxGapsPerDay    int  // default = -1 (unconstrained)
	MaxGapsPerWeek   int  // default = -1 (unconstrained)
	MaxAfternoons    int  // default = -1 (unconstrained)
	LunchBreak       bool // whether the students should have a lunch break
	ForceFirstHour   bool // whether lessons need to start at hour 0
	ClassGroup       Ref  // the Group representing the whole class
}

type Group struct {
	Element
	// These fields do not belong in the JSON object:
	Class Ref `json:"-"`
}

// A Course specifies a collection of resources needed for a set of
// activities ([Lesson] elements). The [Subject] field is a sort of label.
type Course struct {
	Element
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     Ref // [Room], [RoomGroup] or [RoomChoiceGroup] element
	// These fields do not belong in the JSON object:
	Lessons []Ref `json:"-"`
}

func (c *Course) IsSuperCourse() bool {
	return false
}

func (c *Course) AddLesson(lref Ref) {
	c.Lessons = append(c.Lessons, lref)
}

// A SuperCourse specifies a collection of [SubCourse] elements which are
// associated with a set of activities ([Lesson] elements). The [Subject]
// field is a sort of label.
type SuperCourse struct {
	Element
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

// A SubCourse has no Lessons of its own, but shares those of its parent
// [SuperCourse] elements. A SubCourse may blong to more than one
// [SuperCourse]. Otherwise it is much like a [Course], bundling the
// necessary resources.
type SubCourse struct {
	Element
	SuperCourses []Ref
	Subject      Ref
	Groups       []Ref
	Teachers     []Ref
	Room         Ref //  [Room], [RoomGroup] or [RoomChoiceGroup] element
}

// A GeneralRoom covers  [Room], [RoomGroup] and [RoomChoiceGroup].
type GeneralRoom interface {
	IsReal() bool
}

// A Lesson is an activity which needs placing in the timetable.
// Its resources are determined by the course ([Course] or [SuperCourse]) to
// which it belongs.
type Lesson struct {
	Element
	Course   Ref   // [Course] or [SuperCourse] elements
	Duration int   // number of "hours" covered
	Day      int   // 0-based index, -1 for "unplaced"
	Hour     int   // 0-based index
	Fixed    bool  // whether the Lesson is unmovable
	Rooms    []Ref // actually allocated Room elements
	// Flags allows additional directions to be specified for the timetabling
	Flags      []string `json:",omitempty"`
	Background string   // colour, as "#RRGGBB"
	Footnote   string
}

// LessonCourse is a type of course which can have lessons, i.e. a
// [Course] or a [SuperCourse].
type LessonCourse interface {
	IsSuperCourse() bool // whether this is a SuperCourse
	// AddLesson is used to add a lesson to the course. When the data is
	// initially loaded the courses have no attached lessons. This list is
	// built using the course references in the Lesson elements.
	AddLesson(Ref) // add a lesson to the course
}

// Constraint is a rule used in the construction of a timetable.
//
// These can be very varied and they may have very little in common. Each
// implementation must have a distinguishing CType.
type Constraint interface {
	CType() string
}

// There is just one DbTopLevel. It is the root of the database.
// In general, the list fields should be ordered, where this is relevant.
type DbTopLevel struct {
	Info Info
	// ModuleData is for data supplied and managed by other packages
	ModuleData       map[string]any
	Days             []*Day
	Hours            []*Hour
	Teachers         []*Teacher
	Subjects         []*Subject
	Rooms            []*Room
	RoomGroups       []*RoomGroup       `json:",omitempty"`
	RoomChoiceGroups []*RoomChoiceGroup `json:",omitempty"`
	Classes          []*Class
	Groups           []*Group       `json:",omitempty"`
	Courses          []*Course      `json:",omitempty"`
	SuperCourses     []*SuperCourse `json:",omitempty"`
	SubCourses       []*SubCourse   `json:",omitempty"`
	Lessons          []*Lesson      `json:",omitempty"`
	Constraints      []Constraint   `json:",omitempty"`

	// These fields do not belong in the JSON object:
	Elements map[Ref]Elem `json:"-"`
}
