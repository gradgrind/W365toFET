package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"slices"
)

type Tile struct {
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Duration int    `json:"duration"`
	Fraction int    `json:"fraction"`
	Offset   int    `json:"offset"`
	Total    int    `json:"total"`
	Centre   string `json:"centre"`
	TL       string `json:"tl"`
	TR       string `json:"tr"`
	BR       string `json:"br"`
	BL       string `json:"bl"`
}

type Timetable struct {
	Title string
	Info  map[string]any
	Plan  string
	Pages [][]any
}

type ttHour struct {
	Hour  string
	Start string
	End   string
}

func orderResources(ttinfo *ttbase.TtInfo) map[base.Ref]int {
	// Needed for sorting teachers, groups and rooms
	db := ttinfo.Db
	i := 0
	olist := map[base.Ref]int{}
	for _, t := range db.Teachers {
		olist[t.Id] = i
		i++
	}
	for _, r := range db.Rooms {
		olist[r.Id] = i
		i++
	}
	for _, c := range db.Classes {
		olist[c.ClassGroup] = i
		i++
		for _, div := range ttinfo.ClassDivisions[c.Id] {
			for _, gref := range div {
				olist[gref] = i
				i++
			}
		}
	}
	return olist
}

func sortList(
	ordering map[base.Ref]int,
	ref2tag map[base.Ref]string,
	list []base.Ref,
) []string {
	olist := []string{}
	if len(list) > 1 {
		slices.SortFunc(list, func(a, b base.Ref) int {
			if ordering[a] < ordering[b] {
				return -1
			}
			return 1
		})
		for _, ref := range list {
			olist = append(olist, ref2tag[ref])
		}
	} else if len(list) == 1 {
		olist = append(olist, ref2tag[list[0]])
	}
	return olist
}

/*
func PrepareData(ttinfo *ttbase.TtInfo) TimetableData {
	ref2id := ttinfo.Ref2Tag
	// Get the rooms contained in room-groups
	//TODO??
	room_groups := map[int][]string{}
	for _, ri := range wzdb.TableMap["ROOMS"] {
		rg := wzdb.GetNode(ri).(wzbase.Room).SUBROOMS
		if len(rg) != 0 {
			rglist := []string{}
			for _, r := range rg {
				rglist = append(rglist, ref2id[r])
			}
			slices.Sort(rglist)
			room_groups[ri] = rglist
		}
	}

	// Class-group infrastructure
	divmap := map[string][][]string{}
	for c, ad := range wzdb.ActiveDivisions {
		divlist := [][]string{}
		for _, div := range ad {
			gs := []string{}
			for _, g := range div {
				gs = append(gs, ref2id[g])
			}
			divlist = append(divlist, gs)
		}
		divmap[ref2id[c]] = divlist
		//fmt.Printf(" $$$ AD %s: %+v\n", ref2id[c], divlist)
	}

	lessons := []LessonData{}
	for _, a := range activities {
		if a.Day < 0 {
			// Unplaced activity, skip it.
			continue
		}
		// Gather the rooms.
		rooms := []string{}
		if len(a.Rooms) == 0 {
			// Check whether there are compulsory rooms (possible with
			// undeclared room-group).
			for _, r := range a.RoomNeeds.Compulsory {
				rooms = append(rooms, ref2id[r])
			}
			if len(rooms) > 1 {
				slices.Sort(rooms)
			}
		} else {
			for _, r := range a.Rooms {
				rg, ok := room_groups[r]
				if ok {
					rooms = append(rooms, rg...)
				} else {
					rooms = append(rooms, ref2id[r])
				}
			}
		}
		// Gather the teachers.
		teachers := []string{}
		for _, t := range a.Teachers {
			teachers = append(teachers, ref2id[t])
		}

		//TODO: Is there any way of associating teachers with particular
		// (sub)groups? Probably not (with the current data structures).

		// Gather student groups, dividing them for the class view.
		classes := map[string][]string{} // mapping: class -> list of groups
		for _, cg := range a.Groups {
			c := ref2id[cg.CIX]
			g := ref2id[cg.GIX]
			// Assume the groups are valid
			if g == "" {
				classes[c] = nil
			} else {
				classes[c] = append(classes[c], g)
			}
		}
		cgroups := map[string][]TTGroup{}
		for c, glist := range classes {
			var ttgroups []TTGroup
			if len(glist) == 0 {
				// whole class
				ttgroups = []TTGroup{{nil, 0, 1, 1}}
			} else {
				n := 0
				start := 0
				gs := []string{}
				for _, div := range divmap[c] {
					for i, g := range div {
						if slices.Contains(glist, g) {
							n += 1
							if (start + len(gs)) == i {
								gs = append(gs, g)
								continue
							}
							if len(gs) > 0 {
								ttgroups = append(ttgroups,
									TTGroup{gs, start, len(gs), len(div)})
							}
							gs = []string{g}
							start = i
						}
					}
					if len(gs) > 0 {
						ttgroups = append(ttgroups,
							TTGroup{gs, start, len(gs), len(div)})
					}
					if n != 0 {
						if n != len(glist) {
							log.Fatalf("Groups in activity for class %s"+
								" not in one division: %+v\n", c, glist)
						}
						break
					}
				}
				if n == 0 {
					log.Fatalf("Invalid groups in activity for class %s: %+v\n",
						c, glist)
				}
			}
			cgroups[c] = ttgroups
		}
		lessons = append(lessons, LessonData{
			Duration:  a.Duration,
			Subject:   ref2id[a.Subject],
			Teacher:   teachers,
			Students:  cgroups,
			RealRooms: rooms,
			Day:       a.Day,
			Hour:      a.Hour,
		})
	}

	info := map[string]string{
		"School": wzdb.Schooldata["SchoolName"].(string),
	}
	// Assume the classes table is sorted!
	clist := []IdName{}
	for _, ci := range wzdb.TableMap["CLASSES"] {
		node := wzdb.GetNode(ci).(wzbase.Class)
		clist = append(clist, IdName{
			node.ID,
			node.NAME,
		})
	}
	// Assume the teacher table is sorted!
	tlist := []IdName{}
	for _, ti := range wzdb.TableMap["TEACHERS"] {
		node := wzdb.GetNode(ti).(wzbase.Teacher)
		tlist = append(tlist, IdName{
			node.ID,
			node.FIRSTNAMES + " " + node.LASTNAME,
		})
	}
	// Assume the room table is sorted!
	rlist := []IdName{}
	for _, ri := range wzdb.TableMap["ROOMS"] {
		// Keep only "real" rooms
		if _, ok := room_groups[ri]; !ok {
			node := wzdb.GetNode(ri).(wzbase.Room)
			rlist = append(rlist, IdName{
				node.ID,
				node.NAME,
			})
			//fmt.Printf("$ ROOM: %s\n", ref2id[ri])
		}
	}
	return TimetableData{
		Info:        info,
		ClassList:   clist,
		TeacherList: tlist,
		RoomList:    rlist,
		Lessons:     lessons,
	}
}
*/
