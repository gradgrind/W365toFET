package ttbase

import (
	"W365toFET/base"
	"slices"
)

type DaysBetweenConstraint struct {
	Weight               int
	HoursGap             int
	ConsecutiveIfSameDay bool
	LessonUnits          []LessonUnitIndex
}

// processDaysBetween collects days-between constraints according to gaps,
// reporting and ignoring conflicts. If there is no constraint for gap = 1,
// and no other hard constraint, the default is used. The constraints are
// returned as a list of [DaysBetweenLessons] constraints.
func (ttinfo *TtInfo) processDaysBetween(
	ag *ActivityGroup,
) []DaysBetweenLessons {
	gapmap := map[int]DaysBetweenLessons{}
	hardgap := 0 // Record any such hard constraint
	// Include hard-parallel courses
	dgc := ttinfo.DayGapConstraints
	daygapmap := dgc.CourseConstraints
	for _, pcref := range ag.Courses {
		for _, dbl := range daygapmap[pcref] {
			gap := dbl.DayGap
			if dbl.Weight == base.MAXWEIGHT {
				// A hard constraint – there should be at most one
				if hardgap == 0 {
					hardgap = gap
				} else {
					base.Error.Printf("Multiple hard DAYS_BETWEEN"+
						" constraints for a course (possibly with"+
						" parallel courses):\n -- %s\n",
						ttinfo.View(ttinfo.CourseInfo[pcref]))
					continue
				}
			}
			if _, nok := gapmap[gap]; nok {
				base.Error.Printf("Multiple DAYS_BETWEEN constraints"+
					" with gap %d for a course (possibly with parallel"+
					" courses):\n  -- %s\n",
					gap, ttinfo.View(ttinfo.CourseInfo[pcref]))
				continue
			}
			gapmap[gap] = dbl
		}
	}
	// Check for ineffective constraints and whether the default is needed
	dbllist := []DaysBetweenLessons{}
	ng1dbc := hardgap == 0 // whether a constraint with gap = 1 is needed
	for gap, dbl := range gapmap {
		if gap < hardgap {
			base.Error.Printf("DAYS_BETWEEN constraint with gap smaller"+
				" than hard gap, course (or parallel course):\n -- %s\n",
				ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]))
			continue
		}
		if gap == 1 {
			ng1dbc = false
		}
		dbllist = append(dbllist, dbl)
	}
	if ng1dbc {
		dbllist = append(dbllist, DaysBetweenLessons{
			Weight:               dgc.DefaultDifferentDaysWeight,
			ConsecutiveIfSameDay: dgc.DefaultDifferentDaysConsecutiveIfSameDay,
			DayGap:               1,
		})
	}
	return dbllist
}

// processCrossDaysBetween collects cross-days-between on the same principle
// as the days-between constraints, but with the extra complication of there
// being two courses involved. These must be processed after all ActivityGroup
// elements have been generated, so that their TtLesson indexes are available.
func (ttinfo *TtInfo) processCrossDaysBetween() {
	ttplaces := ttinfo.Placements
	dgc := ttinfo.DayGapConstraints
	xdaygapmap := dgc.CrossCourseConstraints
	for _, ag := range ttplaces.ActivityGroups {
		// First collect the constraints for each cross-course. Note that
		// the ActivityGroup indexes are used rather than the course
		// references, in case of hard-parallel courses.

		xdblmap := map[ActivityGroupIndex][]CrossDaysBetweenLessons{}
		for _, pcref := range ag.Courses {
			// ... for each course in the ActivityGroup

			for _, xdbl := range xdaygapmap[pcref] {
				xcourse := xdbl.CrossCourse

				if slices.Contains(ag.Courses, xcourse) {
					base.Error.Printf("DAYS_BETWEEN_JOIN constraint acts"+
						" on parallel courses:\n  -- %s\n  -- %s\n",
						ttinfo.View(ttinfo.CourseInfo[pcref]),
						ttinfo.View(ttinfo.CourseInfo[xcourse]),
					)
					continue
				}
				xix := ttplaces.CourseActivityGroup[xcourse]
				xdblmap[xix] = append(xdblmap[xix], xdbl)
			}
		}
		xdbllist0 := []CrossDaysBetweenLessons{}
		// Deal with each cross-course separately
		for xix, xdbllist := range xdblmap {
			// Divide according to gap
			gapmap := map[int]CrossDaysBetweenLessons{}
			hardgap := 0 // only one hard constraint is permissible
			for _, xdbl := range xdbllist {
				gap := xdbl.DayGap
				if xdbl.Weight == base.MAXWEIGHT {
					// A hard constraint
					if hardgap == 0 {
						hardgap = gap
					} else {
						xcourse := ttplaces.ActivityGroups[xix].Courses[0]
						base.Error.Printf("Multiple hard DAYS_BETWEEN_JOIN"+
							" constraints for a course (possibly with"+
							" parallel courses):\n -- %s\n",
							ttinfo.View(ttinfo.CourseInfo[xcourse]),
						)
						continue
					}
				}
				if _, nok := gapmap[gap]; nok {
					xcourse := ttplaces.ActivityGroups[xix].Courses[0]
					base.Error.Printf("Multiple DAYS_BETWEEN_JOIN"+
						" constraints with gap %d for a course (possibly"+
						" with parallel courses):\n  -- %s",
						gap, ttinfo.View(ttinfo.CourseInfo[xcourse]),
					)
					continue
				}
				gapmap[gap] = xdbl
			}

			// Check for ineffective constraints
			for gap, xdbl := range gapmap {
				if gap < hardgap {
					xcinfo := ttinfo.CourseInfo[xdbl.CrossCourse]
					base.Error.Printf("DAYS_BETWEEN constraint with gap smaller"+
						" than hard gap, course:\n -- %s\n",
						ttinfo.View(xcinfo))
					continue
				}
				xdbllist0 = append(xdbllist0, xdbl)
			}

		}

		// Each constraint in xdbllist0 concerns the activity group ag and
		// the one referenced in the constraint. Every lesson of each of the
		// two activity groups needs a [DaysBetweenConstraint] listing the
		// lessons of the other activity group.
		for _, xdbl := range xdbllist0 {
			xix1 := ttplaces.CourseActivityGroup[xdbl.CrossCourse]
			xag := ttplaces.ActivityGroups[xix1]
			// Add constraints to the lessons of this activity group
			for _, lix := range xag.LessonUnits {
				xl := ttplaces.TtLessons[lix]
				xl.DaysBetween = append(xl.DaysBetween, DaysBetweenConstraint{
					Weight: xdbl.Weight,
					// HoursGap assumes the use of an adequately long
					// DayLength, including buffer space at the end of the
					// real lesson slots
					HoursGap:             xdbl.DayGap * ttinfo.DayLength,
					ConsecutiveIfSameDay: xdbl.ConsecutiveIfSameDay,
					LessonUnits:          ag.LessonUnits,
				})
			}
			// Add constraints to the lessons of the ag activity group
			for _, lix := range xag.LessonUnits {
				l := ttplaces.TtLessons[lix]
				l.DaysBetween = append(l.DaysBetween, DaysBetweenConstraint{
					Weight: xdbl.Weight,
					// HoursGap assumes the use of an adequately long
					// DayLength, including buffer space at the end of the
					// real lesson slots
					HoursGap:             xdbl.DayGap * ttinfo.DayLength,
					ConsecutiveIfSameDay: xdbl.ConsecutiveIfSameDay,
					LessonUnits:          xag.LessonUnits,
				})
			}
		}
	}
}
