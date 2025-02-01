package ttbase

import (
	"W365toFET/base"
	"slices"
)

// A DayGapConstraints encapsulates the data arising from the constraints
// "AUTOMATIC_DIFFERENT_DAYS", "DAYS_BETWEEN" and "DAYS_BETWEEN_JOIN".
type DayGapConstraints struct {
	DefaultDifferentDaysWeight               int
	DefaultDifferentDaysConsecutiveIfSameDay bool

	// CourseConstraints maps an [ActivityGroupIndex] to a list of
	// [DaysBetweenConstraint] constraints for that activity group (between
	// the lessons of that activity group).
	CourseConstraints map[ActivityGroupIndex][]DaysBetweenConstraint

	// CrossCourseConstraints maps an [ActivityGroupIndex] to a map (which
	// is unlikely to contain more than one entry ...) from a second
	// [ActivityGroupIndex] to a list of [DaysBetweenConstraint] constraints
	// for this pair of activity groups (between the lessons of the two
	// activity groups).
	CrossCourseConstraints map[ActivityGroupIndex]map[ActivityGroupIndex][]DaysBetweenConstraint
}

type DaysBetweenConstraint struct {
	Weight               int
	DayGap               int
	ConsecutiveIfSameDay bool
}

func (ttinfo *TtInfo) processConstraints() {
	// Some constraints can be "preprocessed" into more convenient structures.
	db := ttinfo.Db

	// For parallel courses
	ttinfo.HardParallelCourses = map[Ref][]Ref{}
	ttinfo.SoftParallelCourses = []*base.ParallelCourses{}
	pclists := map[Ref][]Ref{} // to check for duplicate parallel constraints

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
			cn, ok := c.(*base.ParallelCourses)
			if ok {
				ttinfo.addParallelCoursesConstraint(pclists, cn)
				continue
			}
		}
		// Collect the other constraints according to type
		ctype := c.CType()
		ttinfo.Constraints[ctype] = append(ttinfo.Constraints[ctype], c)
	}
	// If there has been no different-days override, set to the default
	if dayGapConstraints.DefaultDifferentDaysWeight < 0 {
		dayGapConstraints.DefaultDifferentDaysWeight = base.MAXWEIGHT
	}
}

// constraintAutomaticDifferentDays handles an "AUTOMATIC_DIFFERENT_DAYS"
// constraint, of which there may be at most one. It sepcifies that the
// lessons of a course should take place on distinct days. When this
// constraint is not specified, it is made a hard constraint automatically.
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

// addParallelCoursesConstraint constrains the lessons of the given courses
// to start at the same time (constraint "PARALLEL_COURSES").
// The courses must have the same number of lessons and the durations of the
// corresponding lessons must also be the same.
func (ttinfo *TtInfo) addParallelCoursesConstraint(
	pclists map[Ref][]Ref, // to check for duplicate constraints
	c *base.ParallelCourses,
) {
	// Check lesson lengths
	footprint := []int{} // lesson sizes
	ll := 0              // number of lessons in each course
	for i, cref := range c.Courses {
		cinfo := ttinfo.CourseInfo[cref]
		if i == 0 {
			ll = len(cinfo.Lessons)
		} else if len(cinfo.Lessons) != ll {
			base.Error.Printf("Parallel courses have different"+
				" lessons: %s\n", ttinfo.View(cinfo))
			return
		}
		for j, l := range cinfo.Lessons {
			if i == 0 {
				footprint = append(footprint, l.Duration)
			} else if l.Duration != footprint[j] {
				clist := []string{}
				for _, cr := range c.Courses {
					clist = append(clist, string(cr))
				}
				base.Error.Printf("Parallel courses have lesson"+
					" mismatch: %s\n", ttinfo.View(cinfo))
				return
			}
		}

		// Check for duplicate constraints
		pc, ok := pclists[cref]
		if ok {
			for _, cr := range c.Courses {
				if cr == cref {
					continue
				}
				if slices.Contains(pc, cr) {
					base.Error.Printf("Courses subject to more than one"+
						" parallel constraint:\n  -- %s\n  -- %s\n",
						ttinfo.View(cinfo),
						ttinfo.View(ttinfo.CourseInfo[cr]))
					return
				}
				pclists[cref] = append(pclists[cref], cr)
			}
		}
	}
	// Add the constraint
	for _, cref := range c.Courses {
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
			// For soft constraints, simply add to the list
			ttinfo.SoftParallelCourses = append(
				ttinfo.SoftParallelCourses, c)
		}
	}
}
