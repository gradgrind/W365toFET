package timetable

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
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
	Info  map[string]string
	Plan  string
	Pages [][]any
}

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

func PrepareData(wzdb *wzbase.WZdata,
	activities []wzbase.Activity,
	// ) []LessonData {
) TimetableData {
	ref2id := wzdb.Ref2IdMap()
	// Get the rooms contained in room-groups
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

func PrintClassTimetables(
	ttdata TimetableData,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	pages := [][]any{}
	type chip struct {
		class  string
		groups []string
		offset int
		parts  int
		total  int
	}
	// Generate the tiles.
	classTiles := map[string][]Tile{}
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(output))
}
