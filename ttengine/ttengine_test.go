package ttengine

import (
	"W365toFET/base"
	"W365toFET/readxml"
	"W365toFET/ttbase"
	"fmt"
	"slices"
	"testing"
)

//TODO: At some point the possibility to read from _db.json files
// would be good!

var inputfiles = []string{
	"../testdata/readxml/Demo1.xml",
	"../testdata/readxml/x01.xml",
}

func TestTtEngine(t *testing.T) {
	base.OpenLog("")
	for _, fxml := range inputfiles {
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

		tt(cdata.Db())

		/*
			stempath := strings.TrimSuffix(fxml, filepath.Ext(fxml))
			fjson := stempath + "_db.json"
			if cdata.Db().SaveDb(fjson) {
				fmt.Printf("\n ***** Written to: %s\n", fjson)
			} else {
				fmt.Println("\n ***** Write to JSON failed")
				continue
			}

			stempath = strings.TrimSuffix(stempath, "_w365")
			//toFET(cdata.db, stempath)
		*/
	}
}

func tt(db *base.DbTopLevel) {
	ttinfo := ttbase.MakeTtInfo(db)
	//ttinfo.PrintAtomicGroups()

	readDb(ttinfo)
}
