package timetable

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strings"
)

// A CourseInfo is a representation of a course (Course or SuperCourse) for
// the timetable.
// Activities within a course are (already) ordered, highest duration first,
// and ActivityGroup has the same order.
type CourseInfo struct {
	Id           NodeRef // Course or SuperCourse
	Subject      string
	Groups       []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroups []ResourceIndex
	Teachers     []ResourceIndex
	FixedRooms   []ResourceIndex
	RoomChoices  [][]ResourceIndex
	Activities   []ActivityIndex
}

// Make a shortish string view of a CourseInfo – can be useful in tests
func (tt_data *TtData) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, tt_data.Resources[t].GetResourceTag())
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, g.Tag)
	}
	return fmt.Sprintf("<Course %s/%s:%s>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		cinfo.Subject,
	)
}

// Collect courses (Course and SuperCourse) and their activities.
// Build a list of CourseInfo structures.
func (tt_data *TtData) CollectCourses() []CourseInfo {
	db := tt_data.Db
	tt_data.Activities = []*TtActivity{{}} // first entry is empty

	// Gather the SuperCourses.
	for _, spc := range db.SuperCourses {
		groups := []*base.Group{}
		agroups := []ResourceIndex{}
		teachers := []ResourceIndex{}
		rooms := []ResourceIndex{}
		crooms := [][]ResourceIndex{}
		for _, sbc := range spc.SubCourses {
			// Add groups
			for _, gref := range sbc.Groups {
				g, ok := db.GetElement(gref).(*base.Group)
				if !ok {
					panic("Invalid Group ref: " + gref)
				}
				groups = append(groups, g)
				agroups = append(agroups, tt_data.AtomicGroups[gref]...)
			}
			// Add teachers
			for _, tref := range sbc.Teachers {
				t, ok := tt_data.TeacherIndex[tref]
				if !ok {
					panic("Invalid Teacher ref: " + tref)
				}
				teachers = append(teachers, t)
			}
			// Add rooms
			if sbc.Room != "" {
				r, ok := tt_data.RoomIndex[sbc.Room]
				if ok {
					rooms = append(rooms, r)
					continue
				}

				// Not a `Room` – it can be a RoomGroup or RoomChoiceGroup

				gr := db.GetElement(sbc.Room)
				rg, ok := gr.(*base.RoomGroup)
				if ok {
					for _, rr := range rg.Rooms {
						r, ok = tt_data.RoomIndex[rr]
						if !ok {
							base.Bug.Fatalf(
								"Unknown room in RoomGroup %s: %s",
								rr, sbc.Room)
						}
						rooms = append(rooms, r)
					}
					continue
				}

				rcg, ok := gr.(*base.RoomChoiceGroup)
				if ok {
					roomlist := []ResourceIndex{}
					for _, rr := range rcg.Rooms {
						r, ok = tt_data.RoomIndex[rr]
						if !ok {
							base.Bug.Fatalf(
								"Unknown room in RoomChoiceGroup %s: %s",
								rr, sbc.Room)
						}
						roomlist = append(roomlist, r)
					}

					//TODO: don't add if it is a duplicate
					crooms = append(crooms, roomlist)
					continue
				}

				panic("Expecting room element, found: " + sbc.Room)
			}
		}

		// Eliminate duplicate resources by sorting and then compacting
		slices.Sort(agroups)
		slices.Sort(teachers)
		slices.Sort(rooms)
		sbj, ok := db.GetElement(spc.Subject).(*base.Subject)
		if !ok {
			panic("Invalid Subject ref: " + spc.Subject)
		}
		tt_data.CourseInfoList = append(tt_data.CourseInfoList, &CourseInfo{
			Id:           spc.Id,
			Subject:      sbj.Tag,
			Groups:       groups,
			AtomicGroups: slices.Compact(agroups),
			Teachers:     slices.Compact(teachers),
			FixedRooms:   slices.Compact(rooms),
			RoomChoices:  crooms,

			//TODO

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
		tt_data.CourseInfoList = append(tt_data.CourseInfoList, &CourseInfo{
			Id:         cref,
			Subject:    c.Subject,
			Groups:     c.Groups,
			Teachers:   c.Teachers,
			Rooms:      rooms,
			Activities: c.Lessons,
		})
	}
}

// `MakeActivities` creates the `TtActivity` structures ...
func (tt_data *TtData) MakeActivities() {
	db := tt_data.Db
	tt_data.Activities = []*TtActivity{{}} // first entry is empty
	for _, cinfo := range tt_data.CourseInfoList {
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

		// Build a TtActivity for each Activity – they are already sorted
		// with the longest first.
		for _, l := range cinfo.Activities {
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
