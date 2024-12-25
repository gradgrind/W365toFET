package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

func getTeachers(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string, // basic name part of source file
) string {
	data := getTeacherData(ttinfo, datadir, stemfile)
	pages := []ttPage{}
	for _, t := range ttinfo.Db.Teachers {
		ttiles, ok := data[t.Id]
		if !ok {
			continue
		}
		pages = append(pages, ttPage{
			Name:       t.Firstname + " " + t.Name,
			Short:      t.Tag,
			Activities: ttiles,
		})
	}
	tt := timetable(ttinfo.Db, pages, "Teacher")
	f := stemfile + "_teachers"
	makeTypstJson(tt, datadir, f)
	return f
}

func getOneTeacher(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string, // basic name part of source file
	e *base.Teacher,
) string {
	data := getTeacherData(ttinfo, datadir, stemfile)
	tiles, ok := data[e.Id]
	if !ok {
		tiles = []Tile{} // Avoid none in JSON if table empty
	}
	pages := []ttPage{
		ttPage{
			Name:       e.Firstname + " " + e.Name,
			Short:      e.Tag,
			Activities: tiles,
		},
	}
	tt := timetable(ttinfo.Db, pages, "Teacher")
	f := stemfile + "_teacher_" + e.Tag
	makeTypstJson(tt, datadir, f)
	return f
}

func getTeacherData(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string, // basic name part of source file
) map[base.Ref][]Tile {
	db := ttinfo.Db
	// Generate the tiles.
	teacherTiles := map[base.Ref][]Tile{}
	type tdata struct { // for SuperCourses
		groups   map[base.Ref]bool
		rooms    map[base.Ref]bool
		teachers map[base.Ref]bool
	}
	for cref, cinfo := range ttinfo.CourseInfo {
		subject := ttinfo.Ref2Tag[cinfo.Subject]
		// For SuperCourses gather the resources from the relevant SubCourses.
		sc, ok := db.Elements[cref].(*base.SuperCourse)
		if ok {
			tmap := map[base.Ref]tdata{}
			for _, subref := range sc.SubCourses {
				sub := db.Elements[subref].(*base.SubCourse)
				for _, tref := range sub.Teachers {
					tdata1, ok := tmap[tref]
					if !ok {
						tdata1 = tdata{
							map[base.Ref]bool{},
							map[base.Ref]bool{},
							map[base.Ref]bool{},
						}
					}
					// If there is more than one teacher, add the others
					for _, tref1 := range sub.Teachers {
						if tref1 != tref {
							tdata1.teachers[tref1] = true
						}
					}
					if sub.Room != "" {
						tdata1.rooms[sub.Room] = true
					}
					for _, gref := range sub.Groups {
						tdata1.groups[gref] = true
					}
					tmap[tref] = tdata1
				}
			}
			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				if l.Day < 0 {
					continue
				}
				rooms := l.Rooms
				for tref, tdata1 := range tmap {
					tlist := []base.Ref{}
					for t := range tdata1.teachers {
						tlist = append(tlist, t)
					}
					glist := []base.Ref{}
					for g := range tdata1.groups {
						glist = append(glist, g)
					}
					// Choose the rooms in "rooms" which could be relevant for
					// the list of (general) rooms in tdata1.rooms.
					rlist := []base.Ref{}
					for rref := range tdata1.rooms {
						rx := db.Elements[rref]
						_, ok := rx.(*base.Room)
						if ok {
							if slices.Contains(rooms, rref) {
								rlist = append(rlist, rref)
							}
							continue
						}
						rg, ok := rx.(*base.RoomGroup)
						if ok {
							for _, rr := range rg.Rooms {
								if slices.Contains(rooms, rr) {
									rlist = append(rlist, rr)
								}
							}
							continue
						}
						rc, ok := rx.(*base.RoomChoiceGroup)
						if ok {
							for _, rr := range rc.Rooms {
								if slices.Contains(rooms, rr) {
									rlist = append(rlist, rr)
								}
							}
							continue
						}
						base.Bug.Fatalf("Not a room: %s\n", rref)
					}
					gstrings := ttinfo.SortList(glist)
					tstrings := ttinfo.SortList(tlist)
					rstrings := ttinfo.SortList(rlist)
					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						//Fraction: 1,
						//Offset:   0,
						//Total:    1,
						Subject:  subject,
						Groups:   gstrings,
						Teachers: tstrings,
						Rooms:    rstrings,
					}
					teacherTiles[tref] = append(teacherTiles[tref], tile)
				}
			}
		} else {
			// A normal Course
			glist := []base.Ref{}
			glist = append(glist, cinfo.Groups...)
			gstrings := ttinfo.SortList(glist)

			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				if l.Day < 0 {
					continue
				}
				rlist := []base.Ref{}
				rlist = append(rlist, l.Rooms...)
				rstrings := ttinfo.SortList(rlist)

				for _, tref := range cinfo.Teachers {
					// If there is more than one teacher, list the others
					tlist := []base.Ref{}
					if len(cinfo.Teachers) > 1 {
						for _, tref1 := range cinfo.Teachers {
							if tref1 != tref {
								tlist = append(tlist, tref1)
							}
						}
					}
					tstrings := ttinfo.SortList(tlist)
					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						//Fraction: 1,
						//Offset:   0,
						//Total:    1,
						Subject:  subject,
						Groups:   gstrings,
						Teachers: tstrings,
						Rooms:    rstrings,
					}
					teacherTiles[tref] = append(teacherTiles[tref], tile)
				}
			}
		}
	}
	return teacherTiles
}
