package ttbase

import (
	"W365toFET/base"
	"strings"
)

// A DayGapConstraints encapsulates the data arising from the constraints
// "AUTOMATIC_DIFFERENT_DAYS", "DAYS_BETWEEN" and "DAYS_BETWEEN_JOIN".
type DayGapConstraints struct {
	DefaultDifferentDaysWeight               int
	DefaultDifferentDaysConsecutiveIfSameDay bool
	// Constraints maps the course references ([base.Course] and
	// [base.SuperCourse]) to a list of [base.DaysBetween] constraints
	// which affect them.
	CourseConstraints map[Ref][]*base.DaysBetween

	//TODO
	CrossCourseConstraints []MinDaysBetweenLessons // ???
}

// TODO?
type MinDaysBetweenLessons struct {
	// Result of processing constraints DifferentDays and DaysBetween
	Weight               int
	ConsecutiveIfSameDay bool
	Lessons              []int
	MinDays              int
}

// TODO?
type ParallelLessons struct {
	Weight       int
	LessonGroups [][]ActivityIndex
}

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
				dayGapConstraints.addCrossConstraintDaysBetween(ttinfo, cn)
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
// to the list for the course concerned.
func (dgdata *DayGapConstraints) addConstraintDaysBetween(
	c *base.DaysBetween,
) {
	for _, cref := range c.Courses {
		dgdata.CourseConstraints[cref] = append(
			dgdata.CourseConstraints[cref], c)
	}
}

//TODO: This probably shouldn't jump immediately to lesson handling as the
// lesson handling is now different, handled by the activity groups ...

// addCrossConstraintDaysBetween handles "DAYS_BETWEEN_JOIN" constraints,
// adding them to ??TODO?? the list for the course concerned.
func (dgdata *DayGapConstraints) addCrossConstraintDaysBetween(
	ttinfo *TtInfo,
	c *base.DaysBetweenJoin,
) {
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
}

//TODO: This probably needs adjusting for the new activity group processing

// addParallelCoursesConstraint constrains the lessons of the given courses
// to start at the same time (constraint "PARALLEL_COURSES").
// The courses must have the same number of lessons and the durations of the
// corresponding lessons must also be the same.
func (ttinfo *TtInfo) addParallelCoursesConstraint(c *base.ParallelCourses) {
	// Check lesson lengths
	footprint := []int{} // lesson sizes
	ll := 0              // number of lessons in each course
	var llists [][]int   // collect the parallel lessons
	for i, cref := range c.Courses {
		cinfo := ttinfo.CourseInfo[cref]
		if i == 0 {
			ll = len(cinfo.Lessons)
			llists = make([][]int, ll)
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
			llists[j] = append(llists[j], lix)
		}
	}
	// llists is now a list of lists of parallel TtLesson indexes.
	ttinfo.ParallelLessons = append(ttinfo.ParallelLessons,
		ParallelLessons{
			Weight:       c.Weight,
			LessonGroups: llists,
		})
}
