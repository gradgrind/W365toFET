package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
)

//const CLASS_GROUP_SEP = "."

func PrintTeacherTimetables(
	//ttdata TimetableData,
	ttinfo *ttbase.TtInfo,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	pages := [][]any{}
	ref2id := ttinfo.Ref2Tag
	// Generate the tiles.
	teacherTiles := map[base.Ref][]Tile{}

	for cref, cinfo := range ttinfo.CourseInfo {

		//

		subject := ref2id[cinfo.Subject]
		groups := []string{}
		// For SuperCourses gather the resources from the relevant SubCourses.
		_, ok := ttinfo.SuperSubs[cref]
		if ok {
			//TODO

		} else {
			for _, gref := range cinfo.Groups {
				groups = append(groups, ref2id[gref])
			}

			// The rooms are associated with the lessons
			for _, lix := range cinfo.Lessons {
				rooms := []string{}
				l := ttinfo.Activities[lix].Lesson
				for _, rref := range l.Rooms {
					rooms = append(rooms, ref2id[rref])
				}

				//TODO

				// The lists should be sorted, somehow ...

				// Limit list lengths?

				for _, tref := range cinfo.Teachers {

					// If there is more than one teacher, list the others
					teachers := []string{}
					if len(cinfo.Teachers) > 1 {
						for _, tref1 := range cinfo.Teachers {
							if tref1 != tref {
								teachers = append(teachers, ref2id[tref1])
							}
						}
					}

					tile := Tile{
						Day:      l.Day,
						Hour:     l.Hour,
						Duration: l.Duration,
						Fraction: 1,
						Offset:   0,
						Total:    1,
						Centre:   strings.Join(groups, ","),
						TL:       subject,
						TR:       strings.Join(teachers, ","),
						BR:       strings.Join(rooms, ","),
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
	db := ttinfo.Db
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
