package base

const MAXWEIGHT = 100

func (db *DbTopLevel) addConstraint(c Constraint) {
	db.Constraints = append(db.Constraints, c)
}

// ++ AutomaticDifferentDays
// This Constraint applies to all courses (with more than one Lesson).
// If not present, all courses will by default apply it as a hard constraint,
// except for courses which have an overriding DAYS_BETWEEN constraint.
type AutomaticDifferentDays struct {
	Constraint           string
	Weight               int
	ConsecutiveIfSameDay bool
}

func (c *AutomaticDifferentDays) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewAutomaticDifferentDays() *AutomaticDifferentDays {
	c := &AutomaticDifferentDays{Constraint: "AutomaticDifferentDays"}
	db.addConstraint(c)
	return c
}

// ++ DaysBetween
// This constraint applies between the lessons of the individual courses.
// It does not connect the courses. If DaysBetween = 1, this constraint
// overrides the global AutomaticDifferentDays constraint for these courses.

type DaysBetween struct {
	Constraint           string
	Weight               int
	Courses              []Ref // Courses or SuperCourses
	DaysBetween          int
	ConsecutiveIfSameDay bool
}

func (c *DaysBetween) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewDaysBetween() *DaysBetween {
	c := &DaysBetween{Constraint: "DaysBetween"}
	db.addConstraint(c)
	return c
}

// ++ DaysBetweenJoin
// This constraint applies between the individual lessons of the two courses,
// not between the lessons of a course itself. That is, between course 1,
// lesson 1 and course 2 lesson 1; between course 1, lesson 1 and course 2,
// lesson 2, etc.

type DaysBetweenJoin struct {
	Constraint           string
	Weight               int
	Course1              Ref // Course or SuperCourse
	Course2              Ref // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
}

func (c *DaysBetweenJoin) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewDaysBetweenJoin() *DaysBetweenJoin {
	c := &DaysBetweenJoin{Constraint: "DaysBetweenJoin"}
	db.addConstraint(c)
	return c
}

// ++ LessonsEndDay

type LessonsEndDay struct {
	Constraint string
	Course     Ref
	Weight     int
}

func (c *LessonsEndDay) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewLessonsEndsDay() *LessonsEndDay {
	c := &LessonsEndDay{Constraint: "LessonsEndDay"}
	db.addConstraint(c)
	return c
}
