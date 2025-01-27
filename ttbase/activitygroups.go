package ttbase

import (
	"W365toFET/base"
	"slices"
)

type ActivityGroupIndex = int
type LessonUnitIndex = int

// An ActivityGroup manages placement of the lessons of a course and
// any hard-parallel courses by combining their resources and activities.
type ActivityGroup struct {
	Resources   []ResourceIndex
	Courses     []Ref
	LessonUnits []LessonUnitIndex

	//TODO--? PossiblePlacements [][]SlotIndex
}

type DaysBetweenConstraint struct {
	Weight               int
	HoursGap             int
	ConsecutiveIfSameDay bool
	LessonUnits          []LessonUnitIndex
}

type TtLesson struct {
	Resources []ResourceIndex // same as ActivityGroup Resources,
	// it is not dynamic so it is just be a "copy" of the ActivityGroup field
	Placement   SlotIndex
	Fixed       bool
	DaysBetween []DaysBetweenConstraint
}

type TtPlacement struct {
	ActivityGroups []*ActivityGroup
	// CourseActivityGroup maps a course to the index of its ActivityGroup
	CourseActivityGroup map[Ref]ActivityGroupIndex
	TtLessons           []*TtLesson
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

func (ttinfo *TtInfo) PrepareActivityGroups() {
	// The handling of the first encountered course of a hard-parallel set
	// will also deal with the other courses in the set.
	dgc := ttinfo.DayGapConstraints
	daygapmap := dgc.CourseConstraints
	xdaygapmap := dgc.CrossCourseConstraints
	defaultDaysBetween := DaysBetweenLessons{
		Weight:               dgc.DefaultDifferentDaysWeight,
		ConsecutiveIfSameDay: dgc.DefaultDifferentDaysConsecutiveIfSameDay,
		DayGap:               1,
	}
	ttplaces := &TtPlacement{
		ActivityGroups:      []*ActivityGroup{},
		CourseActivityGroup: map[Ref]ActivityGroupIndex{},
		// The first entry of TtLessons is invalid
		TtLessons: []*TtLesson{nil},
	}
	ttinfo.Placements = ttplaces
	for _, cinfo := range ttinfo.LessonCourses {
		cref := cinfo.Id
		if _, nok := ttplaces.CourseActivityGroup[cref]; nok {
			// This course has already been handled
			continue
		}
		agindex := len(ttplaces.ActivityGroups)
		// Reference ActivityGroup, also mark this course as "handled"
		ttplaces.CourseActivityGroup[cref] = agindex
		// Gather the course's Activity elements for comparison with those of
		// the parallel courses
		alist := []*Activity{}
		for _, aix := range cinfo.Lessons {
			alist = append(alist, ttinfo.Activities[aix])
		}
		// Seek hard-parallel courses. The process will need to be restarted
		// from scratch when fields of a parallel course override those of the
		// first course: Fixed and Placement fields should be the same in the
		// Activity items of all hard-parallel courses.
	restart:
		// Gather hard-parallel courses and all their resources
		pcourses := []Ref{cref}
		resources := append([]ResourceIndex{}, cinfo.Resources...)

		for _, hpc := range ttinfo.HardParallelCourses[cinfo.Id] {
			// One (or more!) of the activities may have a placement.
			// This should be taken for the TtLesson – and checked for
			// conflicts

			// Reference ActivityGroup, also mark this course as "handled"
			ttplaces.CourseActivityGroup[hpc] = agindex

			// Add the parallel course and its resources
			pcourses = append(pcourses, hpc)
			pcinfo := ttinfo.CourseInfo[hpc]
			resources = append(resources, pcinfo.Resources...)

			// Get Activities – these should have the same number and duration
			// in every parallel course, which should have been checked in
			// [addParallelCoursesConstraint], called via [processConstraints]
			// from [PrepareCoreData].
			update := false // flag whether the loop needs restarting
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
				goto restart
			}
		}

		// Create the ActivityGroup and add it to the list
		ag := &ActivityGroup{
			Resources:   resources,
			Courses:     pcourses,
			LessonUnits: []LessonUnitIndex{},
		}
		ttplaces.ActivityGroups = append(ttplaces.ActivityGroups, ag)

		// "DAYS_BETWEEN" constraints:
		// 1) Collect days-between constraints according to gaps, reporting
		// and ignoring conflicts.
		// 2) If there is no constraint for gap = 1, and no other hard
		//    constraint, use the default.
		// Later, because all course must have ActivityGroups for this:
		// 3) Collect cross-days-between on the same principle, but with the
		//    extra complication of there being two courses involved.

		dblmap := map[int]DaysBetweenLessons{}
		harddbc := 0 // Record any such hard constraint
		// Include hard-parallel courses
		for _, pcref := range pcourses {
			for _, dbl := range daygapmap[pcref] {
				gap := dbl.DayGap
				if dbl.Weight == base.MAXWEIGHT {
					// A hard constraint – there should be at most one
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

		ttli0 := len(ttplaces.TtLessons)
		for i, a := range alist {

			// Collect days-between constraints
			dbclist := []DaysBetweenConstraint{}
			lulist := []int{}
			for j := range alist {
				if j != i {
					lulist = append(lulist, ttli0+j)
				}
			}
			for _, dbl := range dbllist {
				dbclist = append(dbclist, DaysBetweenConstraint{
					Weight: dbl.Weight,
					// HoursGap assumes the use of an adequately long
					// DayLength, including buffer space at the end of the
					// real lesson slots
					HoursGap:             dbl.DayGap * ttinfo.DayLength,
					ConsecutiveIfSameDay: dbl.ConsecutiveIfSameDay,
					LessonUnits:          lulist,
				})
			}

			ttl := &TtLesson{
				Resources:   ag.Resources,
				Placement:   a.Placement,
				Fixed:       a.Fixed,
				DaysBetween: dbclist,
			}
			ttplaces.TtLessons = append(ttplaces.TtLessons, ttl)
			ag.LessonUnits = append(ag.LessonUnits, ttli0+i)
		}
	}

	// The cross-course days-between must be done after all ActivityGroup
	// elements have been generated, so that their TtLesson indexes are
	// available.

	//xdaygapmap map[Ref][]CrossDaysBetweenLessons
	//xdblmap := map[int]DaysBetweenXLessons{}

	for _, ag := range ttplaces.ActivityGroups {
		//xdbllist := []CrossDaysBetweenLessons{}
		// First collect the constraints for each cross-course

		xdblmap := map[Ref][]CrossDaysBetweenLessons{}
		for _, pcref := range ag.Courses {

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

				xdblmap[xcourse] = append(xdblmap[xcourse], xdbl)
			}

			for xcourse, xdbllist := range xdblmap {
				xdblmap := map[int]CrossDaysBetweenLessons{}
				harddbc := 0
				for _, xdbl := range xdbllist {
					gap := xdbl.DayGap
					if xdbl.Weight == base.MAXWEIGHT {
						// A hard constraint
						if harddbc == 0 {
							harddbc = gap
						} else {
							base.Error.Printf("Multiple hard DAYS_BETWEEN_JOIN"+
								" constraints for a course (possibly with"+
								" parallel courses):\n -- %s\n  -- %s\n",
								ttinfo.View(ttinfo.CourseInfo[pcref]),
								ttinfo.View(ttinfo.CourseInfo[xcourse]),
							)
							continue
						}
					}
					if _, nok := xdblmap[gap]; nok {
						base.Error.Printf("Multiple DAYS_BETWEEN_JOIN"+
							" constraints with gap %d for a course (possibly"+
							" with parallel courses):\n  -- %s\n  -- %s\n",
							gap,
							ttinfo.View(ttinfo.CourseInfo[pcref]),
							ttinfo.View(ttinfo.CourseInfo[xcourse]),
						)
						continue
					}
					xdblmap[gap] = xdbl
				}
			}
		}

		//TODO: something like ...

		for i, a := range alist {

			// Collect days-between constraints
			dbclist := []DaysBetweenConstraint{}
			lulist := []int{}
			for j := range alist {
				if j != i {
					lulist = append(lulist, ttli0+j)
				}
			}
			for _, dbl := range dbllist {
				dbclist = append(dbclist, DaysBetweenConstraint{
					Weight: dbl.Weight,
					// HoursGap assumes the use of an adequately long
					// DayLength, including buffer space at the end of the
					// real lesson slots
					HoursGap:             dbl.DayGap * ttinfo.DayLength,
					ConsecutiveIfSameDay: dbl.ConsecutiveIfSameDay,
					LessonUnits:          lulist,
				})
			}

		}

	}

}
