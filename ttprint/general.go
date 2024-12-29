package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Tile struct {
	Day        int      `json:"day"`
	Hour       int      `json:"hour"`
	Duration   int      `json:"duration,omitempty"`
	Fraction   int      `json:"fraction,omitempty"`
	Offset     int      `json:"offset,omitempty"`
	Total      int      `json:"total,omitempty"`
	Subject    string   `json:"subject"`
	Groups     []string `json:"groups,omitempty"`
	Teachers   []string `json:"teachers,omitempty"`
	Rooms      []string `json:"rooms,omitempty"`
	Background string   `json:"background,omitempty"`
}

type Timetable struct {
	TableType string
	Info      map[string]any
	Typst     map[string]any `json:",omitempty"`
	Pages     []ttPage
}

type ttDay struct {
	Name  string
	Short string
}

type ttHour struct {
	Name  string
	Short string
	Start string
	End   string
}

type ttPage struct {
	Name       string
	Short      string
	Activities []Tile
}

func GenTimetables(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string,
	commands []base.PrintCommand,
	genpdf string,
) {
	var f string
	var tt Timetable
	for _, cmd := range commands {
		f = cmd.TypstJson
		switch cmd.Type {
		case "Class":
			pages := getClasses(ttinfo)
			tt = timetable(ttinfo.Db, pages, "Class")
			if f == "" {
				f = stemfile + "_classes"
			}
		case "Teacher":
			pages := getTeachers(ttinfo)
			tt = timetable(ttinfo.Db, pages, "Class")
			if f == "" {
				f = stemfile + "_classes"
			}
		case "Room":
			pages := getRooms(ttinfo)
			tt = timetable(ttinfo.Db, pages, "Class")
			if f == "" {
				f = stemfile + "_classes"
			}
		default:
			// Table for individual element
			var tag string
			tt, tag = genTypstOneElement(ttinfo, cmd.TypstJson)
			if f == "" {
				f = stemfile + tag
			}
		}
		makeTypstJson(tt, datadir, f)
		if genpdf != "" {
			//TODO: Leave off endings in command?
			tmpl := cmd.TypstTemplate
			pdf := cmd.Pdf
			if pdf == "" {
				if strings.HasSuffix(tmpl, "_overview") {
					pdf = f + "_overview"
				} else {
					pdf = f
				}
			}
			makePdf(tmpl+".typ", datadir, f, pdf+".pdf", genpdf)
		}
	}
}

func genTypstOneElement(ttinfo *ttbase.TtInfo, id string,
) (Timetable, string) {
	var tt Timetable
	ref := base.Ref(id)
	e := ttinfo.Db.Elements[ref]
	c, ok := e.(*base.Class)
	if ok {
		// Make class JSON, but with only one class
		pages := getOneClass(ttinfo, c)
		tt = timetable(ttinfo.Db, pages, "Class")
		return tt, "class_" + c.Tag
	}
	t, ok := e.(*base.Teacher)
	if ok {
		// Make teacher JSON, but with only one teacher
		pages := getOneTeacher(ttinfo, t)
		tt = timetable(ttinfo.Db, pages, "Teacher")
		return tt, "teacher_" + t.Tag
	}
	r, ok := e.(*base.Room)
	if ok {
		// Make room JSON, but with only one room
		pages := getOneRoom(ttinfo, r)
		tt = timetable(ttinfo.Db, pages, "Room")
		return tt, "room_" + r.Tag
	}
	base.Error.Fatalf("Can't print timetable for invalid element: %+v\n", e)
	return tt, ""
}

func timetable(
	db *base.DbTopLevel,
	pages []ttPage,
	tabletype string, // "Class" or "Teacher" or "Room"
) Timetable {
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
	return Timetable{
		TableType: tabletype,
		Info:      info,
		Typst:     db.PrintOptions.Typst,
		Pages:     pages,
	}
}

func makeTypstJson(tt Timetable, datadir string, outfile string) {
	b, err := json.MarshalIndent(tt, "", "  ")
	if err != nil {
		base.Error.Fatal(err)
	}
	// os.Stdout.Write(b)
	outdir := filepath.Join(datadir, "_data")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Fatal(err)
		}
	}
	jsonpath := filepath.Join(outdir, outfile+".json")
	err = os.WriteFile(jsonpath, b, 0666)
	if err != nil {
		base.Error.Fatal(err)
	}
	base.Message.Printf("Wrote: %s\n", jsonpath)
}

func makePdf(
	script string,
	datadir string,
	stemfile string,
	outfile string,
	typst string,
) {
	outdir := filepath.Join(datadir, "_pdf")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Fatalln(err)
		}
	}
	outpath := filepath.Join(outdir, outfile+".pdf")

	cmd := exec.Command(typst, "compile",
		"--font-path", filepath.Join(datadir, "_fonts"),
		"--root", datadir,
		"--input", "ifile="+filepath.Join("/_data", stemfile+".json"),
		filepath.Join(datadir, "scripts", script),
		outpath)
	//fmt.Printf(" ::: %s\n", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		base.Error.Println("(Typst) " + string(output))
		base.Error.Fatal(err)
	}
	base.Message.Printf("Timetable written to: %s\n", outpath)
}
