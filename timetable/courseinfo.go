package timetable

import (
	"W365toFET/base"
	"fmt"
	"slices"
	"strings"
)

// A CourseInfo is a representation of a course (Course or SuperCourse) for
// the timetable.
// Lessons within a course are (already) ordered, highest duration first,
// and the Activities field has the same order.
type CourseInfo struct {
	Id           NodeRef // Course or SuperCourse
	Subject      string
	Groups       []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroups []AtomicIndex
	Teachers     []TeacherIndex
	FixedRooms   []RoomIndex
	RoomChoices  [][]RoomIndex
	Lessons      []*base.Lesson
	Activities   []ActivityIndex
}

type Activity struct {
	CourseInfo *CourseInfo
	Lesson     *base.Lesson
	Placement  TtSlot
	Duration   int16
	Fixed      bool
}

// Make a shortish string view of a CourseInfo – can be useful in tests
func (tt_shared_data *TtSharedData) View(cinfo *CourseInfo) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, tt_shared_data.Db.Teachers[t].GetTag())
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, base.GroupTag(g))
	}
	return fmt.Sprintf("<Course %s/%s:%s>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		cinfo.Subject,
	)
}

// Collect courses (Course and SuperCourse) and their activities.
// Build a list of CourseInfo structures.
func (tt_shared_data *TtSharedData) CollectCourses() {
	db := tt_shared_data.Db
	tt_shared_data.Ref2CourseInfo = map[NodeRef]*CourseInfo{}
	tt_shared_data.Activities = []*Activity{{}} // first entry is empty

	// *** Gather the SuperCourses. ***
	for _, spc := range db.SuperCourses {
		cref := spc.Id
		groups := []*base.Group{}
		agroups := []AtomicIndex{}
		teachers := []TeacherIndex{}
		rooms := []RoomIndex{}
		crooms := [][]RoomIndex{}
		for _, sbc := range spc.SubCourses {
			// Add groups
			for _, gref := range sbc.Groups {
				g, ok := db.GetElement(gref).(*base.Group)
				if !ok {
					panic("Invalid Group ref: " + gref)
				}
				if !slices.Contains(groups, g) {
					groups = append(groups, g)
					agroups = append(agroups, tt_shared_data.AtomicGroups[gref]...)
				}
			}
			// Add teachers
			for _, tref := range sbc.Teachers {
				t, ok := tt_shared_data.TeacherIndex[tref]
				if !ok {
					panic("Invalid Teacher ref: " + tref)
				}
				teachers = append(teachers, t)
			}
			// Add rooms
			if sbc.Room != "" {
				r, ok := tt_shared_data.RoomIndex[sbc.Room]
				if ok {
					rooms = append(rooms, r)
					continue
				}

				// Not a `Room` – it can be a RoomGroup or RoomChoiceGroup

				gr := db.GetElement(sbc.Room)
				rg, ok := gr.(*base.RoomGroup)
				if ok {
					for _, rr := range rg.Rooms {
						r, ok = tt_shared_data.RoomIndex[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Bug: Unknown room in RoomGroup %s: %s",
								rr, sbc.Room))
						}
						rooms = append(rooms, r)
					}
					continue
				}

				rcg, ok := gr.(*base.RoomChoiceGroup)
				if ok {
					roomlist := []RoomIndex{}
					for _, rr := range rcg.Rooms {
						r, ok = tt_shared_data.RoomIndex[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Bug: Unknown room in RoomChoiceGroup %s: %s",
								rr, sbc.Room))
						}
						roomlist = append(roomlist, r)
					}
					slices.Sort(roomlist)
					// Don't add if it is a duplicate
					for _, rl := range crooms {
						if slices.Equal(rl, roomlist) {
							goto skip
						}
					}
					crooms = append(crooms, roomlist)
				skip:
					continue
				}

				panic("Bug: Expecting room element, found: " + sbc.Room)
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

		//fmt.Printf("^^^^ %s\n\n", sbj.Tag)

		cinfo := &CourseInfo{
			Id:           cref,
			Subject:      sbj.Tag,
			Groups:       groups,
			AtomicGroups: slices.Compact(agroups),
			Teachers:     slices.Compact(teachers),
			FixedRooms:   slices.Compact(rooms),
			RoomChoices:  crooms,
			Lessons:      spc.Lessons,
			//Activities
		}

		// Filter out any "necessary" rooms from the choices
		tt_shared_data.roomChoiceFilter(cinfo)

		tt_shared_data.makeActivities(cinfo)
		tt_shared_data.CourseInfoList = append(
			tt_shared_data.CourseInfoList, cinfo)
		tt_shared_data.Ref2CourseInfo[cref] = cinfo
	}

	// *** Gather the plain Courses. ***
	for _, c := range db.Courses {
		cref := c.Id

		// Get groups
		groups := []*base.Group{}
		agroups := []AtomicIndex{}
		for _, gref := range c.Groups {
			g, ok := db.GetElement(gref).(*base.Group)
			if !ok {
				panic("Invalid Group ref: " + gref)
			}
			groups = append(groups, g)
			agroups = append(agroups, tt_shared_data.AtomicGroups[gref]...)
		}

		// Get teachers
		teachers := []TeacherIndex{}
		for _, tref := range c.Teachers {
			t, ok := tt_shared_data.TeacherIndex[tref]
			if !ok {
				panic("Invalid Teacher ref: " + tref)
			}
			teachers = append(teachers, t)
		}

		// Get rooms
		rooms := []RoomIndex{}
		crooms := [][]RoomIndex{}
		if c.Room != "" {
			r, ok := tt_shared_data.RoomIndex[c.Room]
			if ok {
				rooms = append(rooms, r)
			} else {
				// Not a `Room` – it can be a RoomGroup or RoomChoiceGroup
				gr := db.GetElement(c.Room)
				rg, ok := gr.(*base.RoomGroup)
				if ok {
					for _, rr := range rg.Rooms {
						r, ok = tt_shared_data.RoomIndex[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Unknown room in RoomGroup %s: %s",
								rr, c.Room))
						}
						rooms = append(rooms, r)
					}
				} else {
					rcg, ok := gr.(*base.RoomChoiceGroup)
					if ok {
						roomlist := []RoomIndex{}
						for _, rr := range rcg.Rooms {
							r, ok = tt_shared_data.RoomIndex[rr]
							if !ok {
								panic(fmt.Sprintf(
									"Unknown room in RoomChoiceGroup %s: %s",
									rr, c.Room))
							}
							roomlist = append(roomlist, r)
						}
						crooms = append(crooms, roomlist)
					} else {
						panic("Expecting room element, found: " + c.Room)
					}
				}
			}
		}

		sbj, ok := db.GetElement(c.Subject).(*base.Subject)
		if !ok {
			panic("Invalid Subject ref: " + c.Subject)
		}

		//fmt.Printf("^^__ %s\n\n", sbj.Tag)

		// Sort and compact lists
		slices.Sort(agroups)
		slices.Sort(teachers) // shouldn't need compacting
		slices.Sort(rooms)    // shouldn't need compacting
		cinfo := &CourseInfo{
			Id:           cref,
			Subject:      sbj.Tag,
			Groups:       groups,
			AtomicGroups: slices.Compact(agroups),
			Teachers:     teachers,
			FixedRooms:   rooms,
			RoomChoices:  crooms,
			Lessons:      c.Lessons,
			//Activities
		}

		tt_shared_data.makeActivities(cinfo)
		tt_shared_data.CourseInfoList = append(
			tt_shared_data.CourseInfoList, cinfo)
		tt_shared_data.Ref2CourseInfo[cref] = cinfo
	}
}

// Build an `Activity` for each `Lesson` – they are already sorted
// with the longest first.
func (tt_shared_data *TtSharedData) makeActivities(cinfo *CourseInfo) {
	for _, l := range cinfo.Lessons {
		p := -1
		if l.Day >= 0 {
			p = l.Day*tt_shared_data.NHours + l.Hour
		}
		aix := ActivityIndex(len(tt_shared_data.Activities))
		ttl := &Activity{
			CourseInfo: cinfo,
			Lesson:     l,
			Placement:  TtSlot(p),
			Duration:   int16(l.Duration),
			Fixed:      l.Fixed,
		}
		cinfo.Activities = append(cinfo.Activities, aix)
		tt_shared_data.Activities = append(tt_shared_data.Activities, ttl)
	}
}
