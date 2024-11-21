package base

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
