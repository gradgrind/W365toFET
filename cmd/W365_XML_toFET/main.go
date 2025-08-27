package main

import (
	"W365toFET/autotimetable"
	"W365toFET/base"
	"W365toFET/readxml"
	"W365toFET/timetable"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

var ifiles = []string{
	//"../../_testdata_N1/Demo1/Demo1.xml",
	"../../_testdata_N1/x01/x01.xml",
}

func main() {
	base.OpenLog("")
	for _, fxml := range ifiles {
		fmt.Println("\n ++++++++++++++++++++++")
		cdata := readxml.ConvertToDb(fxml)
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
				continue
			}
		}
		fmt.Printf("*** Using Schedule '%s'\n", sname)
		if !cdata.ReadSchedule(sname) {
			fmt.Println(" ... failed ...")
			continue
		}

		db := cdata.Db()
		db.PrepareDb()

		tt_data := timetable.BasicSetup(db)
		fmt.Printf("Resources: %d\n", len(tt_data.Resources))
		fmt.Printf("Activities: %d\n", len(tt_data.Activities)-1)

		abspath, err := filepath.Abs(fxml)
		if err != nil {
			log.Fatalf("*ERROR* Couldn't resolve file path: %s\n", fxml)
		}

		stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))
		autotimetable.SteerGeneration(tt_data, stempath)
	}
}
