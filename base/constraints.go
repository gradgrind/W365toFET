package base

const MAXWEIGHT = 100

func (db *DbTopLevel) addConstraint(c Constraint) {
	db.Constraints = append(db.Constraints, c)
}

// ++ ActivitiesEndDay

type ActivitiesEndDay struct {
	Constraint string
	Weight     int
	Course     Ref
}

func (c *ActivitiesEndDay) CType() string {
	return c.Constraint
}

func (c *ActivitiesEndDay) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewActivitiesEndDay() *ActivitiesEndDay {
	c := &ActivitiesEndDay{Constraint: "ActivitiesEndDay"}
	db.addConstraint(c)
	return c
}

// ++ BeforeAfterHour
// Permissible hours are before or after the specified hour, not including
// the specified hour.

type BeforeAfterHour struct {
	Constraint string
	Weight     int
	Courses    []Ref
	After      bool
	Hour       int
}

func (c *BeforeAfterHour) CType() string {
	return c.Constraint
}

func (c *BeforeAfterHour) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewBeforeAfterHour() *BeforeAfterHour {
	c := &BeforeAfterHour{Constraint: "BeforeAfterHour"}
	db.addConstraint(c)
	return c
}

// ++ AutomaticDifferentDays
// This Constraint applies to all courses (with more than one Activity).
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

func (c *AutomaticDifferentDays) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewAutomaticDifferentDays() *AutomaticDifferentDays {
	c := &AutomaticDifferentDays{Constraint: "AutomaticDifferentDays"}
	db.addConstraint(c)
	return c
}

// ++ DaysBetween
// This constraint applies between the activitys of the individual courses.
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

func (c *DaysBetween) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewDaysBetween() *DaysBetween {
	c := &DaysBetween{Constraint: "DaysBetween"}
	db.addConstraint(c)
	return c
}

// ++ DaysBetweenJoin
// This constraint applies between the individual activities of the two courses,
// not between the activities of a course itself. That is, between course 1,
// activity 1 and course 2 activity 1; between course 1, activity 1 and course 2,
// activity 2, etc.

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

func (c *DaysBetweenJoin) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewDaysBetweenJoin() *DaysBetweenJoin {
	c := &DaysBetweenJoin{Constraint: "DaysBetweenJoin"}
	db.addConstraint(c)
	return c
}

// ++ ParallelCourses
// The activities of the courses specified here should be at the same time.
// To avoid complications, it is required that the number and lengths of
// activities be the same in each course.

type ParallelCourses struct {
	Constraint string
	Weight     int
	Courses    []Ref // Courses or SuperCourses
}

func (c *ParallelCourses) CType() string {
	return c.Constraint
}

func (c *ParallelCourses) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewParallelCourses() *ParallelCourses {
	c := &ParallelCourses{Constraint: "ParallelCourses"}
	db.addConstraint(c)
	return c
}

// ++ DoubleActivityNotOverBreaks

// There should be at most one of these. The breaks are immediately before
// the specified hours.

type DoubleActivityNotOverBreaks struct {
	Constraint string
	Weight     int
	Hours      []int
}

func (c *DoubleActivityNotOverBreaks) CType() string {
	return c.Constraint
}

func (c *DoubleActivityNotOverBreaks) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewDoubleActivityNotOverBreaks() *DoubleActivityNotOverBreaks {
	c := &DoubleActivityNotOverBreaks{Constraint: "DoubleActivityNotOverBreaks"}
	db.addConstraint(c)
	return c
}

/* TODO: Is this really useful? The W365 front end doesn't currently support it
// and the MinHoursFollowing may be more useful.
// ++ NotOnSameDay

type NotOnSameDay struct {
	Constraint string
	Weight     int
	Subjects   []Ref
}

func (c *NotOnSameDay) CType() string {
	return c.Constraint
}

func (db *DbTopLevel) NewNotOnSameDay() *NotOnSameDay {
	c := &NotOnSameDay{Constraint: "NotOnSameDay"}
	db.addConstraint(c)
	return c
}
*/

//TODO ... more?

// ++ MinHoursFollowing

type MinHoursFollowing struct {
	Constraint string
	Weight     int
	Course1    Ref // Course or SuperCourse
	Course2    Ref // Course or SuperCourse
	Hours      int
}

func (c *MinHoursFollowing) CType() string {
	return c.Constraint
}

func (c *MinHoursFollowing) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

func (db *DbTopLevel) NewMinHoursFollowing() *MinHoursFollowing {
	c := &MinHoursFollowing{Constraint: "MinHoursFollowing"}
	db.addConstraint(c)
	return c
}
