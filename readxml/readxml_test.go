package readxml

import (
	"W365toFET/base"
	"W365toFET/fet"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func getXMLfile() string {
	/*
		const defaultPath = "../_testdata/*.xml"
		f365, err := zenity.SelectFile(
			zenity.Filename(defaultPath),
			zenity.FileFilter{
				Name:     "Waldorf-365 TT-export",
				Patterns: []string{"*.xml"},
				CaseFold: false,
			})
		if err != nil {
			log.Fatal(err)
		}
	*/

	//f365 := "../_testdata/x01_w365.xml"
	//f365 := "../_testdata/fms1_w365.xml"
	f365 := "../_testdata/Demo1_w365.xml"
	return f365
}

func Test2JSON(t *testing.T) {
	base.OpenLog("")
	fxml := getXMLfile()
	cdata := ConvertToDb(fxml)
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
	stempath := strings.TrimSuffix(fxml, filepath.Ext(fxml))
	fjson := stempath + "_db.json"
	if cdata.db.SaveDb(fjson) {
		fmt.Printf("\n ***** Written to: %s\n", fjson)
	} else {
		fmt.Println("\n ***** Write to JSON failed")
		return
	}

	stempath = strings.TrimSuffix(stempath, "_w365")
	toFET(cdata.db, stempath)
}

func toFET(db *base.DbTopLevel, fetpath string) {
	db.PrepareDb()

	// ********** Build the fet file **********
	xmlitem, lessonIdMap := fet.MakeFetFile(db)

	// Write FET file
	fetfile := fetpath + ".fet"
	f, err := os.Create(fetfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", fetfile)
	}
	defer f.Close()
	_, err = f.WriteString(xmlitem)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", fetfile)
	}
	base.Message.Printf("FET file written to: %s\n", fetfile)

	// Write Id-map file.
	mapfile := fetpath + ".map"
	fm, err := os.Create(mapfile)
	if err != nil {
		base.Bug.Fatalf("Couldn't open output file: %s\n", mapfile)
	}
	defer fm.Close()
	_, err = fm.WriteString(lessonIdMap)
	if err != nil {
		base.Bug.Fatalf("Couldn't write fet output to: %s\n", mapfile)
	}
	base.Message.Printf("Id-map written to: %s\n", mapfile)
	base.Message.Println("OK")
}
