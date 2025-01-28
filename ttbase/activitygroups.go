package ttbase

import (
	"W365toFET/base"
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

// TODO: Should I link to the Activities? Maybe the Activities should link
// to their TtLessons, possibly via the ActivityGroups?

// TODO: To be able to handle unplacement I would need XRooms. Would accessing
// these via the Activities be too inefficient? Probably the inner loops
// should be handled in TtLesson as far as possible. On the other hand,
// unplacing would take place far less frequently than testing.

// PrepareActivityGroups creates the [ActivityGroup] items from the
// courses listed in [TtInfo.LessonCourses].
// TODO: This will need intensive testing!
func (ttinfo *TtInfo) PrepareActivityGroups() {
	// The handling of the first encountered course of a hard-parallel set
	// will also deal with the other courses in the set.
	ttplaces := &TtPlacement{
		ActivityGroups:      []*ActivityGroup{},
		CourseActivityGroup: map[Ref]ActivityGroupIndex{},
		// The first entry of TtLessons is invalid, allowing 0 to be used
		// as a null index.
		TtLessons: []*TtLesson{nil},
	}
	ttinfo.Placements = ttplaces
	for _, cinfo := range ttinfo.LessonCourses {
		cref := cinfo.Id
		if _, nok := ttplaces.CourseActivityGroup[cref]; nok {
			// This course has already been handled, meaning that it is
			// hard-parallel to an earlier one.
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
		// Gather hard-parallel courses and all their resources.
		// The process will need to be restarted from scratch when fields
		// of a parallel course override those of the first course: Fixed
		// and Placement fields should be the same in the Activity items of
		// all hard-parallel courses.
	restart:
		pcourses := []Ref{cref} // list of parallel courses
		resources := append([]ResourceIndex{}, cinfo.Resources...)

		for _, hpc := range ttinfo.HardParallelCourses[cinfo.Id] {
			// One (or more!) of the activities may have a placement.
			// This should be taken for the TtLesson – and checked for
			// conflicts.

			// Reference ActivityGroup, also mark this course as "handled"
			ttplaces.CourseActivityGroup[hpc] = agindex

			// Add the parallel course and its resources
			pcourses = append(pcourses, hpc)
			pcinfo := ttinfo.CourseInfo[hpc]
			resources = append(resources, pcinfo.Resources...)

			// Get Activities – these should have the same number and duration
			// in every parallel course. That should have been checked in
			// [addParallelCoursesConstraint], called via [processConstraints]
			// from [PrepareCoreData]. Here it is assumed correct!
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
					// Patch the data to avoid conflict
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

		// Add the TtLessons
		ttli0 := len(ttplaces.TtLessons)
		for i, a := range alist {
			ttl := &TtLesson{
				Resources: ag.Resources,
				Placement: a.Placement,
				Fixed:     a.Fixed,
				//DaysBetween: TODO
			}
			ttplaces.TtLessons = append(ttplaces.TtLessons, ttl)
			ag.LessonUnits = append(ag.LessonUnits, ttli0+i)
		}
	}
}

// TODO
func (ttinfo *TtInfo) processDaysBetweenConstraints() {
	// Sort the constraints according to their activity groups.
	agconstraints := map[ActivityGroupIndex][]*base.DaysBetween{}
	for _, c := range ttinfo.Constraints["DAYS_BETWEEN"] {
		dbc := c.(*base.DaysBetween)
		for _, cref := range dbc.Courses {
			agix := ttinfo.Placements.CourseActivityGroup[cref]
			agconstraints[agix] = append(agconstraints[agix], dbc)
		}
	}
	//TODO ...

	// TODO ...
	//
	//	ttinfo.Constraints["DAYS_BETWEEN_JOIN"]
}

/*TODO

	// Process DAYS_BETWEEN constraints (including automatic ones)
	dbllist := ttinfo.processDaysBetween(ag)

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

    // Process DAYS_BETWEEN_JOIN constraints
	ttinfo.processCrossDaysBetween()

*/
