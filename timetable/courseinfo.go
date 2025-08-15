package timetable

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strings"
)

// TODO: Activities should (already) be ordered, highest duration first,
// and ActivityGroup should have the same order.
// A CourseInfo is an intermediate representation of a course (Course or
// SuperCourse) for the timetable.
type CourseInfo struct {
	Id            NodeRef
	Subject       NodeRef
	Groups        []NodeRef
	Teachers      []NodeRef
	Rooms         []NodeRef // items can be Room, RoomGroup or RoomChoiceGroup
	Activities    []NodeRef
	ActivityGroup []ActivityIndex
}

// Make a shortish string view of a CourseInfo – can be useful in tests
func View(db *base.DbTopLevel, cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, db.Ref2Tag(t))
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, db.Ref2Tag(g))
	}
	return fmt.Sprintf("<Course %s/%s:%s>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		db.Ref2Tag(cinfo.Subject),
	)
}

// Collect courses (Course and SuperCourse) and their activities.
// Build a list of CourseInfo structures.
func CollectCourses(db *base.DbTopLevel) []*CourseInfo {
	cinfo_list := []*CourseInfo{}

	// Gather the SuperCourses.
	for _, spc := range db.SuperCourses {
		cref := spc.Id
		groups := []NodeRef{}
		teachers := []NodeRef{}
		rooms := []NodeRef{}
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
		// Eliminate duplicate resources by sorting and then compacting
		slices.Sort(groups)
		slices.Sort(teachers)
		slices.Sort(rooms)
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:         cref,
			Subject:    spc.Subject,
			Groups:     slices.Compact(groups),
			Teachers:   slices.Compact(teachers),
			Rooms:      slices.Compact(rooms),
			Activities: spc.Lessons,
		})
	}

	// Gather the plain Courses.
	for _, c := range db.Courses {
		cref := c.Id
		rooms := []NodeRef{}
		if c.Room != "" {
			rooms = append(rooms, c.Room)
		}
		cinfo_list = append(cinfo_list, &CourseInfo{
			Id:         cref,
			Subject:    c.Subject,
			Groups:     c.Groups,
			Teachers:   c.Teachers,
			Rooms:      rooms,
			Activities: c.Lessons,
		})
	}

	return cinfo_list
}

// `MakeActivities` creates the `TtActivity` structures ...
func (tt_data *TtData) MakeActivities(db *base.DbTopLevel, cinfo_list []*CourseInfo) {
	tt_data.Activities = []*TtActivity{{}} // first entry is empty
	for _, cinfo := range cinfo_list {
		// Get resource indexes
		resources := []ResourceIndex{}
		for _, r := range cinfo.Groups {
			agilist, ok := tt_data.AtomicGroups[r]
			if !ok {
				base.Bug.Fatalf("Unknown group: %s", r)
			}
			resources = append(resources, agilist...)
		}
		for _, r := range cinfo.Teachers {
			agi, ok := tt_data.TeacherIndex[r]
			if !ok {
				base.Bug.Fatalf("Unknown teacher: %s", r)
			}
			resources = append(resources, agi)
		}
		roomchoices := [][]ResourceIndex{}
		for _, r := range cinfo.Rooms {
			agi, ok := tt_data.RoomIndex[r]
			if ok {
				resources = append(resources, agi)
				continue
			}
			// Not a Room – it can be a RoomGroup or RoomChoiceGroup
			elem, ok := db.Elements[r]
			if !ok {
				base.Bug.Fatalf("Unknown room: %s", r)
			}
			rg, ok := elem.(*base.RoomGroup)
			if ok {
				for _, rr := range rg.Rooms {
					agi, ok = tt_data.RoomIndex[rr]
					if !ok {
						base.Bug.Fatalf(
							"Unknown room in RoomGroup %s: %s", r, rr)
					}
					resources = append(resources, agi)
				}
				continue
			}
			rcg, ok := elem.(*base.RoomChoiceGroup)
			if ok {
				rooms := []ResourceIndex{}
				for _, rr := range rcg.Rooms {
					agi, ok = tt_data.RoomIndex[rr]
					if !ok {
						base.Bug.Fatalf(
							"Unknown room in RoomChoiceGroup %s: %s", r, rr)
					}
					rooms = append(rooms, agi)
				}
				roomchoices = append(roomchoices, rooms)
				continue
			}
			base.Bug.Fatalf("Expecting room element, found: %s", r)
		}
		// Check for duplicate resources
		//TODO: Is this a bug, or can it happen "legally"?
		// If the latter, duplicates would need to be removed.
		rset := map[ResourceIndex]bool{}
		for _, ri := range resources {
			if rset[ri] {
				base.Bug.Fatalf("Duplicate resource: %+v\n in course: %+v",
					tt_data.Resources[ri], cinfo)
			}
		}

		// Build a TtActivity for each lesson.
		// First sort according to length.
		activities := make([]*base.Lesson, 0, len(cinfo.Activities))
		for _, lref := range cinfo.Activities {
			l := db.Elements[lref].(*base.Lesson)
			if slices.Contains(l.Flags, "SubstitutionService") {
				cinfo.Groups = nil
			}
			d := l.Duration
			var i int = 0
			for _, a := range activities {
				if a.Duration <= d {
					break
				}
				i++
			}
			activities = slices.Insert(activities, i, l)
		}

		for _, l := range activities {
			//p := -1
			//if l.Day >= 0 {
			//	p = l.Day*tt_data.NHours + l.Hour
			//}
			aix := ActivityIndex(len(tt_data.Activities))
			ttl := &TtActivity{
				Id: aix,
				//Placement:  p,
				Duration:    int16(l.Duration),
				Fixed:       l.Fixed,
				Resources:   resources,
				RoomChoices: roomchoices,
				//Lesson:     l,
				//CourseInfo: cinfo,
			}
			cinfo.ActivityGroup = append(cinfo.ActivityGroup, aix)
			tt_data.Activities = append(tt_data.Activities, ttl)
		}
	}
}
