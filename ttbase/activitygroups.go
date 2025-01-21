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
}

// PrepareActivityGroups creates the [ActivityGroup] items from the
// [Activity] items, taking their duration, courses, hard-parallel and
// hard-different-day constraints into account.
func (ttinfo *TtInfo) PrepareActivityGroups() {

	for _, cinfo := range ttinfo.LessonCourses {

		// Seek hard-parallel courses
		for _, hpc := range ttinfo.ParallelCourses[cinfo.Id] {
			//TODO

			fmt.Printf("??? %+v\n", hpc)

			if hpc.Weight == base.MAXWEIGHT {
				//TODO: These courses are hard-parallel, join them into
				// a single activity group.

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
