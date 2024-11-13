package base

import "slices"

/* TODO
func (db *DbTopLevel) NewId() Ref {
	return Ref(fmt.Sprintf("#%d", db.MaxId+1))
}
*/

func (db *DbTopLevel) InitDb() {

	// Build the Element reference map.

	for _, n := range db.Days {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Hours {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Teachers {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Subjects {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Rooms {
		db.Elements[n.Id] = n
	}
	for _, n := range db.RoomGroups {
		db.Elements[n.Id] = n
	}
	for _, n := range db.RoomChoiceGroups {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Classes {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Groups {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Courses {
		db.Elements[n.Id] = n
	}
	for _, n := range db.SuperCourses {
		db.Elements[n.Id] = n
	}
	for _, n := range db.SubCourses {
		db.Elements[n.Id] = n
	}
	for _, n := range db.Lessons {
		db.Elements[n.Id] = n
	}

	// *** Initializations

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
		spcref := sbc.SuperCourse
		spc := db.Elements[spcref].(*SuperCourse)
		spc.SubCourses = append(spc.SubCourses, sbc.Id)
	}

	// Collect the Lessons for each Course and SuperCourse
	for _, l := range db.Lessons {
		db.Elements[l.Course].(LessonCourse).AddLesson(l.Id)
	}

	if db.Constraints == nil {
		db.Constraints = make(map[string]any)
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
			Error.Printf("Group not in Class: %s\n", g.Id)
			//TODO: Remove it?
		}
	}
}

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
