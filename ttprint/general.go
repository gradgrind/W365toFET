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

func DEFAULT_PRINT_TABLES() []*base.PrintTable {
	return []*base.PrintTable{
		{Type: "Teacher", TypstTemplate: "print_timetable"},
		{Type: "Teacher", TypstTemplate: "print_overview"},
		{Type: "Class", TypstTemplate: "print_timetable"},
		{Type: "Class", TypstTemplate: "print_overview"},
		{Type: "Room", TypstTemplate: "print_timetable"},
		{Type: "Room", TypstTemplate: "print_overview"},
	}
}

type Tile struct {
	Day        int
	Hour       int
	Duration   int `json:",omitempty"`
	Fraction   int `json:",omitempty"`
	Offset     int `json:",omitempty"`
	Total      int `json:",omitempty"`
	Subject    string
	Groups     [][]string `json:",omitempty"`
	Teachers   []string   `json:",omitempty"`
	Rooms      []string   `json:",omitempty"`
	Background string     `json:",omitempty"`
	Footnote   string     `json:",omitempty"`
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

// ttPage basic fields:
//
//	Name       []string // to allow ["FirstName", "LastName"]
//	Short      string
//	Activities []Tile
//
// Others can be added from the Pages entries in the input PrintTable.
type ttPage map[string]any

type xPage struct {
	key   string
	value any
}

func (page ttPage) extendPage(x []xPage) {
	for _, xp := range x {
		page[xp.key] = xp.value
	}
}

func GenTimetables(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string,
	commands []*base.PrintTable,
	genpdf string,
) {
	var f string
	var tt Timetable
	var typstData map[string]any
	if len(commands) == 0 {
		commands = DEFAULT_PRINT_TABLES()
	}

	for _, cmd := range commands {
		// Collect the "Pages" data from the PrintTable
		pageData := map[base.Ref][]xPage{}
		for _, pd := range cmd.Pages {
			ref := base.Ref(pd["Id"].(string))
			pdlist := make([]xPage, len(pd)-1)
			i := 0
			for k, v := range pd {
				if k != "Id" {
					pdlist[i] = xPage{k, v}
					i += 1
				}
			}
			pageData[ref] = pdlist
		}
		f = cmd.TypstJson
		typstData = cmd.Typst
		switch cmd.Type {
		case "Class":
			pages := getClasses(ttinfo, pageData)
			tt = timetable(ttinfo.Db, pages, typstData, "Class")
			if f == "" {
				f = stemfile + "_classes"
			}
		case "Teacher":
			pages := getTeachers(ttinfo, pageData)
			tt = timetable(ttinfo.Db, pages, typstData, "Teacher")
			if f == "" {
				f = stemfile + "_teachers"
			}
		case "Room":
			pages := getRooms(ttinfo, pageData)
			tt = timetable(ttinfo.Db, pages, typstData, "Room")
			if f == "" {
				f = stemfile + "_rooms"
			}
		default:
			// Table for individual element
			var tag string
			tt, tag = genTypstOneElement(ttinfo, pageData, cmd)
			if f == "" {
				f = stemfile + tag
			}
		}
		makeTypstJson(tt, datadir, f)

		if genpdf != "" {
			tmpl := cmd.TypstTemplate
			pdf := cmd.Pdf
			if pdf == "" {
				if strings.HasSuffix(tmpl, "_overview") {
					pdf = f + "_overview"
				} else {
					pdf = f
				}
			}
			makePdf(tmpl, datadir, f, pdf, genpdf)
		}
	}
}

func genTypstOneElement(
	ttinfo *ttbase.TtInfo,
	pagemap map[base.Ref][]xPage,
	cmd *base.PrintTable,
) (Timetable, string) {
	var tt Timetable
	ref := base.Ref(cmd.Type)
	typstData := cmd.Typst
	e := ttinfo.Db.Elements[ref]
	c, ok := e.(*base.Class)
	if ok {
		// Make class JSON, but with only one class
		pages := getOneClass(ttinfo, pagemap, c)
		tt = timetable(ttinfo.Db, pages, typstData, "Class")
		return tt, "class_" + c.Tag
	}
	t, ok := e.(*base.Teacher)
	if ok {
		// Make teacher JSON, but with only one teacher
		pages := getOneTeacher(ttinfo, pagemap, t)
		tt = timetable(ttinfo.Db, pages, typstData, "Teacher")
		return tt, "teacher_" + t.Tag
	}
	r, ok := e.(*base.Room)
	if ok {
		// Make room JSON, but with only one room
		pages := getOneRoom(ttinfo, pagemap, r)
		tt = timetable(ttinfo.Db, pages, typstData, "Room")
		return tt, "room_" + r.Tag
	}
	base.Error.Fatalf("Can't print timetable for invalid element: %+v\n", e)
	return tt, ""
}

func timetable(
	db *base.DbTopLevel,
	pages []ttPage,
	typstData map[string]any,
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
	clist := [][]any{}
	for _, e := range db.Classes {
		clist = append(clist, []any{
			e.Tag,
			e.Name,
		})
	}
	tlist := [][]any{}
	for _, e := range db.Teachers {
		tlist = append(tlist, []any{
			e.Tag,
			[]string{e.Firstname, e.Name},
		})
	}
	rlist := [][]any{}
	for _, e := range db.Rooms {
		rlist = append(rlist, []any{
			e.Tag,
			e.Name,
		})
	}
	slist := [][]any{}
	for _, e := range db.Subjects {
		slist = append(slist, []any{
			e.Tag,
			e.Name,
		})
	}
	info := map[string]any{
		"Institution":  db.Info.Institution,
		"Days":         dlist,
		"Hours":        hlist,
		"ClassNames":   clist,
		"TeacherNames": tlist,
		"RoomNames":    rlist,
		"SubjectNames": slist,
	}
	return Timetable{
		TableType: tabletype,
		Info:      info,
		Typst:     typstData,
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
		filepath.Join(datadir, "scripts", script+".typ"),
		outpath)
	//fmt.Printf(" ::: %s\n", cmd.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		base.Error.Println("(Typst) " + string(output))
		base.Error.Fatal(err)
	}
	base.Message.Printf("Timetable written to: %s\n", outpath)
}

func splitGroups(glist []string) [][]string {
	gplist := [][]string{}
	for _, g := range glist {
		gs := strings.Split(g, ttbase.CLASS_GROUP_SEP)
		if len(gs) == 1 {
			gs = append(gs, "")
		}
		gplist = append(gplist, gs)
	}
	return gplist
}
