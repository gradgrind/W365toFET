package base

import (
	"slices"

	"github.com/gofrs/uuid/v5"
)

func NewDb() *DbTopLevel {
	db := &DbTopLevel{}
	db.Elements = map[Ref]any{}
	return db
}

func (db *DbTopLevel) NewDay(ref Ref) *Day {
	e := &Day{}
	e.Id = db.addElement(ref, e)
	db.Days = append(db.Days, e)
	return e
}

func (db *DbTopLevel) NewHour(ref Ref) *Hour {
	e := &Hour{}
	e.Id = db.addElement(ref, e)
	db.Hours = append(db.Hours, e)
	return e
}

func (db *DbTopLevel) NewTeacher(ref Ref) *Teacher {
	e := &Teacher{}
	e.Id = db.addElement(ref, e)
	db.Teachers = append(db.Teachers, e)
	return e
}

func (db *DbTopLevel) NewSubject(ref Ref) *Subject {
	e := &Subject{}
	e.Id = db.addElement(ref, e)
	db.Subjects = append(db.Subjects, e)
	return e
}

func (db *DbTopLevel) NewRoom(ref Ref) *Room {
	e := &Room{}
	e.Id = db.addElement(ref, e)
	db.Rooms = append(db.Rooms, e)
	return e
}

func (db *DbTopLevel) NewRoomGroup(ref Ref) *RoomGroup {
	e := &RoomGroup{}
	e.Id = db.addElement(ref, e)
	db.RoomGroups = append(db.RoomGroups, e)
	return e
}

func (db *DbTopLevel) NewRoomChoiceGroup(ref Ref) *RoomChoiceGroup {
	e := &RoomChoiceGroup{}
	e.Id = db.addElement(ref, e)
	db.RoomChoiceGroups = append(db.RoomChoiceGroups, e)
	return e
}

func (db *DbTopLevel) NewClass(ref Ref) *Class {
	e := &Class{}
	e.Id = db.addElement(ref, e)
	db.Classes = append(db.Classes, e)
	return e
}

func (db *DbTopLevel) NewGroup(ref Ref) *Group {
	e := &Group{}
	e.Id = db.addElement(ref, e)
	db.Groups = append(db.Groups, e)
	return e
}

func (db *DbTopLevel) NewCourse(ref Ref) *Course {
	e := &Course{}
	e.Id = db.addElement(ref, e)
	db.Courses = append(db.Courses, e)
	return e
}

func (db *DbTopLevel) NewSuperCourse(ref Ref) *SuperCourse {
	e := &SuperCourse{}
	e.Id = db.addElement(ref, e)
	db.SuperCourses = append(db.SuperCourses, e)
	return e
}

func (db *DbTopLevel) NewSubCourse(ref Ref) *SubCourse {
	e := &SubCourse{}
	e.Id = db.addElement(ref, e)
	db.SubCourses = append(db.SubCourses, e)
	return e
}

func (db *DbTopLevel) NewLesson(ref Ref) *Lesson {
	e := &Lesson{}
	e.Id = db.addElement(ref, e)
	db.Lessons = append(db.Lessons, e)
	return e
}

func (db *DbTopLevel) newId() Ref {
	// Create a Version 4 UUID.
	u2, err := uuid.NewV4()
	if err != nil {
		Error.Fatalf("Failed to generate UUID: %v", err)
	}
	return Ref(u2.String())
}

func (db *DbTopLevel) addElement(ref Ref, element any) Ref {
	if ref == "" {
		ref = db.newId()
	}
	_, nok := db.Elements[ref]
	if nok {
		Error.Fatalf("Element Id defined more than once:\n  %s\n", ref)
	}
	db.Elements[ref] = element
	return ref
}

// TODO???
func (db *DbTopLevel) PrepareDb() {
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
	for _, sbc := range db.SubCourses {
		for _, spcref := range sbc.SuperCourses {
			spc := db.Elements[spcref].(*SuperCourse)
			spc.SubCourses = append(spc.SubCourses, sbc.Id)
		}
	}

	// Collect the Lessons for each Course and SuperCourse
	for _, l := range db.Lessons {
		db.Elements[l.Course].(LessonCourse).AddLesson(l.Id)
	}

	// Expand Group information
	for _, c := range db.Classes {
		db.Elements[c.ClassGroup].(*Group).Class = c.Id // Tag is empty.
		for _, d := range c.Divisions {
			for _, gref := range d.Groups {
				db.Elements[gref].(*Group).Class = c.Id
			}
		}
	}
	// Check that all groups belong to a class
	for _, g := range db.Groups {
		if g.Class == "" {
			// This is a loader failure, it should not be possible.
			Bug.Fatalf("Group not in Class: %s\n", g.Id)
		}
	}
}

func (db *DbTopLevel) CheckDbBasics() {
	// This function is provided for use by code which needs the following
	// Elements to be provided.
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
}

// Interface for Course and SubCourse elements
type CourseInterface interface {
	GetId() Ref
	GetGroups() []Ref
	GetTeachers() []Ref
	GetSubject() Ref
	GetRoom() Ref
}

func (c *Course) GetId() Ref            { return c.Id }
func (c *SubCourse) GetId() Ref         { return c.Id }
func (c *Course) GetGroups() []Ref      { return c.Groups }
func (c *SubCourse) GetGroups() []Ref   { return c.Groups }
func (c *Course) GetTeachers() []Ref    { return c.Teachers }
func (c *SubCourse) GetTeachers() []Ref { return c.Teachers }
func (c *Course) GetSubject() Ref       { return c.Subject }
func (c *SubCourse) GetSubject() Ref    { return c.Subject }
func (c *Course) GetRoom() Ref          { return c.Room }
func (c *SubCourse) GetRoom() Ref       { return c.Room }
