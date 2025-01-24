package ttbase

import (
	"W365toFET/base"
	"fmt"
)

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
// The room choices should perhaps be handled like SuperCourses?
// Should I link to the Activities?
// To be able to handle unplacement I would need XRooms. Would accessing
// these via the Activities be too inefficient? Probably the inner loops
// should be handled in TtLesson as far as possible.
func (ttinfo *TtInfo) PrepareActivityGroups() {
	parallels := map[Ref]bool{} // handled parallel courses
	for _, cinfo := range ttinfo.LessonCourses {
		ag := &ActivityGroup{}
		// Seek hard-parallel courses
		for _, hpc := range ttinfo.ParallelCourses[cinfo.Id] {
			//TODO: One (or more!) of the activities may have a placement.
			// This should be taken for the TtLesson

			fmt.Printf("??? %+v\n", hpc)

			if hpc.Weight == base.MAXWEIGHT {
				//TODO: These courses are hard-parallel, join them into
				// a single activity group.
				for _, c := range hpc.Courses {
					parallels[c] = true

					// Add resources from the parallel course
					cinfo1 := ttinfo.CourseInfo[c]
					// ... via one of the Activities
					l0 := cinfo1.Lessons[0]
					a := ttinfo.Activities[l0]
					ag.Resources = append(ag.Resources, a.Resources...)
				}

			}
		}

	}

	/*
			// Add activities to CourseInfo
			llist := clessons[i]
			for _, lref := range llist {
				l := db.Elements[lref].(*base.Lesson)
				if slices.Contains(l.Flags, "SubstitutionService") {
					cinfo.Groups = nil
				}
				// Index of new Activity:
				ttlix := len(ttinfo.Activities)
				p := -1
				if l.Day >= 0 {
					p = l.Day*ttinfo.DayLength + l.Hour
				}
				ttl := &Activity{
					Index:      ttlix,
					Placement:  p,
					Duration:   l.Duration,
					Fixed:      l.Fixed,
					Lesson:     l,
					CourseInfo: cinfo,
				}
				ttinfo.Activities = append(ttinfo.Activities, ttl)
				cinfo.Lessons = append(cinfo.Lessons, ttlix)
			}

			// Add to CourseInfo map
			ttinfo.CourseInfo[cinfo.Id] = cinfo
		}
		//return roomData
	*/
}
