package ttbase

import (
	"W365toFET/base"
)

//TODO: Make the filtered constraints available to the (fet) backend,
// presumably in ttinfo.DayGapConstraints.

// The days-between constraints must be processed after all ActivityGroup
// elements have been generated, so that their TtLesson indexes are available.

type DaysBetweenLessons struct {
	Weight               int
	HoursGap             int
	ConsecutiveIfSameDay bool
	LessonUnits          []LessonUnitIndex
}

// TODO: call this from PrepareCoreData (in base) after addActivityInfo()?
// Cut addActivityInfo() before PrepareActivityGroups() and call this
// after that?
func (ttinfo *TtInfo) processDaysBetweenConstraints() {
	// Handle the DAYS_BETWEEN constraints. Sort them according to their
	// activity groups.
	agconstraints := map[ActivityGroupIndex][]*base.DaysBetween{}
	for _, c := range ttinfo.Constraints["DAYS_BETWEEN"] {
		dbc := c.(*base.DaysBetween)
		for _, cref := range dbc.Courses {
			agix := ttinfo.Placements.CourseActivityGroup[cref]
			agconstraints[agix] = append(agconstraints[agix], dbc)
		}
	}
	ttinfo.processDaysBetween(agconstraints)

	// Handle the DAYS_BETWEEN_JOIN constraints. Sort them according to their
	// activity groups.
	agxconstraints := map[ActivityGroupIndex][]*base.DaysBetweenJoin{}
	for _, c := range ttinfo.Constraints["DAYS_BETWEEN_JOIN"] {
		dbc := c.(*base.DaysBetweenJoin)
		agix := ttinfo.Placements.CourseActivityGroup[dbc.Course1]
		if ttinfo.Placements.CourseActivityGroup[dbc.Course2] == agix {
			base.Error.Printf("DAYS-BETWEEN_JOIN constraint between"+
				" parallel courses:\n  -- %s\n  -- %s\n",
				ttinfo.View(ttinfo.CourseInfo[dbc.Course1]),
				ttinfo.View(ttinfo.CourseInfo[dbc.Course2]),
			)
			continue
		}
		agxconstraints[agix] = append(agxconstraints[agix], dbc)
	}
	ttinfo.processCrossDaysBetween(agxconstraints)
}

// processDaysBetween collects days-between constraints according to gaps,
// reporting and ignoring conflicts. If there is no constraint for gap = 1,
// and no other hard constraint, the default is used.
// The constraints are applied to the [TtLesson] items of the activity
// groups.
func (ttinfo *TtInfo) processDaysBetween(
	agconstraints map[ActivityGroupIndex][]*base.DaysBetween,
) {
	dgc := ttinfo.DayGapConstraints
	dgc.CourseConstraints = map[ActivityGroupIndex][]DaysBetweenConstraint{}
	ttplaces := ttinfo.Placements
	for agix, dbclist0 := range agconstraints {
		ag := ttplaces.ActivityGroups[agix]
		// Sort according to gap
		hardgap := 0 // Record any hard constraint
		gapmap := map[int]*base.DaysBetween{}
		for _, dbc := range dbclist0 {
			gap := dbc.DayGap
			if dbc.Weight == base.MAXWEIGHT {
				// A hard constraint – there should be at most one
				if hardgap == 0 {
					hardgap = gap
				} else {
					base.Error.Printf("Multiple hard DAYS_BETWEEN"+
						" constraints for a course (possibly with"+
						" parallel courses):\n -- %s\n",
						ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]))
					continue
				}
			}
			if _, nok := gapmap[agix]; nok {
				base.Error.Printf("Multiple DAYS_BETWEEN constraints"+
					" with gap %d for a course (possibly with parallel"+
					" courses):\n  -- %s\n",
					gap, ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]))
				continue
			}
			gapmap[agix] = dbc
		}

		// Check for ineffective constraints and whether the default is needed
		dbclist := []*base.DaysBetween{}
		ng1dbc := hardgap == 0 // whether a constraint with gap = 1 is needed
		for gap, dbc := range gapmap {
			if gap < hardgap {
				base.Error.Printf("DAYS_BETWEEN constraint with gap smaller"+
					" than hard gap, course (or parallel course):\n -- %s\n",
					ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]))
				continue
			}
			if gap == 1 {
				ng1dbc = false
			}
			dbclist = append(dbclist, dbc)
		}
		if ng1dbc {
			dbclist = append(dbclist, &base.DaysBetween{
				Weight:               dgc.DefaultDifferentDaysWeight,
				ConsecutiveIfSameDay: dgc.DefaultDifferentDaysConsecutiveIfSameDay,
				DayGap:               1,
			})
		}

		// The constraints for ActivityGroup ag are in dbclist.
		// Add DaysBetweenConstraint items for the activity group
		for _, dbc := range dbclist {
			dbc0 := DaysBetweenConstraint{
				Weight:               dbc.Weight,
				DayGap:               dbc.DayGap,
				ConsecutiveIfSameDay: dbc.ConsecutiveIfSameDay,
			}
			dgc.CourseConstraints[agix] = append(
				dgc.CourseConstraints[agix], dbc0)
		}

		// Add the constraints to the lessons.
		for _, ttl := range ag.LessonUnits {
			// Collect days-between constraints
			dbllist := []DaysBetweenLessons{}
			lulist := []int{}
			for ttl1 := range ag.LessonUnits {
				if ttl1 != ttl {
					lulist = append(lulist, ttl1)
				}
			}
			for _, dbc := range dbclist {
				dbllist = append(dbllist, DaysBetweenLessons{
					Weight: dbc.Weight,
					// HoursGap assumes the use of an adequately long
					// DayLength, including buffer space at the end of the
					// real lesson slots
					HoursGap:             dbc.DayGap * ttinfo.DayLength,
					ConsecutiveIfSameDay: dbc.ConsecutiveIfSameDay,
					LessonUnits:          lulist,
				})
			}
			ttplaces.TtLessons[ttl].DaysBetween = dbllist
		}
	}
}

