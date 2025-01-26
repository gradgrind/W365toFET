package ttbase

import "W365toFET/base"

type LessonUnitIndex = int

// An ActivityGroup manages placement of the lessons of a course and
// any hard-parallel courses.
type ActivityGroup struct {
	Resources          []ResourceIndex
	LessonUnits        []LessonUnitIndex
	PossiblePlacements [][]SlotIndex
}

type TtLesson struct {
	Resources []ResourceIndex // same as ActivityGroup Resources
	// ... if not dynamic, it could just be a "copy"
	Placement SlotIndex
	Fixed     bool
}

type TtPlacement struct {
	TtLessons []*TtLesson
}

// PrepareActivityGroups creates the [ActivityGroup] items from the
// [Activity] items, taking their duration, courses, hard-parallel and
// different-day/days-between constraints into account.

// TODO: Should I link to the Activities? Maybe the Activities should link
// to their TtLessons, possibly via the ActivityGroups?

// TODO: To be able to handle unplacement I would need XRooms. Would accessing
// these via the Activities be too inefficient? Probably the inner loops
// should be handled in TtLesson as far as possible. On the other hand,
// unplacing would take place far less frequently than testing.

func (ttinfo *TtInfo) PrepareActivityGroups() []*ActivityGroup {
	aglist := []*ActivityGroup{}
	parallels := map[Ref]bool{} // handled parallel courses
	dgc := ttinfo.DayGapConstraints
	daygapmap := dgc.CourseConstraints
	defaultDaysBetween := DaysBetweenLessons{
		Weight:               dgc.DefaultDifferentDaysWeight,
		ConsecutiveIfSameDay: dgc.DefaultDifferentDaysConsecutiveIfSameDay,
		DayGap:               1,
	}
	ttplaces := &TtPlacement{
		// The first entry of TtLessons is invalid
		TtLessons: []*TtLesson{nil},
	}
	ttinfo.Placements = ttplaces
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

		dblmap := map[int]DaysBetweenLessons{}
		harddbc := 0
		// Include hard-parallel courses
		pcourses := append([]Ref{cref}, ttinfo.HardParallelCourses[cref]...)
		for _, pcref := range pcourses {
			for _, dbl := range daygapmap[pcref] {
				gap := dbl.DayGap
				if dbl.Weight == base.MAXWEIGHT {
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
				if _, nok := dblmap[gap]; nok {
					base.Error.Printf("Multiple DAYS_BETWEEN constraints"+
						" with gap %d for a course (possibly with parallel"+
						" courses):\n  -- %s\n",
						gap, ttinfo.View(cinfo))
					continue
				}
				dblmap[gap] = dbl
			}
		}
		// Check for ineffective constraints and whether the default is needed
		dbllist := []DaysBetweenLessons{}
		ng1dbc := harddbc == 0 // whether a constraint with gap = 1 is needed
		for gap, dbl := range dblmap {
			if gap < harddbc {
				base.Error.Printf("DAYS_BETWEEN constraint with gap smaller"+
					" than hard gap, course:\n -- %s\n",
					ttinfo.View(cinfo))
				continue
			}
			if gap == 1 {
				ng1dbc = false
			}
			dbllist = append(dbllist, dbl)
		}
		if ng1dbc {
			dbllist = append(dbllist, defaultDaysBetween)
		}

		// Add the TtLessons
		//TODO: Do I need a TtLesson list? Probably ...
		// ... maybe in package ttplacement
		//TODO: Add links ...
		for _, a := range alist {
			ttl := &TtLesson{
				Resources: ag.Resources,
				Placement: a.Placement,
				Fixed:     a.Fixed,
			}
			i := len(ttplaces.TtLessons)
			ttplaces.TtLessons = append(ttplaces.TtLessons, ttl)
			ag.LessonUnits = append(ag.LessonUnits, i)
		}

		//TODO: Generate the possible placements lists?

	}

	return aglist
}
