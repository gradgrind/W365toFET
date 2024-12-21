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
	plan_name string,
	flags map[string]bool,
) {
	genTypstClassData(ttinfo, plan_name, datadir, stemfile, flags)
	genTypstTeacherData(ttinfo, plan_name, datadir, stemfile, flags)
	genTypstRoomData(ttinfo, plan_name, datadir, stemfile, flags)
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

func MakePdf(script string, datadir string, stemfile string, typst string) {
	outdir := filepath.Join(datadir, "_pdf")
	if _, err := os.Stat(outdir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(outdir, os.ModePerm)
		if err != nil {
			base.Error.Fatalln(err)
		}
	}
	outpath := filepath.Join(outdir, stemfile+".pdf")

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
