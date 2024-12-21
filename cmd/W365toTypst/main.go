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
	// Define and read command-line flags

	nocheck := flag.Bool("x", false, "Don't check for invalid placements")
	typstexec := flag.String("typst", "typst", "Typst executable")

	flag.Parse()

	// Get command-line argument: input file
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
	// Open logger
	logpath := stempath + ".log"
	base.OpenLog(logpath)
	stempath = strings.TrimSuffix(stempath, "_w365")

	// Read input file
	db := base.NewDb()
	w365tt.LoadJSON(db, abspath)
	db.PrepareDb()
	ttinfo := ttbase.MakeTtInfo(db)

	if !*nocheck {
		// Among other things (which are not relevant for the printing),
		// this checks placements
		ttinfo.PrepareCoreData()
	}

	datadir := filepath.Join(filepath.Dir(abspath), "typst_files")
	stemfile := filepath.Base(stempath)

	// Generate Typst data
	typst_files := ttprint.GenTypstData(ttinfo, datadir, stemfile)

	// Generate PDF files
	for _, tfile := range typst_files {
		ttprint.MakePdf(
			"print_timetable.typ", datadir, tfile, *typstexec)
	}

	base.Message.Println("OK")
}
