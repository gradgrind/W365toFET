package ttbase

import (
	"W365toFET/base"
	"slices"
)

type LessonIndex int // index to TtLessons vector

type TtLesson struct {
	Index LessonIndex
	//? Duration int
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
}

type CourseInfo struct {
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     VirtualRoom
	Lessons  []LessonIndex
}

func gatherCourseInfo(ttinfo *TtInfo) {
	// Gather the Groups, Teachers and "rooms" for the Courses and
	// SuperCourses with lessons (only).
	// Gather the Lessons for these Courses and SuperCourses.
	// Also, the SuperCourses (with lessons) get a list of their
	// SubCourses.
	db := ttinfo.Db
	ttinfo.SuperSubs = map[Ref][]Ref{}
	ttinfo.CourseInfo = map[Ref]*CourseInfo{}
	ttinfo.TtLessons = []TtLesson{}

	// Collect Courses with Lessons.
	roomData := collectCourses(ttinfo)

	// Now find the SubCourses, adding their groups, teachers and rooms.
	findSubCourses(ttinfo, roomData)

	// Prepare the internal room structure, filtering the room lists of
	// the SuperCourses.
	for cref, crlist := range roomData {
		// Join all Rooms and the Rooms from RoomGroups into a "compulsory"
		// list. Then go through the RoomChoiceGroups. If one contains a
		// compulsory room, ignore the choice.
		// The result is a list of Rooms and a list of Room-choice-lists,
		// which can be converted into a tt virtual room.
		rooms := []Ref{}
		roomChoices := [][]Ref{}
		for _, rref := range crlist {
			rx := db.Elements[rref]
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

		// Add virtual room to CourseInfo item
		ttinfo.CourseInfo[cref].Room = VirtualRoom{
			Rooms:       rooms,
			RoomChoices: roomChoices,
		}
	}
}

func collectCourses(ttinfo *TtInfo) map[Ref][]Ref {
	// Collect Courses with Lessons.
	roomData := map[Ref][]Ref{} // course -> []room (any sort of "room")
	db := ttinfo.Db
	for _, l := range db.Lessons {
		// Make a new TtLesson (CourseInfo will be added later)
		ttlix := LessonIndex(len(ttinfo.TtLessons))
		ttl := TtLesson{
			Index:  ttlix,
			Lesson: l,
		}
		ttinfo.TtLessons = append(ttinfo.TtLessons, ttl)

		lcref := l.Course
		cinfo, ok := ttinfo.CourseInfo[lcref]
		if ok {
			// If the course has already been handled, just add the lesson.
			cinfo.Lessons = append(cinfo.Lessons, ttlix)
			ttinfo.CourseInfo[lcref] = cinfo
			continue
		}

		// First encounter with the course.
		var subject Ref
		var groups []Ref
		var teachers []Ref
		var rooms []Ref

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
			ttinfo.SuperSubs[lcref] = []Ref{}
			subject = spc.Subject
			groups = []Ref{}
			teachers = []Ref{}
			rooms = []Ref{}
		}
		ttinfo.CourseInfo[lcref] = &CourseInfo{
			Subject:  subject,
			Groups:   groups,
			Teachers: teachers,
			//Rooms: filled later
			Lessons: []LessonIndex{ttlix},
		}
		roomData[lcref] = rooms
	}
	return roomData
}

func findSubCourses(ttinfo *TtInfo, roomData map[Ref][]Ref) {
	// Find the SubCourses for each lesson-containing SuperCourse.
	for _, sbc := range ttinfo.Db.SubCourses {
		for _, spc := range sbc.SuperCourses {
			// Only fill SuperCourses which have Lessons
			cinfo, ok := ttinfo.CourseInfo[spc]
			if ok {
				ttinfo.SuperSubs[spc] = append(ttinfo.SuperSubs[spc], sbc.Id)

				// Add groups
				if len(sbc.Groups) != 0 {
					cglist := append(cinfo.Groups, sbc.Groups...)
					slices.Sort(cglist)
					cglist = slices.Compact(cglist)
					cinfo.Groups = make([]Ref, len(cglist))
					copy(cinfo.Groups, cglist)
				}

				// Add teachers
				if len(sbc.Teachers) != 0 {
					ctlist := append(cinfo.Teachers, sbc.Teachers...)
					slices.Sort(ctlist)
					ctlist = slices.Compact(ctlist)
					cinfo.Teachers = make([]Ref, len(ctlist))
					copy(cinfo.Teachers, ctlist)
				}

				// Add rooms
				if sbc.Room != "" {
					crlist := append(roomData[spc], sbc.Room)
					slices.Sort(crlist)
					crlist = slices.Compact(crlist)
					roomData[spc] = crlist
				}

				ttinfo.CourseInfo[spc] = cinfo
			}
		}
	}
}
