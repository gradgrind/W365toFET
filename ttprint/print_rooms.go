package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"fmt"
	"strings"
)

func genTypstRoomData(
	ttinfo *ttbase.TtInfo,
	plan_name string,
	datadir string,
	stemfile string, // basic name part of source file
	flags map[string]bool,
) {
	db := ttinfo.Db
	pages := [][]any{}
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

				//TODO: A room can be part of sub.Room OR
				// added in the lessons.

				if sub.Room != "" {
					rref := sub.Room

					//

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
			// The rooms are associated with the lessons

			//TODO: Check on this. I think I had decided that the rooms
			// in the lesson's list are the chosen ones only, not the
			// compulsory ones.

			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				if l.Day < 0 {
					continue
				}
				//?? rooms := l.Rooms
				for rref, rdata1 := range rmap {

					//TODO: Note that this is probably a GENERAL room ref

					tlist := []base.Ref{}
					for t := range rdata1.teachers {
						tlist = append(tlist, t)
					}
					glist := []base.Ref{}
					for g := range rdata1.groups {
						glist = append(glist, g)
					}
					// Choose the rooms in "rooms" which could be relevant for
					// the list of (general) rooms in tdata1.rooms.
					rlist := []base.Ref{}

					/*????
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
					*/
					gstrings := ttinfo.SortList(glist)
					tstrings := ttinfo.SortList(tlist)
					rstrings := ttinfo.SortList(rlist)
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
					roomTiles[rref] = append(roomTiles[rref], tile)
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
						Fraction: 1,
						Offset:   0,
						Total:    1,
						Centre:   strings.Join(gstrings, ","),
						TL:       subject,
						TR:       strings.Join(tstrings, ","),
						BR:       strings.Join(rstrings, ","),
					}
					roomTiles[tref] = append(roomTiles[tref], tile)
				}
			}
		}
	}

	for _, t := range db.Teachers {
		ctiles, ok := roomTiles[t.Id]
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

/*
func PrintRoomTimetables(
	ttdata TimetableData,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	pages := [][]any{}
	// Generate the tiles.
	roomTiles := map[string][]Tile{}
	for _, l := range ttdata.Lessons {
		// Limit the length of the teachers list.
		var teacher string
		if len(l.Teacher) > 6 {
			teacher = strings.Join(l.Teacher[:5], ",") + "..."
		} else {
			teacher = strings.Join(l.Teacher, ",")
		}
		// Gather student groups.
		var students string
		type c_ttg struct {
			class string
			ttg   []TTGroup
		}
		var c_ttg_list []c_ttg
		if len(l.Students) > 1 {
			// Multiple classes, which need sorting
			for _, c := range ttdata.ClassList {
				ttgroups, ok := l.Students[c.Id]
				if ok {
					c_ttg_list = append(c_ttg_list, c_ttg{c.Id, ttgroups})
				}
			}
		} else {
			for c, ttgroups := range l.Students {
				c_ttg_list = []c_ttg{{c, ttgroups}}
			}
		}
		cgroups := []string{}
		for _, cg := range c_ttg_list {
			for _, ttg := range cg.ttg {
				if len(ttg.Groups) == 0 {
					cgroups = append(cgroups, cg.class)
				} else {
					for _, g := range ttg.Groups {
						cgroups = append(cgroups, cg.class+CLASS_GROUP_SEP+g)
					}
				}
			}
		}
		if len(cgroups) > 10 {
			students = strings.Join(cgroups[:9], ",") + "..."
		} else {
			students = strings.Join(cgroups, ",")
		}
		// Go through the rooms.
		for _, r := range l.RealRooms {
			tile := Tile{
				Day:      l.Day,
				Hour:     l.Hour,
				Duration: l.Duration,
				Fraction: 1,
				Offset:   0,
				Total:    1,
				Centre:   students,
				TL:       l.Subject,
				BR:       teacher,
			}
			roomTiles[r] = append(roomTiles[r], tile)
		}
	}
	for _, r := range ttdata.RoomList {
		ctiles, ok := roomTiles[r.Id]
		if !ok {
			continue
		}
		pages = append(pages, []any{
			fmt.Sprintf("%s (%s)", r.Name, r.Id),
			ctiles,
		})
	}
	tt := Timetable{
		Title: "Stundenpläne der Räume",
		Info:  ttdata.Info,
		Plan:  plan_name,
		Pages: pages,
	}
	b, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	//os.Stdout.Write(b)
	jsonfile := filepath.Join("_out", "tmp.json")
	jsonpath := filepath.Join(datadir, jsonfile)
	err = os.WriteFile(jsonpath, b, 0666)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote json to: %s\n", jsonpath)
	cmd := exec.Command("typst", "compile",
		"--root", datadir,
		"--input", "ifile="+filepath.Join("..", jsonfile),
		filepath.Join(datadir, "resources", "print_timetable.typ"),
		outpath)
	fmt.Printf(" ::: %s\n", cmd.String())
	//TODO: I am not getting any output messages from typst here ...
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}

type MiniTileData struct {
	Duration int    `json:"duration"`
	Top      string `json:"top"`
	Middle   string `json:"middle"`
	Bottom   string `json:"bottom"`
}

type MiniTile struct {
	Day  int          `json:"day"`
	Hour int          `json:"hour"`
	Data MiniTileData `json:"data"`
}

type Row struct {
	Header string
	Items  []MiniTile
}

type BigTimetable struct {
	Title string
	Info  map[string]string
	Plan  string
	Rows  []Row
}

func PrintRoomOverview(
	ttdata TimetableData,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	rows := []Row{}
	// Generate the tiles.
	// They will normally be tiny so the amount of info which can be shown
	// is severely limited.
	roomTiles := map[string][]MiniTile{}
	for _, l := range ttdata.Lessons {
		// Show only one teacher ...
		var teacher string
		if len(l.Teacher) > 1 {
			teacher = l.Teacher[0] + " +"
		} else {
			teacher = l.Teacher[0]
		}
		// Gather student groups.
		var students string
		type c_ttg struct {
			class string
			ttg   []TTGroup
		}
		var c_ttg_list []c_ttg
		if len(l.Students) > 1 {
			// Multiple classes, which need sorting
			for _, c := range ttdata.ClassList {
				ttgroups, ok := l.Students[c.Id]
				if ok {
					c_ttg_list = append(c_ttg_list, c_ttg{c.Id, ttgroups})
				}
			}
		} else {
			for c, ttgroups := range l.Students {
				c_ttg_list = []c_ttg{{c, ttgroups}}
			}
		}
		cgroups := []string{}
		for _, cg := range c_ttg_list {
			for _, ttg := range cg.ttg {
				if len(ttg.Groups) == 0 {
					cgroups = append(cgroups, cg.class)
				} else {
					for _, g := range ttg.Groups {
						cgroups = append(cgroups, cg.class+CLASS_GROUP_SEP+g)
					}
				}
			}
		}
		// Show only one student group ...
		if len(cgroups) > 1 {
			students = cgroups[0] + " +"
		} else {
			students = cgroups[0]
		}
		// Go through the rooms.
		for _, r := range l.RealRooms {
			tile := MiniTile{
				Day:  l.Day,
				Hour: l.Hour,
				Data: MiniTileData{
					Duration: l.Duration,
					Top:      l.Subject,
					Middle:   students,
					Bottom:   teacher,
				},
			}
			roomTiles[r] = append(roomTiles[r], tile)
		}
	}
	for _, r := range ttdata.RoomList {
		ctiles, ok := roomTiles[r.Id]
		if !ok {
			continue
		}
		rows = append(rows, Row{
			fmt.Sprintf("%s (%s)", r.Name, r.Id),
			ctiles,
		})
	}
	tt := BigTimetable{
		Title: "Gesamtansicht der Räume",
		Info:  ttdata.Info,
		Plan:  plan_name,
		Rows:  rows,
	}
	b, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	//os.Stdout.Write(b)
	jsonfile := filepath.Join("_out", "tmp.json")
	jsonpath := filepath.Join(datadir, jsonfile)
	err = os.WriteFile(jsonpath, b, 0666)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wrote json to: %s\n", jsonpath)
	cmd := exec.Command("typst", "compile",
		"--root", datadir,
		"--input", "ifile="+filepath.Join("..", jsonfile),
		filepath.Join(datadir, "resources", "print_whole_timetable.typ"),
		outpath)
	fmt.Printf(" ::: %s\n", cmd.String())
	//TODO: I am not getting any output messages from typst here ...
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
*/
