package ttbase

import "W365toFET/base"

// An ActivityGroup manages placement of the lessons of a course and
// any hard-parallel courses.
type ActivityGroup struct {
	Resources          []ResourceIndex
	LessonUnits        []*TtLesson //TODO: or []LessonUnitIndex?
	PossiblePlacements [][]SlotIndex
}

type TtLesson struct {
	Resources *[]ResourceIndex // points to ActivityGroup Resources?
	// ... if not dynamic, it could just be a "copy"
	Placement SlotIndex
	Fixed     bool
}

// PrepareActivityGroups creates the [ActivityGroup] items from the
// [Activity] items, taking their duration, courses, hard-parallel and
// hard-different-day constraints into account.
// TODO: Perhaps also soft days-between, etc.?
// Should I link to the Activities?
// To be able to handle unplacement I would need XRooms. Would accessing
// these via the Activities be too inefficient? Probably the inner loops
// should be handled in TtLesson as far as possible.
func (ttinfo *TtInfo) PrepareActivityGroups() []*ActivityGroup {
	aglist := []*ActivityGroup{}
	parallels := map[Ref]bool{} // handled parallel courses
	daygapmap := ttinfo.DayGapConstraints.CourseConstraints
	defaultDaysBetween := &base.DaysBetween{
		// The other fields are not needed here
		Weight: ttinfo.DayGapConstraints.DefaultDifferentDaysWeight,
		ConsecutiveIfSameDay: ttinfo.DayGapConstraints.
			DefaultDifferentDaysConsecutiveIfSameDay,
		DayGap: 1,
	}
	for _, cinfo := range ttinfo.LessonCourses {
		cref := cinfo.Id
		if _, nok := parallels[cref]; nok {
			// This course has already been handled
			continue
		}
		ag := &ActivityGroup{}
		aglist = append(aglist, ag)
		alist := []*Activity{}
		for _, aix := range cinfo.Lessons {
			alist = append(alist, ttinfo.Activities[aix])
		}
		// Seek hard-parallel courses
	repeat:
		ag.Resources = append([]ResourceIndex{}, cinfo.Resources...)

		for _, hpc := range ttinfo.HardParallelCourses[cinfo.Id] {
			//TODO: One (or more!) of the activities may have a placement.
			// This should be taken for the TtLesson – and checked for
			// conflicts

			//TODO: I suppose fixed activities should be excluded from the
			// possible placement lists?

			parallels[hpc] = true

			// Add resources from the parallel course
			pcinfo := ttinfo.CourseInfo[hpc]
			ag.Resources = append(ag.Resources, pcinfo.Resources...)

			// Get Activities
			update := false // flag whether the loop needs repeating
			for i, aix := range pcinfo.Lessons {
				a := ttinfo.Activities[aix]
				a0 := alist[i]
				if a.Fixed {
					if !a0.Fixed {
						a0.Fixed = true
						update = true
					}
				} else {
					if a0.Fixed {
						a.Fixed = true
					}
				}
				if a.Placement < 0 {
					if a0.Placement >= 0 {
						a.Placement = a0.Placement
					}
				} else if a0.Placement < 0 {
					a0.Placement = a.Placement
					update = true
				} else if a.Placement != a0.Placement {
					base.Error.Printf("Parallel Activities with"+
						"different placements in courses\n"+
						" -- %s\n -- %s\n",
						ttinfo.View(cinfo), ttinfo.View(pcinfo))
					if update {
						// a0 was previously unfixed, a fixed
						a0.Placement = a.Placement
					} else {
						a.Placement = a0.Placement
					}
				}
			}
			if update {
				// The activities of the first course have been changed,
				// the processing of parallel courses should be repeated
				goto repeat
			}
		}

		// "DAYS_BETWEEN" constraints:
		// 1) Collect days-between according to gaps, reporting and ignoring
		//    conflicts.
		// 2) If there is no constraint for gap = 1, and no other hard
		//    constraint, use the default.
		//TODO:
		// 3) Collect cross-days-between on the same principle, but with the
		//    extra complication of there being two courses involved.

		dbcmap := map[int]*base.DaysBetween{}
		harddbc := 0
		// Include hard-parallel courses
		pcourses := append([]Ref{cref}, ttinfo.HardParallelCourses[cref]...)
		for _, pcref := range pcourses {
			for _, dbc := range daygapmap[pcref] {
				gap := dbc.DayGap
				if dbc.Weight == base.MAXWEIGHT {
					// A hard constraint
					if harddbc == 0 {
						harddbc = gap
					} else {
						base.Error.Printf("Multiple hard DAYS_BETWEEN"+
							" constraints for a course (possibly with"+
							" parallel courses):\n -- %s\n",
							ttinfo.View(cinfo))
						continue
					}
				}
				if _, nok := dbcmap[gap]; nok {
					base.Error.Printf("Multiple DAYS_BETWEEN constraints"+
						" with gap %d for a course (possibly with parallel"+
						" courses):\n  -- %s\n",
						gap, ttinfo.View(cinfo))
					continue
				}
				dbcmap[gap] = dbc
			}
		}
		dbclist := []*base.DaysBetween{}
		ng1dbc := harddbc == 0 // whether a constraint with gap = 1 is needed
		for gap, dbc := range dbcmap {
			if gap < harddbc {
				base.Error.Printf("DAYS_BETWEEN constraint with gap smaller"+
					" than hard gap, course:\n -- %s\n",
					ttinfo.View(cinfo))
				continue
			}
			if gap == 1 {
				ng1dbc = false
			}
			dbclist = append(dbclist, dbc)
		}
		if ng1dbc {
			dbclist = append(dbclist, defaultDaysBetween)
		}

	}

	//TODO: Add the TtLessons

	return aglist
}
