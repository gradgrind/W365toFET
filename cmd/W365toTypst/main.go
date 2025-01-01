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

func main() {
	// Define and read command-line flags

	nocheck := flag.Bool("x", false, "Don't check for invalid placements")
	typstexec := flag.String("typst", "typst", "Typst executable")
	nopdf := flag.Bool("np", false, "Don't run Typst")

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

	// Commands:
	printTables := ttinfo.Db.PrintTables
	if len(printTables) == 0 {
		printTables = DEFAULT_PRINT_TABLES()
	}

	// Generate Typst data and, if not suppressed, PDF output.
	genpdf := *typstexec
	if *nopdf {
		genpdf = ""
	}
	ttprint.GenTimetables(ttinfo, datadir, stemfile, printTables, genpdf)
	base.Message.Println("OK")
}
