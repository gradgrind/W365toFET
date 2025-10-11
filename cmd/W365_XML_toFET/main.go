package main

import (
	"W365toFET/autotimetable"
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/readxml"
	"W365toFET/timetable"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

//Input files:
//  "../../_testdata_N1/Demo1/Demo1.xml"
//  "../../_testdata_N1/x01/x01.xml"

func main() {

	flag.BoolVar(&base.CONSOLE, "c", false, "enable progress output")
	flag.BoolVar(&autotimetable.TESTING, "T", false, "run in testing mode")
	timeout := flag.Int("t", 300, "set timeout")
	nprocesses := flag.Int("p", 0, "max. parallel processes")

	flag.Parse()

	if *nprocesses > 0 {
		autotimetable.MAXPROCESSES = *nprocesses
	}

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

	//base.OpenLog("")
	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
	logpath := stempath + ".log"
	base.OpenLog(logpath)
	//stempath = strings.TrimSuffix(stempath, "_w365")

	cdata := readxml.ConvertToDb(abspath)
	fmt.Println("*** Available Schedules:")
	slist := cdata.ScheduleNames()
	for _, sname := range slist {
		fmt.Printf("  -- %s\n", sname)
	}
	sname := "Vorlage"
	if !slices.Contains(slist, sname) {
		if len(slist) != 0 {
			sname = slist[0]
		} else {
			fmt.Println(" ... stopping ...")
			return
		}
	}
	fmt.Printf("*** Using Schedule '%s'\n", sname)
	if !cdata.ReadSchedule(sname) {
		fmt.Println(" ... failed ...")
		return
	}

	// This allows for an option to select different generator back-ends
	fet.Setup()

	db := cdata.Db()
	db.PrepareDb()

	db.SaveDb(stempath + "_DB.json")

	// May want to change this with a different back-end ...
	workingdir := stempath + "_fet"

	tt_data := timetable.BasicSetup(db, workingdir)
	base.Report(fmt.Sprintf("Atomic Groups: %d\n",
		len(tt_data.SharedData.AtomicNodes)))
	base.Report(fmt.Sprintf("Teachers: %d\n",
		len(tt_data.SharedData.Db.Teachers)))
	base.Report(fmt.Sprintf("Rooms: %d\n",
		len(tt_data.SharedData.Db.Rooms)))
	base.Report(fmt.Sprintf("Activities: %d\n",
		len(tt_data.SharedData.Activities)-1))

	autotimetable.StartGeneration(tt_data, *timeout)
}
