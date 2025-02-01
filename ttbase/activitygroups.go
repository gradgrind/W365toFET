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
}

type TtLesson struct {
	Resources []ResourceIndex // same as ActivityGroup Resources,
	// it is not dynamic so it is just be a "copy" of the ActivityGroup field
	Placement   SlotIndex
	Fixed       bool
	DaysBetween []DaysBetweenLessons
	// XRooms is a list of chosen rooms, distinct from the required rooms
	XRooms        []ResourceIndex
	PossibleSlots []SlotIndex
}

type TtPlacement struct {
	ActivityGroups []*ActivityGroup
	// CourseActivityGroup maps a course to the index of its ActivityGroup
	CourseActivityGroup map[Ref]ActivityGroupIndex
	TtLessons           []*TtLesson
	TtSlots             []LessonUnitIndex
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
	// Initialize th [TtPlacement] structure
	nResources := len(ttinfo.Resources)
	ttplaces := &TtPlacement{
		ActivityGroups:      []*ActivityGroup{},
		CourseActivityGroup: map[Ref]ActivityGroupIndex{},
		// The first entry of TtLessons is invalid, allowing 0 to be used
		// as a null index.
		TtLessons: []*TtLesson{nil},
		// Each resource has a vector of time-slots covering the whole week
		TtSlots: make([]LessonUnitIndex, nResources*ttinfo.SlotsPerWeek),
	}
	ttinfo.Placements = ttplaces

	// Add pseudo-lessons to block certain slots:
	ttinfo.blockPadding() // block the extra slots at the end of each day
	ttinfo.addBlockers()  // block the resources' not-available slots

	// Build activity groups from the [CouurseInfo] items.
	// The handling of the first encountered course of a hard-parallel set
	// will also deal with the other courses in the set.
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
		// Gather the course's lessons for comparison with those of
		// the parallel courses
		llist0 := cinfo.Lessons
		// Gather hard-parallel courses and all their resources.
		// The process will need to be restarted from scratch when fields
		// of a parallel course override those of the first course: Fixed
		// and Placement fields should be the same in the lessons of
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

			// Get lessons – these should have the same number and duration
			// in every parallel course. That should have been checked in
			// [addParallelCoursesConstraint], called via [processConstraints]
			// from [MakeTtInfo]. Here it is assumed correct!
			update := false // flag whether the loop needs restarting
			for i, l := range pcinfo.Lessons {
				l0 := llist0[i]
				if l.Fixed {
					if !l0.Fixed {
						l0.Fixed = true
						update = true
					}
				} else {
					if l0.Fixed {
						l.Fixed = true
					}
				}
				if l.Day < 0 {
					if l0.Day >= 0 {
						l.Day = l0.Day
						l.Hour = l0.Hour
					}
				} else if l0.Day < 0 {
					l0.Day = l.Day
					l0.Hour = l.Hour
					update = true
				} else if l.Day != l0.Day || l.Hour != l0.Hour {
					base.Error.Printf("Parallel Activities with"+
						"different placements in courses\n"+
						" -- %s\n -- %s\n",
						ttinfo.View(cinfo), ttinfo.View(pcinfo))
					// Patch the data to avoid conflict
					if update {
						// a0 was previously unfixed, a fixed
						l0.Day = l.Day
						l0.Hour = l.Hour
					} else {
						l.Day = l0.Day
						l.Hour = l0.Hour
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
		for i, l := range llist0 {
			p := -1 // placement slot
			if l.Day >= 0 {
				p = l.Day * ttinfo.DayLength * l.Hour
			}
			ttl := &TtLesson{
				Resources: ag.Resources,
				Placement: p,
				Fixed:     l.Fixed,
				//DaysBetween: will be added later
			}
			ttplaces.TtLessons = append(ttplaces.TtLessons, ttl)
			ag.LessonUnits = append(ag.LessonUnits, ttli0+i)
		}
	}
}
