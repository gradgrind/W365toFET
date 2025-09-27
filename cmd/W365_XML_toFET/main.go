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
	flag.Parse()
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

	base.OpenLog("")

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

	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
	db.SaveDb(stempath + "_DB.json")

	// May want to change this with a different back-end ...
	workingdir := stempath + "_fet"

	tt_data := timetable.BasicSetup(db, workingdir)
	fmt.Printf("Resources: %d\n", len(tt_data.SharedData.Resources))
	fmt.Printf("Activities: %d\n", len(tt_data.SharedData.Activities)-1)

	timeout := 200 // seconds
	autotimetable.StartGeneration(tt_data, timeout)
}
