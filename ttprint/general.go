package ttprint

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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

func GenTypstData(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stemfile string,
) {
	//genTypstClassData(ttinfo, "", datadir, stemfile)
	genTypstTeacherData(ttinfo, "", datadir, stemfile)
}

/* TODO--
func PrintTimetables(
	ttinfo *ttbase.TtInfo,
	datadir string,
	stempath string,
) {
	outdir := filepath.Join(filepath.Dir(stempath), "_pdf")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Println(err)
		}
	}
	outpath := filepath.Join(outdir, filepath.Base(stempath))
	PrintClassTimetables(
		ttinfo, "", datadir, outpath+"_Klassen.pdf")
	PrintTeacherTimetables(
		ttinfo, "", datadir, outpath+"_Lehrer.pdf")
}
*/

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
			base.Error.Println(err)
		}
	}
	jsonpath := filepath.Join(outdir, outfile+".json")
	err = os.WriteFile(jsonpath, b, 0666)
	if err != nil {
		base.Error.Fatal(err)
	}
}

func MakePdf(script string, datadir string, stemfile string, typst string) {
	if typst == "" {
		typst = "typst"
	}
	outdir := filepath.Join(datadir, "_pdf")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Fatalln(err)
		}
	}
	outpath := filepath.Join(outdir, stemfile+".pdf")
	//TODO: The output file path should somehow reflect the .typ script.
	//"testx01_teachers.json"
	//"print_timetable.typ"

	cmd := exec.Command(typst, "compile",
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
