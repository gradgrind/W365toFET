package ttbase

import (
	"W365toFET/base"
	"slices"
	"strings"
)

// A DayGapConstraints encapsulates the data arising from the constraints
// "AUTOMATIC_DIFFERENT_DAYS", "DAYS_BETWEEN" and "DAYS_BETWEEN_JOIN".
type DayGapConstraints struct {
	DefaultDifferentDaysWeight               int
	DefaultDifferentDaysConsecutiveIfSameDay bool
	// CourseConstraints maps the course references ([base.Course] and
	// [base.SuperCourse]) to a list of [base.DaysBetween] constraints
	// which affect them.
	CourseConstraints map[Ref][]*base.DaysBetween
	// CrossCourseConstraints maps the course references ([base.Course] and
	// [base.SuperCourse]) to a list of [base.DaysBetweenJoin] constraints
	// which affect them.
	CrossCourseConstraints map[Ref][]*base.DaysBetweenJoin
}

// TODO?
type MinDaysBetweenLessons struct {
	// Result of processing constraints DifferentDays and DaysBetween
	Weight               int
	ConsecutiveIfSameDay bool
	Lessons              []int
	MinDays              int
}

/* TODO?
type ParallelLessons struct {
	Weight       int
	LessonGroups [][]ActivityIndex
}
*/

func (ttinfo *TtInfo) processConstraints() {
	// Some constraints can be "preprocessed" into more convenient structures.
	db := ttinfo.Db

	dayGapConstraints := &DayGapConstraints{
		DefaultDifferentDaysWeight: -1, // uninitialized
		//DefaultDifferentDaysConsecutiveIfSameDay: false,
	}
	ttinfo.DayGapConstraints = dayGapConstraints
	ttinfo.Constraints = map[string][]any{}
	for _, c := range db.Constraints {
		{
			cn, ok := c.(*base.AutomaticDifferentDays)
			if ok {
				dayGapConstraints.constraintAutomaticDifferentDays(cn)
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetween)
			if ok {
				dayGapConstraints.addConstraintDaysBetween(cn)
				continue
			}
		}
		{
			cn, ok := c.(*base.DaysBetweenJoin)
			if ok {
				dayGapConstraints.addCrossConstraintDaysBetween(cn)
				continue
			}
		}
		{
			//TODO: This must happen BEFORE the result is used to add stuff
			// to the Activities!

			cn, ok := c.(*base.ParallelCourses)
			if ok {
				ttinfo.addParallelCoursesConstraint(cn)
				continue
			}
		}
		// Collect the other constraints according to type
		ctype := c.CType()
		ttinfo.Constraints[ctype] = append(ttinfo.Constraints[ctype], c)
	}
	// Resolve the differentDays constraints into days-between-lessons.
	if dayGapConstraints.DefaultDifferentDaysWeight < 0 {
		dayGapConstraints.DefaultDifferentDaysWeight = base.MAXWEIGHT
	}
}

// constraintAutomaticDifferentDays handles an "AUTOMATIC_DIFFERENT_DAYS"
// constraint, of which there may be at most 1. It sepcifies that the lessons
// of a course should take place on distinct days. When this constraint is
// not specified, it is made a hard constraint automatically.
func (dgdata *DayGapConstraints) constraintAutomaticDifferentDays(
	c *base.AutomaticDifferentDays,
) {
	if dgdata.DefaultDifferentDaysWeight < 0 {
		dgdata.DefaultDifferentDaysWeight = c.Weight
		dgdata.DefaultDifferentDaysConsecutiveIfSameDay =
			c.ConsecutiveIfSameDay
	} else {
		base.Bug.Fatalln(
			"More than one AutomaticDifferentDays constraint")
	}
}

// addConstraintDaysBetween handles "DAYS_BETWEEN" constraints, adding them
// to the list for each of the courses concerned.
func (dgdata *DayGapConstraints) addConstraintDaysBetween(
	c *base.DaysBetween,
) {
	for _, cref := range c.Courses {
		dgdata.CourseConstraints[cref] = append(
			dgdata.CourseConstraints[cref], c)
	}
}

