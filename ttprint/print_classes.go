package ttprint

/*
// TODO: Try to find a form suitable for both fet and w365 which can be
// passed into the timetable generator.
type TTGroup struct {
	// Represents the groups in a tile in the class view
	Groups []string
	Offset int
	Size   int
	Total  int
}

type LessonData struct {
	Duration  int
	Subject   string
	Teacher   []string
	Students  map[string][]TTGroup // mapping: class -> list of groups
	RealRooms []string
	Day       int
	Hour      int
}

type IdName struct {
	Id   string
	Name string
}

type TimetableData struct {
	Info        map[string]string
	ClassList   []IdName
	TeacherList []IdName
	RoomList    []IdName
	Lessons     []LessonData
}

func PrintClassTimetables(
	ttinfo *ttbase.TtInfo,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	db := ttinfo.Db
	pages := [][]any{}
	ref2id := ttinfo.Ref2Tag
	type cdata struct { // for SuperCourses
		groups   map[base.Ref]bool
		rooms    map[base.Ref]bool
		teachers map[base.Ref]bool
	}
	type chip struct {
		class  string
		groups []string
		offset int
		parts  int
		total  int
	}
	// Generate the tiles.
	classTiles := map[string][]Tile{}
	for cref, cinfo := range ttinfo.CourseInfo {
		subject := ref2id[cinfo.Subject]
		// For SuperCourses gather the resources from the relevant SubCourses.
		subrefs, ok := ttinfo.SuperSubs[cref]
		if ok {
			cmap := map[base.Ref]cdata{}
			for _, subref := range subrefs {
				sub := db.Elements[subref].(*base.SubCourse)
				for _, gref := range sub.Groups {
					g := db.Elements[gref].(*base.Group)
					cref := g.Class
					c := db.Elements[cref].(*base.Class)

					cdata1, ok := cmap[cref]
					if !ok {
						cdata1 = cdata{
							map[base.Ref]bool{},
							map[base.Ref]bool{},
							map[base.Ref]bool{},
						}
					}

					for _, gref1 := range sub.Groups {
						cdata1.groups[gref1] = true
					}
					for _, tref := range sub.Teachers {
						cdata1.teachers[tref] = true
					}
					if sub.Room != "" {
						cdata1.rooms[sub.Room] = true
					}
					cmap[cref] = cdata1
				}
			}
			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				l := ttinfo.Activities[lix].Lesson
				rooms := l.Rooms
				for cref, cdata1 := range cmap {
					tlist := []base.Ref{}
					for t := range cdata1.teachers {
						tlist = append(tlist, t)
					}
					glist := []base.Ref{}
					for g := range cdata1.groups {
						glist = append(glist, g)
					}
					// Choose the rooms in "rooms" which could be relevant for
					// the list of (general) rooms in cdata1.rooms.
					rlist := []base.Ref{}
					for rref := range cdata1.rooms {
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

					//TODO: The groups need special handling, to determine
					// tile fractions (with the groups from the current class)
					chips := ttinfo.SortClassGroups(cref, glist)

					//gstrings := sortList(ordering, ref2id, glist)

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

			//TODO ...

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



			for _, l := range ttdata.Lessons {
		// Limit the length of the room list.
		var room string
		if len(l.RealRooms) > 6 {
			room = strings.Join(l.RealRooms[:5], ",") + "..."
		} else {
			room = strings.Join(l.RealRooms, ",")
		}
		// Limit the length of the teachers list.
		var teacher string
		if len(l.Teacher) > 6 {
			teacher = strings.Join(l.Teacher[:5], ",") + "..."
		} else {
			teacher = strings.Join(l.Teacher, ",")
		}
		chips := []chip{}
		for cl, ttglist := range l.Students {
			for _, ttg := range ttglist {
				chips = append(chips, chip{cl,
					ttg.Groups, ttg.Offset, ttg.Size, ttg.Total})
			}
		}
		for _, c := range chips {
			tile := Tile{
				Day:      l.Day,
				Hour:     l.Hour,
				Duration: l.Duration,
				Fraction: c.parts,
				Offset:   c.offset,
				Total:    c.total,
				Centre:   l.Subject,
				TL:       teacher,
				TR:       strings.Join(c.groups, ","),
				BR:       room,
			}
			classTiles[c.class] = append(classTiles[c.class], tile)
		}
	}

	for _, cl := range ttdata.ClassList {
		ctiles, ok := classTiles[cl.Id]
		if !ok {
			continue
		}
		pages = append(pages, []any{
			fmt.Sprintf("Klasse %s", cl.Id),
			ctiles,
		})
	}
	tt := Timetable{
		Title: "Stundenpläne der Klassen",
		//Info:  ttdata.Info,
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
*/
