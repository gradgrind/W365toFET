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
	teachers := flag.Bool("T", false, "Print individual teacher tables")
	classes := flag.Bool("C", false, "Print individual class tables")

	nocheck := flag.Bool("x", false, "Don't check for invalid placements")
	typstexec := flag.String("typst", "typst", "Typst executable")
	without_times := flag.Bool("nt", false, "Don't show lesson-period times")
	without_breaks := flag.Bool("nb", false, "Don't show breaks")

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

	//TODO: Provide plan_name somehow?
	plan_name := ""

	extraflags := map[string]bool{}
	if !*without_times {
		extraflags["WithTimes"] = true
	}
	if !*without_breaks {
		extraflags["WithBreaks"] = true
	}

	// Generate Typst data
	ttprint.GenTypstData(ttinfo, datadir, stemfile, plan_name, extraflags)

	// Optionally generate PDF files
	if *teachers {
		ttprint.MakePdf(
			"print_timetable.typ", datadir, stemfile+"_teachers", *typstexec)
	}
	if *classes {
		ttprint.MakePdf(
			"print_timetable.typ", datadir, stemfile+"_classes", *typstexec)
	}

	base.Message.Println("OK")
}
