package ttprint

import (
	"W365toFET/base"
	"W365toFET/fet"
	"W365toFET/ttbase"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var inputfiles = []string{
	"../testdata/readxml/Demo1_db.json",
	"../testdata/readxml/x01_db.json",
}

func TestPrint(t *testing.T) {
	base.OpenLog("")
	datadir, err := filepath.Abs("../typst_files/")
	if err != nil {
		base.Error.Fatal(err)
	}
	fmt.Println("\n############## TestPrint")
	for _, f := range inputfiles {
		f, err := filepath.Abs(f)
		if err != nil {
			base.Error.Fatal(err)
		}
		fmt.Println("\n ++++++++++++++++++++++")
		db := base.LoadDb(f)
		db.PrepareDb()
		ttinfo := ttbase.MakeTtInfo(db)

		// Disabling this will prevent testing of placements.
		// Note that times loaded from the activities.xml file are not checked
		// anyway!
		ttinfo.PrepareCoreData()

		stempath := strings.TrimSuffix(f, filepath.Ext(f))
		stempath = strings.TrimSuffix(stempath, "_db")
		doPrinting(ttinfo, datadir, stempath)
	}
}

func doPrinting(ttinfo *ttbase.TtInfo, datadir string, stempath string) {
	// Get activity mapping
	mapfile := stempath + ".map"
	activityMap := readActivityMap(mapfile)

	// Get placements
	pfile := stempath + "_activities.xml"
	pmap := map[base.Ref]fet.ActivityPlacement{}
	for _, p := range fet.ReadPlacements(ttinfo, pfile) {
		//fmt.Printf("### %+v\n  -- %s\n", p, activityMap[p.Id])
		pmap[activityMap[p.Id]] = p
	}
	for aix := 1; aix < len(ttinfo.Activities); aix++ {
		a := ttinfo.Activities[aix]
		l := a.Lesson
		p, ok := pmap[l.Id]
		if !ok {
			fmt.Printf("### %+v\n  -%v- %s\n", p, ok, l.Id)
		}
		l.Rooms = p.Rooms
		l.Day = p.Day
		l.Hour = p.Hour
		a.Placement = p.Day*ttinfo.NHours + p.Hour
	}
	plan_name := "Test Plan"

	stemfile := filepath.Base(stempath)
	flags := map[string]bool{
		"WithTimes":  true,
		"WithBreaks": true,
	}
	GenTypstData(ttinfo, datadir, stemfile, plan_name, flags)

	typst := "typst"
	MakePdf("print_timetable.typ", datadir, stemfile+"_teachers", typst)
	MakePdf("print_timetable.typ", datadir, stemfile+"_classes", typst)

	/*
		PrintRoomTimetables(lessons, plan_name, datadir,
			strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume.pdf")
		PrintRoomOverview(lessons, plan_name, datadir,
			strings.TrimSuffix(abspath, filepath.Ext(abspath))+"_Räume-gesamt.pdf")
	*/
}

func readActivityMap(mapfile string) map[int]base.Ref {
	amap := map[int]base.Ref{}
	infile, err := os.Open(mapfile)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer infile.Close()
	// Read the file line by line using scanner
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		iref := line
		i, err := strconv.Atoi(strings.TrimSpace(iref[0]))
		if err != nil {
			base.Error.Fatalf("Invalid line in %s:\n  \"%s\"\n", mapfile, line)
		}
		//fmt.Printf("line: %d: %+v\n", i, iref[1])
		amap[i] = base.Ref(strings.TrimSpace(iref[1]))
	}
	if err := scanner.Err(); err != nil {
		base.Error.Fatal(err)
	}
	return amap
}
