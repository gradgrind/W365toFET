package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
)

func genTypstRoomData(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string, // basic name part of source file
) string {
	db := ttinfo.Db
	pages := []ttPage{}
	// Generate the tiles.
	roomTiles := map[base.Ref][]Tile{}
	type rdata struct { // for SuperCourses
		groups   map[base.Ref]bool
		teachers map[base.Ref]bool
	}

	for cref, cinfo := range ttinfo.CourseInfo {
		subject := ttinfo.Ref2Tag[cinfo.Subject]
		// For SuperCourses gather the resources from the relevant SubCourses.
		subrefs, ok := ttinfo.SuperSubs[cref]
		if ok {
			rmap := map[base.Ref]rdata{}
			for _, subref := range subrefs {
				sub := db.Elements[subref].(*base.SubCourse)
				if sub.Room != "" {
					rlist := []base.Ref{}
					r0 := db.Elements[sub.Room]
					r, ok := r0.(*base.Room)
					if ok {
						rlist = append(rlist, r.Id)
					} else {
						r, ok := r0.(*base.RoomGroup)
						if ok {
							rlist = append(rlist, r.Rooms...)
						} else {
							r, ok := r0.(*base.RoomChoiceGroup)
							if ok {
								rlist = append(rlist, r.Rooms...)
							} else {
								base.Bug.Fatalf("Invalid room in course %s:"+
									"\n  %+v\n", ttinfo.View(cinfo), r0)
							}
						}
					}
					for _, rref := range rlist {
						rdata1, ok := rmap[rref]
						if !ok {
							rdata1 = rdata{
								map[base.Ref]bool{},
								map[base.Ref]bool{},
							}
						}
						for _, tref1 := range sub.Teachers {
							rdata1.teachers[tref1] = true
						}
						for _, gref := range sub.Groups {
							rdata1.groups[gref] = true
						}
						rmap[rref] = rdata1
					}
				}
			}

			// The actual rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				if l.Day < 0 {
					continue
				}
				for _, rref := range l.Rooms {
					tlist := []base.Ref{}
					glist := []base.Ref{}
					rdata1, ok := rmap[rref]
					if ok {
						for t := range rdata1.teachers {
							tlist = append(tlist, t)
						}
						for g := range rdata1.groups {
							glist = append(glist, g)
						}
					}
					gstrings := ttinfo.SortList(glist)
					tstrings := ttinfo.SortList(tlist)
					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						//Fraction: 1,
						//Offset: 0,
						//Total:    1,
						Subject:  subject,
						Groups:   gstrings,
						Teachers: tstrings,
						//TODO: Background:
					}
					roomTiles[rref] = append(roomTiles[rref], tile)
				}
			}
		} else {
			// A normal Course
			glist := []base.Ref{}
			glist = append(glist, cinfo.Groups...)
			gstrings := ttinfo.SortList(glist)
			tlist := []base.Ref{}
			tlist = append(tlist, cinfo.Teachers...)
			tstrings := ttinfo.SortList(tlist)

			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				if l.Day < 0 {
					continue
				}
				for _, rref := range l.Rooms {
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
						//TODO: Background:
					}
					roomTiles[rref] = append(roomTiles[rref], tile)
				}
			}
		}
	}

	for _, r := range db.Rooms {
		rtiles, ok := roomTiles[r.Id]
		if !ok {
			continue
		}
		pages = append(pages, ttPage{
			Name:       r.Name,
			Short:      r.Tag,
			Activities: rtiles,
		})
	}
	dlist := []ttDay{}
	for _, d := range db.Days {
		dlist = append(dlist, ttDay{
			Name:  d.Name,
			Short: d.Tag,
		})
	}
	hlist := []ttHour{}
	for _, h := range db.Hours {
		hlist = append(hlist, ttHour{
			Name:  h.Name,
			Short: h.Tag,
			Start: h.Start,
			End:   h.End,
		})
	}
	info := map[string]any{
		"Institution": db.Info.Institution,
		"Days":        dlist,
		"Hours":       hlist,
	}
	tt := Timetable{
		TableType: "Room",
		Info:      info,
		Typst:     db.PrintOptions.Typst,
		Pages:     pages,
	}
	f := stemfile + "_rooms"
	makeTypstJson(tt, datadir, f)
	return f
}
