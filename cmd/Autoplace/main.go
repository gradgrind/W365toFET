package main

import (
	"W365toFET/base"
	"W365toFET/readxml"
	"W365toFET/ttbase"
	"W365toFET/ttengine"
	"fmt"
	"slices"
)

var ifiles = []string{
	"../testdata/readxml/Demo1.xml",
	"../testdata/readxml/x01.xml",
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
		ttbase.MakeTtInfo(db)
		ttinfo := ttbase.MakeTtInfo(db)
		ttinfo.PrepareCoreData()

		alist := ttengine.CollectCourseLessons(ttinfo)
		//fmt.Printf("??? %+v\n", alist)

		ttengine.PlaceLessons(ttinfo, alist)
	}
}