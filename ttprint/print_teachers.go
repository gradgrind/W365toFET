package timetable

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"strings"
)

const CLASS_GROUP_SEP = "."

func PrintTeacherTimetables(
	ttdata TimetableData,
	plan_name string,
	datadir string,
	outpath string, // full path to output pdf
) {
	pages := [][]any{}
	// Generate the tiles.
	teacherTiles := map[string][]Tile{}
	for _, l := range ttdata.Lessons {
		// Limit the length of the room list.
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
		// Go through the teachers.
		for _, t := range l.Teacher {
			tile := Tile{
				Day:      l.Day,
				Hour:     l.Hour,
				Duration: l.Duration,
				Fraction: 1,
				Offset:   0,
				Total:    1,
				Centre:   students,
				TL:       l.Subject,
				BR:       room,
			}
			teacherTiles[t] = append(teacherTiles[t], tile)
		}
	}
	for _, t := range ttdata.TeacherList {
		ctiles, ok := teacherTiles[t.Id]
		if !ok {
			continue
		}
		pages = append(pages, []any{
			fmt.Sprintf("%s (%s)", t.Name, t.Id),
			ctiles,
		})
	}
	tt := Timetable{
		Title: "Stundenpläne der Lehrer",
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