// addCrossConstraintDaysBetween handles "DAYS_BETWEEN_JOIN" constraints,
// adding them to the list for the courses concerned.
func (dgdata *DayGapConstraints) addCrossConstraintDaysBetween(
	//	ttinfo *TtInfo,
	c *base.DaysBetweenJoin,
) {
	dgdata.CrossCourseConstraints[c.Course1] = append(
		dgdata.CrossCourseConstraints[c.Course1], c)
	dgdata.CrossCourseConstraints[c.Course2] = append(
		dgdata.CrossCourseConstraints[c.Course2], c)

	/*TODO--
	c1 := ttinfo.CourseInfo[c.Course1]
	c2 := ttinfo.CourseInfo[c.Course2]
	for _, l1ref := range c1.Lessons {
		l1fixed := ttinfo.Activities[l1ref].Fixed
		for _, l2ref := range c2.Lessons {
			if l1fixed && ttinfo.Activities[l2ref].Fixed {
				// both fixed => no constraint
				continue
			}
			dgdata.CrossCourseConstraints = append(
				dgdata.CrossCourseConstraints, MinDaysBetweenLessons{
					Weight:               c.Weight,
					ConsecutiveIfSameDay: c.ConsecutiveIfSameDay,
					Lessons:              []int{l1ref, l2ref},
					MinDays:              c.DayGap,
				},
			)
		}
	}
	*/
}

// addParallelCoursesConstraint constrains the lessons of the given courses
// to start at the same time (constraint "PARALLEL_COURSES").
// The courses must have the same number of lessons and the durations of the
// corresponding lessons must also be the same.
func (ttinfo *TtInfo) addParallelCoursesConstraint(c *base.ParallelCourses) {
	ttinfo.HardParallelCourses = map[Ref][]Ref{}
	ttinfo.SoftParallelCourses = map[Ref][]*base.ParallelCourses{}

	pclists := map[Ref][]Ref{} // for checking for duplicate constraints
	// Check lesson lengths
	footprint := []int{} // lesson sizes
	ll := 0              // number of lessons in each course
	//var llists [][]int   // collect the parallel lessons
	for i, cref := range c.Courses {
		cinfo := ttinfo.CourseInfo[cref]
		if i == 0 {
			ll = len(cinfo.Lessons)
			//llists = make([][]int, ll)
		} else if len(cinfo.Lessons) != ll {
			clist := []string{}
			for _, cr := range c.Courses {
				clist = append(clist, string(cr))
			}
			base.Error.Fatalf("Parallel courses have different"+
				" lessons: %s\n",
				strings.Join(clist, ","))
		}
		for j, lix := range cinfo.Lessons {
			a := ttinfo.Activities[lix]
			if i == 0 {
				footprint = append(footprint, a.Duration)
			} else if a.Duration != footprint[j] {
				clist := []string{}
				for _, cr := range c.Courses {
					clist = append(clist, string(cr))
				}
				base.Error.Fatalf("Parallel courses have lesson"+
					" mismatch: %s\n",
					strings.Join(clist, ","))
			}
			//llists[j] = append(llists[j], lix)
		}

		// Check for duplicate constraints
		pc, ok := pclists[cref]
		if ok {
			for _, cr := range c.Courses {
				if cr == cref {
					continue
				}
				if slices.Contains(pc, cr) {
					base.Error.Fatalf("Courses subject to more than one"+
						" parallel constraint:\n  -- %s\n  -- %s\n",
						ttinfo.View(cinfo),
						ttinfo.View(ttinfo.CourseInfo[cr]))
				}
				pclists[cref] = append(pclists[cref], cr)
			}
		}

		// Treat weight = MAXWEIGHT as a special case
		if c.Weight == base.MAXWEIGHT {
			// For hard constraints, link each course to its parallel courses
			for _, cr := range c.Courses {
				if cr == cref {
					continue
				}
				ttinfo.HardParallelCourses[cref] = append(
					ttinfo.HardParallelCourses[cref], cr)
			}
		} else {
			// For soft constraints, link to the constraint from each of the
			// courses concerned
			ttinfo.SoftParallelCourses[cref] = append(
				ttinfo.SoftParallelCourses[cref], c)
		}
	}
	/*
		// llists is now a list of lists of parallel TtLesson indexes.
		ttinfo.ParallelLessons = append(ttinfo.ParallelLessons,
			ParallelLessons{
				Weight:       c.Weight,
				LessonGroups: llists,
			})
	*/
}
