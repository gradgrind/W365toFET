package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"fmt"
	"slices"
	"strings"
)

func genTypstTeacherData(
	ttinfo *ttbase.TtInfo,
	plan_name string,
	datadir string,
	stemfile string, // basic name part of source file
	flags map[string]bool,
) {
	db := ttinfo.Db
	pages := [][]any{}
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
		subrefs, ok := ttinfo.SuperSubs[cref]
		if ok {
			tmap := map[base.Ref]tdata{}
			for _, subref := range subrefs {
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
					//TODO: Rather pass lists and let the Typst template
					// decide how to join or shorten them?
					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						Fraction: 1,
						Offset:   0,
						Total:    1,
						Centre:   strings.Join(gstrings, ","),
						TL:       subject,
						TR:       strings.Join(tstrings, ","),
						BR:       strings.Join(rstrings, ","),
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
					//TODO: Rather pass lists and let the Typst template
					// decide how to join or shorten them?
					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						Fraction: 1,
						Offset:   0,
						Total:    1,
						Centre:   strings.Join(gstrings, ","),
						TL:       subject,
						TR:       strings.Join(tstrings, ","),
						BR:       strings.Join(rstrings, ","),
					}
					teacherTiles[tref] = append(teacherTiles[tref], tile)
				}
			}
		}
	}

	for _, t := range db.Teachers {
		ctiles, ok := teacherTiles[t.Id]
		if !ok {
			continue
		}
		pages = append(pages, []any{
			fmt.Sprintf("%s %s (%s)", t.Firstname, t.Name, t.Tag),
			ctiles,
		})
	}
	dlist := []string{}
	for _, d := range db.Days {
		dlist = append(dlist, d.Name)
	}
	hlist := []ttHour{}
	for _, h := range db.Hours {
		hlist = append(hlist, ttHour{
			Hour:  h.Tag,
			Start: h.Start,
			End:   h.End,
		})
	}
	info := map[string]any{
		"School":     db.Info.Institution,
		"Days":       dlist,
		"Hours":      hlist,
		"WithTimes":  flags["WithTimes"],
		"WithBreaks": flags["WithBreaks"],
	}
	tt := Timetable{
		Title: "Stundenpläne der Lehrer",
		Info:  info,
		Plan:  plan_name,
		Pages: pages,
	}
	makeTypstJson(tt, datadir, stemfile+"_teachers")
}
