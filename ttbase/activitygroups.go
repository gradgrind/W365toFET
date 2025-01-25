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
	for _, cinfo := range ttinfo.LessonCourses {
		if _, nok := parallels[cinfo.Id]; nok {
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
			for i, aix := range pcinfo.Lessons {
				a := ttinfo.Activities[aix]
				a0 := alist[i]
				update := false
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
				if update {
					goto repeat
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

	return aglist
}