// processCrossDaysBetween collects cross-days-between on the same principle
// as the days-between constraints, but with the extra complication of there
// being two courses involved.
func (ttinfo *TtInfo) processCrossDaysBetween(
	agconstraints map[ActivityGroupIndex][]*base.DaysBetweenJoin,
) {
	dgc := ttinfo.DayGapConstraints
	dgc.CrossCourseConstraints = map[ActivityGroupIndex]map[ActivityGroupIndex][]DaysBetweenConstraint{}
	ttplaces := ttinfo.Placements
	for agix, dbclist0 := range agconstraints {
		ag := ttplaces.ActivityGroups[agix]

		// Sort according to "partner"
		agixmap := map[ActivityGroupIndex][]*base.DaysBetweenJoin{}
		for _, dbc := range dbclist0 {
			agix2 := ttplaces.CourseActivityGroup[dbc.Course2]
			agixmap[agix2] = append(agixmap[agix2], dbc)
		}
		// Handle each partner separately
		for agix2, dbclist0 := range agixmap {
			ag2 := ttplaces.ActivityGroups[agix2]
			// Sort according to gap
			hardgap := 0 // Record any hard constraint
			gapmap := map[int]*base.DaysBetweenJoin{}
			for _, dbc := range dbclist0 {
				gap := dbc.DayGap
				if dbc.Weight == base.MAXWEIGHT {
					// A hard constraint – there should be at most one
					if hardgap == 0 {
						hardgap = gap
					} else {
						base.Error.Printf("Multiple hard DAYS_BETWEEN_JOIN"+
							" constraints for courses (possibly with"+
							" parallel courses):\n -- %s\n -- %s\n",
							ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]),
							ttinfo.View(ttinfo.CourseInfo[ag2.Courses[0]]),
						)
						continue
					}
				}
				if _, nok := gapmap[agix]; nok {
					base.Error.Printf("Multiple DAYS_BETWEEN_JOIN"+
						" constraints with gap %d for courses (possibly"+
						" with parallel courses):\n  -- %s\n  -- %s\n",
						gap,
						ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]),
						ttinfo.View(ttinfo.CourseInfo[ag2.Courses[0]]),
					)
					continue
				}
				gapmap[agix] = dbc
			}

			// Check for ineffective constraints
			dbclist := []*base.DaysBetweenJoin{}
			//ng1dbc := hardgap == 0 // whether a constraint with gap = 1 is needed
			for gap, dbc := range gapmap {
				if gap < hardgap {
					base.Error.Printf("DAYS_BETWEEN_JOIN constraint with gap"+
						" smaller than hard gap, courses (or parallel"+
						" courses):\n -- %s\n -- %s\n",
						ttinfo.View(ttinfo.CourseInfo[ag.Courses[0]]),
						ttinfo.View(ttinfo.CourseInfo[ag2.Courses[0]]),
					)
					continue
				}
				dbclist = append(dbclist, dbc)
			}

			// Each constraint in dbclist concerns the activity groups ag and
			// ag2. Every lesson of each of the two activity groups needs a
			// [DaysBetweenLessons] constraint listing the
			// lessons of the other activity group.
			for _, dbc := range dbclist {
				// Collect the parameters
				gap := dbc.DayGap
				hgap := gap * ttinfo.DayLength
				weight := dbc.Weight
				cisd := dbc.ConsecutiveIfSameDay

				// Add a DaysBetweenConstraint for the activity group pair
				dbc0 := DaysBetweenConstraint{
					Weight:               weight,
					DayGap:               gap,
					ConsecutiveIfSameDay: cisd,
				}
				dgc.CrossCourseConstraints[agix][agix2] = append(
					dgc.CrossCourseConstraints[agix][agix2], dbc0)

				// Add constraints to the lessons of activity group ag
				for _, lix := range ag.LessonUnits {
					xl := ttplaces.TtLessons[lix]
					xl.DaysBetween = append(xl.DaysBetween,
						DaysBetweenLessons{
							Weight: weight,
							// HoursGap assumes the use of an adequately long
							// DayLength, including buffer space at the end of
							// the real lesson slots.
							HoursGap:             hgap,
							ConsecutiveIfSameDay: cisd,
							LessonUnits:          ag2.LessonUnits,
						})
				}
				// Add constraints to the lessons of activity group ag2
				for _, lix := range ag2.LessonUnits {
					xl := ttplaces.TtLessons[lix]
					xl.DaysBetween = append(xl.DaysBetween,
						DaysBetweenLessons{
							Weight: dbc.Weight,
							// HoursGap assumes the use of an adequately long
							// DayLength, including buffer space at the end of
							// the real lesson slots.
							HoursGap:             dbc.DayGap * ttinfo.DayLength,
							ConsecutiveIfSameDay: dbc.ConsecutiveIfSameDay,
							LessonUnits:          ag.LessonUnits,
						})
				}
			}
		}
	}
}
