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

func GenTypstData(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string,
) []string {
	typst_files := []string{}
	printTables := ttinfo.Db.PrintOptions.PrintTables
	if len(printTables) == 0 {
		printTables = []string{
			"Class", "Teacher", "Room",
			"Class_overview", "Teacher_overview", "Room_overview",
		}
	}
	// The same JSON is used for overview tables as for individual tables,
	// so suppress generation of doubles.
	done := map[string]string{}
	var f string
	var ok bool
	for _, ptable := range printTables {
		p, overview := strings.CutSuffix(ptable, "_overview")
		f, ok = done[p]
		if !ok {
			switch p {
			case "Class":
				f = getClasses(ttinfo, datadir, stemfile)
			case "Teacher":
				f = getTeachers(ttinfo, datadir, stemfile)
			case "Room":
				f = getRooms(ttinfo, datadir, stemfile)
			default:
				// Table for individual element
				f = genTypstOneElement(ttinfo, datadir, stemfile, p)
			}
			done[p] = f
		}
		if overview {
			f += "_overview"
		}
		typst_files = append(typst_files, f)
	}
	return typst_files
}

func genTypstOneElement(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string, // basic name part of source file
	id string,
) string {
	ref := base.Ref(id)
	e := ttinfo.Db.Elements[ref]
	c, ok := e.(*base.Class)
	if ok {
		// Make class JSON, but with only one class
		return getOneClass(ttinfo, datadir, stemfile, c)
	}
	t, ok := e.(*base.Teacher)
	if ok {
		// Make teacher JSON, but with only one teacher
		return getOneTeacher(ttinfo, datadir, stemfile, t)
	}
	r, ok := e.(*base.Room)
	if ok {
		// Make room JSON, but with only one room
		return getOneRoom(ttinfo, datadir, stemfile, r)
	}
	base.Error.Fatalf("Can't print timetable for invalid element: %+v\n", e)
	return ""
}

// TODO
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

func MakePdf(
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
