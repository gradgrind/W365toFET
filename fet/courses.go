package fet

/* TODO--

import (
	"W365toFET/base"
	"encoding/xml"
	"slices"
	"strconv"
)

func gatherCourseInfo(fetinfo *fetInfo) {
	// Gather the Groups, Teachers and "rooms" for the Courses and
	// SuperCourses with lessons (only).
	// Gather the Lessons for these Courses and SuperCourses.
	// Also, the SuperCourses (with lessons) get a list of their
	// SubCourses.
	db := fetinfo.db
	fetinfo.superSubs = make(map[Ref][]Ref)
	fetinfo.courseInfo = make(map[Ref]courseInfo)
	roomData := map[Ref][]Ref{} // course-Ref -> []room-Ref (any sort of "room")

	// Collect Courses with Lessons.
	for _, l := range db.Lessons {
		lcref := l.Course
		cinfo, ok := fetinfo.courseInfo[lcref]
		if ok {
			// If the course has already been handled, just add the lesson.
			cinfo.lessons = append(cinfo.lessons, l)
			fetinfo.courseInfo[lcref] = cinfo
			continue
		}
		// First encounter with the course.
		var subject Ref
		var groups []Ref
		var teachers []Ref
		var rooms []Ref
		lessons := []*base.Lesson{l}
		actids := []int{}

		c := db.Elements[lcref] // can be Course or SuperCourse
		cnode, ok := c.(*base.Course)
		if ok {
			if slices.Contains(l.Flags, "SubstitutionService") {
				groups = nil
			} else {
				groups = cnode.Groups
			}
			subject = cnode.Subject
			teachers = cnode.Teachers
			rooms = []Ref{}
			if cnode.Room != "" {
				rooms = append(rooms, cnode.Room)
			}
		} else {
			spc, ok := c.(*base.SuperCourse)
			if !ok {
				base.Error.Fatalf(
					"Invalid Course in Lesson %s:\n  %s\n",
					l.Id, lcref)
			}
			fetinfo.superSubs[lcref] = []Ref{}
			subject = spc.Subject
			groups = []Ref{}
			teachers = []Ref{}
			rooms = []Ref{}
		}
		fetinfo.courseInfo[lcref] = courseInfo{
			subject:  subject,
			groups:   groups,
			teachers: teachers,
			//rooms: filled later
			lessons:    lessons,
			activities: actids,
		}
		roomData[lcref] = rooms
	}

	// Now find the SubCourses, adding their groups, teachers and rooms.
	for _, sbc := range db.SubCourses {
		for _, spc := range sbc.SuperCourses {
			// Only fill SuperCourses which have Lessons
			cinfo, ok := fetinfo.courseInfo[spc]
			if ok {
				fetinfo.superSubs[spc] = append(fetinfo.superSubs[spc], sbc.Id)

				// Add groups
				if len(sbc.Groups) != 0 {
					cglist := append(cinfo.groups, sbc.Groups...)
					slices.Sort(cglist)
					cglist = slices.Compact(cglist)
					cinfo.groups = make([]Ref, len(cglist))
					copy(cinfo.groups, cglist)
				}

				// Add teachers
				if len(sbc.Teachers) != 0 {
					ctlist := append(cinfo.teachers, sbc.Teachers...)
					slices.Sort(ctlist)
					ctlist = slices.Compact(ctlist)
					cinfo.teachers = make([]Ref, len(ctlist))
					copy(cinfo.teachers, ctlist)
				}

				// Add rooms
				if sbc.Room != "" {
					crlist := append(roomData[spc], sbc.Room)
					slices.Sort(crlist)
					crlist = slices.Compact(crlist)
					roomData[spc] = crlist
				}

				fetinfo.courseInfo[spc] = cinfo
			}
		}
	}

	// Prepare the internal room structure, filtering the room lists of
	// the SuperCourses.
	for cref, crlist := range roomData {
		// Join all Rooms and the Rooms from RoomGroups into a "compulsory"
		// list. Then go through the RoomChoiceGroups. If one contains a
		// compulsory room, ignore the choice.
		// The result is a list of Rooms and a list of Room-choice-lists,
		// which can be converted into a fet virtual room.
		rooms := []Ref{}
		roomChoices := [][]Ref{}
		for _, rref := range crlist {
			rx := fetinfo.db.Elements[rref]
			_, ok := rx.(*base.Room)
			if ok {
				rooms = append(rooms, rref)
			} else {
				rg, ok := rx.(*base.RoomGroup)
				if ok {
					rooms = append(rooms, rg.Rooms...)
				} else {
					rc, ok := rx.(*base.RoomChoiceGroup)
					if !ok {
						base.Bug.Fatalf(
							"Invalid room in course %s:\n  %s\n",
							cref, rref)
					}
					roomChoices = append(roomChoices, rc.Rooms)
				}
			}
		}
		// Remove duplicates in Room list.
		slices.Sort(rooms)
		rooms = slices.Compact(rooms)
		// Filter choice lists.
		roomChoices = slices.DeleteFunc(roomChoices, func(rcl []Ref) bool {
			for _, rc := range rcl {
				if slices.Contains(rooms, rc) {
					return true
				}
			}
			return false
		})
		cinfo := fetinfo.courseInfo[cref]
		cinfo.room = virtualRoom{
			rooms:       rooms,
			roomChoices: roomChoices,
		}
		fetinfo.courseInfo[cref] = cinfo
	}
}
*/
