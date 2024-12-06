package main

import (
	"W365toFET/base"
	"W365toFET/ttbase"
	"W365toFET/ttprint"
	"W365toFET/w365tt"
	"flag"
	"log"
	"path/filepath"
	"strings"
)

func main() {

	//wordPtr := flag.String("word", "foo", "a string")
	//numbPtr := flag.Int("numb", 42, "an int")
	//forkPtr := flag.Bool("fork", false, "a bool")
	//var svar string
	//flag.StringVar(&svar, "svar", "bar", "a string var")

	teachers := flag.Bool("t", false, "Print individual teacher tables")

	flag.Parse()

	//fmt.Println("word:", *wordPtr)
	//fmt.Println("numb:", *numbPtr)
	//fmt.Println("fork:", *forkPtr)
	//fmt.Println("svar:", svar)
	//fmt.Println("args:", flag.Args())

	args := flag.Args()
	if len(args) != 1 {
		if len(args) == 0 {
			log.Fatalln("ERROR* No input file")
		}
		log.Fatalf("*ERROR* Too many command-line arguments:\n  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("*ERROR* Couldn't resolve file path: %s\n", args[0])
	}

	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
	logpath := stempath + ".log"
	base.OpenLog(logpath)
	stempath = strings.TrimSuffix(stempath, "_w365")

	db := base.NewDb()
	w365tt.LoadJSON(db, abspath)
	db.PrepareDb()
	ttinfo := ttbase.MakeTtInfo(db)

	datadir := filepath.Join(filepath.Dir(abspath), "typst_files")
	stemfile := filepath.Base(stempath)
	ttprint.GenTypstData(ttinfo, datadir, stemfile)

	//TODO option to pass in typst path
	if *teachers {
		ttprint.MakePdf("print_timetable.typ", datadir, stemfile+"_teachers", "")
	}
	base.Message.Println("OK")
}
