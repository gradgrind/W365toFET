package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func PrintTeacherTimetables(
	//ttdata TimetableData,
	ttinfo *ttbase.TtInfo,
	ordering map[base.Ref]int,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	db := ttinfo.Db
	pages := [][]any{}
	ref2id := ttinfo.Ref2Tag
	// Generate the tiles.
	teacherTiles := map[base.Ref][]Tile{}
	type tdata struct { // for SuperCourses
		groups   map[base.Ref]bool
		rooms    map[base.Ref]bool
		teachers map[base.Ref]bool
	}
	for cref, cinfo := range ttinfo.CourseInfo {
		subject := ref2id[cinfo.Subject]
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
					gstrings := sortList(ordering, ref2id, glist)
					tstrings := sortList(ordering, ref2id, tlist)
					rstrings := sortList(ordering, ref2id, rlist)
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
			gstrings := sortList(ordering, ref2id, glist)

			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				rlist := []base.Ref{}
				l := ttinfo.Activities[lix].Lesson
				rlist = append(rlist, l.Rooms...)
				rstrings := sortList(ordering, ref2id, rlist)

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
					tstrings := sortList(ordering, ref2id, tlist)
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

	{
		/* Limit the length of the room list.
		var room string
		if len(l.RealRooms) > 6 {
			room = strings.Join(l.RealRooms[:5], ",") + "..."
		} else {
			room = strings.Join(l.RealRooms, ",")
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
		*/
	}
	for _, t := range db.Teachers {
		ctiles, ok := teacherTiles[t.Id]
		if !ok {
			continue
		}
		pages = append(pages, []any{
			fmt.Sprintf("%s (%s)", t.Name, t.Tag),
			ctiles,
		})
	}
	info := map[string]string{
		"School": db.Info.Institution,
	}
	tt := Timetable{
		Title: "Stundenpläne der Lehrer",
		Info:  info,
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
		base.Error.Fatal(err)
	}
	fmt.Printf("Wrote json to: %s\n", jsonpath)
	cmd := exec.Command("typst", "compile",
		"--root", datadir,
		"--input", "ifile="+filepath.Join("..", jsonfile),
		filepath.Join(datadir, "resources", "print_timetable.typ"),
		outpath)
	fmt.Printf(" ::: %s\n", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("(Typst) " + string(output))
		base.Error.Fatal(err)
	}
}
