package ttbase

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strings"
)

type CourseInfo struct {
	Id       Ref
	Subject  Ref
	Groups   []Ref
	Teachers []Ref
	Room     VirtualRoom
	Lessons  []ActivityIndex
}

// Make a shortish string view of a CourseInfo – can be useful in tests
func (ttinfo *TtInfo) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, ttinfo.Ref2Tag[t])
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		gx, ok := ttinfo.Ref2Tag[g]
		if !ok {
			base.Bug.Fatalf("No Ref2Tag for %s\n", g)
		}
		glist = append(glist, gx)
	}

	return fmt.Sprintf("<Course %s/%s:%s>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		ttinfo.Ref2Tag[cinfo.Subject],
	)
}

func gatherCourseInfo(ttinfo *TtInfo) {
	// Gather the Groups, Teachers and "rooms" for the Courses and
	// SuperCourses with lessons (only).
	// Gather the Lessons for these Courses and SuperCourses.
	// Also, the SuperCourses (with lessons) get a list of their
	// SubCourses.
	db := ttinfo.Db
	//--ttinfo.SuperSubs = map[Ref][]Ref{}
	ttinfo.CourseInfo = map[Ref]*CourseInfo{}
	ttinfo.Activities = make([]*Activity, 1) // 1-based indexing, 0 is invalid

	// Collect Courses with Lessons.
	roomData := collectCourses(ttinfo)

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

	// Create the CourseInfos and Activities.
	// Gather first the SuperCourses, then the Courses.

	cinfo_list := []*CourseInfo{}
	clessons := [][]Ref{}
	for _, spc := range db.SuperCourses {
		cref := spc.Id
		groups := []Ref{}
		teachers := []Ref{}
		rooms := []Ref{}
		for _, sbcref := range spc.SubCourses {
			sbc := db.Elements[sbcref].(*base.SubCourse)
			// Add groups
			if len(sbc.Groups) != 0 {
				groups = append(groups, sbc.Groups...)
			}
			// Add teachers
			if len(sbc.Teachers) != 0 {
				teachers = append(teachers, sbc.Teachers...)
			}
			// Add rooms
			if sbc.Room != "" {
				rooms = append(rooms, sbc.Room)
			}
		}
		// Eliminate duplicates by sorting and then compacting
		slices.Sort(groups)
		slices.Sort(teachers)
		slices.Sort(rooms)
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:       cref,
			Subject:  spc.Subject,
			Groups:   slices.Compact(groups),
			Teachers: slices.Compact(teachers),
			//Room: filled later
			Lessons: []ActivityIndex{},
		})
		clessons = append(clessons, spc.Lessons)
		roomData[cref] = slices.Compact(rooms)
	}
	for _, c := range db.Courses {
		cref := c.Id
		rooms := []Ref{}
		if c.Room != "" {
			rooms = append(rooms, c.Room)
		}
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:       cref,
			Subject:  c.Subject,
			Groups:   c.Groups,
			Teachers: c.Teachers,
			//Room: filled later
			Lessons: []ActivityIndex{},
		})
		clessons = append(clessons, c.Lessons)
		roomData[cref] = rooms
	}

	// Retain this ordered list of courses (with lessons)
	ttinfo.LessonCourses = cinfo_list

	for i, cinfo := range cinfo_list {
		// Add lessons to CourseInfo
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
				p = l.Day*ttinfo.NHours + l.Hour
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
	return roomData
}
