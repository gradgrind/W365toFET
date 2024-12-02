package ttprint

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
