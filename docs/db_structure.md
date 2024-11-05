# Current Database Structure

Date: 31.10.2024

```
package db

// The structures used for the "database"
//TODO: Currently dealing only with the elements needed for the timetable

type Ref string // Element reference

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

type TimeSlot struct {
	Day  int
	Hour int
}

type Teacher struct {
	Id               Ref
	Name             string
	Tag              string
	Firstname        string
	NotAvailable     []TimeSlot
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxDays          int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
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

type RoomGroup struct {
	Id    Ref
	Name  string
	Tag   string
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
	Tag              string
	Year             int
	Letter           string
	NotAvailable     []TimeSlot
	Divisions        []Division
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
	LunchBreak       bool
	ForceFirstHour   bool
}

type Group struct {
	Id  Ref
	Tag string
}

type Division struct {
	Name   string
	Groups []Ref
}

type Course struct {
	Id       Ref
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type SuperCourse struct {
	Id      Ref
	Subject Ref
}

type SubCourse struct {
	Id          Ref
	SuperCourse Ref
	Subject     Ref
	Groups      []Ref
	Teachers    []Ref
	Room        Ref // Room, RoomGroup or RoomChoiceGroup Element
}

type Lesson struct {
	Id       Ref
	Course   Ref // Course or SuperCourse Elements
	Duration int
	Day      int
	Hour     int
	Fixed    bool
	Rooms    []Ref // only Room Elements
}

type DbTopLevel struct {
	Info             Info
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
}
```