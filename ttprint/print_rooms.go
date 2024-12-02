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
